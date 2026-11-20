package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestVerifyPKCE(t *testing.T) {
	verifier := strings.Repeat("a", 20) + "-._~" + strings.Repeat("Z9", 12)
	challenge := challengeFor(verifier)
	if !ValidCodeChallenge(challenge) {
		t.Fatalf("challenge %q should be valid", challenge)
	}
	if !VerifyPKCE(verifier, challenge) {
		t.Fatal("matching verifier rejected")
	}
	for name, v := range map[string]string{
		"wrong verifier": strings.Repeat("b", 43),
		"too short":      "abc",
		"too long":       strings.Repeat("a", 129),
		"invalid chars":  verifier[:42] + "!",
		"plain method":   challenge,
	} {
		if VerifyPKCE(v, challenge) {
			t.Errorf("%s: verifier accepted", name)
		}
	}
	if ValidCodeChallenge("short") || ValidCodeChallenge(strings.Repeat("+", 43)) {
		t.Error("invalid challenge accepted")
	}
}

func TestValidateRedirectURI(t *testing.T) {
	valid := []string{
		"https://claude.ai/api/mcp/auth_callback",
		"https://chatgpt.com/connector_platform_oauth_redirect?x=1",
		"http://localhost:33418/callback",
		"http://127.0.0.1:8765/cb",
		"http://[::1]:8765/cb",
	}
	for _, u := range valid {
		if err := ValidateRedirectURI(u); err != nil {
			t.Errorf("%s rejected: %v", u, err)
		}
	}
	invalid := []string{
		"",
		"http://example.com/callback",
		"https://example.com/cb#frag",
		"https://user:pw@example.com/cb",
		"/relative",
		"javascript:alert(1)",
		"cursor://callback",
		"http://localhost.evil.com/cb",
		" https://example.com/cb",
		"https://" + strings.Repeat("a", MaxRedirectURILength) + ".com",
	}
	for _, u := range invalid {
		if err := ValidateRedirectURI(u); err == nil {
			t.Errorf("%q accepted", u)
		}
	}
}

func TestSameResource(t *testing.T) {
	canonical := "https://kegel.example.com/mcp"
	for _, r := range []string{"https://kegel.example.com/mcp", "https://kegel.example.com/mcp/", "HTTPS://Kegel.Example.com/mcp"} {
		if !SameResource(r, canonical) {
			t.Errorf("%s should match", r)
		}
	}
	for _, r := range []string{"https://evil.example.com/mcp", "https://kegel.example.com/", "https://kegel.example.com/mcp?x=1", "http://kegel.example.com/mcp"} {
		if SameResource(r, canonical) {
			t.Errorf("%s should not match", r)
		}
	}
}

func TestScopeAndState(t *testing.T) {
	if !ValidScopeSyntax("") || !ValidScopeSyntax("mcp offline_access") {
		t.Error("valid scope rejected")
	}
	if ValidScopeSyntax("mcp\n") || ValidScopeSyntax(`a"b`) {
		t.Error("invalid scope accepted")
	}
	if !ValidState("abc-123_XYZ.~") || ValidState("a\nb") || ValidState(strings.Repeat("a", MaxStateLength+1)) {
		t.Error("state validation wrong")
	}
}

func TestAppendQueryKeepsExistingQuery(t *testing.T) {
	got := AppendQuery("https://client.example/cb?tenant=1", map[string]string{"code": "c", "state": "s&x", "empty": ""})
	if got != "https://client.example/cb?code=c&state=s%26x&tenant=1" {
		t.Errorf("unexpected redirect %q", got)
	}
}
