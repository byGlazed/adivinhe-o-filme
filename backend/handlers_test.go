package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// chamar executa um handler direto, sem subir o servidor.
func chamar(h http.HandlerFunc, corpo, id string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/", strings.NewReader(corpo))
	if id != "" {
		req.SetPathValue("id", id)
	}
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

// semRede garante que o teste não usa chaves reais nem acessa TMDB e Gemini.
func semRede(t *testing.T) {
	t.Helper()
	t.Setenv("TMDB_TOKEN", "")
	t.Setenv("GEMINI_API_KEY", "")
	reiniciarPartidas(t)
}

func TestCriarPartidaValidaOLimite(t *testing.T) {
	semRede(t)

	for _, corpo := range []string{`{"max_perguntas": 0}`, `{"max_perguntas": 51}`, `isso não é json`} {
		rec := chamar(criarPartidaHandler, corpo, "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("corpo %q: status %d, esperado 400", corpo, rec.Code)
		}
	}
}

func TestPerguntaCurtaDemais(t *testing.T) {
	semRede(t)

	rec := chamar(perguntarHandler, `{"pergunta": "oi"}`, "qualquer")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, esperado 400", rec.Code)
	}
}

func TestFluxoDaPartidaSemRede(t *testing.T) {
	semRede(t)

	// 1. Cria a partida (sem TMDB, cai no plano B)
	rec := chamar(criarPartidaHandler, `{"max_perguntas": 3}`, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar partida: status %d, corpo %s", rec.Code, rec.Body.String())
	}
	var criada RespostaNovaPartida
	if err := json.NewDecoder(rec.Body).Decode(&criada); err != nil {
		t.Fatal(err)
	}
	if len(criada.ID) != 32 {
		t.Errorf("ID com %d caracteres, esperado 32", len(criada.ID))
	}

	// 2. A IA está indisponível: erro 502 e a pergunta é devolvida
	rec = chamar(perguntarHandler, `{"pergunta": "É uma animação?"}`, criada.ID)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("pergunta sem IA: status %d, esperado 502", rec.Code)
	}
	mutex.Lock()
	feitas := partidas[criada.ID].PerguntasFeitas
	mutex.Unlock()
	if feitas != 0 {
		t.Errorf("a pergunta deveria ter sido devolvida, mas PerguntasFeitas = %d", feitas)
	}

	// 3. Palpite errado: devolve o filme e acertou=false
	rec = chamar(palpiteHandler, `{"palpite": "xyzxyz"}`, criada.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("palpite: status %d, corpo %s", rec.Code, rec.Body.String())
	}
	var resultado RespostaPalpite
	if err := json.NewDecoder(rec.Body).Decode(&resultado); err != nil {
		t.Fatal(err)
	}
	if resultado.Acertou || resultado.Filme == "" {
		t.Errorf("resultado inesperado: %+v", resultado)
	}

	// 4. A partida já acabou
	rec = chamar(palpiteHandler, `{"palpite": "xyzxyz"}`, criada.ID)
	if rec.Code != http.StatusConflict {
		t.Errorf("segundo palpite: status %d, esperado 409", rec.Code)
	}
}