package homework

import (
	"fmt"
	"github.com/gorilla/mux"
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
)

type Handler struct {
	store types.HomeworkStore
}

func NewHandler(store types.HomeworkStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/upload/homework", h.handleUploadHomework).Methods(http.MethodPost)
	router.HandleFunc("/upload/homework-add", h.handleAddHomework).Methods(http.MethodPost)
	router.HandleFunc("/homework/teacher", h.GetHomeworkTeacher).Methods(http.MethodPost)
	router.HandleFunc("/homework/teacher/count", h.CountHomeworkWithStatus).Methods(http.MethodPost)
	router.HandleFunc("/homework/{homeworkID}", h.handleUpdateHomeworkStatus).Methods(http.MethodPatch)
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

	err = h.store.UpdateHomeworkStatus(homeworkID, 2)

	utils.WriteJSON(w, http.StatusCreated, "ok")
}

func (h *Handler) handleUpdateHomeworkStatus(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateHomeworkPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		log.Println("invalid payload")
		return
	}

	vars := mux.Vars(r)
	str, ok := vars["homeworkID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing homework ID"))
		log.Println("missind homeworkID")
		return
	}

	homeworkID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user homeworkID"))
		log.Println("invalid homeworkID")
		return
	}

	err = h.store.UpdateHomeworkStatus(homeworkID, payload.Status)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cannot update homework"))
		log.Println("cannot update homework")
		return
	}

	//if payload.Status == 1 || payload.Status == 4 {
	//	files, err := h.store.GetHomeworkFilesByHomeworkID(homeworkID)
	//	if err != nil {
	//		log.Println("cant get homework files")
	//		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cant get homework files"))
	//		return
	//	}
	//	for _, v := range files {
	//		_, err = h.store.DeleteHomeworkFileByID(v.ID)
	//		if err != nil {
	//			log.Println(err)
	//		}
	//		if err := os.Remove(v.FilePath); err != nil {
	//			log.Printf("error deleting file: %v\n", err)
	//		}
	//	}
	//}
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

	filepath, err := h.store.GetHomeworkPathByID(fileID)
	if err != nil {
		fmt.Println("error getting path")
		return
	}

	if err := os.Remove(filepath); err != nil {
		log.Printf("error deleting file: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error deleting file"))
		return
	}

	homeworkID, err := h.store.DeleteHomeworkFileByID(fileID)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, fmt.Errorf("error getting homework files"))
		return
	}

	files, err := h.store.GetHomeworkFilesByHomeworkID(*homeworkID)
	if err != nil {
		log.Println(err)
		return
	}

	if len(files) == 0 {
		//err = h.store.UpdateHomeworkStatus(*homeworkID, 3)
		err = h.store.DeleteHomework(*homeworkID)
		if err != nil {
			log.Println(err)
		}
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

	filepath, err := h.store.GetHomeworkPathByID(fileID)
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
