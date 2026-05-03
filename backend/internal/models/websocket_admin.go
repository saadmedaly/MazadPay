package models

// AdminEvent — événements envoyés aux admins en temps réel
type AdminEvent struct {
	Type    string      `json:"type"` // new_request | request_updated | request_reviewed | new_bid | auction_ended
	Payload interface{} `json:"payload"`
}

// NewRequestPayload — nouvelle demande reçue
type NewRequestPayload struct {
	RequestID   string `json:"request_id"`
	RequestType string `json:"request_type"` // auction | banner
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Title       string `json:"title"`
	CreatedAt   string `json:"created_at"`
}

// RequestUpdatedPayload — demande mise à jour
type RequestUpdatedPayload struct {
	RequestID   string `json:"request_id"`
	RequestType string `json:"request_type"`
	Status      string `json:"status"`
	UpdatedBy   string `json:"updated_by"`
	UpdatedAt   string `json:"updated_at"`
}
