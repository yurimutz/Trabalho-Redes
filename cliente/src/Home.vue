<script setup>
  import { onMounted, ref } from 'vue';
import Header from './components/Header.vue';
import Carousel from './components/ui/carousel/Carousel.vue';
import CarouselContent from './components/ui/carousel/CarouselContent.vue';
import CarouselItem from './components/ui/carousel/CarouselItem.vue';
import CarouselNext from './components/ui/carousel/CarouselNext.vue';
import CarouselPrevious from './components/ui/carousel/CarouselPrevious.vue';

  // const urlCatalogo = "http://localhost:8050/catalogo";
  // const urlVideos = "http://localhost:8080/"
  const ipDoServidor = window.location.hostname;
  const urlCatalogo = `http://${ipDoServidor}:8050/catalogo`;
  const urlVideos = `http://${ipDoServidor}:8080/`
  //`http://${ipDoServidor}:8080/${mParam}`
  //192.168.100.86

  const videos = ref([]);

  async function buscarCatalogo(){

    try {
      const response = await fetch(urlCatalogo);
      if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);
      const dados = await response.json();
      videos.value = dados;
      console.log(dados);
    } catch (error) {
      console.error("Falha ao buscar os dados JSON:", error.message);
    }
}

  onMounted(() => {
    buscarCatalogo();
  })

</script>

<template>
  
  <Header />

  <main class="mt-9">

    <h1 class="scroll-m-20 text-4xl font-extrabold tracking-tight text-balance">Catálogo</h1>

    <Carousel 
      class="w-9/10 relative pt-8 ml-5 lg:ml-10"
      :opts="{
        align: 'center',
      }"
      >
      <CarouselContent>
        <CarouselItem 
          v-for="video in videos" 
          :key="video.nome" 
          class="sm:basis-1/2 lg:basis-1/3 px-2">
        <div class="p1 w-full flex flex-col items-center">
          <!-- <Card> -->
            <!-- <CardContent class="flex flex-col items-center justify-center p-0"> -->
              <!-- <span class="text-3xl font-semibold">{{ i }}</span> -->
               <div v-if="video.thumbnail" class="w-full overflow-hidden rounded-md drop-shadow-lg">
                <router-link :to="{ path: '/player', query : { m: video.manifesto } }">
                  <img
                    :src="`${urlVideos + video.thumbnail}`"
                    :alt="video.nome"
                    class="w-full aspect-video object-cover rounded-md transition-all duration-300 hover:scale-110"
                  />
                </router-link>
               </div>
              <div
                v-else
                class="w-full aspect-video bg-muted flex items-center justify-center rounded-md drop-shadow-md"
              >
                <span class="text-muted-foreground text-sm">sem thumbnail</span>
              </div>
                <span class="text-sm font-medium py-2">{{ video.nome }}</span>
            <!-- </CardContent> -->
          <!-- </Card> -->
        </div>
        </CarouselItem>
      </CarouselContent>
      <CarouselPrevious
        v-if="videos.length > 3"/>
      <CarouselNext 
        v-if="videos.length > 3"/>
    </Carousel>
  </main>

</template>

<style scoped>

header .wrapper {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
}

@media (min-width: 1024px) {
  header {
    place-items: center;
    width: 100%;
    height: 72px;
    /* padding-right: calc(var(--section-gap) / 2); */
    /* padding-left: calc(var(--section-gap) / 2); */
  }

}
</style>