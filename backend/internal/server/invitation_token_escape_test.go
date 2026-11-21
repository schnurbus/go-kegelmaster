package server

import (
	"net/http"
	"testing"
	"time"
)

// Invitation tokens are padded URL-safe base64 and end with "=". Clients that
// encode the path segment (encodeURIComponent) send "%3D"; the handler must
// look up the decoded token.
func TestGetInvitation_DecodesEscapedToken(t *testing.T) {
	srv, _, invMock := newRegistrationModeServer(t, "invite_only")
	token := "abc-123_xyz="
	// Expired invitation: the handler answers 410 right after the lookup.
	expectInvitationByToken(invMock, token, "neu@example.com", time.Now().Add(-time.Hour), nil)

	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodGet, "/api/invitations/abc-123_xyz%3D", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusGone {
		t.Fatalf("expected 410 for the decoded token, got %d", resp.StatusCode)
	}
	if err := invMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
