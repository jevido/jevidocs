package middleware

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/store"
)

const (
	userKey  = "jevidocs.user"
	tokenKey = "jevidocs.token"
)

// BearerToken reads `Authorization: Bearer <token>`.
func BearerToken(ctx http.Context) string {
	h := ctx.Request().Header("Authorization", "")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// Auth rejects requests without a valid token and puts the user on the
// context for controllers (see User).
func Auth() http.Middleware { return authMiddleware{} }

type authMiddleware struct{}

func (authMiddleware) Signature() string { return "auth" }

func (authMiddleware) Handle(ctx http.Context) {
	{
		u, t, err := store.UserForToken(BearerToken(ctx))
		if err != nil {
			ctx.Request().AbortWithStatusJson(http.StatusUnauthorized, http.Json{"error": "unauthenticated"})
			return
		}
		ctx.WithValue(userKey, u)
		ctx.WithValue(tokenKey, t)
		ctx.Request().Next()
	}
}

// User is the authenticated user set by Auth.
func User(ctx http.Context) models.User {
	u, _ := ctx.Value(userKey).(models.User)
	return u
}

// Token is the token the request authenticated with.
func Token(ctx http.Context) models.ApiToken {
	t, _ := ctx.Value(tokenKey).(models.ApiToken)
	return t
}
