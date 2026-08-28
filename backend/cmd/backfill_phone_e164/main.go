// backfill_phone_e164 est un outil CLI ponctuel (non intégré au serveur HTTP) pour
// renseigner users.phone_e164 / users.phone_country_iso (migration 000044) pour les
// comptes créés avant l'introduction du support téléphonique international. Avant
// cette migration, seule la Mauritanie était supportée : on suppose donc que
// users.phone (colonne héritée, jamais modifiée par cet outil) est un numéro
// mauritanien, on le normalise en E.164 via la même logique que le service
// d'authentification (services.NormalizeE164), et on écrit le résultat dans les
// deux nouvelles colonnes.
//
// Usage :
//
//	go run ./cmd/backfill_phone_e164 --dry-run   (par défaut, aucune écriture)
//	go run ./cmd/backfill_phone_e164 --execute   (exécution réelle)
//
// Idempotence : seules les lignes avec phone_e164 IS NULL sont sélectionnées, donc
// relancer l'outil après une exécution partielle ou complète ne touche jamais deux
// fois la même ligne. Une ligne dont le numéro ne peut pas être normalisé (format
// invalide) est journalisée et ignorée — jamais de panique ni d'arrêt du traitement
// des lignes suivantes.
//
// Ne JAMAIS lancer --execute contre une base de données de production dans le cadre
// de cette tâche : cet outil est livré construit et testé en dry-run uniquement.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mazadpay/backend/internal/config"
	"github.com/mazadpay/backend/internal/services"
)

type userRow struct {
	ID    string `db:"id"`
	Phone string `db:"phone"`
}

func main() {
	execute := flag.Bool("execute", false, "Exécute le backfill réel (UPDATE en base). Sans ce flag : dry-run.")
	flag.Parse()
	dryRun := !*execute

	cfg := config.Load()

	db, err := sqlx.Connect("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s client_encoding=UTF8",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	))
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	var rows []userRow
	err = db.Select(&rows, `SELECT id, phone FROM users WHERE phone_e164 IS NULL ORDER BY created_at`)
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}

	fmt.Println("========================================")
	if dryRun {
		fmt.Println("MODE : DRY-RUN (aucune écriture)")
	} else {
		fmt.Println("MODE : EXECUTE (backfill réel)")
	}
	fmt.Printf("%d utilisateur(s) avec phone_e164 IS NULL\n", len(rows))
	fmt.Println("========================================")

	type planned struct {
		id, e164, iso string
	}

	var toUpdate []planned
	skipped := 0

	for _, r := range rows {
		masked := maskPhone(r.Phone)
		shortID := shortID(r.ID)

		e164, iso, err := services.NormalizeE164(r.Phone, "MR")
		if err != nil {
			skipped++
			fmt.Printf("[SKIP unparseable] user=%s phone=%s error=%v\n", shortID, masked, err)
			continue
		}
		if iso != "MR" {
			// Un numéro pré-existant qui se normaliserait vers un autre pays serait
			// suspect (aucun autre pays n'était jamais accepté avant cette
			// migration) — on préfère l'ignorer plutôt que d'écrire une donnée
			// probablement erronée, à examiner manuellement.
			skipped++
			fmt.Printf("[SKIP unexpected-region] user=%s phone=%s detected_region=%s (attendu MR)\n", shortID, masked, iso)
			continue
		}

		fmt.Printf("[MIGRATABLE] user=%s phone=%s -> phone_e164=%s phone_country_iso=%s\n", shortID, masked, e164, iso)
		toUpdate = append(toUpdate, planned{id: r.ID, e164: e164, iso: iso})
	}

	fmt.Println("========================================")
	fmt.Printf("Résumé : %d ligne(s) au total, %d migrable(s), %d ignorée(s)\n", len(rows), len(toUpdate), skipped)
	fmt.Println("========================================")

	if dryRun {
		fmt.Println("Dry-run terminé. Aucune donnée modifiée. Relancer avec --execute pour le backfill réel.")
		return
	}

	successCount, failCount := 0, 0
	for _, p := range toUpdate {
		_, err := db.Exec(
			`UPDATE users SET phone_e164 = $1, phone_country_iso = $2 WHERE id = $3 AND phone_e164 IS NULL`,
			p.e164, p.iso, p.id,
		)
		if err != nil {
			fmt.Printf("[FAIL db-update] user=%s: %v\n", shortID(p.id), err)
			failCount++
			continue
		}
		successCount++
	}

	fmt.Println("========================================")
	fmt.Printf("Backfill terminé : %d succès, %d échec(s).\n", successCount, failCount)
	fmt.Println("La colonne 'phone' héritée n'a pas été modifiée.")
}

// maskPhone masque un numéro pour les logs, cohérent avec models.User.MaskPhone().
func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "####"
	}
	return "####" + phone[len(phone)-4:]
}

// shortID retourne les 8 premiers caractères d'un UUID, jamais l'identifiant complet.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
