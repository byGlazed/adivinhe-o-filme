package main

import (
	"errors"
	"math/rand"
	"strconv"
	"sync"
)

type Partida struct {
	ID              string
	FilmeSecreto    Filme
	MaxPerguntas    int
	PerguntasFeitas int
	Finalizada      bool
}

var (
	ErrPartidaNaoEncontrada = errors.New("partida não encontrada")
	ErrPartidaFinalizada    = errors.New("partida já finalizada")
	ErrSemPerguntas         = errors.New("limite de perguntas atingido")
)

var (
	partidas  = make(map[string]*Partida)
	mutex     sync.Mutex
	proximoID int
)

func novaPartida(maxPerguntas int) *Partida {
	mutex.Lock()
	defer mutex.Unlock()

	proximoID++
	p := &Partida{
		ID:           strconv.Itoa(proximoID),
		FilmeSecreto: bancoDeFilmes[rand.Intn(len(bancoDeFilmes))],
		MaxPerguntas: maxPerguntas,
	}
	partidas[p.ID] = p
	return p
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

// finalizarPartida confere o palpite e encerra a partida.
func finalizarPartida(id, palpite string) (bool, Filme, error) {
	mutex.Lock()
	defer mutex.Unlock()

	p, ok := partidas[id]
	if !ok {
		return false, Filme{}, ErrPartidaNaoEncontrada
	}
	if p.Finalizada {
		return false, Filme{}, ErrPartidaFinalizada
	}

	p.Finalizada = true
	acertou := normalizar(palpite) == normalizar(p.FilmeSecreto.Titulo)
	return acertou, p.FilmeSecreto, nil
}