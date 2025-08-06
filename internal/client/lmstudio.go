package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/wellitonscheer/ticket-helper/internal/types"
)

const model = "qwen2.5-7b-instruct-1m"

// const model = "deepseek-r1-distill-qwen-7b"

func LmstudioModel(messages *types.LMSMessages) (types.LMSModelOutput, error) {
	if messages == nil {
		return types.LMSModelOutput{}, errors.New("invalid messages")
	}

	modelInput := types.LMSModelInput{
		Model:       model,
		Messages:    *messages,
		Temperature: 0.7,
		MaxTokens:   -1,
		Stream:      false,
	}

	var modelOutput types.LMSModelOutput

	requestBody, err := json.Marshal(modelInput)
	if err != nil {
		return modelOutput, fmt.Errorf("failed to create model request body: %v", err.Error())
	}

	resp, err := http.Post("http://localhost:1234/v1/chat/completions", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return modelOutput, fmt.Errorf("error to reach model in lmstudio: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return modelOutput, fmt.Errorf("failed to read model response body: %v", err.Error())
	}

	if err := json.Unmarshal(body, &modelOutput); err != nil {
		return modelOutput, fmt.Errorf("failed to unmarshal model output: %w, Response body: %s", err, string(body))
	}

	return modelOutput, nil
}
