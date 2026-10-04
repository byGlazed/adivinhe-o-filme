// ---------- Estado ----------
let partidaId = null;
let perguntasRestantes = 0;
let resultadoFinal = null;
let fase = "perguntando"; // "perguntando" | "respondida" | "palpitada"

// Autocomplete do palpite
let filmeEscolhidoId = 0;   // ID do TMDb do filme escolhido na lista (0 = digitou à mão)
let sugestoesAtuais = [];
let indiceAtivo = -1;
let temporizadorBusca = null;
let contadorBusca = 0;      // serve para ignorar respostas de buscas antigas

const TEXTO_RESPOSTA_PADRAO = "Resposta da pergunta";
const ROTULO_PROXIMO_PADRAO = "Ver Resultado ou Próxima pergunta";

// ---------- Elementos ----------
const $ = (id) => document.getElementById(id);

const telaConfig = $("tela-config");
const telaJogo = $("tela-jogo");
const telaResultado = $("tela-resultado");

const campoMax = $("max-perguntas");
const btnIniciar = $("btn-iniciar");
const erroConfig = $("erro-config");

const formPergunta = $("form-pergunta");
const campoPergunta = $("campo-pergunta");
const formPalpite = $("form-palpite");
const campoPalpite = $("campo-palpite");
const listaSugestoes = $("sugestoes");
const resposta = $("resposta");
const contador = $("contador");
const rotuloProximo = $("rotulo-proximo");
const btnProximo = $("btn-proximo");

const btnNovo = $("btn-novo");

// ---------- Funções de apoio ----------
function mostrar(tela) {
  for (const t of [telaConfig, telaJogo, telaResultado]) t.hidden = true;
  tela.hidden = false;
}

function bloquear(form, bloqueado) {
  for (const item of form.elements) item.disabled = bloqueado;
}

function setResposta(texto, destaque = false) {
  resposta.textContent = texto;
  resposta.classList.toggle("preenchida", destaque);
}

function definirProximo(rotulo, ativo) {
  rotuloProximo.textContent = rotulo;
  btnProximo.disabled = !ativo;
  btnProximo.setAttribute("aria-label", rotulo);
}

async function chamarApi(url, corpo) {
  const resp = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(corpo),
  });

  if (!resp.ok) {
    const texto = await resp.text();
    throw new Error(texto.trim());
  }
  return resp.json();
}

// Deixa a tela do jogo coerente com a fase atual
function renderizar(novaFase) {
  fase = novaFase;
  const semPerguntas = perguntasRestantes === 0;

  bloquear(formPergunta, fase !== "perguntando" || semPerguntas);
  bloquear(formPalpite, fase === "palpitada");
  contador.textContent = `Perguntas restantes: ${perguntasRestantes}`;

  if (fase === "palpitada") {
    definirProximo("Ver Resultado", true);
  } else if (fase === "respondida" && !semPerguntas) {
    definirProximo("Próxima pergunta", true);
  } else {
    definirProximo(ROTULO_PROXIMO_PADRAO, false);
  }
}

function mostrarResultado(d) {
  const lista = $("res-detalhes");
  lista.innerHTML = "";

  const linhas = [d.filme, d.diretor, d.ano ? String(d.ano) : "", (d.generos || []).join(", ")];
  for (const texto of linhas) {
    if (!texto) continue; // filmes do plano B podem não ter todos os dados
    const li = document.createElement("li");
    li.textContent = texto;
    lista.appendChild(li);
  }

  const img = $("res-capa");
  img.onerror = () => { img.hidden = true; };
  if (d.capa) {
    img.src = d.capa;
    img.hidden = false;
  } else {
    img.removeAttribute("src");
    img.hidden = true;
  }

  let mensagem;
  if (!d.acertou) {
    mensagem = "Poxa, você errou, foi quase, bora mais uma?";
  } else if (d.perguntas_feitas === 0) {
    mensagem = "Caramba, você acertou sem nem perguntar! Vamos mais uma?";
  } else {
    const palavra = d.perguntas_feitas === 1 ? "pergunta" : "perguntas";
    mensagem = `Caramba, você acertou com ${d.perguntas_feitas} ${palavra}. Vamos mais uma?`;
  }
  $("res-mensagem").textContent = mensagem;

  mostrar(telaResultado);
}

// ---------- Autocomplete do palpite ----------
function fecharSugestoes() {
  listaSugestoes.hidden = true;
  listaSugestoes.innerHTML = "";
  sugestoesAtuais = [];
  indiceAtivo = -1;
  campoPalpite.setAttribute("aria-expanded", "false");
}

function desenharSugestoes(lista) {
  listaSugestoes.innerHTML = "";
  sugestoesAtuais = lista;
  indiceAtivo = -1;

  if (lista.length === 0) {
    fecharSugestoes();
    return;
  }

  lista.forEach((s, i) => {
    const li = document.createElement("li");
    li.setAttribute("role", "option");
    li.textContent = s.ano ? `${s.titulo} (${s.ano})` : s.titulo;
    // mousedown (e não click) para disparar antes de o campo perder o foco
    li.addEventListener("mousedown", (e) => {
      e.preventDefault();
      escolherSugestao(i);
    });
    listaSugestoes.appendChild(li);
  });

  listaSugestoes.hidden = false;
  campoPalpite.setAttribute("aria-expanded", "true");
}

function escolherSugestao(i) {
  const s = sugestoesAtuais[i];
  if (!s) return;
  campoPalpite.value = s.titulo;
  filmeEscolhidoId = s.id;
  fecharSugestoes();
  campoPalpite.focus();
}

function destacar(novoIndice) {
  const itens = listaSugestoes.children;
  if (itens.length === 0) return;
  indiceAtivo = (novoIndice + itens.length) % itens.length;
  [...itens].forEach((li, i) => li.classList.toggle("ativo", i === indiceAtivo));
  itens[indiceAtivo].scrollIntoView({ block: "nearest" });
}

async function buscarSugestoes(termo) {
  const minhaBusca = ++contadorBusca;
  try {
    const resp = await fetch(`/api/filmes/busca?q=${encodeURIComponent(termo)}`);
    if (!resp.ok) return;
    const lista = await resp.json();
    // Ignora se saiu uma busca mais nova ou se o campo foi bloqueado nesse meio tempo
    if (minhaBusca !== contadorBusca || campoPalpite.disabled) return;
    desenharSugestoes(lista);
  } catch {
    // O autocomplete é opcional: se falhar, o jogador digita o título normalmente
  }
}

campoPalpite.addEventListener("input", () => {
  filmeEscolhidoId = 0; // digitou de novo: a escolha anterior não vale mais
  clearTimeout(temporizadorBusca);

  const termo = campoPalpite.value.trim();
  if (termo.length < 2) {
    contadorBusca++;
    fecharSugestoes();
    return;
  }
  // Espera 300 ms sem digitar antes de buscar, para não fazer uma busca por letra
  temporizadorBusca = setTimeout(() => buscarSugestoes(termo), 300);
});

campoPalpite.addEventListener("keydown", (e) => {
  if (listaSugestoes.hidden) return;

  if (e.key === "ArrowDown") {
    e.preventDefault();
    destacar(indiceAtivo + 1);
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    destacar(indiceAtivo < 0 ? listaSugestoes.children.length - 1 : indiceAtivo - 1);
  } else if (e.key === "Enter" && indiceAtivo >= 0) {
    e.preventDefault();
    escolherSugestao(indiceAtivo);
  } else if (e.key === "Escape") {
    fecharSugestoes();
  }
});

campoPalpite.addEventListener("blur", fecharSugestoes);

// ---------- Tela 1: começar ----------
btnIniciar.addEventListener("click", async () => {
  erroConfig.textContent = "";

  const max = Number(campoMax.value);
  if (!Number.isInteger(max) || max < 1 || max > 50) {
    erroConfig.textContent = "Escolha entre 1 e 50 perguntas.";
    return;
  }

  btnIniciar.disabled = true;
  try {
    const partida = await chamarApi("/api/partidas", { max_perguntas: max });

    partidaId = partida.id;
    perguntasRestantes = partida.max_perguntas;
    resultadoFinal = null;

    campoPergunta.value = "";
    campoPalpite.value = "";
    filmeEscolhidoId = 0;
    clearTimeout(temporizadorBusca);
    contadorBusca++;
    fecharSugestoes();

    setResposta(TEXTO_RESPOSTA_PADRAO);
    renderizar("perguntando");
    mostrar(telaJogo);
    campoPergunta.focus();
  } catch (e) {
    erroConfig.textContent = `Ops: ${e.message}`;
  } finally {
    btnIniciar.disabled = false;
  }
});

// ---------- Tela 2: perguntar ----------
formPergunta.addEventListener("submit", async (evento) => {
  evento.preventDefault();

  const pergunta = campoPergunta.value.trim();
  if (!pergunta) return;

  // Trava tudo enquanto a IA pensa, para não gastar duas perguntas num clique duplo
  bloquear(formPergunta, true);
  bloquear(formPalpite, true);
  definirProximo(ROTULO_PROXIMO_PADRAO, false);
  setResposta("A IA está pensando...");

  try {
    const dados = await chamarApi(`/api/partidas/${partidaId}/perguntas`, { pergunta });
    perguntasRestantes = dados.perguntas_restantes;

    let texto = dados.resposta;
    if (perguntasRestantes === 0) {
      texto += "\nAcabaram as perguntas! Dê o seu palpite final.";
    }
    setResposta(texto, true);
    renderizar("respondida");

    if (perguntasRestantes === 0) campoPalpite.focus();
  } catch (e) {
    setResposta(`Ops: ${e.message}`);
    renderizar(fase); // volta ao estado anterior; a pergunta digitada continua no campo
  }
});

// ---------- Tela 2: palpitar ----------
formPalpite.addEventListener("submit", async (evento) => {
  evento.preventDefault();

  const palpite = campoPalpite.value.trim();
  if (!palpite) return;

  // Cancela qualquer busca de sugestões em andamento
  clearTimeout(temporizadorBusca);
  contadorBusca++;
  fecharSugestoes();

  bloquear(formPergunta, true);
  bloquear(formPalpite, true);
  definirProximo(ROTULO_PROXIMO_PADRAO, false);
  setResposta("Conferindo o seu palpite...");

  try {
    resultadoFinal = await chamarApi(`/api/partidas/${partidaId}/palpite`, {
      palpite,
      filme_id: filmeEscolhidoId,
    });
    setResposta("Palpite enviado! Será que você acertou?", true);
    renderizar("palpitada");
  } catch (e) {
    setResposta(`Ops: ${e.message}`);
    renderizar(fase);
  }
});

// ---------- Tela 2: próxima pergunta / ver resultado ----------
btnProximo.addEventListener("click", () => {
  if (fase === "respondida") {
    campoPergunta.value = "";
    setResposta(TEXTO_RESPOSTA_PADRAO);
    renderizar("perguntando");
    campoPergunta.focus();
  } else if (fase === "palpitada") {
    mostrarResultado(resultadoFinal);
  }
});

// ---------- Tela 3: jogar de novo ----------
btnNovo.addEventListener("click", () => {
  erroConfig.textContent = "";
  mostrar(telaConfig);
});