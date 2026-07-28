<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">迭代列表</n-h2>
      <n-button v-if="auth.has('sprint.create')" type="primary" @click="goCreate">新建迭代</n-button>
    </n-space>
    <n-space>
      <n-select
        v-model:value="productId"
        :options="productOptions"
        clearable
        filterable
        placeholder="所属产品"
        style="width: 200px"
        @update:value="onProductChange"
      />
      <n-select
        v-model:value="projectId"
        :options="projectOptions"
        clearable
        filterable
        placeholder="所属项目"
        style="width: 200px"
      />
      <n-select v-model:value="status" :options="statusOptions" clearable placeholder="状态" style="width: 140px" />
      <n-button @click="load">查询</n-button>
    </n-space>
    <n-data-table :columns="columns" :data="list" :loading="loading" :pagination="pagination" remote @update:page="onPage" />
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listSprints, deleteSprint, listProducts, listProjects } from '@/api'
import { sprintStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const list = ref<any[]>([])
const loading = ref(false)
const productId = ref<number | null>(null)
const projectId = ref<number | null>(null)
const status = ref<string | null>(null)
const productOptions = ref<{ label: string; value: number }[]>([])
const projectOptions = ref<{ label: string; value: number }[]>([])
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const statusOptions = Object.entries(sprintStatusMap).map(([value, label]) => ({ label, value }))

function fmtDate(v: string | null | undefined) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name' },
  { title: '项目', key: 'projectName', width: 140 },
  { title: '产品', key: 'productName', width: 120 },
  { title: '状态', key: 'status', width: 90, render: (r) => sprintStatusMap[r.status] || r.status },
  { title: '开始', key: 'begin', width: 110, render: (r) => fmtDate(r.begin) },
  { title: '结束', key: 'end', width: 110, render: (r) => fmtDate(r.end) },
  { title: '需求数', key: 'storyCount', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/sprints/${row.id}`) }, { default: () => '查看' }),
          auth.has('sprint.edit')
            ? h(NButton, { text: true, onClick: () => router.push(`/sprints/${row.id}/edit`) }, { default: () => '编辑' })
            : null,
          auth.has('sprint.delete')
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

function goCreate() {
  const q = projectId.value ? `?projectId=${projectId.value}` : ''
  router.push(`/sprints/new${q}`)
}

async function loadProducts() {
  try {
    const res: any = await listProducts({ page: 1, pageSize: 100, status: 'normal' })
    productOptions.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
  } catch {
    /* ignore */
  }
}

async function loadProjects() {
  try {
    const res: any = await listProjects({
      page: 1,
      pageSize: 100,
      productId: productId.value || undefined,
    })
    projectOptions.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
    if (projectId.value && !projectOptions.value.find((o) => o.value === projectId.value)) {
      projectId.value = null
    }
  } catch {
    /* ignore */
  }
}

function onProductChange() {
  projectId.value = null
  loadProjects()
}

async function load() {
  loading.value = true
  try {
    const res: any = await listSprints({
      page: pagination.page,
      pageSize: pagination.pageSize,
      productId: productId.value || undefined,
      projectId: projectId.value || undefined,
      status: status.value || undefined,
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

function onDelete(row: any) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除迭代「${row.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteSprint(row.id)
        message.success('已删除')
        load()
      } catch (e: any) {
        message.error(e.message)
      }
    },
  })
}

watch(projectId, () => {
  pagination.page = 1
})

onMounted(async () => {
  const q = route.query
  if (q.productId) productId.value = Number(q.productId)
  if (q.projectId) projectId.value = Number(q.projectId)
  if (q.status) status.value = String(q.status)
  await loadProducts()
  await loadProjects()
  load()
})
</script>
