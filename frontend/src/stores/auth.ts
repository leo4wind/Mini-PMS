import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as api from '@/api'

export type MenuNode = {
  id: number
  code: string
  name: string
  type: string
  path?: string | null
  icon?: string | null
  children?: MenuNode[]
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('minipms_token') || '')
  const user = ref<any>(null)
  const roles = ref<any[]>([])
  const menus = ref<MenuNode[]>([])

  const isLogin = computed(() => !!token.value)

  const codes = computed(() => {
    const set = new Set<string>()
    const walk = (nodes: MenuNode[]) => {
      for (const n of nodes) {
        set.add(n.code)
        if (n.children?.length) walk(n.children)
      }
    }
    walk(menus.value)
    return set
  })

  function has(code: string) {
    return codes.value.has(code)
  }

  async function login(account: string, password: string) {
    const res: any = await api.login(account, password)
    token.value = res.data.token
    localStorage.setItem('minipms_token', token.value)
    await loadMe()
  }

  async function loadMe() {
    const res: any = await api.fetchMe()
    user.value = res.data.user
    roles.value = res.data.roles || []
    menus.value = res.data.menus || []
  }

  function logoutLocal() {
    token.value = ''
    user.value = null
    roles.value = []
    menus.value = []
    localStorage.removeItem('minipms_token')
  }

  async function logout() {
    try {
      await api.logout()
    } finally {
      logoutLocal()
    }
  }

  return { token, user, roles, menus, isLogin, has, login, loadMe, logout, logoutLocal }
})
