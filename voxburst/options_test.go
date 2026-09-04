package voxburst

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultBaseURLIsVoxBurstProduction(t *testing.T) {
	if DefaultBaseURL != "https://api.voxburst.io/v1" {
		t.Fatalf("DefaultBaseURL = %q, want https://api.voxburst.io/v1", DefaultBaseURL)
	}
	if strings.Contains(DefaultBaseURL, "socialdispatch") {
		t.Fatalf("DefaultBaseURL must not reference the retired socialdispatch.io domain")
	}

	c := NewClient("vb_test_key")
	if c.baseURL != DefaultBaseURL {
		t.Fatalf("NewClient baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
}

func TestDefaultUserAgentCarriesVersion(t *testing.T) {
	c := NewClient("vb_test_key")
	want := "voxburst-go/" + Version
	if c.config.userAgent != want {
		t.Fatalf("userAgent = %q, want %q", c.config.userAgent, want)
	}
}

func TestWithBaseURLAcceptsHTTPS(t *testing.T) {
	c := NewClient("vb_test_key", WithBaseURL("https://example.test/v1"))
	if c.baseURL != "https://example.test/v1" {
		t.Fatalf("baseURL = %q, want https://example.test/v1", c.baseURL)
	}
}

func TestWithBaseURLRejectsNonHTTPS(t *testing.T) {
	cases := []string{
		"http://api.voxburst.io/v1",
		"http://localhost:3000/v1",
		"http://127.0.0.1:3000/v1",
		"ftp://api.voxburst.io/v1",
		"api.voxburst.io/v1",
		"",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("WithBaseURL(%q) did not panic", u)
				}
			}()
			WithBaseURL(u)
		})
	}
}

// TestWithBaseURLWorksWithTLSTestServer documents the supported pattern for
// local testing now that plaintext base URLs are rejected.
func TestWithBaseURLWorksWithTLSTestServer(t *testing.T) {
	var gotAuth, gotUA string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"pagination":{"hasMore":false}}`))
	}))
	defer srv.Close()

	c := NewClient("vb_test_key",
		WithBaseURL(srv.URL+"/v1"),
		WithHTTPClient(srv.Client()),
		WithNoRetry(),
	)
	if _, err := c.Accounts.List(context.Background(), nil); err != nil {
		t.Fatalf("Accounts.List: %v", err)
	}
	if gotAuth != "Bearer vb_test_key" {
		t.Fatalf("Authorization = %q, want Bearer vb_test_key", gotAuth)
	}
	if gotUA != "voxburst-go/"+Version {
		t.Fatalf("User-Agent = %q, want voxburst-go/%s", gotUA, Version)
	}
}
