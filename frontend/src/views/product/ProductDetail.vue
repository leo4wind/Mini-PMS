<template>
  <n-spin :show="loading">
    <n-space vertical v-if="product">
      <n-space justify="end">
        <n-button v-if="auth.has('product.edit')" @click="$router.push(`/products/${product.id}/edit`)">编辑</n-button>
        <n-button @click="$router.push(`/stories?productId=${product.id}`)">需求</n-button>
        <n-button @click="$router.push(`/bugs?productId=${product.id}`)">缺陷</n-button>
        <n-button @click="$router.push(`/projects?productId=${product.id}`)">项目</n-button>
      </n-space>
      <n-card title="基本信息" size="small">
        <n-descriptions :column="2" label-placement="left">
          <n-descriptions-item label="ID">{{ product.id }}</n-descriptions-item>
          <n-descriptions-item label="代号">{{ product.code || '-' }}</n-descriptions-item>
          <n-descriptions-item label="状态">{{ product.status }}</n-descriptions-item>
          <n-descriptions-item label="创建时间">{{ product.createdAt }}</n-descriptions-item>
          <n-descriptions-item label="描述" :span="2">{{ product.description || '-' }}</n-descriptions-item>
        </n-descriptions>
      </n-card>
      <n-card title="所属项目" size="small">
        <template #header-extra>
          <n-button
            v-if="auth.has('project.create') && product.status === 'normal'"
            size="small"
            type="primary"
            @click="$router.push(`/projects/new?productId=${product.id}`)"
          >
            新建项目
          </n-button>
        </template>
        <n-data-table :columns="projectColumns" :data="projects" :bordered="false" size="small" />
      </n-card>
      <n-alert type="info" title="说明">
        新建产品时会自动创建「{{ product.name }}1.0」项目。
      </n-alert>
    </n-space>
  </n-spin>
</template>

<script setup lang="ts">
import { h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getProduct, listProductProjects } from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useSyncDrawerTitle } from '@/composables/useRouteDrawer'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const auth = useAuthStore()
const product = ref<any>(null)
const projects = ref<any[]>([])
const loading = ref(false)

useSyncDrawerTitle(() => product.value?.name, '产品详情')

const statusMap: Record<string, string> = {
  wait: '未开始',
  doing: '进行中',
  suspended: '已挂起',
  closed: '已关闭',
}

const projectColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name' },
  { title: '状态', key: 'status', width: 90, render: (r) => statusMap[r.status] || r.status },
  { title: '迭代数', key: 'sprintCount', width: 80 },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) =>
      h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/projects/${row.id}`) }, { default: () => '查看' }),
  },
]

async function load() {
  loading.value = true
  try {
    const id = route.params.id as string
    const [pRes, pjRes]: any[] = await Promise.all([getProduct(id), listProductProjects(id)])
    product.value = pRes.data
    projects.value = pjRes.data || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

watch(() => route.params.id, load, { immediate: true })
</script>
