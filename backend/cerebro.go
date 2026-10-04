package main

import "strings"

// responderPergunta é o "cérebro" do jogo.
// Por enquanto é um simulador. Depois vamos trocar por uma chamada a uma IA.
func responderPergunta(filme Filme, pergunta string) string {
	return "irrelevante"
}

var removedorDeAcentos = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c",
)

// normalizar deixa o texto em minúsculas, sem acentos e sem espaços sobrando,
// para "O Rei Leão" e " o rei  leao " serem considerados iguais.
func normalizar(s string) string {
	s = strings.ToLower(s)
	s = removedorDeAcentos.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}