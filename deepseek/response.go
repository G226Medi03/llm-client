package deepseek

type response struct {
	Choices []choice  `json:"choices"`
	Error   *apiError `json:"error,omitempty"`
}

type choice struct {
	Message      respMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type respMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type streamResponse struct {
	Choices []streamChoice `json:"choices"`
	Error   *apiError      `json:"error,omitempty"`
}

type streamChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}
