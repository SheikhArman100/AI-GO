package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type HistoryItem struct {
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Timestamp int64  `json:"timestamp"`
}
type AiRequestPayload struct {
	Question string        `json:"question"`
	History  []HistoryItem `json:"history"`
}
type AiResponse struct {
	AnswerDetails    string        `json:"answer_details"`
	RelatedQuestions []string      `json:"related_questions"`
	Images           []string      `json:"images"`
	Charts           []interface{} `json:"charts"`
}

func GetAiResponse(payload AiRequestPayload) (*AiResponse, error) {
	//converts a Go struct or map into JSON format
	requestBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling payload: %v", err)
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}

	//get ai url from env
	aiUrl := os.Getenv("AI_URL")

	// Prepare HTTP POST request
	req, err := http.NewRequest("POST", aiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	// Parse JSON response into a map
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return nil, err
	}

	// Check if the AI request was successful
	success, ok := raw["success"].(bool)
	if !ok || !success {
		log.Printf("AI unsuccessful response: %v", string(body))
		return nil, errors.New("AI service responded with failure")
	}

	// Convert raw response into structured AiResponse
	aiResp := &AiResponse{
		AnswerDetails:    getString(raw["answer_details"]),
		RelatedQuestions: getStringSlice(raw["related_questions"]),
		Images:           getStringSlice(raw["images"]),
		Charts:           getInterfaceSlice(raw["charts"]),
	}

	log.Printf("AI response: %+v\n", aiResp)
	return aiResp, nil
}

// getString safely converts an interface{} to a string
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// getStringSlice safely converts an interface{} to a []string
func getStringSlice(v interface{}) []string {
	var result []string
	if items, ok := v.([]interface{}); ok {
		for _, item := range items {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}

// getInterfaceSlice safely converts an interface{} to []interface{}
func getInterfaceSlice(v interface{}) []interface{} {
	if items, ok := v.([]interface{}); ok {
		return items
	}
	return []interface{}{}
}
