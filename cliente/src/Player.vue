<script setup>
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import Header from './components/Header.vue';
const videoPlayer = ref(null);
const route = useRoute();

let chunkAtual = 1;
let manifest = null;
let promessaDeLimpeza = null;
let loopAtivo = false;
let videoInitUrl = null;


// Video config
let videoSet = null;
let videoRepresentation = null;
let videoMimeType = null; 
let videoCodecs = null;     
let videoRepId = null;    
let videoSegment = null;
let videoInitTemplate = null; 
let videoMediaTemplate = null;
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
let tamMaxTrava = 3;
let travaABR = 0;

let hist = [];
let manifestUrl = "";
let videosBaseUrl = ""; // Ajuste conforme sua pasta de vídeos

async function buscaManifesto() {
  // Busca e analisa o manifesto
      const response = await fetch(manifestUrl);
      if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);
      const xmlText = await response.text();
      
      const parser = new DOMParser();
      manifest = parser.parseFromString(xmlText, "application/xml");
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
        qualidades.push((parseFloat(rep.getAttribute("bandwidth"))) / 1000);
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

async function criarMediaSource() {
  mediaSource = new MediaSource();
  videoPlayer.value.src = URL.createObjectURL(mediaSource);

  await new Promise((resolve) => {
    mediaSource.addEventListener('sourceopen', async () => {
      mediaSource.duration = tamanhoVideoTotal;

      videoBuffer = mediaSource.addSourceBuffer(`${videoMimeType}; codecs="${videoCodecs}"`);
      audioBuffer = mediaSource.addSourceBuffer(`${audioMimeType}; codecs="${audioCodecs}"`);

      const videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
      const audioInitUrl = audioInitTemplate.replace('$RepresentationID$', audioRepId);

      await Promise.all([
        fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer),
        fetchAndAppend(videosBaseUrl + audioInitUrl, audioBuffer)
      ]);

      resolve();
    }, { once: true });
  });
}

async function rodarGerenciadorDeChunks() {
  // Se o loop já estiver rodando, evita criar um processo duplicado na memória
  if (loopAtivo) return;
  loopAtivo = true;

  console.log("Processo produtor de chunks ativado.");

  videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);

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

    const motivo = await aguardarEspacoNoBuffer(videoPlayer.value, videoBuffer, 10);

    if (motivo === "SEEK_DETECTADO") {
      console.log(`Seek detectado durante o download. Abortando avanço do chunk ${i}. O próximo será ${chunkAtual}.`);
      continue;
    }

    console.log(`Baixando pedaço ${i}...`);
    // let resultadoVideo;
    // let resultadoAudio;

    try{
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
    // historicoBanda[tamanhoBufferBanda] = resultadoVideo.kbps;
    // tamanhoBufferBanda++;

    historicoBanda.push(resultadoVideo.kbps);
    if (historicoBanda.length > 3) {
      historicoBanda.shift(); 
    }

    console.log("Historico de chunks atualizado: " + i);
    //hist.push(i);

    if (i !== chunkAtual) continue;
    await injetarComSeguranca(videoBuffer, resultadoVideo.dados);
    if (i !== chunkAtual) continue;
    await injetarComSeguranca(audioBuffer, resultadoAudio.dados);

    console.log(`Pedaço ${i} injetado! (Video + Audio)`);

    const tempoAgora = videoPlayer.value.currentTime;
    await limparPassado(videoBuffer, tempoAgora);
    await limparPassado(audioBuffer, tempoAgora);

    await logicaABR();
    // if (tamanhoBufferBanda > 2) {
    //   const mediaBanda = (historicoBanda[0] + historicoBanda[1] + historicoBanda[2]) / 3;
    //   tamanhoBufferBanda = 0;
    //   console.log("Media de banda atual:" + mediaBanda);

    //   for (let j = qualidades.length; j >= 0; j--) {
    //     if (mediaBanda >= qualidades[j]) {
    //       if (videoRepId != j) {
    //         console.log("Setando qualidade " + j);
    //         videoRepId = j;
    //         videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
    //         await fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer);
    //       }
    //       break;
    //     }
    //   }
    // }

    if(travaABR > 0){
      travaABR--;
    }
    chunkAtual++;

    } catch {
      console.log("Sem conexao com a internet");
      await new Promise(resolve => setTimeout(resolve, 4000));
      continue;
    }

  }

  // FINAL DO VÍDEO NATURAL
  loopAtivo = false;

  if (mediaSource.readyState === 'open') {
    mediaSource.endOfStream();
    console.log("Stream fechado com sucesso. Vídeo completo na memória!");
  }

  console.log("Todos os chunks baixados. Aguardando o vídeo terminar de tocar...");
  const motivo = await aguardarFimDoVideo(videoPlayer.value);
  
  if(motivo == "FIM"){
    console.log("Todos os chunks foram processados. Fechando a transmissão...");
    //mediaSource.endOfStream();  
    console.log("Stream fechado com sucesso. Vídeo completo na memória!");
  } else {
    console.log("SEEK detectado apos o fim do ciclo do gerenciador");
  }
}

function aguardarFimDoVideo(videoElement) {
  return new Promise((resolve) => {
    
    const aoTerminar = () => {
      console.log("O vídeo chegou ao fim!");
        
      // Remove o escutador imediatamente para que ele não 
      // dispare duplicado caso o usuário dê replay no vídeo depois
      videoElement.removeEventListener('ended', aoTerminar);

      resolve("FIM");
    };

    const aoSeeking = () => {
      limparListeners();
      console.log("Seek detectado durante espera do fim. Abortando...");
      resolve("SEEK_DETECTADO");
    };
  
    const limparListeners = () => {
      videoElement.removeEventListener('ended', aoTerminar);
      videoElement.removeEventListener('seeking', aoSeeking);
    };

    // Atrela o escutador ao player
    videoElement.addEventListener('ended', aoTerminar);
    videoElement.addEventListener('seeking', aoSeeking);
  });
}

function calcularChunksNoBuffer(videoElement, videoBuffer) {
  if (videoBuffer.buffered.length > 0) {
    
    const ultimoBloco = videoBuffer.buffered.length - 1;
    const tempoFim = videoBuffer.buffered.end(ultimoBloco);
    
    // Pega onde o usuário está AGORA
    const tempoAtual = videoElement.currentTime;
    
    const estoqueFuturo = tempoFim - tempoAtual;
    
    const DURACAO_DO_CHUNK = tamanhoSegmentos;
    const estimativaDeChunks = Math.floor(estoqueFuturo / DURACAO_DO_CHUNK);
    
    return estimativaDeChunks;
  }
  
  return 0;
}

async function logicaABR(){
  // if (tamanhoBufferBanda > 2) {
  // Apos mudanca de qualidade, fica tamMaxTrava chunks(4 * tamMaxTrava segundos) travados a fim de estabilizar a rede
  // Os 3 primeiros chunks, que inicializam o player estao protegidos, a fim de garantir uma inicializacao rapida
  if(historicoBanda.length === 3){

        // Variaveis para ajudar em possiveis mudancas
        const pesoRecente = 5;
        const pesoMeio = 3;
        const pesoAntigo = 2;
        const somaPesos = pesoAntigo + pesoMeio + pesoRecente;

        // Media ponderada valorizando o chunk mais recente
        let mediaBanda = ((historicoBanda[0]*pesoAntigo) + (historicoBanda[1]*pesoMeio) + (historicoBanda[2]*pesoRecente)) / somaPesos;
        //tamanhoBufferBanda = 0;
        mediaBanda = mediaBanda * 0.8;
        console.log("Media de banda atual:" + mediaBanda);

        for (let j = qualidades.length-1; j >= 0; j--) {
          if (mediaBanda >= qualidades[j]) {
            if (videoRepId < j && (travaABR == 0)) {
              console.log("Subindo qualidade " + j);
              travaABR = tamMaxTrava;
              videoRepId = j;
              videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
              await fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer);
              break;
            }
            // } else {
            //   console.log(`Qualidade ${j} mantida`);
            // }
            if(videoRepId > j){
              console.log("Qualidade desceu para a " + j);
              travaABR = 0;
              videoRepId = j;
              videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
              await fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer);
              break;
            }
            if(videoRepId === j){
              console.log(`Qualidade ${j} mantida`);
              break;
            }
            //break;
          }
        }
      }
}

// So estava aqui temporariamente, em algum momento vai dar problema
// Acabou dando quando tentei limpar o buffer velho do player
// function tempoJaEstaBaixado(tempoDesejado, videoBuffer, novoChunk) {
//   // So verifica se o id do chunk esta no array
//   if(hist.includes(novoChunk)){
//     console.log("Ja contem esse chunk em memoria");
//     return true;
//   }

//   console.log("Chunk ainda nao esta em memoria " + novoChunk);
//   return false;
// }

function tempoJaEstaBaixado(tempoDesejado, videoBuffer) {
  // Se não tem nada alocado, já sabemos que não está na memória
  if (!videoBuffer || videoBuffer.buffered.length === 0) {
    console.log(`Cache Miss: Buffer está completamente vazio.`);
    return false;
  }

  const ranges = videoBuffer.buffered;
  // Margem de segurança (em segundos) para evitar que o player trave se o usuário 
  // clicar exatamente no milissegundo final de um chunk baixado.
  const margemSeguranca = 0.5; 

  // Percorre o array de "janelas" do buffer pra ver se o tempoDesejado cai dentro de algum bloco
  for (let i = 0; i < ranges.length; i++) {
    const inicioDoBloco = ranges.start(i);
    const fimDoBloco = ranges.end(i);

    if (tempoDesejado >= inicioDoBloco && tempoDesejado <= (fimDoBloco - margemSeguranca)) {
      console.log(`Cache Hit! Tempo ${tempoDesejado.toFixed(1)}s está na RAM (Bloco: ${inicioDoBloco.toFixed(1)}s até ${fimDoBloco.toFixed(1)}s)`);
      return true;
    }
  }

  console.log(`Cache Miss! Tempo ${tempoDesejado.toFixed(1)}s NÃO está na RAM.`);
  return false;
}

function limparPassado(buffer, tempoAtual) {
  return new Promise((resolve) => {
    const margemSeguranca = 10; // 10 segundos no passado

    // Se o buffer estiver ocupado com outra coisa ou vazio, ignora e destrava o loop
    if (!buffer || buffer.updating || buffer.buffered.length === 0) {
      resolve();
      return;
    }

    const tempoCorte = tempoAtual - margemSeguranca;
    const inicioPrimeiroBloco = buffer.buffered.start(0);

    // Se tiver coisa suficiente para ser limpa
    if (inicioPrimeiroBloco < tempoCorte) {
      const aoTerminar = () => {
        buffer.removeEventListener('updateend', aoTerminar);
        resolve();
      };

      buffer.addEventListener('updateend', aoTerminar);

      try {
        console.log(`Faxina: Limpando de ${inicioPrimeiroBloco.toFixed(1)}s até ${tempoCorte.toFixed(1)}s`);
        buffer.remove(inicioPrimeiroBloco, tempoCorte);
      } catch (erro) {
        buffer.removeEventListener('updateend', aoTerminar);
        resolve(); 
      }
    } else {
      // Buffer velho muito pequeno
      resolve(); 
    }
  });
}

function limparBufferSeguro(buffer, duracaoTotal) {
  return new Promise((resolve) => {
    // Se estiver no meio de uma injeção, aborta
    if (buffer.updating) {
      buffer.abort();
    }

    // Se o buffer já estiver completamente vazio, não precisa fazer nada
    if (buffer.buffered.length === 0) {
      resolve();
      return;
    }

    const aoTerminarLimpeza = () => {
      buffer.removeEventListener('updateend', aoTerminarLimpeza);
      console.log("Buffer limpo com sucesso!");
      //hist.length = 0;
      resolve();
    };

    buffer.addEventListener('updateend', aoTerminarLimpeza);

    buffer.remove(0, duracaoTotal);
  });
}

function aguardarEspacoNoBuffer(videoElement, videoBuffer, limiteChunks) {
  return new Promise((resolve) => {
    
    // Se já tiver espaço, nem precisa criar o evento, 
    if (calcularChunksNoBuffer(videoElement, videoBuffer) < limiteChunks) {
      resolve("ESPACO_LIVRE");
      return;
    }

    console.log("Buffer cheio! Aguardando o vídeo tocar para liberar espaço...");

    const checarEstoque = () => {
      if (calcularChunksNoBuffer(videoElement, videoBuffer) < limiteChunks) {
        limparListeners("ESPACO_LIVRE");
      }
    };

    // Destrava caso ocorra um seek enquanto estiver esperando espaco, caso de deadlock
    const acordar = () => {
      console.log("Seek detectado, acordando loop a força.");
      limparListeners("SEEK_DETECTADO");
    }

    const limparListeners = (motivo) => {
      videoElement.removeEventListener('timeupdate', checarEstoque);
      videoElement.removeEventListener('seeking', acordar);
      resolve(motivo);
    }

    videoElement.addEventListener('timeupdate', checarEstoque);
    videoElement.addEventListener('seeking', acordar);
  });
}

async function baixarECalcularBanda(url) {
  const inicio = performance.now();

  const response = await fetch(url);
  if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);

  // Transforma a resposta em binário puro
  const dadosBinarios = await response.arrayBuffer();

  const fim = performance.now();
  const tempoEmMilissegundos = fim - inicio;

  // Transforma binario em bits
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
    // Se o buffer estiver ocupado, rejeitamos para não quebrar o navegador
    if (bufferDoCanal.updating) {
      reject(new Error("O buffer está ocupado processando outro pedaço."));
      return;
    }

    const aoTerminar = () => {
      bufferDoCanal.removeEventListener('updateend', aoTerminar);
      resolve();
    };

    bufferDoCanal.addEventListener('updateend', aoTerminar);
    
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

    const mParam = route.query.m;
    if (!mParam) {
      alert("nenhum video selecionado");
      return;
    }

    const ipDoServidor = window.location.hostname;
    console.log(ipDoServidor);

    manifestUrl = `http://${ipDoServidor}:8080/${mParam}`;

    const ultimaBarra = manifestUrl.lastIndexOf('/');
    videosBaseUrl = manifestUrl.substring(0, ultimaBarra + 1);

    await buscaManifesto();
    console.log("foi o manifesto");
    await extraiInformacoesManifesto();
    console.log("foi a extracao");

    videoPlayer.value.addEventListener('seeking', async() => {
        const tempo = videoPlayer.value.currentTime;
        console.log(`usuario clicou em ${tempo} segundos`);

        const novoChunk = Math.floor(tempo / tamanhoSegmentos) + 1;

        if (!tempoJaEstaBaixado(tempo, videoBuffer, novoChunk)) {
          console.log(`O tempo ${tempo}s não está na memória. Pulando o download para o Chunk ${novoChunk}`);
          chunkAtual = novoChunk;

          promessaDeLimpeza = Promise.all([
            limparBufferSeguro(videoBuffer, tamanhoVideoTotal),
            limparBufferSeguro(audioBuffer, tamanhoVideoTotal)
          ]);
          
          await promessaDeLimpeza;
          
          promessaDeLimpeza = null;

          if (!loopAtivo) {
            console.log("Ressuscitando o loop de chunks para processar o retrocesso...");
            rodarGerenciadorDeChunks();
          }

        } else {
          console.log("Tempo já está na memória! Nenhum download necessário.");
        }

      });

      await criarMediaSource();
      rodarGerenciadorDeChunks();

  } catch (error) {
    console.error("Falha ao buscar o manifesto ou injetar chunks:", error.message);
  }
}

onMounted(() => {
  iniciarStreaming();
});
</script>

<template>

  <Header />

  <div class="player-container">
    <!-- <h2>Player de Vídeo</h2>
    <p>A navegação via Vue Router funcionou com sucesso!</p> -->
    
    <div class="video-wrapper">
      <video ref="videoPlayer" controls autoplay muted></video>
    </div>

    <!-- <a href="/">Voltar para home</a> -->

  </div>
  <a href="/" class="text-lg mt-6 group relative w-max">
      <span class ="p-1 relative z-10 group-hover:text-white">Voltar para home</span>
      <!-- <span class="absolute -bottom-1 left-1/2 w-0 transition-all h-0.5 bg-primary group-hover:w-3/6"></span> -->
      <!-- <span class="absolute -bottom-1 right-1/2 w-0 transition-all h-0.5 bg-primary group-hover:w-3/6"></span> -->
      <span class="absolute left-0 bottom-0 w-full h-0.5 transition-all bg-primary z-0 group-hover:h-full "></span>
  </a>
</template>

<style scoped>

.video-wrapper {
  margin-top: 20px;
  width: 100%;
  max-width: 1024px;
}

video {
  width: 100%;
  background-color: black;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0,0,0,0.5);
}


button {
  padding: 10px 20px;
  cursor: pointer;
}
</style>