<template>
  <n-space vertical v-if="story">
    <n-space justify="end" wrap>
      <n-button v-if="auth.has('story.edit')" @click="$router.push(`/stories/${story.id}/edit`)">编辑</n-button>
      <n-button
        v-if="auth.has('story.edit') && story.status === 'draft'"
        type="primary"
        :loading="actionLoading"
        @click="doStatus('active')"
      >
        激活
      </n-button>
      <n-button
        v-if="auth.has('story.edit') && story.status === 'active'"
        type="warning"
        :loading="actionLoading"
        @click="doStatus('closed')"
      >
        关闭
      </n-button>
      <n-button
        v-if="auth.has('story.edit') && story.status === 'closed'"
        type="primary"
        :loading="actionLoading"
        @click="doStatus('active')"
      >
        重开
      </n-button>
      <n-button
        v-if="auth.has('story.edit') && story.type === 'planning'"
        :loading="actionLoading"
        @click="convertToStory"
      >
        转为可交付
      </n-button>
      <n-button v-if="auth.has('story.delete')" type="error" @click="onDelete">删除</n-button>
    </n-space>

    <n-descriptions bordered :column="2" label-placement="left">
      <n-descriptions-item label="ID">{{ story.id }}</n-descriptions-item>
      <n-descriptions-item label="产品">
        <n-button text type="primary" @click="$router.push(`/products/${story.productId}`)">
          {{ story.productName || story.productId }}
        </n-button>
      </n-descriptions-item>
      <n-descriptions-item label="类型">{{ storyTypeMap[story.type] || story.type }}</n-descriptions-item>
      <n-descriptions-item label="状态">{{ storyStatusMap[story.status] || story.status }}</n-descriptions-item>
      <n-descriptions-item label="优先级">{{ story.pri }}</n-descriptions-item>
      <n-descriptions-item label="估算">{{ story.estimate ?? '-' }}</n-descriptions-item>
      <n-descriptions-item label="指派人">
        {{ story.assignee ? story.assignee.realname || story.assignee.account : '-' }}
      </n-descriptions-item>
      <n-descriptions-item label="更新时间">{{ story.updatedAt }}</n-descriptions-item>
      <n-descriptions-item label="描述" :span="2">{{ story.description || '-' }}</n-descriptions-item>
    </n-descriptions>

    <AttachmentPanel
      object-type="story"
      :object-id="story.id"
      :attachments="story.attachments || []"
      :can-upload="auth.has('story.attach')"
      :can-delete="auth.has('story.attach')"
      @refresh="load"
    />

    <n-card title="关联迭代" size="small">
      <n-empty v-if="!story.sprints?.length" description="尚未被任何迭代拉入" />
      <n-data-table v-else :columns="sprintColumns" :data="story.sprints" :bordered="false" size="small" />
    </n-card>

    <n-space>
      <n-button @click="$router.push(`/bugs?productId=${story.productId}&storyId=${story.id}`)">相关缺陷</n-button>
    </n-space>
  </n-space>
</template>

<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { getStory, updateStory, deleteStory } from '@/api'
import AttachmentPanel from '@/components/AttachmentPanel.vue'
import { storyTypeMap, storyStatusMap, sprintStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'
import { useNotifyListReload, useSyncDrawerTitle } from '@/composables/useRouteDrawer'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const notifyListReload = useNotifyListReload()

const story = ref<any>(null)
const actionLoading = ref(false)

useSyncDrawerTitle(() => story.value?.title, '需求详情')

const backTo = computed(() => {
  const pid = story.value?.productId
  return pid ? `/stories?productId=${pid}` : '/stories'
})

const sprintColumns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '名称', key: 'name' },
  { title: '状态', key: 'status', width: 100, render: (r) => sprintStatusMap[r.status] || r.status },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row) =>
      h(NButton, { text: true, type: 'primary', onClick: () => router.push(`/sprints/${row.id}`) }, { default: () => '查看' }),
  },
]

async function load() {
  try {
    const res: any = await getStory(route.params.id as string)
    story.value = res.data
  } catch (e: any) {
    message.error(e.message)
  }
}

async function doStatus(status: string) {
  actionLoading.value = true
  try {
    const res: any = await updateStory(story.value.id, { status })
    story.value = res.data
    message.success('状态已更新')
    notifyListReload()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    actionLoading.value = false
  }
}

async function convertToStory() {
  actionLoading.value = true
  try {
    const res: any = await updateStory(story.value.id, { type: 'story' })
    story.value = res.data
    message.success('已转为可交付需求')
    notifyListReload()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    actionLoading.value = false
  }
}

function onDelete() {
  dialog.warning({
    title: '确认删除',
    content: `确定删除需求「${story.value.title}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteStory(story.value.id)
        message.success('已删除')
        notifyListReload()
        router.push(backTo.value)
      } catch (e: any) {
        message.error(e.message)
      }
    },
  })
}

watch(() => route.params.id, load, { immediate: true })
</script>
