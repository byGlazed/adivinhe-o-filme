package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type PedidoPergunta struct {
	Pergunta string `json:"pergunta"`
}

type RespostaPergunta struct {
	Resposta           string `json:"resposta"`
	PerguntasFeitas    int    `json:"perguntas_feitas"`
	PerguntasRestantes int    `json:"perguntas_restantes"`
}

type PedidoPalpite struct {
	Palpite string `json:"palpite"`
}

type RespostaPalpite struct {
	Acertou bool   `json:"acertou"`
	Filme   string `json:"filme"`
}

func escreverJSON(w http.ResponseWriter, status int, dados any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dados)
}

func statusDoErro(err error) int {
	switch {
	case errors.Is(err, ErrPartidaNaoEncontrada):
		return http.StatusNotFound
	case errors.Is(err, ErrPartidaFinalizada), errors.Is(err, ErrSemPerguntas):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func perguntarHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var pedido PedidoPergunta
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil || strings.TrimSpace(pedido.Pergunta) == "" {
		http.Error(w, "envie uma pergunta", http.StatusBadRequest)
		return
	}

	p, err := registrarPergunta(id)
	if err != nil {
		http.Error(w, err.Error(), statusDoErro(err))
		return
	}

	resposta := responderPergunta(p.FilmeSecreto, pedido.Pergunta)

	escreverJSON(w, http.StatusOK, RespostaPergunta{
		Resposta:           resposta,
		PerguntasFeitas:    p.PerguntasFeitas,
		PerguntasRestantes: p.MaxPerguntas - p.PerguntasFeitas,
	})
}

func palpiteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var pedido PedidoPalpite
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil || strings.TrimSpace(pedido.Palpite) == "" {
		http.Error(w, "envie um palpite", http.StatusBadRequest)
		return
	}

	acertou, filme, err := finalizarPartida(id, pedido.Palpite)
	if err != nil {
		http.Error(w, err.Error(), statusDoErro(err))
		return
	}

	escreverJSON(w, http.StatusOK, RespostaPalpite{Acertou: acertou, Filme: filme.Titulo})
}