import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import HomeView from './views/HomeView.vue'
import ArticleView from './views/ArticleView.vue'
import AuthView from './views/AuthView.vue'
import AdminView from './views/AdminView.vue'
import './styles.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/article/:slug', component: ArticleView },
    { path: '/auth', component: AuthView },
    { path: '/admin', component: AdminView },
  ],
})

createApp(App).use(router).mount('#app')

