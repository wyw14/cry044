import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

const pinia = createPinia()
const hall = createApp(App)
hall.use(pinia)
hall.use(router)
hall.mount('#app')
