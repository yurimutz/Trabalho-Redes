<template>
    <header class="h-16 flex flex-wrap justify-center sm:justify-start items-center md:py-6 border-b">
        <div class="wrapper py-5 gap-2 lg:gap-4 md:pt-0">
            <img alt="Vue logo" class="logo" src="../assets/redeflix.svg" width="24" height="24" />
            <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">Redeflix</h3>
        </div>
    </header>


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