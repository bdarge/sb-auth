package utils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/bdarge/auth/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/exp/slog"
)

// FileReader file reader interface
type FileReader interface {
	ReadFile(filename string) ([]byte, error)	
}

// FileReaderFunc reads a fle
type FileReaderFunc func(filename string) ([]byte, error)

// ReadFile executes the file reader func
func (f FileReaderFunc) ReadFile(filename string) ([]byte, error) {
	return f(filename)
}

// JwtWrapper jwt object
type JwtWrapper struct {
	PrivateKeyPath        string
	Issuer                string
	TokenExpOn            int
	RefreshTokenSecretKey string
	RefreshTokenExpOn     int
	FileReader FileReader
}

// SbClaims claim struct
type SbClaims struct {
	UserID     uint32   `json:"userId"`
	AccountID  uint32   `json:"accountId"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	BusinessID uint32   `json:"businessId"`
	jwt.RegisteredClaims
}

// GenerateToken generate a token
func (w *JwtWrapper) GenerateToken(user models.User, originalRefreshToken *string) (auth *models.Auth, err error) {
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

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claim)

	key, err := w.FileReader.ReadFile(w.PrivateKeyPath)
	if err != nil {
		slog.Error("Failed to read key", "error", err)
		return nil, err
	}
	ecdsaPrivateKey, err := privateKey(key)
	if err != nil {
		slog.Error("Failed to get private key", "error", err)
		return nil, err
	}

	t, err := token.SignedString(ecdsaPrivateKey)

	if err != nil {
		slog.Error("Signing failed", "error", err)
		return nil, err
	}

	if originalRefreshToken != nil {
		return &models.Auth{
			Token:        t,
			RefreshToken: *originalRefreshToken,
		}, nil
	}
	slog.Info("Generate refresh token")
	refreshTokenExpOn := time.Now().Add(time.Hour * time.Duration(w.RefreshTokenExpOn))
	slog.Info("refresh token", "expires on", refreshTokenExpOn)

	claim.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(refreshTokenExpOn)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodES256, claim)

	rt, err := refreshToken.SignedString(ecdsaPrivateKey)
	if err != nil {
		return nil, err
	}

	return &models.Auth{
		Token:        t,
		RefreshToken: rt,
	}, nil
}

// ValidateToken validates a token
func (w *JwtWrapper) ValidateToken(signedToken string) (claims *SbClaims, err error) {
	slog.Info("Validate token")

	key, err := w.FileReader.ReadFile(fmt.Sprintf("%s.pub", w.PrivateKeyPath))
	if err != nil {
		slog.Error("Failed to read public key", "error", err)
		return nil, err
	}

	pk, err := publicKey(key)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(
		signedToken,
		&SbClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// validate the alg
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return pk, nil
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

// RefreshToken refreshes a token
func (w *JwtWrapper) RefreshToken(refreshToken string) (claims *SbClaims, err error) {
	slog.Info("refresh token helper")

	key, err := w.FileReader.ReadFile(fmt.Sprintf("%s.pub", w.PrivateKeyPath))
	if err != nil {
		slog.Error("Failed to read public key", "error", err)
		return nil, err
	}

	pk, err := publicKey(key)
	if err != nil {
		return nil, err
	}


	token, err := jwt.ParseWithClaims(refreshToken, &SbClaims{}, func(token *jwt.Token) (interface{}, error) {
		// validate the alg
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pk, nil
	})

	if claims, ok := token.Claims.(*SbClaims); ok && token.Valid {
		slog.Info("Claims found", "claims", claims)
		return claims, nil
	}

	slog.Error(err.Error())

	return nil, err
}

func publicKey(key []byte) (k *ecdsa.PublicKey, err error){
	block, _ := pem.Decode(key)
	genericPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pk, ok := genericPublicKey.(*ecdsa.PublicKey)

	if !ok {
		return nil, errors.New("Failed to to get ecdsa public key")
	}

	return pk, nil
}

func privateKey(key []byte) (k *ecdsa.PrivateKey, err error) {
	block, _ := pem.Decode(key)

	// Check if it's a private key
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, errors.New("Failed to decode pem block")
	}

	// Get the encoded bytes
	x509Encoded := block.Bytes

	var parsedKey interface{}

	parsedKey, err = x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		return nil, errors.New("Failed to parse private key")
	}

	ecdsaPrivateKey, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("Failed to get ecdsa private key")
	}
	
	return ecdsaPrivateKey, nil
}
