// Package normalizer provides response normalization utilities
package normalizer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestNormalizeResponse_HappyPath(t *testing.T) {
	t.Parallel()

	// Simulate a typical chat completion response from downstream LLM
	startTime := time.Now().Add(-5 * time.Second)
	responseData := map[string]interface{}{
		"id":    "resp-123",
		"model": "test-model",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello, how can I help you today?",
				},
				"finish_reason": "stop",
			},
		},
	}

	result, err := NormalizeResponse("resp-123", "test-model", startTime, responseData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed: %v", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(result, &resultMap); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	// Verify top-level fields
	if resultMap["id"] != "resp-123" {
		t.Errorf("Expected id='resp-123', got '%v'", resultMap["id"])
	}
	if resultMap["object"] != "chat.completion" {
		t.Errorf("Expected object='chat.completion', got '%v'", resultMap["object"])
	}
	if resultMap["model"] != "test-model" {
		t.Errorf("Expected model='test-model', got '%v'", resultMap["model"])
	}

	// Verify choices array has one element
	choices := resultMap["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(choices))
	}

	choiceMap := choices[0].(map[string]interface{})
	if choiceMap["index"] != float64(0) {
		t.Errorf("Expected index=0, got %v", choiceMap["index"])
	}

	msgMap := choiceMap["message"].(map[string]interface{})
	if msgMap["role"] != "assistant" {
		t.Errorf("Expected role='assistant', got '%v'", msgMap["role"])
	}

	content, ok := msgMap["content"].(string)
	if !ok || content != "Hello, how can I help you today?" {
		t.Errorf("Expected content='Hello...', got %v (type: %T)", msgMap["content"], msgMap["content"])
	}

	if choiceMap["finish_reason"] != "stop" {
		t.Errorf("Expected finish_reason='stop', got '%v'", choiceMap["finish_reason"])
	}
}

func TestNormalizeResponse_MultipleChoices(t *testing.T) {
	t.Parallel()

	startTime := time.Now()
	responseData := map[string]interface{}{
		"model": "test-model",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "First choice response",
				},
				"finish_reason": "stop",
			},
			map[string]interface{}{
				"index": float64(1),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Second choice response",
				},
				"finish_reason": "length",
			},
		},
	}

	result, err := NormalizeResponse("resp-456", "test-model", startTime, responseData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed: %v", err)
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	choices := resultMap["choices"].([]interface{})
	if len(choices) != 2 {
		t.Fatalf("Expected 2 choices, got %d", len(choices))
	}

	// Verify first choice
	firstChoice := choices[0].(map[string]interface{})
	if firstChoice["index"] != float64(0) {
		t.Errorf("First choice index should be 0, got %v", firstChoice["index"])
	}

	// Verify second choice
	secondChoice := choices[1].(map[string]interface{})
	if secondChoice["index"] != float64(1) {
		t.Errorf("Second choice index should be 1, got %v", secondChoice["index"])
	}
	if secondChoice["finish_reason"] != "length" {
		t.Errorf("Second choice finish_reason should be 'length', got '%v'", secondChoice["finish_reason"])
	}
}

func TestNormalizeResponse_StreamingResponse(t *testing.T) {
	t.Parallel()

	startTime := time.Now()
	responseData := map[string]interface{}{
		"model": "streaming-model",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "[streaming response data]",
				},
				"finish_reason": "",
			},
		},
	}

	result, err := NormalizeResponse("stream-resp-123", "streaming-model", startTime, responseData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed: %v", err)
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	choices := resultMap["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("Expected 1 choice for streaming response, got %d", len(choices))
	}

	choice := choices[0].(map[string]interface{})
	msgMap := choice["message"].(map[string]interface{})
	content, _ := msgMap["content"].(string)

	if content != "[streaming response data]" {
		t.Errorf("Streaming content not preserved: got '%s'", content)
	}

	if choice["finish_reason"] != "" {
		t.Errorf("Streaming response should have empty finish_reason, got '%v'", choice["finish_reason"])
	}
}

func TestNormalizeResponse_ErrorCases(t *testing.T) {
	t.Parallel()

	startTime := time.Now()

	// Test 1: Nil input - Marshal(nil) returns "null", which is valid JSON (no error expected)
	result, err := NormalizeResponse("resp-err", "test-model", startTime, nil)
	if err != nil {
		t.Errorf("NormalizeResponse with nil input returned error: %v (expected no error for null)", err)
	} else {
		t.Logf("Got nil result for nil input (expected - Marshal(nil) returns 'null')")
	}

	// Test 2: Partial data - missing choices field (should still work with empty choices)
	partialData := map[string]interface{}{
		"id":    "resp-123",
		"model": "test-model",
		// choices field intentionally omitted
	}

	result, err = NormalizeResponse("resp-123", "test-model", startTime, partialData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed on valid input with missing choices: %v", err)
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	if len(resultMap["choices"].([]interface{})) != 0 {
		t.Errorf("Expected empty choices array for response without choices in input, got %d", len(resultMap["choices"].([]interface{})))
	}
}

func TestNormalizeResponse_EmptyData(t *testing.T) {
	t.Parallel()

	startTime := time.Now()
	responseData := map[string]interface{}{
		"model": "empty-model",
		// choices intentionally omitted
	}

	result, err := NormalizeResponse("resp-empty", "empty-model", startTime, responseData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed: %v", err)
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	if resultMap["id"] != "resp-empty" {
		t.Errorf("Expected id='resp-empty', got '%v'", resultMap["id"])
	}
	if len(resultMap["choices"].([]interface{})) != 0 {
		t.Errorf("Expected empty choices array, got %d", len(resultMap["choices"].([]interface{})))
	}
}

func TestNormalizeResponse_NestedChoicesMap(t *testing.T) {
	t.Parallel()

	startTime := time.Now()

	// Response with deeply nested choices structure
	responseData := map[string]interface{}{
		"model": "nested-model",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "user",
					"content": "Test user message",
				},
				"finish_reason": "tool_calls",
			},
		},
	}

	result, err := NormalizeResponse("resp-nested", "nested-model", startTime, responseData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed: %v", err)
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	choices := resultMap["choices"].([]interface{})
	if len(choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(choices))
	}

	msgMap := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if msgMap["role"] != "user" {
		t.Errorf("Expected role='user', got '%v'", msgMap["role"])
	}

	content, ok := msgMap["content"].(string)
	if !ok || content != "Test user message" {
		t.Errorf("Expected content='Test user message', got %v", msgMap["content"])
	}

	finishReason := choices[0].(map[string]interface{})["finish_reason"].(string)
	if finishReason != "tool_calls" {
		t.Errorf("Expected finish_reason='tool_calls', got '%v'", finishReason)
	}
}

func TestNormalizeResponse_PreservesTimestamp(t *testing.T) {
	t.Parallel()

	expectedTime := time.Now().Add(-10 * time.Second)
	responseData := map[string]interface{}{
		"model": "test-model",
		"choices": []interface{}{
			map[string]interface{}{
				"index": float64(0),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Response content",
				},
				"finish_reason": "stop",
			},
		},
	}

	result, err := NormalizeResponse("resp-timestamp", "test-model", expectedTime, responseData)
	if err != nil {
		t.Fatalf("NormalizeResponse failed: %v", err)
	}

	var resultMap map[string]interface{}
	json.Unmarshal(result, &resultMap)

	created, ok := resultMap["created"].(float64)
	if !ok {
		t.Fatal("Expected 'created' field in response, got nil")
	}

	// Verify timestamp is within expected range (±10 seconds to account for execution time variance)
	createdFloat := float64(expectedTime.Unix())
	if created < createdFloat-10 || created > createdFloat+10 {
		t.Errorf("Timestamp out of expected range: got %d, expected near %d", int64(created), expectedTime.Unix())
	}
}

func BenchmarkNormalizeResponse_Small(b *testing.B) {
	type small struct {
		ID      string   `json:"id"`
		Model   string   `json:"model"`
		Choices []Choice `json:"choices"`
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		responseData := map[string]interface{}{
			"id":    "resp-" + strconv.Itoa(i),
			"model": "test-model",
			"choices": []interface{}{
				map[string]interface{}{
					"index": float64(0),
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Response content for test case " + string(rune(i%10)),
					},
					"finish_reason": "stop",
				},
			},
		}

		_, err := NormalizeResponse("resp-small", "test-model", time.Now(), responseData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNormalizeResponse_Large(b *testing.B) {
	type large struct {
		ID      string   `json:"id"`
		Model   string   `json:"model"`
		Choices []Choice `json:"choices"`
	}

	b.ReportAllocs()

	largeChoices := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		largeChoices[i] = map[string]interface{}{
			"index": float64(i),
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": fmt.Sprintf("Response content for choice %d with longer text to simulate realistic response length", i),
			},
			"finish_reason": "stop",
		}
	}

	for i := 0; i < b.N; i++ {
		responseData := map[string]interface{}{
			"id":      "resp-" + strconv.Itoa(i%100),
			"model":   "large-model",
			"choices": largeChoices,
		}

		_, err := NormalizeResponse("resp-large", "large-model", time.Now(), responseData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
