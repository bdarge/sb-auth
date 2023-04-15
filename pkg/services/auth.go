package services

import (
	"context"
	"log"
	"net/http"

	"github.com/bdarge/auth/out/auth"
	"github.com/bdarge/auth/pkg/db"
	"github.com/bdarge/auth/pkg/models"
	"github.com/bdarge/auth/pkg/utils"
)

type Server struct {
	auth.UnimplementedAuthServiceServer
	DbAccessHandler db.Handler
	Jwt             utils.JwtWrapper
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	var acct models.Account
	if result := s.DbAccessHandler.DB.Where(&models.Account{Email: req.Email}).First(&acct); result.Error == nil {
		log.Printf("found: %v", acct)
		return &auth.RegisterResponse{
			Status: http.StatusConflict,
			Error:  "E-Mail already exists",
		}, nil
	}

	acct.Email = req.Email
	acct.Password = utils.HashPassword(req.Password)

	if dbc := s.DbAccessHandler.DB.Create(&models.User{
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
	var user models.Account

	if result := s.DbAccessHandler.DB.Where(&models.Account{Email: req.Email}).First(&user); result.Error != nil {
		return &auth.LoginResponse{
			Status: http.StatusNotFound,
			Error:  "User not found",
		}, nil
	}

	match := utils.CheckPasswordHash(req.Password, user.Password)

	if !match {
		return &auth.LoginResponse{
			Status: http.StatusNotFound,
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

	if result := s.DbAccessHandler.DB.Where(&models.Account{Email: claims.Email}).First(&user); result.Error != nil {
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
