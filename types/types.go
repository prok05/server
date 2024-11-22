package types

import "time"

type UserStore interface {
	FindUserByEmail(email string) (*User, error)
	FindUserByPhone(phone string) (*User, error)
	FindUserByID(id int) (*UserDTO, error)
	CreateUser(User) error
	GetAllTeachers() ([]*UserDTO, error)
}

type MessageStore interface {
	SaveMessage(message *Message) error
	GetMessages(chatID, limit, offset int) ([]*Message, error)
	IsUserInChat(chatID int, userID int) (bool, error)
}

type HomeworkStore interface {
	SaveHomework(lessonID, studentID, teacherID int) (int, error)
	SaveHomeworkFile(homeworkID int, filepath string) error
	//GetHomework(homeworkID int) (*Homework, error)
	GetHomeworksByLessonAndStudentID(studentID int, lessonIDs []int) (map[int]*HomeworkInfo, error)
	GetHomeworkFilesByHomeworkID(homeworkID int) ([]HomeworkFile, error)
	DeleteHomeworkFileByID(fileID int) error
	GetHomeworkPathByID(fileID int) (string, error)
}

type ChatStore interface {
	CreateChat(chat *Chat, participants []int64) error
	GetAllChats(userID int) ([]AllChatsItem, error)
	GetChatByID(chatID int) (*Chat, error)
	DeleteChat(chatID int) error
	GetChatParticipants(chatID int) ([]Participant, error)
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

type UserDTO struct {
	ID         int    `json:"id"`
	Phone      string `json:"phone"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Role       string `json:"role"`
}

type Message struct {
	ID        int       `json:"id"`
	ChatID    int       `json:"chat_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Homework struct {
	ID         int       `json:"id"`
	StudentID  int       `json:"student_id"`
	LessonID   int       `json:"lesson_id"`
	TeacherID  int       `json:"teacher_id"`
	FilePath   string    `json:"filepath"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type HomeworkInfo struct {
	ID     *int `json:"id"`
	Status int  `json:"status"`
}

type Chat struct {
	ID        int       `json:"id"`
	ChatType  string    `json:"chat_type"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatMember struct {
	ID       int       `json:"id"`
	ChatID   int       `json:"chat_id"`
	UserID   int       `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type Participant struct {
	UserID    int    `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AllChatsResponse struct {
	Count int            `json:"count"`
	Items []AllChatsItem `json:"items"`
}

type AllChatsItem struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	ChatType     string        `json:"chat_type"`
	LastMessage  Message       `json:"last_message"`
	Participants []Participant `json:"participants"`
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
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Balance         string `json:"balance"`
	PaidLessonCount int    `json:"paid_lesson_count"`
}

type GetLessonsPayload struct {
	CustomerID int    `json:"customer_id"`
	TeacherID  int    `json:"teacher_id"`
	Status     int    `json:"status" validate:"required"`
	Page       int    `json:"page"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

type GetLessonsResponse struct {
	Total int                      `json:"total"`
	Count int                      `json:"count"`
	Page  int                      `json:"page"`
	Items []GetLessonsResponseItem `json:"items"`
}

type GetLessonsResponseItem struct {
	ID             int    `json:"id"`
	Status         int    `json:"status"`
	Date           string `json:"date"`
	TimeFrom       string `json:"time_from"`
	TimeTo         string `json:"time_to"`
	SubjectID      int    `json:"subject_id"`
	RoomID         int    `json:"room_id"`
	TeacherIDs     []int  `json:"teacher_ids"`
	Streaming      any    `json:"streaming"`
	Topic          string `json:"topic"`
	Note           string `json:"note"`
	Homework       any    `json:"homework"`
	HomeworkStatus int    `json:"homework_status"`
	HomeworkID     *int   `json:"homework_id"`
}

type AllFutureLessonsResponse struct {
	Count int                      `json:"count"`
	Items []GetLessonsResponseItem `json:"items"`
}

// HOMEWORK

type HomeworkFile struct {
	ID       int    `json:"id"`
	FilePath string `json:"file_path"`
}
