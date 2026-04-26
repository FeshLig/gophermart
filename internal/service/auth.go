package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUniqueViolation    = errors.New("unique violation for test")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type AuthService interface {
	Register(ctx context.Context, login string, password string) (string, error)
	Login(ctx context.Context, login string, password string) (string, error)
	GetJWTSecret() string
}

type AuthServiceImpl struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:      repo,
		jwtSecret: generateRandomSecret(),
	}
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("failed to generate JWT secret")
	}
	return base64.StdEncoding.EncodeToString(b)
}

func (s *AuthServiceImpl) GetJWTSecret() string {
	return s.jwtSecret
}

func (s *AuthServiceImpl) Register(ctx context.Context, login string, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := model.User{
		Login:        login,
		PasswordHash: string(hash),
	}

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		if isUniqueViolation(err) {
			return "", ErrUserAlreadyExists
		}
		return "", err
	}

	return s.generateToken(id)
}

func (s *AuthServiceImpl) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.generateToken(user.ID)
}

func (s *AuthServiceImpl) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func isUniqueViolation(err error) bool {
	if err == ErrUniqueViolation {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
