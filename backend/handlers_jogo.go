package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"log"
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
	if len([]rune(pedido.Pergunta)) > 300 {
		http.Error(w, "pergunta muito longa (máximo de 300 caracteres)", http.StatusBadRequest)
		return
	}

	p, err := registrarPergunta(id)
	if err != nil {
		http.Error(w, err.Error(), statusDoErro(err))
		return
	}

	resposta, err := responderPergunta(p.FilmeSecreto, pedido.Pergunta)
	if err != nil {
		log.Println("erro ao consultar a IA:", err)
		desfazerPergunta(id)
		http.Error(w, "a IA não conseguiu responder agora, tente de novo (esta pergunta não foi contada)", http.StatusBadGateway)
		return
	}

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