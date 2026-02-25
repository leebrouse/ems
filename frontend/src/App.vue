<script setup lang="ts">
/**
 * 根组件：
 * - 提供全局 Element Plus 配置容器
 * - 渲染路由出口（实际页面由 router 控制）
 */
import { RouterView } from 'vue-router'
</script>

<template>
  <el-config-provider>
    <RouterView />
  </el-config-provider>
</template>

<style>
/* Global transitions can go here */
</style>
