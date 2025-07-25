package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	appContext "github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

func GetTextEmbeddings(appCtx appContext.AppContext, inputs *types.ClientEmbeddingInputs) (*types.ClientEmbeddings, error) {
	var embeddings types.ClientEmbeddings

	if len(inputs.Inputs) == 0 {
		return &embeddings, fmt.Errorf("client embeddings inputs cannot be empty (inputs=%+v)", inputs)
	}

	for _, input := range inputs.Inputs {
		if input == "" {
			return &embeddings, fmt.Errorf("client embeddings inputs cannot have empty strings (inputs=%+v)", inputs)
		}
	}

	inputsBytes, err := json.Marshal(inputs)
	if err != nil {
		return nil, fmt.Errorf("error to marchal inputs: %v", err)
	}

	embedUrl := fmt.Sprintf("http://%s:%s/embed", appCtx.Config.Common.BaseUrl, appCtx.Config.Embed.Port)

	resp, err := http.Post(embedUrl, "application/json", bytes.NewBuffer(inputsBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch embeddings: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response body: %v", err)
	}

	err = json.Unmarshal(body, &embeddings)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal embeddings body (string(body)=%v): %v", string(body), err)
	}

	return &embeddings, nil
}
