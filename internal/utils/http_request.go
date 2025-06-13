package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PostRequestInput struct {
	Url         string
	ContentType string
	Body        any
}

func PostRequest[T any](input *PostRequestInput) (T, error) {
	if input.ContentType == "" {
		input.ContentType = "application/json"
	}

	var embeddings T

	requestBody, err := json.Marshal(input.Body)
	if err != nil {
		return embeddings, fmt.Errorf("failed to create embedding request body: %v", err.Error())
	}

	resp, err := http.Post(input.Url, input.ContentType, bytes.NewBuffer(requestBody))
	if err != nil {
		return embeddings, fmt.Errorf("error to get input embedding: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return embeddings, fmt.Errorf("failed to read embedded request body: %v", err.Error())
	}

	if err := json.Unmarshal(body, &embeddings); err != nil {
		return embeddings, fmt.Errorf("failed to unmarshal embedded inputs: %w, Response body: %s", err, string(body))
	}
}
