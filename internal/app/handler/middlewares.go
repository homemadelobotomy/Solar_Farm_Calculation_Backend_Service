package handler

import (
	"errors"
	"lab/internal/app/ds"
	"lab/internal/app/role"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
)

const (
	jwtPrefix   = "Bearer "
	userIdKey   = "userId"
	userRoleKey = "userRole"
)

func (h *Handler) WithAuthCheck(roles ...role.Role) func(ctx *gin.Context) {

	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")

		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]
		err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
		if err == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)

			return
		}
		if !errors.Is(err, redis.Nil) {
			ctx.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {

			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Set(userIdKey, claims.UserId)
		ctx.Set(userRoleKey, claims.IsModerator)

		if len(roles) > 0 {
			hasRole := false
			for _, oneRole := range roles {
				if claims.IsModerator == oneRole {
					hasRole = true
					break
				}
			}
			if !hasRole {
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		ctx.Next()
	}
}

func GetUserIdFromContext(ctx *gin.Context) (uint, bool) {
	userId, exists := ctx.Get(userIdKey)
	if !exists {
		return 0, false
	}
	id, ok := userId.(uint)
	return id, ok
}

func GetUserRoleFromContext(ctx *gin.Context) (role.Role, bool) {
	userRole, exists := ctx.Get(userRoleKey)
	if !exists {
		return role.User, false
	}
	role := userRole.(role.Role)
	return role, true
}

func (h *Handler) WithOptionalAuth() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")

		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			ctx.Next()
			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]

		err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
		if err == nil {
			ctx.Next()
			return
		}
		if !errors.Is(err, redis.Nil) {
			ctx.Next()
			return
		}
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.Token), nil
		})
		if err != nil {
			ctx.Next()
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			ctx.Next()
			return
		}
		ctx.Set(userIdKey, claims.UserId)
		ctx.Set(userRoleKey, claims.IsModerator)

		ctx.Next()
	}
}
