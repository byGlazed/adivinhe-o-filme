package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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

// debug é lido a cada chamada (e não numa variável global), porque o .env
// só é carregado depois que o programa começa a rodar.
func debug() bool {
	return os.Getenv("DEBUG") == "1"
}

func saudeHandler(w http.ResponseWriter, r *http.Request) {
	escreverJSON(w, http.StatusOK, Resposta{Mensagem: "Backend no ar!"})
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

	filme := sortearFilme()
	p, err := novaPartida(filme, pedido.MaxPerguntas)
	if err != nil {
		http.Error(w, err.Error(), statusDoErro(err))
		return
	}

	if debug() {
		log.Printf("[dev] partida %s: %s (%d), dir. %s", p.ID, filme.Titulo, filme.Ano, filme.Diretor)
	}

	escreverJSON(w, http.StatusCreated, RespostaNovaPartida{ID: p.ID, MaxPerguntas: p.MaxPerguntas})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: arquivo .env não encontrado, usando variáveis do sistema")
	}

	go limparPartidasAntigas()

	// Limites por visitante: requisições por segundo e "rajada" permitida
	limCriar := novoLimitadorPorIP(0.1, 3)     // 1 partida a cada 10 s
	limPerguntar := novoLimitadorPorIP(0.5, 5) // 1 pergunta a cada 2 s
	limPalpite := novoLimitadorPorIP(0.5, 5)
	limBusca := novoLimitadorPorIP(5, 10) // o autocomplete faz várias buscas

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/saude", saudeHandler)
	mux.HandleFunc("POST /api/partidas", limitar(limCriar, criarPartidaHandler))
	mux.HandleFunc("POST /api/partidas/{id}/perguntas", limitar(limPerguntar, perguntarHandler))
	mux.HandleFunc("POST /api/partidas/{id}/palpite", limitar(limPalpite, palpiteHandler))
	mux.HandleFunc("GET /api/filmes/busca", limitar(limBusca, buscarFilmesHandler))
	mux.Handle("/", http.FileServer(http.Dir("../frontend")))

	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	servidor := &http.Server{
		Addr:              ":" + porta,
		Handler:           limitarCorpo(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Servidor rodando em http://localhost:" + porta)
	log.Fatal(servidor.ListenAndServe())
}