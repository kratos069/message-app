package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
)

// ListUsers returns all users (admin only)
func (server *Server) listUsers(ctx *gin.Context) {
	limit := ctx.DefaultQuery("limit", "50")
	offset := ctx.DefaultQuery("offset", "0")

	users, err := server.store.GetAllUsers(
		ctx, db.GetAllUsersParams{
			Limit:  parseInt32(limit, 50),
			Offset: parseInt32(offset, 0),
		})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users":   users,
		"count":   len(users),
		"message": "Users retrieved successfully",
	})
}

type userIDStruct struct {
	UserID int64 `uri:"user_id" binding:"required,min=1"`
}

// GetUser returns user details (admin only)
func (server *Server) getUser(ctx *gin.Context) {
	var req userIDStruct
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	user, err := server.store.GetUserByID(ctx, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "User retrieved successfully",
	})
}

// BanUserRequest represents ban user request
type banUserRequest struct {
	Reason string `json:"reason" binding:"required,min=1"`
}

// BanUser bans a user (admin only)
func (server *Server) banUser(ctx *gin.Context) {
	var reqUser userIDStruct
	if err := ctx.ShouldBindUri(&reqUser); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	userBefore, err := server.store.GetUserByID(ctx, reqUser.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, "User not found")
		return
	}

	var req banUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, "wrong input")
		return
	}

	// Note: You'll need to create this query
	err = server.store.BanUser(ctx, db.BanUserParams{
		ID: userBefore.ID,
		BannedReason: pgtype.Text{
			String: req.Reason,
			Valid:  true,
		},
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			"Failed to ban user")
		return
	}

	userAfter, err := server.store.GetUserByID(ctx, reqUser.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, "User not found")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":    userAfter,
		"message": "User banned successfully",
	})
}

// UnbanUser unbans a user (admin only)
func (server *Server) unbanUser(ctx *gin.Context) {
	var reqUser userIDStruct
	if err := ctx.ShouldBindUri(&reqUser); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	userBefore, err := server.store.GetUserByID(ctx, reqUser.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, "User not found")
		return
	}

	err = server.store.UnbanUser(ctx, userBefore.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			"Failed to unban user")
		return
	}

	userAfter, err := server.store.GetUserByID(ctx, reqUser.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, "User not found")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user":    userAfter,
		"message": "User un-banned successfully",
	})
}

// GetStats returns system statistics (admin only)
func (server *Server) getStats(ctx *gin.Context) {
	totalUsers, err := server.store.GetTotalUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get total users"})
		return
	}

	totalConversations, err := server.store.GetTotalConversations(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get total conversations"})
		return
	}

	totalMessages, err := server.store.GetTotalMessages(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get total messages"})
		return
	}

	onlineUsers, err := server.store.GetOnlineUsersCount(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to get online users"})
		return
	}

	stats := gin.H{
		"total_users":         totalUsers,
		"total_conversations": totalConversations,
		"total_messages":      totalMessages,
		"online_users":        onlineUsers,
	}

	ctx.JSON(http.StatusOK, stats)
}
