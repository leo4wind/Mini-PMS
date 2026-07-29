import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'minipms.productId'

const PRODUCT_LIST_ROOTS = ['/stories', '/projects', '/sprints', '/bugs'] as const

function readStored(): number | null {
  const v = localStorage.getItem(STORAGE_KEY)
  if (!v) return null
  const n = Number(v)
  return Number.isFinite(n) && n > 0 ? n : null
}

export const useProductFilterStore = defineStore('productFilter', () => {
  const productId = ref<number | null>(readStored())

  function setProductId(id: number | null) {
    productId.value = id
    if (id == null) localStorage.removeItem(STORAGE_KEY)
    else localStorage.setItem(STORAGE_KEY, String(id))
  }

  /** 列表根路径且已选产品时，附带 productId query */
  function withProductQuery(path: string): { path: string; query?: { productId: string } } {
    const base = path.split('?')[0]
    if (!PRODUCT_LIST_ROOTS.includes(base as (typeof PRODUCT_LIST_ROOTS)[number])) {
      return { path }
    }
    if (productId.value == null) return { path: base }
    return { path: base, query: { productId: String(productId.value) } }
  }

  function isProductListRoot(path: string) {
    return PRODUCT_LIST_ROOTS.includes(path as (typeof PRODUCT_LIST_ROOTS)[number])
  }

  return {
    productId,
    setProductId,
    withProductQuery,
    isProductListRoot,
  }
})
