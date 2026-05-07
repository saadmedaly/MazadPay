package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type MessageHandler struct {
	chatSvc services.ChatService
	logger  *zap.Logger
}

func NewMessageHandler(chatSvc services.ChatService, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		chatSvc: chatSvc,
		logger:  logger,
	}
}

func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	var req models.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// Validation
	if req.Type == "" {
		return BadRequest(c, "Message type is required")
	}

	// Pour les messages texte, le contenu est requis
	if req.Type == "text" && (req.Content == nil || *req.Content == "") {
		return BadRequest(c, "Content is required for text messages")
	}

	// Pour les fichiers, l'URL est requise
	if (req.Type == "audio" || req.Type == "video" || req.Type == "image" || req.Type == "file") &&
		(req.FileURL == nil || *req.FileURL == "") {
		return BadRequest(c, "File URL is required for media messages")
	}

	message, err := h.chatSvc.SendMessage(c.Context(), conversationID, userID, &req)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return Created(c, message)
}

// GetMessages récupère les messages d'une conversation
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	messages, err := h.chatSvc.GetMessages(c.Context(), conversationID, userID, limit, offset)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	// Marquer les messages comme lus en arrière-plan avec timeout
	if len(messages) > 0 {
		lastMessageID := messages[0].ID
		// Utiliser un context avec timeout pour éviter les goroutines orphelines
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			h.chatSvc.MarkMessagesAsRead(ctx, conversationID, userID, &lastMessageID)
		}()
	}

	return OK(c, messages)
}

// EditMessage modifie un message
func (h *MessageHandler) EditMessage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	messageID, err := uuid.Parse(c.Params("message_id"))
	if err != nil {
		return BadRequest(c, "Invalid message ID")
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if req.Content == "" {
		return BadRequest(c, "Content is required")
	}

	message, err := h.chatSvc.EditMessage(c.Context(), messageID, userID, req.Content)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, message)
}

// DeleteMessage supprime un message
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	messageID, err := uuid.Parse(c.Params("message_id"))
	if err != nil {
		return BadRequest(c, "Invalid message ID")
	}

	if err := h.chatSvc.DeleteMessage(c.Context(), messageID, userID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Message deleted"})
}

// MarkAsRead marque les messages comme lus
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	var req models.MarkReadRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	if err := h.chatSvc.MarkMessagesAsRead(c.Context(), conversationID, userID, req.MessageID); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Messages marked as read"})
}

// UploadChatMedia handles file uploads for chat messages (images, videos, audio, files)
func (h *MessageHandler) UploadChatMedia(c *fiber.Ctx) error {
	h.logger.Info("[UploadChatMedia] Starting chat media upload")

	// Get user ID from JWT
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return Unauthorized(c)
	}

	// Parse conversation ID
	conversationID, err := uuid.Parse(c.Params("conversation_id"))
	if err != nil {
		return BadRequest(c, "Invalid conversation ID")
	}

	// Get media service from context
	mediaSvc, ok := c.Locals("mediaService").(services.MediaService)
	if !ok {
		h.logger.Error("[UploadChatMedia] Media service not available in context")
		return InternalError(c, "Media service not available")
	}

	// Parse multipart form (max 1 file, max 50MB for videos)
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("[UploadChatMedia] Failed to get file", zap.Error(err))
		return BadRequest(c, "No file provided")
	}

	// Validate file size (max 50MB)
	if file.Size > 50*1024*1024 {
		h.logger.Warn("[UploadChatMedia] File too large", zap.Int64("size", file.Size))
		return BadRequest(c, "File too large (max 50MB)")
	}

	// Get file type from form
	fileType := c.FormValue("type", "file") // image, video, audio, file

	// Validate file type by extension
	ext := filepath.Ext(file.Filename)
	allowedTypes := map[string]map[string]bool{
		"image": {".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true},
		"video": {".mp4": true, ".webm": true, ".mov": true},
		"audio": {".mp3": true, ".wav": true, ".ogg": true, ".m4a": true},
		"file":  {".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".zip": true},
	}

	if typeExts, ok := allowedTypes[fileType]; !ok || !typeExts[ext] {
		h.logger.Warn("[UploadChatMedia] Invalid file type", zap.String("type", fileType), zap.String("ext", ext))
		return BadRequest(c, "Invalid file type for "+fileType)
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		h.logger.Error("[UploadChatMedia] Failed to open file", zap.Error(err))
		return InternalError(c, "Failed to open file")
	}
	defer fileReader.Close()

	// Upload to R2: chats/{conversationID}/{userID}_{timestamp}_{filename}
	folder := fmt.Sprintf("chats/%s", conversationID.String())

	h.logger.Info("[UploadChatMedia] Uploading to R2",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()),
		zap.String("folder", folder),
		zap.String("filename", file.Filename))

	url, err := mediaSvc.UploadFile(c.Context(), fileReader, file, folder)
	if err != nil {
		h.logger.Error("[UploadChatMedia] R2 upload failed",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.String("conversation_id", conversationID.String()))
		return InternalError(c, "Failed to upload file to Cloudflare R2: "+err.Error())
	}

	h.logger.Info("[UploadChatMedia] R2 upload successful",
		zap.String("user_id", userID.String()),
		zap.String("conversation_id", conversationID.String()),
		zap.String("url", url))

	return OK(c, fiber.Map{
		"message": "File uploaded successfully",
		"url":     url,
		"type":    fileType,
		"size":    file.Size,
		"name":    file.Filename,
	})
}
