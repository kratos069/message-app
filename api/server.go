package api

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/token"
	"github.com/kratos069/message-app/util"
	"github.com/kratos069/message-app/worker"
	"github.com/rs/zerolog/log"
)

// servers HTTP requests
type Server struct {
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	router          *gin.Engine
	server          *http.Server
	taskDistributor worker.TaskDistributor
}

// Creates HTTP server and Setup Routing
func NewServer(config util.Config, store db.Store,
	taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	// ============================Simulation============================
	// ============================for Production========================

	// Limit CPU usage for realistic production simulation
	maxProcs := config.MaxProcs
	if maxProcs == 0 {
		maxProcs = 4 // Default to 4 cores for production-like testing
	}
	runtime.GOMAXPROCS(maxProcs)
	log.Info().Int("max_procs", maxProcs).Msg("Set CPU limit for production simulation")

	// ============================Simulation============================

	server := &Server{
		config:          config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}

	// Routes
	server.setupRoutes()

	// HTTP server with timeouts
	server.server = &http.Server{
		Addr:           config.HTTPServerAddress,
		Handler:        server.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return server, nil
}

func (server *Server) setupRoutes() {
	router := gin.Default()

	// Track active requests
	router.Use(ActiveRequestsMiddleware())

	// Health check endpoint
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"success": true,
			"message": "Server is healthy",
		})
	})

	// routes

	// public
	router.POST("/register", server.register)
	router.POST("/login", server.loginUser)

	// Email verification (public - accessed via email link)
	router.GET("/verify_email", server.VerifyEmail)
	router.POST("/resend_verification", server.ResendVerificationEmail)

	// for both users and admins
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker,
		[]string{util.AdminRole, util.CustomerRole}))
	authRoutes.POST("/logout", server.logoutUser)
	authRoutes.GET("/users/:id", server.getUserByID)
	authRoutes.POST("/users/search", server.SearchUsers)

	authRoutes.GET("/conversations", server.listConversations)
	authRoutes.POST(
		"/conversations/:other_user_id",
		server.GetOrCreateDirectConversation)
	authRoutes.GET("/conversations/:id", server.getConversation)
	authRoutes.GET("/debug/:conversation_id", server.debugConversation)

	authRoutes.GET("/messages/:conversation_id", server.getMessages)
	authRoutes.POST("/messages/:conversation_id", server.sendMessage)

	// for only admins
	adminRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker,
		[]string{util.AdminRole}))
	adminRoutes.GET("/admin", server.listUsers)
	adminRoutes.GET("/admin/:user_id", server.getUser)
	adminRoutes.POST("/admin/ban/:user_id", server.banUser)
	adminRoutes.POST("/admin/unban/:user_id", server.unbanUser)
	adminRoutes.GET("/admin/stats", server.getStats)

	server.router = router
}

// Starts and runs HTTP server on a specific address
func (server *Server) Start(address string) error {
	server.server.Addr = address
	return server.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (server *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Initiating graceful shutdown...")

	// Set the server to not accept new connections
	if err := server.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info().Msg("All active connections closed")
	return nil
}

func errResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func parseInt32(s string, defaultVal int32) int32 {
	var val int32
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return defaultVal
	}
	return val
}
