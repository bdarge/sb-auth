package services

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"log"
	"net/http"

	"github.com/bdarge/auth/out/auth"
	"github.com/bdarge/auth/pkg/db"
	"github.com/bdarge/auth/pkg/models"
	"github.com/bdarge/auth/pkg/utils"
)

type Server struct {
	auth.UnimplementedAuthServiceServer
	DBHandler db.Handler
	Jwt       utils.JwtWrapper
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
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

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	var user models.User

	result := s.DBHandler.DB.Model(&user).
		Preload("Roles").
		Joins("Account").
		Where("email = ?", req.Email).
		First(&user)

	if result.Error != nil {
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

	log.Printf("account found %v", user)

	match := utils.CheckPasswordHash(req.Password, user.Account.Password)

	if !match {
		return &auth.LoginResponse{
			Status: http.StatusForbidden,
			Error:  "User not found",
		}, nil
	}

	token, _ := s.Jwt.GenerateToken(user)

	return &auth.LoginResponse{
		Status: http.StatusOK,
		Token:  token,
	}, nil
}

func (s *Server) Validate(ctx context.Context, req *auth.ValidateRequest) (*auth.ValidateResponse, error) {
	claims, err := s.Jwt.ValidateToken(req.Token)

	if err != nil {
		return &auth.ValidateResponse{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		}, nil
	}

	var user models.Account

	if result := s.DBHandler.DB.Where(&models.Account{Email: claims.Email}).First(&user); result.Error != nil {
		return &auth.ValidateResponse{
			Status: http.StatusNotFound,
			Error:  "User not found",
		}, nil
	}

	return &auth.ValidateResponse{
		Status: http.StatusOK,
		UserId: int64(user.ID),
	}, nil
}
