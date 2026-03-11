package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/topboyasante/prompts/internal/config"
	"github.com/topboyasante/prompts/internal/identities"
	"github.com/topboyasante/prompts/internal/users"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	googleoauth "golang.org/x/oauth2/google"
)

type ProviderUser struct {
	Provider       string
	ProviderUserID string
	Username       string
	Email          string
	EmailVerified  bool
	AvatarURL      string
}

type Provider interface {
	AuthCodeURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	FetchUser(ctx context.Context, accessToken string) (*ProviderUser, error)
}

type AuthService struct {
	providers      map[string]Provider
	usersRepo      users.Repository
	identitiesRepo identities.Repository
	jwtSecret      string
	jwtExpiry      int
}

func NewAuthService(cfg *config.Config, usersRepo users.Repository, identitiesRepo identities.Repository) *AuthService {
	providers := map[string]Provider{}
	if cfg.GithubClientID != "" && cfg.GithubClientSecret != "" {
		providers["github"] = newGitHubProvider(cfg)
	}
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		providers["google"] = newGoogleProvider(cfg)
	}

	return &AuthService{
		providers:      providers,
		usersRepo:      usersRepo,
		identitiesRepo: identitiesRepo,
		jwtSecret:      cfg.JWTSecret,
		jwtExpiry:      cfg.JWTExpiryHours,
	}
}

func (s *AuthService) IsProviderSupported(provider string) bool {
	_, ok := s.providers[provider]
	return ok
}

func (s *AuthService) GetLoginURL(provider, state string) (string, error) {
	p, err := s.getProvider(provider)
	if err != nil {
		return "", err
	}
	return p.AuthCodeURL(state), nil
}

func (s *AuthService) ExchangeCode(ctx context.Context, provider, code string) (*oauth2.Token, error) {
	p, err := s.getProvider(provider)
	if err != nil {
		return nil, err
	}
	return p.ExchangeCode(ctx, code)
}

func (s *AuthService) GetProviderUser(ctx context.Context, provider, accessToken string) (*ProviderUser, error) {
	p, err := s.getProvider(provider)
	if err != nil {
		return nil, err
	}
	return p.FetchUser(ctx, accessToken)
}

func (s *AuthService) UpsertUser(ctx context.Context, providerUser *ProviderUser) (*users.User, error) {
	identity, err := s.identitiesRepo.FindByProviderUserID(ctx, providerUser.Provider, providerUser.ProviderUserID)
	if err != nil {
		return nil, err
	}
	if identity != nil {
		return s.usersRepo.FindByID(ctx, identity.UserID)
	}

	var user *users.User
	if providerUser.EmailVerified && providerUser.Email != "" {
		user, err = s.usersRepo.FindByEmail(ctx, providerUser.Email)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		user, err = s.usersRepo.Create(ctx, &users.User{
			Username:  fallbackUsername(providerUser),
			Email:     strings.TrimSpace(providerUser.Email),
			AvatarURL: strings.TrimSpace(providerUser.AvatarURL),
		})
		if err != nil {
			return nil, err
		}
	}

	_, err = s.identitiesRepo.Create(ctx, &identities.UserIdentity{
		UserID:         user.ID,
		Provider:       providerUser.Provider,
		ProviderUserID: providerUser.ProviderUserID,
		Email:          strings.TrimSpace(providerUser.Email),
		EmailVerified:  providerUser.EmailVerified,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) IssueJWT(user *users.User) (string, error) {
	return SignToken(user.ID, s.jwtSecret, s.jwtExpiry)
}

func (s *AuthService) GetUserByID(ctx context.Context, id string) (*users.User, error) {
	return s.usersRepo.FindByID(ctx, id)
}

func (s *AuthService) getProvider(provider string) (Provider, error) {
	p, ok := s.providers[strings.ToLower(strings.TrimSpace(provider))]
	if !ok {
		return nil, fmt.Errorf("unsupported oauth provider: %s", provider)
	}
	return p, nil
}

func fallbackUsername(user *ProviderUser) string {
	if strings.TrimSpace(user.Username) != "" {
		return strings.TrimSpace(user.Username)
	}
	if strings.TrimSpace(user.Email) != "" {
		parts := strings.Split(user.Email, "@")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	return user.Provider + "-user"
}

type githubProvider struct {
	oauth      *oauth2.Config
	httpClient *http.Client
}

func newGitHubProvider(cfg *config.Config) Provider {
	return &githubProvider{
		oauth: &oauth2.Config{
			ClientID:     cfg.GithubClientID,
			ClientSecret: cfg.GithubClientSecret,
			RedirectURL:  cfg.GithubRedirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		},
		httpClient: &http.Client{},
	}
}

func (p *githubProvider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state)
}

func (p *githubProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauth.Exchange(ctx, code)
}

func (p *githubProvider) FetchUser(ctx context.Context, accessToken string) (*ProviderUser, error) {
	type githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	type githubEmail struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("github user request failed: %s", resp.Status)
	}

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("github user id missing")
	}

	email := strings.TrimSpace(user.Email)
	emailVerified := false

	if email == "" {
		reqEmails, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
		if err == nil {
			reqEmails.Header.Set("Authorization", "Bearer "+accessToken)
			reqEmails.Header.Set("Accept", "application/vnd.github+json")
			respEmails, err := p.httpClient.Do(reqEmails)
			if err == nil {
				defer respEmails.Body.Close()
				if respEmails.StatusCode < http.StatusBadRequest {
					var emails []githubEmail
					if err := json.NewDecoder(respEmails.Body).Decode(&emails); err == nil {
						for _, e := range emails {
							if e.Primary && e.Verified {
								email = strings.TrimSpace(e.Email)
								emailVerified = true
								break
							}
						}
						if email == "" {
							for _, e := range emails {
								if e.Verified {
									email = strings.TrimSpace(e.Email)
									emailVerified = true
									break
								}
							}
						}
					}
				}
			}
		}
	}

	return &ProviderUser{
		Provider:       "github",
		ProviderUserID: fmt.Sprintf("%d", user.ID),
		Username:       strings.TrimSpace(user.Login),
		Email:          email,
		EmailVerified:  emailVerified,
		AvatarURL:      strings.TrimSpace(user.AvatarURL),
	}, nil
}

type googleProvider struct {
	oauth      *oauth2.Config
	httpClient *http.Client
}

func newGoogleProvider(cfg *config.Config) Provider {
	return &googleProvider{
		oauth: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleoauth.Endpoint,
		},
		httpClient: &http.Client{},
	}
}

func (p *googleProvider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state)
}

func (p *googleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.oauth.Exchange(ctx, code)
}

func (p *googleProvider) FetchUser(ctx context.Context, accessToken string) (*ProviderUser, error) {
	type googleUser struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Picture       string `json:"picture"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("google user request failed: %s", resp.Status)
	}

	var user googleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	if strings.TrimSpace(user.Sub) == "" {
		return nil, fmt.Errorf("google user id missing")
	}

	return &ProviderUser{
		Provider:       "google",
		ProviderUserID: strings.TrimSpace(user.Sub),
		Username:       strings.TrimSpace(user.Name),
		Email:          strings.TrimSpace(user.Email),
		EmailVerified:  user.EmailVerified,
		AvatarURL:      strings.TrimSpace(user.Picture),
	}, nil
}
