package alpha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prok05/ecom/types"
	"io"
	"log"
	"net/http"
)

// GetUserID Получение ID пользователя по номеру телефона
func GetUserIDByPhone(phone, role, token string) (*types.GetUserResponseItem, error) {
	var url string

	switch role {
	case "teacher", "supervisor":
		url = "https://centriym.s20.online/v2api/1/teacher/index"
	case "student":
		url = "https://centriym.s20.online/v2api/1/customer/index"
	}

	request := map[string]string{
		"phone": phone,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-ALFACRM-TOKEN", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var getUserResponse types.GetUserResponse
	err = json.Unmarshal(body, &getUserResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	if len(getUserResponse.Items) == 0 {
		return nil, fmt.Errorf("no user with such phone: %s", phone)
	}
	return &getUserResponse.Items[0], nil
}

func GetUserById(id int, token string) (*types.GetUserResponseItem, error) {
	url := "https://centriym.s20.online/v2api/1/customer/index"

	request := map[string]int{
		"id": id,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Println(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-ALFACRM-TOKEN", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("received non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	var getUserResponse types.GetUserResponse
	err = json.Unmarshal(body, &getUserResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	if len(getUserResponse.Items) == 0 {
		return nil, fmt.Errorf("no user with such id: %s", id)
	}
	return &getUserResponse.Items[0], nil
}
