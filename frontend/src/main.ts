/**
 * 前端应用入口：
 * - 初始化 Vue 应用
 * - 注册 Pinia / Router / Element Plus
 * - 引入全局样式与暗黑主题
 */
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import './assets/main.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

// 统一启用暗黑模式（配合 Element Plus dark css-vars）
document.documentElement.classList.add('dark')

app.mount('#app')
