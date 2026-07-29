<template>
  <n-layout has-sider style="height: 100%">
    <n-layout-sider bordered :width="siderWidth" :collapsed-width="siderWidth" :show-trigger="false">
      <div class="brand">MP</div>
      <div class="sider-nav">
        <button
          v-for="item in topMenus"
          :key="item.key"
          type="button"
          class="sider-item"
          :class="{ active: item.key === activeTopKey }"
          @click="onTopClick(item)"
        >
          <span class="sider-label">{{ item.label }}</span>
        </button>
      </div>
    </n-layout-sider>

    <n-layout>
      <n-layout-header bordered class="header">
        <div class="header-left">
          <nav v-if="secondaryMenus.length" class="subnav">
            <button
              v-for="m in secondaryMenus"
              :key="m.key"
              type="button"
              class="subnav-item"
              :class="{ active: m.key === activeSecondaryKey }"
              @click="onSecondaryClick(m.key)"
            >
              {{ m.label }}
            </button>
          </nav>
        </div>
        <n-space align="center" :wrap="false">
          <span class="user-name">{{ auth.user?.realname || auth.user?.account }}</span>
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
import { useAuthStore, type MenuNode } from '@/stores/auth'
import { useProductFilterStore } from '@/stores/productFilter'

type NavItem = {
  key: string
  label: string
  path?: string | null
  children: NavItem[]
}

const siderWidth = 72
const auth = useAuthStore()
const productFilter = useProductFilterStore()
const route = useRoute()
const router = useRouter()

function toNavItems(nodes: MenuNode[]): NavItem[] {
  return nodes
    .filter((n) => n.type !== 'button')
    .map((n) => ({
      key: n.code,
      label: n.name,
      path: n.path,
      children: n.children?.length ? toNavItems(n.children) : [],
    }))
}

const topMenus = computed(() => toNavItems(auth.menus))

function collectPaths(item: NavItem): string[] {
  const paths: string[] = []
  if (item.path) paths.push(item.path)
  for (const c of item.children) paths.push(...collectPaths(c))
  return paths
}

function matchScore(basePath: string, current: string): number {
  if (!basePath) return -1
  if (current === basePath || current.startsWith(basePath + '/')) return basePath.length
  return -1
}

const activeTopKey = computed(() => {
  const path = route.path
  let best: { key: string; score: number } | null = null
  for (const item of topMenus.value) {
    for (const p of collectPaths(item)) {
      const score = matchScore(p, path)
      if (score > (best?.score ?? -1)) best = { key: item.key, score }
    }
    // 顶级自身 path（如工作台）
    if (item.path) {
      const score = matchScore(item.path, path)
      if (score > (best?.score ?? -1)) best = { key: item.key, score }
    }
  }
  return best?.key || topMenus.value[0]?.key || ''
})

const activeTopMenu = computed(() => topMenus.value.find((m) => m.key === activeTopKey.value) || null)

const secondaryMenus = computed(() => {
  const top = activeTopMenu.value
  if (!top) return []
  return top.children.filter((c) => !!c.path)
})

const activeSecondaryKey = computed(() => {
  const path = route.path
  let best: { key: string; score: number } | null = null
  for (const m of secondaryMenus.value) {
    if (!m.path) continue
    const score = matchScore(m.path, path)
    if (score > (best?.score ?? -1)) best = { key: m.key, score }
  }
  return best?.key || secondaryMenus.value[0]?.key || ''
})

function firstLeafPath(item: NavItem): string | null {
  if (item.path) return item.path
  for (const c of item.children) {
    const p = firstLeafPath(c)
    if (p) return p
  }
  return null
}

function navigateWithProduct(path: string) {
  const target = productFilter.withProductQuery(path)
  router.push(target)
}

function onTopClick(item: NavItem) {
  const target = firstLeafPath(item)
  if (target) navigateWithProduct(target)
}

function onSecondaryClick(key: string) {
  const m = secondaryMenus.value.find((x) => x.key === key)
  if (m?.path) navigateWithProduct(m.path)
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
  font-size: 16px;
  letter-spacing: 0.5px;
  border-bottom: 1px solid #efeff5;
}
.sider-nav {
  padding: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.sider-item {
  appearance: none;
  border: 0;
  background: transparent;
  cursor: pointer;
  margin: 0 8px;
  padding: 10px 4px;
  border-radius: 8px;
  color: #333;
  text-align: center;
  line-height: 1.25;
}
.sider-item:hover {
  background: #f3f3f5;
}
.sider-item.active {
  background: #e8f3ff;
  color: #2080f0;
  font-weight: 600;
}
.sider-label {
  display: block;
  font-size: 13px;
  word-break: break-all;
}
.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: #fff;
  gap: 16px;
}
.header-left {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  display: flex;
  align-items: stretch;
  height: 100%;
}
.subnav {
  display: flex;
  align-items: stretch;
  gap: 4px;
  height: 100%;
}
.subnav-item {
  appearance: none;
  border: 0;
  background: transparent;
  cursor: pointer;
  padding: 0 14px;
  color: #666;
  font-size: 14px;
  position: relative;
}
.subnav-item:hover {
  color: #2080f0;
}
.subnav-item.active {
  color: #2080f0;
  font-weight: 600;
}
.subnav-item.active::after {
  content: '';
  position: absolute;
  left: 10px;
  right: 10px;
  bottom: 0;
  height: 2px;
  background: #2080f0;
  border-radius: 1px;
}
.user-name {
  color: #666;
  white-space: nowrap;
}
</style>
