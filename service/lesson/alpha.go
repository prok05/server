package lesson

import (
	"bytes"
	"encoding/json"
	"github.com/prok05/ecom/types"
	"io"
	"log"
	"net/http"
	"sync"
)

func GetAlphaLessons(customerID, status, page int, alphaToken, dateFrom, dateTo string, wg *sync.WaitGroup, ch chan []types.GetLessonsResponseItem) {
	defer wg.Done()
	var allLessons []types.GetLessonsResponseItem
	p := page

	for {
		// Формируем тело запроса
		requestBody := types.GetLessonsPayload{
			CustomerID: customerID,
			Status:     status,
			Page:       p,
			DateTo:     dateTo,
			DateFrom:   dateFrom,
		}

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
