package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Verrouillage automatique après échecs de PIN successifs (Auth/OTP Rate Limiting
// Hardening) : 6 échecs en 15 minutes déclenchent un blocage temporaire de 15 minutes.
// Valeurs fixes (pas de variable d'env dédiée) car aucune n'existait déjà pour ce cas
// précis — cohérent avec la demande de ne pas introduire de config supplémentaire non
// nécessaire.
const (
	loginFailWindow    = 15 * time.Minute
	loginMaxFailures   = 6
	loginLockoutPeriod = 15 * time.Minute
)

// AuthService expose DEUX contrats de contenu Register/ResetPassword côte à côte —
// "Legacy" (v1, /v1/api/auth/...) et le contrat par défaut (v2, /v2/api/auth/...) — pour
// permettre au Backend d'être déployé AVANT que la mise à jour Flutter n'atteigne tous
// les utilisateurs de la version déjà publiée sur Google Play (API Versioning Phase 1).
//
// La version publiée actuelle de l'app envoie déjà : country_code parmi
// {+222,+221,+212,+216} uniquement (jamais country_iso), et un PIN à 4 chiffres
// numériques. Casser cet appel avant que l'app mise à jour ne soit largement installée
// bloquerait l'inscription ET la réinitialisation de mot de passe pour TOUS les
// utilisateurs existants — inacceptable. RegisterLegacy/ResetPasswordLegacy reproduisent
// donc EXACTEMENT le comportement du contrat publié (voir historique git, commit de
// baseline avant cette fonctionnalité), sans aucune régression, y compris le refus des
// pays hors de la liste fermée historique.
//
// Register/ResetPassword (sans suffixe) restent le contrat v2 strict déjà implémenté
// (country_iso obligatoire, validation libphonenumber par région, mot de passe 8-72
// caractères) — c'est le contrat que la nouvelle version de Flutter doit utiliser
// exclusivement (voir mobile/lib/services/auth_api.dart, endpoints /v2/api/auth/...).
//
// v1 est TEMPORAIRE : à retirer une fois l'ancienne version de l'app suffisamment
// désinstallée/mise à jour (suivre le taux d'appels sur /v1/api/auth/register en
// production avant de planifier son retrait — ne pas retirer sur une base arbitraire).
type AuthService interface {
	// Register (v2, /v2/api/auth/register) — countryISO doit être un code région ISO-2
	// (ex: "MR", "TN", "US", "CA") — requis pour distinguer les pays qui partagent un
	// même indicatif téléphonique (ex: +1 pour US/CA/plusieurs Caraïbes). Voir
	// NormalizeE164 (phone_service.go). N'accepte QUE ce contrat strict — un appelant ne
	// peut jamais obtenir le comportement legacy (4 pays, sans country_iso) en omettant
	// simplement le champ : country_iso est requis dès la validation du DTO
	// (auth_handler.go RegisterRequest), avant même d'atteindre cette fonction.
	Register(ctx context.Context, phone, pin, fullName, email, city, countryISO string) error
	// RegisterLegacy (v1, /v1/api/auth/register) — reproduit EXACTEMENT le contrat publié
	// avant cette fonctionnalité : countryCode est un indicatif d'appel (ex: "+222"),
	// restreint aux 4 pays historiques (validé ici, pas seulement côté DTO), PIN 4
	// chiffres numériques (la contrainte de longueur reste au niveau DTO
	// RegisterRequestLegacy, mais la validation de force reste identique). Ne JAMAIS
	// étendre ce contrat — toute nouvelle capacité va exclusivement dans Register (v2).
	RegisterLegacy(ctx context.Context, phone, pin, fullName, email, city, countryCode string) error
	// countryISO est optionnel ici : un appelant ancien (app non mise à jour) qui ne
	// l'envoie pas retombe sur la recherche legacy par 'phone' brut (pont de
	// compatibilité — voir Login). Un appelant à jour doit toujours l'envoyer. Login est
	// PARTAGÉ entre v1 et v2 (un seul endpoint /auth/login pour les deux, voir routes.go)
	// car il accepte déjà indifféremment un PIN 4 chiffres ou un mot de passe 8+.
	Login(ctx context.Context, phone, countryISO, pin string) (string, *models.User, error) // token, user, err
	SendOTP(ctx context.Context, phone, purpose, ip string) error
	VerifyOTP(ctx context.Context, phone, code, purpose string) error
	// ResetPassword est PARTAGÉE entre v1 et v2 (contrairement à Register) : sa logique
	// interne est identique au comportement publié (correspondance exacte de 'phone',
	// aucune normalisation E.164 obligatoire) — seule la politique de longueur de newPin
	// diffère, et elle est imposée en amont par la validation du DTO (voir
	// ResetPasswordRequestLegacy en 4 chiffres vs ResetPasswordRequest en 8-72
	// caractères dans auth_handler.go), jamais dans cette fonction elle-même.
	ResetPassword(ctx context.Context, phone, code, newPin string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPin, newPin string) error
	GenerateJWT(userID uuid.UUID, role string, isSuperAdmin bool) (string, error)
	ValidateJWT(tokenString string) (*JWTClaims, error)
	TrackPasswordReset(ctx context.Context, phone, ip string) error
}

// ValidCountryCodesLegacy liste les 4 codes pays acceptés par le contrat v1
// (/v1/api/auth/...) — c'est EXACTEMENT la liste fermée du contrat publié sur Google
// Play avant cette fonctionnalité (voir git history). Ne jamais l'étendre : toute
// nouvelle capacité (nouveaux pays) n'existe que dans le contrat v2 (Register).
var ValidCountryCodesLegacy = map[string]string{
	"+222": "MR", // Mauritanie
	"+221": "SN", // Sénégal
	"+212": "MA", // Maroc
	"+216": "TN", // Tunisie
}

type JWTClaims struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

// RevokedBeforeKey retourne la clé Redis portant l'horodatage de révocation globale
// d'un utilisateur (Session Security Phase 1 — user-level revocation). Tout JWT émis
// (IssuedAt) avant cet horodatage doit être rejeté par le middleware JWT, même s'il
// est par ailleurs valide (signature correcte, non expiré, non blacklisté
// individuellement). Fonction partagée entre services (écriture, lors d'un blocage ou
// d'une réinitialisation de PIN) et middleware (lecture, à chaque requête).
func RevokedBeforeKey(userID uuid.UUID) string {
	return fmt.Sprintf("revoked_before:user:%s", userID.String())
}

// revocationTTLSafetyMargin est ajouté à la durée de vie max d'un JWT pour calculer le
// TTL de la clé revoked_before : couvre les dérives d'horloge et la latence réseau
// entre l'émission d'un token et son premier contrôle côté serveur.
const revocationTTLSafetyMargin = 1 * time.Hour

// RevokeUserSessions marque tous les JWT actuellement émis pour cet utilisateur comme
// invalides (Session Security Phase 1). N'échoue jamais bruyamment : si Redis est
// indisponible, l'appelant (blocage utilisateur, réinitialisation de PIN) continue
// normalement — cette révocation est une couche de défense supplémentaire, pas une
// condition bloquante pour l'opération principale qui l'a déclenchée.
//
// jwtExpiryHours doit être la même valeur que cfg.JWT.ExpiryHours utilisée pour
// signer les tokens (JWT_EXPIRY_HOURS) : le TTL ne doit jamais être un nombre fixe
// indépendant de cette configuration, sous peine qu'un changement futur de
// JWT_EXPIRY_HOURS rende ce TTL trop court et laisse une fenêtre de réutilisation
// d'un token pourtant révoqué (voir revue Session Security Phase 1).
func RevokeUserSessions(ctx context.Context, rdb *redis.Client, logger *zap.Logger, userID uuid.UUID, jwtExpiryHours int) {
	if err := RevokeUserSessionsErr(ctx, rdb, userID, jwtExpiryHours); err != nil {
		if logger != nil {
			logger.Error("RevokeUserSessions: failed to write revocation marker", zap.Error(err), zap.String("user_id", userID.String()))
		}
	}
}

// RevokeUserSessionsErr est la même opération que RevokeUserSessions mais remonte
// l'erreur Redis à l'appelant au lieu de l'avaler (Blocked Phone Phase 1C) : utilisée
// uniquement là où un échec silencieux de révocation serait trompeur pour l'admin
// (BlockPhone doit savoir si les sessions n'ont pas pu être invalidées).
func RevokeUserSessionsErr(ctx context.Context, rdb *redis.Client, userID uuid.UUID, jwtExpiryHours int) error {
	if rdb == nil {
		return fmt.Errorf("redis unavailable: cannot revoke sessions")
	}
	key := RevokedBeforeKey(userID)
	ttl := time.Duration(jwtExpiryHours)*time.Hour + revocationTTLSafetyMargin
	return rdb.Set(ctx, key, time.Now().Unix(), ttl).Err()
}

type authService struct {
	userRepo       repository.UserRepository
	jwtSecret      string
	jwtExpiry      int
	env            string
	developmentOTP string // Code OTP de développement
	smsService     SMSService
	smsEnabled     bool // true only when SMS credentials are configured
	otpLength      int
	otpTTL         time.Duration
	rdb            *redis.Client
	logger         *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry int, env string, devOTP string, sms SMSService, otpLength int, otpTTLMinutes int, rdb *redis.Client, logger *zap.Logger) AuthService {
	// En production, devOTP doit être vide
	if env != "development" {
		devOTP = ""
	}
	if otpLength == 0 {
		otpLength = 4 // Default 4 digits
	}
	if otpTTLMinutes <= 0 {
		otpTTLMinutes = 5
	}
	// Detect if SMS credentials are actually configured
	smsEnabled := sms != nil && sms.IsConfigured()
	return &authService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
		env:            env,
		developmentOTP: devOTP,
		smsService:     sms,
		smsEnabled:     smsEnabled,
		otpLength:      otpLength,
		otpTTL:         time.Duration(otpTTLMinutes) * time.Minute,
		rdb:            rdb,
		logger:         logger,
	}
}

func (s *authService) Register(ctx context.Context, phone, pin, fullName, email, city, countryISO string) error {
	// Blocage admin au niveau du numéro (Blocked Phone Phase 1A) — table blocked_phones,
	// distincte de users.blocked_until. Erreur générique : ne révèle jamais la raison,
	// la durée, ni si le numéro est par ailleurs déjà enregistré.
	if blocked, err := s.userRepo.IsPhoneBlocked(ctx, phone); err == nil && blocked {
		return apperr.ErrPhoneUnavailable
	}

	// Support téléphonique international (migration 000044, corrigé en Phase 2) :
	// countryISO doit être un code région ISO-2 explicite (ex: "US" vs "CA"), jamais un
	// simple indicatif d'appel — plusieurs pays partagent le même indicatif (+1, +7, ...)
	// et un indicatif seul ne suffit pas à identifier le pays réel. NormalizeE164 valide
	// le numéro spécifiquement pour cette région ; il échoue si le numéro n'est pas
	// valide pour ELLE, même s'il serait valide pour un autre pays du même indicatif.
	e164, isoRegion, err := NormalizeE164(phone, countryISO)
	if err != nil {
		return apperr.ErrInvalidPhone
	}

	// Vérifier si le numéro (normalisé) existe déjà — recherche par E.164 canonique,
	// robuste aux différents formats d'entrée pour un même numéro logique.
	existing, _ := s.userRepo.FindByPhoneE164(ctx, e164)
	if existing != nil {
		return apperr.ErrDuplicatePhone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Auto-verify user when SMS is not configured so login works immediately.
	// When Wablas credentials are set, user must verify via OTP first.
	isVerified := !s.smsEnabled

	user := &models.User{
		ID: uuid.New(),
		// Phone (legacy) reste peuplé avec la valeur normalisée pour compatibilité avec
		// tout code existant qui lit encore ce champ (recherche, affichage, OTP).
		Phone:           e164,
		PhoneE164:       &e164,
		PhoneCountryISO: &isoRegion,
		PasswordHash:    string(hash),
		FullName:        stringPtr(fullName),
		Email:           stringPtr(email),
		City:            stringPtr(city),
		CountryCode:     stringPtr(countryISO),
		LanguagePref:    "ar",
		Role:            "user",
		IsActive:        true,
		IsVerified:      isVerified,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		// AccountCountryISO (migration 000046): the canonical MazadPay account
		// market, set explicitly from the client-supplied, server-validated
		// country_iso (required, len=2, see auth_handler.go RegisterRequest) --
		// NOT from isoRegion (phone-number-region metadata, see PhoneCountryISO
		// above). These should agree in practice (NormalizeE164 validates the
		// phone specifically for the claimed countryISO region), but are
		// structurally distinct fields serving distinct purposes; account market
		// must reflect what the user explicitly selected, not what was detected
		// from their phone number.
		AccountCountryISO: stringPtr(countryISO),
	}

	return s.userRepo.Create(ctx, user)
}

// RegisterLegacy (API Versioning Phase 1, /v1/api/auth/register) reproduit EXACTEMENT
// le comportement du contrat publié sur Google Play avant l'introduction du support
// téléphonique international : liste fermée à 4 pays, countryCode = indicatif d'appel
// (pas ISO-2), phone stocké tel qu'envoyé par le client (pas de normalisation E.164
// obligatoire — un numéro mal formé passait déjà avant, et doit continuer à passer
// exactement pareil ici, sinon des inscriptions qui fonctionnaient hier échoueraient
// après ce déploiement backend, cassant la version déjà publiée).
//
// Amélioration SANS RISQUE ajoutée ici : on tente, en best-effort, de renseigner
// phone_e164/phone_country_iso via NormalizeE164 avec le pays legacy déjà connu
// (ValidCountryCodesLegacy[countryCode]) — mais un échec de cette tentative n'affecte
// JAMAIS le résultat de l'inscription (pas de "return" sur erreur ici), contrairement à
// Register (v2) où NormalizeE164 est bloquante. Ainsi on "normalise et stocke de façon
// sûre autant que possible" (comme demandé) sans jamais rejeter une inscription que le
// contrat v1 publié aurait acceptée.
func (s *authService) RegisterLegacy(ctx context.Context, phone, pin, fullName, email, city, countryCode string) error {
	if blocked, err := s.userRepo.IsPhoneBlocked(ctx, phone); err == nil && blocked {
		return apperr.ErrPhoneUnavailable
	}

	if countryCode == "" {
		countryCode = "+222" // Default historique : Mauritanie
	}
	isoRegion, valid := ValidCountryCodesLegacy[countryCode]
	if !valid {
		return fmt.Errorf("code pays non supporté: %s (supportés: +222, +221, +212, +216)", countryCode)
	}

	// Recherche de doublon IDENTIQUE au contrat publié : correspondance exacte de
	// 'phone' tel qu'envoyé, PAS de normalisation E.164 (une normalisation obligatoire
	// ici pourrait rejeter un numéro que le contrat v1 acceptait auparavant).
	existing, _ := s.userRepo.FindByPhone(ctx, phone)
	if existing != nil {
		return apperr.ErrDuplicatePhone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	isVerified := !s.smsEnabled

	user := &models.User{
		ID:           uuid.New(),
		Phone:        phone,
		PasswordHash: string(hash),
		FullName:     stringPtr(fullName),
		Email:        stringPtr(email),
		City:         stringPtr(city),
		CountryCode:  stringPtr(countryCode),
		LanguagePref: "ar",
		Role:         "user",
		IsActive:     true,
		IsVerified:   isVerified,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		// AccountCountryISO (migration 000046): isoRegion is already known and
		// reliable here (from ValidCountryCodesLegacy, the closed 4-country
		// legacy set), so the account market is set explicitly rather than left
		// NULL-to-fallback -- more precise than the generic MR-only fallback for
		// the 3 legacy non-MR countries (SN/MA/TN).
		AccountCountryISO: &isoRegion,
	}

	// Best-effort seulement : ne bloque jamais l'inscription, voir commentaire de
	// fonction. isoRegion vient de ValidCountryCodesLegacy, donc déjà connu et fiable —
	// on ne fait confiance qu'à NormalizeE164 pour la validité du NUMÉRO lui-même dans
	// cette région, pas pour déterminer la région (déjà fixée par countryCode ici).
	if e164, normalizedISO, normErr := NormalizeE164(phone, isoRegion); normErr == nil {
		user.PhoneE164 = &e164
		user.PhoneCountryISO = &normalizedISO
	}

	return s.userRepo.Create(ctx, user)
}

// loginFailKey retourne la clé Redis du compteur d'échecs de PIN pour un téléphone
// donné — distincte des clés de rate limiting par route (voir middleware/ratelimit.go).
func loginFailKey(phone string) string {
	return fmt.Sprintf("login_fail:%s", phone)
}

func (s *authService) Login(ctx context.Context, phone, countryISO, pin string) (string, *models.User, error) {
	// Blocage admin au niveau du numéro (Blocked Phone Phase 1A) — vérifié avant même
	// FindByPhone pour ne rien révéler de plus que l'erreur générique existante.
	if blocked, err := s.userRepo.IsPhoneBlocked(ctx, phone); err == nil && blocked {
		return "", nil, apperr.ErrPhoneUnavailable
	}

	// Pont de compatibilité E.164 (migration 000044, corrigé en Phase 2) : on tente
	// d'abord une recherche par le numéro normalisé — UNIQUEMENT si countryISO est fourni
	// (un client à jour l'envoie toujours). NormalizeE164 exige désormais une région
	// explicite ; sans elle on n'essaie même pas la normalisation, et on retombe
	// directement sur la correspondance exacte de l'ancien champ 'phone' tel qu'envoyé
	// par le client — c'est le pont de compatibilité pour les appels d'une éventuelle
	// ancienne version de l'app qui n'enverrait pas encore country_iso, et pour les
	// comptes pas encore backfillés (voir cmd/backfill_phone_e164). La normalisation
	// reste best-effort même quand countryISO est fourni : un échec ne bloque jamais le
	// login, on utilise simplement le fallback.
	var user *models.User
	if countryISO != "" {
		if e164, _, normErr := NormalizeE164(phone, countryISO); normErr == nil {
			if u, findErr := s.userRepo.FindByPhoneE164(ctx, e164); findErr == nil {
				user = u
			}
		}
	}
	if user == nil {
		u, err := s.userRepo.FindByPhone(ctx, phone)
		if err != nil {
			return "", nil, apperr.ErrUserNotFound
		}
		user = u
	}

	// Vérifier si le compte est actif
	if !user.IsActive {
		return "", nil, apperr.ErrAccountBlocked
	}

	// Vérifier le blocage temporaire (déjà actif, posé soit manuellement par un admin,
	// soit automatiquement ci-dessous après des échecs de PIN répétés)
	if user.BlockedUntil != nil && time.Now().Before(*user.BlockedUntil) {
		return "", nil, apperr.ErrAccountBlocked
	}

	if !user.IsVerified {
		return "", nil, apperr.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pin)); err != nil {
		s.recordFailedLogin(ctx, phone)
		return "", nil, apperr.ErrInvalidPin
	}

	// PIN correct : réinitialiser le compteur d'échecs
	if s.rdb != nil {
		s.rdb.Del(ctx, loginFailKey(phone))
	}

	token, err := s.GenerateJWT(user.ID, user.Role, user.IsSuperAdmin)
	if err != nil {
		return "", nil, err
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log l'erreur mais ne bloque pas la connexion
		fmt.Println("Warning: failed to update last login for user", user.ID, ":", err)
	}
	return token, user, nil
}

// recordFailedLogin incrémente le compteur d'échecs de PIN pour ce téléphone et pose
// un blocage temporaire (via BlockedUntil, déjà utilisé par le blocage manuel admin)
// une fois le seuil atteint — verrouillage automatique déclenché uniquement par un
// échec de PIN réel, jamais par le rate limiting de route (Auth/OTP Rate Limiting
// Hardening).
func (s *authService) recordFailedLogin(ctx context.Context, phone string) {
	if s.rdb == nil {
		return
	}
	key := loginFailKey(phone)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		if s.logger != nil {
			s.logger.Error("recordFailedLogin: Redis error", zap.Error(err))
		}
		return
	}
	if count == 1 {
		s.rdb.Expire(ctx, key, loginFailWindow)
	}
	if count >= loginMaxFailures {
		until := time.Now().Add(loginLockoutPeriod)
		if err := s.userRepo.SetBlockedUntil(ctx, phone, until); err != nil {
			if s.logger != nil {
				s.logger.Error("recordFailedLogin: failed to set BlockedUntil", zap.Error(err))
			}
			return
		}
		s.rdb.Del(ctx, key)
		if s.logger != nil {
			s.logger.Warn("Account auto-locked after repeated failed PIN attempts",
				zap.String("phone", phone),
				zap.Int64("attempts", count),
			)
		}
	}
}

func (s *authService) SendOTP(ctx context.Context, phone, purpose, ip string) error {
	// Blocage admin au niveau du numéro (Blocked Phone Phase 1A) — empêche l'envoi
	// d'un OTP à un numéro bloqué, quel que soit purpose (register/reset_password).
	if blocked, err := s.userRepo.IsPhoneBlocked(ctx, phone); err == nil && blocked {
		return apperr.ErrPhoneUnavailable
	}

	code := GenerateOTP(s.otpLength)

	// Vérifier s'il existe déjà un OTP en cours de validité
	existingOTP, _ := s.userRepo.FindLatestOTP(ctx, phone, purpose)
	if existingOTP != nil && time.Now().Before(existingOTP.ExpiresAt) {
		// Un OTP valide existe déjà, retourner un message clair
		return apperr.ErrOTPRateLimited
	}

	// Save new OTP to database for verification
	otp := &models.OTPVerification{
		ID:          uuid.New(),
		Phone:       phone,
		Code:   code,
		Purpose:     purpose,
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(s.otpTTL),
		IPAddress:   &ip,
	}

	if err := s.userRepo.CreateOTP(ctx, otp); err != nil {
		return err
	}

	// Send OTP via WhatsApp (Wablas)
	if s.smsService != nil {
		if err := s.smsService.SendOTP(phone, code); err != nil {
			if s.env == "development" {
				fmt.Printf("DEBUG: WhatsApp OTP sending failed in development (%v). Continuing because development OTP is allowed.\n", err)
				return nil
			}
			return err
		}
	}

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, phone, code, purpose string) error {
	otp, err := s.userRepo.FindLatestOTP(ctx, phone, purpose)
	if err != nil {
		return apperr.ErrOTPInvalid
	}

	if time.Now().After(otp.ExpiresAt) {
		return apperr.ErrOTPExpired
	}

	if otp.Attempts >= otp.MaxAttempts {
		return apperr.ErrOTPMaxAttempts
	}

	// Vérifier le code OTP
	// En développement, accepter le code configuré
	// En production, rejeter les codes de développement
	if s.env == "development" && s.developmentOTP != "" && code == s.developmentOTP {
		// OK - Mode développement avec code autorisé
	} else if code != otp.Code {
		// En production ou si code incorrect, incrémenter les tentatives
		if err := s.userRepo.IncrementOTPAttempts(ctx, otp.ID); err != nil {
			return err
		}
		return apperr.ErrOTPInvalid
	}

	if err := s.userRepo.MarkOTPVerified(ctx, otp.ID); err != nil {
		return err
	}

	if purpose == "register" {
		if err := s.userRepo.SetVerified(ctx, phone); err != nil {
			return err
		}
	}

	return nil
}

// ResetPassword exige une preuve OTP valide ("reset_password") avant toute
// modification du PIN — la vérification vit désormais ici, dans le service, et non
// plus seulement dans le handler appelant (OTP Security Phase 1C). Auparavant, le
// handler (auth_handler.go) appelait VerifyOTP puis ResetPassword séparément : cela
// fonctionnait tant que ResetPassword n'avait qu'un seul appelant discipliné, mais
// rien n'empêchait un futur appelant (nouvel endpoint, script, tâche interne) de
// contourner la vérification en appelant directement le service. Le contrat est
// maintenant imposé par la signature elle-même.
func (s *authService) ResetPassword(ctx context.Context, phone, code, newPin string) error {
	if err := s.VerifyOTP(ctx, phone, code, "reset_password"); err != nil {
		return err
	}

	// Résolution du compte par correspondance exacte de 'phone' tel qu'envoyé par le
	// client — sans indice de région ici (contrairement à Register/Login), NormalizeE164
	// ne peut plus être tentée en aveugle depuis la Phase 2 (elle exige désormais une
	// région explicite). C'est sans impact : ce numéro doit déjà correspondre exactement
	// à otp_verifications.phone (VerifyOTP juste au-dessus utilise la même valeur brute),
	// donc la correspondance exacte historique reste correcte ici.
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return apperr.ErrNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePin(ctx, user.ID, string(hash)); err != nil {
		return err
	}

	// Un PIN réinitialisé implique souvent une prise de contrôle de compte suspectée —
	// invalider toute session déjà émise plutôt que de laisser un éventuel attaquant
	// garder l'accès via un JWT encore valide (Session Security Phase 1).
	RevokeUserSessions(ctx, s.rdb, s.logger, user.ID, s.jwtExpiry)

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPin, newPin string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPin)); err != nil {
		return apperr.ErrInvalidPin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePin(ctx, userID, string(hash))
}

func (s *authService) GenerateJWT(userID uuid.UUID, role string, isSuperAdmin bool) (string, error) {
	claims := JWTClaims{
		UserID:       userID.String(),
		Role:         role,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// TrackPasswordReset enregistre une tentative de réinitialisation de mot de passe
// pour détecter les abus (brute force attempts)
func (s *authService) TrackPasswordReset(ctx context.Context, phone, ip string) error {
	return s.userRepo.LogPasswordResetAttempt(ctx, phone, ip)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
