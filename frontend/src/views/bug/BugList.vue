<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">缺陷列表</n-h2>
      <n-button v-if="auth.has('bug.create')" type="primary" @click="goCreate">新建缺陷</n-button>
    </n-space>
    <n-space wrap>
      <ProductFilterSelect :model-value="productId" @update:model-value="onProductFilterChange" />
      <n-select
        v-model:value="projectId"
        :options="projectOptions"
        clearable
        filterable
        placeholder="项目"
        style="width: 180px"
        @update:value="onProjectChange"
      />
      <n-select
        v-model:value="sprintId"
        :options="sprintOptions"
        clearable
        filterable
        placeholder="迭代"
        style="width: 180px"
      />
      <n-select
        v-model:value="storyId"
        :options="storyOptions"
        clearable
        filterable
        placeholder="需求"
        style="width: 180px"
      />
      <n-select v-model:value="status" :options="statusOptions" clearable placeholder="状态" style="width: 120px" />
      <n-select v-model:value="severity" :options="priOptions" clearable placeholder="严重程度" style="width: 120px" />
      <n-select v-model:value="pri" :options="priOptions" clearable placeholder="优先级" style="width: 120px" />
      <n-select
        v-model:value="assignedTo"
        :options="assigneeOptions"
        clearable
        filterable
        placeholder="指派人"
        style="width: 160px"
      />
      <n-input v-model:value="keyword" placeholder="搜索标题" style="width: 200px" clearable />
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
import { NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listBugs, deleteBug, listProjects, listSprints, listStories, listUsers } from '@/api'
import { bugStatusMap } from '@/constants/labels'
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
  listPath: '/bugs',
  drawerNames: ['bug-new', 'bug-detail', 'bug-edit'],
  keepQueryKeys: ['productId', 'projectId', 'sprintId', 'storyId', 'status', 'assignedTo'],
  titles: {
    'bug-new': '新建缺陷',
    'bug-detail': '缺陷详情',
    'bug-edit': '编辑缺陷',
  },
})

const list = ref<any[]>([])
const loading = ref(false)
const projectId = ref<number | null>(null)
const sprintId = ref<number | null>(null)
const storyId = ref<number | null>(null)
const status = ref<string | null>(null)
const severity = ref<string | null>(null)
const pri = ref<string | null>(null)
const assignedTo = ref<string | null>(null)
const keyword = ref('')

const projectOptions = ref<{ label: string; value: number }[]>([])
const sprintOptions = ref<{ label: string; value: number }[]>([])
const storyOptions = ref<{ label: string; value: number }[]>([])
const assigneeOptions = ref<{ label: string; value: string }[]>([])
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const statusOptions = Object.entries(bugStatusMap).map(([value, label]) => ({ label, value }))
const priOptions = [1, 2, 3, 4].map((v) => ({ label: String(v), value: String(v) }))

function renderUser(u: any) {
  if (!u) return '-'
  return u.realname || u.account || '-'
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '产品', key: 'productName', width: 120 },
  { title: '严重程度', key: 'severity', width: 90 },
  { title: '优先级', key: 'pri', width: 80 },
  { title: '状态', key: 'status', width: 90, render: (r) => bugStatusMap[r.status] || r.status },
  { title: '指派人', key: 'assignee', width: 100, render: (r) => renderUser(r.assignee) },
  { title: '创建人', key: 'creator', width: 100, render: (r) => renderUser(r.creator) },
  { title: '关联需求', key: 'storyId', width: 90, render: (r) => (r.storyId ? `#${r.storyId}` : '-') },
  { title: '附件', key: 'attachCount', width: 70, render: (r) => r.attachCount ?? 0 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/bugs/${row.id}`) }, { default: () => '查看' }),
          auth.has('bug.edit')
            ? h(NButton, { text: true, onClick: () => router.push(`/bugs/${row.id}/edit`) }, { default: () => '编辑' })
            : null,
          auth.has('bug.delete') && row.status === 'active'
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

function goCreate() {
  const q = new URLSearchParams()
  if (productId.value) q.set('productId', String(productId.value))
  if (projectId.value) q.set('projectId', String(projectId.value))
  if (sprintId.value) q.set('sprintId', String(sprintId.value))
  if (storyId.value) q.set('storyId', String(storyId.value))
  const qs = q.toString()
  router.push(`/bugs/new${qs ? `?${qs}` : ''}`)
}

async function loadProjects() {
  const res: any = await listProjects({
    page: 1,
    pageSize: 100,
    productId: productId.value || undefined,
  })
  projectOptions.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
}

async function loadSprints() {
  const res: any = await listSprints({
    page: 1,
    pageSize: 100,
    projectId: projectId.value || undefined,
  })
  sprintOptions.value = (res.data.list || []).map((s: any) => ({ label: s.name, value: s.id }))
}

async function loadStories() {
  if (!productId.value) {
    storyOptions.value = []
    return
  }
  const res: any = await listStories({ page: 1, pageSize: 100, productId: productId.value })
  storyOptions.value = (res.data.list || []).map((s: any) => ({ label: s.title, value: s.id }))
}

async function loadUsers() {
  const res: any = await listUsers({ page: 1, pageSize: 100, status: 'active' })
  const opts = (res.data.list || []).map((u: any) => ({
    label: `${u.realname || u.account} (${u.account})`,
    value: String(u.id),
  }))
  assigneeOptions.value = [{ label: '指派给我', value: 'me' }, ...opts]
}

function onProductFilterChange(id: number | null) {
  setProductId(id)
  projectId.value = null
  sprintId.value = null
  storyId.value = null
  pagination.page = 1
  loadProjects()
  loadSprints()
  loadStories()
  load()
}

function onProjectChange() {
  sprintId.value = null
  loadSprints()
}

async function load() {
  loading.value = true
  try {
    const res: any = await listBugs({
      page: pagination.page,
      pageSize: pagination.pageSize,
      productId: productId.value || undefined,
      projectId: projectId.value || undefined,
      sprintId: sprintId.value || undefined,
      storyId: storyId.value || undefined,
      status: status.value || undefined,
      severity: severity.value || undefined,
      pri: pri.value || undefined,
      assignedTo: assignedTo.value || undefined,
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
    content: `确定删除缺陷「${row.title}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteBug(row.id)
        message.success('已删除')
        load()
      } catch (e: any) {
        message.error(e.message)
      }
    },
  })
}

onMounted(async () => {
  const q = route.query
  if (q.projectId) projectId.value = Number(q.projectId)
  if (q.sprintId) sprintId.value = Number(q.sprintId)
  if (q.storyId) storyId.value = Number(q.storyId)
  if (q.status) status.value = String(q.status)
  if (q.assignedTo) assignedTo.value = String(q.assignedTo)
  initFromRouteAndStore()
  try {
    await loadUsers()
  } catch {
    // ignore
  }
  await loadProjects()
  await loadSprints()
  if (productId.value) await loadStories()
  load()
})
</script>
