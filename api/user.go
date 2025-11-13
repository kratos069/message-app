package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/token"
	"github.com/kratos069/message-app/util"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type userResponse struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// Register creates a new user account
func (server *Server) register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	// Create user
	user, err := server.store.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	resp := newUserResponse(user)
	ctx.JSON(http.StatusOK, resp)
}

type loginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var input loginUserRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	user, err := server.store.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	err = util.CheckPassword(input.Password, user.PasswordHash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}

	// creating access token
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.ID,
		user.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	// creating refresh token
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		user.ID,
		user.Role,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	// create session
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     accessPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiredAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	// Update online status
	_ = server.store.UpdateUserOnlineStatus(ctx, db.UpdateUserOnlineStatusParams{
		ID: user.ID,
		IsOnline: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	})

	resp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, resp)
}

// Logout logs out a user
func (server *Server) logoutUser(ctx *gin.Context) {
	// Get user info from auth middleware
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Find and block all active sessions for this user
	err := server.store.BlockUserSessions(ctx, authPayload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	// Update user online status to offline
	err = server.store.UpdateUserOnlineStatus(ctx, db.UpdateUserOnlineStatusParams{
		ID: authPayload.UserID,
		IsOnline: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

type inputUserID struct {
	UserID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getUserByID(ctx *gin.Context) {
	var input inputUserID

	if err := ctx.ShouldBindUri(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	user, err := server.store.GetUserByID(ctx, input.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if user.ID != authPayload.UserID {
		err := errors.New("account doesnot belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errResponse(err))
		return
	}

	resp := newUserResponse(user)

	ctx.JSON(http.StatusOK, resp)
}

// SearchUsers searches for users by username
func (server *Server) SearchUsers(ctx *gin.Context) {
	query := ctx.Query("username")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, "user not found")
		return
	}

	users, err := server.store.SearchUsersByUsername(ctx, db.SearchUsersByUsernameParams{
		Username: "%" + query + "%",
		Limit:    20,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users":   users,
		"count":   len(users),
		"message": "Search completed successfully",
	})
}
