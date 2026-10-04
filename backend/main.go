package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Resposta struct {
	Mensagem string `json:"mensagem"`
}

func saudeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Resposta{Mensagem: "Backend no ar!"})
}

func main() {
	http.HandleFunc("/api/saude", saudeHandler)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}