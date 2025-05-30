package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sinfirst/Ref-System/internal/config"
	"github.com/sinfirst/Ref-System/internal/models"
)

type Claims struct {
	jwt.RegisteredClaims
	UserName string
}

func BuildJWTString(user string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenExp)),
		},
		UserName: user,
	})

	tokenString, err := token.SignedString([]byte(config.SecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims,
			func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(config.SecretKey), nil
			})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctxK := models.CtxKey("userName")
		ctx := context.WithValue(r.Context(), ctxK, claims.UserName)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
