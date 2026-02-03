package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service handles password hashing and JWT issuance.
type Service struct {
	secret         []byte
	sessionTTL     time.Duration
	rememberMeTTL  time.Duration
	nowFn          func() time.Time
}

func NewService(secret string, ttlMinutes int, rememberMeDays int) *Service {
	sessionTTL := time.Duration(ttlMinutes) * time.Minute
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}
	rememberMeTTL := time.Duration(rememberMeDays) * 24 * time.Hour
	if rememberMeDays <= 0 {
		rememberMeTTL = 30 * 24 * time.Hour
	}
	return &Service{
		secret:        []byte(secret),
		sessionTTL:    sessionTTL,
		rememberMeTTL: rememberMeTTL,
		nowFn:         time.Now,
	}
}

// TokenTTL returns the JWT/cookie TTL for the given remember-me choice.
func (s *Service) TokenTTL(remember bool) time.Duration {
	if remember {
		return s.rememberMeTTL
	}
	return s.sessionTTL
}

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *Service) ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func (s *Service) GenerateToken(userID string, remember bool) (string, error) {
	ttl := s.TokenTTL(remember)
	now := s.nowFn()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

var ErrInvalidToken = errors.New("invalid token")

func (s *Service) ParseToken(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, ErrInvalidToken
	}

	return *claims, nil
}
