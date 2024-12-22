package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Input struct {
	Inputs []string `json:"inputs"`
}

type Embeddings [][]float32

func GetTextEmbeddings(input *Input) (Embeddings, error) {
	requestBody, err := json.Marshal(*input)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request body: %v", err.Error())
	}

	resp, err := http.Post("http://127.0.0.1:5000/embed", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error to get input embedding: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded request body: %v", err.Error())
	}

	var embeddings Embeddings
	if err := json.Unmarshal(body, &embeddings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedded inputs: %w, Response body: %s", err, string(body))
	}

	return embeddings, nil
}
