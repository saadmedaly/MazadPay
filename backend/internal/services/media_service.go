package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/config"
	"go.uber.org/zap"
)

// allowedUploadExtensions et allowedUploadMimeTypes forment une liste blanche centrale
// appliquée à TOUT upload passant par MediaService, quel que soit le handler appelant
// (audit de sécurité — certains handlers comme UploadReceipt ou UploadBannerRequestImage
// n'avaient historiquement aucune validation propre). Ceci est un dernier filet de
// sécurité ; les handlers peuvent en plus appliquer des règles plus strictes.
var allowedUploadExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".pdf": true, ".mp4": true,
}

// allowedUploadMimeTypes associe chaque extension autorisée aux magic bytes / types MIME
// réels qu'elle doit avoir (détectés via http.DetectContentType sur le contenu, pas le
// header Content-Type envoyé par le client qui est falsifiable).
var allowedUploadMimeTypes = map[string]map[string]bool{
	".jpg":  {"image/jpeg": true},
	".jpeg": {"image/jpeg": true},
	".png":  {"image/png": true},
	".webp": {"image/webp": true},
	".pdf":  {"application/pdf": true},
	".mp4":  {"video/mp4": true, "application/octet-stream": true}, // certains encodeurs mp4 sont détectés comme octet-stream
}

// maxUploadSizeBytes est un plafond de dernier recours (dernier filet de sécurité) pour
// les images/PDF passant par MediaService, indépendamment de la limite propre à chaque
// handler. Les vidéos (.mp4) ont une limite plus large car elles passent par des chemins
// (tutoriels admin) déjà limités à 100MB au niveau du handler appelant.
const maxUploadSizeBytes = 20 * 1024 * 1024
const maxVideoUploadSizeBytes = 100 * 1024 * 1024

type MediaService interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error)
	UploadAuctionImages(ctx context.Context, files []multipart.File, headers []*multipart.FileHeader, auctionID uuid.UUID) ([]string, error)
	DeleteFile(ctx context.Context, key string) error
	GetPublicURL(key string) string
	// ExtractKey extrait la clé objet (folder/filename) à partir d'une URL publique
	// complète — utilisé pour les reçus dont on ne veut plus stocker/exposer l'URL
	// publique directement (audit de sécurité).
	ExtractKey(url string) string
	// GetPresignedURL génère une URL R2 temporaire (valable `expiry`, max recommandé
	// 5 minutes) pour accéder à un objet du bucket privé des reçus (audit de sécurité —
	// Private Receipts Bucket Separation).
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	// GetReceiptURL génère une URL présignée pour un reçu de paiement, quel que soit son
	// format de stockage : un object key (nouveaux reçus → bucket privé R2_BUCKET_RECEIPTS)
	// ou une ancienne URL publique complète (reçus créés avant la séparation des buckets
	// → bucket média public R2_BUCKET_MEDIA, en lecture seule, temporaire jusqu'à la
	// migration effective des anciens reçus). Ne renvoie jamais l'URL publique brute.
	GetReceiptURL(ctx context.Context, storedValue string, expiry time.Duration) (string, error)
	// UploadPrivateFile upload un fichier vers le bucket R2 privé des reçus de paiement
	// (R2_BUCKET_RECEIPTS) et retourne uniquement l'object key (ex: "receipts/<uuid>.jpg"),
	// jamais une URL publique — le bucket n'a pas d'accès public configuré (audit de
	// sécurité — Private Receipts Bucket Separation).
	UploadPrivateFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error)
}

type mediaService struct {
	client         *s3.Client
	presignClient  *s3.PresignClient
	bucket         string
	bucketReceipts string
	publicURL      string
	appPort        string
	appEnv         string
	useLocal       bool // true when R2 is not configured
	logger         *zap.Logger
}

// isR2Configured returns true only when real R2 credentials are present
func isR2Configured(cfg *config.Config) bool {
	return cfg.R2.AccessKey != "" &&
		cfg.R2.AccessKey != "your_r2_access_key" &&
		cfg.R2.SecretKey != "" &&
		cfg.R2.SecretKey != "your_r2_secret_key" &&
		!strings.Contains(cfg.R2.Endpoint, "xxxxxxxx")
}

func NewMediaService(cfg *config.Config, logger *zap.Logger) MediaService {
	if !isR2Configured(cfg) {
		if cfg.App.Env == "production" {
			logger.Error("[MediaService] R2 not configured in production — uploads will fail. Set R2_ENDPOINT, R2_ACCESS_KEY, R2_SECRET_KEY, R2_BUCKET_MEDIA, R2_PUBLIC_URL.")
		} else {
			logger.Warn("[MediaService] R2 not configured — using local storage under ./uploads/")
		}
		return &mediaService{
			useLocal:  true,
			appPort:   cfg.App.Port,
			appEnv:    cfg.App.Env,
			publicURL: "",
			logger:    logger,
		}
	}

	// Configure AWS SDK for R2 (S3-compatible)
	awsCfg := aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.R2.AccessKey,
			cfg.R2.SecretKey,
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			endpoint := cfg.R2.Endpoint
			if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
				endpoint = "https://" + endpoint
			}
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
			}, nil
		}),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	presignClient := s3.NewPresignClient(client)

	if cfg.R2.BucketReceipts == "" {
		if cfg.App.Env == "production" {
			logger.Error("[MediaService] R2_BUCKET_RECEIPTS not set in production — receipt uploads will be rejected (no fallback to the public media bucket).")
		} else {
			logger.Warn("[MediaService] R2_BUCKET_RECEIPTS not set — falling back to the public media bucket for receipts (development only).")
		}
	}

	return &mediaService{
		client:         client,
		presignClient:  presignClient,
		bucket:         cfg.R2.BucketMedia,
		bucketReceipts: cfg.R2.BucketReceipts,
		publicURL:      cfg.R2.PublicURL,
		appPort:        cfg.App.Port,
		appEnv:         cfg.App.Env,
		useLocal:       false,
		logger:         logger,
	}
}

// validatedUpload contient le résultat de la validation d'un fichier uploadé (liste
// blanche d'extension + magic bytes réels), prêt à être écrit sur disque ou R2.
type validatedUpload struct {
	filename    string
	content     []byte
	contentType string
}

// validateUpload applique la liste blanche centrale d'extensions et vérifie les vrais
// magic bytes du contenu (et non le Content-Type déclaré par le client, falsifiable) —
// dernier filet de sécurité partagé par UploadFile et UploadPrivateFile.
func (s *mediaService) validateUpload(file multipart.File, header *multipart.FileHeader) (*validatedUpload, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}

	if !allowedUploadExtensions[ext] {
		return nil, fmt.Errorf("file type not allowed")
	}

	maxSize := int64(maxUploadSizeBytes)
	if ext == ".mp4" {
		maxSize = maxVideoUploadSizeBytes
	}
	if header.Size > maxSize {
		return nil, fmt.Errorf("file too large (max %d bytes)", maxSize)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if int64(len(content)) > maxSize {
		return nil, fmt.Errorf("file too large (max %d bytes)", maxSize)
	}

	detectHeader := content
	if len(detectHeader) > 512 {
		detectHeader = detectHeader[:512]
	}
	detectedType := http.DetectContentType(detectHeader)
	if idx := strings.Index(detectedType, ";"); idx != -1 {
		detectedType = detectedType[:idx]
	}
	if !allowedUploadMimeTypes[ext][detectedType] {
		s.logger.Warn("[Upload] Rejected: content does not match declared extension",
			zap.String("filename", header.Filename),
			zap.String("extension", ext),
			zap.String("detected_type", detectedType))
		return nil, fmt.Errorf("file content does not match a valid %s file", ext)
	}

	return &validatedUpload{
		filename:    uuid.New().String() + ext,
		content:     content,
		contentType: detectedType,
	}, nil
}

func (s *mediaService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	defer file.Close()

	v, err := s.validateUpload(file, header)
	if err != nil {
		return "", err
	}

	// --- Local storage fallback ---
	if s.useLocal {
		if s.appEnv == "production" {
			return "", fmt.Errorf("file upload unavailable in production: R2 credentials not configured")
		}
		localDir := filepath.Join("uploads", folder)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create upload dir: %w", err)
		}
		localPath := filepath.Join(localDir, v.filename)
		dst, err := os.Create(localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}
		defer dst.Close()
		if _, err := dst.Write(v.content); err != nil {
			return "", fmt.Errorf("failed to save file: %w", err)
		}
		port := s.appPort
		if port == "" {
			port = "8082"
		}
		publicURL := fmt.Sprintf("http://localhost:%s/uploads/%s/%s", port, folder, v.filename)
		s.logger.Info("[LocalStorage] File saved", zap.String("path", localPath), zap.String("url", publicURL))
		return publicURL, nil
	}

	key := fmt.Sprintf("%s/%s", folder, v.filename)
	if err := s.putObject(ctx, s.bucket, key, v); err != nil {
		return "", err
	}

	publicURL := s.GetPublicURL(key)
	s.logger.Info("[R2 Upload] File uploaded successfully",
		zap.String("key", key), zap.String("url", publicURL), zap.String("bucket", s.bucket))
	return publicURL, nil
}

// UploadPrivateFile upload vers le bucket R2 privé des reçus (jamais le bucket média
// public) et retourne uniquement l'object key, jamais une URL — audit de sécurité,
// Private Receipts Bucket Separation.
func (s *mediaService) UploadPrivateFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	defer file.Close()

	if s.useLocal {
		return "", fmt.Errorf("private uploads unavailable: R2 not configured")
	}

	bucket := s.bucketReceipts
	if bucket == "" {
		if s.appEnv == "production" {
			// Aucun fallback silencieux vers le bucket public en production — un reçu de
			// paiement ne doit jamais atterrir dans un bucket accessible publiquement.
			return "", fmt.Errorf("receipt uploads unavailable: R2_BUCKET_RECEIPTS not configured")
		}
		s.logger.Warn("[UploadPrivateFile] R2_BUCKET_RECEIPTS not set, falling back to public media bucket (development only)")
		bucket = s.bucket
	}

	v, err := s.validateUpload(file, header)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s/%s", folder, v.filename)
	if err := s.putObject(ctx, bucket, key, v); err != nil {
		return "", err
	}

	s.logger.Info("[R2 Upload] Private file uploaded successfully",
		zap.String("key", key), zap.String("bucket", bucket))
	return key, nil
}

// putObject écrit le contenu validé dans le bucket R2 spécifié.
func (s *mediaService) putObject(ctx context.Context, bucket, key string, v *validatedUpload) error {
	s.logger.Info("[R2 Upload] Starting upload",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.String("content_type", v.contentType),
		zap.Int("content_length", len(v.content)))

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(v.content),
		ContentType: aws.String(v.contentType),
	})
	if err != nil {
		s.logger.Error("[R2 Upload] Cloudflare R2 upload failed",
			zap.Error(err), zap.String("bucket", bucket), zap.String("key", key))
		return fmt.Errorf("failed to upload file to Cloudflare R2: %w", err)
	}
	return nil
}

// getEndpoint returns the R2 endpoint for logging (without credentials)
func (s *mediaService) getEndpoint() string {
	return s.publicURL
}

func (s *mediaService) UploadAuctionImages(ctx context.Context, files []multipart.File, headers []*multipart.FileHeader, auctionID uuid.UUID) ([]string, error) {
	s.logger.Info("[R2 UploadAuctionImages] Starting batch upload",
		zap.String("auction_id", auctionID.String()),
		zap.Int("total_files", len(files)))

	var urls []string
	var uploadErrors []error

	for i, file := range files {
		if file == nil || headers[i] == nil {
			s.logger.Warn("[R2 UploadAuctionImages] Skipping nil file",
				zap.Int("index", i))
			continue
		}

		// Folder structure: mazad-mwdia/auctions/{auctionID}
		folder := fmt.Sprintf("auctions/%s", auctionID.String())
		s.logger.Info("[R2 UploadAuctionImages] Uploading image",
			zap.Int("index", i+1),
			zap.Int("total", len(files)),
			zap.String("filename", headers[i].Filename),
			zap.String("folder", folder))

		url, err := s.UploadFile(ctx, file, headers[i], folder)
		if err != nil {
			s.logger.Error("[R2 UploadAuctionImages] Failed to upload image",
				zap.Error(err),
				zap.Int("index", i+1),
				zap.String("auction_id", auctionID.String()),
				zap.String("filename", headers[i].Filename))
			uploadErrors = append(uploadErrors, err)
			// Continue with other files instead of failing completely
			continue
		}
		urls = append(urls, url)
		s.logger.Info("[R2 UploadAuctionImages] Image uploaded successfully",
			zap.Int("index", i+1),
			zap.String("url", url))
	}

	if len(urls) == 0 {
		s.logger.Error("[R2 UploadAuctionImages] No images were successfully uploaded",
			zap.String("auction_id", auctionID.String()),
			zap.Errors("errors", uploadErrors))
		return nil, fmt.Errorf("no images were successfully uploaded to Cloudflare R2: %v", uploadErrors)
	}

	s.logger.Info("[R2 UploadAuctionImages] Batch upload completed",
		zap.String("auction_id", auctionID.String()),
		zap.Int("success_count", len(urls)),
		zap.Int("error_count", len(uploadErrors)))

	return urls, nil
}

func (s *mediaService) DeleteFile(ctx context.Context, key string) error {
	// Extract key from full URL if needed
	if strings.HasPrefix(key, "http") {
		key = s.extractKeyFromURL(key)
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		s.logger.Error("Failed to delete file from R2", zap.Error(err), zap.String("key", key))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Info("File deleted successfully", zap.String("key", key))
	return nil
}

func (s *mediaService) GetPublicURL(key string) string {
	if s.publicURL == "" {
		return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", s.bucket, s.bucket, key)
	}
	return fmt.Sprintf("%s/%s", s.publicURL, key)
}

func (s *mediaService) extractKeyFromURL(url string) string {
	if s.publicURL != "" && strings.HasPrefix(url, s.publicURL) {
		return strings.TrimPrefix(url, s.publicURL+"/")
	}
	// Fallback: try to extract from URL path (assuming https://domain/bucket/key)
	parts := strings.Split(url, "/")
	if len(parts) >= 4 {
		return strings.Join(parts[3:], "/")
	}
	return url
}

// ExtractKey retourne l'object key à partir d'une valeur stockée dans receipt_url.
// Compatible avec les deux formats possibles : les nouveaux reçus stockent directement
// l'object key (ex: "receipts/<uuid>.jpg", voir UploadPrivateFile), tandis que les
// anciens reçus (avant Private Receipts Bucket Separation) stockent encore l'URL
// publique complète de l'ancien bucket média — dans ce cas on l'extrait comme avant.
func (s *mediaService) ExtractKey(value string) string {
	if !strings.HasPrefix(value, "http") {
		return value // déjà un object key
	}
	return s.extractKeyFromURL(value)
}

// presignInBucket est l'implémentation partagée de génération d'URL présignée pour un
// bucket donné.
func (s *mediaService) presignInBucket(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return req.URL, nil
}

// GetPresignedURL génère une URL R2 temporaire valable `expiry` pour un objet du bucket
// privé des reçus (R2_BUCKET_RECEIPTS). En développement, si ce bucket n'est pas
// configuré, retombe sur le bucket média (pour ne pas casser les tests locaux) — jamais
// en production, où ce fallback créerait une confusion sur ce qui est réellement privé.
func (s *mediaService) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if s.useLocal {
		return "", fmt.Errorf("presigned URLs unavailable: R2 not configured, local storage in use")
	}
	bucket := s.bucketReceipts
	if bucket == "" {
		if s.appEnv == "production" {
			return "", fmt.Errorf("receipt access unavailable: R2_BUCKET_RECEIPTS not configured")
		}
		bucket = s.bucket
	}
	return s.presignInBucket(ctx, bucket, key, expiry)
}

// GetReceiptURL choisit automatiquement le bon bucket selon le format de storedValue :
//   - une URL publique complète (legacy, ancien reçu créé avant la séparation des
//     buckets) → on extrait la clé et on présigne depuis le bucket média public
//     (R2_BUCKET_MEDIA), où le fichier existe réellement encore. Solution temporaire
//     jusqu'à la migration effective des anciens reçus.
//   - un object key simple (nouveau reçu, ex: "receipts/<uuid>.jpg") → on présigne
//     depuis le bucket privé des reçus (R2_BUCKET_RECEIPTS).
//
// Ne renvoie jamais l'URL publique brute au client, dans les deux cas.
func (s *mediaService) GetReceiptURL(ctx context.Context, storedValue string, expiry time.Duration) (string, error) {
	if s.useLocal {
		return "", fmt.Errorf("presigned URLs unavailable: R2 not configured, local storage in use")
	}

	if strings.HasPrefix(storedValue, "http") {
		// Legacy : ancien reçu avec URL publique complète stockée avant la séparation
		// des buckets. Le fichier existe encore dans l'ancien bucket média public.
		key := s.extractKeyFromURL(storedValue)
		return s.presignInBucket(ctx, s.bucket, key, expiry)
	}

	// Nouveau reçu : storedValue est déjà un object key, dans le bucket privé.
	bucket := s.bucketReceipts
	if bucket == "" {
		if s.appEnv == "production" {
			return "", fmt.Errorf("receipt access unavailable: R2_BUCKET_RECEIPTS not configured")
		}
		bucket = s.bucket
	}
	return s.presignInBucket(ctx, bucket, storedValue, expiry)
}
