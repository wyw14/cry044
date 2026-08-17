package httptransport

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry044/internal/application"
	"net/http"
	"strings"
)

type Services struct {
	Reviews       *application.ReviewService
	Deliberations *application.DeliberationService
	Transfers     *application.TransferService
	Ready         func() error
}
type Handler struct {
	services Services
	rules    *validator.Validate
}
type reviewOperation struct {
	method string
	path   string
	roles  []string
	handle func(*reviewRequest)
}

func New(services Services, middleware ...gin.HandlerFunc) *gin.Engine {
	hub := &Handler{services: services, rules: validator.New()}
	engine := gin.New()
	engine.Use(middleware...)
	engine.GET("/healthz", func(c *gin.Context) { c.JSON(200, map[string]string{"status": "ok", "service": "review-standard"}) })
	engine.GET("/readyz", hub.readiness)
	hub.mountStandards(engine.Group("/api/v1"))
	return engine
}

func (h *Handler) readiness(c *gin.Context) {
	if h.services.Ready != nil {
		if err := h.services.Ready(); err != nil {
			writeProblem(c, 503, "DATABASE_NOT_READY", err, nil)
			return
		}
	}
	c.JSON(200, map[string]string{"status": "ready"})
}

func (h *Handler) mountStandards(api *gin.RouterGroup) {
	operations := []reviewOperation{
		{http.MethodPost, "/standards/:id/versions/:version/clones", []string{"maintainer"}, (*reviewRequest).cloneTemplate},
		{http.MethodPost, "/standards/:id/versions/:version/publications", []string{"template_approver"}, (*reviewRequest).publishTemplate},
		{http.MethodGet, "/standards/:id/versions/:version/exports", []string{"maintainer", "template_approver"}, (*reviewRequest).exportTemplate},
		{http.MethodPost, "/review-batches/:id/starts", []string{"coordinator"}, (*reviewRequest).startBatch},
		{http.MethodPost, "/review-batches/:id/reviews", []string{"reviewer"}, (*reviewRequest).submitReview},
		{http.MethodGet, "/review-batches/:id/deliberation", []string{"coordinator", "reviewer"}, (*reviewRequest).deliberationBoard},
		{http.MethodPost, "/review-batches/:id/materials/:material_id/final-decisions", []string{"coordinator"}, (*reviewRequest).resolveMaterial},
		{http.MethodPost, "/review-batches/:id/completions", []string{"coordinator"}, (*reviewRequest).completeBatch},
		{http.MethodGet, "/review-batches/:id/printable-sheet", []string{"coordinator", "auditor"}, (*reviewRequest).printableSheet},
		{http.MethodPost, "/template-imports", []string{"maintainer"}, (*reviewRequest).importTemplate},
		{http.MethodGet, "/review-statistics", []string{"maintainer", "auditor"}, (*reviewRequest).statistics},
	}
	for _, operation := range operations {
		api.Handle(operation.method, operation.path, h.anyOf(operation.roles...), h.endpoint(operation.handle))
	}
}

func (h *Handler) endpoint(action func(*reviewRequest)) gin.HandlerFunc {
	return func(c *gin.Context) { action(&reviewRequest{context: c, services: h.services, rules: h.rules}) }
}

func (h *Handler) anyOf(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[strings.TrimSpace(c.GetHeader("X-Actor-Role"))]; !ok {
			writeProblem(c, 403, "ROLE_DENIED", errors.New("actor role is not permitted for this review operation"), nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
