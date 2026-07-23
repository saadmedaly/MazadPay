package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	AdminID    uuid.UUID  `db:"admin_id" json:"admin_id"`
	Action     string     `db:"action" json:"action"`
	EntityType string     `db:"entity_type" json:"entity_type"`
	EntityID   *uuid.UUID `db:"entity_id" json:"entity_id"`
	Details    string     `db:"details" json:"details"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`

	// Colonnes ajoutées par 000042_audit_logs_schema_improvement (Audit Logs Schema
	// Improvement, Phase A) — toutes nullable/avec défaut, aucune n'est requise pour
	// les écritures existantes via AuditService.Log. Activation progressive prévue en
	// Phase B (voir functional options WithActorType/WithIP/WithUserAgent/
	// WithDetailsJSON/WithEntityKey dans audit_service.go).
	ActorID     *uuid.UUID `db:"actor_id"     json:"actor_id,omitempty"`
	ActorType   string     `db:"actor_type"   json:"actor_type"`
	IPAddress   *string    `db:"ip_address"   json:"ip_address,omitempty"`
	UserAgent   *string    `db:"user_agent"   json:"user_agent,omitempty"`
	DetailsJSON JSONB      `db:"details_json" json:"details_json,omitempty"`
	EntityKey   *string    `db:"entity_key"   json:"entity_key,omitempty"`
}