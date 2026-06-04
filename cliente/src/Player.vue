<script setup>
import { onMounted, ref } from 'vue';
const videoPlayer = ref(null);

const manifestUrl = "http://localhost:8080/manifesto/manifesto.mpd";
const videosBaseUrl = "http://localhost:8080/videos/"; // Ajuste conforme sua pasta de vídeos

async function iniciarStreaming() {
  try {
    // 1. Inicia o MediaSource e vincula à tag de vídeo
    const mediaSource = new MediaSource();
    videoPlayer.value.src = URL.createObjectURL(mediaSource);

    // Só podemos começar a injetar coisas quando o source estiver aberto
    mediaSource.addEventListener('sourceopen', async () => {
      
      // 2. Busca e analisa o manifesto (O que você já tinha feito!)
      const response = await fetch(manifestUrl);
      if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);
      const xmlText = await response.text();
      
      const parser = new DOMParser();
      const manifest = parser.parseFromString(xmlText, "application/xml");

      // Extraindo informações de vídeo
      const videoSet = manifest.querySelector('AdaptationSet[contentType="video"]');
      // console.log(videoSet);
      const videoRepresentation = videoSet.querySelector("Representation");
      const videoMimeType = videoRepresentation.getAttribute("mimeType"); // ex: video/mp4
      const videoCodecs = videoRepresentation.getAttribute("codecs");     // ex: avc1.64001e
      const videoRepId = videoRepresentation.getAttribute("id");          // ex: 2

      const videoSegment = manifest.querySelector("SegmentTemplate");
      const videoInitTemplate = videoSegment.getAttribute("initialization"); // ex: init_$RepresentationID$.mp4
      const videoMediaTemplate = videoSegment.getAttribute("media");         // ex: chunk_$RepresentationID$_$Number$.m4s

      const segmentTimeline = manifest.querySelector("SegmentTimeline");
      const sTags = segmentTimeline.querySelectorAll("S");
      let quantidadeDeChunks = 0;

      // Extraindo informações do áudio
      const audioSet = manifest.querySelector('AdaptationSet[contentType="audio"]');
      const audioRep = audioSet.querySelector("Representation");
      const audioMimeType = audioRep.getAttribute("mimeType"); 
      const audioCodecs = audioRep.getAttribute("codecs");     
      const audioRepId = audioRep.getAttribute("id");          

      const audioSegment = audioSet.querySelector("SegmentTemplate");
      const audioInitTemplate = audioSegment.getAttribute("initialization"); 
      const audioMediaTemplate = audioSegment.getAttribute("media");

      
      // Calculando a quantidade de chunks
      sTags.forEach(sTag => {
        const repeatAttr = sTag.getAttribute("r");

        // Quantidade de repeats do chunk
        const repeats = repeatAttr ? parseInt(repeatAttr) : 0;

        quantidadeDeChunks += (1 + repeats);
      })
      console.log("Total de chunks: ", quantidadeDeChunks);


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
      const videoInitUrl = videoInitTemplate.replace('$RepresentationID$', videoRepId);
      const audioInitUrl = audioInitTemplate.replace('$RepresentationID$', audioRepId);


      console.log("Baixando inits...");

      // Promisse para baixar os dois ao mesmo tempo
      await Promise.all([
        fetchAndAppend(videosBaseUrl + videoInitUrl, videoBuffer),
        fetchAndAppend(videosBaseUrl + audioInitUrl, audioBuffer)
      ]);

      // 6. O famoso Loop de Chunks (Exemplo: buscando os 5 primeiros pedaços)
      
      for (let i = 1; i <= quantidadeDeChunks; i++) {
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

        await Promise.all([
          fetchAndAppend(videosBaseUrl + videoChunkUrl, videoBuffer),
          fetchAndAppend(videosBaseUrl + audioChunkUrl, audioBuffer)
        ]);
        
        console.log(`Pedaço ${i} injetado! (Video + Audio)`);
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