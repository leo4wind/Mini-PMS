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
      <button v-if="allowVideo" type="button" :disabled="uploading" @click="pickVideo">视频</button>
    </div>
    <editor-content :editor="editor" class="desc-content" />
    <input ref="imageInput" type="file" accept="image/*" multiple hidden @change="onImagePick" />
    <input ref="videoInput" type="file" accept="video/mp4,.mp4" hidden @change="onVideoPick" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { Node, mergeAttributes } from '@tiptap/core'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'
import { useMessage } from 'naive-ui'
import {
  uploadStoryAttachment,
  uploadStoryRemarkAttachment,
  uploadBugAttachment,
  uploadBugRemarkAttachment,
} from '@/api'
import {
  attachmentPreviewPath,
  htmlForEditorDisplay,
  htmlForStorage,
  withAuthPreviewUrl,
} from '@/utils/markdownAttachment'

export type EditorObjectType = 'story' | 'bug'

const Video = Node.create({
  name: 'video',
  group: 'block',
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      src: { default: null },
      controls: { default: true },
    }
  },
  parseHTML() {
    return [{ tag: 'video' }]
  },
  renderHTML({ HTMLAttributes }) {
    return [
      'video',
      mergeAttributes(HTMLAttributes, {
        controls: 'true',
        style: 'max-width:100%;display:block;margin:8px 0;',
      }),
    ]
  },
})

const props = withDefaults(
  defineProps<{
    modelValue?: string | null
    /** story / bug 主体；与 remarkId 组合决定上传目标 */
    objectType?: EditorObjectType
    objectId?: number | string | null
    /** @deprecated 用 objectType=story + objectId */
    storyId?: number | string | null
    remarkId?: number | string | null
    canUpload?: boolean
    allowVideo?: boolean
    height?: string
    maxImageMB?: number
    maxVideoMB?: number
    placeholder?: string
  }>(),
  {
    modelValue: '',
    objectType: 'story',
    objectId: null,
    storyId: null,
    remarkId: null,
    canUpload: false,
    allowVideo: false,
    height: '300px',
    maxImageMB: 5,
    maxVideoMB: 100,
    placeholder: '填写内容，可直接粘贴图片（将上传到服务器）…',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const message = useMessage()
const imageInput = ref<HTMLInputElement | null>(null)
const videoInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const pendingByBlob = new Map<string, File>()
let syncing = false

function resolvedObjectId() {
  if (props.objectId != null && props.objectId !== '') return props.objectId
  if (props.storyId != null && props.storyId !== '') return props.storyId
  return null
}

async function uploadFile(file: File, objectId: number | string, remarkId?: number | string | null) {
  const formData = new FormData()
  formData.append('file', file)
  const type = props.objectType
  if (remarkId != null && remarkId !== '') {
    if (type === 'bug') {
      return uploadBugRemarkAttachment(objectId, remarkId, formData)
    }
    return uploadStoryRemarkAttachment(objectId, remarkId, formData)
  }
  if (type === 'bug') {
    return uploadBugAttachment(objectId, formData)
  }
  return uploadStoryAttachment(objectId, formData)
}

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
  const oid = resolvedObjectId()
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
      if (oid != null) {
        try {
          const res: any = await uploadFile(file, oid, props.remarkId)
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

async function insertVideoFile(file: File) {
  if (!props.canUpload) {
    message.warning('无上传附件权限，无法插入视频')
    return
  }
  if (!file.name.toLowerCase().endsWith('.mp4') && file.type !== 'video/mp4') {
    message.warning('仅支持 mp4 视频')
    return
  }
  const maxBytes = props.maxVideoMB * 1024 * 1024
  if (file.size > maxBytes) {
    message.warning(`视频过大（>${props.maxVideoMB}MB）`)
    return
  }
  const oid = resolvedObjectId()
  uploading.value = true
  try {
    if (oid != null) {
      const res: any = await uploadFile(file, oid, props.remarkId)
      const stored = attachmentPreviewPath(res.data.id)
      const display = withAuthPreviewUrl(stored)
      editor.value?.chain().focus().insertContent({ type: 'video', attrs: { src: display, controls: true } }).run()
    } else {
      const blobUrl = URL.createObjectURL(file)
      pendingByBlob.set(blobUrl, file)
      editor.value?.chain().focus().insertContent({ type: 'video', attrs: { src: blobUrl, controls: true } }).run()
    }
  } catch (e: any) {
    message.error(e.message || '视频上传失败')
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
    Video,
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
  imageInput.value?.click()
}
function pickVideo() {
  videoInput.value?.click()
}

function onImagePick(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.length) void insertImageFiles(input.files)
  input.value = ''
}
function onVideoPick(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.[0]) void insertVideoFile(input.files[0])
  input.value = ''
}

/** 把 blob: 换成服务端预览地址 */
async function flushPending(
  objectId: number | string,
  html: string,
  remarkId?: number | string | null,
): Promise<string> {
  if (pendingByBlob.size === 0) return htmlForStorage(html)
  let result = html
  for (const [blobUrl, file] of pendingByBlob) {
    if (!result.includes(blobUrl)) continue
    const res: any = await uploadFile(file, objectId, remarkId)
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
.desc-content :deep(.desc-prose img),
.desc-content :deep(.desc-prose video) {
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
