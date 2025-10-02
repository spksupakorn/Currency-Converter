package services

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/internal/models"
	"github.com/spksupakorn/Currency-Converter/internal/repositories"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"github.com/spksupakorn/Currency-Converter/pkg/utils"
)

type AuthService interface {
	Register(email, password string) error
	Login(email, password string) (string, *models.User, error)
	Logout(userID uint) error
	ParseToken(token string) (*jwt.Token, *TokenClaims, error)
}

type authService struct {
	cfg      config.Config
	userRepo repositories.UserRepository
	log      *logger.Logger
}

type TokenClaims struct {
	UserID       uint   `json:"uid"`
	Email        string `json:"email"`
	TokenVersion int    `json:"ver"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg config.Config, userRepo repositories.UserRepository, log *logger.Logger) AuthService {
	return &authService{cfg: cfg, userRepo: userRepo, log: log}
}

func (s *authService) Register(email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return errors.New("email and password are required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}
	_, err := s.userRepo.FindByEmail(email)
	if err == nil {
		return errors.New("email already registered")
	}

	salt, err := utils.GenerateSalt(16)
	if err != nil {
		return errors.New("failed to generate salt")
	}
	hash := utils.HashPasswordArgon2(password, salt)

	u := &models.User{
		Email:        email,
		Password:     string(hash),
		TokenVersion: 0,
	}
	return s.userRepo.Create(u)
}

func (s *authService) Login(email, password string) (string, *models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}
	if !utils.VerifyPasswordArgon2(password, u.Password) {
		return "", nil, errors.New("invalid email or password")
	}
	//*Invalidate previous sessions by incrementing token version
	if err := s.userRepo.IncrementTokenVersion(u.ID); err != nil {
		return "", nil, err
	}
	//*Reload user to get new token version
	u, err = s.userRepo.FindByID(u.ID)
	if err != nil {
		return "", nil, err
	}

	claims := TokenClaims{
		UserID:       u.ID,
		Email:        u.Email,
		TokenVersion: u.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTExpiry)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := tok.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", nil, err
	}
	return ss, u, nil
}

func (s *authService) Logout(userID uint) error {
	return s.userRepo.IncrementTokenVersion(userID)
}

func (s *authService) ParseToken(token string) (*jwt.Token, *TokenClaims, error) {
	claims := &TokenClaims{}
	tok, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil || !tok.Valid {
		return nil, nil, errors.New("invalid token")
	}
	return tok, claims, nil
}

func isValidEmail(s string) bool {
	// Minimal email check
	return strings.Count(s, "@") == 1 && len(s) >= 6 && strings.Contains(s, ".")
}
