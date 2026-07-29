<template>
  <n-select
    :value="modelValue"
    :options="options"
    :loading="loading"
    clearable
    filterable
    placeholder="产品"
    style="width: 180px"
    @update:value="onUpdate"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { listProducts } from '@/api'

defineProps<{
  modelValue: number | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const message = useMessage()
const loading = ref(false)
const options = ref<{ label: string; value: number }[]>([])

function onUpdate(v: number | null) {
  emit('update:modelValue', v ?? null)
}

async function loadOptions() {
  loading.value = true
  try {
    const res: any = await listProducts({ page: 1, pageSize: 200 })
    options.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
  } catch (e: any) {
    message.error(e.message || '加载产品失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadOptions)
</script>
