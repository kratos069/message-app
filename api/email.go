package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/worker"
	"github.com/rs/zerolog/log"
)

// VerifyEmailRequest represents the email verification request
type VerifyEmailRequest struct {
	EmailID    int64  `form:"email_id" binding:"required,min=1"`
	SecretCode string `form:"secret_code" binding:"required,len=10"`
}

// VerifyEmailResponse represents the response
type VerifyEmailResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	IsVerified bool   `json:"is_verified"`
	Username   string `json:"username,omitempty"`
	Email      string `json:"email,omitempty"`
}

// VerifyEmail handles email verification
func (server *Server) VerifyEmail(ctx *gin.Context) {
	var req VerifyEmailRequest

	// Bind query parameters
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, VerifyEmailResponse{
			Success: false,
			Message: "Invalid verification link. Please check your email.",
		})
		return
	}

	// Execute verification transaction
	result, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailID:    req.EmailID,
		SecretCode: req.SecretCode,
	})

	if err != nil {
		log.Error().Err(err).
			Int64("email_id", req.EmailID).
			Msg("Failed to verify email")

		// Check for specific errors
		if err.Error() == "verification code expired" {
			ctx.JSON(http.StatusBadRequest, VerifyEmailResponse{
				Success: false,
				Message: "Verification link has expired. Please request a new one.",
			})
			return
		}

		if err.Error() == "verification code already used" {
			ctx.JSON(http.StatusBadRequest, VerifyEmailResponse{
				Success: false,
				Message: "This verification link has already been used.",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, VerifyEmailResponse{
			Success: false,
			Message: "Failed to verify email. Please try again.",
		})
		return
	}

	// Success response
	ctx.JSON(http.StatusOK, VerifyEmailResponse{
		Success:    true,
		Message:    "Email verified successfully!",
		IsVerified: true,
		Username:   result.User.Username,
		Email:      result.User.Email,
	})

	log.Info().
		Str("username", result.User.Username).
		Str("email", result.User.Email).
		Msg("Email verified successfully")
}

// ResendVerificationEmailRequest represents resend request
type ResendVerificationEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResendVerificationEmail resends verification email
func (server *Server) ResendVerificationEmail(ctx *gin.Context) {
	var req ResendVerificationEmailRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid email address",
		})
		return
	}

	// Get user by email
	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("User not found")
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "No account found with this email",
		})
		return
	}

	// Check if already verified
	if user.IsEmailVerified {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email is already verified",
		})
		return
	}

	// Distribute task to send verification email
	// Note: You need to pass the task distributor to the server
	err = server.taskDistributor.DistributeTaskSendBerifyEmail(
		ctx,
		&worker.PayloadSendVerifyEmail{
			Username: user.Username,
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to distribute verification email task")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to send verification email",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification email sent. Please check your inbox.",
	})
}
