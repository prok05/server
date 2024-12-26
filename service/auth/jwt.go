package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prok05/ecom/config"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"strconv"
	"time"
)

func CreateJWT(secret []byte, userID int, role string) (string, error) {
	expiration := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(userID),
		"expiredAt": time.Now().Add(expiration).Unix(),
		"role":      role,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func WithJWTAuth(handlerFunc http.HandlerFunc, store types.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := GetTokenFromRequest(r)

		token, err := ValidateToken(tokenString)
		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		if !token.Valid {
			log.Println("invalid token")
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		str := claims["userID"].(string)

		userID, _ := strconv.Atoi(str)

		u, err := store.FindUserByID(userID)
		if err != nil {
			log.Printf("failed to get user by id: %v", err)
			permissionDenied(w)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", u.ID)
		r = r.WithContext(ctx)

		handlerFunc(w, r)
	}
}

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

func GetUserIDFromContext(ctx context.Context) int {
	userID := ctx.Value("userID").(int)
	return userID
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}
