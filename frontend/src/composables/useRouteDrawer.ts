import { computed, inject, onBeforeUnmount, provide, ref, watch, type InjectionKey, type Ref } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'

export const drawerTitleKey: InjectionKey<Ref<string>> = Symbol('drawerTitle')
export const drawerCloseKey: InjectionKey<() => void> = Symbol('drawerClose')

export type RouteDrawerOptions = {
  listPath: string
  drawerNames: string[]
  /** 关闭时保留的 query 键；不传则保留当前全部 query */
  keepQueryKeys?: string[]
  titles: Record<string, string>
  widths?: Partial<Record<string, number>>
}

const DEFAULT_FORM_WIDTH = 640
const DEFAULT_DETAIL_WIDTH = 800

export function useRouteDrawer(options: RouteDrawerOptions) {
  const route = useRoute()
  const router = useRouter()
  const drawerTitle = ref('')
  provide(drawerTitleKey, drawerTitle)

  const drawerOpen = computed(() => options.drawerNames.includes(String(route.name)))

  const width = computed(() => {
    const name = String(route.name || '')
    if (options.widths?.[name]) return options.widths[name]!
    if (name.endsWith('-new') || name.endsWith('-edit')) return DEFAULT_FORM_WIDTH
    return DEFAULT_DETAIL_WIDTH
  })

  const title = computed(() => {
    if (drawerTitle.value) return drawerTitle.value
    return options.titles[String(route.name)] || ''
  })

  function pickQuery(): LocationQuery {
    if (!options.keepQueryKeys) return { ...route.query }
    const q: LocationQuery = {}
    for (const key of options.keepQueryKeys) {
      if (route.query[key] != null) q[key] = route.query[key]
    }
    return q
  }

  function closeDrawer() {
    drawerTitle.value = ''
    router.push({ path: options.listPath, query: pickQuery() })
  }

  provide(drawerCloseKey, closeDrawer)

  function onUpdateShow(show: boolean) {
    if (!show && drawerOpen.value) closeDrawer()
  }

  return {
    drawerOpen,
    width,
    title,
    closeDrawer,
    onUpdateShow,
  }
}

export function useCloseDrawer() {
  return inject(drawerCloseKey, null)
}

/** 详情页把实体标题同步到抽屉；卸载时清空覆盖 */
export function useSyncDrawerTitle(getter: () => string | undefined | null, fallback: string) {
  const titleRef = inject(drawerTitleKey, null)
  if (!titleRef) return

  watch(
    () => getter() || fallback,
    (v) => {
      titleRef.value = v
    },
    { immediate: true },
  )

  onBeforeUnmount(() => {
    titleRef.value = ''
  })
}
