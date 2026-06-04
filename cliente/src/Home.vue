<script setup>
  import { onMounted } from 'vue';
import Card from './components/ui/card/Card.vue';
import CardContent from './components/ui/card/CardContent.vue';
import Carousel from './components/ui/carousel/Carousel.vue';
import CarouselContent from './components/ui/carousel/CarouselContent.vue';
import CarouselItem from './components/ui/carousel/CarouselItem.vue';
import CarouselNext from './components/ui/carousel/CarouselNext.vue';
import CarouselPrevious from './components/ui/carousel/CarouselPrevious.vue';

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
  
  <header class="h-16 flex flex-wrap justify-center sm:justify-start items-center md:py-6 px-4 border-b">
    <div class="wrapper py-5 gap-4 lg:pt-0">
      <img alt="Vue logo" class="logo" src="./assets/redeflix.svg" width="24" height="24" />
      <h3 
        class="scroll-m-20 text-2xl font-semibold tracking-tight"
        >
        Redeflix</h3>
      <!-- <HelloWorld msg="RedeFlix" /> -->
    </div>
  </header>

  
  <main>
    <a href="/player" style="padding: 10px; cursor: pointer;">Ir para o Player de Vídeo</a>

    <Carousel 
      class="relative w-full max-w-5xl pt-12"
      :opts="{
        align: 'start',
      }"
      >
      <CarouselContent>
        <CarouselItem v-for="i in 3" :key="i" class="basis 1 sm:basis-1/2 lg:basis-1/3">
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
  /* display: flex;
  place-items: center;
  line-height: 1.5;
  padding-bottom: 2rem;
  gap: 5px; */
}


header .wrapper {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
}

@media (min-width: 1024px) {
  header {
    place-items: center;
    width: 1280px;
    height: 72px;
    /* padding-right: calc(var(--section-gap) / 2); */
    /* padding-left: calc(var(--section-gap) / 2); */
  }

}
</style>