package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var clienteHTTP = &http.Client{Timeout: 10 * time.Second}

type respostaLista struct {
	Results []struct {
		ID int `json:"id"`
	} `json:"results"`
}

type respostaDetalhes struct {
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title"`
	ReleaseDate   string `json:"release_date"`
	Overview      string `json:"overview"`
	Runtime       int    `json:"runtime"`
	Genres        []struct {
		Name string `json:"name"`
	} `json:"genres"`
	Credits struct {
		Cast []struct {
			Name string `json:"name"`
		} `json:"cast"`
		Crew []struct {
			Name string `json:"name"`
			Job  string `json:"job"`
		} `json:"crew"`
	} `json:"credits"`
}

// tmdbGet faz um GET na API do TMDb e decodifica o JSON em "destino".
func tmdbGet(caminho string, destino any) error {
	token := os.Getenv("TMDB_TOKEN")
	if token == "" {
		return errors.New("variável TMDB_TOKEN não definida")
	}

	req, err := http.NewRequest("GET", "https://api.themoviedb.org/3"+caminho, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := clienteHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TMDb respondeu com status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(destino)
}

// sortearFilmeTMDb escolhe um filme entre os 100 mais votados (os mais famosos)
// e busca os detalhes completos dele.
func sortearFilmeTMDb() (Filme, error) {
	pagina := rand.Intn(5) + 1

	var lista respostaLista
	caminho := fmt.Sprintf("/discover/movie?language=pt-BR&sort_by=vote_count.desc&include_adult=false&page=%d", pagina)
	if err := tmdbGet(caminho, &lista); err != nil {
		return Filme{}, err
	}
	if len(lista.Results) == 0 {
		return Filme{}, errors.New("TMDb devolveu uma lista vazia")
	}
	id := lista.Results[rand.Intn(len(lista.Results))].ID

	var d respostaDetalhes
	caminho = fmt.Sprintf("/movie/%d?language=pt-BR&append_to_response=credits", id)
	if err := tmdbGet(caminho, &d); err != nil {
		return Filme{}, err
	}

	f := Filme{
		Titulo:         d.Title,
		TituloOriginal: d.OriginalTitle,
		Sinopse:        d.Overview,
		DuracaoMin:     d.Runtime,
	}

	if len(d.ReleaseDate) >= 4 {
		f.Ano, _ = strconv.Atoi(d.ReleaseDate[:4])
	}
	for _, g := range d.Genres {
		f.Generos = append(f.Generos, g.Name)
	}
	for _, c := range d.Credits.Crew {
		if c.Job == "Director" {
			f.Diretor = c.Name
			break
		}
	}
	for i, a := range d.Credits.Cast {
		if i == 5 {
			break
		}
		f.Elenco = append(f.Elenco, a.Name)
	}
	return f, nil
}