package types

import "time"

type UserStore interface {
	FindUserByEmail(email string) (*User, error)
	FindUserByPhone(phone string) (*User, error)
	FindUserByID(id int) (*User, error)
	CreateUser(User) error
}

type MessageStore interface {
	SaveMessage(message *Message) error
	GetMessages(chatID, limit, offset int) ([]*Message, error)
	IsUserInChat(chatID int, userID int) (bool, error)
}

type ChatStore interface {
	CreateChat(chat *Chat) error
	GetAllChats() ([]*Chat, error)
	GetChatByID(chatID int) (*Chat, error)
	DeleteChat(chatID int) error
}

type User struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	MiddleName string    `json:"middleName"`
	Phone      string    `json:"phone"`
	Password   string    `json:"-"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Message struct {
	ID        int       `json:"id"`
	ChatID    int       `json:"chat_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Chat struct {
	ID        int       `json:"id"`
	ChatType  string    `json:"chat_type"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterUserPayload struct {
	Phone    string `json:"phone" validate:"required"`
	Role     string `json:"role" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginUserPayload struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Alpha CRM
type AlphaAuthRequest struct {
	Email  string `json:"email"`
	APIKey string `json:"api_key"`
}

type AlphaAuthResponse struct {
	Token string `json:"token"`
}

type GetUserResponse struct {
	Total int                   `json:"total"`
	Count int                   `json:"count"`
	Page  int                   `json:"page"`
	Items []GetUserResponseItem `json:"items"`
}

type GetUserResponseItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
}

type GetLessonsPayload struct {
	CustomerID int `json:"customer_id" validate:"required"`
	Status     int `json:"status" validate:"required"`
	Page       int `json:"page"`
}

type GetLessonsResponse struct {
	Total int                      `json:"total"`
	Count int                      `json:"count"`
	Page  int                      `json:"page"`
	Items []GetLessonsResponseItem `json:"items"`
}

type GetLessonsResponseItem struct {
	ID         int      `json:"id"`
	LessonType int      `json:"lesson_type_id"`
	Date       string   `json:"date"`
	TimeFrom   string   `json:"time_from"`
	TimeTo     string   `json:"time_to"`
	SubjectID  int      `json:"subject_id"`
	RoomID     int      `json:"room_id"`
	TeacherIDs []int    `json:"teacher_ids"`
	Streaming  []string `json:"streaming"`
}

type AllFutureLessonsResponse struct {
	Count int                      `json:"count"`
	Items []GetLessonsResponseItem `json:"items"`
}
