<template>
  <n-form @submit.prevent="onSubmit">
    <n-grid :cols="2" :x-gap="16" :y-gap="0">
      <n-gi>
        <n-form-item label="所属产品" required>
          <n-select
            v-model:value="form.productId"
            :options="productOptions"
            filterable
            placeholder="选择产品"
            :disabled="isEdit"
            @update:value="onProductChange"
          />
        </n-form-item>
      </n-gi>
      <n-gi>
        <n-form-item label="指派人">
          <n-select
            v-model:value="form.assignedTo"
            :options="userOptions"
            clearable
            filterable
            placeholder="可选"
          />
        </n-form-item>
      </n-gi>

      <n-gi :span="2">
        <n-form-item label="标题" required>
          <n-input v-model:value="form.title" />
        </n-form-item>
      </n-gi>

      <n-gi>
        <n-form-item label="严重程度">
          <n-input-number v-model:value="form.severity" :min="1" :max="4" style="width: 100%" />
        </n-form-item>
      </n-gi>
      <n-gi>
        <n-form-item label="优先级">
          <n-input-number v-model:value="form.pri" :min="1" :max="4" style="width: 100%" />
        </n-form-item>
      </n-gi>

      <n-gi>
        <n-form-item label="所属项目">
          <n-select
            v-model:value="form.projectId"
            :options="projectOptions"
            clearable
            filterable
            placeholder="可选"
            @update:value="onProjectChange"
          />
        </n-form-item>
      </n-gi>
      <n-gi>
        <n-form-item label="所属迭代">
          <n-select
            v-model:value="form.sprintId"
            :options="sprintOptions"
            clearable
            filterable
            placeholder="可选"
          />
        </n-form-item>
      </n-gi>

      <n-gi>
        <n-form-item label="关联需求">
          <n-select
            v-model:value="form.storyId"
            :options="storyOptions"
            clearable
            filterable
            placeholder="可选"
          />
        </n-form-item>
      </n-gi>
      <n-gi />

      <n-gi :span="2">
        <n-form-item label="重现步骤">
          <MarkdownEditor
            ref="mdEditorRef"
            v-model="form.steps"
            object-type="bug"
            :object-id="isEdit ? String(route.params.id) : null"
            :can-upload="auth.has('bug.attach')"
            allow-video
            height="300px"
            placeholder="填写重现步骤，可插入图片/mp4…"
          />
        </n-form-item>
      </n-gi>

      <n-gi v-if="isEdit && bugId" :span="2">
        <AttachmentPanel
          object-type="bug"
          :object-id="bugId"
          :attachments="attachments"
          :can-upload="auth.has('bug.attach')"
          :can-delete="auth.has('bug.attach')"
          @refresh="reloadAttachments"
        />
      </n-gi>

      <n-gi v-else-if="auth.has('bug.attach')" :span="2">
        <n-form-item label="附件">
          <n-upload v-model:file-list="pendingFileList" multiple :default-upload="false" :max="20">
            <n-button>选择文件（可多选，含 doc/图片/mp4）</n-button>
          </n-upload>
        </n-form-item>
      </n-gi>
    </n-grid>

    <n-space style="margin-top: 12px">
      <n-button type="primary" attr-type="submit" :loading="loading">保存</n-button>
      <n-button @click="onCancel">取消</n-button>
    </n-space>
  </n-form>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, type UploadFileInfo } from 'naive-ui'
import {
  createBug,
  getBug,
  updateBug,
  listProducts,
  listProjects,
  listSprints,
  listStories,
  listUsers,
  uploadBugAttachment,
} from '@/api'
import { useAuthStore } from '@/stores/auth'
import { useCloseDrawer, useNotifyListReload } from '@/composables/useRouteDrawer'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import AttachmentPanel from '@/components/AttachmentPanel.vue'
import type { AttachmentItem } from '@/components/AttachmentPanel.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const auth = useAuthStore()
const closeDrawer = useCloseDrawer()
const notifyListReload = useNotifyListReload()
const loading = ref(false)
const isEdit = computed(() => !!route.params.id && route.name === 'bug-edit')
const bugId = computed(() => (isEdit.value ? Number(route.params.id) : 0))
const mdEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)
const pendingFileList = ref<UploadFileInfo[]>([])
const attachments = ref<AttachmentItem[]>([])

const productOptions = ref<{ label: string; value: number }[]>([])
const projectOptions = ref<{ label: string; value: number }[]>([])
const sprintOptions = ref<{ label: string; value: number }[]>([])
const storyOptions = ref<{ label: string; value: number }[]>([])
const userOptions = ref<{ label: string; value: number }[]>([])

const form = reactive({
  productId: null as number | null,
  title: '',
  steps: '' as string,
  severity: 3 as number | null,
  pri: 3 as number | null,
  projectId: null as number | null,
  sprintId: null as number | null,
  storyId: null as number | null,
  assignedTo: null as number | null,
})

async function loadProducts() {
  const res: any = await listProducts({ page: 1, pageSize: 100, status: 'normal' })
  productOptions.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
}

async function loadProjects() {
  if (!form.productId) {
    projectOptions.value = []
    return
  }
  const res: any = await listProjects({ page: 1, pageSize: 100, productId: form.productId })
  projectOptions.value = (res.data.list || []).map((p: any) => ({ label: p.name, value: p.id }))
}

async function loadSprints() {
  if (!form.projectId) {
    sprintOptions.value = []
    return
  }
  const res: any = await listSprints({ page: 1, pageSize: 100, projectId: form.projectId })
  sprintOptions.value = (res.data.list || []).map((s: any) => ({ label: s.name, value: s.id }))
}

async function loadStories() {
  if (!form.productId) {
    storyOptions.value = []
    return
  }
  const res: any = await listStories({ page: 1, pageSize: 100, productId: form.productId })
  storyOptions.value = (res.data.list || []).map((s: any) => ({ label: s.title, value: s.id }))
}

async function loadUsers() {
  const res: any = await listUsers({ page: 1, pageSize: 100, status: 'active' })
  userOptions.value = (res.data.list || []).map((u: any) => ({
    label: `${u.realname || u.account} (${u.account})`,
    value: u.id,
  }))
}

function onProductChange() {
  form.projectId = null
  form.sprintId = null
  form.storyId = null
  loadProjects()
  loadSprints()
  loadStories()
}

function onProjectChange() {
  form.sprintId = null
  loadSprints()
}

async function load() {
  if (!isEdit.value) {
    const q = route.query
    if (q.productId) form.productId = Number(q.productId)
    if (q.projectId) form.projectId = Number(q.projectId)
    if (q.sprintId) form.sprintId = Number(q.sprintId)
    if (q.storyId) form.storyId = Number(q.storyId)
    if (form.productId) {
      await loadProjects()
      await loadSprints()
      await loadStories()
    }
    return
  }
  const res: any = await getBug(route.params.id as string)
  const b = res.data
  form.productId = b.productId
  form.title = b.title
  form.steps = b.steps || ''
  form.severity = b.severity
  form.pri = b.pri
  form.projectId = b.projectId
  form.sprintId = b.sprintId
  form.storyId = b.storyId
  form.assignedTo = b.assignedTo
  attachments.value = b.attachments || []
  if (b.productName && !productOptions.value.find((o) => o.value === b.productId)) {
    productOptions.value.push({ label: b.productName, value: b.productId })
  }
  await loadProjects()
  await loadSprints()
  await loadStories()
}

async function reloadAttachments() {
  if (!isEdit.value) return
  const res: any = await getBug(route.params.id as string)
  attachments.value = res.data.attachments || []
}

function onCancel() {
  if (isEdit.value) {
    router.push(`/bugs/${route.params.id}`)
  } else {
    closeDrawer?.()
  }
}

async function onSubmit() {
  if (!form.productId) {
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
        steps: form.steps || null,
        severity: form.severity,
        pri: form.pri,
        projectId: form.projectId || null,
        sprintId: form.sprintId || null,
        storyId: form.storyId || null,
      }
      if (form.assignedTo == null) {
        payload.clearAssign = true
      } else {
        payload.assignedTo = form.assignedTo
      }
      await updateBug(route.params.id as string, payload)
      message.success('已保存')
      notifyListReload()
      router.push(`/bugs/${route.params.id}`)
    } else {
      const res: any = await createBug({
        productId: form.productId,
        title: form.title,
        steps: form.steps || null,
        severity: form.severity,
        pri: form.pri,
        projectId: form.projectId || null,
        sprintId: form.sprintId || null,
        storyId: form.storyId || null,
        assignedTo: form.assignedTo || null,
      })
      const newId = res.data.id
      if (mdEditorRef.value?.hasPending()) {
        const rewritten = await mdEditorRef.value.flushPending(newId, form.steps || '')
        if (rewritten !== (form.steps || '')) {
          await updateBug(newId, { steps: rewritten || null })
        }
      }
      for (const f of pendingFileList.value) {
        if (!f.file) continue
        const fd = new FormData()
        fd.append('file', f.file)
        await uploadBugAttachment(newId, fd)
      }
      message.success('已创建')
      notifyListReload()
      router.push(`/bugs/${newId}`)
    }
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    await loadProducts()
  } catch (e: any) {
    message.error(e.message || '加载产品失败')
  }
  try {
    await loadUsers()
  } catch {
    // ignore
  }
  try {
    await load()
  } catch (e: any) {
    message.error(e.message)
  }
})
</script>
