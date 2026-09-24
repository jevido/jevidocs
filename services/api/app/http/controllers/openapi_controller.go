package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/store"
)

// OpenAPIController generates API reference pages from OpenAPI documents.
type OpenAPIController struct{}

func NewOpenAPIController() *OpenAPIController { return &OpenAPIController{} }

// Import takes `{spec, prefix?}` (spec as JSON or YAML text) and syncs the
// generated pages into the project under prefix.
func (r *OpenAPIController) Import(ctx http.Context) http.Response {
	p, err := store.FindProject(ctx.Request().Route("project"), true)
	if err != nil {
		return fail(ctx, err)
	}
	var in struct {
		Spec   string `json:"spec"`
		Prefix string `json:"prefix"`
	}
	if err := ctx.Request().Bind(&in); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{"error": "invalid body"})
	}
	if in.Spec == "" {
		return fail(ctx, store.ValidationError{Msg: "spec is required"})
	}
	res, err := store.ImportOpenAPI(p, []byte(in.Spec), in.Prefix)
	if err != nil {
		return fail(ctx, err)
	}
	return ok(ctx, res)
}
