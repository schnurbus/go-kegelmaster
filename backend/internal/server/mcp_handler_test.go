package server

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func mcpRequest(auth string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	return req
}

func TestMCP_RequiresBearerToken(t *testing.T) {
	srv, _, authSvc := newTestServer(t)

	resp := doRequest(t, srv, mcpRequest(""))
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("WWW-Authenticate"), "Bearer") {
		t.Errorf("expected WWW-Authenticate header, got %q", resp.Header.Get("WWW-Authenticate"))
	}

	// A browser session cookie is not enough: MCP is for API tokens.
	cookieToken, _ := authSvc.GenerateToken(uuid.NewString(), false)
	req := mcpRequest("")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: cookieToken})
	resp = doRequest(t, srv, req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with cookie only, got %d", resp.StatusCode)
	}
}

func TestMCP_ListsToolsWithAPIToken(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	userID, tokenID := uuid.NewString(), uuid.NewString()

	expectAPITokenLookup(mock, tokenID, userID, time.Now().UTC().Add(-time.Hour), nil)
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`UPDATE api_tokens`).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, mcpRequest("Bearer "+testAPIToken))
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, raw)
	}
	for _, tool := range []string{"klubs_auflisten", "spieltag_erfassen", "spieltag_veroeffentlichen", "buchung_erfassen"} {
		if !strings.Contains(string(raw), `"`+tool+`"`) {
			t.Errorf("tools/list misses %s", tool)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
