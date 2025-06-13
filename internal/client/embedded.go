package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wellitonscheer/ticket-helper/internal/config"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

type EmbeddingClient struct {
	Config *config.Config
}

func NewEmbeddingClient(appContext context.AppContext) *EmbeddingClient {
	return &EmbeddingClient{
		Config: appContext.Config,
	}
}

type GetTextEmbeddingsInput struct {
	Inputs []string `json:"inputs"`
}

func (e *EmbeddingClient) GetTextEmbeddings(input *GetTextEmbeddingsInput) (types.Embeddings, error) {
	requestBody, err := json.Marshal(*input)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request body: %v", err.Error())
	}

	embeddedUrl := fmt.Sprintf("http://%s:%d/embed", e.Config.Common.BaseUrl, e.Config.Embed.Port)

	resp, err := http.Post(embeddedUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error to get input embedding: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded request body: %v", err.Error())
	}

	var embeddings types.Embeddings
	if err := json.Unmarshal(body, &embeddings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedded inputs: %w, Response body: %s", err, string(body))
	}

	return embeddings, nil
}
