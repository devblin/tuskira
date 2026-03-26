package service

import (
	"fmt"
	"time"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *repository.UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

func NewAuthService(repo *repository.UserRepository, secret string, expiry time.Duration) *AuthService {
	return &AuthService{repo: repo, jwtSecret: []byte(secret), jwtExpiry: expiry}
}

func (s *AuthService) Register(email, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:    email,
		Password: string(hash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	return s.generateToken(user.ID)
}

func (s *AuthService) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
