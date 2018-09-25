package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func NewContent(message string) ([]byte, error) {
	msg := map[string]interface{}{
		"message": message,
	}

	content, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func WriteResponse(w http.ResponseWriter, content []byte, status int, logger *log.Logger) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(status)
	_, err := w.Write(content)
	if err != nil {
		logger.Printf("[ERR] handler: Failed to write content: %s", err.Error())
	}

	return
}
