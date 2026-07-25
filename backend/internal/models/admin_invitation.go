package models

import (
	"time"

	"github.com/google/uuid"
)

type AdminInvitation struct {
	ID        uuid.UUID  `db:"id"         json:"id"`
	Token     string     `db:"token"      json:"token"`
	CreatedBy uuid.UUID  `db:"created_by" json:"created_by"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt    *time.Time `db:"used_at"    json:"used_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`

	// Ajoutés par 000043_admin_invitations_target_phone (Admin Authorization Phase
	// 1C-A) — nullable, aucune vérification n'est encore appliquée (voir Phase 1C-B).
	// Jamais le numéro complet en clair : TargetPhoneHash sert uniquement à la
	// comparaison (hash), TargetPhoneMasked uniquement à l'affichage/audit.
	TargetPhoneHash   *string `db:"target_phone_hash"   json:"target_phone_hash,omitempty"`
	TargetPhoneMasked *string `db:"target_phone_masked" json:"target_phone_masked,omitempty"`
}
