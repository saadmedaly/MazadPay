package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/config"
	"go.uber.org/zap"
)

type MediaService interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error)
	UploadAuctionImages(ctx context.Context, files []multipart.File, headers []*multipart.FileHeader, auctionID uuid.UUID) ([]string, error)
	DeleteFile(ctx context.Context, key string) error
	GetPublicURL(key string) string
}

type mediaService struct {
	client    *s3.Client
	bucket    string
	publicURL string
	appPort   string
	appEnv    string
	useLocal  bool // true when R2 is not configured
	logger    *zap.Logger
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

	return &mediaService{
		client:    client,
		bucket:    cfg.R2.BucketMedia,
		publicURL: cfg.R2.PublicURL,
		appPort:   cfg.App.Port,
		appEnv:    cfg.App.Env,
		useLocal:  false,
		logger:    logger,
	}
}

func (s *mediaService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := uuid.New().String() + ext

	// --- Local storage fallback ---
	if s.useLocal {
		if s.appEnv == "production" {
			return "", fmt.Errorf("file upload unavailable in production: R2 credentials not configured")
		}
		localDir := filepath.Join("uploads", folder)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create upload dir: %w", err)
		}
		localPath := filepath.Join(localDir, filename)
		dst, err := os.Create(localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			return "", fmt.Errorf("failed to save file: %w", err)
		}
		port := s.appPort
		if port == "" {
			port = "8082"
		}
		publicURL := fmt.Sprintf("http://localhost:%s/uploads/%s/%s", port, folder, filename)
		s.logger.Info("[LocalStorage] File saved", zap.String("path", localPath), zap.String("url", publicURL))
		return publicURL, nil
	}

	// --- R2 upload ---
	key := fmt.Sprintf("%s/%s", folder, filename)

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("[R2 Upload] Failed to read file content",
			zap.Error(err),
			zap.String("filename", header.Filename),
			zap.Int64("size", header.Size))
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	s.logger.Info("[R2 Upload] Starting upload",
		zap.String("bucket", s.bucket),
		zap.String("key", key),
		zap.String("content_type", contentType),
		zap.Int("content_length", len(content)))

	// Upload to R2
	output, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		// Detailed Cloudflare R2 error logging
		s.logger.Error("[R2 Upload] Cloudflare R2 upload failed",
			zap.Error(err),
			zap.String("bucket", s.bucket),
			zap.String("key", key),
			zap.String("endpoint", s.getEndpoint()),
			zap.String("error_type", "R2_UPLOAD_ERROR"),
			zap.String("filename", header.Filename),
			zap.Int64("file_size", header.Size))
		return "", fmt.Errorf("failed to upload file to Cloudflare R2: %w", err)
	}

	publicURL := s.GetPublicURL(key)
	s.logger.Info("[R2 Upload] File uploaded successfully",
		zap.String("key", key),
		zap.String("url", publicURL),
		zap.String("etag", aws.ToString(output.ETag)),
		zap.String("bucket", s.bucket))
	return publicURL, nil
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
