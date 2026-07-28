<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">需求列表</n-h2>
      <n-button v-if="auth.has('story.create')" type="primary" @click="goCreate">新建需求</n-button>
    </n-space>

    <n-space>
      <n-select
        v-model:value="productId"
        :options="productOptions"
        clearable
        filterable
        placeholder="所属产品"
        style="width: 200px"
      />
      <n-select v-model:value="type" :options="typeOptions" clearable placeholder="类型" style="width: 120px" />
      <n-select v-model:value="status" :options="statusOptions" clearable placeholder="状态" style="width: 120px" />
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

    <n-data-table
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      remote
      @update:page="onPage"
    />
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { listStories, deleteStory, listProducts, listUsers } from '@/api'
import { storyTypeMap, storyStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const list = ref<any[]>([])
const loading = ref(false)
const productId = ref<number | null>(null)
const type = ref<string | null>(null)
const status = ref<string | null>(null)
const assignedTo = ref<string | null>(null)
const keyword = ref('')
const productOptions = ref<{ label: string; value: number }[]>([])
const assigneeOptions = ref<{ label: string; value: string }[]>([])
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })

const typeOptions = Object.entries(storyTypeMap).map(([value, label]) => ({ label, value }))
const statusOptions = Object.entries(storyStatusMap).map(([value, label]) => ({ label, value }))

function renderUser(u: any) {
  if (!u) return '-'
  return u.realname || u.account || '-'
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
  { title: '附件', key: 'attachCount', width: 70, render: (r) => r.attachCount ?? 0 },
  { title: '更新时间', key: 'updatedAt', width: 170 },
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

async function loadProducts() {
  try {
    const res: any = await listProducts({ page: 1, pageSize: 100, status: 'normal' })
    productOptions.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
  } catch {
    /* ignore */
  }
}

async function loadUsers() {
  try {
    const res: any = await listUsers({ page: 1, pageSize: 100, status: 'active' })
    const opts = (res.data.list || []).map((u: any) => ({
      label: `${u.realname || u.account} (${u.account})`,
      value: String(u.id),
    }))
    assigneeOptions.value = [{ label: '指派给我', value: 'me' }, ...opts]
  } catch {
    /* ignore */
  }
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

onMounted(async () => {
  const q = route.query
  if (q.productId) productId.value = Number(q.productId)
  if (q.assignedTo) assignedTo.value = String(q.assignedTo)
  if (q.type) type.value = String(q.type)
  if (q.status) status.value = String(q.status)
  await Promise.all([loadProducts(), loadUsers()])
  load()
})
</script>
