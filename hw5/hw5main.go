package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type User struct {
	Id       uint64 `json:"id"`
	Username string `json:"username"`
}

//func CreateToken(userId uint64) {
//	atClaims := jwt.MapClaims{
//		"authorized": true,
//		"user_id": userId,
//	}
//	os.Setenv("ACCEPT_SECRET", "awuiefhawjeh")
//	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
//	token, err := at.SigningString([]byte(os.))
//}

func Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieAuth)
		switch err {
		case nil:
		case http.ErrNoCookie:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if c.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		idCtx := context.WithValue(r.Context(), userID, cookieVal(c.Value))

		handler.ServeHTTP(w, r.WithContext(idCtx))
	}
	return http.HandlerFunc(fn)
}

func SendTokenToUser(w http.ResponseWriter, r *http.Request) {

}

func main() {
	server := http.Server{
		Addr:        ":5000",
		Handler:     nil,
		ReadTimeout: time.Second,
	}
	// logger := log.New()
	authRoute := chi.NewRouter()
	authRoute.Use(middleware.RequestID)
	authRoute.Use(middleware.Logger)
	authRoute.Post("/getToken", SendTokenToUser)

	// logger.Info("start authRoute")
	mainRoute := chi.NewRouter()
	mainRoute.Use(middleware.RequestID)
	mainRoute.Use(middleware.Logger)

	log.Fatal(server.ListenAndServe())
}
