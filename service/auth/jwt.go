package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prok05/ecom/config"
	"log"
	"net/http"
	"strconv"
	"time"
)

func CreateJWT(secret []byte, userID int) (string, error) {
	expiration := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(userID),
		"expiredAt": time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//func WithJWTAuth(handlerFunc http.HandlerFunc, store)

func GetTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err != nil {
		log.Println(err)
		return ""
	}
	return cookie.Value
}

func ValidateToken(t string) (*jwt.Token, error) {
	return jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.Envs.JWTSecret), nil
	})
}
