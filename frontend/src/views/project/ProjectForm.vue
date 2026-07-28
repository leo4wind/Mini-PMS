<template>
  <n-card :title="isEdit ? '编辑项目' : '新建项目'" style="max-width: 640px">
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
      <n-form-item label="名称" required>
        <n-input v-model:value="form.name" />
      </n-form-item>
      <n-form-item label="代号">
        <n-input v-model:value="form.code" />
      </n-form-item>
      <n-form-item v-if="isEdit" label="状态">
        <n-select v-model:value="form.status" :options="statusOptions" />
      </n-form-item>
      <n-form-item label="开始日期">
        <n-date-picker v-model:formatted-value="form.begin" value-format="yyyy-MM-dd" type="date" clearable style="width: 100%" />
      </n-form-item>
      <n-form-item label="结束日期">
        <n-date-picker v-model:formatted-value="form.end" value-format="yyyy-MM-dd" type="date" clearable style="width: 100%" />
      </n-form-item>
      <n-form-item label="项目经理">
        <n-select v-model:value="form.pm" :options="userOptions" clearable filterable placeholder="可选" />
      </n-form-item>
      <n-form-item label="描述">
        <n-input v-model:value="form.description" type="textarea" />
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
import { createProject, getProject, updateProject, listProducts, listUsers } from '@/api'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const loading = ref(false)
const isEdit = computed(() => !!route.params.id && route.name === 'project-edit')
const productOptions = ref<{ label: string; value: number }[]>([])
const userOptions = ref<{ label: string; value: number }[]>([])

const statusMap: Record<string, string> = {
  wait: '未开始',
  doing: '进行中',
  suspended: '已挂起',
  closed: '已关闭',
}

const nextStatus: Record<string, string[]> = {
  wait: ['wait', 'doing', 'closed'],
  doing: ['doing', 'suspended', 'closed'],
  suspended: ['suspended', 'doing', 'closed'],
  closed: ['closed'],
}

const form = reactive({
  productId: null as number | null,
  name: '',
  code: '' as string | null,
  begin: null as string | null,
  end: null as string | null,
  pm: null as number | null,
  description: '' as string | null,
  status: 'wait',
  originalStatus: 'wait',
})

const statusOptions = computed(() => {
  const allow = nextStatus[form.originalStatus] || [form.originalStatus]
  return allow.map((v) => ({ label: statusMap[v] || v, value: v }))
})

function sliceDate(v: string | null | undefined) {
  if (!v) return null
  return String(v).slice(0, 10)
}

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
  const res: any = await getProject(route.params.id as string)
  const p = res.data
  form.productId = p.productId
  form.name = p.name
  form.code = p.code
  form.begin = sliceDate(p.begin)
  form.end = sliceDate(p.end)
  form.pm = p.pm
  form.description = p.description
  form.status = p.status
  form.originalStatus = p.status
  // ensure closed product still shows in select when editing
  if (p.productName && !productOptions.value.find((o) => o.value === p.productId)) {
    productOptions.value.push({ label: p.productName, value: p.productId })
  }
}

async function onSubmit() {
  if (!isEdit.value && !form.productId) {
    message.warning('请选择所属产品')
    return
  }
  if (!form.name.trim()) {
    message.warning('请填写名称')
    return
  }
  loading.value = true
  try {
    if (isEdit.value) {
      const payload: Record<string, unknown> = {
        name: form.name,
        code: form.code || null,
        begin: form.begin || '',
        end: form.end || '',
        description: form.description || null,
        status: form.status,
      }
      if (form.pm == null) {
        payload.clearPm = true
      } else {
        payload.pm = form.pm
      }
      await updateProject(route.params.id as string, payload)
      message.success('已保存')
      router.push(`/projects/${route.params.id}`)
    } else {
      const res: any = await createProject({
        productId: form.productId,
        name: form.name,
        code: form.code || null,
        begin: form.begin || null,
        end: form.end || null,
        pm: form.pm || null,
        description: form.description || null,
      })
      message.success('已创建')
      router.push(`/projects/${res.data.id}`)
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
