<template>
  <n-space vertical v-if="project">
    <n-space justify="end" wrap>
      <n-button v-if="auth.has('project.edit')" @click="$router.push(`/projects/${project.id}/edit`)">编辑</n-button>
      <n-button @click="$router.push(`/sprints?projectId=${project.id}`)">迭代</n-button>
      <n-button @click="$router.push(`/bugs?projectId=${project.id}`)">缺陷</n-button>
    </n-space>

    <n-tabs type="line" v-model:value="tab">
      <n-tab-pane name="overview" tab="概览">
        <n-descriptions bordered :column="2" label-placement="left">
          <n-descriptions-item label="ID">{{ project.id }}</n-descriptions-item>
          <n-descriptions-item label="代号">{{ project.code || '-' }}</n-descriptions-item>
          <n-descriptions-item label="所属产品">
            <n-button text type="primary" @click="$router.push(`/products/${project.productId}`)">
              {{ project.productName || project.productId }}
            </n-button>
          </n-descriptions-item>
          <n-descriptions-item label="状态">{{ statusMap[project.status] || project.status }}</n-descriptions-item>
          <n-descriptions-item label="PM">
            {{ project.pmUser ? project.pmUser.realname || project.pmUser.account : '-' }}
          </n-descriptions-item>
          <n-descriptions-item label="迭代数">{{ project.sprintCount }}</n-descriptions-item>
          <n-descriptions-item label="开始">{{ fmtDate(project.begin) }}</n-descriptions-item>
          <n-descriptions-item label="结束">{{ fmtDate(project.end) }}</n-descriptions-item>
          <n-descriptions-item label="描述" :span="2">{{ project.description || '-' }}</n-descriptions-item>
        </n-descriptions>

        <n-space v-if="auth.has('project.edit') && nextActions.length" style="margin-top: 16px">
          <n-button
            v-for="a in nextActions"
            :key="a.value"
            :type="a.value === 'closed' ? 'warning' : 'primary'"
            :loading="statusLoading"
            @click="changeStatus(a.value)"
          >
            {{ a.label }}
          </n-button>
        </n-space>
      </n-tab-pane>

      <n-tab-pane name="sprints" tab="迭代">
        <n-space justify="end" style="margin-bottom: 12px">
          <n-button
            v-if="auth.has('sprint.create')"
            size="small"
            type="primary"
            @click="$router.push(`/sprints/new?projectId=${project.id}`)"
          >
            新建迭代
          </n-button>
        </n-space>
        <n-data-table :columns="sprintColumns" :data="sprints" :loading="sprintsLoading" :bordered="false" size="small" />
      </n-tab-pane>

      <n-tab-pane name="stories" tab="需求">
        <n-data-table :columns="storyColumns" :data="stories" :loading="storiesLoading" :bordered="false" size="small" />
      </n-tab-pane>

      <n-tab-pane name="bugs" tab="缺陷">
        <n-space justify="space-between" style="margin-bottom: 12px">
          <n-text depth="3">本项目下的缺陷</n-text>
          <n-space>
            <n-button size="small" @click="$router.push(`/bugs?projectId=${project.id}`)">查看全部</n-button>
            <n-button
              v-if="auth.has('bug.create')"
              size="small"
              type="primary"
              @click="$router.push(`/bugs/new?productId=${project.productId}&projectId=${project.id}`)"
            >
              新建缺陷
            </n-button>
          </n-space>
        </n-space>
        <n-data-table :columns="bugColumns" :data="bugs" :loading="bugsLoading" :bordered="false" size="small" />
      </n-tab-pane>
    </n-tabs>
  </n-space>
</template>

<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getProject, updateProject, listProjectSprints, listProjectStories, listBugs } from '@/api'
import { storyTypeMap, storyStatusMap, sprintStatusMap, bugStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'
import { useNotifyListReload, useSyncDrawerTitle } from '@/composables/useRouteDrawer'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const notifyListReload = useNotifyListReload()
const project = ref<any>(null)
const tab = ref('overview')
const statusLoading = ref(false)
const sprints = ref<any[]>([])
const sprintsLoading = ref(false)
const stories = ref<any[]>([])
const storiesLoading = ref(false)
const bugs = ref<any[]>([])
const bugsLoading = ref(false)

useSyncDrawerTitle(() => project.value?.name, '项目详情')

const statusMap: Record<string, string> = {
  wait: '未开始',
  doing: '进行中',
  suspended: '已挂起',
  closed: '已关闭',
}

const actionLabel: Record<string, string> = {
  doing: '开始',
  suspended: '挂起',
  closed: '关闭',
}

const statusNext: Record<string, string[]> = {
  wait: ['doing', 'closed'],
  doing: ['suspended', 'closed'],
  suspended: ['doing', 'closed'],
  closed: [],
}

const nextActions = computed(() => {
  if (!project.value) return []
  return (statusNext[project.value.status] || []).map((v) => ({
    value: v,
    label: actionLabel[v] || v,
  }))
})

function fmtDate(v: string | null | undefined) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

const sprintColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name' },
  { title: '状态', key: 'status', width: 90, render: (r) => sprintStatusMap[r.status] || r.status },
  { title: '开始', key: 'begin', width: 110, render: (r) => fmtDate(r.begin) },
  { title: '结束', key: 'end', width: 110, render: (r) => fmtDate(r.end) },
  { title: '需求数', key: 'storyCount', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) =>
      h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/sprints/${row.id}`) }, { default: () => '查看' }),
  },
]

const storyColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '标题', key: 'title', ellipsis: { tooltip: true } },
  { title: '类型', key: 'type', width: 90, render: (r) => storyTypeMap[r.type] || r.type },
  { title: '状态', key: 'status', width: 80, render: (r) => storyStatusMap[r.status] || r.status },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) =>
      h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/stories/${row.id}`) }, { default: () => '查看' }),
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

async function load() {
  try {
    const res: any = await getProject(route.params.id as string)
    project.value = res.data
    sprints.value = []
    stories.value = []
    bugs.value = []
    tab.value = 'overview'
  } catch (e: any) {
    message.error(e.message)
  }
}

async function loadSprints() {
  sprintsLoading.value = true
  try {
    const res: any = await listProjectSprints(route.params.id as string)
    sprints.value = res.data || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    sprintsLoading.value = false
  }
}

async function loadStories() {
  storiesLoading.value = true
  try {
    const res: any = await listProjectStories(route.params.id as string)
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
    const res: any = await listBugs({ page: 1, pageSize: 20, projectId: route.params.id })
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
    const res: any = await updateProject(project.value.id, { status })
    project.value = res.data
    message.success('状态已更新')
    notifyListReload()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    statusLoading.value = false
  }
}

watch(tab, (name) => {
  if (name === 'sprints' && !sprints.value.length) loadSprints()
  if (name === 'stories' && !stories.value.length) loadStories()
  if (name === 'bugs' && !bugs.value.length) loadBugs()
})

watch(() => route.params.id, load, { immediate: true })
</script>
