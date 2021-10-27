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
	"sync"
	"syscall"
	"time"
)

const cookieAuth = "auth"
const userID cookieValue = "ID"

var CountUser uint64 = 0

type Message struct {
	Author  string
	Message string
}

type NewMessage struct {
	Message string `json:"message"`
}

type MessageToPerson struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type cookieValue string

type Users struct {
	Slice []User
	mx    sync.Mutex
}

func NewUsers() Users {
	return Users{
		Slice: make([]User, 0),
	}
}

type GlobalChatS struct {
	Slice []Message
	mx    sync.Mutex
}

func NewGlobalChatS() GlobalChatS {
	return GlobalChatS{
		Slice: make([]Message, 0),
	}
}

// PersonalChatsS храню отправленные юзеру сообщения
type PersonalChatsS struct {
	Map map[string][]Message
	mx  sync.Mutex
}

func NewPersonalChatsS() PersonalChatsS {
	return PersonalChatsS{
		Map: make(map[string][]Message),
	}
}

func Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello"))
}

func Hello2(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieValue)
	if !ok {
		log.Error("Error read context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello2, " + id))
}

func (chat *PersonalChatsS) PostPersonalChatHandler(w http.ResponseWriter, r *http.Request) {
	name, ok := r.Context().Value(userID).(cookieValue)
	if !ok {
		log.Error("Error read context, PostPersonalChatHandler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("User %s, want to send message", name))
	text, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	var message Message
	var messageToPerson MessageToPerson
	if err != nil {
		log.Error("Can't read body request. PostPersonalChatHandler.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(text, &messageToPerson)
	if err != nil {
		log.Error("Unmarshall error PostPersonalChatHandle")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	message.Author = string(name)
	message.Message = messageToPerson.Message
	log.Info(fmt.Sprintf("New message was unmarshaled %s, %s", message.Author, message.Message))
	chat.mx.Lock()
	if chat.Map[messageToPerson.Recipient] != nil {
		_ = append(chat.Map[messageToPerson.Recipient], message)
		log.Info(fmt.Sprintf("New message was delivered to %s", messageToPerson.Recipient))
	} else {
		chat.Map[messageToPerson.Recipient] = []Message{message}
		log.Info(fmt.Sprintf("Create new history of messaging. New message was delivered to %s", messageToPerson.Recipient))
	}
	chat.mx.Unlock()
}

func (chat *PersonalChatsS) GetPersonalChatHandler(w http.ResponseWriter, r *http.Request) {
	name, ok := r.Context().Value(userID).(cookieValue)
	if !ok {
		log.Error("Error read context, GetPersonalChatHandler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("User %s, ask about personal messages", name))
	chat.mx.Lock()
	if len(chat.Map[string(name)]) == 0 {
		_, _ = w.Write([]byte("You haven't messages"))
		log.Info("End GetPersonalChatHandler. No message")
		return
	}
	for _, message := range chat.Map[string(name)] {
		_, _ = w.Write([]byte(fmt.Sprintf("from: %s, message: %s\n", message.Author, message.Message)))
	}
	chat.mx.Unlock()
	log.Info("End GetPersonalChatHandler. have message")
}

func (chat *GlobalChatS) PostGroupChatHandler(w http.ResponseWriter, r *http.Request) {
	name, ok := r.Context().Value(userID).(cookieValue)
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
	if err != nil {
		log.Error("Unmarshall error PostGroupChatHandle")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info(fmt.Sprintf("Read new Message %s", nMessage.Message))
	message.Author = string(name)
	message.Message = nMessage.Message
	chat.mx.Lock()
	chat.Slice = append(chat.Slice, message)
	chat.mx.Unlock()
	log.Info(fmt.Sprintf("New Message %+v, was add to global chat.", message))
}

func (chat *GlobalChatS) GetGroupChatHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(userID).(cookieValue)
	if !ok {
		log.Error("Error read context. GetGroupChatHandler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Info("Search global chat messages")
	chat.mx.Lock()
	for _, message := range chat.Slice {
		_, err := w.Write([]byte(fmt.Sprintf("User: %s, Message: %s\n", message.Author, message.Message)))
		if err != nil {
			log.Error("Unexpected error, read global Message")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	chat.mx.Unlock()
	log.Info("End send global chat messages")
}

type LoginStruct struct {
	Login string `json:"login"`
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
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
	u.mx.Lock()
	u.Slice = append(u.Slice, user)
	u.mx.Unlock()
	log.Info(fmt.Sprintf("Add new User: %s, summury count users: %d", user.Username, CountUser))
	log.Info(fmt.Sprintf("Add new cookie to user %s", user.Username))
	log.Info(fmt.Sprintf("User %s get new cookie: %s", user.Username, user.CookieValue))
	http.SetCookie(w, c)
}

func (us *Users) Auth(handler http.Handler) http.Handler {
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
		flag := false
		log.Info(fmt.Sprintf("Get cookie:%s", c.Value))
		us.mx.Lock()
		for _, user := range us.Slice {
			log.Info(fmt.Sprintf("Check user %s, cookie:%s", user.Username, user.CookieValue))
			if user.CookieValue == c.Value {
				log.Info(fmt.Sprintf("User %s found", user.Username))
				flag = true
				break
			}
		}
		us.mx.Unlock()
		if flag == false {
			log.Error("User doesn't found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userCtx := context.WithValue(r.Context(), userID, cookieValue(c.Value))

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
	_, cancel := context.WithCancel(context.Background())
	authRoute := chi.NewRouter()
	server := http.Server{
		Addr:        ":4999",
		Handler:     authRoute,
		ReadTimeout: time.Second,
	}

	logger := log.New()

	users := NewUsers()
	globalChat := NewGlobalChatS()
	personalChats := NewPersonalChatsS()

	authRoute.Use(middleware.RequestID)
	authRoute.Use(middleware.Logger)
	authRoute.Post("/login", users.Login)
	authRoute.Get("/", Hello)

	logger.Info("start authRoute")
	mainRoute := chi.NewRouter()
	mainRoute.Use(middleware.RequestID)
	mainRoute.Use(middleware.Logger)
	mainRoute.Use(users.Auth)
	mainRoute.Get("/messages", globalChat.GetGroupChatHandler)
	mainRoute.Post("/messages", globalChat.PostGroupChatHandler)
	mainRoute.Get("/messages/me", personalChats.GetPersonalChatHandler)
	mainRoute.Post("/messages/newMessage", personalChats.PostPersonalChatHandler)
	mainRoute.Get("/hello", Hello2)

	authRoute.Mount("/auth", mainRoute)

	c := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	stopAppCh := make(chan struct{})
	go func() {
		log.Info("Get captured signal: ", <-c)
		cancel()

		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("Can't shutdown server")
		}
		stopAppCh <- struct{}{}
	}()

	log.Info("Wait requests")
	log.Fatal(server.ListenAndServe())
	<-stopAppCh
}
