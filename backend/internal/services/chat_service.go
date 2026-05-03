package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	ws "github.com/mazadpay/backend/internal/websocket"
	"go.uber.org/zap"
)

// ChatService définit les opérations de messagerie
type ChatService interface {
	// Conversations
	CreateConversation(ctx context.Context, req *models.CreateConversationRequest, creatorID uuid.UUID) (*models.Conversation, error)
	GetConversation(ctx context.Context, conversationID, userID uuid.UUID) (*models.Conversation, []models.ConversationParticipant, error)
	GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error)
	GetAdminConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error)
	GetSupportConversations(ctx context.Context, limit, offset int) ([]models.UserConversation, error)
	JoinConversation(ctx context.Context, conversationID, userID uuid.UUID) error
	LeaveConversation(ctx context.Context, conversationID, userID uuid.UUID) error
	GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error)

	// Messages
	SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, req *models.SendMessageRequest) (*models.Message, error)
	GetMessages(ctx context.Context, conversationID, userID uuid.UUID, limit, offset int) ([]models.Message, error)
	GetMessage(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
	EditMessage(ctx context.Context, messageID, userID uuid.UUID, newContent string) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error
	MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, upToMessageID *uuid.UUID) error

	// Typing indicators
	SetTyping(ctx context.Context, conversationID, userID uuid.UUID, isTyping bool)
}

type chatService struct {
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
	userRepo         repository.UserRepository
	notificationSvc  NotificationService
	logger           *zap.Logger
	wsHub            *ws.ChatHub
	pushSem          chan struct{} // Sémaphore pour limiter les goroutines de push
}

func NewChatService(
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	notificationSvc NotificationService,
	logger *zap.Logger,
	wsHub *ws.ChatHub,
) ChatService {
	return &chatService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		userRepo:         userRepo,
		notificationSvc:  notificationSvc,
		logger:           logger,
		wsHub:            wsHub,
		pushSem:          make(chan struct{}, 100), // Limite à 100 push concurrents
	}
}

 func (s *chatService) CreateConversation(ctx context.Context, req *models.CreateConversationRequest, creatorID uuid.UUID) (*models.Conversation, error) {
 	if req.Type == "group" {
		creator, err := s.userRepo.FindByID(ctx, creatorID)
		if err != nil {
			return nil, err
		}
		if creator.Role != "admin" && creator.Role != "super_admin" {
			return nil, fmt.Errorf("unauthorized: only admins can create group conversations")
		}
	}

	// Pour une conversation directe, vérifier si elle existe déjà
	if req.Type == "direct" && len(req.UserIDs) == 1 {
		existing, err := s.conversationRepo.GetDirectConversation(ctx, creatorID, req.UserIDs[0])
		if err != nil {
			return nil, err
		}
		if existing != nil {
			// Populate participants for the existing conversation
			_, participants, _ := s.conversationRepo.GetByIDWithParticipants(ctx, existing.ID)
			existing.Participants = participants
			return existing, nil
		}
	}

	// Créer la conversation
	conversation := &models.Conversation{
		Type:      req.Type,
		Title:     req.Title,
		CreatedBy: &creatorID,
	}

	if err := s.conversationRepo.Create(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Ajouter le créateur comme participant
	ownerRole := "owner"
	if req.Type == "direct" {
		ownerRole = "member"
	}

	creatorParticipant := &models.ConversationParticipant{
		ConversationID: conversation.ID,
		UserID:         creatorID,
		Role:           ownerRole,
	}
	if err := s.conversationRepo.AddParticipant(ctx, creatorParticipant); err != nil {
		return nil, fmt.Errorf("failed to add creator: %w", err)
	}

	// Ajouter les autres participants
	for _, userID := range req.UserIDs {
		if userID == creatorID {
			continue
		}

		participant := &models.ConversationParticipant{
			ConversationID: conversation.ID,
			UserID:         userID,
			Role:           "member",
		}
		if err := s.conversationRepo.AddParticipant(ctx, participant); err != nil {
			s.logger.Error("failed to add participant", zap.Error(err), zap.String("user_id", userID.String()))
			continue
		}
	}

	// Envoyer message initial si fourni
	if req.InitialMessage != nil && *req.InitialMessage != "" {
		msgReq := &models.SendMessageRequest{
			Type:    "text",
			Content: req.InitialMessage,
		}
		_, err := s.SendMessage(ctx, conversation.ID, creatorID, msgReq)
		if err != nil {
			s.logger.Error("failed to send initial message", zap.Error(err))
		}
	}

	s.logger.Info("conversation created",
		zap.String("conversation_id", conversation.ID.String()),
		zap.String("type", conversation.Type),
		zap.String("creator_id", creatorID.String()))

	// Populate participants before returning
	_, participants, _ := s.conversationRepo.GetByIDWithParticipants(ctx, conversation.ID)
	conversation.Participants = participants

	return conversation, nil
}

// GetConversation récupère une conversation avec ses participants (si l'utilisateur est participant)
func (s *chatService) GetConversation(ctx context.Context, conversationID, userID uuid.UUID) (*models.Conversation, []models.ConversationParticipant, error) {
	// Vérifier l'appartenance
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !isParticipant {
		return nil, nil, fmt.Errorf("unauthorized: user is not a participant")
	}

	return s.conversationRepo.GetByIDWithParticipants(ctx, conversationID)
}

func (s *chatService) GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error) {
	return s.conversationRepo.GetUserConversations(ctx, userID, limit, offset)
}

func (s *chatService) GetAdminConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserConversation, error) {
	return s.conversationRepo.GetUserConversations(ctx, userID, limit, offset)
}

func (s *chatService) GetSupportConversations(ctx context.Context, limit, offset int) ([]models.UserConversation, error) {
	return s.conversationRepo.GetSupportConversations(ctx, limit, offset)
}

// JoinConversation permet à un utilisateur de rejoindre une conversation (pour les groupes)
func (s *chatService) JoinConversation(ctx context.Context, conversationID, userID uuid.UUID) error {
	conversation, err := s.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}

	// Seuls les groupes peuvent être rejoints librement
	if conversation.Type != "group" && conversation.Type != "support" {
		return fmt.Errorf("can only join group or support conversations")
	}

	// Vérifier si déjà participant
	_, err = s.conversationRepo.GetParticipant(ctx, conversationID, userID)
	if err == nil {
		return fmt.Errorf("already a participant")
	}

	participant := &models.ConversationParticipant{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           "member",
	}

	return s.conversationRepo.AddParticipant(ctx, participant)
}

// LeaveConversation permet à un utilisateur de quitter une conversation
func (s *chatService) LeaveConversation(ctx context.Context, conversationID, userID uuid.UUID) error {
	return s.conversationRepo.RemoveParticipant(ctx, conversationID, userID)
}

// GetOrCreateDirectConversation récupère ou crée une conversation directe entre deux utilisateurs
func (s *chatService) GetOrCreateDirectConversation(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error) {
	// Vérifier si existe déjà
	existing, err := s.conversationRepo.GetDirectConversation(ctx, userID1, userID2)
	if err != nil {
		return nil, err
	}
	if existing != nil {
 		participants, _ := s.conversationRepo.GetParticipants(ctx, existing.ID)
		existing.Participants = participants
		return existing, nil
	}

 	req := &models.CreateConversationRequest{
		Type:    "direct",
		UserIDs: []uuid.UUID{userID2},
	}

	return s.CreateConversation(ctx, req, userID1)
}

// SendMessage envoie un message dans une conversation
func (s *chatService) SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, req *models.SendMessageRequest) (*models.Message, error) {
	// Vérifier que l'expéditeur est participant
	_, err := s.conversationRepo.GetParticipant(ctx, conversationID, senderID)
	if err != nil {
 		user, uerr := s.userRepo.FindByID(ctx, senderID)
		if uerr == nil && (user.Role == "admin" || user.Role == "super_admin") {
			// Auto-join l'admin
			participant := &models.ConversationParticipant{
				ConversationID: conversationID,
				UserID:         senderID,
				Role:           "admin",
			}
			s.conversationRepo.AddParticipant(ctx, participant)
		} else {
			return nil, fmt.Errorf("sender is not a participant: %w", err)
		}
	}

	// Vérifier la taille du fichier si présent
	if req.FileSize != nil && *req.FileSize > 10*1024*1024 {
		return nil, fmt.Errorf("file size exceeds 10MB limit")
	}

	// Créer le message
	message := &models.Message{
		ConversationID: conversationID,
		SenderID:       &senderID,
		Type:           req.Type,
		Content:        req.Content,
		FileName:       req.FileName,
		FileURL:        req.FileURL,
		FileSize:       req.FileSize,
		FileDuration:   req.FileDuration,
		MimeType:       req.MimeType,
		ThumbnailURL:   req.ThumbnailURL,
		ReplyToID:      req.ReplyToID,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Mettre à jour le dernier message de la conversation
	preview := s.generatePreview(req)
	if err := s.conversationRepo.UpdateLastMessage(ctx, conversationID, preview, senderID); err != nil {
		s.logger.Error("failed to update last message", zap.Error(err))
	}

	// Incrémenter le compteur non lu pour les autres participants
	if err := s.conversationRepo.IncrementUnreadCount(ctx, conversationID, senderID); err != nil {
		s.logger.Error("failed to increment unread count", zap.Error(err))
	}

	// Créer le statut "sent" pour l'expéditeur
	status := &models.MessageStatus{
		MessageID: message.ID,
		UserID:    senderID,
		Status:    "sent",
	}
	if err := s.messageRepo.CreateMessageStatus(ctx, status); err != nil {
		s.logger.Error("failed to create message status", zap.Error(err))
	}

	// Notifier via WebSocket tous les participants en ligne
	if s.wsHub != nil {
		msgWithSender, _ := s.messageRepo.GetByIDWithRelations(ctx, message.ID)
		if msgWithSender != nil {
			// Récupérer tous les participants pour leur envoyer le message individuellement
			// (car ils ne sont pas forcément "join" dans la room du hub)
			participants, _ := s.conversationRepo.GetParticipants(ctx, conversationID)
			for _, p := range participants {
				s.wsHub.BroadcastToUser(p.UserID, models.ChatEventMessageNew, msgWithSender)
			}
		}
	}

	// Envoyer notification push aux participants hors ligne (avec sémaphore)
	select {
	case s.pushSem <- struct{}{}:
		go func() {
			defer func() { <-s.pushSem }()
			s.sendPushNotifications(conversationID, senderID, message)
		}()
	default:
		// Si le sémaphore est plein, logger un warning et ne pas bloquer
		s.logger.Warn("push notification semaphore full, skipping push", zap.String("conversation_id", conversationID.String()))
	}

	return message, nil
}

// GetMessages récupère les messages d'une conversation (si l'utilisateur est participant)
func (s *chatService) GetMessages(ctx context.Context, conversationID, userID uuid.UUID, limit, offset int) ([]models.Message, error) {
	// Vérifier l'appartenance
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, fmt.Errorf("unauthorized: user is not a participant of this conversation")
	}

	return s.messageRepo.GetByConversation(ctx, conversationID, limit, offset)
}

// GetMessage récupère un message par ID
func (s *chatService) GetMessage(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	return s.messageRepo.GetByIDWithRelations(ctx, messageID)
}

// EditMessage modifie un message existant
func (s *chatService) EditMessage(ctx context.Context, messageID, userID uuid.UUID, newContent string) (*models.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Vérifier que l'utilisateur est l'expéditeur
	if message.SenderID == nil || *message.SenderID != userID {
		return nil, fmt.Errorf("unauthorized: can only edit own messages")
	}

	// Vérifier que c'est un message texte
	if message.Type != "text" {
		return nil, fmt.Errorf("can only edit text messages")
	}

	message.Content = &newContent
	if err := s.messageRepo.Update(ctx, message); err != nil {
		return nil, err
	}

	return s.messageRepo.GetByIDWithRelations(ctx, messageID)
}

// DeleteMessage supprime un message (soft delete)
func (s *chatService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Vérifier que l'utilisateur est l'expéditeur ou admin
	if message.SenderID == nil || *message.SenderID != userID {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user.Role != "admin" && user.Role != "super_admin" {
			return fmt.Errorf("unauthorized: can only delete own messages")
		}
	}

	return s.messageRepo.MarkAsDeleted(ctx, messageID)
}

// MarkMessagesAsRead marque les messages comme lus
func (s *chatService) MarkMessagesAsRead(ctx context.Context, conversationID, userID uuid.UUID, upToMessageID *uuid.UUID) error {
	// Vérifier l'appartenance
	isParticipant, err := s.conversationRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil || !isParticipant {
		return fmt.Errorf("unauthorized: user is not a participant")
	}

	// Mettre à jour le statut des messages
	if err := s.messageRepo.MarkMessagesAsRead(ctx, conversationID, userID, upToMessageID); err != nil {
		return err
	}

	// Réinitialiser le compteur non lu
	if err := s.conversationRepo.MarkAsRead(ctx, conversationID, userID, upToMessageID); err != nil {
		return err
	}

	// Notifier via WebSocket tous les participants en ligne
	if s.wsHub != nil {
		eventData := map[string]interface{}{
			"user_id":         userID,
			"conversation_id": conversationID,
			"read_at":         time.Now(),
			"message_id":      upToMessageID,
		}
		
		participants, _ := s.conversationRepo.GetParticipants(ctx, conversationID)
		for _, p := range participants {
			s.wsHub.BroadcastToUser(p.UserID, models.ChatEventMessageRead, eventData)
		}
	}

	return nil
}

// SetTyping gère les indicateurs de frappe
func (s *chatService) SetTyping(ctx context.Context, conversationID, userID uuid.UUID, isTyping bool) {
	if s.wsHub == nil {
		return
	}

	event := models.ChatEventTypingStop
	if isTyping {
		event = models.ChatEventTypingStart
	}

	s.wsHub.BroadcastToConversation(conversationID, event, map[string]interface{}{
		"user_id": userID,
		"typing":  isTyping,
	})
}

// Helper functions

func (s *chatService) generatePreview(req *models.SendMessageRequest) string {
	switch req.Type {
	case "text":
		if req.Content != nil {
			preview := *req.Content
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			return preview
		}
		return ""
	case "audio":
		return "🎤 Message audio"
	case "video":
		return "🎥 Vidéo"
	case "image":
		return "📷 Image"
	case "file":
		return "📎 Fichier"
	default:
		return "Nouveau message"
	}
}

func (s *chatService) sendPushNotifications(conversationID, senderID uuid.UUID, message *models.Message) {
	// Récupérer les participants
	participants, err := s.conversationRepo.GetParticipants(context.Background(), conversationID)
	if err != nil {
		s.logger.Error("failed to get participants for push", zap.Error(err))
		return
	}

	// Récupérer le nom de l'expéditeur
	sender, err := s.userRepo.FindByID(context.Background(), senderID)
	if err != nil {
		s.logger.Error("failed to get sender", zap.Error(err))
		return
	}

	senderName := ""
	if sender.FullName != nil {
		senderName = *sender.FullName
	}
	if senderName == "" {
		senderName = sender.Phone
	}

	// Envoyer notification à chaque participant (sauf l'expéditeur)
	for _, p := range participants {
		if p.UserID == senderID {
			continue
		}

		title := "Nouveau message"
		body := fmt.Sprintf("%s: %s", senderName, s.generatePreview(&models.SendMessageRequest{
			Type:    message.Type,
			Content: message.Content,
		}))

		data := map[string]string{
			"type":            "new_message",
			"conversation_id": conversationID.String(),
			"message_id":      message.ID.String(),
			"sender_id":       senderID.String(),
		}

		if err := s.notificationSvc.SendPush(context.Background(), p.UserID, title, body, "new_message", data); err != nil {
			s.logger.Error("failed to send push notification", zap.Error(err))
		}
	}
}
