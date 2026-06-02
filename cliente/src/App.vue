<script setup>
  import { ref, onMounted, onBeforeUpdate } from 'vue'

  import HelloWorld from './components/HelloWorld.vue'
  import TheWelcome from './components/TheWelcome.vue'

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
      const chunk = await fetch("http://localhost:8080/videos/chunk_2_1.m4s");
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
  <header>
    <img alt="Vue logo" class="logo" src="./assets/logo.svg" width="125" height="125" />

    <div class="wrapper">
      <HelloWorld msg="You did it!" />
    </div>
  </header>

  <main>
    <TheWelcome />
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
