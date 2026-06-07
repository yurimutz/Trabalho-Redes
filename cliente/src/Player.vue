<script setup>
import { onMounted, ref } from 'vue';
const videoPlayer = ref(null);

let chunkAtual = 1;
let manifest = null;
let promessaDeLimpeza = null;
let loopAtivo = false;

// Video config
let videoSet = null;
let videoRepresentation = null;
let videoMimeType = null; 
let videoCodecs = null;     
let videoRepId = null;    
let videoSegment = null;
let videoInitTemplate = null; 
let videoMediaTemplate = null;   // ex: chunk_$RepresentationID$_$Number$.m4s
let segmentTimeline = null;
let sTags = null;
let quantidadeDeChunks = 0;
let tamanhoVideoTotal = null;
let qualidades = [];
let totalQualidades = 0;
let tamanhoSegmentos = null;

// Audio config
let audioSet = null;
let audioRep = null;
let audioMimeType = null; 
let audioCodecs = null;     
let audioRepId = null;   
let audioSegment = null;
let audioInitTemplate = null; 
let audioMediaTemplate = null;

// Instâncias do MediaSource e SourceBuffers (criadas em iniciarStreaming)
let mediaSource = null;
let videoBuffer = null;
let audioBuffer = null;

// Dados para controlar o buffer do mediaSource
let historicoBanda = [];
let tamanhoBufferBanda = 0;
let tamMaxBufferBanda = 5;

let hist = [];

const manifestUrl = "http://localhost:8080/manifesto/lol/manifesto1.mpd";
const videosBaseUrl = "http://localhost:8080/videos/lol/"; // Ajuste conforme sua pasta de vídeos

async function buscaManifesto() {
  // 2. Busca e analisa o manifesto (O que você já tinha feito!)
      const response = await fetch(manifestUrl);
      if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);
      const xmlText = await response.text();
      
      const parser = new DOMParser();
      manifest = parser.parseFromString(xmlText, "application/xml");

    // // 2. Busca e analisa o manifesto (O que você já tinha feito!)
      // const response = await fetch(manifestUrl);
      // if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);
      // const xmlText = await response.text();
      
      // const parser = new DOMParser();
      // const manifest = parser.parseFromString(xmlText, "application/xml");
}

function traduzDuracaoParaSegundos(string) {
  if (!string) return 0;
  const regex = /PT(?:([0-9]+)H)?(?:([0-9]+)M)?(?:([0-9]+(?:\.[0-9]+)?)S)?/;
  const matches = string.match(regex);
  
  const horas = parseFloat(matches[1] || 0);
  const minutos = parseFloat(matches[2] || 0);
  const segundos = parseFloat(matches[3] || 0);
  
  // tamanhoVideoTotal = (horas * 3600) + (minutos * 60) + segundos;
  return (horas * 3600) + (minutos * 60) + segundos;
}

async function extraiInformacoesManifesto(){
  // Extraindo informações de vídeo
      videoSet = manifest.querySelector('AdaptationSet[contentType="video"]');
      // console.log(videoSet);
      videoRepresentation = videoSet.querySelector("Representation");
      videoMimeType = videoRepresentation.getAttribute("mimeType"); // ex: video/mp4
      videoCodecs = videoRepresentation.getAttribute("codecs");     // ex: avc1.64001e
      videoRepId = videoRepresentation.getAttribute("id");          // ex: 2

      videoSegment = manifest.querySelector("SegmentTemplate");
      videoInitTemplate = videoSegment.getAttribute("initialization"); // ex: init_$RepresentationID$.mp4
      videoMediaTemplate = videoSegment.getAttribute("media");         // ex: chunk_$RepresentationID$_$Number$.m4s

      segmentTimeline = manifest.querySelector("SegmentTimeline");
      sTags = segmentTimeline.querySelectorAll("S");

      // Extraindo informações do áudio
      audioSet = manifest.querySelector('AdaptationSet[contentType="audio"]');
      audioRep = audioSet.querySelector("Representation");
      audioMimeType = audioRep.getAttribute("mimeType"); 
      audioCodecs = audioRep.getAttribute("codecs");     
      audioRepId = audioRep.getAttribute("id");          

      audioSegment = audioSet.querySelector("SegmentTemplate");
      audioInitTemplate = audioSegment.getAttribute("initialization"); 
      audioMediaTemplate = audioSegment.getAttribute("media");

      // Gambi para pegar todas as qualidades do manifesto
      let representationALL = videoSet.querySelectorAll('Representation');
      representationALL.forEach((rep) => {
        qualidades.push((parseFloat(rep.getAttribute("bandwidth"))) / 10);
        //qualidades[totalQualidades] = (parseFloat(rep.getAttribute("bandwidth"))) / 10;
        //totalQualidades++;
      });

      // Forca bruta mesmo pq nao to com criatividade para encontrar o 360p em um vetor generico
      videoRepId = 1;

      qualidades.forEach((q) => {
        console.log("Qualidade: " + q); 
      });
      

      // Calculando a quantidade de chunks
      sTags.forEach(sTag => {
        const repeatAttr = sTag.getAttribute("r");

        // Quantidade de repeats do chunk
        const repeats = repeatAttr ? parseInt(repeatAttr) : 0;

        quantidadeDeChunks += (1 + repeats);
      })
      console.log("Total de chunks: ", quantidadeDeChunks);

      // Extrai tempo total do video direto do manifesto
      const mpd = manifest.querySelector('MPD');
      let aux = mpd.getAttribute("mediaPresentationDuration");
      console.log(aux);
      tamanhoVideoTotal = traduzDuracaoParaSegundos(aux);
      console.log("Tamanho total do video em segundos: " + tamanhoVideoTotal);

      
      let aux2 = mpd.getAttribute("maxSegmentDuration");
      console.log(aux2);
      tamanhoSegmentos = traduzDuracaoParaSegundos(aux2);
      console.log("Tamanho de cada segmento: " + tamanhoSegmentos);

}

async function rodarGerenciadorDeChunks() {
  // Se o loop já estiver rodando, evita criar um processo duplicado na memória
  if (loopAtivo) return;
  loopAtivo = true;

  console.log("Processo produtor de chunks ativado.");

  let videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);

  while (chunkAtual <= quantidadeDeChunks) {

    if (promessaDeLimpeza) {
      console.log("Loop em pausa aguardando a faxina dos buffers terminar...");
      await promessaDeLimpeza;
    }

    let i = chunkAtual;

    // Monta o nome do arquivo substituindo as variaveis dinâmicas do XML
    let videoChunkUrl = videoMediaTemplate
                          .replace('$RepresentationID$', videoRepId)
                          .replace('$Number$', i);

    let audioChunkUrl = audioMediaTemplate
                          .replace('$RepresentationID$', audioRepId)
                          .replace('$Number$', i);

    const motivo = await aguardarEspacoNoBuffer(videoPlayer.value, videoBuffer, 2);

    if (motivo === "SEEK_DETECTADO") {
      console.log(`Seek detectado durante o download. Abortando avanço do chunk ${i}. O próximo será ${chunkAtual}.`);
      continue;
    }

    console.log(`Baixando pedaço ${i}...`);

    let [resultadoVideo, resultadoAudio] = await Promise.all([
      baixarECalcularBanda(videosBaseUrl + videoChunkUrl),
      baixarECalcularBanda(videosBaseUrl + audioChunkUrl)
    ]);

    if (i !== chunkAtual) {
      console.log(`[Defesa] Download do chunk ${i} descartado. O seek alterou a rota para ${chunkAtual}.`);
      continue;
    }

    // Exibe a banda calculada no console
    console.log(`Banda do Vídeo: ${resultadoVideo.kbps.toFixed(2)} Kbps`);
    console.log(`Banda do Áudio: ${resultadoAudio.kbps.toFixed(2)} Kbps`);

    // Manipulações para a qualidade sob demanda
    historicoBanda[tamanhoBufferBanda] = resultadoVideo.kbps;
    tamanhoBufferBanda++;

    console.log("Historico de chunks atualizado: " + i);
    hist.push(i);

    if (i !== chunkAtual) continue;
    await injetarComSeguranca(videoBuffer, resultadoVideo.dados);
    if (i !== chunkAtual) continue;
    await injetarComSeguranca(audioBuffer, resultadoAudio.dados);

    console.log(`Pedaço ${i} injetado! (Video + Audio)`);

    if (tamanhoBufferBanda > 3) {
      const mediaBanda = (historicoBanda[0] + historicoBanda[1] + historicoBanda[2]) / 3;
      tamanhoBufferBanda = 0;
      console.log("Media de banda atual:" + mediaBanda);

      for (let j = qualidades.length; j >= 0; j--) {
        if (mediaBanda >= qualidades[j]) {
          if (videoRepId != j) {
            console.log("Setando qualidade " + j);
            videoRepId = j;
            videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
            fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer);
          }
          break;
        }
      }
    }

    chunkAtual++;
  }

  console.log("Todos os chunks baixados. Aguardando o vídeo terminar de tocar...");
  const motivo = await aguardarFimDoVideo(videoPlayer.value);

  // FINAL DO VÍDEO NATURAL
  loopAtivo = false;

  if(motivo == "FIM"){
    console.log("Todos os chunks foram processados. Fechando a transmissão...");
    if (mediaSource.readyState === 'open') {
      mediaSource.endOfStream();
      console.log("Stream fechado com sucesso. Vídeo completo na memória!");
    }
  } else {
    console.log("SEEK detectado apos o fim do ciclo do gerenciador");
  }
}

function aguardarFimDoVideo(videoElement) {
  return new Promise((resolve) => {
    
    // 1. Criamos a função que vai reagir ao evento
    const aoTerminar = () => {
      console.log("O vídeo chegou ao fim!");
      
      // 2. Removemos o escutador imediatamente para que ele não 
      // dispare duplicado caso o usuário dê replay no vídeo depois.
      videoElement.removeEventListener('ended', aoTerminar);
      
      // 3. Resolve a Promise retornando o 'false' que você pediu
      resolve("FIM");
    };

    const aoSeeking = () => {
      limparListeners();
      console.log("Seek detectado durante espera do fim. Abortando...");
      resolve("SEEK_DETECTADO");
    };

    // 4. Atrelamos o escutador ao player
    //videoElement.addEventListener('ended', aoTerminar);

    const limparListeners = () => {
      videoElement.removeEventListener('ended', aoTerminar);
      videoElement.removeEventListener('seeking', aoSeeking);
    };

    videoElement.addEventListener('ended', aoTerminar);
    videoElement.addEventListener('seeking', aoSeeking);
  });
}

function calcularChunksNoBuffer(videoElement, videoBuffer) {
  if (videoBuffer.buffered.length > 0) {
    
    // 1. Pega onde o buffer termina (ex: 40 segundos)
    // Usamos .length - 1 para pegar sempre o último bloco de tempo, caso haja buracos
    const ultimoBloco = videoBuffer.buffered.length - 1;
    const tempoFim = videoBuffer.buffered.end(ultimoBloco);
    
    // 2. Pega onde o usuário está AGORA (ex: 20 segundos)
    const tempoAtual = videoElement.currentTime;
    
    // 3. A MÁGICA: O estoque real é apenas o que falta tocar! (40 - 20 = 20 segundos)
    const estoqueFuturo = tempoFim - tempoAtual;
    
    // 4. Divide pelo tamanho do chunk (20 / 4 = 5 chunks restantes)
    const DURACAO_DO_CHUNK = tamanhoSegmentos;
    const estimativaDeChunks = Math.floor(estoqueFuturo / DURACAO_DO_CHUNK);
    
    return estimativaDeChunks;
  }
  
  return 0;
}

function tempoJaEstaBaixado(tempoDesejado, videoBuffer, novoChunk) {
  // O buffer pode ter vários "blocos" de tempo separados (TimeRanges) se o usuário pular muito
  if(hist.includes(novoChunk)){
    console.log("Ja contem esse chunk em memoria");
    return true;
  }

  // for (let i = 0; i < videoBuffer.buffered.length; i++) {
  //   const inicio = videoBuffer.buffered.start(i);
  //   const fim = videoBuffer.buffered.end(i);
    
  //   // Se o tempo desejado estiver dentro deste bloco, já temos o vídeo!
  //   if (tempoDesejado >= inicio && tempoDesejado <= fim) {
  //     return true;
  //   }
  // }
  console.log("Chunk ainda nao esta em memoria " + novoChunk);
  return false;
}

function limparBufferSeguro(buffer, duracaoTotal) {
  return new Promise((resolve) => {
    // 1. Se estiver no meio de uma injeção, mandamos parar (abortar)
    if (buffer.updating) {
      buffer.abort();
    }

    // 2. Se o buffer já estiver completamente vazio, não precisa fazer nada
    if (buffer.buffered.length === 0) {
      resolve();
      return;
    }

    // 3. Criamos um "escutador" para avisar quando a faxina terminar
    const aoTerminarLimpeza = () => {
      buffer.removeEventListener('updateend', aoTerminarLimpeza);
      console.log("Buffer limpo com sucesso!");
      hist.length = 0;
      resolve(); // Destrava a Promise
    };

    buffer.addEventListener('updateend', aoTerminarLimpeza);

    // 4. Dá a ordem de apagar tudo: do segundo 0 até o final do vídeo!
    buffer.remove(0, duracaoTotal);
  });
}

function aguardarEspacoNoBuffer(videoElement, videoBuffer, limiteChunks) {
  return new Promise((resolve) => {
    
    // 1. Checagem imediata: se já tiver espaço, nem precisa criar o evento, 
    // resolve a Promise na hora e deixa o loop seguir.
    if (calcularChunksNoBuffer(videoElement, videoBuffer) <= limiteChunks) {
      resolve("ESPACO_LIVRE");
      return;
    }

    console.log("Buffer cheio! Aguardando o vídeo tocar para liberar espaço...");

    // 2. A função que será chamada a cada milissegundo que o vídeo tocar
    const checarEstoque = () => {
      if (calcularChunksNoBuffer(videoElement, videoBuffer) <= limiteChunks) {
        limparListeners("ESPACO_LIVRE");
      }
    };

    const acordar = () => {
      console.log("Seek detectado, acordando loop a força.");
      limparListeners("SEEK_DETECTADO");
    }

    const limparListeners = (motivo) => {
      videoElement.removeEventListener('timeupdate', checarEstoque);
      videoElement.removeEventListener('seeking', acordar);
      resolve(motivo); // Destrava o seu loop while!
    }

    // 3. Atrela a função ao evento de tempo do player
    videoElement.addEventListener('timeupdate', checarEstoque);
    videoElement.addEventListener('seeking', acordar);
  });
}

async function baixarECalcularBanda(url) {
  // 1. Inicia o cronômetro de alta precisão
  const inicio = performance.now();

  const response = await fetch(url);
  if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);

  // Transforma a resposta em binário puro
  const dadosBinarios = await response.arrayBuffer();

  // 2. Para o cronômetro
  const fim = performance.now();
  const tempoEmMilissegundos = fim - inicio;

  // 3. A Matemática da Rede
  const bytes = dadosBinarios.byteLength;
  const bits = bytes * 8;
  
  // Fórmula: (bits / tempo_em_segundos) = bps
  // Dividimos por 1000 para transformar em Kbps
  const tempoEmSegundos = tempoEmMilissegundos / 1000;
  const bandaKbps = (bits / tempoEmSegundos) / 1000;

  return {
    dados: dadosBinarios,
    kbps: bandaKbps
  };
}

async function injetarComSeguranca(bufferDoCanal, dados) {
  return new Promise((resolve, reject) => {
    // Se o cano estiver ocupado, rejeitamos para não quebrar o navegador
    if (bufferDoCanal.updating) {
      reject(new Error("O buffer está ocupado processando outro pedaço."));
      return;
    }

    // Cria o ouvinte que avisa quando terminou
    const aoTerminar = () => {
      bufferDoCanal.removeEventListener('updateend', aoTerminar);
      resolve(); // Destrava o código!
    };

    bufferDoCanal.addEventListener('updateend', aoTerminar);
    
    // Injeta os dados
    bufferDoCanal.appendBuffer(dados);
  });
}

// Baixa um pedaço em formato binário e injeta no SourceBuffer
async function fetchAndAppend(url, targetBuffer) {
  const res = await fetch(url);
  const buffer = await res.arrayBuffer();

  return new Promise((resolve) => {
    targetBuffer.addEventListener('updateend', resolve, { once: true });
    targetBuffer.appendBuffer(buffer);
  });
}

async function iniciarStreaming() {
  try {

    await buscaManifesto();
    console.log("foi o manifesto");
    await extraiInformacoesManifesto();
    console.log("foi a extracao");

    // 1. Inicia o MediaSource e vincula à tag de vídeo
    mediaSource = new MediaSource();
    videoPlayer.value.src = URL.createObjectURL(mediaSource);

    // Só podemos começar a injetar coisas quando o source estiver aberto
    mediaSource.addEventListener('sourceopen', async () => {

      // Defini o tempo total do video
      mediaSource.duration = tamanhoVideoTotal;

      // Criando os buffers de video e audio
      videoBuffer = mediaSource.addSourceBuffer(`${videoMimeType}; codecs="${videoCodecs}"`);
      audioBuffer = mediaSource.addSourceBuffer(`${audioMimeType}; codecs="${audioCodecs}"`);

      videoPlayer.value.addEventListener('seeking', async() => {
        const tempo = videoPlayer.value.currentTime;
        console.log(`usuario clicou em ${tempo} segundos`);

        const novoChunk = Math.floor(tempo / tamanhoSegmentos) + 1;

        if (!tempoJaEstaBaixado(tempo, videoBuffer, novoChunk)) {
          //const novoChunk = Math.floor(tempo / tamanhoSegmentos) + 1;

          console.log(`O tempo ${tempo}s não está na memória. Pulando o download para o Chunk ${novoChunk}`);
          chunkAtual = novoChunk;

          // if (mediaSource.readyState === 'ended') {
          //   console.log("Reabrindo MediaSource fechado para permitir re-injeção...");
          //   mediaSource.duration = tamanhoVideoTotal; 
          // }

          promessaDeLimpeza = Promise.all([
            limparBufferSeguro(videoBuffer, mediaSource.duration),
            limparBufferSeguro(audioBuffer, mediaSource.duration)
          ]);
          
          // 2. Espera a faxina terminar
          await promessaDeLimpeza;
          
          // 3. Limpa a variável para avisar que a faxina acabou!
          promessaDeLimpeza = null;

          if (!loopAtivo) {
            console.log("Ressuscitando o loop de chunks para processar o retrocesso...");
            console.log("Reabrindo MediaSource fechado para permitir re-injeção...");
            mediaSource.duration = tamanhoVideoTotal;
            rodarGerenciadorDeChunks();
          }

        } else {
          console.log("Tempo já está na memória! Nenhum download necessário.");
        }

      });

      // 5. Baixa e Injeta o Cabeçalho (Init) - Obrigatório antes dos chunks!
      // Substitui a variavel $RepresentationID$ pelo ID real
      let videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
      const audioInitUrl = audioInitTemplate.replace('$RepresentationID$', audioRepId);

      console.log("Baixando inits...");

      // Promisse para baixar os dois ao mesmo tempo
      await Promise.all([
        fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer),
        fetchAndAppend(videosBaseUrl + audioInitUrl, audioBuffer)
      ]);

      // 6. Inicia o gerenciador de chunks como função reutilizável
      rodarGerenciadorDeChunks();
    });

  } catch (error) {
    console.error("Falha ao buscar o manifesto ou injetar chunks:", error.message);
  }
}

onMounted(() => {
  iniciarStreaming();
});
</script>

<template>
  <div class="player-container">
    <h2>Player de Vídeo</h2>
    <p>A navegação via Vue Router funcionou com sucesso!</p>
    <a href="/">home</a>

    <div class="video-wrapper">
      <video ref="videoPlayer" controls autoplay muted></video>
    </div>
  </div>
</template>

<style scoped>
.player-container {
  padding: 20px;
  /* min-height: 100vh; */
}

.video-wrapper {
  margin-top: 20px;
  width: 100%;
  max-width: 800px;
}

video {
  width: 100%;
  background-color: black;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0,0,0,0.5);
}

a {
  color: #42b883;
  margin-bottom: 10px;
}

button {
  padding: 10px 20px;
  cursor: pointer;
}
</style>