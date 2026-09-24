package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support"

	"dev.jevido/jevidocs/services/api/app/facades"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// Show reports that the API is up and can reach its database. Deploy health
// checks hit this, so a container that cannot talk to Postgres never goes
// live.
func (r *HealthController) Show(ctx http.Context) http.Response {
	database := "ok"
	if err := pingDatabase(ctx); err != nil {
		// The reason stays in the log; the public answer does not describe
		// the network behind the API.
		facades.Log().Errorf("health: database unreachable: %v", err)
		database = "unreachable"
	}

	status, code := "ok", http.StatusOK
	if database != "ok" {
		status, code = "degraded", http.StatusServiceUnavailable
	}

	return ctx.Response().Json(code, http.Json{
		"status":    status,
		"database":  database,
		"framework": support.Version,
	})
}

func pingDatabase(parent context.Context) (err error) {
	// With no usable database configuration Goravel leaves the ORM unset and
	// calling it panics; that is still just "unreachable" to a health check.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("database not configured: %v", r)
		}
	}()
	db, err := facades.Orm().DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}
