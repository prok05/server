package lesson

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/service/alpha"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
	"sync"
)

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/lessons/future", handleGetAllLessons).Methods(http.MethodPost)
	router.HandleFunc("/lessons/teacher", handleGetAllLessonsTeacher).Methods(http.MethodPost)
}

func handleGetAllLessons(w http.ResponseWriter, r *http.Request) {
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

func handleGetAllLessonsTeacher(w http.ResponseWriter, r *http.Request) {
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
