<template>
  <n-form @submit.prevent="onSubmit">
    <n-form-item label="所属项目" required>
      <n-select
        v-model:value="form.projectId"
        :options="projectOptions"
        filterable
        placeholder="选择项目"
        :disabled="isEdit"
      />
    </n-form-item>
    <n-form-item label="名称" required>
      <n-input v-model:value="form.name" />
    </n-form-item>
    <n-form-item label="开始日期">
      <n-date-picker v-model:formatted-value="form.begin" value-format="yyyy-MM-dd" type="date" clearable style="width: 100%" />
    </n-form-item>
    <n-form-item label="结束日期">
      <n-date-picker v-model:formatted-value="form.end" value-format="yyyy-MM-dd" type="date" clearable style="width: 100%" />
    </n-form-item>
    <n-form-item label="目标">
      <n-input v-model:value="form.goal" type="textarea" />
    </n-form-item>
    <n-form-item v-if="isEdit" label="状态">
      <n-select v-model:value="form.status" :options="statusOptions" />
    </n-form-item>
    <n-space>
      <n-button type="primary" attr-type="submit" :loading="loading">保存</n-button>
      <n-button @click="onCancel">取消</n-button>
    </n-space>
  </n-form>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { createSprint, getSprint, updateSprint, listProjects } from '@/api'
import { sprintStatusMap } from '@/constants/labels'
import { useCloseDrawer } from '@/composables/useRouteDrawer'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const closeDrawer = useCloseDrawer()
const loading = ref(false)
const isEdit = computed(() => !!route.params.id && route.name === 'sprint-edit')
const projectOptions = ref<{ label: string; value: number }[]>([])

const sprintStatusNext: Record<string, string[]> = {
  wait: ['wait', 'doing', 'closed'],
  doing: ['doing', 'done', 'closed'],
  done: ['done', 'doing', 'closed'],
  closed: ['closed'],
}

const form = reactive({
  projectId: null as number | null,
  name: '',
  begin: null as string | null,
  end: null as string | null,
  goal: '' as string | null,
  status: 'wait',
  originalStatus: 'wait',
})

const statusOptions = computed(() => {
  const allow = sprintStatusNext[form.originalStatus] || [form.originalStatus]
  return allow.map((v) => ({ label: sprintStatusMap[v] || v, value: v }))
})

function sliceDate(v: string | null | undefined) {
  if (!v) return null
  return String(v).slice(0, 10)
}

async function loadProjects() {
  const res: any = await listProjects({ page: 1, pageSize: 100 })
  projectOptions.value = (res.data.list || []).map((p: any) => ({
    label: `${p.name}${p.productName ? ` (${p.productName})` : ''}`,
    value: p.id,
  }))
}

async function load() {
  if (!isEdit.value) {
    const q = route.query.projectId
    if (q) form.projectId = Number(q)
    return
  }
  const res: any = await getSprint(route.params.id as string)
  const s = res.data
  form.projectId = s.projectId
  form.name = s.name
  form.begin = sliceDate(s.begin)
  form.end = sliceDate(s.end)
  form.goal = s.goal
  form.status = s.status
  form.originalStatus = s.status
  if (s.projectName && !projectOptions.value.find((o) => o.value === s.projectId)) {
    projectOptions.value.push({ label: s.projectName, value: s.projectId })
  }
}

function onCancel() {
  if (isEdit.value) {
    router.push(`/sprints/${route.params.id}`)
  } else {
    closeDrawer?.()
  }
}

async function onSubmit() {
  if (!isEdit.value && !form.projectId) {
    message.warning('请选择所属项目')
    return
  }
  if (!form.name.trim()) {
    message.warning('请填写名称')
    return
  }
  loading.value = true
  try {
    if (isEdit.value) {
      await updateSprint(route.params.id as string, {
        name: form.name,
        begin: form.begin || '',
        end: form.end || '',
        goal: form.goal || null,
        status: form.status,
      })
      message.success('已保存')
      router.push(`/sprints/${route.params.id}`)
    } else {
      const res: any = await createSprint({
        projectId: form.projectId,
        name: form.name,
        begin: form.begin || null,
        end: form.end || null,
        goal: form.goal || null,
      })
      message.success('已创建')
      router.push(`/sprints/${res.data.id}`)
    }
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadProjects()
    await load()
  } catch (e: any) {
    message.error(e.message)
  }
})
</script>
