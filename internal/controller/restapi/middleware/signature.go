package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gct/config"
	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/usecase"

	"github.com/gin-gonic/gin"
)

// SignatureMiddleware validates request integrity using a SHA256 signature.
// It uses API keys stored in the database for verification.
// Requirements:
// X-Time-Unix: 10 digit unix timestamp
// X-Request-ID: Unique ID for the request
// X-API-KEY: The API key (used as SignKey)
// X-Sign: SHA256 hash of (X-Time-Unix + raw_api_key + X-Request-ID)
func SignatureMiddleware(cfg *config.Config, uc *usecase.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip signature verification for non-API routes (admin panel, static assets, docs, health)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/docs") ||
			strings.HasPrefix(path, "/health") ||
			path == "/" {
			c.Next()
			return
		}

		// 1. Get and Validate Headers
		timeUnixStr := c.GetHeader(consts.HeaderXTimeUnix)
		if timeUnixStr == consts.EmptyString {
			response.RespondWithError(c, errors.New(consts.MsgSignTimeEmpty), http.StatusUnauthorized)
			c.Abort()
			return
		}

		requestID := c.GetHeader(consts.HeaderXRequestID)
		if requestID == consts.EmptyString {
			response.RespondWithError(c, errors.New(consts.MsgSignRequestIDEmpty), http.StatusUnauthorized)
			c.Abort()
			return
		}

		apiKey := c.GetHeader(consts.HeaderXAPIKey)
		if apiKey == consts.EmptyString {
			// Fallback to query param
			apiKey = c.Query(consts.ParamAPIKey)
		}

		if apiKey == consts.EmptyString {
			response.RespondWithError(c, errors.New(consts.MsgMissingAPIKey), http.StatusUnauthorized)
			c.Abort()
			return
		}

		sign := c.GetHeader(consts.HeaderXSign)
		if sign == consts.EmptyString {
			response.RespondWithError(c, errors.New(consts.MsgSignEmpty), http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 2. Validate Timestamp
		timeUnix, err := strconv.ParseInt(timeUnixStr, 10, 64)
		if err != nil {
			response.RespondWithError(c, errors.New(consts.MsgSignInvalidTime), http.StatusUnauthorized)
			c.Abort()
			return
		}

		now := time.Now().Unix()
		expireTime := cfg.APIKeys.SignExpireTime
		if expireTime == 0 {
			expireTime = 10 // Default
		}

		// Return 499 for expired time
		if timeUnix < (now - expireTime) {
			response.RespondWithError(c, errors.New(consts.MsgSignTimeExpired), 499)
			c.Abort()
			return
		}

		// Future-dated requests prevention
		if timeUnix > (now + expireTime) {
			response.RespondWithError(c, errors.New(consts.MsgSignInvalidTime), http.StatusUnauthorized)
			c.Abort()
			return
		}

		// 3. Database Check: Validate API Key
		// This also ensures the key is active and not expired in DB
		integrationKey, err := uc.Integration.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			// If key is invalid in DB, the signature will also be invalid/unauthorized
			response.RespondWithError(c, errors.New(consts.MsgSignInvalid), 498)
			c.Abort()
			return
		}

		// 4. Signature Verification
		// Logic: X-Sign = SHA256(X-Time-Unix + raw_api_key + X-Request-Id)
		data := timeUnixStr + apiKey + requestID
		hash := sha256.Sum256([]byte(data))
		expectedSign := hex.EncodeToString(hash[:])

		if sign != expectedSign {
			response.RespondWithError(c, errors.New(consts.MsgSignInvalid), 498)
			c.Abort()
			return
		}

		// 5. Success: Store identity in context
		c.Set(consts.CtxIntegrationID, integrationKey.IntegrationID)
		c.Set(consts.CtxAPIKeyID, integrationKey.ID)

		c.Next()
	}
}
