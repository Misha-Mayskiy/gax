package middleware

import (
	"api_gateway/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
		next.ServeHTTP(w, r)
	})
}

type middlewareResponse struct {
	Status_code uint   `json:"status_code"`
	Message     string `json:"message"`
}

func GetUserID(token string) (string, error) {
	jwtPayload, ok := utils.JwtPayloadsFromToken(token)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	userID, ok := jwtPayload["sub"].(string)
	if !ok {
		return "", errors.New("cant find sub from claims")
	}
	return userID, nil
}

func ProtectedHandler(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		token, err := r.Cookie("jwtToken")

		// Вынести во фронт
		if err != nil {
			message := "token not found"
			response := middlewareResponse{Status_code: http.StatusUnauthorized, Message: message}
			data, _ := json.Marshal(response)
			w.Write(data)
			log.Println("authorization error:", message)
			return
		}

		tokenString := token.String()[9:]
		err = utils.VerifyToken(tokenString)
		if err != nil {
			message := "invalid token"
			response := middlewareResponse{Status_code: http.StatusUnauthorized, Message: message}
			data, _ := json.Marshal(response)
			w.Write(data)
			log.Println("authorization error:", message)
			return
		}
		userID, err := GetUserID(tokenString)
		if err != nil {
			message := err.Error()
			response := middlewareResponse{Status_code: http.StatusUnauthorized, Message: message}
			data, _ := json.Marshal(response)
			w.Write(data)
			log.Println("authorization error:", message)
			return
		}
		ctx := context.WithValue(r.Context(), utils.UserIDKey, userID)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
