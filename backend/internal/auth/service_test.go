package auth

import (
	"testing"
	"time"
)

func TestHashAndComparePassword(t *testing.T) {
	svc := NewService("secret", 60)

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
	svc := NewService("secret", 60)
	token, err := svc.GenerateToken("user-123")
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
	svc := NewService("secret", 5) // 5 minutes
	token, err := svc.GenerateToken("user-ttl")
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
