<script setup>
  import { onMounted } from 'vue';
import HelloWorld from './components/HelloWorld.vue';
import Card from './components/ui/card/Card.vue';
import CardContent from './components/ui/card/CardContent.vue';
import Carousel from './components/ui/carousel/Carousel.vue';
import CarouselContent from './components/ui/carousel/CarouselContent.vue';
import CarouselItem from './components/ui/carousel/CarouselItem.vue';
import CarouselNext from './components/ui/carousel/CarouselNext.vue';
import CarouselPrevious from './components/ui/carousel/CarouselPrevious.vue';

import { useColorMode } from '@vueuse/core';
const mode = useColorMode();
mode.value = 'auto'

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
    <div class="wrapper">
      <img alt="Vue logo" class="logo" src="./assets/redeflix.svg" width="75" height="75" />
      <!-- <h2>Redeflix</h2> -->
      <HelloWorld msg="RedeFlix" />
    </div>
  </header>

  
  <main>
    <a href="/player" style="padding: 10px; cursor: pointer;">Ir para o Player de Vídeo</a>

    <Carousel class="w-full">
      <CarouselContent>
        <CarouselItem v-for="i in 5" :key="i">
        <div class="p1">
          <Card>
            <CardContent class="flex items-center justify-center p-6">
              <span class="text-3xl font-semibold">{{ i }}</span>
            </CardContent>
          </Card>
        </div>
        </CarouselItem>
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  </main>

</template>

<style scoped>
header {
  line-height: 1.5;
  /* padding-bottom: 2rem; */
  gap: 5px;
  border-bottom: 1px solid var(--vt-c-divider-dark-1);
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    width: 1280px;
    height: 72px;
    /* padding-right: calc(var(--section-gap) / 2); */
    padding-left: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
    justify-content: center;
  }
}
</style>