<template>
  <n-form @submit.prevent="onSubmit">
    <n-form-item label="名称" required>
      <n-input v-model:value="form.name" />
    </n-form-item>
    <n-form-item label="代号">
      <n-input v-model:value="form.code" />
    </n-form-item>
    <n-form-item label="描述">
      <n-input v-model:value="form.description" type="textarea" />
    </n-form-item>
    <n-space>
      <n-button type="primary" attr-type="submit" :loading="loading">保存</n-button>
      <n-button @click="onCancel">取消</n-button>
    </n-space>
  </n-form>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { createProduct, getProduct, updateProduct } from '@/api'
import { useCloseDrawer, useNotifyListReload } from '@/composables/useRouteDrawer'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const closeDrawer = useCloseDrawer()
const notifyListReload = useNotifyListReload()
const loading = ref(false)
const isEdit = computed(() => !!route.params.id && route.name === 'product-edit')

const form = reactive({
  name: '',
  code: '' as string | null,
  description: '' as string | null,
})

async function load() {
  if (!isEdit.value) return
  const res: any = await getProduct(route.params.id as string)
  form.name = res.data.name
  form.code = res.data.code
  form.description = res.data.description
}

function onCancel() {
  if (isEdit.value) {
    router.push(`/products/${route.params.id}`)
  } else {
    closeDrawer?.()
  }
}

async function onSubmit() {
  if (!form.name.trim()) {
    message.warning('请填写名称')
    return
  }
  loading.value = true
  try {
    const payload = {
      name: form.name,
      code: form.code || null,
      description: form.description || null,
    }
    if (isEdit.value) {
      await updateProduct(route.params.id as string, payload)
      message.success('已保存')
      notifyListReload()
      router.push(`/products/${route.params.id}`)
    } else {
      const res: any = await createProduct(payload)
      const p = res.data.product
      const proj = res.data.defaultProject
      message.success(`已创建产品，并自动生成项目「${proj.name}」`)
      notifyListReload()
      router.push(`/products/${p.id}`)
    }
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
