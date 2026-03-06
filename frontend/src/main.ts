import { createApp } from 'vue'
import App from './App.vue'
import Antd from 'ant-design-vue';
import 'ant-design-vue/dist/reset.css';
import router from './router'
import '@/styles/global.scss'
import '@/styles/style.css'
import {loadLang} from "@/locales"
import pinia from '@/stores';
async function start() {
    const app = createApp(App)
    app.use(pinia)
    // 全局语言包加载
    await loadLang(app)
    app.use(router)
    app.use(Antd)

    // // 全局注册
    // directives(app) // 指令
    // registerIcons(app) // icons
    app.mount('#app')

    // app.config.globalProperties.eventBus = mitt()
}

start()