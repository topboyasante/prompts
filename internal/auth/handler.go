package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topboyasante/prompts/internal/config"
	"github.com/topboyasante/prompts/internal/server"
)

const oauthStateCookieName = "prompts_oauth_state"
const oauthCLICookieName = "prompts_oauth_cli"

type Handler struct {
	service *AuthService
	config  *config.Config
}

func NewHandler(service *AuthService, cfg *config.Config) *Handler {
	return &Handler{service: service, config: cfg}
}

func (h *Handler) Login(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if !h.service.IsProviderSupported(provider) {
		server.LoggerFromContext(c).WithField("provider", provider).Warn("unsupported oauth provider")
		server.RespondValidationError(c, map[string]string{"provider": "unsupported provider"})
		return
	}

	state := strings.TrimSpace(c.Query("state"))
	if state == "" {
		var err error
		state, err = randomState(32)
		if err != nil {
			server.LoggerFromContext(c).WithError(err).Error("failed to generate oauth state")
			server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate oauth state")
			return
		}
	}

	payload := provider + "|" + state
	signed := signState(payload, h.config.JWTSecret)
	c.SetCookie(oauthStateCookieName, signed, 300, "/", "", false, true)
	if c.Query("cli") == "true" {
		c.SetCookie(oauthCLICookieName, "true", 300, "/", "", false, true)
	} else {
		c.SetCookie(oauthCLICookieName, "false", 300, "/", "", false, true)
	}
	loginURL, err := h.service.GetLoginURL(provider, state)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("provider", provider).Error("failed to generate oauth login url")
		server.RespondValidationError(c, map[string]string{"provider": "unsupported provider"})
		return
	}
	server.LoggerFromContext(c).WithFields(logrus.Fields{"provider": provider, "cli": c.Query("cli") == "true"}).Info("oauth login redirect")
	c.Redirect(http.StatusFound, loginURL)
}

func (h *Handler) Callback(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if !h.service.IsProviderSupported(provider) {
		server.LoggerFromContext(c).WithField("provider", provider).Warn("unsupported oauth provider")
		server.RespondValidationError(c, map[string]string{"provider": "unsupported provider"})
		return
	}

	state := strings.TrimSpace(c.Query("state"))
	code := strings.TrimSpace(c.Query("code"))
	if state == "" || code == "" {
		server.LoggerFromContext(c).Warn("oauth callback missing state or code")
		server.RespondError(c, http.StatusBadRequest, "VALIDATION_FAILED", "missing oauth state or code")
		return
	}

	signedState, err := c.Cookie(oauthStateCookieName)
	payload := provider + "|" + state
	if err != nil || !verifyState(signedState, payload, h.config.JWTSecret) {
		server.LoggerFromContext(c).WithError(err).Warn("invalid oauth state")
		server.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid oauth state")
		return
	}

	token, err := h.service.ExchangeCode(c.Request.Context(), provider, code)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("provider", provider).Warn("oauth code exchange failed")
		server.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "oauth exchange failed")
		return
	}

	providerUser, err := h.service.GetProviderUser(c.Request.Context(), provider, token.AccessToken)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("provider", provider).Warn("failed to fetch provider user")
		server.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "failed to fetch provider user")
		return
	}

	user, err := h.service.UpsertUser(c.Request.Context(), providerUser)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("provider", provider).Error("failed to upsert user from provider")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to upsert user")
		return
	}

	jwtToken, err := h.service.IssueJWT(user)
	if err != nil {
		server.LoggerFromContext(c).WithError(err).WithField("user_id", user.ID).Error("failed to issue jwt")
		server.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to issue token")
		return
	}

	cliFlow := c.Query("cli") == "true"
	if !cliFlow {
		if cliCookie, cookieErr := c.Cookie(oauthCLICookieName); cookieErr == nil && cliCookie == "true" {
			cliFlow = true
		}
	}

	if cliFlow {
		redirectURL := fmt.Sprintf("http://localhost:%s/callback?token=%s&state=%s", h.config.CLIOAuthPort, url.QueryEscape(jwtToken), url.QueryEscape(state))
		server.LoggerFromContext(c).WithFields(logrus.Fields{"provider": provider, "user_id": user.ID, "cli": true}).Info("oauth callback success")
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	c.SetCookie("prompts_token", jwtToken, h.config.JWTExpiryHours*3600, "/", "", false, true)
	server.LoggerFromContext(c).WithFields(logrus.Fields{"provider": provider, "user_id": user.ID, "cli": false}).Info("oauth callback success")
	server.RespondJSON(c, http.StatusOK, gin.H{"token": jwtToken})
}

func randomState(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func signState(state, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(state))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return state + "." + sig
}

func verifyState(signed, state, secret string) bool {
	parts := strings.Split(signed, ".")
	if len(parts) != 2 {
		return false
	}
	if parts[0] != state {
		return false
	}
	return hmac.Equal([]byte(signState(state, secret)), []byte(signed))
}
