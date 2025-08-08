package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

func LmstudioModel(appCtx context.AppContext, messages *types.LMSMessages) (types.LMSModelOutput, error) {
	if messages == nil {
		return types.LMSModelOutput{}, errors.New("invalid messages")
	}

	modelInput := types.LMSModelInput{
		Model:       appCtx.Config.LLM.LLMModel,
		Messages:    *messages,
		Temperature: appCtx.Config.LLM.LLMTemperature,
		MaxTokens:   appCtx.Config.LLM.LLMMaxTokens,
		Stream:      appCtx.Config.LLM.LLMStream,
	}

	var modelOutput types.LMSModelOutput

	requestBody, err := json.Marshal(modelInput)
	if err != nil {
		return modelOutput, fmt.Errorf("failed to create model request body: %v", err.Error())
	}

	llmUrl := fmt.Sprintf("http://%s:%s/v1/chat/completions", appCtx.Config.Common.BaseUrl, appCtx.Config.LLM.LLMPort)
	resp, err := http.Post(llmUrl, "application/json", bytes.NewBuffer(requestBody))
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
