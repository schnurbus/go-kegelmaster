package auth

import (
	"testing"
	"time"
)

func TestHashAndComparePassword(t *testing.T) {
	svc := NewService("secret", 60, 30)

	hash, err := svc.HashPassword("super-secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := svc.ComparePassword(hash, "super-secret"); err != nil {
		t.Fatalf("compare password: %v", err)
	}

	if err := svc.ComparePassword(hash, "wrong"); err == nil {
		t.Fatalf("expected compare to fail for wrong password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	svc := NewService("secret", 60, 30)
	token, err := svc.GenerateToken("user-123", true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.UserID != "user-123" || claims.Subject != "user-123" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestGenerateTokenUsesTTL(t *testing.T) {
	svc := NewService("secret", 5, 30) // 5 minutes session TTL
	token, err := svc.GenerateToken("user-ttl", false)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if diff < 4*time.Minute || diff > 6*time.Minute {
		t.Fatalf("expected ttl of ~5m, got %s", diff)
	}
}

func TestIsRememberMe(t *testing.T) {
	svc := NewService("secret", 60, 30)
	for _, remember := range []bool{true, false} {
		token, err := svc.GenerateToken("user-remember", remember)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}
		claims, err := svc.ParseToken(token)
		if err != nil {
			t.Fatalf("parse token: %v", err)
		}
		if got := svc.IsRememberMe(claims); got != remember {
			t.Fatalf("IsRememberMe = %v, want %v", got, remember)
		}
	}
}

func TestGenerateTokenSetsAuthTime(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	svc := NewService("secret", 60, 30)
	svc.nowFn = func() time.Time { return now }

	token, err := svc.GenerateToken("user-1", false)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	at, ok := claims.AuthenticatedAt()
	if !ok || !at.Equal(now) {
		t.Fatalf("auth_time = %v (ok=%v), want %v", at, ok, now)
	}
}

// A reissued session (e.g. a sliding refresh) gets a new lifetime but keeps
// the original auth_time, so it does not count as a re-authentication.
func TestGenerateTokenWithAuthTimeKeepsAuthTime(t *testing.T) {
	authTime := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	svc := NewService("secret", 60, 30)

	token, err := svc.GenerateTokenWithAuthTime("user-1", true, authTime)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if at, ok := claims.AuthenticatedAt(); !ok || !at.Equal(authTime) {
		t.Fatalf("auth_time = %v (ok=%v), want %v", at, ok, authTime)
	}
	if !claims.IssuedAt.Time.After(authTime) {
		t.Fatalf("iat = %v should be the reissue time, not auth_time", claims.IssuedAt.Time)
	}
	if !svc.IsRememberMe(claims) {
		t.Fatalf("expected remember-me lifetime to be kept")
	}
}

func TestAuthenticatedAtMissingClaim(t *testing.T) {
	if _, ok := (Claims{}).AuthenticatedAt(); ok {
		t.Fatalf("expected no auth_time for claims without the claim")
	}
}
