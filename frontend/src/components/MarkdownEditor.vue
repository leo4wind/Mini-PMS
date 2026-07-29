<template>
  <div class="desc-editor" :style="{ minHeight: height }">
    <div v-if="editor" class="desc-toolbar">
      <button type="button" :class="{ active: editor.isActive('bold') }" @click="editor.chain().focus().toggleBold().run()">
        B
      </button>
      <button type="button" :class="{ active: editor.isActive('italic') }" @click="editor.chain().focus().toggleItalic().run()">
        I
      </button>
      <button type="button" :class="{ active: editor.isActive('strike') }" @click="editor.chain().focus().toggleStrike().run()">
        S
      </button>
      <button type="button" @click="editor.chain().focus().toggleBulletList().run()">• 列表</button>
      <button type="button" @click="editor.chain().focus().toggleOrderedList().run()">1. 列表</button>
      <button type="button" :disabled="uploading" @click="pickImage">{{ uploading ? '上传中…' : '图片' }}</button>
    </div>
    <editor-content :editor="editor" class="desc-content" />
    <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onFilePick" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'
import { useMessage } from 'naive-ui'
import { uploadStoryAttachment, uploadStoryRemarkAttachment } from '@/api'
import {
  attachmentPreviewPath,
  htmlForEditorDisplay,
  htmlForStorage,
  withAuthPreviewUrl,
} from '@/utils/markdownAttachment'

const props = withDefaults(
  defineProps<{
    modelValue?: string | null
    storyId?: number | string | null
    /** 备注图片上传目标；有值时优先于 story 级上传 */
    remarkId?: number | string | null
    canUpload?: boolean
    height?: string
    maxImageMB?: number
    placeholder?: string
  }>(),
  {
    modelValue: '',
    storyId: null,
    remarkId: null,
    canUpload: false,
    height: '300px',
    maxImageMB: 5,
    placeholder: '填写内容，可直接粘贴图片（将上传到服务器）…',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const message = useMessage()
const fileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
/** create 时 blob: → File，保存后再上传 */
const pendingByBlob = new Map<string, File>()
let syncing = false

function emitStorageHtml() {
  if (!editor.value) return
  const html = editor.value.isEmpty ? '' : htmlForStorage(editor.value.getHTML())
  emit('update:modelValue', html)
}

async function insertImageFiles(files: FileList | File[]) {
  if (!props.canUpload) {
    message.warning('无上传附件权限，无法插入图片')
    return
  }
  const list = Array.from(files)
  const maxBytes = props.maxImageMB * 1024 * 1024
  uploading.value = true
  try {
    for (const file of list) {
      if (!file.type.startsWith('image/')) {
        message.warning(`跳过非图片：${file.name}`)
        continue
      }
      if (file.size > maxBytes) {
        message.warning(`图片过大（>${props.maxImageMB}MB）：${file.name}`)
        continue
      }

      if (props.remarkId != null && props.remarkId !== '' && props.storyId != null && props.storyId !== '') {
        try {
          const formData = new FormData()
          formData.append('file', file)
          const res: any = await uploadStoryRemarkAttachment(props.storyId, props.remarkId, formData)
          const stored = attachmentPreviewPath(res.data.id)
          const display = withAuthPreviewUrl(stored)
          editor.value?.chain().focus().setImage({ src: display }).run()
        } catch (e: any) {
          message.error(e.message || '图片上传失败')
        }
      } else if (props.storyId != null && props.storyId !== '' && (props.remarkId == null || props.remarkId === '')) {
        try {
          const formData = new FormData()
          formData.append('file', file)
          const res: any = await uploadStoryAttachment(props.storyId, formData)
          const stored = attachmentPreviewPath(res.data.id)
          const display = withAuthPreviewUrl(stored)
          editor.value?.chain().focus().setImage({ src: display }).run()
        } catch (e: any) {
          message.error(e.message || '图片上传失败')
        }
      } else {
        const blobUrl = URL.createObjectURL(file)
        pendingByBlob.set(blobUrl, file)
        editor.value?.chain().focus().setImage({ src: blobUrl }).run()
      }
    }
  } finally {
    uploading.value = false
  }
}

const editor = useEditor({
  extensions: [
    StarterKit,
    Image.configure({
      allowBase64: false,
      inline: false,
    }),
    Placeholder.configure({
      placeholder: props.placeholder,
    }),
  ],
  content: htmlForEditorDisplay(props.modelValue || ''),
  editorProps: {
    attributes: {
      class: 'desc-prose',
    },
    handlePaste(_view, event) {
      const items = event.clipboardData?.items
      if (!items) return false
      const files: File[] = []
      for (const item of items) {
        if (item.type.startsWith('image/')) {
          const f = item.getAsFile()
          if (f) files.push(f)
        }
      }
      if (!files.length) return false
      event.preventDefault()
      void insertImageFiles(files)
      return true
    },
    handleDrop(_view, event) {
      const files = event.dataTransfer?.files
      if (!files?.length) return false
      const images = Array.from(files).filter((f) => f.type.startsWith('image/'))
      if (!images.length) return false
      event.preventDefault()
      void insertImageFiles(images)
      return true
    },
  },
  onUpdate: () => {
    if (syncing) return
    emitStorageHtml()
  },
})

watch(
  () => props.modelValue,
  (v) => {
    if (!editor.value) return
    const nextStored = v || ''
    const curStored = editor.value.isEmpty ? '' : htmlForStorage(editor.value.getHTML())
    if (nextStored === curStored) return
    syncing = true
    editor.value.commands.setContent(htmlForEditorDisplay(nextStored), { emitUpdate: false })
    syncing = false
  },
)

function pickImage() {
  fileInput.value?.click()
}

function onFilePick(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.length) void insertImageFiles(input.files)
  input.value = ''
}

/** 把 blob: 换成服务端预览地址；remarkId 有值则挂到备注 */
async function flushPending(
  storyId: number | string,
  html: string,
  remarkId?: number | string | null,
): Promise<string> {
  if (pendingByBlob.size === 0) return htmlForStorage(html)
  let result = html
  for (const [blobUrl, file] of pendingByBlob) {
    if (!result.includes(blobUrl)) continue
    const formData = new FormData()
    formData.append('file', file)
    const res: any =
      remarkId != null && remarkId !== ''
        ? await uploadStoryRemarkAttachment(storyId, remarkId, formData)
        : await uploadStoryAttachment(storyId, formData)
    const stored = attachmentPreviewPath(res.data.id)
    result = result.split(blobUrl).join(stored)
    URL.revokeObjectURL(blobUrl)
  }
  pendingByBlob.clear()
  return htmlForStorage(result)
}

function hasPending() {
  return pendingByBlob.size > 0
}

onBeforeUnmount(() => {
  for (const url of pendingByBlob.keys()) URL.revokeObjectURL(url)
  pendingByBlob.clear()
})

defineExpose({ flushPending, hasPending })
</script>

<style scoped>
.desc-editor {
  width: 100%;
  border: 1px solid var(--n-border-color, #e0e0e6);
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
}
.desc-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--n-border-color, #e0e0e6);
  background: #fafafa;
}
.desc-toolbar button {
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 4px;
  padding: 2px 8px;
  cursor: pointer;
  font-size: 12px;
  line-height: 1.6;
}
.desc-toolbar button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.desc-toolbar button.active {
  background: #e8f3ff;
  border-color: #2080f0;
  color: #2080f0;
}
.desc-content {
  padding: 8px 12px;
  min-height: 220px;
  max-height: 480px;
  overflow: auto;
}
.desc-content :deep(.desc-prose) {
  outline: none;
  min-height: 200px;
}
.desc-content :deep(.desc-prose p) {
  margin: 0.4em 0;
}
.desc-content :deep(.desc-prose img) {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 8px 0;
  border-radius: 4px;
}
.desc-content :deep(.is-editor-empty:first-child::before) {
  color: #aaa;
  content: attr(data-placeholder);
  float: left;
  height: 0;
  pointer-events: none;
}
</style>
