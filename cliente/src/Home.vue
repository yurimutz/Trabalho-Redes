<script setup>
  import { onMounted } from 'vue';
import Card from './components/ui/card/Card.vue';
import CardContent from './components/ui/card/CardContent.vue';
import Carousel from './components/ui/carousel/Carousel.vue';
import CarouselContent from './components/ui/carousel/CarouselContent.vue';
import CarouselItem from './components/ui/carousel/CarouselItem.vue';
import CarouselNext from './components/ui/carousel/CarouselNext.vue';
import CarouselPrevious from './components/ui/carousel/CarouselPrevious.vue';

  const url = "http://localhost:8050/catalogo";

async function testeGet(){

    try{

      const teste = await fetch(url);
      if (!teste.ok) throw new Error(`Erro HTTP: ${teste.status}`);
      const dados = await teste.json();
      console.log(dados);

    } catch (error) {
      console.error("Falha ao buscar os dados JSON:", error.message);
    }
}

  onMounted(() => {
    testeGet();
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