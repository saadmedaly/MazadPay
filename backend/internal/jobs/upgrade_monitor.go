package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UpgradeSignal struct {
	Tool      string
	Reason    string
	Current   float64
	Threshold float64
	Urgent    bool
}

// CheckUpgradeNeeded évalue si le volume de données ou de requêtes nécessite de quitter le Free Tier
func CheckUpgradeNeeded(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) []UpgradeSignal {
	signals := []UpgradeSignal{}

	// 1. NEON — taille BDD
	var dbSizeMB float64
	err := pool.QueryRow(ctx, "SELECT pg_database_size(current_database()) / 1024.0 / 1024.0").Scan(&dbSizeMB)
	if err == nil && dbSizeMB > 400 { // 80% de 512MB
		signals = append(signals, UpgradeSignal{
			Tool:      "Neon DB",
			Reason:    "Stockage BDD critique (> 400MB)",
			Current:   dbSizeMB,
			Threshold: 512,
			Urgent:    dbSizeMB > 460,
		})
	}

	// 2. USERS ACTIFS — Volume trop élevé pour le Free Tier
	var activeUsers int
	// Note: Adapt query to your exact schema. We assume a 'sessions' table here.
	err = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT user_id) FROM sessions WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&activeUsers)
	if err == nil && activeUsers > 1500 {
		signals = append(signals, UpgradeSignal{
			Tool:      "Fly.io + Neon",
			Reason:    "Volume d'utilisateurs actifs élevé",
			Current:   float64(activeUsers),
			Threshold: 2000,
			Urgent:    activeUsers > 1800,
		})
	}

	// 3. COMPUTE HOURS NEON — Estimation grossière
	var hoursUsed float64
	err = pool.QueryRow(ctx, `
		SELECT (EXTRACT(EPOCH FROM (NOW() - date_trunc('month', NOW()))) / 3600.0)
		* (SELECT LEAST(COUNT(*), 100) FROM pg_stat_activity) / 100.0
	`).Scan(&hoursUsed)
	if err == nil && hoursUsed > 153 { // 80% de 191.9h
		signals = append(signals, UpgradeSignal{
			Tool:      "Neon Compute",
			Reason:    "Compute hours mensuels en danger",
			Current:   hoursUsed,
			Threshold: 191.9,
			Urgent:    hoursUsed > 172,
		})
	}

	return signals
}

// NotifyUpgrade envoie une alerte Telegram si un upgrade est indispensable
func NotifyUpgrade(signals []UpgradeSignal, logger *zap.Logger) {
	if len(signals) == 0 {
		return
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		logger.Warn("Telegram Webhook manquant pour envoyer les signaux d'upgrade")
		return
	}

	for _, s := range signals {
		emoji := "⚠️"
		if s.Urgent {
			emoji = "🚨"
		}
		
		msg := fmt.Sprintf("%s *UPGRADE REQUIS : %s*\n_%s_\nValeur actuelle : %.1f\nSeuil Free Tier : %.1f",
			emoji, s.Tool, s.Reason, s.Current, s.Threshold)

		payload := map[string]interface{}{
			"chat_id":    chatID,
			"text":       msg,
			"parse_mode": "Markdown",
		}
		
		payloadBytes, _ := json.Marshal(payload)
		_, err := http.Post(
			fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
			"application/json",
			bytes.NewBuffer(payloadBytes),
		)
		if err != nil {
			logger.Error("Erreur d'envoi de l'alerte d'upgrade", zap.Error(err))
		}
	}
}
