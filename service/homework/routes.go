package homework

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/service/auth"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	store     types.HomeworkStore
	userStore types.UserStore
}

func NewHandler(store types.HomeworkStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/upload/homework", h.handleUploadHomework).Methods(http.MethodPost)
	router.HandleFunc("/upload/homework-add", h.handleAddHomework).Methods(http.MethodPost)
	//router.HandleFunc("/homework/teacher", h.GetHomeworkTeacher).Methods(http.MethodPost)

	// Преподаватель
	router.HandleFunc("/homework/teacher", h.handleAssignHomework).Methods(http.MethodPost)
	router.HandleFunc("/homework/teacher/files/{homeworkID}", h.handleGetHomeworkTeacherFiles).Methods(http.MethodGet)
	router.HandleFunc("/homework/teacher", auth.WithJWTAuth(h.handleGetTeacherHomework, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/homework/teacher/solutions/{homeworkID}", auth.WithJWTAuth(h.handleGetTeacherHomeworkSolutions, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/homework/solution/{solutionID}", auth.WithJWTAuth(h.handleUpdateSolutionStatus, h.userStore)).Methods(http.MethodPatch)
	router.HandleFunc("/homework/teacher/file/{fileID}/download", h.handleDownloadTeacherHomeworkFile).Methods(http.MethodGet)

	// Ученик
	// Получение ДЗ ученика
	router.HandleFunc("/homework/student", auth.WithJWTAuth(h.handleGetStudentHomework, h.userStore)).Methods(http.MethodGet)
	// Отправка решения
	router.HandleFunc("/homework/student/solution", auth.WithJWTAuth(h.handleAssignSolution, h.userStore)).Methods(http.MethodPost)
	// Получение решения
	router.HandleFunc("/homework/student/solution/{homeworkID}", auth.WithJWTAuth(h.handleGetStudentSolution, h.userStore)).Methods(http.MethodGet)

	router.HandleFunc("/homework/{lessonID}", h.handleGetHomework).Methods(http.MethodGet)
	router.HandleFunc("/homework/teacher/count", h.CountHomeworkWithStatus).Methods(http.MethodPost)
	router.HandleFunc("/homework/files/{homeworkID}", h.handleGetHomeworkFiles).Methods(http.MethodGet)
	router.HandleFunc("/homework/files/{fileID}", h.handleDeleteHomeworkFiles).Methods(http.MethodDelete)
	router.HandleFunc("/homework/file/{fileID}/download", h.handleDownloadHomeworkFile).Methods(http.MethodGet)
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

	homeworkID, err := h.store.SaveHomework(lessonID, studentID, teacherID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cannot save homework"))
		log.Println(err)
		return
	}

	err = h.store.SaveHomeworkFile(homeworkID, tempFile.Name())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cannot save homework file"))
		log.Println(err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "ok")
}

func (h *Handler) handleAddHomework(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

	// парсинг ключей с формы
	homeworkID, err := strconv.Atoi(r.FormValue("homework_id"))
	if err != nil {
		fmt.Println("wrong homework_id")
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

	err = h.store.SaveHomeworkFile(homeworkID, tempFile.Name())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cannot save homework file"))
		log.Println(err)
		return
	}

	err = h.store.UpdateSolutionStatus(homeworkID, 2)

	utils.WriteJSON(w, http.StatusCreated, "ok")
}

func (h *Handler) handleUpdateSolutionStatus(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateSolutionPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		log.Println("invalid payload")
		return
	}

	vars := mux.Vars(r)
	str, ok := vars["solutionID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing solution ID"))
		log.Println("missing solutionID")
		return
	}

	solutionID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid solutionID"))
		log.Println("invalid solutionID")
		return
	}

	err = h.store.UpdateSolutionStatus(solutionID, payload.Status)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cannot update homework"))
		log.Println("cannot update homework")
		return
	}

	utils.WriteJSON(w, http.StatusOK, "ok")
}

func (h *Handler) handleGetHomeworkFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["homeworkID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing homework ID"))
		return
	}

	homeworkID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user homeworkID"))
		return
	}

	homeworkFiles, err := h.store.GetHomeworkFilesByHomeworkID(homeworkID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("error getting homework files"))
		return
	}

	fmt.Println(homeworkFiles)

	utils.WriteJSON(w, http.StatusOK, homeworkFiles)
}

func (h *Handler) handleDeleteHomeworkFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["fileID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing file ID"))
		return
	}

	fileID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user fileID"))
		return
	}

	filepath, err := h.store.GetHomeworkFilePathByID(fileID)
	if err != nil {
		fmt.Println("error getting path")
		return
	}

	if err := os.Remove(filepath); err != nil {
		log.Printf("error deleting file: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting file"))
		return
	}

	_, err = h.store.DeleteHomeworkFileByID(fileID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("error getting homework files"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "file deleted successfully"})
}

func (h *Handler) handleDownloadHomeworkFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["fileID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing file ID"))
		return
	}

	fileID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file ID"))
		return
	}

	filepath, err := h.store.GetHomeworkFilePathByID(fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("file not found"))
		} else {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting file path"))
		}
		return
	}

	// Открываем файл

	filepath = strings.ReplaceAll(filepath, "\\", "/")
	filename := path.Base(filepath)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("file not found"))
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error opening file"))
		return
	}
	defer file.Close()

	log.Println("filename:", filename)

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Передаем содержимое файла в ответ
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("error writing file to response: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error writing file to response"))
		return
	}
}

func (h *Handler) handleDownloadTeacherHomeworkFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["fileID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing file ID"))
		return
	}

	fileID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file ID"))
		return
	}

	filepath, err := h.store.GetHomeworkTeacherFilePathByID(fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("file not found"))
		} else {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting file path"))
		}
		return
	}

	// Открываем файл

	filepath = strings.ReplaceAll(filepath, "\\", "/")
	filename := path.Base(filepath)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("file not found"))
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("error opening file: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error opening file"))
		return
	}
	defer file.Close()

	log.Println("filename:", filename)

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Передаем содержимое файла в ответ
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("error writing file to response: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error writing file to response"))
		return
	}
}

func (h *Handler) GetHomeworkTeacher(w http.ResponseWriter, r *http.Request) {
	var payload types.HomeworkPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	homeworks, err := h.store.GetHomeworksByTeacherAndLessonID(payload.LessonID, payload.TeacherID, payload.Students)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	utils.WriteJSON(w, http.StatusOK, homeworks)
}

func (h *Handler) handleAssignHomework(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

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

	students := r.PostForm["student_ids"]
	studentIDs, err := utils.StringsToInts(students)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("wrong student ids"))
		log.Println("wrong studentIDs")
		return
	}

	description := r.FormValue("description")

	subjectTitle := r.FormValue("subject_title")

	lessonDate := r.FormValue("lesson_date")
	lessonDateTime, err := time.Parse(time.RFC3339, lessonDate)
	if err != nil {
		log.Println("error while parsing date", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	lessonTopic := r.FormValue("lesson_topic")

	files := r.MultipartForm.File["files"]

	homeworkData := types.HomeworkAssignment{
		TeacherID:    teacherID,
		LessonID:     lessonID,
		StudentIDs:   studentIDs,
		SubjectTitle: subjectTitle,
		LessonTopic:  lessonTopic,
		Description:  description,
		TeacherFiles: files,
		LessonDate:   lessonDateTime,
	}

	_, err = h.store.AssignHomework(homeworkData)
	if err != nil {
		log.Println("error while assigning homework:", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "ok")
}

func (h *Handler) handleGetHomeworkTeacherFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["homeworkID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing file ID"))
		return
	}

	homeworkID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file ID"))
		return
	}

	files, err := h.store.GetHomeworkTeacherFiles(homeworkID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unable to get homework teacher files"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, files)
}

func (h *Handler) handleAssignSolution(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())

	r.ParseMultipartForm(50 << 20)

	solutionID, err := strconv.Atoi(r.FormValue("solution_id"))
	if err != nil {
		fmt.Println("wrong solution_id")
		return
	}
	homeworkID, err := strconv.Atoi(r.FormValue("homework_id"))
	if err != nil {
		fmt.Println("wrong homework_id")
		return
	}

	solution := r.FormValue("solution")

	files := r.MultipartForm.File["files"]

	solutionData := types.SolutionAssignment{
		SolutionID:    solutionID,
		StudentID:     studentID,
		HomeworkID:    homeworkID,
		Solution:      solution,
		SolutionFiles: files,
	}

	err = h.store.AssignSolution(solutionData)
	if err != nil {
		log.Println("error while assigning solution:", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "ok")
}

func (h *Handler) handleGetTeacherHomework(w http.ResponseWriter, r *http.Request) {
	teacherID := auth.GetUserIDFromContext(r.Context())

	homeworks, err := h.store.GetHomeworksByTeacherID(teacherID)
	if err != nil {
		log.Printf("failed to retrieve homeworks for teacher: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve homeworks"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, homeworks)
}

func (h *Handler) handleGetStudentHomework(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())

	homeworks, err := h.store.GetHomeworksByStudentID(studentID)
	if err != nil {
		log.Printf("failed to retrieve homeworks for student: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve homeworks"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, homeworks)
}

func (h *Handler) handleGetStudentSolution(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())

	vars := mux.Vars(r)
	str, ok := vars["homeworkID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing homeworkID"))
		return
	}

	homeworkID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file ID"))
		return
	}

	solution, err := h.store.GetHomeworkSolutionByStudent(homeworkID, studentID)
	if err != nil {
		log.Printf("failed to retrieve homeworks for student: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve homeworks"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, solution)
}

func (h *Handler) handleGetTeacherHomeworkSolutions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["homeworkID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing homeworkID"))
		return
	}

	homeworkID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file ID"))
		return
	}

	solutions, err := h.store.GetHomeworkSolutions(homeworkID)
	if err != nil {
		log.Printf("unable to get homework solutions: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("unable to get homework solutions"))
		return
	}
	utils.WriteJSON(w, http.StatusOK, solutions)
}

func (h *Handler) handleGetHomework(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["lessonID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing lesson ID"))
		return
	}

	lessonID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid file ID"))
		return
	}
	homework, err := h.store.GetHomeworkByLessonID(lessonID)
	if err != nil {
		log.Println("error while getting homework: ", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, homework)
}

func (h *Handler) CountHomeworkWithStatus(w http.ResponseWriter, r *http.Request) {
	var payload types.HomeworkPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	count, err := h.store.CountHomeworksWithStatus(payload.LessonID, payload.TeacherID, payload.Status)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"count": count})
}
