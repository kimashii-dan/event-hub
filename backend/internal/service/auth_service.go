package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Fixsbreaker/event-hub/backend/internal/domain"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/pkg/jwt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Register creates a new user account
func (s *AuthService) Register(req *domain.CreateUserRequest) (*domain.User, error) {
	// 1 Check if email already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// 2 Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3 Create user entity
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         "user", // Default role
		// CreatedAt и UpdatedAt — автоматически от GORM
	}

	// 4 Validate user data
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 5 Save to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns JWT token
func (s *AuthService) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// 1 Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 2 Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 3 Generate JWT token
	token, err := jwt.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 4 Prepare response
	response := &domain.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.jwtExpiry),
		User:      user,
	}

	return response, nil
}

// ValidateToken validates a JWT token and returns user claims
func (s *AuthService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	claims, err := jwt.ValidateToken(tokenString, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

// GetUserByID retrieves user information by ID
func (s *AuthService) GetUserByID(userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// PromoteToOrganizer promotes a user to organizer role
func (s *AuthService) PromoteToOrganizer(userID string) error {
	return s.userRepo.UpdateRole(userID, "organizer")
}

// PromoteToAdmin promotes a user to admin role
func (s *AuthService) PromoteToAdmin(userID string) error {
	return s.userRepo.UpdateRole(userID, "admin")
}
