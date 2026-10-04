package main

import (
	"errors"
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
  func novaPartida(filme Filme, maxPerguntas int) *Partida {
  	mutex.Lock()
  	defer mutex.Unlock()

  	proximoID++
  	p := &Partida{
  		ID:           strconv.Itoa(proximoID),
  		FilmeSecreto: filme,
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


// desfazerPergunta devolve uma pergunta ao jogador quando a IA falha.
func desfazerPergunta(id string) {
	mutex.Lock()
	defer mutex.Unlock()

	if p, ok := partidas[id]; ok && p.PerguntasFeitas > 0 {
		p.PerguntasFeitas--
	}
}

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
		// Escolheu da lista: compara pelo ID do TMDb
		acertou = filmeID == p.FilmeSecreto.ID
	} else {
		// Digitou à mão (ou o filme é do plano B): compara pelo título
		acertou = normalizar(palpite) == normalizar(p.FilmeSecreto.Titulo) ||
			normalizar(palpite) == normalizar(p.FilmeSecreto.TituloOriginal)
	}
	return acertou, *p, nil
}