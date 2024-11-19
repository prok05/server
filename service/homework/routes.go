package homework

import (
	"fmt"
	"github.com/gorilla/mux"
	lesson2 "github.com/prok05/ecom/service/lesson"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Handler struct {
	store types.HomeworkStore
}

func NewHandler(store types.HomeworkStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/upload/homework", h.handleUploadHomework).Methods(http.MethodPost)
}

func (h *Handler) handleUploadHomework(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

	// парсинг ключей с формы
	lessonID, err := strconv.Atoi(r.FormValue("lesson_id"))
	if err != nil {
		fmt.Println("wrong lesson_id")
		return
	}
	teacherID, err := strconv.Atoi(r.FormValue("teacher_id"))
	if err != nil {
		fmt.Println("wrong teacher_id")
		return
	}
	studentID, err := strconv.Atoi(r.FormValue("student_id"))
	if err != nil {
		fmt.Println("wrong student_id")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// получение разрешения файла
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		http.Error(w, "Не удалось определить расширение файла", http.StatusBadRequest)
		return
	}

	tempFile, err := os.CreateTemp("../uploads", fmt.Sprintf("homework-*%s", ext))
	if err != nil {
		fmt.Println("createdTemp", err)
		return
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	tempFile.Write(fileBytes)

	err = h.store.SaveHomework(lessonID, studentID, teacherID, tempFile.Name())
	if err != nil {
		fmt.Println("cant save homework")
		return
	}

	err = AlphaUpdateHomeworkStatus(lessonID, 2)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error updating homework"))
		return
	}

	lesson, err := lesson2.GetLessonByIDAlpha(lessonID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting lesson"))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, lesson)
}
