package httptransport

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry044/internal/domain"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

type reviewRequest struct {
	context  *gin.Context
	services Services
	rules    *validator.Validate
}
type cloneCommand struct {
	NewID string `json:"new_id" validate:"required"`
}
type startCommand struct {
	TemplateID string            `json:"template_id" validate:"required"`
	Version    int               `json:"version" validate:"min=1"`
	Variables  map[string]string `json:"variables"`
}
type reviewCommand struct {
	MaterialID string         `json:"material_id" validate:"required"`
	Opinion    string         `json:"opinion" validate:"required"`
	Scores     []domain.Score `json:"scores" validate:"min=1"`
}
type decisionCommand struct {
	ExpectedRevision int64  `json:"expected_revision" validate:"min=1"`
	Passed           bool   `json:"passed"`
	Conclusion       string `json:"conclusion" validate:"required"`
	Rationale        string `json:"rationale"`
}

func (r *reviewRequest) cloneTemplate() {
	version, ok := r.version()
	if !ok {
		return
	}
	command := cloneCommand{}
	if !r.decode(&command) {
		return
	}
	value, err := r.services.Reviews.CloneTemplate(r.context, r.context.Param("id"), version, command.NewID, r.actor())
	r.reply(201, value, err)
}

func (r *reviewRequest) publishTemplate() {
	version, ok := r.version()
	if !ok {
		return
	}
	revision, err := strconv.ParseInt(strings.TrimSpace(r.context.GetHeader("If-Match")), 10, 64)
	if err != nil || revision < 1 {
		r.reject(422, "REVISION_REQUIRED", errors.New("If-Match must contain a positive revision"), map[string]string{"If-Match": "positive integer required"})
		return
	}
	err = r.services.Reviews.PublishTemplate(r.context, r.context.Param("id"), version, revision, r.actor())
	r.reply(200, map[string]bool{"published": err == nil}, err)
}

func (r *reviewRequest) startBatch() {
	command := startCommand{}
	if !r.decode(&command) {
		return
	}
	err := r.services.Reviews.StartBatch(r.context, r.context.Param("id"), command.TemplateID, command.Version, command.Variables, r.actor())
	r.reply(200, map[string]bool{"started": err == nil}, err)
}
func (r *reviewRequest) submitReview() {
	command := reviewCommand{}
	if !r.decode(&command) {
		return
	}
	err := r.services.Reviews.SubmitReview(r.context, r.context.Param("id"), command.MaterialID, r.actor(), command.Scores, command.Opinion)
	r.reply(201, map[string]bool{"submitted": err == nil}, err)
}

func (r *reviewRequest) deliberationBoard() {
	items, err := r.services.Deliberations.Board(r.context, r.context.Param("id"))
	if err != nil {
		r.reply(200, nil, err)
		return
	}
	page, size, ok := r.pagination()
	if !ok {
		return
	}
	order := r.context.DefaultQuery("sort", "material_id")
	if order != "material_id" && order != "score" && order != "reviewer_count" {
		r.reject(422, "SORT_NOT_ALLOWED", errors.New("unsupported deliberation ordering"), map[string]string{"sort": "material_id|score|reviewer_count"})
		return
	}
	visible := make([]domain.DeliberationCase, 0, len(items))
	for _, item := range items {
		if r.context.Query("vetoed") != "true" || item.Preliminary.Vetoed {
			visible = append(visible, item)
		}
	}
	sort.SliceStable(visible, func(i, j int) bool {
		if order == "score" {
			return visible[i].Preliminary.WeightedScore > visible[j].Preliminary.WeightedScore
		}
		if order == "reviewer_count" {
			return visible[i].SubmittedReviewers > visible[j].SubmittedReviewers
		}
		return visible[i].MaterialID < visible[j].MaterialID
	})
	start := min((page-1)*size, len(visible))
	end := min(start+size, len(visible))
	r.reply(200, map[string]any{"items": visible[start:end], "page": page, "page_size": size, "total": len(visible), "sort": order}, nil)
}

func (r *reviewRequest) resolveMaterial() {
	command := decisionCommand{}
	if !r.decode(&command) {
		return
	}
	value, err := r.services.Deliberations.Resolve(r.context, r.context.Param("id"), r.context.Param("material_id"), r.actor(), command.ExpectedRevision, command.Passed, command.Conclusion, command.Rationale)
	r.reply(200, value, err)
}
func (r *reviewRequest) completeBatch() {
	value, err := r.services.Reviews.Complete(r.context, r.context.Param("id"), r.actor())
	r.reply(200, map[string]any{"results": value}, err)
}

func (r *reviewRequest) importTemplate() {
	document, err := io.ReadAll(io.LimitReader(r.context.Request.Body, 2<<20+1))
	if err != nil || len(document) > 2<<20 {
		r.reject(413, "IMPORT_TOO_LARGE", errors.New("template import exceeds 2 MiB"), nil)
		return
	}
	value, err := r.services.Transfers.ImportTemplate(r.context, document, r.context.GetHeader("Idempotency-Key"))
	r.reply(201, value, err)
}
func (r *reviewRequest) exportTemplate() {
	version, ok := r.version()
	if !ok {
		return
	}
	path, err := r.services.Transfers.ExportTemplate(r.context, r.context.Param("id"), version)
	r.reply(200, map[string]string{"local_path": path}, err)
}
func (r *reviewRequest) printableSheet() {
	path, err := r.services.Transfers.PrintReviewSheet(r.context, r.context.Param("id"))
	r.reply(200, map[string]string{"local_path": path}, err)
}

func (r *reviewRequest) statistics() {
	to := time.Now().UTC()
	from := to.AddDate(0, -1, 0)
	var err error
	if raw := strings.TrimSpace(r.context.Query("from")); raw != "" {
		from, err = time.Parse(time.DateOnly, raw)
		if err != nil {
			r.reject(422, "INVALID_FROM", err, nil)
			return
		}
	}
	if raw := strings.TrimSpace(r.context.Query("to")); raw != "" {
		var day time.Time
		day, err = time.Parse(time.DateOnly, raw)
		if err != nil {
			r.reject(422, "INVALID_TO", err, nil)
			return
		}
		to = day.Add(24 * time.Hour)
	}
	value, err := r.services.Transfers.Statistics(r.context, from, to)
	r.reply(200, value, err)
}

func (r *reviewRequest) actor() string {
	value := strings.TrimSpace(r.context.GetHeader("X-Actor-ID"))
	if value == "" {
		return "demo-reviewer"
	}
	return value
}
func (r *reviewRequest) version() (int, bool) {
	value, err := strconv.Atoi(r.context.Param("version"))
	if err != nil || value < 1 {
		r.reject(422, "INVALID_VERSION", errors.New("version must be a positive integer"), nil)
		return 0, false
	}
	return value, true
}
func (r *reviewRequest) decode(target any) bool {
	if err := r.context.ShouldBindJSON(target); err != nil {
		r.reject(400, "INVALID_JSON", err, nil)
		return false
	}
	if err := r.rules.Struct(target); err != nil {
		r.reject(422, "VALIDATION_FAILED", err, nil)
		return false
	}
	return true
}
func (r *reviewRequest) pagination() (int, int, bool) {
	page, size := 1, 20
	if raw := r.context.Query("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			r.reject(422, "INVALID_PAGE", errors.New("page must be positive"), nil)
			return 0, 0, false
		}
		page = value
	}
	if raw := r.context.Query("page_size"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			r.reject(422, "INVALID_PAGE_SIZE", errors.New("page_size must be between 1 and 100"), nil)
			return 0, 0, false
		}
		size = value
	}
	return page, size, true
}
func (r *reviewRequest) reply(status int, value any, err error) {
	if err != nil {
		r.reject(409, "REVIEW_WORKFLOW_CONFLICT", err, nil)
		return
	}
	r.context.JSON(status, map[string]any{"data": value, "request_id": r.context.GetString("request_id")})
}
func (r *reviewRequest) reject(status int, code string, err error, fields map[string]string) {
	writeProblem(r.context, status, code, err, fields)
}

func writeProblem(c *gin.Context, status int, code string, err error, fields map[string]string) {
	if fields == nil {
		fields = map[string]string{}
	}
	message := "review workflow request failed"
	if err != nil {
		message = err.Error()
	}
	c.JSON(status, map[string]any{"code": code, "message": message, "field_errors": fields, "request_id": c.GetString("request_id")})
}
