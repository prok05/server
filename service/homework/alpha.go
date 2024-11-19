package homework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prok05/ecom/service/alpha"
	"log"
	"net/http"
)

func AlphaUpdateHomeworkStatus(lessonID, homeworkStatus int) error {
	url := fmt.Sprintf("https://centriym.s20.online/v2api/1/lesson/update?id=%d", lessonID)

	token, err := alpha.GetAlphaToken()
	if err != nil {
		fmt.Println("error token in updatehomeworkstatus")
	}

	data := map[string]any{
		"custom_homework_status": fmt.Sprint(homeworkStatus),
	}

	// парсинг тела запроса в json
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshalling request body: %v", err)
		return fmt.Errorf("error update homework")
	}

	// создание запроса
	req, err := http.NewRequest("POST",
		url,
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error creating reques: %v", err)
		//return nil, fmt.Errorf("error creating request: %v", err)
	}

	// установка залоговков
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ALFACRM-TOKEN", token)

	//  создание клиента и отправка post запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error while updating homework status")
	}

	return nil

	//// чтение тела запроса
	//body, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("error reading response: %v", err)
	//	//return nil, fmt.Errorf("error reading response: %v", err)
	//}
}
