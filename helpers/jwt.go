package helpers

import (
	"aliagha/config"
	"time"

	"github.com/golang-jwt/jwt"
)

type CustomClaims struct {
	UserID    int32  `json:"user_id"`
	Cellphone string `json:"cellphone"`
	jwt.StandardClaims
}

func GenerateJwtToken(userID int32, cellphone string, jwtConfig *config.JWT) (string, error) {
	claims := &CustomClaims{
		UserID:    userID,
		Cellphone: cellphone,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jwtConfig.ExpiresIn).Unix(),
			Issuer:    "Aliagha",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtConfig.SecretKey)

	return tokenString, err
}
