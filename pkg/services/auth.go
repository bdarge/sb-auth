package services

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"

	"github.com/bdarge/auth/out/auth"
	"github.com/bdarge/auth/pkg/db"
	"github.com/bdarge/auth/pkg/models"
	"github.com/bdarge/auth/pkg/utils"
	"golang.org/x/exp/slog"
)

// Server struct
type Server struct {
	auth.UnimplementedAuthServiceServer
	DBHandler db.Handler
	Jwt       utils.JwtWrapper
}

// Register register a user
func (s *Server) Register(_ context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	var acct models.Account
	if result := s.DBHandler.DB.Where(&models.Account{Email: req.Email}).First(&acct); result.Error == nil {
		log.Printf("found: %v", acct)
		return &auth.RegisterResponse{
			Status: http.StatusConflict,
			Error:  "E-Mail already exists",
		}, nil
	}

	acct.Email = req.Email
	acct.Password = utils.HashPassword(req.Password)

	if dbc := s.DBHandler.DB.Create(&models.User{
		Account: acct,
	}); dbc.Error != nil {
		return &auth.RegisterResponse{
			Status: http.StatusInternalServerError,
			Error:  "Failed to create user",
		}, nil
	}

	return &auth.RegisterResponse{
		Status: http.StatusCreated,
	}, nil
}

// Login auth a user
func (s *Server) Login(_ context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	var user models.User

	result := s.DBHandler.DB.Model(&user).
		Preload("Roles").
		Joins("Account").
		Where("email = ?", req.Email).
		First(&user)

	if result.Error != nil {
		log.Printf("User not found for %s", req.Email)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &auth.LoginResponse{
				Status: http.StatusNotFound,
				Error:  "User not found",
			}, nil
		}
		return &auth.LoginResponse{
			Status: http.StatusInternalServerError,
			Error:  "User not found",
		}, nil
	}

	log.Printf("User found %v", user)

	match := utils.CheckPasswordHash(req.Password, user.Account.Password)

	if !match {
		log.Printf("Invalid password for %v", user)
		return &auth.LoginResponse{
			Status: http.StatusForbidden,
			Error:  "User not found",
		}, nil
	}

	authObj, _ := s.Jwt.GenerateToken(user, nil)

	return &auth.LoginResponse{
		Status:       http.StatusOK,
		Token:        authObj.Token,
		RefreshToken: authObj.RefreshToken,
	}, nil
}

// ValidateToken validate a token
func (s *Server) ValidateToken(_ context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	claims, err := s.Jwt.ValidateToken(req.Token)

	if err != nil {
		return &auth.ValidateTokenResponse{
			Status: http.StatusBadRequest,
			Error:  "Failed to validate token",
		}, nil
	}

	var user models.Account

	if result := s.DBHandler.DB.Where(&models.Account{Email: claims.Email}).First(&user); result.Error != nil {
		return &auth.ValidateTokenResponse{
			Status: http.StatusNotFound,
			Error:  "User not found",
		}, nil
	}

	return &auth.ValidateTokenResponse{
		Status: http.StatusOK,
		UserId: int64(user.ID),
	}, nil
}

// RefreshToken refresh a token
func (s *Server) RefreshToken(_ context.Context, req *auth.RefreshTokenRequest) (*auth.LoginResponse, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("refresh token service")

	claims, err := s.Jwt.RefreshToken(req.Token)

	if err != nil {
		slog.Error(err.Error())
		return &auth.LoginResponse{
			Status: http.StatusBadRequest,
			Error:  "Failed to refresh token",
		}, nil
	}

	slog.Info("Found claims", "claims", claims)

	var user models.User

	if s.DBHandler.DB.
		Model(models.User{Model: models.Model{ID: claims.UserId}}).
		Preload("Account").First(&user).Error != nil {
		return &auth.LoginResponse{
			Status: http.StatusForbidden,
			Error:  "Invalid claim attributes",
		}, nil
	}

	// generate a new token
	slog.Info("generate a new token", "user", user)
	authObj, _ := s.Jwt.GenerateToken(user, &req.Token)

	return &auth.LoginResponse{
		Status:       http.StatusOK,
		Token:        authObj.Token,
		RefreshToken: authObj.RefreshToken,
	}, nil
}
