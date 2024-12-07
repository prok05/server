package lesson

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/cache"
	"github.com/prok05/ecom/service/alpha"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"sync"
)

type Handler struct {
	lessonStore   types.LessonStore
	homeworkStore types.HomeworkStore
	chatStore     types.LessonStore
	tokenCache    *cache.TokenCache
}

func NewHandler(homeworkStore types.HomeworkStore, lessonStore types.LessonStore, tokenCache *cache.TokenCache) *Handler {
	return &Handler{
		lessonStore:   lessonStore,
		homeworkStore: homeworkStore,
		tokenCache:    tokenCache,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/lessons/student", h.handleGetAllLessonsStudent).Methods(http.MethodPost)
	router.HandleFunc("/lessons/homework/student", h.handleGetLessonsHomeworkStudent).Methods(http.MethodPost)
	router.HandleFunc("/lessons/teacher", h.handleGetAllLessonsTeacher).Methods(http.MethodPost)
	router.HandleFunc("/lessons/homework/teacher", h.handleGetLessonsHomeworkTeacher).Methods(http.MethodPost)

	router.HandleFunc("/lessons/rating", h.handleGetLessonRates).Methods(http.MethodGet)
	router.HandleFunc("/lessons/rating", h.handleRateLesson).Methods(http.MethodPost)
}

func (h *Handler) handleGetAllLessonsStudent(w http.ResponseWriter, r *http.Request) {
	var payload types.GetLessonsPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// получение токена alpha CRM
	//alphaToken, err := alpha.GetAlphaToken()
	//if err != nil {
	//	log.Printf("error getting alpha token: %v\n", err)
	//	utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting alpha token"))
	//	return
	//}

	alphaToken, err := h.tokenCache.GetToken()
	if err != nil {
		log.Printf("error getting alpha token: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting alpha token"))
		return
	}

	lessons := []types.GetLessonsResponseItem{}
	lessonsCh := make(chan []types.GetLessonsResponseItem, 3)
	wg := sync.WaitGroup{}
	wg.Add(3)

	go GetAlphaLessons(
		payload.CustomerID,
		1,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"student")
	go GetAlphaLessons(
		payload.CustomerID,
		2,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"student")
	go GetAlphaLessons(
		payload.CustomerID,
		3,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"student")

	wg.Wait()
	close(lessonsCh)
	for v := range lessonsCh {
		lessons = append(lessons, v...)
	}

	// ответ запроса
	utils.WriteJSON(w, http.StatusOK, types.AllFutureLessonsResponse{
		Count: len(lessons),
		Items: lessons,
	})
}

func (h *Handler) handleGetAllLessonsTeacher(w http.ResponseWriter, r *http.Request) {
	var payload types.GetLessonsPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// получение токена alpha CRM
	alphaToken, err := alpha.GetAlphaToken()
	if err != nil {
		log.Printf("error getting alpha token: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting alpha token"))
		return
	}

	lessons := []types.GetLessonsResponseItem{}
	lessonsCh := make(chan []types.GetLessonsResponseItem, 3)
	wg := sync.WaitGroup{}
	wg.Add(3)

	go GetAlphaLessons(
		payload.TeacherID,
		1,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"teacher")
	go GetAlphaLessons(
		payload.TeacherID,
		2,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"teacher")
	go GetAlphaLessons(
		payload.TeacherID,
		3,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"teacher")

	wg.Wait()
	close(lessonsCh)
	for v := range lessonsCh {
		lessons = append(lessons, v...)
	}

	// ответ запроса
	utils.WriteJSON(w, http.StatusOK, types.AllFutureLessonsResponse{
		Count: len(lessons),
		Items: lessons,
	})
}

func (h *Handler) handleGetLessonsHomeworkStudent(w http.ResponseWriter, r *http.Request) {
	var payload types.GetLessonsPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// получение токена alpha CRM
	alphaToken, err := alpha.GetAlphaToken()
	if err != nil {
		log.Printf("error getting alpha token: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting alpha token"))
		return
	}
	lessons := []types.GetLessonsResponseItem{}
	lessonsCh := make(chan []types.GetLessonsResponseItem, 3)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go GetAlphaLessons(
		payload.CustomerID,
		3,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"student")

	wg.Wait()
	close(lessonsCh)
	for v := range lessonsCh {
		lessons = append(lessons, v...)
	}

	lessonIDs := make([]int, len(lessons))
	for i, lesson := range lessons {
		lessonIDs[i] = lesson.ID
	}

	homeworks, err := h.homeworkStore.GetHomeworksByLessonAndStudentID(payload.CustomerID, lessonIDs)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error"))
		return
	}

	for i, lesson := range lessons {
		if homeworkInfo, exists := homeworks[lesson.ID]; exists {
			// Если есть в таблице `homeworks`, берем статус из нее
			lessons[i].HomeworkStatus = homeworkInfo.Status
			lessons[i].HomeworkID = homeworkInfo.ID
		} else {
			// Если нет в таблице, задаем статус 3
			lessons[i].HomeworkStatus = 3
			lessons[i].HomeworkID = nil
		}
	}

	// ответ запроса
	utils.WriteJSON(w, http.StatusOK, types.AllFutureLessonsResponse{
		Count: len(lessons),
		Items: lessons,
	})
}

func (h *Handler) handleGetLessonsHomeworkTeacher(w http.ResponseWriter, r *http.Request) {
	var payload types.GetLessonsPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// получение токена alpha CRM
	alphaToken, err := alpha.GetAlphaToken()
	if err != nil {
		log.Printf("error getting alpha token: %v\n", err)
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting alpha token"))
		return
	}

	lessons := []types.GetLessonsResponseItem{}
	lessonsCh := make(chan []types.GetLessonsResponseItem, 3)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go GetAlphaLessons(
		payload.TeacherID,
		3,
		payload.Page,
		alphaToken,
		payload.DateFrom,
		payload.DateTo,
		&wg,
		lessonsCh,
		"teacher")

	wg.Wait()
	close(lessonsCh)
	for v := range lessonsCh {
		lessons = append(lessons, v...)
	}

	// ответ запроса
	utils.WriteJSON(w, http.StatusOK, types.AllFutureLessonsResponse{
		Count: len(lessons),
		Items: lessons,
	})
}

func (h *Handler) handleRateLesson(w http.ResponseWriter, r *http.Request) {
	var payload types.RateLessonPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		log.Println("handleRateLesson:", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	rateExists, err := h.lessonStore.CheckRateExists(payload.StudentID, payload.TeacherID, payload.LessonID)
	if err != nil {
		log.Println("handleRateLesson:", err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if rateExists {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("lesson rate already exists"))
		return
	} else {
		if err := h.lessonStore.SaveLessonRate(payload.StudentID,
			payload.TeacherID,
			payload.LessonID,
			payload.LessonDate,
			payload.Rate); err != nil {
			log.Println("handleRateLesson:", err)
			utils.WriteError(w, http.StatusInternalServerError, err)
		}
		utils.WriteJSON(w, http.StatusCreated, nil)
	}
}

func (h *Handler) handleGetLessonRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.lessonStore.GetLessonRates()
	if err != nil {
		log.Printf("error getting rates: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, rates)
}
