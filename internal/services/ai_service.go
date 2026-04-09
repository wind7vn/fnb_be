package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wind7vn/fnb_be/pkg/common/errors"
)

type AIService struct {
	apiKey string
}

func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey: apiKey,
	}
}

// Structs for Gemini API payload
type geminiPayload struct {
	Contents []geminiContent `json:"contents"`
	GenerationConfig geminiConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

type geminiConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
}

func (s *AIService) ExtractMenuFromImage(imgBytes []byte, mimeType string) ([]map[string]interface{}, *errors.AppError) {
	base64Str := base64.StdEncoding.EncodeToString(imgBytes)

	promptText := `Bạn là chuyên gia. Hãy trích xuất danh sách các món ăn, danh mục, và giá tiền từ ảnh menu này.
Trả về dữ liệu dưới định dạng JSON array chứa các object:
[
  {
    "name": "Tên món ăn",
    "category": "Danh mục món",
    "price": 50000,
    "description": "Mô tả ngắn"
  }
]
- Giá tiền ghi số nguyên. Nếu không có giá, để 0.
- Trả về ĐÚNG MỘT CẤU TRÚC JSON MẢNG, KHÔNG PHẢI MARKDOWN, KHÔNG CÓ BẤT CỨ ĐỊNH DẠNG HOẶC KÝ TỰ THỪA NÀO KHÁC.`

	payload := geminiPayload{
		GenerationConfig: geminiConfig{
			ResponseMimeType: "application/json",
		},
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{InlineData: &inlineData{MimeType: mimeType, Data: base64Str}},
					{Text: promptText},
				},
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent?key=%s", s.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Request failed", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Network failed", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Gemini Request fail: "+string(bodyBytes), nil)
	}

	// Parse Gemini Response
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Parse failed: "+err.Error(), nil)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return []map[string]interface{}{}, nil // Empty results
	}

	jsonText := result.Candidates[0].Content.Parts[0].Text
	
	var parsedItems []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonText), &parsedItems); err != nil {
		return nil, errors.NewBadRequest(errors.ErrCodeValidationFailed, "Không hiểu được JSON từ AI: "+err.Error(), err)
	}

	return parsedItems, nil
}
