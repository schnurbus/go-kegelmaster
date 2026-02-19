package handlers

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

//go:embed assistant_context.txt
var assistantContext string

const (
	maxChatMessageLength   = 2000
	maxChatHistoryEntries  = 5
	geminiRequestTimeout   = 30 * time.Second
	defaultGeminiModel     = "gemini-2.5-flash"
	geminiMaxRetries       = 1
)

type chatHistoryEntry struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type chatRequest struct {
	Message     string            `json:"message"`
	History     []chatHistoryEntry `json:"history"`
	CurrentPage string            `json:"current_page"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

// Gemini API request/response structures (subset we need).
type geminiGenerateRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// HandleChat handles POST /api/chat: proxy to Gemini with app-only system prompt.
func (h *Handler) HandleChat(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	_, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	if strings.TrimSpace(h.Config.GeminiAPIKey) == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"message": "Der Assistent ist derzeit nicht konfiguriert.",
		})
	}

	var req chatRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Nachricht darf nicht leer sein")
	}
	if len(msg) > maxChatMessageLength {
		return fiber.NewError(fiber.StatusBadRequest, "Nachricht ist zu lang")
	}

	history := req.History
	if len(history) > maxChatHistoryEntries {
		history = history[len(history)-maxChatHistoryEntries:]
	}
	for i := range history {
		r := strings.TrimSpace(history[i].Role)
		if r != "user" && r != "assistant" {
			r = "user"
		}
		history[i].Role = r
	}

	currentPage := strings.TrimSpace(req.CurrentPage)

	reply, err := h.callGemini(ctx, msg, history, currentPage)
	if err != nil {
		slog.Error("gemini chat", "error", err)
		return fiber.NewError(fiber.StatusBadGateway, "Assistent vorübergehend nicht erreichbar")
	}

	return c.JSON(chatResponse{Reply: reply})
}

func (h *Handler) callGemini(ctx context.Context, userMessage string, history []chatHistoryEntry, currentPage string) (string, error) {
	systemPrompt := assistantContext
	if currentPage != "" {
		systemPrompt += "\n\nAktuelle Seite des Nutzers: " + currentPage
	}

	contents := make([]geminiContent, 0, len(history)+1)
	for _, e := range history {
		role := e.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: strings.TrimSpace(e.Text)}},
		})
	}
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: userMessage}},
	})

	body := geminiGenerateRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: contents,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	model := strings.TrimSpace(h.Config.GeminiModel)
	if model == "" {
		model = defaultGeminiModel
	}
	url := "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent?key=" + h.Config.GeminiAPIKey

	var lastErr error
	for attempt := 0; attempt <= geminiMaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: geminiRequestTimeout}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			lastErr = errors.New("gemini API returned " + resp.Status)
			continue
		}

		var geminiResp geminiGenerateResponse
		if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
			_ = resp.Body.Close()
			lastErr = err
			continue
		}
		_ = resp.Body.Close()

		if len(geminiResp.Candidates) == 0 ||
			len(geminiResp.Candidates[0].Content.Parts) == 0 {
			lastErr = errors.New("gemini returned no text")
			continue
		}

		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", lastErr
}
