package homework

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prok05/ecom/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Handler struct {
	store types.HomeworkStore
}

func NewHandler(store types.HomeworkStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/upload/homework", h.handleUploadHomework).Methods(http.MethodPost)
}

func (h *Handler) handleUploadHomework(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(50 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		http.Error(w, "Не удалось определить расширение файла", http.StatusBadRequest)
		return
	}

	tempFile, err := os.CreateTemp("../uploads", fmt.Sprintf("homework-*%s", ext))
	if err != nil {
		fmt.Println("createdTemp", err)
		return
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	tempFile.Write(fileBytes)
}
