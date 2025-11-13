package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/token"
)

type conversationIDStruct struct {
	ConversationID int64 `uri:"id" binding:"required,min=1"`
}

// GetMessages returns messages for a conversation
func (server *Server) getMessages(ctx *gin.Context) {
	// Get user info from auth middleware
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	var req conversationIDStruct
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	// Check if user is participant
	isParticipant, err := server.store.IsUserInConversation(
		ctx, db.IsUserInConversationParams{
			ConversationID: req.ConversationID,
			UserID:         authPayload.UserID,
		})

	if err != nil || !isParticipant {
		ctx.JSON(http.StatusForbidden,
			gin.H{"error": "access denied"})
		return
	}

	// Get pagination params
	limit := ctx.DefaultQuery("limit", "50")
	offset := ctx.DefaultQuery("offset", "0")

	messages, err := server.store.GetConversationMessages(
		ctx, db.GetConversationMessagesParams{
			ConversationID: req.ConversationID,
			Limit:          parseInt32(limit, 50),
			Offset:         parseInt32(offset, 0),
		})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get messages"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
		"message":  "Messages retrieved successfully",
	})
}

// SendMessageRequest defines the expected JSON payload
type sendMessageRequest struct {
	EncryptedContent string `json:"encrypted_content" binding:"required"`
}

// SendMessage sends a new encrypted message in a conversation
func (server *Server) sendMessage(ctx *gin.Context) {
	// Extract authenticated user payload from context
	authPayload, ok := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if !ok {
		ctx.JSON(http.StatusUnauthorized,
			gin.H{"error": "unauthorized"})
		return
	}

	// Parse conversation ID from URI
	var uri struct {
		ConversationID int64 `uri:"id" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid conversation ID"})
		return
	}

	// Parse and validate message body
	var req sendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message payload"})
		return
	}

	// Verify that the user is a participant in this conversation
	isParticipant, err := server.store.IsUserInConversation(ctx, db.IsUserInConversationParams{
		ConversationID: uri.ConversationID,
		UserID:         authPayload.UserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify participant"})
		return
	}
	if !isParticipant {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you are not a participant in this conversation"})
		return
	}

	// Generate unique client_message_id for idempotency
	clientMessageID := uuid.NewString()

	// Send message transactionally
	result, err := server.store.SendMessageTx(ctx, db.SendMessageTxParams{
		ConversationID:   uri.ConversationID,
		SenderID:         authPayload.UserID,
		EncryptedContent: req.EncryptedContent,
		ClientMessageID:  &clientMessageID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	// Respond with message metadata
	ctx.JSON(http.StatusCreated, gin.H{
		"message_id": result.Message.MessagesID,
		"client_id":  clientMessageID,
		"sent_at":    result.Message.SentAt.Time,
		"content": gin.H{
			"conversation_id": result.Message.ConversationID,
			"sender_id":       result.Message.SenderID,
			"encrypted_data":  result.Message.EncryptedContent,
		},
		"status":  "sent",
		"message": "Message sent successfully",
	})
}
