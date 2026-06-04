import { createRouter, createWebHistory } from 'vue-router'

// Telas principais
import Home from '../Home.vue' 
import Player from '../Player.vue'

const rotas = [
  {
    path: '/',
    name: 'home',
    component: Home
  },
  {
    path: '/player',
    name: 'player',
    component: Player
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: rotas
})

export default router