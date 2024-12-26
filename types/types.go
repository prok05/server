package types

import (
	"mime/multipart"
	"time"
)

type UserStore interface {
	FindUserByEmail(email string) (*User, error)
	FindUserByPhone(phone string) (*User, error)
	FindUserByID(id int) (*UserDTO, error)
	CreateUser(User) error
	GetAllTeachers() ([]*UserDTO, error)
	GetAllStudents() ([]*UserDTO, error)
	FindUsersByIDs(ids []int) (*[]UserDTO, error)
}

type MessageStore interface {
	SaveMessage(message *Message) (int, error)
	GetMessages(chatID, limit, offset int) ([]*Message, error)
	IsUserInChat(chatID int, userID int) (bool, error)
}

type HomeworkStore interface {
	SaveHomework(lessonID, studentID, teacherID int) (int, error)
	AssignHomework(data HomeworkAssignment) (int, error)
	DeleteHomework(homeworkID int) error
	SaveHomeworkFile(homeworkID int, filepath string) error
	UpdateHomeworkStatus(homeworkID, status int) error
	CountHomeworksWithStatus(lessonID, teacherID, status int) (int, error)
	GetHomeworksByLessonAndStudentID(studentID int, lessonIDs []int) (map[int]*HomeworkInfo, error)
	GetHomeworkFilesByHomeworkID(homeworkID int) ([]HomeworkFile, error)
	DeleteHomeworkFileByID(fileID int) (*int, error)
	GetHomeworkPathByID(fileID int) (string, error)
	GetHomeworksByTeacherAndLessonID(lessonID, teacherID int, studentIDs []int) ([]HomeworkResponse, error)
	GetHomeworkByLessonID(lessonID int) (*Homework, error)
	GetHomeworksByTeacherID(teacherID int) ([]Homework, error)
}

type ChatStore interface {
	CreateChat(chat *Chat, participants []int) error
	GetAllChatsByUserID(userID int) ([]AllChatsItem, error)
	GetChatByID(chatID int) (*Chat, error)
	DeleteChat(chatID int) error
	GetChatParticipants(chatID int) ([]Participant, error)
	GetChatByUserIDs(user1ID, user2ID int) (*Chat, error)
}

type LessonStore interface {
	SaveLessonRate(studentID, teacherID, lessonID int, lessonDate time.Time, rate int8) error
	CheckRateExists(studentID, teacherID, lessonID int) (bool, error)
	GetLessonRates() ([]LessonRate, error)
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
	TeacherID int       `json:"teacher_id"`
	ChatID    int       `json:"chat_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type MessagePayload struct {
	Type    string  `json:"type"`
	UserID  int     `json:"user_id"`
	Message Message `json:"message"`
}

type Homework struct {
	ID               int       `json:"id"`
	LessonID         int       `json:"lesson_id"`
	LessonDate       time.Time `json:"lesson_date"`
	LessonTopic      string    `json:"lesson_topic"`
	TeacherID        int       `json:"teacher_id"`
	SubjectTitle     string    `json:"subject_title"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at_at"`
	UnderReviewCount int       `json:"under_review_count"`
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
	Role            string `json:"role"`
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
	CustomerIDs    []int  `json:"customer_ids"`
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

type LessonRate struct {
	ID                int       `json:"id"`
	StudentID         int       `json:"student_id"`
	StudentFirstName  string    `json:"student_first_name"`
	StudentLastName   string    `json:"student_last_name"`
	StudentMiddleName string    `json:"student_middle_name"`
	TeacherID         int       `json:"teacher_id"`
	TeacherFirstName  string    `json:"teacher_first_name"`
	TeacherLastName   string    `json:"teacher_last_name"`
	TeacherMiddleName string    `json:"teacher_middle_name"`
	LessonID          string    `json:"lesson_id"`
	LessonDate        time.Time `json:"lesson_date"`
	Rate              int8      `json:"rate"`
}

type RateLessonPayload struct {
	StudentID  int       `json:"student_id"`
	TeacherID  int       `json:"teacher_id"`
	LessonID   int       `json:"lesson_id"`
	LessonDate time.Time `json:"lesson_date"`
	Rate       int8      `json:"rate"`
}

// HOMEWORK

type HomeworkPayload struct {
	TeacherID int   `json:"teacher_id"`
	LessonID  int   `json:"lesson_id"`
	Status    int   `json:"status"`
	Students  []int `json:"student_ids"`
}

type HomeworkAssignment struct {
	TeacherID    int
	LessonID     int
	StudentIDs   []int
	SubjectTitle string
	LessonTopic  string
	LessonDate   time.Time
	Description  string
	TeacherFiles []*multipart.FileHeader
}

type UpdateHomeworkPayload struct {
	Status int `json:"status"`
}

type HomeworkResponse struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	HomeworkID     int    `json:"homework_id"`
	HomeworkStatus int    `json:"homework_status"`
	Files          []int  `json:"file_id"`
}

type HomeworkFile struct {
	ID       int    `json:"id"`
	FilePath string `json:"file_path"`
}
