<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">产品列表</n-h2>
      <n-button v-if="auth.has('product.create')" type="primary" @click="$router.push('/products/new')">
        新建产品
      </n-button>
    </n-space>
    <n-space>
      <n-select v-model:value="status" :options="statusOptions" clearable placeholder="状态" style="width: 140px" />
      <n-input v-model:value="keyword" placeholder="搜索名称/代号" style="width: 220px" clearable />
      <n-button @click="load">查询</n-button>
    </n-space>
    <n-data-table :columns="columns" :data="list" :loading="loading" :pagination="pagination" remote @update:page="onPage" />

    <EntityDrawer :show="drawerOpen" :title="title" :width="width" @update:show="onUpdateShow">
      <router-view />
    </EntityDrawer>
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NSpace, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listProducts, deleteProduct, updateProduct } from '@/api'
import { useAuthStore } from '@/stores/auth'
import EntityDrawer from '@/components/EntityDrawer.vue'
import { provideListReload, useRouteDrawer } from '@/composables/useRouteDrawer'

const auth = useAuthStore()
const router = useRouter()
const message = useMessage()
const { drawerOpen, width, title, onUpdateShow } = useRouteDrawer({
  listPath: '/products',
  drawerNames: ['product-new', 'product-detail', 'product-edit'],
  titles: {
    'product-new': '新建产品',
    'product-detail': '产品详情',
    'product-edit': '编辑产品',
  },
  widths: { 'product-detail': 900 },
})
const list = ref<any[]>([])
const loading = ref(false)
const status = ref<string | null>(null)
const keyword = ref('')
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const statusOptions = [
  { label: '正常', value: 'normal' },
  { label: '关闭', value: 'closed' },
]

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '名称', key: 'name' },
  { title: '代号', key: 'code' },
  { title: '状态', key: 'status', width: 100 },
  {
    title: '创建人',
    key: 'creator',
    width: 100,
    render: (r) => (r.creator ? r.creator.realname || r.creator.account : '-'),
  },
  { title: '创建时间', key: 'createdAt', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 260,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/products/${row.id}`) }, { default: () => '查看' }),
          auth.has('product.edit')
            ? h(NButton, { text: true, onClick: () => router.push(`/products/${row.id}/edit`) }, { default: () => '编辑' })
            : null,
          auth.has('product.edit')
            ? h(
                NButton,
                {
                  text: true,
                  onClick: () => toggleStatus(row),
                },
                { default: () => (row.status === 'normal' ? '关闭' : '启用') },
              )
            : null,
          auth.has('product.delete')
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

async function load() {
  loading.value = true
  try {
    const res: any = await listProducts({
      page: pagination.page,
      pageSize: pagination.pageSize,
      status: status.value || undefined,
      keyword: keyword.value || undefined,
    })
    list.value = res.data.list || []
    pagination.itemCount = res.data.total || 0
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

provideListReload(load)

function onPage(p: number) {
  pagination.page = p
  load()
}

async function toggleStatus(row: any) {
  try {
    await updateProduct(row.id, { status: row.status === 'normal' ? 'closed' : 'normal' })
    message.success('已更新')
    load()
  } catch (e: any) {
    message.error(e.message)
  }
}

async function onDelete(row: any) {
  try {
    await deleteProduct(row.id)
    message.success('已删除')
    load()
  } catch (e: any) {
    message.error(e.message)
  }
}

onMounted(load)
</script>
