package main

type Filme struct {
	Titulo string `json:"titulo"`
	Ano    int    `json:"ano"`
}

// Banco temporário. Depois trocar pela API do TMDb.
var bancoDeFilmes = []Filme{
	{"Matrix", 1999},
	{"O Rei Leão", 1994},
	{"O Iluminado", 1980},
	{"Vingadores", 2012},
	{"Titanic", 1997},
}
