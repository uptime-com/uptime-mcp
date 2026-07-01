package app

import "testing"

func TestAPIBaseURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "default",
			cfg:  Config{},
			want: "https://uptime.com/api/v1/",
		},
		{
			name: "derived from uptime-url",
			cfg:  Config{UptimeURL: "https://example.com/"},
			want: "https://example.com/api/v1/",
		},
		{
			name: "api-url override used verbatim",
			cfg:  Config{UptimeURL: "https://example.com", APIURL: "http://uptime.svc.cluster.local/api/v1/"},
			want: "http://uptime.svc.cluster.local/api/v1/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.APIBaseURL(); got != tt.want {
				t.Errorf("APIBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOAuthIssuer(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "empty",
			cfg:  Config{},
			want: "",
		},
		{
			name: "falls back to uptime-url",
			cfg:  Config{UptimeURL: "https://example.com"},
			want: "https://example.com",
		},
		{
			name: "oauth-url override used verbatim",
			cfg:  Config{UptimeURL: "http://internal", OAuthURL: "https://public.example.com"},
			want: "https://public.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.OAuthIssuer(); got != tt.want {
				t.Errorf("OAuthIssuer() = %q, want %q", got, tt.want)
			}
		})
	}
}
