package main

import (
	"log"
	"math/rand"
)

// Filme guarda tudo o que sabemos sobre o filme secreto.
// Sem tags json de propósito: este struct nunca deve ser enviado ao navegador.
type Filme struct {
	ID             int
	Titulo         string
	TituloOriginal string
	Ano            int
	Generos        []string
	Diretor        string
	Elenco         []string
	DuracaoMin     int
	Sinopse        string
	Capa 		   string
}

// Plano B, caso o TMDb esteja fora do ar ou sem token.
var bancoDeFilmes = []Filme{
	{Titulo: "Matrix", TituloOriginal: "The Matrix", Ano: 1999},
	{Titulo: "O Rei Leão", TituloOriginal: "The Lion King", Ano: 1994},
	{Titulo: "O Iluminado", TituloOriginal: "The Shining", Ano: 1980},
	{Titulo: "Vingadores", TituloOriginal: "The Avengers", Ano: 2012},
	{Titulo: "Titanic", TituloOriginal: "Titanic", Ano: 1997},
}

// sortearFilme tenta o TMDb e, se falhar, cai no banco local.
func sortearFilme() Filme {
	f, err := sortearFilmeTMDb()
	if err != nil {
		log.Println("AVISO: TMDb falhou, usando banco local:", err)
		return bancoDeFilmes[rand.Intn(len(bancoDeFilmes))]
	}
	return f
}