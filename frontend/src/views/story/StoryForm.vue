<template>
  <n-card :title="isEdit ? '编辑需求' : '新建需求'" style="max-width: 640px">
    <n-form @submit.prevent="onSubmit">
      <n-form-item label="所属产品" required>
        <n-select
          v-model:value="form.productId"
          :options="productOptions"
          filterable
          placeholder="选择产品"
          :disabled="isEdit"
        />
      </n-form-item>
      <n-form-item label="类型" required>
        <n-select v-model:value="form.type" :options="typeOptions" :disabled="isEdit && form.originalType === 'story'" />
      </n-form-item>
      <n-form-item label="标题" required>
        <n-input v-model:value="form.title" />
      </n-form-item>
      <n-form-item label="描述">
        <n-input v-model:value="form.description" type="textarea" />
      </n-form-item>
      <n-form-item label="优先级">
        <n-input-number v-model:value="form.pri" :min="1" :max="4" style="width: 100%" />
      </n-form-item>
      <n-form-item label="估算">
        <n-input-number v-model:value="form.estimate" :min="0" :step="0.5" style="width: 100%" />
      </n-form-item>
      <n-form-item label="指派人">
        <n-select v-model:value="form.assignedTo" :options="userOptions" clearable filterable placeholder="可选" />
      </n-form-item>
      <n-form-item v-if="isEdit" label="状态">
        <n-select v-model:value="form.status" :options="statusOptions" />
      </n-form-item>
      <n-space>
        <n-button type="primary" attr-type="submit" :loading="loading">保存</n-button>
        <n-button @click="$router.back()">取消</n-button>
      </n-space>
    </n-form>
  </n-card>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { createStory, getStory, updateStory, listProducts, listUsers } from '@/api'
import { storyStatusMap } from '@/constants/labels'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const loading = ref(false)
const isEdit = computed(() => !!route.params.id && route.name === 'story-edit')
const productOptions = ref<{ label: string; value: number }[]>([])
const userOptions = ref<{ label: string; value: number }[]>([])

const typeOptions = [
  { label: '规划', value: 'planning' },
  { label: '可交付', value: 'story' },
]

const storyStatusNext: Record<string, string[]> = {
  draft: ['draft', 'active'],
  active: ['active', 'closed'],
  closed: ['closed', 'active'],
}

const form = reactive({
  productId: null as number | null,
  type: 'planning',
  originalType: 'planning',
  title: '',
  description: '' as string | null,
  pri: 3 as number | null,
  estimate: null as number | null,
  assignedTo: null as number | null,
  status: 'draft',
  originalStatus: 'draft',
})

const statusOptions = computed(() => {
  const allow = storyStatusNext[form.originalStatus] || [form.originalStatus]
  return allow.map((v) => ({ label: storyStatusMap[v] || v, value: v }))
})

async function loadOptions() {
  const [prodRes, userRes]: any[] = await Promise.all([
    listProducts({ page: 1, pageSize: 100, status: 'normal' }),
    listUsers({ page: 1, pageSize: 100, status: 'active' }),
  ])
  productOptions.value = (prodRes.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
  userOptions.value = (userRes.data.list || []).map((u: any) => ({
    label: `${u.realname || u.account} (${u.account})`,
    value: u.id,
  }))
}

async function load() {
  if (!isEdit.value) {
    const q = route.query.productId
    if (q) form.productId = Number(q)
    return
  }
  const res: any = await getStory(route.params.id as string)
  const s = res.data
  form.productId = s.productId
  form.type = s.type
  form.originalType = s.type
  form.title = s.title
  form.description = s.description
  form.pri = s.pri
  form.estimate = s.estimate
  form.assignedTo = s.assignedTo
  form.status = s.status
  form.originalStatus = s.status
  if (s.productName && !productOptions.value.find((o) => o.value === s.productId)) {
    productOptions.value.push({ label: s.productName, value: s.productId })
  }
}

async function onSubmit() {
  if (!isEdit.value && !form.productId) {
    message.warning('请选择所属产品')
    return
  }
  if (!form.title.trim()) {
    message.warning('请填写标题')
    return
  }
  loading.value = true
  try {
    if (isEdit.value) {
      const payload: Record<string, unknown> = {
        title: form.title,
        description: form.description || null,
        pri: form.pri,
        estimate: form.estimate,
        type: form.type,
        status: form.status,
      }
      if (form.assignedTo == null) {
        payload.clearAssign = true
      } else {
        payload.assignedTo = form.assignedTo
      }
      await updateStory(route.params.id as string, payload)
      message.success('已保存')
      router.push(`/stories/${route.params.id}`)
    } else {
      const res: any = await createStory({
        productId: form.productId,
        type: form.type,
        title: form.title,
        description: form.description || null,
        pri: form.pri,
        estimate: form.estimate,
        assignedTo: form.assignedTo || null,
      })
      message.success('已创建')
      router.push(`/stories/${res.data.id}`)
    }
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadOptions()
    await load()
  } catch (e: any) {
    message.error(e.message)
  }
})
</script>
