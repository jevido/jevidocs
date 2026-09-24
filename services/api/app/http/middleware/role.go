package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// Role must run after Auth. It rejects requests the user's role may not
// make, as decided by store.RequiredRole from the method and path.
func Role() http.Middleware { return roleMiddleware{} }

type roleMiddleware struct{}

func (roleMiddleware) Signature() string { return "role" }

func (roleMiddleware) Handle(ctx http.Context) {
	need := store.RequiredRole(ctx.Request().Method(), ctx.Request().Path())
	if !store.HasRole(User(ctx), need) {
		ctx.Request().AbortWithStatusJson(http.StatusForbidden, http.Json{"error": "your role (" + store.RoleOf(User(ctx)) + ") cannot do this; it needs " + need})
		return
	}
	ctx.Request().Next()
}
