package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// stdioOAuthConfig holds the parameters for the stdio OAuth2 browser flow.
type stdioOAuthConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// stdioOAuthFlow performs a full OAuth2 authorization code flow with PKCE via the browser.
// It starts a temporary local HTTP server to receive the callback, opens the browser to
// the authorization URL, and exchanges the code for tokens.
func stdioOAuthFlow(ctx context.Context, logger *slog.Logger, cfg stdioOAuthConfig) (*oauth2.Token, error) {
	issuer := strings.TrimRight(cfg.Issuer, "/")

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  issuer + "/o/authorize/",
			TokenURL: issuer + "/o/token/",
		},
	}

	// Generate PKCE code verifier (43-128 chars of unreserved characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("generating code verifier: %w", err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// S256 challenge
	h := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

	// Generate state parameter for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Start temporary local server on random port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)
	oauthCfg.RedirectURL = redirectURI

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
		return nil, ctx.Err()
	case result := <-resultCh:
		if result.err != nil {
			return nil, result.err
		}

		// Exchange code for tokens
		token, err := oauthCfg.Exchange(ctx, result.code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		)
		if err != nil {
			return nil, fmt.Errorf("exchanging authorization code: %w", err)
		}

		logger.Info("authorization successful")
		return token, nil
	}
}

// startTokenRefresh starts a background goroutine that refreshes the OAuth2 token
// before it expires. It updates the tokenHolder with the new token.
func startTokenRefresh(ctx context.Context, logger *slog.Logger, holder *tokenHolder, cfg stdioOAuthConfig) {
	issuer := strings.TrimRight(cfg.Issuer, "/")
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  issuer + "/o/authorize/",
			TokenURL: issuer + "/o/token/",
		},
	}

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
