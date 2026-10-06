package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	duracaoMaxPartida = 30 * time.Minute // depois disso a partida é apagada
	maxPartidasAtivas = 5000             // teto de partidas na memória
)

type Partida struct {
	ID              string
	FilmeSecreto    Filme
	MaxPerguntas    int
	PerguntasFeitas int
	Finalizada      bool
	CriadaEm        time.Time
}

var (
	ErrPartidaNaoEncontrada = errors.New("partida não encontrada")
	ErrPartidaFinalizada    = errors.New("partida já finalizada")
	ErrSemPerguntas         = errors.New("limite de perguntas atingido")
	ErrServidorCheio        = errors.New("servidor cheio, tente novamente em instantes")
)

var (
	partidas = make(map[string]*Partida)
	mutex    sync.Mutex
)

// gerarID cria um identificador imprevisível (128 bits aleatórios).
func gerarID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func novaPartida(filme Filme, maxPerguntas int) (*Partida, error) {
	id, err := gerarID()
	if err != nil {
		return nil, err
	}

	mutex.Lock()
	defer mutex.Unlock()

	if len(partidas) >= maxPartidasAtivas {
		return nil, ErrServidorCheio
	}

	p := &Partida{
		ID:           id,
		FilmeSecreto: filme,
		MaxPerguntas: maxPerguntas,
		CriadaEm:     time.Now(),
	}
	partidas[id] = p
	return p, nil
}

// registrarPergunta confere se ainda dá para perguntar e conta a pergunta.
// Devolve uma CÓPIA da partida, para o resto do código usar sem segurar o mutex.
func registrarPergunta(id string) (Partida, error) {
	mutex.Lock()
	defer mutex.Unlock()

	p, ok := partidas[id]
	if !ok {
		return Partida{}, ErrPartidaNaoEncontrada
	}
	if p.Finalizada {
		return Partida{}, ErrPartidaFinalizada
	}
	if p.PerguntasFeitas >= p.MaxPerguntas {
		return Partida{}, ErrSemPerguntas
	}

	p.PerguntasFeitas++
	return *p, nil
}

// desfazerPergunta devolve uma pergunta ao jogador quando a IA falha.
func desfazerPergunta(id string) {
	mutex.Lock()
	defer mutex.Unlock()

	if p, ok := partidas[id]; ok && p.PerguntasFeitas > 0 {
		p.PerguntasFeitas--
	}
}

// finalizarPartida confere o palpite e encerra a partida.
func finalizarPartida(id, palpite string, filmeID int) (bool, Partida, error) {
	mutex.Lock()
	defer mutex.Unlock()

	p, ok := partidas[id]
	if !ok {
		return false, Partida{}, ErrPartidaNaoEncontrada
	}
	if p.Finalizada {
		return false, Partida{}, ErrPartidaFinalizada
	}

	p.Finalizada = true

	var acertou bool
	if filmeID != 0 && p.FilmeSecreto.ID != 0 {
		// Escolheu da lista: compara pelo ID do TMDB
		acertou = filmeID == p.FilmeSecreto.ID
	} else {
		// Digitou à mão (ou o filme é do plano B): compara pelo título
		acertou = normalizar(palpite) == normalizar(p.FilmeSecreto.Titulo) ||
			normalizar(palpite) == normalizar(p.FilmeSecreto.TituloOriginal)
	}
	return acertou, *p, nil
}

// limparPartidasAntigas roda em segundo plano e apaga as partidas expiradas.
func limparPartidasAntigas() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		removidas := 0

		mutex.Lock()
		for id, p := range partidas {
			if time.Since(p.CriadaEm) > duracaoMaxPartida {
				delete(partidas, id)
				removidas++
			}
		}
		ativas := len(partidas)
		mutex.Unlock()

		if removidas > 0 {
			log.Printf("limpeza: %d partida(s) expirada(s) removida(s), %d ativa(s)", removidas, ativas)
		}
	}
}
