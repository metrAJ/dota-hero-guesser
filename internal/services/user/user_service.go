package user

import (
	"context"
	"doproj/internal/auth"
	"doproj/internal/models"
	"errors"
	"regexp"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type UserRepository interface {
	GetTopPlayers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
}

type userService struct {
	repo         UserRepository
	tokenManager *auth.TokenManager
}

func NewUserService(repo UserRepository, tm *auth.TokenManager) *userService {
	return &userService{
		repo:         repo,
		tokenManager: tm,
	}
}

// Ger User struct by username
func (s *userService) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	if name == "" {
		return nil, errors.New("username cannot be empty")
	}
	return s.repo.GetUserByUsername(ctx, name)
}

// Get User struct by ID
func (s *userService) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {

	if !validUsername.MatchString(user.Username) {
		return errors.New("Invalid username : must be 3-20 characters and contain only letters, numbers, or underscores")
	}
	if err := validatePasswordComplexity(user.Password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("An error occurred while processing your password")
	}
	user.Password = string(hashedPassword)
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) LoginUser(ctx context.Context, username, plaintextPassword string) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plaintextPassword))
	if err != nil {
		return "", errors.New("invalid username or password")
	}
	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		return "", errors.New("an error occurred while generating your session")
	}
	return token, nil
}

func (s *userService) GetTopPlayers(ctx context.Context) ([]models.User, error) {
	return s.repo.GetTopPlayers(ctx)
}

func validatePasswordComplexity(password string) error {
	// Count runes
	lenght := utf8.RuneCountInString(password)
	if lenght < 8 {
		return errors.New("Password must be at lest 8 characters long")
	}
	if lenght > 72 {
		return errors.New("Password cannot exceed 72 characters")
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
		if hasUpper && hasLower && hasNumber && hasSpecial {
			return nil
		}
	}
	return errors.New("Password must contain at least one number, one uppercase and lowercase letters, one special character")
}
