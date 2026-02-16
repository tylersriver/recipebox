package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/tyler/recipebox/internal/application/command"
	"github.com/tyler/recipebox/internal/domain/entity"
	"github.com/tyler/recipebox/internal/domain/repository"
)

// AuthService handles user registration and login.
type AuthService struct {
	userRepo repository.UserRepository
}

// NewAuthService creates a new auth service.
func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, cmd command.RegisterCommand) (*entity.User, error) {
	existing, err := s.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := entity.NewUser(cmd.Email, string(hash))
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login verifies credentials and returns the user.
func (s *AuthService) Login(ctx context.Context, cmd command.LoginCommand) (*entity.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}
