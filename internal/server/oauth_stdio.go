package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"golang.org/x/oauth2"
)

// stdioOAuthConfig holds the parameters for the stdio OAuth2 browser flow.
type stdioOAuthConfig struct {
	Issuer string
	// ClientID is a pre-registered client; empty registers one per flow.
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// stdioOAuthFlow performs a full OAuth2 authorization code flow with PKCE via the browser.
// It starts a temporary local HTTP server to receive the callback, opens the browser to
// the authorization URL, and exchanges the code for tokens.
//
// When cfg.ClientID is empty, the client registers itself for this run's redirect URI
// (see registerStdioClient). The returned config is the client that obtained the token;
// refreshing must use it.
func stdioOAuthFlow(ctx context.Context, logger *slog.Logger, cfg stdioOAuthConfig) (*oauth2.Token, *oauth2.Config, error) {
	// Generate PKCE code verifier (43-128 chars of unreserved characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, nil, fmt.Errorf("generating code verifier: %w", err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// S256 challenge
	h := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

	// Generate state parameter for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, nil, fmt.Errorf("generating state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Start temporary local server on random port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, fmt.Errorf("starting callback server: %w", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	var oauthCfg *oauth2.Config
	if cfg.ClientID != "" {
		oauthCfg = cfg.oauth2Config()
		oauthCfg.RedirectURL = redirectURI
	} else {
		oauthCfg, err = registerStdioClient(ctx, cfg, redirectURI)
		if err != nil {
			return nil, nil, err
		}
		logger.Info("registered OAuth2 client", "client_id", oauthCfg.ClientID)
	}

	type callbackResult struct {
		code string
		err  error
	}
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			desc := r.URL.Query().Get("error_description")
			resultCh <- callbackResult{err: fmt.Errorf("oauth error: %s: %s", errMsg, desc)}
			fmt.Fprintf(w, "<html><body><h1>Authorization failed</h1><p>%s</p><p>You can close this window.</p></body></html>", desc)
			return
		}

		if gotState := r.URL.Query().Get("state"); gotState != state {
			resultCh <- callbackResult{err: fmt.Errorf("state mismatch: got %q, want %q", gotState, state)}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			resultCh <- callbackResult{err: fmt.Errorf("no authorization code in callback")}
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		resultCh <- callbackResult{code: code}
		fmt.Fprint(w, "<html><body><h1>Authorization successful</h1><p>You can close this window and return to the terminal.</p></body></html>")
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("callback server error", "error", err)
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	// Build authorization URL with PKCE
	authURL := oauthCfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	logger.Info("opening browser for authorization", "url", authURL)

	if err := openBrowserFunc(authURL); err != nil {
		logger.Warn("failed to open browser", "error", err)
	}

	// Wait for callback
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case result := <-resultCh:
		if result.err != nil {
			return nil, nil, result.err
		}

		// Exchange code for tokens
		token, err := oauthCfg.Exchange(ctx, result.code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("exchanging authorization code: %w", err)
		}

		logger.Info("authorization successful")
		return token, oauthCfg, nil
	}
}

// oauth2Config returns the client for a pre-registered client ID, whose
// endpoints are the authorization server's fixed django-oauth-toolkit paths.
func (cfg stdioOAuthConfig) oauth2Config() *oauth2.Config {
	issuer := strings.TrimRight(cfg.Issuer, "/")
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  issuer + "/o/authorize/",
			TokenURL: issuer + "/o/token/",
		},
	}
}

// registerStdioClient registers a public client for redirectURI at the
// registration endpoint the issuer advertises in its RFC 8414 metadata
// (RFC 7591), and returns it with the advertised endpoints.
//
// The registration is not persisted: every process that runs the browser flow
// registers a new client, and the authorization server rate-limits
// registration per address.
func registerStdioClient(ctx context.Context, cfg stdioOAuthConfig, redirectURI string) (*oauth2.Config, error) {
	meta, err := discoverAuthServer(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	if meta.RegistrationEndpoint == "" {
		return nil, fmt.Errorf("authorization server %s does not offer client registration; set -client-id", cfg.Issuer)
	}

	reg, err := oauthex.RegisterClient(ctx, meta.RegistrationEndpoint, &oauthex.ClientRegistrationMetadata{
		RedirectURIs:            []string{redirectURI},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Uptime.com MCP Server",
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("registering OAuth2 client: %w", err)
	}

	return &oauth2.Config{
		ClientID:     reg.ClientID,
		ClientSecret: reg.ClientSecret,
		Scopes:       cfg.Scopes,
		RedirectURL:  redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:   meta.AuthorizationEndpoint,
			TokenURL:  meta.TokenEndpoint,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}, nil
}

// discoverAuthServer fetches the issuer's RFC 8414 metadata. The issuer it
// names may differ from the configured one by a trailing slash: uptime.com
// publishes "https://uptime.com/" while -uptime-url is written without one.
func discoverAuthServer(ctx context.Context, issuer string) (*oauthex.AuthServerMeta, error) {
	base := strings.TrimRight(issuer, "/")
	metadataURL := base + "/.well-known/oauth-authorization-server"

	var errs []error
	for _, candidate := range []string{base + "/", base} {
		meta, err := oauthex.GetAuthServerMeta(ctx, metadataURL, candidate, nil)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if meta == nil {
			return nil, fmt.Errorf("authorization server %s publishes no metadata at %s; set -client-id", issuer, metadataURL)
		}
		return meta, nil
	}
	return nil, fmt.Errorf("discovering authorization server: %w", errors.Join(errs...))
}

// startTokenRefresh starts a background goroutine that refreshes the OAuth2 token
// before it expires. It updates the tokenHolder with the new token.
func startTokenRefresh(ctx context.Context, logger *slog.Logger, holder *tokenHolder, oauthCfg *oauth2.Config) {
	go func() {
		for {
			token := holder.Token()
			if token == nil {
				return
			}

			// Refresh 60 seconds before expiry
			refreshAt := token.Expiry.Add(-60 * time.Second)
			sleepDuration := time.Until(refreshAt)
			if sleepDuration <= 0 {
				sleepDuration = 30 * time.Second
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(sleepDuration):
				logger.Debug("refreshing OAuth2 token")

				src := oauthCfg.TokenSource(ctx, token)
				newToken, err := src.Token()
				if err != nil {
					logger.Error("failed to refresh token", "error", err)
					continue
				}

				holder.Update(newToken)
				logger.Info("token refreshed", "expiry", newToken.Expiry)
			}
		}
	}()
}

// openBrowserFunc opens the given URL in the default browser.
// It is a variable to allow overriding in tests.
var openBrowserFunc = openBrowserDefault

func openBrowserDefault(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
