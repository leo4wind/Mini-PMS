import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProductFilterStore } from '@/stores/productFilter'

/** 列表页：产品筛选与 URL / store / localStorage 同步 */
export function useListProductFilter() {
  const route = useRoute()
  const router = useRouter()
  const store = useProductFilterStore()
  const productId = ref<number | null>(null)

  function applyToUrl(id: number | null) {
    const query: Record<string, any> = { ...route.query }
    if (id == null) delete query.productId
    else query.productId = String(id)

    const urlHas = route.query.productId != null ? String(route.query.productId) : null
    const want = id != null ? String(id) : null
    if (urlHas !== want) {
      router.replace({ query })
    }
  }

  /** 进入列表：URL 优先，否则用 store，并回写 URL */
  function initFromRouteAndStore() {
    if (route.query.productId != null && String(route.query.productId) !== '') {
      const id = Number(route.query.productId)
      if (Number.isFinite(id) && id > 0) {
        store.setProductId(id)
        productId.value = id
        return
      }
    }
    productId.value = store.productId
    applyToUrl(productId.value)
  }

  function setProductId(id: number | null) {
    store.setProductId(id)
    productId.value = id
    applyToUrl(id)
  }

  return {
    productId,
    initFromRouteAndStore,
    setProductId,
  }
}
