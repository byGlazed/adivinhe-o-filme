package main

import (
	"math/rand"
	"strconv"
	"sync"
)

type Partida struct {
	ID              string
	FilmeSecreto    Filme
	MaxPerguntas    int
	PerguntasFeitas int
}

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