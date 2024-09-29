package alpha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prok05/ecom/config"
	"github.com/prok05/ecom/types"
	"io"
	"net/http"
)

func GetAlphaToken() (string, error) {
	url := "https://centriym.s20.online/v2api/auth/login"

	authRequest := types.AlphaAuthRequest{
		Email:  config.Envs.AlphaEmail,
		APIKey: config.Envs.AlphaApiKey,
	}

	requestBody, err := json.Marshal(authRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-APP-KEY", config.Envs.AlphaXAppKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("received non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var authRespone types.AlphaAuthResponse
	err = json.Unmarshal(body, &authRespone)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return authRespone.Token, nil
}
