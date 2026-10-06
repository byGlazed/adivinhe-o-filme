package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var clienteGemini = &http.Client{Timeout: 30 * time.Second}

type gemParte struct {
	Text string `json:"text"`
}

type gemConteudo struct {
	Role  string     `json:"role,omitempty"`
	Parts []gemParte `json:"parts"`
}

type gemPedido struct {
	SystemInstruction *gemConteudo  `json:"systemInstruction,omitempty"`
	Contents          []gemConteudo `json:"contents"`
}

type gemResposta struct {
	Candidates []struct {
		Content gemConteudo `json:"content"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

// perguntarAoGemini manda uma instrução fixa + um texto e devolve o texto da resposta.
func perguntarAoGemini(instrucao, texto string) (string, error) {
	chave := os.Getenv("GEMINI_API_KEY")
	if chave == "" {
		return "", errors.New("variável GEMINI_API_KEY não definida")
	}
	modelo := os.Getenv("GEMINI_MODEL")
	if modelo == "" {
		modelo = "gemini-3.5-flash"
	}

	corpo, err := json.Marshal(gemPedido{
		SystemInstruction: &gemConteudo{Parts: []gemParte{{Text: instrucao}}},
		Contents:          []gemConteudo{{Role: "user", Parts: []gemParte{{Text: texto}}}},
	})
	if err != nil {
		return "", err
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/" + modelo + ":generateContent"
	req, err := http.NewRequest("POST", url, bytes.NewReader(corpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", chave)

	resp, err := clienteGemini.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detalhe, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return "", fmt.Errorf("resposta do Gemini com status %d: %s", resp.StatusCode, strings.TrimSpace(string(detalhe)))
	}

	var r gemResposta
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if len(r.Candidates) == 0 {
		return "", fmt.Errorf("o Gemini não devolveu resposta (bloqueio: %q)", r.PromptFeedback.BlockReason)
	}

	var sb strings.Builder
	for _, parte := range r.Candidates[0].Content.Parts {
		sb.WriteString(parte.Text)
	}
	return strings.TrimSpace(sb.String()), nil
}
