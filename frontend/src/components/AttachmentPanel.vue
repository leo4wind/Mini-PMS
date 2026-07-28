<template>
  <n-card title="附件" size="small">
    <template #header-extra>
      <n-upload
        v-if="canUpload"
        :show-file-list="false"
        :custom-request="handleUpload"
      >
        <n-button size="small" type="primary" :loading="uploading">上传附件</n-button>
      </n-upload>
    </template>
    <n-empty v-if="!attachments.length" description="暂无附件" />
    <n-data-table
      v-else
      :columns="columns"
      :data="attachments"
      :bordered="false"
      size="small"
    />
    <n-modal v-model:show="previewVisible" preset="card" title="图片预览" style="width: auto; max-width: 90vw">
      <img v-if="previewUrl" :src="previewUrl" style="max-width: 100%; max-height: 70vh" alt="preview" />
    </n-modal>
  </n-card>
</template>

<script setup lang="ts">
import { h, onUnmounted, ref, watch } from 'vue'
import type { UploadCustomRequestOptions } from 'naive-ui'
import { NButton, NSpace, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  deleteAttachment,
  uploadStoryAttachment,
  uploadBugAttachment,
} from '@/api'
import { formatBytes, isImageExt } from '@/constants/labels'
import { downloadAttachment, previewAttachmentUrl } from '@/utils/attachment'

export type AttachmentItem = {
  id: number
  originalName: string
  ext: string
  sizeBytes: number
  uploadedBy?: number
  uploaderName?: string
  createdAt: string
}

const props = defineProps<{
  objectType: 'story' | 'bug'
  objectId: number | string
  attachments: AttachmentItem[]
  canUpload?: boolean
  canDelete?: boolean
}>()

const emit = defineEmits<{ refresh: [] }>()

const message = useMessage()
const uploading = ref(false)
const previewVisible = ref(false)
const previewUrl = ref('')

watch(previewVisible, (v) => {
  if (!v && previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
})

onUnmounted(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
})

const columns: DataTableColumns<AttachmentItem> = [
  { title: '文件名', key: 'originalName', ellipsis: { tooltip: true } },
  { title: '大小', key: 'sizeBytes', width: 90, render: (r) => formatBytes(r.sizeBytes) },
  {
    title: '上传人',
    key: 'uploadedBy',
    width: 100,
    render: (r) => r.uploaderName || (r.uploadedBy ? `#${r.uploadedBy}` : '-'),
  },
  { title: '时间', key: 'createdAt', width: 160 },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          isImageExt(row.ext)
            ? h(NButton, { text: true, type: 'primary', onClick: () => onPreview(row) }, { default: () => '预览' })
            : null,
          h(NButton, { text: true, onClick: () => onDownload(row) }, { default: () => '下载' }),
          props.canDelete
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

async function handleUpload({ file, onFinish, onError }: UploadCustomRequestOptions) {
  if (!file.file) return
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file.file)
    if (props.objectType === 'story') {
      await uploadStoryAttachment(props.objectId, formData)
    } else {
      await uploadBugAttachment(props.objectId, formData)
    }
    message.success('上传成功')
    emit('refresh')
    onFinish()
  } catch (e: any) {
    message.error(e.message)
    onError()
  } finally {
    uploading.value = false
  }
}

async function onPreview(row: AttachmentItem) {
  try {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = await previewAttachmentUrl(row.id)
    previewVisible.value = true
  } catch (e: any) {
    message.error(e.message)
  }
}

async function onDownload(row: AttachmentItem) {
  try {
    await downloadAttachment(row.id, row.originalName)
  } catch (e: any) {
    message.error(e.message)
  }
}

async function onDelete(row: AttachmentItem) {
  try {
    await deleteAttachment(row.id)
    message.success('已删除')
    emit('refresh')
  } catch (e: any) {
    message.error(e.message)
  }
}
</script>
