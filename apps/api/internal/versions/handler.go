package versions

import (
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topboyasante/prompts/internal/auth"
	"github.com/topboyasante/prompts/internal/prompts"
	"github.com/topboyasante/prompts/internal/server"
	"github.com/topboyasante/prompts/internal/storage"
)

var semverRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type Handler struct {
	repo       Repository
	prompts    prompts.Repository
	storage    storage.Client
	maxBodyLen int64
}

func NewHandler(repo Repository, promptsRepo prompts.Repository, storageClient storage.Client) *Handler {
	return &Handler{repo: repo, prompts: promptsRepo, storage: storageClient, maxBodyLen: 10 << 20}
}

func (h *Handler) Upload(c *gin.Context) {
	userID, ok := auth.UserIDFromGin(c)
	if !ok {
		server.LoggerFromContext(c).Warn("missing authenticated user in upload version")
		server.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing user in context")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBodyLen)
	promptID := strings.TrimSpace(c.Param("id"))
	if promptID == "" {
		server.LoggerFromContext(c).Warn("missing prompt id in upload version")
		server.RespondValidationError(c, map[string]string{"id": "prompt id is required"})
		return
	}

	pr, err := h.prompts.FindByID(c.Request.Context(), promptID)
	if err != nil || pr == nil {
		server.LoggerFromContext(c).WithError(err).WithField("prompt_id", promptID).Warn("prompt not found for version upload")
		server.RespondError(c, http.StatusNotFound, "NOT_FOUND", "prompt not found")
		return
	}
	if pr.OwnerID != userID {
		server.LoggerFromContext(c).WithFields(logrus.Fields{"prompt_id": promptID, "owner_id": pr.OwnerID, "user_id": userID}).Warn("unauthorized version upload attempt")
		server.RespondError(c, http.StatusForbidden, "UNAUTHORIZED", "not prompt owner")
		return
	}

	version := strings.TrimSpace(c.PostForm("version"))
	if !semverRE.MatchString(version) {
		server.LoggerFromContext(c).WithField("version", version).Warn("invalid semver for upload")
		server.RespondValidationError(c, map[string]string{"version": "must match ^\\d+\\.\\d+\\.\\d+$"})
		return
	}

	if existing, _ := h.repo.FindByVersion(c.Request.Context(), promptID, version); existing != nil {
		server.LoggerFromContext(c).WithFields(logrus.Fields{"prompt_id": promptID, "version": version}).Warn("version already exists")
		server.RespondError(c, http.StatusConflict, "CONFLICT", "version already exists")
		return
	}

	file, header, err := c.Request.FormFile("tarball")
	if err != nil {
		server.LoggerFromContext(c).WithError(err).Warn("missing tarball form file")
		server.RespondValidationError(c, map[string]string{"tarball": "multipart tarball file is required"})
		return
	}
	defer file.Close()

	key := path.Join(pr.OwnerID, pr.Name, version+".tar.gz")
	if err := h.storage.Upload(c.Request.Context(), key, file, header.Size, "application/gzip"); err != nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"prompt_id": promptID, "version": version, "s3_key": key, "size": header.Size}).Error("tarball upload failed")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to upload tarball")
		return
	}

	created, err := h.repo.Create(c.Request.Context(), &PromptVersion{PromptID: promptID, Version: version, TarballURL: key})
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"prompt_id": promptID, "version": version}).Error("failed to persist prompt version")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist version")
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"prompt_id": promptID, "version": version, "version_id": created.ID}).Info("prompt version uploaded")

	server.RespondJSON(c, http.StatusCreated, created)
}

func (h *Handler) List(c *gin.Context) {
	owner := strings.TrimSpace(c.Param("owner"))
	name := strings.TrimSpace(c.Param("name"))

	pr, err := h.prompts.FindByOwnerAndName(c.Request.Context(), owner, name)
	if err != nil || pr == nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"owner": owner, "name": name}).Warn("prompt not found for versions list")
		server.RespondError(c, http.StatusNotFound, "NOT_FOUND", "prompt not found")
		return
	}

	rows, err := h.repo.FindByPromptID(c.Request.Context(), pr.ID)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"owner": owner, "name": name, "prompt_id": pr.ID}).Error("failed to fetch versions")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch versions")
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"owner": owner, "name": name, "count": len(rows)}).Info("versions listed")

	server.RespondJSON(c, http.StatusOK, gin.H{"items": rows})
}

func (h *Handler) Download(c *gin.Context) {
	owner := strings.TrimSpace(c.Param("owner"))
	name := strings.TrimSpace(c.Param("name"))
	version := strings.TrimSpace(c.Param("version"))

	pr, err := h.prompts.FindByOwnerAndName(c.Request.Context(), owner, name)
	if err != nil || pr == nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"owner": owner, "name": name}).Warn("prompt not found for download")
		server.RespondError(c, http.StatusNotFound, "NOT_FOUND", "prompt not found")
		return
	}

	pv, err := h.repo.FindByVersion(c.Request.Context(), pr.ID, version)
	if err != nil || pv == nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"prompt_id": pr.ID, "version": version}).Warn("prompt version not found for download")
		server.RespondError(c, http.StatusNotFound, "NOT_FOUND", "version not found")
		return
	}

	_ = h.repo.RecordDownload(c.Request.Context(), pr.ID, pv.ID)
	presigned, err := h.storage.GetPresignedURL(c.Request.Context(), pv.TarballURL, 15*time.Minute)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithFields(logrus.Fields{"prompt_id": pr.ID, "version": version, "s3_key": pv.TarballURL}).Error("failed to create presigned url")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate download url")
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"prompt_id": pr.ID, "version": version}).Info("download redirect generated")

	c.Redirect(http.StatusFound, presigned)
}
