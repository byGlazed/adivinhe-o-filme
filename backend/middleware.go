package main

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Teto global de chamadas à IA (cerca de 10 por minuto, com rajada de 5), para
// proteger a cota gratuita do Gemini mesmo que muita gente jogue ao mesmo tempo.
var limiteGlobalIA = rate.NewLimiter(rate.Every(6*time.Second), 5)

type visitante struct {
	limitador *rate.Limiter
	visto     time.Time
}

// limitadorPorIP guarda um limitador separado para cada visitante.
type limitadorPorIP struct {
	mu         sync.Mutex
	visitantes map[string]*visitante
	limite     rate.Limit
	rajada     int
}

func novoLimitadorPorIP(porSegundo float64, rajada int) *limitadorPorIP {
	l := &limitadorPorIP{
		visitantes: make(map[string]*visitante),
		limite:     rate.Limit(porSegundo),
		rajada:     rajada,
	}
	go l.limparAntigos()
	return l
}

func (l *limitadorPorIP) permitir(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitantes[ip]
	if !ok {
		v = &visitante{limitador: rate.NewLimiter(l.limite, l.rajada)}
		l.visitantes[ip] = v
	}
	v.visto = time.Now()
	return v.limitador.Allow()
}

// limparAntigos esquece visitantes inativos, para o mapa não crescer para sempre.
func (l *limitadorPorIP) limparAntigos() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		for ip, v := range l.visitantes {
			if time.Since(v.visto) > 10*time.Minute {
				delete(l.visitantes, ip)
			}
		}
		l.mu.Unlock()
	}
}

// saltosDeProxy diz quantos proxies confiáveis estão na frente do servidor.
// Lido a cada chamada porque o .env só é carregado depois do início do programa.
func saltosDeProxy() int {
	n, err := strconv.Atoi(os.Getenv("PROXY_HOPS"))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// ipDoCliente descobre o IP do visitante.
// Com PROXY_HOPS=N, usa a N-ésima entrada, contando do fim, do X-Forwarded-For:
// é a que o nosso proxy acrescentou, e o cliente não consegue forjar.
func ipDoCliente(r *http.Request) string {
	if saltos := saltosDeProxy(); saltos > 0 {
		partes := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		if len(partes) >= saltos {
			if ip := strings.TrimSpace(partes[len(partes)-saltos]); ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// limitar embrulha um handler: se o visitante passou do limite, responde 429.
func limitar(l *limitadorPorIP, proximo http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.permitir(ipDoCliente(r)) {
			w.Header().Set("Retry-After", "5")
			http.Error(w, "muitas requisições, espere um pouco e tente de novo", http.StatusTooManyRequests)
			return
		}
		proximo(w, r)
	}
}

// limitarCorpo recusa corpos de requisição maiores que 4 KB.
func limitarCorpo(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
		proximo.ServeHTTP(w, r)
	})
}