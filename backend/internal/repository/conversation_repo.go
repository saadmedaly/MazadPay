package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type ConversationRepository interface {
	Create(ctx context.Context, conversation *models.Conversation) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	GetByIDWithParticipants(ctx context.Context, id uuid.UUID) (*models.Conversation, []models.ConversationParticipant, error)
	GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error)
	GetSupportConversations(ctx context.Context, limit, offset int) ([]models.UserConversation, error)
	GetAllConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error)
	Update(ctx context.Context, conversation *models.Conversation) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Participants
	AddParticipant(ctx context.Context, participant *models.ConversationParticipant) error
	RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error
	GetParticipants(ctx context.Context, conversationID uuid.UUID) ([]models.ConversationParticipant, error)
	GetParticipant(ctx context.Context, conversationID, userID uuid.UUID) (*models.ConversationParticipant, error)
	IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	UpdateParticipant(ctx context.Context, participant *models.ConversationParticipant) error
	UpdateUnreadCount(ctx context.Context, conversationID, userID uuid.UUID, count int) error
	IncrementUnreadCount(ctx context.Context, conversationID uuid.UUID, excludeUserID uuid.UUID) error
	MarkAsRead(ctx context.Context, conversationID, userID uuid.UUID, messageID *uuid.UUID) error
	
	// Direct conversation
	GetDirectConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error)
	
	// Last message
	UpdateLastMessage(ctx context.Context, conversationID uuid.UUID, preview string, senderID uuid.UUID) error
}

type conversationRepository struct {
	db *sqlx.DB
}

func NewConversationRepository(db *sqlx.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, conversation *models.Conversation) error {
	query := `
		INSERT INTO conversations (id, type, title, created_by, created_at, updated_at, is_active, metadata)
		VALUES (:id, :type, :title, :created_by, :created_at, :updated_at, :is_active, :metadata)
	`
	
	if conversation.ID == uuid.Nil {
		conversation.ID = uuid.New()
	}
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()
	conversation.IsActive = true
	
	_, err := r.db.NamedExecContext(ctx, query, conversation)
	return err
}

func (r *conversationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	query := `SELECT * FROM conversations WHERE id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &conversation, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	}
	return &conversation, err
}

func (r *conversationRepository) GetByIDWithParticipants(ctx context.Context, id uuid.UUID) (*models.Conversation, []models.ConversationParticipant, error) {
	conversation, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	
	participants, err := r.GetParticipants(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	
	return conversation, participants, nil
}

func (r *conversationRepository) GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error) {
	var conversations []models.UserConversation
	query := `
		SELECT * FROM user_conversations 
		WHERE user_id = $1 
		ORDER BY last_message_at DESC NULLS LAST, updated_at DESC 
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &conversations, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Fetch participants for each conversation
	for i := range conversations {
		p, err := r.GetParticipants(ctx, conversations[i].ConversationID)
		if err == nil {
			conversations[i].Participants = p
		}
	}

	return conversations, nil
}

func (r *conversationRepository) GetSupportConversations(ctx context.Context, limit, offset int) ([]models.UserConversation, error) {
	var conversations []models.UserConversation
	query := `
		SELECT 
			NULL as user_id,
			c.id as conversation_id,
			c.type,
			c.title,
			c.last_message_at,
			c.last_message_preview,
			c.last_message_sender_id,
			c.is_active,
			'admin' as role,
			c.created_at as joined_at,
			NULL as last_read_at,
			0 as unread_count,
			FALSE as is_muted
		FROM conversations c
		WHERE c.type = 'support' AND c.is_active = TRUE
		ORDER BY c.last_message_at DESC NULLS LAST, c.updated_at DESC
		LIMIT $1 OFFSET $2
	`
	err := r.db.SelectContext(ctx, &conversations, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Fetch participants for each conversation
	for i := range conversations {
		p, err := r.GetParticipants(ctx, conversations[i].ConversationID)
		if err == nil {
			conversations[i].Participants = p
		}
	}

	return conversations, nil
}

func (r *conversationRepository) GetAllConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error) {
	var conversations []models.UserConversation
	query := `
		SELECT 
			$1 as user_id,
			c.id as conversation_id,
			c.type,
			c.title,
			c.last_message_at,
			c.last_message_preview,
			c.last_message_sender_id,
			c.is_active,
			COALESCE(cp.role, 'admin') as role,
			COALESCE(cp.joined_at, c.created_at) as joined_at,
			cp.last_read_at,
			COALESCE(cp.unread_count, 0) as unread_count,
			COALESCE(cp.is_muted, FALSE) as is_muted
		FROM conversations c
		LEFT JOIN conversation_participants cp ON c.id = cp.conversation_id AND cp.user_id = $1
		WHERE c.is_active = TRUE
		ORDER BY c.last_message_at DESC NULLS LAST, c.updated_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &conversations, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Fetch participants for each conversation
	for i := range conversations {
		p, err := r.GetParticipants(ctx, conversations[i].ConversationID)
		if err == nil {
			conversations[i].Participants = p
		}
	}

	return conversations, nil
}

func (r *conversationRepository) Update(ctx context.Context, conversation *models.Conversation) error {
	query := `
		UPDATE conversations 
		SET type = :type, title = :title, updated_at = :updated_at, 
		    last_message_at = :last_message_at, last_message_preview = :last_message_preview,
		    last_message_sender_id = :last_message_sender_id, is_active = :is_active,
		    metadata = :metadata
		WHERE id = :id
	`
	conversation.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, conversation)
	return err
}

func (r *conversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE conversations SET is_active = false, updated_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

// Participants

func (r *conversationRepository) AddParticipant(ctx context.Context, participant *models.ConversationParticipant) error {
	query := `
		INSERT INTO conversation_participants 
		(id, conversation_id, user_id, role, joined_at, last_read_at, is_muted, unread_count)
		VALUES (:id, :conversation_id, :user_id, :role, :joined_at, :last_read_at, :is_muted, :unread_count)
	`
	
	if participant.ID == uuid.Nil {
		participant.ID = uuid.New()
	}
	participant.JoinedAt = time.Now()
	participant.UnreadCount = 0
	
	_, err := r.db.NamedExecContext(ctx, query, participant)
	return err
}

func (r *conversationRepository) RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error {
	query := `DELETE FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, conversationID, userID)
	return err
}

func (r *conversationRepository) GetParticipants(ctx context.Context, conversationID uuid.UUID) ([]models.ConversationParticipant, error) {
	var results []struct {
		models.ConversationParticipant
		UserID        uuid.UUID `db:"u_id"`
		FullName      *string   `db:"u_full_name"`
		ProfilePicURL *string   `db:"u_profile_pic_url"`
		UserRole      string    `db:"u_role"`
		IsActive      bool      `db:"u_is_active"`
	}

	query := `
		SELECT cp.*, u.id as u_id, u.full_name as u_full_name, 
		       u.profile_pic_url as u_profile_pic_url, u.role as u_role,
		       u.is_active as u_is_active
		FROM conversation_participants cp
		LEFT JOIN users u ON cp.user_id = u.id
		WHERE cp.conversation_id = $1
	`
	err := r.db.SelectContext(ctx, &results, query, conversationID)
	if err != nil {
		return nil, err
	}

	participants := make([]models.ConversationParticipant, len(results))
	for i, res := range results {
		participants[i] = res.ConversationParticipant
		participants[i].User = &models.User{
			ID:            res.UserID,
			FullName:      res.FullName,
			ProfilePicURL: res.ProfilePicURL,
			Role:          res.UserRole,
			IsActive:      res.IsActive,
		}
	}

	return participants, err
}

func (r *conversationRepository) GetParticipant(ctx context.Context, conversationID, userID uuid.UUID) (*models.ConversationParticipant, error) {
	var participant models.ConversationParticipant
	query := `SELECT * FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &participant, query, conversationID, userID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("participant not found")
	}
	return &participant, err
}

func (r *conversationRepository) IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, conversationID, userID)
	return exists, err
}

func (r *conversationRepository) UpdateParticipant(ctx context.Context, participant *models.ConversationParticipant) error {
	query := `
		UPDATE conversation_participants 
		SET role = :role, last_read_at = :last_read_at, last_read_message_id = :last_read_message_id,
		    is_muted = :is_muted, unread_count = :unread_count
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, participant)
	return err
}

func (r *conversationRepository) UpdateUnreadCount(ctx context.Context, conversationID, userID uuid.UUID, count int) error {
	query := `
		UPDATE conversation_participants 
		SET unread_count = $1 
		WHERE conversation_id = $2 AND user_id = $3
	`
	_, err := r.db.ExecContext(ctx, query, count, conversationID, userID)
	return err
}

func (r *conversationRepository) IncrementUnreadCount(ctx context.Context, conversationID uuid.UUID, excludeUserID uuid.UUID) error {
	query := `
		UPDATE conversation_participants 
		SET unread_count = unread_count + 1 
		WHERE conversation_id = $1 AND user_id != $2
	`
	_, err := r.db.ExecContext(ctx, query, conversationID, excludeUserID)
	return err
}

func (r *conversationRepository) MarkAsRead(ctx context.Context, conversationID, userID uuid.UUID, messageID *uuid.UUID) error {
	query := `
		UPDATE conversation_participants 
		SET last_read_at = $1, last_read_message_id = $2, unread_count = 0
		WHERE conversation_id = $3 AND user_id = $4
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), messageID, conversationID, userID)
	return err
}

// Direct conversation

func (r *conversationRepository) GetDirectConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	query := `
		SELECT c.* FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id AND cp1.user_id = $1
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id AND cp2.user_id = $2
		WHERE c.type = 'direct' AND c.is_active = true
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &conversation, query, userID1, userID2)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &conversation, err
}

// Last message

func (r *conversationRepository) UpdateLastMessage(ctx context.Context, conversationID uuid.UUID, preview string, senderID uuid.UUID) error {
	query := `
		UPDATE conversations 
		SET last_message_at = $1, last_message_preview = $2, last_message_sender_id = $3, updated_at = $1
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), preview, senderID, conversationID)
	return err
}
