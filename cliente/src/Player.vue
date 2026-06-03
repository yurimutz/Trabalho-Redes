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

      // 3. Extrai as informações vitais do XML
      const representation = manifest.querySelector("Representation");
      const mimeType = representation.getAttribute("mimeType"); // ex: video/mp4
      const codecs = representation.getAttribute("codecs");     // ex: avc1.64001e
      const repId = representation.getAttribute("id");          // ex: 2

      const segmento = manifest.querySelector("SegmentTemplate");
      const initTemplate = segmento.getAttribute("initialization"); // ex: init_$RepresentationID$.mp4
      const mediaTemplate = segmento.getAttribute("media");         // ex: chunk_$RepresentationID$_$Number$.m4s

      // 4. Cria o SourceBuffer no formato correto
      const codecString = `${mimeType}; codecs="${codecs}"`;
      const sourceBuffer = mediaSource.addSourceBuffer(codecString);

      // Função Mágica: Baixa um pedaço em formato binário e espera o buffer engolir
      const fetchAndAppend = async (url) => {
        const res = await fetch(url);
        const buffer = await res.arrayBuffer(); // Pega o binário (ArrayBuffer)

        return new Promise((resolve) => {
          // Só resolve a promessa quando o buffer gritar "terminei de atualizar!"
          sourceBuffer.addEventListener('updateend', resolve, { once: true });
          sourceBuffer.appendBuffer(buffer);
        });
      };

      // 5. Baixa e Injeta o Cabeçalho (Init) - Obrigatório antes dos chunks!
      // Substitui a variavel $RepresentationID$ pelo ID real
      const initFinalUrl = initTemplate.replace('$RepresentationID$', repId);
      console.log("Baixando Init:", initFinalUrl);
      await fetchAndAppend(videosBaseUrl + initFinalUrl);

      // 6. O famoso Loop de Chunks (Exemplo: buscando os 5 primeiros pedaços)
      const quantidadeDeChunks = 4; 
      
      for (let i = 1; i <= quantidadeDeChunks; i++) {
        // Monta o nome do arquivo substituindo as variaveis dinâmicas do XML
        // Dependendo de como você gerou no FFmpeg, o $Number$ pode ser $Number%05d$ (com zeros). 
        // Adapte o replace abaixo se necessário.
        let chunkFinalUrl = mediaTemplate
                              .replace('$RepresentationID$', repId)
                              .replace('$Number$', i);

        console.log(`Baixando pedaço ${i}...`);
        
        // O "await" aqui é crucial. Ele garante que não vamos atropelar o buffer.
        await fetchAndAppend(videosBaseUrl + chunkFinalUrl);
        
        console.log(`Pedaço ${i} injetado! (+4 segundos de vídeo no buffer)`);
      }

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