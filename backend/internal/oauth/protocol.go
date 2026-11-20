// Package oauth implements Kegelmaster's OAuth 2.1 authorization server for
// MCP clients (KEG-12): dynamically registered public clients, authorization
// code flow with mandatory PKCE (S256), short-lived access tokens and
// rotating refresh tokens. Access tokens are only accepted by the MCP server
// (see InProcessKey); codes and tokens are stored as SHA-256 hashes only.
package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	// Token prefixes make leaked secrets recognisable (e.g. by secret scanners)
	// and let the API distinguish OAuth access tokens from personal API tokens.
	AccessTokenPrefix  = "kgo_"
	RefreshTokenPrefix = "kgr_"
	CodePrefix         = "kgc_"

	// Scope is the only scope Kegelmaster grants: access to the MCP server
	// with the rights of the user.
	Scope = "mcp"

	CodeTTL         = 10 * time.Minute
	AccessTokenTTL  = time.Hour
	RefreshTokenTTL = 30 * 24 * time.Hour

	randomBytes = 32

	MaxRedirectURIs      = 5
	MaxRedirectURILength = 2000
	MaxClientNameLength  = 100
	MaxStateLength       = 1024
	MaxScopeLength       = 1024
)

// InProcessKey is the fasthttp user value (Fiber local) the MCP server sets on
// the REST requests it dispatches in-process. OAuth access tokens are bound to
// the MCP resource and are only accepted on requests carrying this marker; an
// external HTTP request cannot set user values.
type InProcessKey struct{}

// HashToken returns the value stored in the database for a code or token.
// They carry 256 bits of randomness, so a plain SHA-256 is sufficient.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generate(prefix string) (string, error) {
	buf := make([]byte, randomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

var (
	ErrInvalidRedirectURI = errors.New("invalid redirect_uri")
)

// ValidateRedirectURI accepts https URLs and http URLs on a loopback host
// (localhost, 127.0.0.1, [::1]) for native clients. Fragments, user info and
// relative URLs are rejected.
func ValidateRedirectURI(raw string) error {
	if raw == "" || len(raw) > MaxRedirectURILength || strings.TrimSpace(raw) != raw {
		return ErrInvalidRedirectURI
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.Opaque != "" || u.User != nil || u.Fragment != "" || strings.Contains(raw, "#") {
		return ErrInvalidRedirectURI
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if IsLoopbackHost(u.Hostname()) {
			return nil
		}
	}
	return ErrInvalidRedirectURI
}

// IsLoopbackHost reports whether host is localhost or a loopback IP literal.
func IsLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// ValidCodeChallenge reports whether challenge looks like an S256 challenge
// (base64url without padding of a SHA-256 digest).
func ValidCodeChallenge(challenge string) bool {
	if len(challenge) != 43 {
		return false
	}
	for _, r := range challenge {
		if !isBase64URLChar(r) {
			return false
		}
	}
	return true
}

// VerifyPKCE checks a code_verifier (RFC 7636) against an S256 challenge.
func VerifyPKCE(verifier, challenge string) bool {
	if len(verifier) < 43 || len(verifier) > 128 {
		return false
	}
	for _, r := range verifier {
		if !isBase64URLChar(r) && r != '.' && r != '~' {
			return false
		}
	}
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

func isBase64URLChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
}

// ValidScopeSyntax reports whether scope is a syntactically valid
// space-separated list of scope tokens (RFC 6749, section 3.3). Kegelmaster
// always grants Scope; other requested values are ignored as the RFC allows.
func ValidScopeSyntax(scope string) bool {
	if len(scope) > MaxScopeLength {
		return false
	}
	for _, r := range scope {
		if r == ' ' {
			continue
		}
		if r < 0x21 || r == 0x22 || r == 0x5C || r > 0x7E {
			return false
		}
	}
	return true
}

// ValidState reports whether state can be passed through safely.
func ValidState(state string) bool {
	if len(state) > MaxStateLength {
		return false
	}
	for _, r := range state {
		if r < 0x20 || r > 0x7E {
			return false
		}
	}
	return true
}

// SameResource compares a resource indicator (RFC 8707) with the canonical
// resource URL, ignoring a trailing slash and the case of scheme and host.
func SameResource(given, canonical string) bool {
	g, err1 := url.Parse(given)
	c, err2 := url.Parse(canonical)
	if err1 != nil || err2 != nil || g.Fragment != "" || g.RawQuery != "" {
		return false
	}
	return strings.EqualFold(g.Scheme, c.Scheme) &&
		strings.EqualFold(g.Host, c.Host) &&
		strings.TrimSuffix(g.Path, "/") == strings.TrimSuffix(c.Path, "/")
}

// AppendQuery adds params to a registered redirect URI, keeping its own query.
func AppendQuery(redirectURI string, params map[string]string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
