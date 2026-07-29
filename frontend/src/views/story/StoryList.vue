<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">需求列表</n-h2>
      <n-button v-if="auth.has('story.create')" type="primary" @click="goCreate">新建需求</n-button>
    </n-space>

    <n-space>
      <ProductFilterSelect :model-value="productId" @update:model-value="onProductChange" />
      <n-select v-model:value="type" :options="typeOptions" clearable placeholder="类型" style="width: 120px" />
      <n-select v-model:value="status" :options="statusOptions" clearable placeholder="状态" style="width: 120px" />
      <n-input v-model:value="keyword" placeholder="搜索标题" style="width: 220px" clearable />
      <n-button @click="load">查询</n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      remote
      @update:page="onPage"
    />

    <EntityDrawer :show="drawerOpen" :title="title" :width="width" @update:show="onUpdateShow">
      <router-view />
    </EntityDrawer>
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listStories, deleteStory } from '@/api'
import { storyTypeMap, storyStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'
import EntityDrawer from '@/components/EntityDrawer.vue'
import ProductFilterSelect from '@/components/ProductFilterSelect.vue'
import { provideListReload, useRouteDrawer } from '@/composables/useRouteDrawer'
import { useListProductFilter } from '@/composables/useListProductFilter'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const { productId, initFromRouteAndStore, setProductId } = useListProductFilter()
const { drawerOpen, width, title, onUpdateShow } = useRouteDrawer({
  listPath: '/stories',
  drawerNames: ['story-new', 'story-detail', 'story-edit'],
  keepQueryKeys: ['productId', 'assignedTo', 'type', 'status'],
  titles: {
    'story-new': '新建需求',
    'story-detail': '需求详情',
    'story-edit': '编辑需求',
  },
})

const list = ref<any[]>([])
const loading = ref(false)
const type = ref<string | null>(null)
const status = ref<string | null>(null)
const keyword = ref('')
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const typeOptions = Object.entries(storyTypeMap).map(([value, label]) => ({ label, value }))
const statusOptions = Object.entries(storyStatusMap).map(([value, label]) => ({ label, value }))

function renderUser(u: any) {
  if (!u) return '-'
  return u.realname || u.account || '-'
}

function fmtDate(v: string | null | undefined) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '产品', key: 'productName', width: 120, render: (r) => r.productName || '-' },
  { title: '类型', key: 'type', width: 90, render: (r) => storyTypeMap[r.type] || r.type },
  { title: '优先级', key: 'pri', width: 80 },
  { title: '状态', key: 'status', width: 80, render: (r) => storyStatusMap[r.status] || r.status },
  { title: '估算', key: 'estimate', width: 80, render: (r) => (r.estimate != null ? r.estimate : '-') },
  { title: '指派人', key: 'assignee', width: 100, render: (r) => renderUser(r.assignee) },
  { title: '创建人', key: 'creator', width: 100, render: (r) => renderUser(r.creator) },
  { title: '附件', key: 'attachCount', width: 70, render: (r) => r.attachCount ?? 0 },
  { title: '更新时间', key: 'updatedAt', width: 110, render: (r) => fmtDate(r.updatedAt) },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/stories/${row.id}`) }, { default: () => '查看' }),
          auth.has('story.edit')
            ? h(NButton, { text: true, onClick: () => router.push(`/stories/${row.id}/edit`) }, { default: () => '编辑' })
            : null,
          auth.has('story.delete')
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

function goCreate() {
  const q = productId.value ? `?productId=${productId.value}` : ''
  router.push(`/stories/new${q}`)
}

function onProductChange(id: number | null) {
  setProductId(id)
  pagination.page = 1
  load()
}

async function load() {
  loading.value = true
  try {
    const res: any = await listStories({
      page: pagination.page,
      pageSize: pagination.pageSize,
      productId: productId.value || undefined,
      type: type.value || undefined,
      status: status.value || undefined,
      assignedTo: route.query.assignedTo ? String(route.query.assignedTo) : undefined,
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

function onDelete(row: any) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除需求「${row.title}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteStory(row.id)
        message.success('已删除')
        load()
      } catch (e: any) {
        message.error(e.message)
      }
    },
  })
}

onMounted(() => {
  if (route.query.type) type.value = String(route.query.type)
  if (route.query.status) status.value = String(route.query.status)
  initFromRouteAndStore()
  load()
})
</script>
