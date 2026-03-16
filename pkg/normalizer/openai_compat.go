// Package normalizer provides response normalization utilities
package normalizer

import (
	"encoding/json"
	"time"
)

// OpenAICompatibleResponse is the target response format
type OpenAICompatibleResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Choice represents a single chat completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Message represents the response message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NormalizeResponse converts a backend response to OpenAI-compatible format
func NormalizeResponse(id string, model string, startTime time.Time, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		return nil, err
	}

	// Convert to OpenAI format
	response := OpenAICompatibleResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{},
	}

	if choices, ok := parsed["choices"].([]interface{}); ok {
		for _, choice := range choices {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				if message, ok := choiceMap["message"].(map[string]interface{}); ok {
					var content string
					if msgContent, ok := message["content"].(string); ok {
						content = msgContent
					}

					response.Choices = append(response.Choices, Choice{
						Index:        int(choiceMap["index"].(float64)),
						Message:      Message{Role: message["role"].(string), Content: content},
						FinishReason: choiceMap["finish_reason"].(string),
					})
				}
			}
		}
	}

	return json.Marshal(response)
}
