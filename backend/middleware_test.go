package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitadorPorIPSeparaVisitantes(t *testing.T) {
	l := novoLimitadorPorIP(0.0001, 1) // recarrega tão devagar que, no teste, não recarrega

	if !l.permitir("1.1.1.1") {
		t.Error("a primeira requisição deveria passar")
	}
	if l.permitir("1.1.1.1") {
		t.Error("a segunda do mesmo IP deveria ser bloqueada")
	}
	if !l.permitir("2.2.2.2") {
		t.Error("outro IP não pode ser afetado")
	}
}

func TestIPDoClienteIgnoraCabecalhoSemProxyConfigurado(t *testing.T) {
	t.Setenv("PROXY_HOPS", "")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.7:5555"
	req.Header.Set("X-Forwarded-For", "6.6.6.6")

	if ip := ipDoCliente(req); ip != "203.0.113.7" {
		t.Errorf("ip = %q, esperado 203.0.113.7 (o cabeçalho forjado deve ser ignorado)", ip)
	}
}

func TestIPDoClienteAtrasDeUmProxy(t *testing.T) {
	t.Setenv("PROXY_HOPS", "1")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:5555" // o proxy
	// O cliente tentou se passar por 6.6.6.6, e o proxy acrescentou o IP real no fim
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.7")

	if ip := ipDoCliente(req); ip != "203.0.113.7" {
		t.Errorf("ip = %q, esperado 203.0.113.7", ip)
	}
}

func TestLimitarDevolve429(t *testing.T) {
	t.Setenv("PROXY_HOPS", "")
	ok := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	h := limitar(novoLimitadorPorIP(0.0001, 2), ok)

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "203.0.113.7:1000"
		rec := httptest.NewRecorder()
		h(rec, req)

		esperado := http.StatusOK
		if i == 3 {
			esperado = http.StatusTooManyRequests
		}
		if rec.Code != esperado {
			t.Errorf("requisição %d: status %d, esperado %d", i, rec.Code, esperado)
		}
	}
}

func TestLimitarCorpoRejeitaCorpoGrande(t *testing.T) {
	h := limitarCorpo(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]string
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, "inválido", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	grande := `{"pergunta":"` + strings.Repeat("a", 10<<10) + `"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(grande))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, esperado 400 para um corpo de 10 KB", rec.Code)
	}
}