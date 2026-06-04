<script setup>
  import { onMounted } from 'vue';
import HelloWorld from './components/HelloWorld.vue';
  const url = "http://localhost:8080/manifesto/manifesto.mpd";

  async function buscarManifesto() {
    try {

      // --------------------------------------------------------------------
      // Teste do fecth do manifesto
      const response = await fetch(url); 
      if (!response.ok) {
        throw new Error(`Erro HTTP: ${response.status}`);
      }
 
      const result = await response.text();
      //console.log(result);

      const parser = new DOMParser();
      const manifest = parser.parseFromString(result, "application/xml");

      //Pega <> do manifesto
      const segmento = manifest.querySelector("SegmentTemplate");
      console.log(segmento);
      // const segmentoAll = manifest.querySelectorAll("SegmentTemplate");
      // console.log(segmentoAll);

      // Extrai informcoes do <>
      const nomeChunk = segmento.getAttribute("media");
      console.log(nomeChunk);

      // ---------------------------------------------------------------------
      // Teste do fetch dos chunks
      const chunk = await fetch("http://localhost:8080/videos/chunk_2_2.m4s");
      if(chunk.ok){
        console.log("Busquei o pedaco");
      } else {
        console.log("Nao deu certo")
      }

    } catch (error) {
      console.error("Falha ao buscar o manifesto:", error.message);
    }
}

  onMounted(() => {
    buscarManifesto();
  })

</script>

<template>
  <!-- <header>
    <img alt="Vue logo" class="logo" src="./assets/logo.svg" width="125" height="125" />
    <div class="wrapper">
      <HelloWorld msg="You did it!" />
    </div>
  </header>
  <RouterLink to="/player/">Go to player</RouterLink>
  <main>
    <TheWelcome />
  </main> -->
    <!-- <div style="padding: 20px;">
    <h2>Redes - Tela Inicial</h2>
    <p>Se você está lendo isso, a tela deixou de ser invisível!</p>
    
    <RouterLink to="/player">
      <button style="padding: 10px; cursor: pointer;">Ir para o Player</button>
    </RouterLink>
  </div> -->
  
  <header>
    <img alt="Vue logo" class="logo" src="./assets/redeflix.svg" width="75" height="75" />
    <div class="wrapper">
      <HelloWorld msg="RedeFlix" />
    </div>
  </header>

  <a href="/player" style="padding: 10px; cursor: pointer;">Ir para o Player de Vídeo</a>

  <main>
    <!-- <TheWelcome /> -->
  </main>

</template>

<style scoped>
header {
  line-height: 1.5;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>