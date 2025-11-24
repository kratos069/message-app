package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/kratos069/message-app/token"
	"github.com/rs/zerolog/log"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// returns auth middlware function
func authMiddleware(tokenMaker token.Maker, accessibleRoles []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}

		// first part is auth type (bearer, o-auth etc), 2nd is token
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}

		// first value of header is auth-type
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}

		// second value of header is token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errResponse(err))
			return
		}

		// store payload in key
		ctx.Set(authorizationPayloadKey, payload)

		if !hasPermissions(payload.Role, accessibleRoles) {
			err := fmt.Errorf("permission denied")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				errResponse(err))
			return
		}

		// forward the request to the next handler
		ctx.Next()
	}
}

func hasPermissions(userRole string, accessibleRoles []string) bool {
	for _, role := range accessibleRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

var (
	activeRequests int64
)

// ActiveRequestsMiddleware tracks number of active requests
func ActiveRequestsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment active requests
		count := atomic.AddInt64(&activeRequests, 1)
		log.Debug().Int64("active_requests", count).Msg("Request started")

		// Process request
		c.Next()

		// Decrement active requests
		count = atomic.AddInt64(&activeRequests, -1)
		log.Debug().Int64("active_requests", count).Msg("Request completed")
	}
}

// GetActiveRequests returns the current number of active requests
func GetActiveRequests() int64 {
	return atomic.LoadInt64(&activeRequests)
}
