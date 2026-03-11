package prompts

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topboyasante/prompts/internal/auth"
	"github.com/topboyasante/prompts/internal/server"
)

var nameSlugRE = regexp.MustCompile(`^[a-z0-9-]+$`)

type Handler struct {
	repo Repository
}

type latestVersionFinder interface {
	FindLatestByPromptID(ctx context.Context, promptID string) (any, error)
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

type createPromptRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := auth.UserIDFromGin(c)
	if !ok {
		server.LoggerFromContext(c).Warn("missing authenticated user in prompt create")
		server.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing user in context")
		return
	}

	var req createPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		server.LoggerFromContext(c).WithError(err).Warn("invalid create prompt payload")
		server.RespondValidationError(c, map[string]string{"body": "invalid JSON payload"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if !nameSlugRE.MatchString(req.Name) {
		server.LoggerFromContext(c).WithField("name", req.Name).Warn("invalid prompt slug")
		server.RespondValidationError(c, map[string]string{"name": "must match ^[a-z0-9-]+$"})
		return
	}

	p, err := h.repo.Create(c.Request.Context(), &Prompt{
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		OwnerID:     userID,
	}, req.Tags)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("name", req.Name).Warn("prompt create conflict")
		server.RespondError(c, http.StatusConflict, "CONFLICT", "prompt already exists")
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"prompt_id": p.ID, "name": p.Name, "owner_id": p.OwnerID}).Info("prompt created")

	server.RespondJSON(c, http.StatusCreated, p)
}

func (h *Handler) Get(c *gin.Context) {
	owner := strings.TrimSpace(c.Param("owner"))
	name := strings.TrimSpace(c.Param("name"))
	prompt, err := h.repo.FindByOwnerAndName(c.Request.Context(), owner, name)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"owner": owner, "name": name}).Error("failed to load prompt")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load prompt")
		return
	}
	if prompt == nil {
		server.LoggerFromContext(c).WithFields(logrus.Fields{"owner": owner, "name": name}).Warn("prompt not found")
		server.RespondError(c, http.StatusNotFound, "NOT_FOUND", "prompt not found")
		return
	}

	server.RespondJSON(c, http.StatusOK, prompt)
}

func (h *Handler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	limit := parseInt(c.Query("limit"), 20)
	offset := parseInt(c.Query("offset"), 0)

	var (
		rows []Prompt
		err  error
	)
	if query == "" {
		rows, err = h.repo.List(c.Request.Context(), limit, offset)
	} else {
		rows, err = h.repo.Search(c.Request.Context(), query, limit, offset)
	}
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"q": query, "limit": limit, "offset": offset}).Error("prompt search failed")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "search failed")
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"q": query, "limit": limit, "offset": offset, "count": len(rows)}).Info("prompt search completed")

	server.RespondJSON(c, http.StatusOK, gin.H{
		"items":  rows,
		"query":  query,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := auth.UserIDFromGin(c)
	if !ok {
		server.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing user in context")
		return
	}

	id := c.Param("id")
	prompt, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("prompt_id", id).Error("failed to load prompt")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load prompt")
		return
	}
	if prompt == nil {
		server.RespondError(c, http.StatusNotFound, "NOT_FOUND", "prompt not found")
		return
	}
	if prompt.OwnerID != userID {
		server.RespondError(c, http.StatusForbidden, "FORBIDDEN", "not the prompt owner")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("prompt_id", id).Error("failed to delete prompt")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete prompt")
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"prompt_id": id, "owner_id": userID}).Info("prompt deleted")

	c.Status(http.StatusNoContent)
}

func parseInt(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
