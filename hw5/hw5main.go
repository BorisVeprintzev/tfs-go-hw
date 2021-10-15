package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const cookieAuth = "auth"
const userID = "ID"

var CountUser uint64 = 0

type Message struct {
	Author  string
	Message string
}

type NewMessage struct {
	Message string `json:"message"`
}

var UserSlice = make([]User, 0)
var GlobalChat = make([]Message, 0)

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
	_, _ = w.Write([]byte("Hello"))
}

func Hello2(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(string)
	if !ok {
		log.Error("Error read context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello2, " + id))
}

func PostPersonalChatHandler(w http.ResponseWriter, r *http.Request) {

}

func GetPersonalChatHandler(w http.ResponseWriter, r *http.Request) {

}

func PostGroupChatHandler(w http.ResponseWriter, r *http.Request) {
	name, ok := r.Context().Value(userID).(string)
	if !ok {
		log.Error("Error read context, PostGroupChatHandler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("Start add new Message. Read body.")
	text, err := io.ReadAll(r.Body)
	log.Info(fmt.Sprintf("body request: %s", string(text)))
	defer r.Body.Close()
	var nMessage NewMessage
	var message Message
	if err != nil {
		log.Error("Can't read body request. PostGroupChatHandler.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(text, &nMessage)
	log.Info(fmt.Sprintf("Read new Message %s", nMessage.Message))
	message.Author = name
	message.Message = nMessage.Message
	GlobalChat = append(GlobalChat, message)
	log.Info(fmt.Sprintf("New Message %+v, was add to global chat.", message))
}

func GetGroupChatHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(userID).(string)
	if !ok {
		log.Error("Error read context. GetGroupChatHandler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("Search global chat messages")
	for _, message := range GlobalChat {
		_, err := w.Write([]byte(fmt.Sprintf("User: %s, Message: %s\n", message.Author, message.Message)))
		if err != nil {
			log.Error("Unexpected error, read global Message")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	log.Info("End send global chat messages")
}

type LoginStruct struct {
	Login string `json:"login"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	text, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error("Can't read body request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var login LoginStruct
	var user User
	err = json.Unmarshal(text, &login)
	if err != nil || login.Login == "" {
		log.Error("Can't read login")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user.Username = login.Login
	user.ID = CountUser
	CountUser++
	c := &http.Cookie{
		Name:  cookieAuth,
		Value: user.Username,
		Path:  "/",
	}
	user.CookieValue = c.Value
	UserSlice = append(UserSlice, user)
	log.Info(fmt.Sprintf("Add new User: %s, summury count users: %d", user.Username, CountUser))
	log.Info(fmt.Sprintf("Add new cookie to user %s", user.Username))
	log.Info(fmt.Sprintf("User %s get new cookie: %s", user.Username, user.CookieValue))
	http.SetCookie(w, c)
}

func Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieAuth)
		switch err {
		case nil:
		case http.ErrNoCookie:
			log.Error("Empty cookie")
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			log.Error("Unexpected error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var u User
		u.ID = 100
		log.Info(fmt.Sprintf("Get cookie:%s", c.Value))
		for _, user := range UserSlice {
			log.Info(fmt.Sprintf("Check user %s, cookie:%s", user.Username, user.CookieValue))
			if user.CookieValue == c.Value {
				log.Info(fmt.Sprintf("User %s found", user.Username))
				u = user
				break
			}
		}
		if u.ID == 100 {
			log.Error("User doesn't found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userCtx := context.WithValue(r.Context(), userID, c.Value)

		handler.ServeHTTP(w, r.WithContext(userCtx))
	}
	return http.HandlerFunc(fn)
}

type User struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	CookieValue string `json:"cookie"`
}

func main() {
	authRoute := chi.NewRouter()
	server := http.Server{
		Addr:        ":4999",
		Handler:     authRoute,
		ReadTimeout: time.Second,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	logger := log.New()

	authRoute.Use(middleware.RequestID)
	authRoute.Use(middleware.Logger)
	authRoute.Post("/login", Login)
	authRoute.Get("/", Hello)

	logger.Info("start authRoute")
	mainRoute := chi.NewRouter()
	mainRoute.Use(middleware.RequestID)
	mainRoute.Use(middleware.Logger)
	mainRoute.Use(Auth)
	mainRoute.Get("/messages", GetGroupChatHandler)
	mainRoute.Post("/messages", PostGroupChatHandler)
	mainRoute.Get("/messages/me", GetPersonalChatHandler)
	mainRoute.Post("/messages/{userName}", PostPersonalChatHandler)
	mainRoute.Get("/hello", Hello2)

	authRoute.Mount("/auth", mainRoute)
	log.Info("Wait requests")
	log.Fatal(server.ListenAndServe())
	<-c
	time.Sleep(3 * time.Second)
}
