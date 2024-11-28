package lesson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prok05/ecom/service/alpha"
	"github.com/prok05/ecom/types"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func GetAlphaLessons(id, status, page int, alphaToken, dateFrom, dateTo string, wg *sync.WaitGroup, ch chan []types.GetLessonsResponseItem, role string) {
	defer wg.Done()
	var allLessons []types.GetLessonsResponseItem
	p := page

	for {
		var requestBody types.GetLessonsPayload

		if role == "student" {
			requestBody = types.GetLessonsPayload{
				CustomerID: id,
				Status:     status,
				Page:       p,
				DateTo:     dateTo,
				DateFrom:   dateFrom,
			}
		} else if role == "teacher" {
			requestBody = types.GetLessonsPayload{
				TeacherID: id,
				Status:    status,
				Page:      p,
				DateTo:    dateTo,
				DateFrom:  dateFrom,
			}
		}
		// Формируем тело запроса

		// парсинг тела запроса в json
		reqBodyJSON, err := json.Marshal(requestBody)
		if err != nil {
			log.Printf("error marshalling request body: %v", err)
			//return nil, fmt.Errorf("error marshalling request body: %v", err)
		}

		// создание запроса
		req, err := http.NewRequest("POST",
			"https://centriym.s20.online/v2api/1/lesson/index",
			bytes.NewBuffer(reqBodyJSON))
		if err != nil {
			log.Printf("error creating reques: %v", err)
			//return nil, fmt.Errorf("error creating request: %v", err)
		}

		// установка залоговков
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-ALFACRM-TOKEN", alphaToken)

		//  создание клиента и отправка post запроса
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("error sending request: %v", err)
			//return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		// чтение тела запроса
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading response: %v", err)
			//return nil, fmt.Errorf("error reading response: %v", err)
		}

		// создание структуры ответа от сервера и распарсинг ответа
		var lessonsResponse types.GetLessonsResponse
		if err := json.Unmarshal(body, &lessonsResponse); err != nil {
			log.Printf("error parsing response: %v", err)
			//return nil, fmt.Errorf("error parsing response: %v", err)
		}

		// добавление уроков во все уроки путем распаковки lessonsResponse.Items
		allLessons = append(allLessons, lessonsResponse.Items...)

		// условие, чтобы остановить цикл
		if len(lessonsResponse.Items) == 0 || lessonsResponse.Count == 0 {
			break
		}
		p++
	}
	ch <- allLessons
}

func GetLessonByIDAlpha(lessonID int) (*types.GetLessonsResponseItem, error) {
	token, err := alpha.GetAlphaToken()
	if err != nil {
		return nil, fmt.Errorf("error getting alha token")
	}

	data := map[string]any{
		"id": lessonID,
	}

	var lesson types.GetLessonsResponseItem

	reqBodyJSON, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshalling request body: %v", err)
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	req, err := http.NewRequest("POST",
		"https://centriym.s20.online/v2api/1/lesson/index",
		bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		log.Printf("error creating reques: %v", err)
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ALFACRM-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
		//return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response: %v", err)
		//return nil, fmt.Errorf("error reading response: %v", err)
	}

	if err := json.Unmarshal(body, &lesson); err != nil {
		log.Printf("error parsing response: %v", err)
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &lesson, nil
}

func GetStudentTeachersIDs(userID, status, page int, alphaToken string) ([]int, error) {
	p := page

	now := time.Now().UTC()
	end := now.Add(30 * (-24 * time.Hour))
	nowstr := now.Format("2006-01-02")
	endstr := end.Format("2006-01-02")

	var allLessons []types.GetLessonsResponseItem
	for {
		//requestBody := types.GetLessonsPayload{
		//	CustomerID: userID,
		//	Status:     status,
		//	Page:       p,
		//	DateTo:     nowstr,
		//	DateFrom:   endstr,
		//}

		bodyMap := map[string]any{
			"customer_id": userID,
			"status":      status,
			"page":        p,
			"date_from":   endstr,
			"date_to":     nowstr,
		}
		// парсинг тела запроса в json
		reqBodyJSON, err := json.Marshal(bodyMap)
		if err != nil {
			log.Printf("error marshalling request body: %v", err)
			//return nil, fmt.Errorf("error marshalling request body: %v", err)
		}

		// создание запроса
		req, err := http.NewRequest("POST",
			"https://centriym.s20.online/v2api/1/lesson/index",
			bytes.NewBuffer(reqBodyJSON))
		if err != nil {
			log.Printf("error creating reques: %v", err)
			//return nil, fmt.Errorf("error creating request: %v", err)
		}

		// установка залоговков
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-ALFACRM-TOKEN", alphaToken)

		//  создание клиента и отправка post запроса
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("error sending request: %v", err)
			//return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		// чтение тела запроса
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading response: %v", err)
			//return nil, fmt.Errorf("error reading response: %v", err)
		}

		// создание структуры ответа от сервера и распарсинг ответа
		var lessonsResponse types.GetLessonsResponse
		if err := json.Unmarshal(body, &lessonsResponse); err != nil {
			log.Printf("error parsing response: %v", err)
			//return nil, fmt.Errorf("error parsing response: %v", err)
		}

		// добавление уроков во все уроки путем распаковки lessonsResponse.Items
		allLessons = append(allLessons, lessonsResponse.Items...)

		// условие, чтобы остановить цикл
		if len(lessonsResponse.Items) == 0 || lessonsResponse.Count == 0 {
			break
		}
		p++
	}
	teachersMap := make(map[int]struct{})
	for _, v := range allLessons {
		teachersMap[v.TeacherIDs[0]] = struct{}{}
	}

	teachersIds := make([]int, 0)
	for id, _ := range teachersMap {
		teachersIds = append(teachersIds, id)
	}

	return teachersIds, nil
}
