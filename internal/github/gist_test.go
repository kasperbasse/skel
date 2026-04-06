package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- ParseSource tests ---

func TestParseSourceGistURL(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
	}{
		{"user/id", "https://gist.github.com/kasperbasse/abc123def456789012345678", "abc123def456789012345678"},
		{"trailing slash", "https://gist.github.com/kasperbasse/abc123def456789012345678/", "abc123def456789012345678"},
		{"bare id", "https://gist.github.com/abc123def456789012345678", "abc123def456789012345678"},
		{"with revision", "https://gist.github.com/user/abc123def456789012345678/aabbccddee00112233445566778899aabbccddee", "abc123def456789012345678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSource(tt.input)
			if err != nil {
				t.Fatalf("ParseSource(%q) error: %v", tt.input, err)
			}
			if got != tt.wantID {
				t.Errorf("ParseSource(%q) = %q, want %q", tt.input, got, tt.wantID)
			}
		})
	}
}

func TestParseSourceShorthand(t *testing.T) {
	got, err := ParseSource("github:kasperbasse/abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestParseSourceShorthandMissingID(t *testing.T) {
	_, err := ParseSource("github:kasperbasse/")
	if err == nil {
		t.Error("expected error for missing gist ID")
	}
}

func TestParseSourceShorthandNoSlash(t *testing.T) {
	_, err := ParseSource("github:justuser")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseSourceInvalid(t *testing.T) {
	_, err := ParseSource("not-a-url")
	if err == nil {
		t.Error("expected error for invalid source")
	}
}

// --- FetchGist tests ---

func TestFetchGistSuccess(t *testing.T) {
	gist := Gist{
		ID:      "abc123",
		HTMLURL: "https://gist.github.com/user/abc123",
		Files: map[string]GistFile{
			"test.json": {Filename: "test.json", Size: 42, Content: `{"name":"test"}`},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gists/abc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header")
		}
		w.WriteHeader(200)
		err := json.NewEncoder(w).Encode(gist)
		if err != nil {
			return
		}
	}))
	defer srv.Close()

	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	got, err := FetchGist("abc123")
	if err != nil {
		t.Fatalf("FetchGist error: %v", err)
	}
	if got.ID != "abc123" {
		t.Errorf("got ID %q, want %q", got.ID, "abc123")
	}
	if len(got.Files) != 1 {
		t.Errorf("got %d files, want 1", len(got.Files))
	}
}

func TestFetchGist404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, err := w.Write([]byte(`{"message":"Not Found"}`))
		if err != nil {
			return
		}
	}))
	defer srv.Close()

	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	_, err := FetchGist("nonexistent")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestFetchGistRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(403)
		_, err := w.Write([]byte(`{"message":"rate limit"}`))
		if err != nil {
			return
		}
	}))
	defer srv.Close()

	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	_, err := FetchGist("abc123")
	if err == nil {
		t.Fatal("expected error for rate limit")
	}
	if !contains(err.Error(), "rate limit") {
		t.Errorf("error should mention 'rate limit', got: %v", err)
	}
}

// --- CreateGist tests ---

func TestCreateGistSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var req CreateGistRequest
		decodeErr := json.NewDecoder(r.Body).Decode(&req)
		if decodeErr != nil {
			return
		}
		if req.Description == "" {
			t.Error("missing description")
		}

		w.WriteHeader(201)
		encodeErr := json.NewEncoder(w).Encode(Gist{
			ID:      "new123",
			HTMLURL: "https://gist.github.com/user/new123",
		})
		if encodeErr != nil {
			return
		}
	}))
	defer srv.Close()

	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	gist, err := CreateGist("test-token", &CreateGistRequest{
		Description: "test",
		Public:      true,
		Files: map[string]CreateGistFile{
			"test.json": {Content: `{"name":"test"}`},
		},
	})
	if err != nil {
		t.Fatalf("CreateGist error: %v", err)
	}
	if gist.ID != "new123" {
		t.Errorf("got ID %q, want %q", gist.ID, "new123")
	}
}

func TestCreateGist401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, err := w.Write([]byte(`{"message":"Bad credentials"}`))
		if err != nil {
			return
		}
	}))
	defer srv.Close()

	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	_, err := CreateGist("bad-token", &CreateGistRequest{
		Files: map[string]CreateGistFile{"f.json": {Content: "{}"}},
	})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !contains(err.Error(), "authentication") {
		t.Errorf("error should mention 'authentication', got: %v", err)
	}
}

// --- FindProfileJSON tests ---

func TestFindProfileJSONSuccess(t *testing.T) {
	gist := &Gist{
		Files: map[string]GistFile{
			"my-skel.json": {Filename: "my-skel.json", Size: 15, Content: `{"name":"test"}`},
		},
	}
	content, err := FindProfileJSON(gist, 50*1024*1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != `{"name":"test"}` {
		t.Errorf("unexpected content: %s", content)
	}
}

func TestFindProfileJSONNoJSON(t *testing.T) {
	gist := &Gist{
		Files: map[string]GistFile{
			"readme.md": {Filename: "readme.md", Size: 10, Content: "# hi"},
		},
	}
	_, err := FindProfileJSON(gist, 50*1024*1024)
	if err == nil {
		t.Fatal("expected error for no json files")
	}
	if !contains(err.Error(), "no .json") {
		t.Errorf("error should mention 'no .json', got: %v", err)
	}
}

func TestFindProfileJSONMultiple(t *testing.T) {
	gist := &Gist{
		Files: map[string]GistFile{
			"a.json": {Filename: "a.json", Size: 5, Content: "{}"},
			"b.json": {Filename: "b.json", Size: 5, Content: "{}"},
		},
	}
	_, err := FindProfileJSON(gist, 50*1024*1024)
	if err == nil {
		t.Fatal("expected error for multiple json files")
	}
	if !contains(err.Error(), "multiple") {
		t.Errorf("error should mention 'multiple', got: %v", err)
	}
}

func TestFindProfileJSONTooLarge(t *testing.T) {
	gist := &Gist{
		Files: map[string]GistFile{
			"big.json": {Filename: "big.json", Size: 100 * 1024 * 1024, Content: ""},
		},
	}
	_, err := FindProfileJSON(gist, 50*1024*1024)
	if err == nil {
		t.Fatal("expected error for too large file")
	}
	if !contains(err.Error(), "too large") {
		t.Errorf("error should mention 'too large', got: %v", err)
	}
}

func TestFindProfileJSONFetchesRawURL(t *testing.T) {
	// In production, raw_url fetching requires a gist.githubusercontent.com URL.
	// Here we exercise the inline-content path (Content != ""), which is the common case.
	gist := &Gist{
		Files: map[string]GistFile{
			"big.json": {Filename: "big.json", Size: 18, Content: `{"name":"fetched"}`},
		},
	}
	content, err := FindProfileJSON(gist, 50*1024*1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != `{"name":"fetched"}` {
		t.Errorf("unexpected content: %s", content)
	}
}

// --- validateRawURL tests ---

func TestValidateRawURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{"valid gist raw URL", "https://gist.githubusercontent.com/user/abc123/raw/file.json", false},
		{"non-https scheme", "http://gist.githubusercontent.com/user/abc123/raw/file.json", true},
		{"wrong host", "https://evil.com/gist.githubusercontent.com/raw", true},
		{"internal address", "https://169.254.169.254/latest/meta-data/", true},
		{"file scheme", "file:///etc/passwd", true},
		{"invalid URL", "://not a url", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRawURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRawURL(%q) error = %v, wantErr = %v", tt.rawURL, err, tt.wantErr)
			}
		})
	}
}

func TestFindProfileJSONRejectsInvalidRawURL(t *testing.T) {
	gist := &Gist{
		Files: map[string]GistFile{
			"profile.json": {
				Filename: "profile.json",
				Size:     100,
				Content:  "", // empty content triggers raw URL fetch
				RawURL:   "http://evil.com/raw",
			},
		},
	}
	_, err := FindProfileJSON(gist, 50*1024*1024)
	if err == nil {
		t.Fatal("expected error for invalid raw URL")
	}
	if !contains(err.Error(), "invalid raw URL") {
		t.Errorf("error should mention 'invalid raw URL', got: %v", err)
	}
}

// --- Helper tests ---

func TestIsHex(t *testing.T) {
	if !isHex("abc123DEF") {
		t.Error("expected true for valid hex")
	}
	if isHex("xyz") {
		t.Error("expected false for non-hex")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("got %q", got)
	}
	if got := truncate("hi", 10); got != "hi" {
		t.Errorf("got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
