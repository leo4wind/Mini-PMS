<template>
  <n-layout has-sider style="height: 100%">
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="64" :width="220" show-trigger>
      <div class="brand">MiniPMS</div>
      <n-menu
        :value="activeKey"
        :options="menuOptions"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        @update:value="onMenu"
      />
    </n-layout-sider>
    <n-layout>
      <n-layout-header bordered class="header">
        <div />
        <n-space align="center">
          <span>{{ auth.user?.realname || auth.user?.account }}</span>
          <n-button text @click="$router.push('/profile/password')">改密</n-button>
          <n-button text type="error" @click="onLogout">退出</n-button>
        </n-space>
      </n-layout-header>
      <n-layout-content content-style="padding: 16px;">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { MenuOption } from 'naive-ui'
import { useAuthStore, type MenuNode } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeKey = computed(() => route.path)

function toOptions(nodes: MenuNode[]): MenuOption[] {
  return nodes
    .filter((n) => n.type !== 'button')
    .map((n) => {
      const opt: MenuOption = {
        label: n.name,
        key: n.path || n.code,
      }
      if (n.children?.length) {
        const kids = toOptions(n.children)
        if (kids.length) opt.children = kids
      }
      return opt
    })
}

const menuOptions = computed(() => toOptions(auth.menus))

function onMenu(key: string) {
  if (key.startsWith('/')) router.push(key)
}

async function onLogout() {
  await auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.brand {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 18px;
  letter-spacing: 0.5px;
}
.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: #fff;
}
</style>
