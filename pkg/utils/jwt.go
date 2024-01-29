package utils

import (
	"errors"
	"fmt"
	"github.com/bdarge/auth/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/exp/slog"
	"os"
	"time"
)

type JwtWrapper struct {
	TokenSecretKey        string
	Issuer                string
	TokenExpOn            int
	RefreshTokenSecretKey string
	RefreshTokenExpOn     int
}

type SbClaims struct {
	UserId     uint32   `json:"userId"`
	AccountId  uint32   `json:"accountId"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	BusinessId uint32   `json:"businessId"`
	jwt.RegisteredClaims
}

func (w *JwtWrapper) GenerateToken(user models.User, originalRefreshToken *string) (auth *models.Auth, err error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	expireOn := time.Now().Add(time.Minute * time.Duration(w.TokenExpOn))
	slog.Info("generate token", "expires on", expireOn)

	var roles []string
	for _, value := range user.Roles {
		roles = append(roles, value.Name)
	}

	claim := SbClaims{
		user.Account.ID,
		user.ID,
		user.Account.Email,
		roles,
		user.BusinessID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireOn),
			Issuer:    w.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	t, err := token.SignedString([]byte(w.TokenSecretKey))

	if err != nil {
		return nil, err
	}

	if originalRefreshToken != nil {
		return &models.Auth{
			Token:        t,
			RefreshToken: *originalRefreshToken,
		}, nil
	}

	refreshTokenExpOn := time.Now().Add(time.Hour * time.Duration(w.RefreshTokenExpOn))
	slog.Info("refresh token", "expires on", refreshTokenExpOn)

	claim.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(refreshTokenExpOn)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	rt, err := refreshToken.SignedString([]byte(w.RefreshTokenSecretKey))
	if err != nil {
		return nil, err
	}

	return &models.Auth{
		Token:        t,
		RefreshToken: rt,
	}, nil
}

func (w *JwtWrapper) ValidateToken(signedToken string) (claims *SbClaims, err error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Validate token")

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SbClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// validate the alg
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(w.TokenSecretKey), nil
		},
	)

	if err != nil {
		slog.Error(err.Error())
		return nil, errors.New("token has expired or invalid")
	}

	if claims, ok := token.Claims.(*SbClaims); ok && token.Valid {
		slog.Info("Claims found", "claims", claims)
		return claims, nil
	}

	return nil, errors.New("token has expired or invalid")
}

func (w *JwtWrapper) RefreshToken(refreshToken string) (claims *SbClaims, err error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("refresh token helper")

	token, err := jwt.ParseWithClaims(refreshToken, &SbClaims{}, func(token *jwt.Token) (interface{}, error) {
		// validate the alg
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(w.RefreshTokenSecretKey), nil
	})

	if claims, ok := token.Claims.(*SbClaims); ok && token.Valid {
		slog.Info("Claims found", "claims", claims)
		return claims, nil
	}

	slog.Error(err.Error())

	return nil, err
}
