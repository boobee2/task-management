package service

import (
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"task-management-system/internal/model"
	"task-management-system/internal/repository"
)

var (
	ErrEmailTaken         = errors.New("email sudah digunakan")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrAccountInactive    = errors.New("akun dinonaktifkan")
)

type AuthService struct {
	UserRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{UserRepo: userRepo}
}

func (s *AuthService) Register(name, email, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" {
		return nil, errors.New("nama wajib diisi")
	}

	_, err := s.UserRepo.GetByEmail(email)
	if err == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     "USER", // ditetapkan backend, bukan dari klien
		IsActive: true,
	}
	if err := s.UserRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(email, password string) (string, *model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.UserRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// dicek setelah password benar supaya status akun tidak bocor ke orang lain
	if !user.IsActive {
		return "", nil, ErrAccountInactive
	}

	token, err := generateToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) GetProfile(id int64) (*model.User, error) {
	return s.UserRepo.GetByID(id)
}

func generateToken(user *model.User) (string, error) {
	hours, err := strconv.Atoi(os.Getenv("JWT_EXPIRE_HOURS"))
	if err != nil || hours <= 0 {
		hours = 24
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(hours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
