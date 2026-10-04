# Adivinhe o Filme

**Qual filme eu estou pensando?** Um jogo interativo em formato de site, inspirado no clássico "20 perguntas", mas com filmes. Você define quantas perguntas pode fazer, descobre o filme secreto com respostas de Sim, Não ou Irrelevante e dá o palpite final antes que as perguntas acabem.

## Modos de jogo

1. **Você adivinha** (funcionando): o site sorteia um filme famoso em segredo. Você faz perguntas de sim ou não e, quando quiser (ou quando as perguntas acabarem), dá o palpite. Ao final, o site mostra a capa, o diretor, o ano e o gênero do filme.
2. **O site adivinha** (em desenvolvimento): você pensa em um filme, o site faz as perguntas e, no fim, tenta acertar.

Em ambos os modos, o limite de perguntas é configurável.

## Status

Em desenvolvimento. O modo 1 está completo e jogável localmente.

## Tecnologias

- **Go**: backend e API REST, usando só a biblioteca padrão (mais o `godotenv` para ler o `.env`)
- **HTML, CSS e JavaScript**: interface do site em 3 telas, sem frameworks
- **TMDB API**: sorteio dos filmes, detalhes, capas e busca para o autocomplete
- **Gemini API (Google)**: responde as perguntas do jogador
- **Python**: protótipo inicial de terminal

## Decisões de projeto

- **O filme secreto nunca sai do servidor** até o palpite ser dado. O navegador só conhece o ID da partida.
- **Não confiamos na IA**: a resposta do Gemini é filtrada, e só `Sim`, `Não` ou `Irrelevante` passam. Qualquer outra coisa vira `Irrelevante`.
- **A pergunta do jogador é tratada como dado, não como instrução**, para dificultar tentativas de enganar a IA.
- **O palpite escolhido na lista é comparado pelo ID do TMDB**, e não pelo texto, então funciona em português ou em inglês.
- **Se o TMDB ou a IA falharem**, o jogo continua: o filme vem de uma lista local e a pergunta não é descontada.

## Estrutura do projeto

```
adivinhe-o-filme/
  backend/     servidor e API em Go
  frontend/    páginas HTML, CSS e JavaScript
```

## Como rodar

Requisitos: [Go](https://go.dev/dl), um token da [API do TMDB](https://www.themoviedb.org/settings/api) e uma chave da API do Gemini (Google AI Studio).

```
git clone https://github.com/byGlazed/adivinhe-o-filme.git
cd adivinhe-o-filme/backend
```

Crie o arquivo `backend/.env`, usando o `backend/.env.example` como modelo, com as suas chaves:

```
TMDB_TOKEN=seu_token_do_tmdb
GEMINI_API_KEY=sua_chave_do_gemini
GEMINI_MODEL=gemini-3.5-flash
```

Depois, ainda dentro de `backend`:

```
go run .
```

Abra `http://localhost:8080` no navegador. O servidor precisa ser iniciado de dentro da pasta `backend`, porque é de lá que ele lê o `.env` e encontra o `frontend`.

O arquivo `.env` está no `.gitignore` e nunca deve ser commitado.

## API

| Método | Rota | O que faz |
| --- | --- | --- |
| GET | `/api/saude` | Confere se o servidor está no ar |
| POST | `/api/partidas` | Cria uma partida (`max_perguntas`) |
| POST | `/api/partidas/{id}/perguntas` | Faz uma pergunta e recebe Sim, Não ou Irrelevante |
| POST | `/api/partidas/{id}/palpite` | Dá o palpite final e recebe o resultado |
| GET | `/api/filmes/busca?q=` | Sugestões de filmes para o autocomplete |

## Roadmap

- [x] Estrutura inicial do repositório
- [x] Lógica do jogo (limite de perguntas, palpite, contagem)
- [x] API do backend
- [x] Interface do site (3 telas)
- [x] Integração com o TMDB
- [x] Integração com IA (modo 1)
- [x] Autocomplete de filmes no palpite
- [ ] Modo 2: o site adivinha o filme
- [ ] Preparar para produção (porta via variável de ambiente, IDs aleatórios, expiração de partidas, limite de requisições)
- [ ] Testes automatizados
- [ ] Deploy

## Créditos

Dados e capas de filmes: [TMDB](https://www.themoviedb.org). Este produto usa a API do TMDB, mas não é endossado nem certificado pelo TMDB.

## Autor

Gabriel