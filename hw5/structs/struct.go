package structs

import "sync"

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

func (u *Users) getUsers() []User {
	u.mx.Lock()
	defer u.mx.Unlock()
	return u.Slice
}

func (u *Users) setUsers(users []User) {
	u.mx.Lock()
	defer u.mx.Unlock()
	u.Slice = users
}

func (u *Users) addUser(user User) {
	u.mx.Lock()
	defer u.mx.Unlock()
	u.Slice = append(u.Slice, user)
}

type GlobalChatS struct {
	Slice []Message
	mx    sync.Mutex
}

func (s *GlobalChatS) getGlobalChat() []Message {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.Slice
}

func (s *GlobalChatS) setGlobalChat(messages []Message) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Slice = messages
}

func (s *GlobalChatS) addToGlobalChat(message Message) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Slice = append(s.Slice, message)
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

func (ch *PersonalChatsS) getPersonalChat() map[string][]Message {
	ch.mx.Lock()
	defer ch.mx.Unlock()
	return ch.Map
}

func (ch *PersonalChatsS) setPersonalChat(chat map[string][]Message) {
	ch.mx.Lock()
	defer ch.mx.Unlock()
	ch.Map = chat
}

func (ch *PersonalChatsS) getMessagesFromPersonalChat(name string) []Message {
	ch.mx.Lock()
	defer ch.mx.Unlock()
	for key, value := range ch.Map {
		if name == key {
			return value
		}
	}
	return nil
}

type LoginStruct struct {
	Login string `json:"login"`
}

type User struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	CookieValue string `json:"cookie"`
}
