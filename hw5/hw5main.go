package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

const SecretKey = "ABOBA"

type Token string

var CountUser uint64 = 0

type IDMessageType map[int]string
type LoginTokenType map[string]Token
type TokenLoginType map[Token]string
type IDPersonalMessageType map[int]map[int]string

var UserSlice = make([]User, 10)
var IdMessage = make(IDMessageType)
var LoginToken = make(LoginTokenType)
var IdPersonalMessage = make(IDPersonalMessageType)
var TokenLogin = make(TokenLoginType)

//func Auth(handler http.Handler) http.Handler {
//	fn := func(w http.ResponseWriter, r *http.Request) {
//		c, err := r.Cookie(cookieAuth)
//		switch err {
//		case nil:
//		case http.ErrNoCookie:
//			w.WriteHeader(http.StatusUnauthorized)
//			return
//		default:
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//		if c.Value == "" {
//			w.WriteHeader(http.StatusUnauthorized)
//			return
//		}
//		idCtx := context.WithValue(r.Context(), userID, cookieVal(c.Value))
//
//		handler.ServeHTTP(w, r.WithContext(idCtx))
//	}
//	return http.HandlerFunc(fn)
//}

func Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello"))
}

func PostPersonalChatHandler(w http.ResponseWriter, r *http.Request) {

}

func GetPersonalChatHandler(w http.ResponseWriter, r *http.Request) {

}

func PostGroupChatHandler(w http.ResponseWriter, r *http.Request) {

}

func GetGroupChatHandler(w http.ResponseWriter, r *http.Request) {

}

func (u *User) GenerateToken(w http.ResponseWriter) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": u.ID,
		"exp":    time.Now().Add(time.Hour * 106).Unix(),
	})
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		w.Write([]byte("Token give error"))
		return
	}
	u.Token = Token(tokenString)
}

func SendTokenToUser(w http.ResponseWriter, r *http.Request) {
	text, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var user User
	err = json.Unmarshal(text, &user)
	if err != nil || user.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user.ID = CountUser
	CountUser++
	user.GenerateToken(w)
	UserSlice = append(UserSlice, user)
	answer, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshal"))
	}
	w.WriteHeader(http.StatusOK)
	w.Write(answer)
}

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Token    Token  `json:"token"`
}

func main() {
	server := http.Server{
		Addr:        ":4999",
		Handler:     nil,
		ReadTimeout: time.Second,
	}

	logger := log.New()

	authRoute := chi.NewRouter()
	authRoute.Use(middleware.RequestID)
	authRoute.Use(middleware.Logger)
	authRoute.Post("/getToken", SendTokenToUser)
	authRoute.Get("/", Hello)

	logger.Info("start authRoute")
	//mainRoute := chi.NewRouter()
	//mainRoute.Use(middleware.RequestID)
	//mainRoute.Use(middleware.Logger)
	//// mainRoute.Use(Auth)
	//mainRoute.Get("/messages", GetGroupChatHandler)
	//mainRoute.Post("/messages", PostGroupChatHandler)
	//mainRoute.Get("/messages/me", GetPersonalChatHandler)
	//mainRoute.Post("/messages/{userName}", PostPersonalChatHandler)

	log.Info("Wait requests")
	log.Fatal(server.ListenAndServe())
}
