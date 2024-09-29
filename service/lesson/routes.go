package lesson

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/service/alpha"
	"github.com/prok05/ecom/types"
	"github.com/prok05/ecom/utils"
	"log"
	"net/http"
)

func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/lessons/future", handleGetAllFutureLessons).Methods(http.MethodPost)
	//router.HandleFunc("/lessons/{userID}", h.handleRegister).Methods(http.MethodPost)
}

func handleGetAllFutureLessons(w http.ResponseWriter, r *http.Request) {
	var payload types.GetLessonsPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	fmt.Println(payload)

	authToken, err := alpha.GetAlphaToken()
	if err != nil {
		log.Printf("Error getting alpha token: %v\n", err)
		return
	}

	lessons, err := GetAllFutureLessons(payload.CustomerID, payload.Status, payload.Page, authToken)
	if err != nil {
		log.Printf("Error getting future lessons: %v\n", err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.AllFutureLessonsResponse{
		Count: len(lessons),
		Items: lessons,
	})

}
