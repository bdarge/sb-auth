package utils

import (
	"errors"
	"log"
	"time"

	"github.com/bdarge/auth/pkg/models"
	"github.com/golang-jwt/jwt"
)

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int
}

type jwtClaims struct {
	jwt.StandardClaims
	Id    int64
	Email string
}

func (w *JwtWrapper) GenerateToken(account models.Account) (signedToken string, err error) {
	expireOn := time.Now().Local().Add(time.Minute * time.Duration(w.ExpirationHours)).Unix()
	claims := &jwtClaims{
		Id:    int64(account.ID),
		Email: account.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireOn,
			Issuer:    w.Issuer,
		},
	}

	log.Printf("token expires on: %v", expireOn)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err = token.SignedString([]byte(w.SecretKey))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (w *JwtWrapper) ValidateToken(signedToken string) (claims *jwtClaims, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(w.SecretKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*jwtClaims)

	if !ok {
		return nil, errors.New("couldn't parse claims")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, errors.New("JWT is expired")
	}

	return claims, nil
}
