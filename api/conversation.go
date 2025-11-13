package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/token"
)

// ListConversations returns user's conversations
func (server *Server) listConversations(ctx *gin.Context) {
	// Get user info from auth middleware
	parsedUser := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Get pagination params
	limit := ctx.DefaultQuery("limit", "20")
	offset := ctx.DefaultQuery("offset", "0")

	conversations, err := server.store.GetUserConversationsWithLastMessage(
		ctx, db.GetUserConversationsWithLastMessageParams{
			UserID: parsedUser.UserID,
			Limit:  parseInt32(limit, 20),
			Offset: parseInt32(offset, 0),
		})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"count":         len(conversations),
		"message":       "Conversations retrieved successfully",
	})
}

type otherUserIDStruct struct {
	OtherUserID int64 `uri:"other_user_id" binding:"required,min=1"`
}

// GetOrCreateDirectConversation gets or creates a direct conversation
func (server *Server) GetOrCreateDirectConversation(ctx *gin.Context) {
	// Get user info from auth middleware
	parsedUser := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	var req otherUserIDStruct
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	// Can't create conversation with yourself
	if parsedUser.UserID == req.OtherUserID {
		ctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Cannot create conversation with yourself"})
		return
	}

	// Verify other user exists
	otherUser, err := server.store.GetUserByID(ctx, req.OtherUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound,
				gin.H{"error": "Other user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	// Get or create conversation between current user and other user
	result, err := server.store.GetOrCreateDirectConversationTx(
		ctx, db.GetOrCreateDirectConversationTxParams{
			User1ID: parsedUser.UserID, // Always use current authenticated user
			User2ID: otherUser.ID,
		})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"conversation_id": result.Conversation.ConversationsID,
		"is_new":          result.IsNew,
		"message":         "Conversation ready",
	})
}

// GetConversation returns conversation details
func (server *Server) getConversation(ctx *gin.Context) {
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	var req conversationIDStruct
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	// Check if user is participant
	isParticipant, err := server.store.IsUserInConversation(
		ctx, db.IsUserInConversationParams{
			UserID:         payload.UserID,
			ConversationID: req.ConversationID,
		})
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	if !isParticipant {
		ctx.JSON(http.StatusForbidden,
			gin.H{"error": "You are not a participant"})
		return
	}

	// Get conversation details
	participants, err := server.store.GetConversationWithParticipants(ctx, req.ConversationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"conversation_id": req.ConversationID,
		"participants":    participants,
		"message":         "Conversation retrieved successfully",
	})
}

// participants are added
// but the on who creates is not a participant!
func (server *Server) debugConversation(ctx *gin.Context) {
	var req conversationIDStruct
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	// Get all participants
	participants, err := server.store.GetConversationParticipants(ctx, req.ConversationID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"conversation_id": req.ConversationID,
		"participants":    participants,
		"count":           len(participants),
	})
}
