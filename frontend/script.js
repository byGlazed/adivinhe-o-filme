let partidaId = null;

const telaConfig = document.getElementById("tela-config");
const telaJogo = document.getElementById("tela-jogo");
const telaResultado = document.getElementById("tela-resultado");

const contador = document.getElementById("contador");
const historico = document.getElementById("historico");
const erro = document.getElementById("erro");

const formPergunta = document.getElementById("form-pergunta");
const campoPergunta = document.getElementById("campo-pergunta");
const formPalpite = document.getElementById("form-palpite");
const campoPalpite = document.getElementById("campo-palpite");

function mostrar(tela) {
  for (const t of [telaConfig, telaJogo, telaResultado]) {
    t.hidden = true;
  }
  tela.hidden = false;
}

function mostrarErro(mensagem) {
  erro.textContent = mensagem;
}

function atualizarContador(restantes) {
  contador.textContent = `Perguntas restantes: ${restantes}`;
  // Sem perguntas sobrando, bloqueia o formulário de perguntas
  for (const el of formPergunta.elements) {
    el.disabled = restantes === 0;
  }
}

async function chamarApi(url, corpo) {
  const resposta = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(corpo),
  });

  if (!resposta.ok) {
    const texto = await resposta.text();
    throw new Error(texto.trim());
  }
  return resposta.json();
}

document.getElementById("btn-iniciar").addEventListener("click", async () => {
  mostrarErro("");
  try {
    const max = Number(document.getElementById("max-perguntas").value);
    const partida = await chamarApi("/api/partidas", { max_perguntas: max });

    partidaId = partida.id;
    historico.innerHTML = "";
    atualizarContador(partida.max_perguntas);
    mostrar(telaJogo);
  } catch (e) {
    mostrarErro(e.message);
  }
});

formPergunta.addEventListener("submit", async (evento) => {
  evento.preventDefault();
  mostrarErro("");

  const pergunta = campoPergunta.value.trim();
  if (!pergunta) return;

  try {
    const dados = await chamarApi(`/api/partidas/${partidaId}/perguntas`, { pergunta });

    const item = document.createElement("li");
    item.textContent = `${pergunta} → ${dados.resposta}`;
    historico.appendChild(item);

    campoPergunta.value = "";
    atualizarContador(dados.perguntas_restantes);
  } catch (e) {
    mostrarErro(e.message);
  }
});

formPalpite.addEventListener("submit", async (evento) => {
  evento.preventDefault();
  mostrarErro("");

  const palpite = campoPalpite.value.trim();
  if (!palpite) return;

  try {
    const dados = await chamarApi(`/api/partidas/${partidaId}/palpite`, { palpite });

    document.getElementById("resultado-titulo").textContent =
      dados.acertou ? "Você acertou! 🎉" : "Não foi dessa vez...";
    document.getElementById("resultado-texto").textContent = `O filme era: ${dados.filme}`;

    campoPalpite.value = "";
    mostrar(telaResultado);
  } catch (e) {
    mostrarErro(e.message);
  }
});

document.getElementById("btn-novo").addEventListener("click", () => {
  mostrarErro("");
  mostrar(telaConfig);
});