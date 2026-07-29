<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">项目列表</n-h2>
      <n-button v-if="auth.has('project.create')" type="primary" @click="goCreate">新建项目</n-button>
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
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listProjects, deleteProject } from '@/api'
import { useAuthStore } from '@/stores/auth'
import EntityDrawer from '@/components/EntityDrawer.vue'
import { useRouteDrawer } from '@/composables/useRouteDrawer'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const { drawerOpen, width, title, onUpdateShow } = useRouteDrawer({
  listPath: '/projects',
  drawerNames: ['project-new', 'project-detail', 'project-edit'],
  keepQueryKeys: ['productId'],
  titles: {
    'project-new': '新建项目',
    'project-detail': '项目详情',
    'project-edit': '编辑项目',
  },
  widths: { 'project-detail': 900 },
})
const list = ref<any[]>([])
const loading = ref(false)
const productId = ref<number | null>(null)
const status = ref<string | null>(null)
const keyword = ref('')
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const statusMap: Record<string, string> = {
  wait: '未开始',
  doing: '进行中',
  suspended: '已挂起',
  closed: '已关闭',
}
const statusOptions = Object.entries(statusMap).map(([value, label]) => ({ label, value }))

function fmtDate(v: string | null | undefined) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name' },
  { title: '代号', key: 'code', width: 100, render: (r) => r.code || '-' },
  { title: '所属产品', key: 'productName', width: 140 },
  { title: '状态', key: 'status', width: 90, render: (r) => statusMap[r.status] || r.status },
  {
    title: 'PM',
    key: 'pm',
    width: 100,
    render: (r) => (r.pmUser ? r.pmUser.realname || r.pmUser.account : '-'),
  },
  { title: '开始', key: 'begin', width: 110, render: (r) => fmtDate(r.begin) },
  { title: '结束', key: 'end', width: 110, render: (r) => fmtDate(r.end) },
  { title: '迭代数', key: 'sprintCount', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/projects/${row.id}`) }, { default: () => '查看' }),
          auth.has('project.edit')
            ? h(NButton, { text: true, onClick: () => router.push(`/projects/${row.id}/edit`) }, { default: () => '编辑' })
            : null,
          auth.has('project.delete')
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

function goCreate() {
  const q = productId.value ? `?productId=${productId.value}` : ''
  router.push(`/projects/new${q}`)
}

async function load() {
  loading.value = true
  try {
    const res: any = await listProjects({
      page: pagination.page,
      pageSize: pagination.pageSize,
      productId: productId.value || undefined,
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

function onPage(p: number) {
  pagination.page = p
  load()
}

async function onDelete(row: any) {
  try {
    await deleteProject(row.id)
    message.success('已删除')
    load()
  } catch (e: any) {
    message.error(e.message)
  }
}

onMounted(() => {
  // 从产品详情等入口带入 ?productId=，走 WHERE 筛选
  const q = route.query.productId
  if (q) productId.value = Number(q)
  load()
})
</script>
