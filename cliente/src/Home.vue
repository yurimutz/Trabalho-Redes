<script setup>
import { nextTick, onMounted, ref } from 'vue';
import Header from './components/Header.vue';
import Carousel from './components/ui/carousel/Carousel.vue';
import CarouselContent from './components/ui/carousel/CarouselContent.vue';
import CarouselItem from './components/ui/carousel/CarouselItem.vue';
import CarouselNext from './components/ui/carousel/CarouselNext.vue';
import CarouselPrevious from './components/ui/carousel/CarouselPrevious.vue';

import { gsap } from "gsap";
    
import { ScrollTrigger } from "gsap/ScrollTrigger";
gsap.registerPlugin(ScrollTrigger);

const ipDoServidor = window.location.hostname;
const urlCatalogo = `http://${ipDoServidor}:8050/catalogo`;
const urlVideos = `http://${ipDoServidor}:8080/`;

const videos = ref([]);

async function buscarCatalogo() {
  try {
    const response = await fetch(urlCatalogo);
    if (!response.ok) throw new Error(`Erro HTTP: ${response.status}`);
    
    const dados = await response.json();
    videos.value = dados;
    
    // Aguarda o DOM ser atualizado com os novos elementos do v-for
    await nextTick();
    
    // Chama a função de animação após os vídeos estarem na tela
    animarEntradaDosVideos();

  } catch (error) {
    console.error("Falha ao buscar os dados JSON:", error.message);
  }
}

function animarEntradaDosVideos() {
  // Pega todos os cards de vídeo renderizados
  gsap.from('.video-card', {
    scrollTrigger: {
      trigger: '.catalogo-container', // Elemento que dispara a animação
      start: 'top 80%', // Começa quando o topo do container atingir 80% da altura da tela
    },
    y: 50, // Começa 50px mais abaixo
    opacity: 0, // Começa invisível
    duration: 0.8,
    stagger: 0.15, // Efeito cascata: 0.15s de atraso entre cada vídeo
    ease: 'power.out'
  });
}

onMounted(() => {
  buscarCatalogo();
});
</script>

<template>
  <Header />

  <main class="mt-9 catalogo-container">
    <h1 class="scroll-m-20 text-4xl font-extrabold tracking-tight text-balance pl-5 lg:pl-10">Catálogo</h1>

    <Carousel 
      class="w-9/10 relative pt-8 ml-5 lg:ml-10"
      :opts="{ align: 'center' }"
    >
      <CarouselContent>
        <CarouselItem 
          v-for="video in videos" 
          :key="video.nome" 
          class="sm:basis-1/2 lg:basis-1/3 px-2">
          
          <div class="video-card p-1 w-full flex flex-col items-center">
             <div v-if="video.thumbnail" class="w-full overflow-hidden rounded-md drop-shadow-lg">
              <router-link :to="{ path: '/player', query : { m: video.manifesto } }">
                <img
                  :src="`${urlVideos + video.thumbnail}`"
                  :alt="video.nome"
                  class="w-full aspect-video object-cover rounded-md transition-all duration-300 hover:scale-105"
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
          </div>
        </CarouselItem>
      </CarouselContent>
      <CarouselPrevious v-if="videos.length > 3"/>
      <CarouselNext v-if="videos.length > 3"/>
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
  }
}
</style>