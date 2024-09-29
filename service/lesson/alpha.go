package lesson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prok05/ecom/types"
	"io"
	"log"
	"net/http"
)

func GetAllFutureLessons(customerID, status, page int, authToken string) ([]types.GetLessonsResponseItem, error) {
	var allLessons []types.GetLessonsResponseItem
	p := page

	for {
		log.Println("Запрос")
		// Формируем тело запроса
		requestBody := types.GetLessonsPayload{
			CustomerID: customerID,
			Status:     status,
			Page:       p,
		}
		reqBodyJSON, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("error marshalling request body: %v", err)
		}

		req, err := http.NewRequest("POST",
			"https://centriym.s20.online/v2api/1/lesson/index",
			bytes.NewBuffer(reqBodyJSON))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-ALFACRM-TOKEN", authToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		var lessonsResponse types.GetLessonsResponse
		if err := json.Unmarshal(body, &lessonsResponse); err != nil {
			return nil, fmt.Errorf("error parsing response: %v", err)
		}

		allLessons = append(allLessons, lessonsResponse.Items...)

		if len(lessonsResponse.Items) == 0 || lessonsResponse.Count == 0 {
			break
		}
		fmt.Println(len(lessonsResponse.Items))
		p++
	}
	return allLessons, nil
}
