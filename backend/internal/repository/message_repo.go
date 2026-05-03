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

type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*models.Message, error)
	GetByConversation(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]models.Message, error)
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
	MarkAsDeleted(ctx context.Context, id uuid.UUID) error
	
	// Message status
	CreateMessageStatus(ctx context.Context, status *models.MessageStatus) error
	UpdateMessageStatus(ctx context.Context, messageID, userID uuid.UUID, status string) error
	GetMessageStatus(ctx context.Context, messageID uuid.UUID) ([]models.MessageStatus, error)
	GetUnreadMessages(ctx context.Context, conversationID, userID uuid.UUID) ([]models.Message, error)
	
	// Bulk operations
	MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, upToMessageID *uuid.UUID) error
	GetMessagesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Message, error)
}

type messageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (
			id, conversation_id, sender_id, type, content, file_name, file_url,
			file_size, file_duration, mime_type, thumbnail_url, reply_to_id,
			is_edited, is_deleted, metadata, created_at, updated_at
		) VALUES (
			:id, :conversation_id, :sender_id, :type, :content, :file_name, :file_url,
			:file_size, :file_duration, :mime_type, :thumbnail_url, :reply_to_id,
			:is_edited, :is_deleted, :metadata, :created_at, :updated_at
		)
	`
	
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	message.IsEdited = false
	message.IsDeleted = false
	
	_, err := r.db.NamedExecContext(ctx, query, message)
	return err
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	query := `SELECT * FROM messages WHERE id = $1`
	err := r.db.GetContext(ctx, &message, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	return &message, err
}

func (r *messageRepository) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	query := `
		SELECT m.*, 
		       u.id as "sender.id", u.full_name as "sender.full_name", 
		       u.profile_pic_url as "sender.profile_pic_url",
		       rm.id as "reply_to.id", rm.content as "reply_to.content", rm.type as "reply_to.type"
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		LEFT JOIN messages rm ON m.reply_to_id = rm.id
		WHERE m.id = $1
	`
	err := r.db.GetContext(ctx, &message, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	return &message, err
}

func (r *messageRepository) GetByConversation(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]models.Message, error) {
	var results []struct {
		models.Message
		SenderID      *uuid.UUID `db:"s_id"`
		SenderName    *string    `db:"s_name"`
		SenderPic     *string    `db:"s_pic"`
		ReplyID       *uuid.UUID `db:"r_id"`
		ReplyContent  *string    `db:"r_content"`
		ReplyType     *string    `db:"r_type"`
	}

	query := `
		SELECT m.*, 
		       u.id as s_id, u.full_name as s_name, u.profile_pic_url as s_pic,
		       rm.id as r_id, rm.content as r_content, rm.type as r_type
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		LEFT JOIN messages rm ON m.reply_to_id = rm.id
		WHERE m.conversation_id = $1 AND m.is_deleted = false
		ORDER BY m.created_at ASC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &results, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, len(results))
	for i, res := range results {
		messages[i] = res.Message
		if res.SenderID != nil {
			messages[i].Sender = &models.User{
				ID:            *res.SenderID,
				FullName:      res.SenderName,
				ProfilePicURL: res.SenderPic,
			}
		}
		if res.ReplyID != nil {
			messages[i].ReplyTo = &models.Message{
				ID:      *res.ReplyID,
				Content: res.ReplyContent,
				Type:    *res.ReplyType,
			}
		}
	}

	return messages, nil
}

func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	query := `
		UPDATE messages 
		SET content = :content, is_edited = :is_edited, updated_at = :updated_at
		WHERE id = :id
	`
	message.UpdatedAt = time.Now()
	message.IsEdited = true
	_, err := r.db.NamedExecContext(ctx, query, message)
	return err
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *messageRepository) MarkAsDeleted(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE messages 
		SET is_deleted = true, deleted_at = $1, content = NULL, file_url = NULL
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// Message status

func (r *messageRepository) CreateMessageStatus(ctx context.Context, status *models.MessageStatus) error {
	query := `
		INSERT INTO message_status (id, message_id, user_id, status, updated_at)
		VALUES (:id, :message_id, :user_id, :status, :updated_at)
	`
	
	if status.ID == uuid.Nil {
		status.ID = uuid.New()
	}
	status.UpdatedAt = time.Now()
	
	_, err := r.db.NamedExecContext(ctx, query, status)
	return err
}

func (r *messageRepository) UpdateMessageStatus(ctx context.Context, messageID, userID uuid.UUID, status string) error {
	query := `
		INSERT INTO message_status (id, message_id, user_id, status, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (message_id, user_id) 
		DO UPDATE SET status = $4, updated_at = $5
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New(), messageID, userID, status, time.Now())
	return err
}

func (r *messageRepository) GetMessageStatus(ctx context.Context, messageID uuid.UUID) ([]models.MessageStatus, error) {
	var statuses []models.MessageStatus
	query := `SELECT * FROM message_status WHERE message_id = $1`
	err := r.db.SelectContext(ctx, &statuses, query, messageID)
	return statuses, err
}

func (r *messageRepository) GetUnreadMessages(ctx context.Context, conversationID, userID uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	query := `
		SELECT m.* FROM messages m
		LEFT JOIN message_status ms ON m.id = ms.message_id AND ms.user_id = $1
		WHERE m.conversation_id = $2 
		  AND m.sender_id != $1 
		  AND (ms.status IS NULL OR ms.status != 'read')
		  AND m.is_deleted = false
	`
	err := r.db.SelectContext(ctx, &messages, query, userID, conversationID)
	return messages, err
}

func (r *messageRepository) MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, upToMessageID *uuid.UUID) error {
	var query string
	var args []interface{}
	
	if upToMessageID != nil {
		// Marquer jusqu'à un message spécifique
		query = `
			INSERT INTO message_status (id, message_id, user_id, status, updated_at)
			SELECT uuid_generate_v4(), m.id, $1, 'read', $2
			FROM messages m
			WHERE m.conversation_id = $3 
			  AND m.sender_id != $1 
			  AND m.created_at <= (SELECT created_at FROM messages WHERE id = $4)
			  AND m.is_deleted = false
			ON CONFLICT (message_id, user_id) 
			DO UPDATE SET status = 'read', updated_at = $2
		`
		args = []interface{}{userID, time.Now(), conversationID, *upToMessageID}
	} else {
		// Marquer tous les messages non lus
		query = `
			INSERT INTO message_status (id, message_id, user_id, status, updated_at)
			SELECT uuid_generate_v4(), m.id, $1, 'read', $2
			FROM messages m
			WHERE m.conversation_id = $3 
			  AND m.sender_id != $1 
			  AND m.is_deleted = false
			  AND NOT EXISTS (
				  SELECT 1 FROM message_status ms 
				  WHERE ms.message_id = m.id AND ms.user_id = $1 AND ms.status = 'read'
			  )
			ON CONFLICT (message_id, user_id) 
			DO UPDATE SET status = 'read', updated_at = $2
		`
		args = []interface{}{userID, time.Now(), conversationID}
	}
	
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *messageRepository) GetMessagesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Message, error) {
	if len(ids) == 0 {
		return []models.Message{}, nil
	}
	
	var messages []models.Message
	query, args, err := sqlx.In(`
		SELECT m.*, 
		       u.id as "sender.id", u.full_name as "sender.full_name", 
		       u.profile_pic_url as "sender.profile_pic_url"
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.id IN (?) AND m.is_deleted = false
	`, ids)
	if err != nil {
		return nil, err
	}
	
	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &messages, query, args...)
	return messages, err
}
