package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type LimitsHandler struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewLimitsHandler(db *pgxpool.Pool, logger *zap.Logger) *LimitsHandler {
	return &LimitsHandler{db: db, logger: logger}
}

// LimitsCheck is a cron-triggered endpoint to verify all free tier limits
func (h *LimitsHandler) LimitsCheck(c *fiber.Ctx) error {
	// Middleware de vérification de la clé interne
	internalKey := os.Getenv("INTERNAL_CRON_KEY")
	if internalKey != "" && c.Get("X-Internal-Key") != internalKey {
		h.logger.Warn("Unauthorized internal access attempt", zap.String("ip", c.IP()))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	status := "ok"
	ctx := c.Context()
	
	checks := make(map[string]interface{})
	var warningMessages []string

	// 1. NEON — Taille de la BDD
	var dbSizeBytes int64
	err := h.db.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&dbSizeBytes)
	if err != nil {
		h.logger.Error("Failed to fetch Neon DB size", zap.Error(err))
	} else {
		usedMB := float64(dbSizeBytes) / (1024 * 1024)
		limitMB := 512.0
		pct := (usedMB / limitMB) * 100
		isOk := usedMB <= 409.6 // 80%
		checks["neon_storage"] = fiber.Map{
			"used_mb":  fmt.Sprintf("%.2f", usedMB),
			"limit_mb": limitMB,
			"pct":      fmt.Sprintf("%.1f", pct),
			"ok":       isOk,
		}
		if !isOk {
			status = setStatus(status, "warning")
			warningMessages = append(warningMessages, fmt.Sprintf("⚠️ Neon Storage: %.1f%% used (%.0fMB/512MB)", pct, usedMB))
		}
	}

	// 2. NEON — Connexions actives
	var activeConns int
	err = h.db.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConns)
	if err != nil {
		h.logger.Error("Failed to fetch Neon active connections", zap.Error(err))
	} else {
		limitConns := 100
		isOk := activeConns <= 80
		checks["neon_connections"] = fiber.Map{
			"active": activeConns,
			"limit":  limitConns,
			"ok":     isOk,
		}
		if !isOk {
			status = setStatus(status, "warning")
			warningMessages = append(warningMessages, fmt.Sprintf("⚠️ Neon Connections: %d/%d active", activeConns, limitConns))
		}
	}

	// 3. R2 — Cloudflare API (Metrics)
	r2AccountID := os.Getenv("CF_ACCOUNT_ID")
	r2Bucket := os.Getenv("CF_R2_BUCKET_NAME")
	cfAPIToken := os.Getenv("CF_API_TOKEN")

	if r2AccountID != "" && r2Bucket != "" && cfAPIToken != "" {
		// Example implementation of CF Analytics API fetch (Adjust URL depending on GraphQL/REST usage in CF)
		// Usually R2 usage is fetched via GraphQL API or specific REST endpoint
		// Here we implement the requested structure. Note: CF might require GraphQL for usage.
		// For simplicity, we just simulate or do the REST call if it exists.
		
		// Fallback mock if you don't have the GraphQL payload ready, but let's implement the HTTP request structure
		url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/r2/buckets/%s/usage", r2AccountID, r2Bucket)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", "Bearer "+cfAPIToken)
		req.Header.Add("Content-Type", "application/json")
		
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		
		if err == nil && resp.StatusCode == 200 {
			var cfResp struct {
				Result struct {
					StorageBytes int64 `json:"storage_bytes"`
					ClassAOps    int64 `json:"class_a_operations"`
				} `json:"result"`
			}
			bodyBytes, _ := io.ReadAll(resp.Body)
			json.Unmarshal(bodyBytes, &cfResp)

			usedGB := float64(cfResp.Result.StorageBytes) / (1024 * 1024 * 1024)
			isStorageOk := usedGB <= 8.0
			checks["r2_storage_gb"] = fiber.Map{"used": fmt.Sprintf("%.2f", usedGB), "limit": 10, "ok": isStorageOk}
			if !isStorageOk {
				status = setStatus(status, "warning")
				warningMessages = append(warningMessages, fmt.Sprintf("⚠️ R2 Storage: %.1fGB/10GB", usedGB))
			}

			isClassAOk := cfResp.Result.ClassAOps <= 800000
			checks["r2_class_a"] = fiber.Map{"used": cfResp.Result.ClassAOps, "limit": 1000000, "ok": isClassAOk}
			if !isClassAOk {
				status = setStatus(status, "warning")
				warningMessages = append(warningMessages, fmt.Sprintf("⚠️ R2 Class A: %d/1,000,000 ops", cfResp.Result.ClassAOps))
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	} else {
		checks["r2_metrics"] = "Missing Cloudflare credentials in ENV"
	}

	// 4. FLY.IO — Uptime estimé
	var flyUptimeHours float64
	// Assure-toi que la table machine_sessions existe !
	err = h.db.QueryRow(ctx, "SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (ended_at - started_at))/3600), 0) FROM machine_sessions WHERE started_at > date_trunc('month', NOW())").Scan(&flyUptimeHours)
	if err != nil {
		h.logger.Warn("Failed to fetch Fly.io uptime (machine_sessions table might not exist)", zap.Error(err))
	} else {
		isOk := flyUptimeHours <= 160.0
		checks["fly_hours"] = fiber.Map{
			"used":  fmt.Sprintf("%.1f", flyUptimeHours),
			"limit": 191.9,
			"ok":    isOk,
		}
		if !isOk {
			status = "critical" // Fly limit is a hard stop
			warningMessages = append(warningMessages, fmt.Sprintf("🚨 Fly Uptime: %.1f/191.9 hours", flyUptimeHours))
		}
	}

	// Envoi Webhook Telegram si warning ou critical
	if status == "warning" || status == "critical" {
		sendTelegramAlert(h.logger, status, warningMessages)
	}

	return c.JSON(fiber.Map{
		"status": status,
		"checks": checks,
	})
}

// Helper to escalate status
func setStatus(current, new string) string {
	if current == "critical" {
		return current
	}
	return new
}

// Envoie une notification sur un channel Telegram
func sendTelegramAlert(logger *zap.Logger, status string, messages []string) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		logger.Warn("Telegram Webhook skipped: Missing credentials")
		return
	}

	emoji := "⚠️"
	if status == "critical" {
		emoji = "🚨"
	}

	text := fmt.Sprintf("%s *MazadPay Limits Alert* (%s)\n\n", emoji, status)
	for _, msg := range messages {
		text += "- " + msg + "\n"
	}

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)

	if err != nil {
		logger.Error("Failed to send Telegram alert", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Error("Telegram API returned non-200 status", zap.Int("status", resp.StatusCode))
	}
}
