import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

// 示例 Store：Vite + Vue 模板自带计数器，可用于演示 Pinia 用法
export const useCounterStore = defineStore('counter', () => {
  const count = ref(0)
  const doubleCount = computed(() => count.value * 2)
  function increment() {
    count.value++
  }

  return { count, doubleCount, increment }
})
