package main

import (
	"errors"
	"regexp"
	"testing"
	"time"
)

// reiniciarPartidas esvazia o estado global para um teste não afetar o outro.
func reiniciarPartidas(t *testing.T) {
	t.Helper()
	mutex.Lock()
	partidas = make(map[string]*Partida)
	mutex.Unlock()
}

func TestGerarIDUnicoEImprevisivel(t *testing.T) {
	formato := regexp.MustCompile(`^[0-9a-f]{32}$`)
	vistos := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id, err := gerarID()
		if err != nil {
			t.Fatal(err)
		}
		if !formato.MatchString(id) {
			t.Fatalf("ID com formato inesperado: %q", id)
		}
		if vistos[id] {
			t.Fatalf("ID repetido: %q", id)
		}
		vistos[id] = true
	}
}

func TestLimiteDePerguntasEDevolucao(t *testing.T) {
	reiniciarPartidas(t)
	p, err := novaPartida(Filme{Titulo: "Teste"}, 2)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 2; i++ {
		if _, err := registrarPergunta(p.ID); err != nil {
			t.Fatalf("a pergunta %d deveria ser aceita: %v", i, err)
		}
	}
	if _, err := registrarPergunta(p.ID); !errors.Is(err, ErrSemPerguntas) {
		t.Fatalf("esperado ErrSemPerguntas, veio %v", err)
	}

	desfazerPergunta(p.ID)
	if _, err := registrarPergunta(p.ID); err != nil {
		t.Fatalf("depois de devolver uma pergunta, deveria aceitar: %v", err)
	}
}

func TestPartidaInexistente(t *testing.T) {
	reiniciarPartidas(t)

	if _, err := registrarPergunta("nao-existe"); !errors.Is(err, ErrPartidaNaoEncontrada) {
		t.Errorf("registrarPergunta: esperado ErrPartidaNaoEncontrada, veio %v", err)
	}
	if _, _, err := finalizarPartida("nao-existe", "x", 0); !errors.Is(err, ErrPartidaNaoEncontrada) {
		t.Errorf("finalizarPartida: esperado ErrPartidaNaoEncontrada, veio %v", err)
	}
}

func TestPalpiteComparaPorID(t *testing.T) {
	reiniciarPartidas(t)
	filme := Filme{ID: 42, Titulo: "Django Livre", TituloOriginal: "Django Unchained"}

	certa, _ := novaPartida(filme, 5)
	acertou, _, err := finalizarPartida(certa.ID, "texto qualquer", 42)
	if err != nil || !acertou {
		t.Errorf("com o ID certo deveria acertar (acertou=%v, err=%v)", acertou, err)
	}

	errada, _ := novaPartida(filme, 5)
	acertou, _, err = finalizarPartida(errada.ID, "Django Livre", 7)
	if err != nil || acertou {
		t.Errorf("com o ID errado deveria errar, mesmo com o texto igual (acertou=%v, err=%v)", acertou, err)
	}
}

func TestPalpiteComparaPorTituloSemAcento(t *testing.T) {
	reiniciarPartidas(t)
	filme := Filme{Titulo: "O Rei Leão", TituloOriginal: "The Lion King"} // sem ID, como no plano B

	casos := []struct {
		palpite string
		acerta  bool
	}{
		{"o rei leao", true},
		{"  O REI LEÃO ", true},
		{"The Lion King", true},
		{"Rei Leão", false}, // título parcial não vale
	}

	for _, c := range casos {
		p, _ := novaPartida(filme, 5)
		acertou, _, err := finalizarPartida(p.ID, c.palpite, 0)
		if err != nil {
			t.Fatal(err)
		}
		if acertou != c.acerta {
			t.Errorf("palpite %q: acertou=%v, esperado %v", c.palpite, acertou, c.acerta)
		}
	}
}

func TestPartidaFinalizadaNaoAceitaMais(t *testing.T) {
	reiniciarPartidas(t)
	p, _ := novaPartida(Filme{Titulo: "X"}, 5)
	finalizarPartida(p.ID, "x", 0)

	if _, err := registrarPergunta(p.ID); !errors.Is(err, ErrPartidaFinalizada) {
		t.Errorf("pergunta após o fim: esperado ErrPartidaFinalizada, veio %v", err)
	}
	if _, _, err := finalizarPartida(p.ID, "x", 0); !errors.Is(err, ErrPartidaFinalizada) {
		t.Errorf("segundo palpite: esperado ErrPartidaFinalizada, veio %v", err)
	}
}

func TestServidorCheio(t *testing.T) {
	reiniciarPartidas(t)

	for i := 0; i < maxPartidasAtivas; i++ {
		if _, err := novaPartida(Filme{}, 1); err != nil {
			t.Fatalf("a partida %d deveria caber: %v", i+1, err)
		}
	}
	if _, err := novaPartida(Filme{}, 1); !errors.Is(err, ErrServidorCheio) {
		t.Errorf("esperado ErrServidorCheio, veio %v", err)
	}
}

func TestRemoverExpiradas(t *testing.T) {
	reiniciarPartidas(t)
	velha, _ := novaPartida(Filme{}, 1)
	nova, _ := novaPartida(Filme{}, 1)

	mutex.Lock()
	partidas[velha.ID].CriadaEm = time.Now().Add(-duracaoMaxPartida - time.Minute)
	mutex.Unlock()

	removidas, ativas := removerExpiradas()
	if removidas != 1 || ativas != 1 {
		t.Errorf("removidas=%d ativas=%d, esperado 1 e 1", removidas, ativas)
	}
	if _, err := registrarPergunta(velha.ID); !errors.Is(err, ErrPartidaNaoEncontrada) {
		t.Errorf("a partida velha deveria ter sumido, veio %v", err)
	}
	if _, err := registrarPergunta(nova.ID); err != nil {
		t.Errorf("a partida nova deveria continuar: %v", err)
	}
}