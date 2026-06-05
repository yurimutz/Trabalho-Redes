<script setup>
  import { useColorMode } from '@vueuse/core';
  const mode = useColorMode();
  mode.value = 'auto'
</script>


<template>
  <router-view></router-view>
</template>