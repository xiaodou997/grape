import { createRouter, createWebHistory } from 'vue-router'
import AboutView from '../views/AboutView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/about',
      name: 'about',
      component: AboutView,
    },
  ],
})

export default router
