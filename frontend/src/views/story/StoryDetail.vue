<template>
  <n-space vertical v-if="story">
    <n-space justify="end" wrap>
      <n-button
        v-if="auth.has('story.edit') && story.type === 'planning'"
        @click="$router.push(`/stories/${story.id}/edit`)"
      >
        编辑
      </n-button>
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
      <n-button
        v-if="auth.has('story.edit') && story.type === 'story'"
        type="primary"
        @click="openRemarkModal"
      >
        添加备注
      </n-button>
      <n-button v-if="auth.has('story.delete')" type="error" @click="onDelete">删除</n-button>
    </n-space>

    <n-alert v-if="story.type === 'story'" type="info" :bordered="false">
      可交付需求正文已锁定，沟通请追加备注（备注提交后不可修改）。
    </n-alert>

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
      <n-descriptions-item label="描述" :span="2">
        <MarkdownView :model-value="story.description" />
      </n-descriptions-item>
    </n-descriptions>

    <AttachmentPanel
      object-type="story"
      :object-id="story.id"
      :attachments="story.attachments || []"
      :can-upload="auth.has('story.attach') && story.type === 'planning'"
      :can-delete="auth.has('story.attach') && story.type === 'planning'"
      @refresh="load"
    />

    <n-card v-if="story.type === 'story'" title="备注" size="small">
      <template #header-extra>
        <n-button
          v-if="auth.has('story.edit')"
          size="small"
          type="primary"
          @click="openRemarkModal"
        >
          添加备注
        </n-button>
      </template>
      <n-spin :show="remarksLoading">
        <n-empty v-if="!remarks.length" description="暂无备注" />
        <n-space v-else vertical size="large">
          <n-card
            v-for="r in remarks"
            :key="r.id"
            size="small"
            :bordered="true"
            embedded
          >
            <template #header>
              <n-space justify="space-between" style="width: 100%">
                <span>{{ r.creator?.realname || r.creator?.account || `#${r.createdBy}` }}</span>
                <span style="color: #999; font-weight: normal">{{ r.createdAt }}</span>
              </n-space>
            </template>
            <MarkdownView :model-value="r.content" />
            <div v-if="r.attachments?.length" style="margin-top: 8px">
              <n-text depth="3">附件：</n-text>
              <n-space>
                <n-button
                  v-for="a in r.attachments"
                  :key="a.id"
                  text
                  type="primary"
                  @click="downloadRemarkFile(a)"
                >
                  {{ a.originalName }}
                </n-button>
              </n-space>
            </div>
          </n-card>
        </n-space>
      </n-spin>
    </n-card>

    <n-card title="关联迭代" size="small">
      <n-empty v-if="!story.sprints?.length" description="尚未被任何迭代拉入" />
      <n-data-table v-else :columns="sprintColumns" :data="story.sprints" :bordered="false" size="small" />
    </n-card>

    <n-space>
      <n-button @click="$router.push(`/bugs?productId=${story.productId}&storyId=${story.id}`)">相关缺陷</n-button>
    </n-space>

    <n-modal
      v-model:show="remarkModalShow"
      preset="card"
      title="添加备注"
      :style="{ width: 'min(720px, 92vw)' }"
      :mask-closable="false"
      @after-leave="resetRemarkForm"
    >
      <n-form>
        <n-form-item label="内容">
          <MarkdownEditor
            ref="remarkEditorRef"
            v-model="remarkContent"
            :can-upload="auth.has('story.attach')"
            height="260px"
            placeholder="填写备注，可粘贴图片…"
          />
        </n-form-item>
        <n-form-item v-if="auth.has('story.attach')" label="附件">
          <n-upload v-model:file-list="remarkFileList" multiple :default-upload="false" :max="20">
            <n-button>选择文件（可多选）</n-button>
          </n-upload>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="remarkModalShow = false">取消</n-button>
          <n-button type="primary" :loading="remarkSaving" @click="submitRemark">提交</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, useDialog, useMessage, type UploadFileInfo } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  getStory,
  updateStory,
  deleteStory,
  listStoryRemarks,
  createStoryRemarkDraft,
  finalizeStoryRemark,
  discardStoryRemarkDraft,
  uploadStoryRemarkAttachment,
} from '@/api'
import AttachmentPanel from '@/components/AttachmentPanel.vue'
import MarkdownView from '@/components/MarkdownView.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import { storyTypeMap, storyStatusMap, sprintStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'
import { useNotifyListReload, useSyncDrawerTitle } from '@/composables/useRouteDrawer'
import { downloadAttachment } from '@/utils/attachment'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const notifyListReload = useNotifyListReload()

const story = ref<any>(null)
const actionLoading = ref(false)
const remarks = ref<any[]>([])
const remarksLoading = ref(false)
const remarkModalShow = ref(false)
const remarkContent = ref('')
const remarkFileList = ref<UploadFileInfo[]>([])
const remarkSaving = ref(false)
const remarkEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)

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
    if (story.value?.type === 'story') await loadRemarks()
    else remarks.value = []
  } catch (e: any) {
    message.error(e.message)
  }
}

async function loadRemarks() {
  remarksLoading.value = true
  try {
    const res: any = await listStoryRemarks(story.value.id)
    remarks.value = res.data || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    remarksLoading.value = false
  }
}

function openRemarkModal() {
  resetRemarkForm()
  remarkModalShow.value = true
}

function resetRemarkForm() {
  remarkContent.value = ''
  remarkFileList.value = []
}

async function submitRemark() {
  const html = (remarkContent.value || '').trim()
  const hasFiles = remarkFileList.value.some((f) => f.file)
  const hasPendingImg = remarkEditorRef.value?.hasPending()
  if (!html && !hasFiles && !hasPendingImg) {
    message.warning('请填写备注内容或添加附件')
    return
  }
  remarkSaving.value = true
  let draftId: number | string | null = null
  try {
    const draftRes: any = await createStoryRemarkDraft(story.value.id)
    draftId = draftRes.data.id
    let content = html
    if (remarkEditorRef.value?.hasPending()) {
      content = await remarkEditorRef.value.flushPending(story.value.id, content || '', draftId)
    }
    for (const f of remarkFileList.value) {
      if (!f.file) continue
      const fd = new FormData()
      fd.append('file', f.file)
      await uploadStoryRemarkAttachment(story.value.id, draftId!, fd)
    }
    await finalizeStoryRemark(story.value.id, draftId!, { content: content || '' })
    message.success('备注已添加')
    remarkModalShow.value = false
    await loadRemarks()
  } catch (e: any) {
    message.error(e.message)
    if (draftId != null) {
      try {
        await discardStoryRemarkDraft(story.value.id, draftId)
      } catch {
        // ignore cleanup errors
      }
    }
  } finally {
    remarkSaving.value = false
  }
}

async function downloadRemarkFile(a: { id: number; originalName: string }) {
  try {
    await downloadAttachment(a.id, a.originalName)
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
    await loadRemarks()
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
