package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Resposta struct {
	Mensagem string `json:"mensagem"`
}

type PedidoNovaPartida struct {
	MaxPerguntas int `json:"max_perguntas"`
}

type RespostaNovaPartida struct {
	ID           string `json:"id"`
	MaxPerguntas int    `json:"max_perguntas"`
}

func saudeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Resposta{Mensagem: "Backend no ar!"})
}

func criarPartidaHandler(w http.ResponseWriter, r *http.Request) {
	var pedido PedidoNovaPartida
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if pedido.MaxPerguntas < 1 || pedido.MaxPerguntas > 50 {
		http.Error(w, "max_perguntas deve estar entre 1 e 50", http.StatusBadRequest)
		return
	}

	p := novaPartida(pedido.MaxPerguntas)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RespostaNovaPartida{ID: p.ID, MaxPerguntas: p.MaxPerguntas})
}

func main() {
	http.HandleFunc("GET /api/saude", saudeHandler)
	http.HandleFunc("POST /api/partidas", criarPartidaHandler)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}