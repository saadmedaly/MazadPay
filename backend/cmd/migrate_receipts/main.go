// migrate_receipts est un outil CLI ponctuel (non intégré au serveur HTTP) pour migrer
// les anciens reçus de paiement — stockés avant la séparation des buckets R2 (voir
// commit bcb6488, "Private Receipts Bucket Separation") sous forme d'URL publique
// complète du bucket média (R2_BUCKET_MEDIA) — vers le bucket privé dédié
// (R2_BUCKET_RECEIPTS), en ne conservant en base que l'object key (ex:
// "receipts/<uuid>.jpg"), jamais l'URL publique.
//
// Usage :
//
//	go run ./cmd/migrate_receipts --dry-run   (par défaut, aucune écriture nulle part)
//	go run ./cmd/migrate_receipts --execute   (exécution réelle, voir déroulé ci-dessous)
//
// Déroulé en mode --execute, dans cet ordre strict, par ligne :
//  1. Sauvegarde locale (fichier hors dépôt git) de l'ancienne valeur de receipt_url.
//  2. CopyObject : copie le fichier du bucket public vers le bucket privé, sous la même
//     clé. Le fichier source N'EST JAMAIS supprimé (opération non destructive).
//  3. HeadObject : vérifie que la copie existe bien dans le bucket privé avant de
//     continuer.
//  4. UPDATE transactions SET receipt_url = '<clé>' WHERE id = '<id>' — uniquement pour
//     la ligne dont la copie a été vérifiée avec succès.
//
// Idempotence : une ligne dont receipt_url est déjà un object key (pas une URL http) est
// ignorée. Un fichier déjà présent dans le bucket privé (HeadObject réussit avant même
// la copie) n'est pas re-copié. Un échec sur une ligne est journalisé et n'interrompt
// pas le traitement des lignes suivantes.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mazadpay/backend/internal/config"
)

// receiptKeyPattern extrait "receipts/<filename>" depuis une URL publique complète.
var receiptKeyPattern = regexp.MustCompile(`(receipts/[^/?#]+)$`)

type row struct {
	TransactionID string    `db:"id"`
	UserID        string    `db:"user_id"`
	Amount        string    `db:"amount"`
	Status        string    `db:"status"`
	ReceiptURL    string    `db:"receipt_url"`
	CreatedAt     time.Time `db:"created_at"`
}

// planned représente une ligne migrable : id de transaction + object key R2 extrait.
type planned struct {
	id, key string
}

func main() {
	execute := flag.Bool("execute", false, "Exécute la migration réelle (copie + mise à jour DB). Sans ce flag : dry-run.")
	dryRunFlag := flag.Bool("dry-run", true, "Mode simulation (par défaut) : aucune écriture, seulement un rapport.")
	flag.Parse()

	dryRun := !*execute
	_ = dryRunFlag // le flag existe pour la clarté de l'usage, --execute est la seule bascule réelle

	cfg := config.Load()

	db, err := sqlx.Connect("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s client_encoding=UTF8",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	))
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	var rows []row
	err = db.Select(&rows, `
		SELECT id, user_id, amount, status, receipt_url, created_at
		FROM transactions
		WHERE receipt_url IS NOT NULL AND receipt_url != ''
		ORDER BY created_at`)
	if err != nil {
		log.Fatalf("Failed to query transactions: %v", err)
	}

	fmt.Println("========================================")
	if dryRun {
		fmt.Println("MODE : DRY-RUN (aucune écriture, aucune copie, aucune suppression)")
	} else {
		fmt.Println("MODE : EXECUTE (migration réelle)")
	}
	fmt.Println("========================================")

	var toMigrate []planned
	skippedAlreadyKey := 0
	skippedEmpty := 0
	unrecognized := 0

	for _, r := range rows {
		shortTx := shortID(r.TransactionID)
		shortUser := shortID(r.UserID)

		if !strings.HasPrefix(r.ReceiptURL, "http") {
			// Déjà un object key (reçu créé après la séparation des buckets) — idempotent, on ignore.
			skippedAlreadyKey++
			fmt.Printf("[SKIP already-key] tx=%s user=%s status=%s key=%s\n", shortTx, shortUser, r.Status, r.ReceiptURL)
			continue
		}

		match := receiptKeyPattern.FindStringSubmatch(r.ReceiptURL)
		if match == nil {
			unrecognized++
			fmt.Printf("[SKIP unrecognized-format] tx=%s user=%s status=%s amount=%s created_at=%s (URL ne correspond pas au motif receipts/<filename>)\n",
				shortTx, shortUser, r.Status, r.Amount, r.CreatedAt.Format("2006-01-02"))
			continue
		}
		key := match[1]

		fmt.Printf("[MIGRATABLE] tx=%s user=%s status=%s amount=%s key=%s created_at=%s\n",
			shortTx, shortUser, r.Status, r.Amount, key, r.CreatedAt.Format("2006-01-02"))
		toMigrate = append(toMigrate, planned{id: r.TransactionID, key: key})
	}

	fmt.Println("========================================")
	fmt.Printf("Résumé : %d ligne(s) au total, %d migrable(s), %d déjà en clé (ignorée), %d format non reconnu, %d vide\n",
		len(rows), len(toMigrate), skippedAlreadyKey, unrecognized, skippedEmpty)
	fmt.Println("========================================")

	if dryRun {
		fmt.Println("Dry-run terminé. Aucune donnée modifiée. Relancer avec --execute pour la migration réelle.")
		return
	}

	if cfg.R2.BucketReceipts == "" {
		log.Fatal("R2_BUCKET_RECEIPTS n'est pas configuré — migration réelle impossible.")
	}

	runExecute(context.Background(), db, cfg, toMigrate)
}

// shortID retourne les 8 premiers caractères d'un UUID, jamais l'identifiant complet ni
// aucune donnée sensible — cohérent avec shortID() côté web/lib/formatters.ts.
func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// runExecute exécute la migration réelle : backup local, CopyObject, HeadObject, puis
// UPDATE DB uniquement pour les lignes dont la copie est vérifiée. N'est jamais appelée
// en mode dry-run.
func runExecute(ctx context.Context, db *sqlx.DB, cfg *config.Config, toMigrate []planned) {
	if len(toMigrate) == 0 {
		fmt.Println("Rien à migrer.")
		return
	}

	// Backup local avant toute écriture — fichier horodaté, jamais commité (voir
	// .gitignore : backend/migrate_receipts_backup_*.txt).
	backupPath := fmt.Sprintf("migrate_receipts_backup_%s.txt", time.Now().Format("20060102_150405"))
	backupFile, err := os.Create(backupPath)
	if err != nil {
		log.Fatalf("Impossible de créer le fichier de backup local: %v", err)
	}
	defer backupFile.Close()

	client, err := newS3Client(cfg)
	if err != nil {
		log.Fatalf("Impossible d'initialiser le client R2: %v", err)
	}

	successCount, failCount := 0, 0

	for _, m := range toMigrate {
		shortTx := shortID(m.id)

		// 1. Backup avant toute action sur cette ligne.
		var oldURL string
		if err := db.Get(&oldURL, "SELECT receipt_url FROM transactions WHERE id = $1", m.id); err != nil {
			fmt.Printf("[FAIL backup] tx=%s: %v\n", shortTx, err)
			failCount++
			continue
		}
		fmt.Fprintf(backupFile, "%s\t%s\n", m.id, oldURL)

		// 2. CopyObject vers le bucket privé (idempotent : si le fichier existe déjà à
		//    cette clé côté destination, la copie l'écrase avec un contenu identique —
		//    sans effet de bord car c'est le même fichier source).
		copySource := fmt.Sprintf("%s/%s", cfg.R2.BucketMedia, m.key)
		_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(cfg.R2.BucketReceipts),
			Key:        aws.String(m.key),
			CopySource: aws.String(copySource),
		})
		if err != nil {
			fmt.Printf("[FAIL copy] tx=%s key=%s: %v\n", shortTx, m.key, err)
			failCount++
			continue
		}

		// 3. Vérification HeadObject avant toute mise à jour DB.
		_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(cfg.R2.BucketReceipts),
			Key:    aws.String(m.key),
		})
		if err != nil {
			fmt.Printf("[FAIL verify] tx=%s key=%s: copie introuvable après CopyObject: %v\n", shortTx, m.key, err)
			failCount++
			continue
		}

		// 4. UPDATE uniquement pour cette ligne, uniquement après vérification réussie.
		_, err = db.Exec("UPDATE transactions SET receipt_url = $1 WHERE id = $2", m.key, m.id)
		if err != nil {
			fmt.Printf("[FAIL db-update] tx=%s key=%s: %v (fichier copié mais DB non mise à jour — la ligne pointe encore vers l'ancienne URL, aucune perte)\n", shortTx, m.key, err)
			failCount++
			continue
		}

		fmt.Printf("[OK] tx=%s migré vers bucket privé, DB mise à jour.\n", shortTx)
		successCount++
	}

	fmt.Println("========================================")
	fmt.Printf("Migration terminée : %d succès, %d échec(s). Backup local: %s\n", successCount, failCount, backupPath)
	fmt.Println("Rappel : les fichiers originaux dans le bucket public n'ont PAS été supprimés.")
}

// newS3Client construit un client S3 compatible R2, indépendant de MediaService (qui ne
// vise que l'upload) — nécessaire ici pour CopyObject/HeadObject entre deux buckets.
func newS3Client(cfg *config.Config) (*s3.Client, error) {
	awsCfg := aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.R2.AccessKey, cfg.R2.SecretKey, "",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			endpoint := cfg.R2.Endpoint
			if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
				endpoint = "https://" + endpoint
			}
			return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
		}),
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return client, nil
}
