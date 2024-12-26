package utils

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func StringsToInts(strings []string) ([]int, error) {
	ints := make([]int, 0)

	for _, str := range strings {
		n, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		ints = append(ints, n)
	}
	return ints, nil
}

func WriteFile(fhs *multipart.FileHeader) (string, error) {
	src, err := fhs.Open()
	if err != nil {
		log.Println("error open file")
		return "", err
	}
	defer src.Close()

	ext := filepath.Ext(fhs.Filename)
	fileName := fmt.Sprintf("homework-%d%s", time.Now().UnixNano(), ext)
	dest, err := os.Create("./uploads/" + fileName)
	if err != nil {
		log.Println("error create file", err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		log.Println("error copy file", err)
		return "", err
	}
	return dest.Name(), nil
}
