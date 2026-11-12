package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/TheAmirhosssein/class-manager/internal/types"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SubjectKindUser  = "user"
	SubjectKindAdmin = "admin"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	Kind  string `json:"kind"`
	Role  string `json:"role"`
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

type IssueParams struct {
	SubjectID types.ID
	Kind      string
	Role      string
	Email     string
	Name      string
}

type Service struct {
	secret    []byte
	accessTTL time.Duration
	issuer    string
}

func NewService(secret string, accessTTL time.Duration, issuer string) (*Service, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if accessTTL <= 0 {
		accessTTL = 24 * time.Hour
	}
	if issuer == "" {
		issuer = "classino"
	}
	return &Service{
		secret:    []byte(secret),
		accessTTL: accessTTL,
		issuer:    issuer,
	}, nil
}

func (s *Service) Issue(p IssueParams) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Kind:  p.Kind,
		Role:  p.Role,
		Email: p.Email,
		Name:  p.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(int64(p.SubjectID), 10),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

func (s *Service) Parse(tokenString string) (Claims, error) {
	if tokenString == "" {
		return Claims{}, ErrInvalidToken
	}
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Claims{}, ErrExpiredToken
		}
		return Claims{}, ErrInvalidToken
	}
	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return Claims{}, ErrInvalidToken
	}
	return *claims, nil
}

func (c Claims) SubjectID() (types.ID, error) {
	if c.Subject == "" {
		return 0, ErrInvalidToken
	}
	n, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil || n <= 0 {
		return 0, ErrInvalidToken
	}
	return types.ID(n), nil
}
