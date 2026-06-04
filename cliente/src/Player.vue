<script setup>
import { onMounted, ref } from 'vue';
const videoPlayer = ref(null);

let chunkAtual = 1;

let manifest = null;
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
let audioSet = null;
let audioRep = null;
let audioMimeType = null; 
let audioCodecs = null;     
let audioRepId = null;   
let audioSegment = null;
let audioInitTemplate = null; 
let audioMediaTemplate = null;


const manifestUrl = "http://localhost:8080/manifesto/manifesto1.mpd";
const videosBaseUrl = "http://localhost:8080/videos/"; // Ajuste conforme sua pasta de vídeos

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

      
      // Calculando a quantidade de chunks
      sTags.forEach(sTag => {
        const repeatAttr = sTag.getAttribute("r");

        // Quantidade de repeats do chunk
        const repeats = repeatAttr ? parseInt(repeatAttr) : 0;

        quantidadeDeChunks += (1 + repeats);
      })
      console.log("Total de chunks: ", quantidadeDeChunks);
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

async function iniciarStreaming() {
  try {

    await buscaManifesto();
    console.log("foi o manifesto");
    await extraiInformacoesManifesto();
    console.log("foi a extracao");

    // 1. Inicia o MediaSource e vincula à tag de vídeo
    const mediaSource = new MediaSource();
    videoPlayer.value.src = URL.createObjectURL(mediaSource);

    // Só podemos começar a injetar coisas quando o source estiver aberto
    mediaSource.addEventListener('sourceopen', async () => {

      // Criando os buffers
      const videoBuffer = mediaSource.addSourceBuffer(`${videoMimeType}; codecs="${videoCodecs}"`);
      const audioBuffer = mediaSource.addSourceBuffer(`${audioMimeType}; codecs="${audioCodecs}"`);

      // Função Mágica: Baixa um pedaço em formato binário e espera o buffer engolir
      const fetchAndAppend = async (url, targetBuffer) => {
        const res = await fetch(url);
        const buffer = await res.arrayBuffer(); // Pega o binário (ArrayBuffer)

        return new Promise((resolve) => {
          // Só resolve a promessa quando o buffer gritar "terminei de atualizar!"
          targetBuffer.addEventListener('updateend', resolve, { once: true });
          targetBuffer.appendBuffer(buffer);
        });
      };

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

      // 6. O famoso Loop de Chunks (Exemplo: buscando os 5 primeiros pedaços)
      
      for (let i = chunkAtual; i <= quantidadeDeChunks; i++) {
        // Monta o nome do arquivo substituindo as variaveis dinâmicas do XML
        // Dependendo de como você gerou no FFmpeg, o $Number$ pode ser $Number%05d$ (com zeros). 
        // Adapte o replace abaixo se necessário.
        let videoChunkUrl = videoMediaTemplate
                              .replace('$RepresentationID$', videoRepId)
                              .replace('$Number$', i);

        let audioChunkUrl = audioMediaTemplate
                              .replace('$RepresentationID$', audioRepId)
                              .replace('$Number$', i);

        console.log(`Baixando pedaço ${i}...`);
        
        // O "await" aqui é crucial. Ele garante que não vamos atropelar o buffer.
        // await fetchAndAppend(videosBaseUrl + videoChunkUrl, videoBuffer);

        // await Promise.all([
        //   fetchAndAppend(videosBaseUrl + videoChunkUrl, videoBuffer),
        //   fetchAndAppend(videosBaseUrl + audioChunkUrl, audioBuffer)
        // ]);

        let [resultadoVideo, resultadoAudio] = await Promise.all([
          baixarECalcularBanda(videosBaseUrl + videoChunkUrl),
          baixarECalcularBanda(videosBaseUrl + audioChunkUrl)
        ]);

        // Exibe a banda calculada no console
        console.log(`Banda do Vídeo: ${resultadoVideo.kbps.toFixed(2)} Kbps`);
        console.log(`Banda do Áudio: ${resultadoAudio.kbps.toFixed(2)} Kbps`);

        // Aqui está a mágica da sincronia!
        // Nós fazemos um "await" no vídeo ANTES do áudio. 
        // Isso garante que o áudio não atropele o vídeo e jogue aquele erro no Firefox.
        await injetarComSeguranca(videoBuffer, resultadoVideo.dados);
        await injetarComSeguranca(audioBuffer, resultadoAudio.dados);
              
        console.log(`Pedaço ${i} injetado! (Video + Audio)`);

        if(resultadoVideo.kbps > 30000 && i > 3 && videoRepId != 3){
          console.log("Aumentando a qualidade")
          videoRepId = 3;
          videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
          fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer);
        }

      }

      mediaSource.endOfStream();
      console.log("Loop finalizado. Buffer carregado com sucesso!");
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