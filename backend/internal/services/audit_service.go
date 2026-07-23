package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

// LogOption enrichit un AuditLog avec les colonnes ajoutées par
// 000042_audit_logs_schema_improvement (Phase A). Toutes optionnelles — un appel
// Log(...) sans aucune option se comporte exactement comme avant cette migration
// (actor_type prend simplement la valeur par défaut "unknown" en base/dans Create()).
// Activation progressive prévue en Phase B : ne pas ajouter ces options à tous les
// appels existants dans cette phase.
type LogOption func(*models.AuditLog)

// WithActorType précise explicitement si l'acteur est "admin", "user" ou "system" —
// par défaut ("unknown") si non fourni, cohérent avec le backfill des lignes
// historiques (voir 000042_audit_logs_schema_improvement.up.sql).
func WithActorType(actorType string) LogOption {
	return func(l *models.AuditLog) { l.ActorType = actorType }
}

// WithIP journalise l'adresse IP de la requête à l'origine de l'action.
func WithIP(ip string) LogOption {
	return func(l *models.AuditLog) {
		if ip != "" {
			l.IPAddress = &ip
		}
	}
}

// WithUserAgent journalise le User-Agent de la requête.
func WithUserAgent(ua string) LogOption {
	return func(l *models.AuditLog) {
		if ua != "" {
			l.UserAgent = &ua
		}
	}
}

// WithDetailsJSON attache un détail structuré (JSONB) en plus du champ details
// (TEXT) existant — n'est jamais généré automatiquement à partir de details pour ne
// pas produire de données trompeuses (voir rapport Audit Logs Schema Improvement).
func WithDetailsJSON(v models.JSONB) LogOption {
	return func(l *models.AuditLog) { l.DetailsJSON = v }
}

// WithEntityKey précise l'identifiant d'une entité dont la clé n'est pas un UUID
// (ex: banner.id, category.id — entiers), en complément ou à la place de entity_id.
func WithEntityKey(key string) LogOption {
	return func(l *models.AuditLog) {
		if key != "" {
			l.EntityKey = &key
		}
	}
}

// WithSystemActor efface actor_id (NULL en base) pour une action déclenchée par un
// processus système (ex: le scheduler CloseExpiredAuctions) plutôt que par un
// utilisateur ou un admin réel — à utiliser avec WithActorType("system"). Sans cette
// option, actor_id serait rempli avec la valeur (souvent uuid.Nil) passée en premier
// argument de Log(...), ce qui n'est pas un vrai identifiant d'acteur.
func WithSystemActor() LogOption {
	return func(l *models.AuditLog) { l.ActorID = nil }
}

type AuditService interface {
	Log(ctx context.Context, adminID uuid.UUID, action, entityType string, entityID *uuid.UUID, details string, opts ...LogOption) error
	GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.AuditLog, error)
	List(ctx context.Context, page, perPage int) ([]models.AuditLog, int, error)
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, adminID uuid.UUID, action, entityType string, entityID *uuid.UUID, details string, opts ...LogOption) error {
	log := &models.AuditLog{
		ID:         uuid.New(),
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		ActorID:    &adminID,
	}
	for _, opt := range opts {
		opt(log)
	}
	return s.repo.Create(ctx, log)
}

func (s *auditService) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.AuditLog, error) {
	return s.repo.ListByEntity(ctx, entityType, entityID)
}

func (s *auditService) List(ctx context.Context, page, perPage int) ([]models.AuditLog, int, error) {
	return s.repo.ListPaginated(ctx, page, perPage)
}