<template>
  <n-space vertical v-if="sprint">
    <n-page-header :title="sprint.name" @back="$router.push(backTo)">
      <template #extra>
        <n-space>
          <n-button v-if="auth.has('sprint.edit')" @click="$router.push(`/sprints/${sprint.id}/edit`)">编辑</n-button>
          <n-button
            v-for="a in statusActions"
            :key="a.value"
            :type="a.value === 'closed' ? 'warning' : 'primary'"
            :loading="statusLoading"
            @click="changeStatus(a.value)"
          >
            {{ a.label }}
          </n-button>
          <n-button v-if="auth.has('sprint.delete')" type="error" @click="onDelete">删除</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-descriptions bordered :column="2" label-placement="left">
      <n-descriptions-item label="ID">{{ sprint.id }}</n-descriptions-item>
      <n-descriptions-item label="项目">
        <n-button text type="primary" @click="$router.push(`/projects/${sprint.projectId}`)">
          {{ sprint.projectName || sprint.projectId }}
        </n-button>
      </n-descriptions-item>
      <n-descriptions-item label="产品">
        <n-button v-if="sprint.productId" text type="primary" @click="$router.push(`/products/${sprint.productId}`)">
          {{ sprint.productName || sprint.productId }}
        </n-button>
        <span v-else>-</span>
      </n-descriptions-item>
      <n-descriptions-item label="状态">{{ sprintStatusMap[sprint.status] || sprint.status }}</n-descriptions-item>
      <n-descriptions-item label="开始">{{ fmtDate(sprint.begin) }}</n-descriptions-item>
      <n-descriptions-item label="结束">{{ fmtDate(sprint.end) }}</n-descriptions-item>
      <n-descriptions-item label="需求数">{{ sprint.storyCount }}</n-descriptions-item>
      <n-descriptions-item label="目标" :span="2">{{ sprint.goal || '-' }}</n-descriptions-item>
    </n-descriptions>

    <n-tabs type="line" v-model:value="tab">
      <n-tab-pane name="stories" tab="迭代需求">
        <n-space vertical>
          <n-space justify="space-between">
            <n-text depth="3" v-if="sprint.status !== 'doing'">
              仅进行中的迭代可关联/移除需求
            </n-text>
            <n-button
              v-if="auth.has('sprint.linkStory')"
              type="primary"
              :disabled="sprint.status !== 'doing'"
              @click="openLinkModal"
            >
              关联需求
            </n-button>
          </n-space>
          <n-data-table :columns="storyColumns" :data="stories" :loading="storiesLoading" :bordered="false" size="small" />
        </n-space>
      </n-tab-pane>

      <n-tab-pane name="bugs" tab="相关缺陷">
        <n-space justify="space-between" style="margin-bottom: 12px">
          <n-text depth="3">本迭代下的缺陷</n-text>
          <n-button
            v-if="auth.has('bug.create')"
            size="small"
            type="primary"
            @click="goNewBug"
          >
            新建缺陷
          </n-button>
        </n-space>
        <n-data-table :columns="bugColumns" :data="bugs" :loading="bugsLoading" :bordered="false" size="small" />
      </n-tab-pane>
    </n-tabs>

    <n-modal v-model:show="linkModal" preset="card" title="关联需求" style="width: 640px">
      <n-space vertical>
        <n-alert type="info" :bordered="false">
          仅列出本产品下「类型=可交付」且「状态=激活」的需求；规划需求请先在需求详情转为可交付并激活。
        </n-alert>
        <n-space>
          <n-input v-model:value="candidateKeyword" placeholder="搜索标题" clearable style="width: 220px" />
          <n-button @click="loadCandidates">查询</n-button>
        </n-space>
        <n-empty
          v-if="!candidatesLoading && !candidates.length"
          description="没有可关联的需求（检查类型/状态/所属产品）"
        />
        <n-data-table
          v-else
          :columns="candidateColumns"
          :data="candidates"
          :loading="candidatesLoading"
          :row-key="(r: any) => r.id"
          :checked-row-keys="selectedStoryIds"
          @update:checked-row-keys="(keys: Array<string | number>) => (selectedStoryIds = keys as number[])"
        />
        <n-space justify="end">
          <n-button @click="linkModal = false">取消</n-button>
          <n-button type="primary" :loading="linkLoading" :disabled="!selectedStoryIds.length" @click="confirmLink">
            确认关联
          </n-button>
        </n-space>
      </n-space>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  getSprint,
  updateSprint,
  deleteSprint,
  listSprintStories,
  listSprintStoryCandidates,
  linkSprintStories,
  unlinkSprintStory,
  listBugs,
} from '@/api'
import { storyTypeMap, storyStatusMap, sprintStatusMap, bugStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

const sprint = ref<any>(null)
const tab = ref('stories')
const statusLoading = ref(false)
const stories = ref<any[]>([])
const storiesLoading = ref(false)
const bugs = ref<any[]>([])
const bugsLoading = ref(false)

const linkModal = ref(false)
const candidates = ref<any[]>([])
const candidatesLoading = ref(false)
const candidateKeyword = ref('')
const selectedStoryIds = ref<number[]>([])
const linkLoading = ref(false)

const actionLabel: Record<string, string> = {
  doing: '开始',
  done: '完成',
  closed: '关闭',
}

const statusNext: Record<string, string[]> = {
  wait: ['doing', 'closed'],
  doing: ['done', 'closed'],
  done: ['doing', 'closed'],
  closed: [],
}

const statusActions = computed(() => {
  if (!sprint.value || !auth.has('sprint.edit')) return []
  return (statusNext[sprint.value.status] || []).map((v) => ({
    value: v,
    label: actionLabel[v] || v,
  }))
})

const backTo = computed(() => {
  const pid = sprint.value?.projectId
  return pid ? `/sprints?projectId=${pid}` : '/sprints'
})

function fmtDate(v: string | null | undefined) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

function renderUser(u: any) {
  if (!u) return '-'
  return u.realname || u.account || '-'
}

const storyColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '类型', key: 'type', width: 90, render: (r) => storyTypeMap[r.type] || r.type },
  { title: '状态', key: 'status', width: 80, render: (r) => storyStatusMap[r.status] || r.status },
  { title: '指派人', key: 'assignee', width: 100, render: (r) => renderUser(r.assignee) },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/stories/${row.id}`) }, { default: () => '查看' }),
          auth.has('sprint.linkStory') && sprint.value?.status === 'doing'
            ? h(NButton, { text: true, type: 'error', onClick: () => onUnlink(row) }, { default: () => '移除' })
            : null,
        ],
      })
    },
  },
]

const bugColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '状态', key: 'status', width: 90, render: (r) => bugStatusMap[r.status] || r.status },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) =>
      h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/bugs/${row.id}`) }, { default: () => '查看' }),
  },
]

const candidateColumns: DataTableColumns<any> = [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '类型', key: 'type', width: 90, render: (r) => storyTypeMap[r.type] || r.type },
  { title: '状态', key: 'status', width: 80, render: (r) => storyStatusMap[r.status] || r.status },
]

async function load() {
  try {
    const res: any = await getSprint(route.params.id as string)
    sprint.value = res.data
    await Promise.all([loadStories(), loadBugs()])
  } catch (e: any) {
    message.error(e.message)
  }
}

async function loadStories() {
  storiesLoading.value = true
  try {
    const res: any = await listSprintStories(route.params.id as string)
    stories.value = res.data || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    storiesLoading.value = false
  }
}

async function loadBugs() {
  bugsLoading.value = true
  try {
    const res: any = await listBugs({ page: 1, pageSize: 50, sprintId: route.params.id })
    bugs.value = res.data.list || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    bugsLoading.value = false
  }
}

async function changeStatus(status: string) {
  statusLoading.value = true
  try {
    const res: any = await updateSprint(sprint.value.id, { status })
    sprint.value = res.data
    message.success('状态已更新')
  } catch (e: any) {
    message.error(e.message)
  } finally {
    statusLoading.value = false
  }
}

function onDelete() {
  dialog.warning({
    title: '确认删除',
    content: `确定删除迭代「${sprint.value.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteSprint(sprint.value.id)
        message.success('已删除')
        router.push(backTo.value)
      } catch (e: any) {
        message.error(e.message)
      }
    },
  })
}

function goNewBug() {
  const s = sprint.value
  const q = new URLSearchParams()
  if (s.productId) q.set('productId', String(s.productId))
  q.set('projectId', String(s.projectId))
  q.set('sprintId', String(s.id))
  router.push(`/bugs/new?${q.toString()}`)
}

function openLinkModal() {
  selectedStoryIds.value = []
  candidateKeyword.value = ''
  linkModal.value = true
  loadCandidates()
}

async function loadCandidates() {
  candidatesLoading.value = true
  try {
    const res: any = await listSprintStoryCandidates(route.params.id as string, {
      page: 1,
      pageSize: 50,
      keyword: candidateKeyword.value || undefined,
    })
    candidates.value = res.data.list || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    candidatesLoading.value = false
  }
}

async function confirmLink() {
  linkLoading.value = true
  try {
    await linkSprintStories(route.params.id as string, selectedStoryIds.value)
    message.success('已关联')
    linkModal.value = false
    await loadStories()
    const res: any = await getSprint(route.params.id as string)
    sprint.value = res.data
  } catch (e: any) {
    message.error(e.message)
  } finally {
    linkLoading.value = false
  }
}

function onUnlink(row: any) {
  dialog.warning({
    title: '确认移除',
    content: `确定从迭代中移除需求「${row.title}」吗？`,
    positiveText: '移除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await unlinkSprintStory(route.params.id as string, row.id)
        message.success('已移除')
        await loadStories()
        const res: any = await getSprint(route.params.id as string)
        sprint.value = res.data
      } catch (e: any) {
        message.error(e.message)
      }
    },
  })
}

onMounted(load)
</script>
