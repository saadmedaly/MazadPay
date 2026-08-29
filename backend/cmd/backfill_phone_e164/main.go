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
// Confidentialité (PII) : AUCUN numéro de téléphone complet (ni la colonne 'phone'
// héritée, ni le phone_e164 calculé) n'est jamais imprimé en clair, sur AUCUN chemin —
// dry-run, execute, ou erreur. Voir maskPhone() dans plan.go, appliqué systématiquement.
//
// Conflits : avant tout UPDATE, l'outil détecte deux catégories de collision sur
// phone_e164 — (a) deux candidats du même lot se normalisant vers la même valeur, et
// (b) un candidat se normalisant vers une valeur déjà présente sur un AUTRE utilisateur
// déjà backfillé. TOUTES les lignes impliquées dans une collision (sans exception, sans
// ordre de priorité par created_at/role/statut admin) sont marquées "conflict" et
// EXCLUES de la mise à jour, en dry-run comme en execute. La résolution d'un conflit est
// une décision opérationnelle séparée, hors du périmètre de cet outil automatique — voir
// buildBackfillPlan() dans plan.go.
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
)

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

	var rows []struct {
		ID    string `db:"id"`
		Phone string `db:"phone"`
	}
	err = db.Select(&rows, `SELECT id, phone FROM users WHERE phone_e164 IS NULL ORDER BY created_at`)
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}

	// Load already-normalized rows (phone_e164 IS NOT NULL) so a candidate colliding
	// with an EXISTING row is also caught as a conflict, not just collisions within
	// this batch.
	var existingRows []struct {
		ID   string `db:"id"`
		E164 string `db:"phone_e164"`
	}
	err = db.Select(&existingRows, `SELECT id, phone_e164 FROM users WHERE phone_e164 IS NOT NULL`)
	if err != nil {
		log.Fatalf("Failed to query existing phone_e164 values: %v", err)
	}
	existing := make(existingE164Lookup, len(existingRows))
	for _, r := range existingRows {
		existing[r.ID] = r.E164
	}

	candidates := make([]candidate, 0, len(rows))
	for _, r := range rows {
		candidates = append(candidates, candidate{ID: r.ID, Phone: r.Phone})
	}

	fmt.Println("========================================")
	if dryRun {
		fmt.Println("MODE : DRY-RUN (aucune écriture)")
	} else {
		fmt.Println("MODE : EXECUTE (backfill réel)")
	}
	fmt.Printf("%d utilisateur(s) avec phone_e164 IS NULL\n", len(rows))
	fmt.Println("========================================")

	plan := buildBackfillPlan(candidates, existing)

	// phoneByID lets the per-entry log lines mask the ORIGINAL raw phone (not just
	// the computed E.164), consistent with the previous log format but never
	// unmasked.
	phoneByID := make(map[string]string, len(rows))
	for _, r := range rows {
		phoneByID[r.ID] = r.Phone
	}

	for _, p := range plan {
		id := shortID(p.ID)
		rawMasked := maskPhone(phoneByID[p.ID])
		switch p.Status {
		case "skip_unparseable":
			fmt.Printf("[SKIP unparseable] user=%s phone=%s error=%v\n", id, rawMasked, p.Err)
		case "skip_unexpected_region":
			fmt.Printf("[SKIP unexpected-region] user=%s phone=%s detected_region=%s (attendu MR)\n", id, rawMasked, p.ISO)
		case "conflict":
			fmt.Printf("[CONFLICT] user=%s phone=%s -> phone_e164=%s (colluding with another row, needs manual resolution)\n", id, rawMasked, maskPhone(p.E164))
		case "migratable":
			fmt.Printf("[MIGRATABLE] user=%s phone=%s -> phone_e164=%s phone_country_iso=%s\n", id, rawMasked, maskPhone(p.E164), p.ISO)
		}
	}

	s := summarize(plan)
	fmt.Println("========================================")
	fmt.Printf("Résumé : scanned=%d migratable=%d skipped_invalid=%d conflicts=%d\n",
		s.Scanned, s.Migratable, s.SkippedInvalid, s.Conflicts)
	fmt.Println("========================================")

	if dryRun {
		fmt.Println("Dry-run terminé. Aucune donnée modifiée. Relancer avec --execute pour le backfill réel.")
		return
	}

	successCount, failCount := 0, 0
	for _, p := range plan {
		if p.Status != "migratable" {
			// Conflicts and skips are NEVER written, in execute mode either — no
			// exception, no priority ordering.
			continue
		}
		_, err := db.Exec(
			`UPDATE users SET phone_e164 = $1, phone_country_iso = $2 WHERE id = $3 AND phone_e164 IS NULL`,
			p.E164, p.ISO, p.ID,
		)
		if err != nil {
			fmt.Printf("[FAIL db-update] user=%s: %v\n", shortID(p.ID), err)
			failCount++
			continue
		}
		successCount++
	}

	fmt.Println("========================================")
	fmt.Printf("Backfill terminé : scanned=%d migratable=%d skipped_invalid=%d conflicts=%d updated=%d failed=%d\n",
		s.Scanned, s.Migratable, s.SkippedInvalid, s.Conflicts, successCount, failCount)
	fmt.Println("La colonne 'phone' héritée n'a pas été modifiée.")
	fmt.Println("Les conflits n'ont PAS été appliqués — résolution manuelle séparée requise.")
}
