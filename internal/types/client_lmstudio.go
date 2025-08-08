package types

type LMSRoleMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LMSMessages = []LMSRoleMessage

type LMSModelInput struct {
	Model       string      `json:"model"`
	Messages    LMSMessages `json:"messages"`
	Temperature float32     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
	Stream      bool        `json:"stream"`
}

type LMSChoice struct {
	Index        int            `json:"index"`
	Logprobs     *string        `json:"logprobs"`
	FinishReason string         `json:"finish_reason"`
	Message      LMSRoleMessage `json:"message"`
}

type LMSModelOutput struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	Choices           []LMSChoice `json:"choices"`
	Usage             LMSUsage    `json:"usage"`
	SystemFingerprint string      `json:"system_fingerprint"`
}

type LMSUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
