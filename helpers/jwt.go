package helpers

import (
	"aliagha/config"
	"fmt"
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

func ParseJWTToken(tokenString, secretKey string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token")
		}
		return []byte(secretKey), nil
	})
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	return userID, err
}
