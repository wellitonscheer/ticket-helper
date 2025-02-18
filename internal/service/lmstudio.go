package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ModelInput struct {
	Model       string   `json:"model"`
	Messages    Messages `json:"messages"`
	Temperature float32  `json:"temperature"`
	MaxTokens   int      `json:"max_tokens"`
	Stream      bool     `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Messages = []Message

const model = "qwen2.5-7b-instruct-1m"

// const model = "deepseek-r1-distill-qwen-7b"

type ModelOutput struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type Choice struct {
	Index        int     `json:"index"`
	Logprobs     *string `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func LmstudioModel(messages *Messages) (ModelOutput, error) {
	if messages == nil {
		return ModelOutput{}, errors.New("invalid messages")
	}

	modelInput := ModelInput{
		Model:       model,
		Messages:    *messages,
		Temperature: 0.7,
		MaxTokens:   -1,
		Stream:      false,
	}

	requestBody, err := json.Marshal(modelInput)
	if err != nil {
		return ModelOutput{}, fmt.Errorf("failed to create model request body: %v", err.Error())
	}

	resp, err := http.Post("http://localhost:1234/v1/chat/completions", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return ModelOutput{}, fmt.Errorf("error to reach model in lmstudio: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ModelOutput{}, fmt.Errorf("failed to read model response body: %v", err.Error())
	}

	var modelOutput ModelOutput
	if err := json.Unmarshal(body, &modelOutput); err != nil {
		return ModelOutput{}, fmt.Errorf("failed to unmarshal model output: %w, Response body: %s", err, string(body))
	}

	return modelOutput, nil
}
