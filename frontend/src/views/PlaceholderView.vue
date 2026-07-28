<template>
  <n-card :title="(route.meta.title as string) || '功能开发中'">
    <n-alert type="info" title="占位页">
      该页面路由已按 pages.md 预留，后续按 api.md 对接业务接口。
    </n-alert>
  </n-card>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>
