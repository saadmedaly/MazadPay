package models

import (
	"time"

	"github.com/google/uuid"
)

// Conversation représente une conversation entre utilisateurs
type Conversation struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	Type                string     `db:"type" json:"type"` // 'direct', 'group', 'support'
	Title               *string    `db:"title" json:"title,omitempty"`
	CreatedBy           *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
	LastMessageAt       *time.Time `db:"last_message_at" json:"last_message_at,omitempty"`
	LastMessagePreview  *string    `db:"last_message_preview" json:"last_message_preview,omitempty"`
	LastMessageSenderID *uuid.UUID `db:"last_message_sender_id" json:"last_message_sender_id,omitempty"`
	IsActive            bool       `db:"is_active" json:"is_active"`
	Metadata            JSONB      `db:"metadata" json:"metadata,omitempty"`
	
	// Relations
	Participants        []ConversationParticipant `db:"-" json:"participants,omitempty"`
}

// ConversationParticipant représente un participant à une conversation
type ConversationParticipant struct {
	ID               uuid.UUID  `db:"id" json:"id"`
	ConversationID   uuid.UUID  `db:"conversation_id" json:"conversation_id"`
	UserID           uuid.UUID  `db:"user_id" json:"user_id"`
	Role             string     `db:"role" json:"role"` // 'owner', 'admin', 'member'
	JoinedAt         time.Time  `db:"joined_at" json:"joined_at"`
	LastReadAt       *time.Time `db:"last_read_at" json:"last_read_at,omitempty"`
	LastReadMessageID *uuid.UUID `db:"last_read_message_id" json:"last_read_message_id,omitempty"`
	IsMuted          bool       `db:"is_muted" json:"is_muted"`
	UnreadCount      int        `db:"unread_count" json:"unread_count"`
	
	// Relations
	User *User `db:"user" json:"user,omitempty"`
}

// Message représente un message dans une conversation
type Message struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	ConversationID uuid.UUID  `db:"conversation_id" json:"conversation_id"`
	SenderID       *uuid.UUID `db:"sender_id" json:"sender_id,omitempty"`
	Type           string     `db:"type" json:"type"` // 'text', 'audio', 'video', 'image', 'file', 'system'
	Content        *string    `db:"content" json:"content,omitempty"`
	FileName       *string    `db:"file_name" json:"file_name,omitempty"`
	FileURL        *string    `db:"file_url" json:"file_url,omitempty"`
	FileSize       *int       `db:"file_size" json:"file_size,omitempty"`
	FileDuration   *int       `db:"file_duration" json:"file_duration,omitempty"` // En secondes
	MimeType       *string    `db:"mime_type" json:"mime_type,omitempty"`
	ThumbnailURL   *string    `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	ReplyToID      *uuid.UUID `db:"reply_to_id" json:"reply_to_id,omitempty"`
	IsEdited       bool       `db:"is_edited" json:"is_edited"`
	IsDeleted      bool       `db:"is_deleted" json:"is_deleted"`
	DeletedAt      *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	Metadata       JSONB      `db:"metadata" json:"metadata,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
	
	// Relations
	Sender       *User    `db:"sender" json:"sender,omitempty"`
	ReplyTo      *Message `db:"reply_to" json:"reply_to,omitempty"`
	Status       []MessageStatus `db:"status" json:"status,omitempty"`
}

// MessageStatus représente le statut d'un message pour un utilisateur
type MessageStatus struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	MessageID uuid.UUID  `db:"message_id" json:"message_id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Status    string     `db:"status" json:"status"` // 'sent', 'delivered', 'read'
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// UserConversation vue pour les conversations d'un utilisateur
type UserConversation struct {
	UserID              uuid.UUID  `db:"user_id" json:"user_id"`
	ConversationID      uuid.UUID  `db:"conversation_id" json:"id"`
	Type                string     `db:"type" json:"type"`
	Title               *string    `db:"title" json:"title,omitempty"`
	LastMessageAt       *time.Time `db:"last_message_at" json:"last_message_at,omitempty"`
	LastMessagePreview  *string    `db:"last_message_preview" json:"last_message_preview,omitempty"`
	LastMessageSenderID *uuid.UUID `db:"last_message_sender_id" json:"last_message_sender_id,omitempty"`
	IsActive            bool       `db:"is_active" json:"is_active"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
	Role                string     `db:"role" json:"role"`
	JoinedAt            time.Time  `db:"joined_at" json:"joined_at"`

	LastReadAt          *time.Time `db:"last_read_at" json:"last_read_at,omitempty"`
	UnreadCount         int        `db:"unread_count" json:"unread_count"`
	IsMuted             bool       `db:"is_muted" json:"is_muted"`
	
	// Added for frontend to identify participants
	Participants        []ConversationParticipant `json:"participants,omitempty"`
}

// WebSocket Event Types
const (
	ChatEventMessageNew       = "message:new"
	ChatEventMessageRead      = "message:read"
	ChatEventMessageDelivered = "message:delivered"
	ChatEventTypingStart      = "typing:start"
	ChatEventTypingStop       = "typing:stop"
	ChatEventUserOnline       = "user:online"
	ChatEventUserOffline      = "user:offline"
	ChatEventConversationJoin = "conversation:join"
	ChatEventConversationLeave = "conversation:leave"
	ChatEventError            = "error"
)

// WebSocketMessage représente un message WebSocket
type WebSocketMessage struct {
	Event       string      `json:"event"`
	ConversationID uuid.UUID `json:"conversation_id,omitempty"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
}

// CreateConversationRequest représente une requête de création de conversation
type CreateConversationRequest struct {
	Type      string      `json:"type" validate:"required,oneof=direct group support"`
	Title     *string     `json:"title,omitempty"`
	UserIDs   []uuid.UUID `json:"user_ids" validate:"required,min=1"`
	InitialMessage *string `json:"initial_message,omitempty"`
}

// SendMessageRequest représente une requête d'envoi de message
type SendMessageRequest struct {
	Type         string      `json:"type" validate:"required,oneof=text audio video image file"`
	Content      *string     `json:"content,omitempty"`
	FileName     *string     `json:"file_name,omitempty"`
	FileURL      *string     `json:"file_url,omitempty"`
	FileSize     *int        `json:"file_size,omitempty"`
	FileDuration *int        `json:"file_duration,omitempty"`
	MimeType     *string     `json:"mime_type,omitempty"`
	ThumbnailURL *string     `json:"thumbnail_url,omitempty"`
	ReplyToID    *uuid.UUID  `json:"reply_to_id,omitempty"`
}

// MarkReadRequest représente une requête de marquage comme lu
type MarkReadRequest struct {
	MessageID *uuid.UUID `json:"message_id,omitempty"` // Si vide, marque tous les messages de la conversation
}

// ConversationResponse représente une réponse avec conversation et participants
type ConversationResponse struct {
	Conversation  Conversation                `json:"conversation"`
	Participants  []ConversationParticipant   `json:"participants"`
	UnreadCount   int                         `json:"unread_count"`
}

// MessageResponse représente une réponse avec message et statut
type MessageResponse struct {
	Message  Message          `json:"message"`
	Status   []MessageStatus  `json:"status,omitempty"`
}
