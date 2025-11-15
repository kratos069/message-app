package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/token"
	"github.com/kratos069/message-app/util"
)

// servers HTTP requests for the insta-app
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// Creates HTTP server and Setup Routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	// Routes
	server.setupRoutes()

	return server, nil
}

func (server *Server) setupRoutes() {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"success": true,
			"message": "Server is healthy",
		})
	})

	// routes
	router.POST("/register", server.register)
	router.POST("/login", server.loginUser)

	// router.POST("/tokens/renew_access", server.renewAccessToken)

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

	authRoutes.GET("/messages/:id", server.getMessages)
	authRoutes.POST("/messages/:id", server.sendMessage)

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
	return server.router.Run(address)
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
