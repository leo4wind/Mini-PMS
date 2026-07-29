<template>
  <n-space vertical v-if="bug">
    <n-space justify="end" wrap>
      <n-button
        v-if="auth.has('bug.resolve') && bug.status === 'active'"
        type="primary"
        @click="openResolve"
      >
        解决
      </n-button>
      <n-button
        v-if="auth.has('bug.close') && bug.status === 'resolved'"
        type="warning"
        :loading="actionLoading"
        @click="doClose"
      >
        关闭
      </n-button>
      <n-button
        v-if="auth.has('bug.edit') && bug.status !== 'active'"
        @click="openActivate"
      >
        激活
      </n-button>
      <n-button
        v-if="auth.has('bug.edit')"
        type="primary"
        secondary
        @click="openRemarkModal"
      >
        添加备注
      </n-button>
      <n-button v-if="auth.has('bug.delete') && bug.status === 'active'" type="error" @click="onDelete">
        删除
      </n-button>
    </n-space>

    <n-alert type="info" :bordered="false">
      缺陷创建后正文已锁定，沟通请追加备注（备注提交后不可修改）。
    </n-alert>

    <n-descriptions bordered :column="2" label-placement="left">
      <n-descriptions-item label="ID">{{ bug.id }}</n-descriptions-item>
      <n-descriptions-item label="状态">{{ bugStatusMap[bug.status] || bug.status }}</n-descriptions-item>
      <n-descriptions-item label="产品">
        <n-button text type="primary" @click="$router.push(`/products/${bug.productId}`)">
          {{ bug.productName || bug.productId }}
        </n-button>
      </n-descriptions-item>
      <n-descriptions-item label="严重程度">{{ bug.severity }}</n-descriptions-item>
      <n-descriptions-item label="优先级">{{ bug.pri }}</n-descriptions-item>
      <n-descriptions-item label="指派人">
        {{ bug.assignee ? bug.assignee.realname || bug.assignee.account : '-' }}
      </n-descriptions-item>
      <n-descriptions-item label="创建人">
        {{ bug.creator ? bug.creator.realname || bug.creator.account : '-' }}
      </n-descriptions-item>
      <n-descriptions-item label="项目">
        <n-button v-if="bug.projectId" text type="primary" @click="$router.push(`/projects/${bug.projectId}`)">
          #{{ bug.projectId }}
        </n-button>
        <span v-else>-</span>
      </n-descriptions-item>
      <n-descriptions-item label="迭代">
        <n-button v-if="bug.sprintId" text type="primary" @click="$router.push(`/sprints/${bug.sprintId}`)">
          #{{ bug.sprintId }}
        </n-button>
        <span v-else>-</span>
      </n-descriptions-item>
      <n-descriptions-item label="关联需求">
        <n-button v-if="bug.storyId" text type="primary" @click="$router.push(`/stories/${bug.storyId}`)">
          #{{ bug.storyId }}
        </n-button>
        <span v-else>-</span>
      </n-descriptions-item>
      <n-descriptions-item label="解决方案">
        {{ bug.resolution ? bugResolutionMap[bug.resolution] || bug.resolution : '-' }}
      </n-descriptions-item>
      <n-descriptions-item v-if="bug.resolveComment" label="解决备注" :span="2">
        {{ bug.resolveComment }}
      </n-descriptions-item>
      <n-descriptions-item label="重现步骤" :span="2">
        <MarkdownView :model-value="bug.steps" />
      </n-descriptions-item>
    </n-descriptions>

    <n-card title="备注" size="small">
      <template #header-extra>
        <n-button v-if="auth.has('bug.edit')" size="small" type="primary" @click="openRemarkModal">
          添加备注
        </n-button>
      </template>
      <n-spin :show="remarksLoading">
        <n-empty v-if="!remarks.length" description="暂无备注（解决/激活说明、沟通记录会出现在这里）" />
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
                <span>{{ r.creator?.realname || r.creator?.account || r.createdBy }}</span>
                <span style="color: var(--n-text-color-3); font-weight: normal">{{ r.createdAt }}</span>
              </n-space>
            </template>
            <MarkdownView :model-value="r.content" />
            <n-space v-if="r.attachments?.length" style="margin-top: 8px" wrap>
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
          </n-card>
        </n-space>
      </n-spin>
    </n-card>

    <AttachmentPanel
      object-type="bug"
      :object-id="bug.id"
      :attachments="bug.attachments || []"
      :can-upload="false"
      :can-delete="false"
      @refresh="load"
    />

    <n-modal v-model:show="showResolve" preset="card" title="解决缺陷" style="width: 520px">
      <n-form>
        <n-form-item label="解决方案" required>
          <n-select v-model:value="resolution" :options="resolutionOptions" placeholder="选择解决方案" />
        </n-form-item>
        <n-form-item label="解决备注">
          <n-input v-model:value="resolveComment" type="textarea" :rows="3" placeholder="可选" />
        </n-form-item>
        <n-form-item label="指派给" feedback="默认创建人，可改派他人">
          <n-select
            v-model:value="resolveAssignedTo"
            :options="userOptions"
            filterable
            clearable
            placeholder="默认创建人"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showResolve = false">取消</n-button>
          <n-button type="primary" :loading="actionLoading" :disabled="!resolution" @click="doResolve">
            确认
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showActivate"
      preset="card"
      title="激活缺陷"
      style="width: 720px"
      :mask-closable="false"
      @after-leave="resetActivateForm"
    >
      <n-form>
        <n-form-item label="激活说明" required>
          <MarkdownEditor
            ref="activateEditorRef"
            v-model="activateContent"
            object-type="bug"
            :can-upload="auth.has('bug.attach')"
            height="260px"
            placeholder="说明为何重新激活，可粘贴截图…"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showActivate = false">取消</n-button>
          <n-button type="primary" :loading="actionLoading" @click="doActivate">确认激活</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="remarkModalShow"
      preset="card"
      title="添加备注"
      style="width: 720px"
      :mask-closable="false"
      @after-leave="resetRemarkForm"
    >
      <n-form>
        <n-form-item label="内容">
          <MarkdownEditor
            ref="remarkEditorRef"
            v-model="remarkContent"
            object-type="bug"
            :can-upload="auth.has('bug.attach')"
            allow-video
            height="260px"
            placeholder="填写备注，可插入图片/视频…"
          />
        </n-form-item>
        <n-form-item v-if="auth.has('bug.attach')" label="附件">
          <n-upload v-model:file-list="remarkFileList" multiple :default-upload="false" :max="20">
            <n-button>选择文件</n-button>
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
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDialog, useMessage, type UploadFileInfo } from 'naive-ui'
import {
  getBug,
  resolveBug,
  closeBug,
  activateBug,
  deleteBug,
  listBugRemarks,
  createBugRemarkDraft,
  finalizeBugRemark,
  discardBugRemarkDraft,
  uploadBugRemarkAttachment,
  listUsers,
} from '@/api'
import AttachmentPanel from '@/components/AttachmentPanel.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import MarkdownView from '@/components/MarkdownView.vue'
import { bugStatusMap, bugResolutionMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'
import { useNotifyListReload, useSyncDrawerTitle } from '@/composables/useRouteDrawer'
import { downloadAttachment } from '@/utils/attachment'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const notifyListReload = useNotifyListReload()

const bug = ref<any>(null)
const actionLoading = ref(false)
const showResolve = ref(false)
const showActivate = ref(false)
const activateContent = ref('')
const activateEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)
const resolution = ref<string | null>(null)
const resolveComment = ref('')
const resolveAssignedTo = ref<number | null>(null)
const userOptions = ref<{ label: string; value: number }[]>([])

const remarks = ref<any[]>([])
const remarksLoading = ref(false)
const remarkModalShow = ref(false)
const remarkContent = ref('')
const remarkFileList = ref<UploadFileInfo[]>([])
const remarkSaving = ref(false)
const remarkEditorRef = ref<InstanceType<typeof MarkdownEditor> | null>(null)

useSyncDrawerTitle(() => bug.value?.title, '缺陷详情')

const resolutionOptions = Object.entries(bugResolutionMap).map(([value, label]) => ({ label, value }))

const backTo = computed(() => {
  const pid = bug.value?.productId
  return pid ? `/bugs?productId=${pid}` : '/bugs'
})

async function loadUsers() {
  try {
    const res: any = await listUsers({ page: 1, pageSize: 100, status: 'active' })
    userOptions.value = (res.data.list || []).map((u: any) => ({
      label: `${u.realname || u.account} (${u.account})`,
      value: u.id,
    }))
  } catch {
    userOptions.value = []
  }
}

async function load() {
  try {
    const res: any = await getBug(route.params.id as string)
    bug.value = res.data
    await loadRemarks()
  } catch (e: any) {
    message.error(e.message)
  }
}

async function loadRemarks() {
  remarksLoading.value = true
  try {
    const res: any = await listBugRemarks(bug.value.id)
    remarks.value = res.data || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    remarksLoading.value = false
  }
}

function openResolve() {
  resolution.value = null
  resolveComment.value = ''
  resolveAssignedTo.value = bug.value?.openedBy ?? null
  showResolve.value = true
  if (!userOptions.value.length) void loadUsers()
}

async function doResolve() {
  if (!resolution.value) return
  actionLoading.value = true
  try {
    const res: any = await resolveBug(bug.value.id, {
      resolution: resolution.value,
      resolveComment: resolveComment.value || null,
      assignedTo: resolveAssignedTo.value || bug.value.openedBy,
    })
    bug.value = res.data
    showResolve.value = false
    message.success('已标记为已解决')
    notifyListReload()
    await loadRemarks()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    actionLoading.value = false
  }
}

async function doClose() {
  actionLoading.value = true
  try {
    const res: any = await closeBug(bug.value.id)
    bug.value = res.data
    message.success('已关闭')
    notifyListReload()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    actionLoading.value = false
  }
}

function openActivate() {
  activateContent.value = ''
  showActivate.value = true
}

function resetActivateForm() {
  activateContent.value = ''
}

/** 将编辑器内容定稿为一条备注；失败时清理草稿并抛错 */
async function postRemarkFromEditor(
  editorRef: InstanceType<typeof MarkdownEditor> | null,
  htmlRaw: string,
  files: UploadFileInfo[] = [],
) {
  const html = (htmlRaw || '').trim()
  const hasFiles = files.some((f) => f.file)
  const hasPending = editorRef?.hasPending()
  if (!html && !hasFiles && !hasPending) {
    throw new Error('请填写内容或添加截图')
  }
  let draftId: number | string | null = null
  try {
    const draftRes: any = await createBugRemarkDraft(bug.value.id)
    draftId = draftRes.data.id
    let content = html
    if (editorRef?.hasPending()) {
      content = await editorRef.flushPending(bug.value.id, content || '', draftId)
    }
    for (const f of files) {
      if (!f.file) continue
      const fd = new FormData()
      fd.append('file', f.file)
      await uploadBugRemarkAttachment(bug.value.id, draftId!, fd)
    }
    await finalizeBugRemark(bug.value.id, draftId!, { content: content || '' })
  } catch (e) {
    if (draftId != null) {
      try {
        await discardBugRemarkDraft(bug.value.id, draftId)
      } catch {
        // ignore
      }
    }
    throw e
  }
}

async function doActivate() {
  const html = (activateContent.value || '').trim()
  const hasPending = activateEditorRef.value?.hasPending()
  if (!html && !hasPending) {
    message.warning('请填写激活说明或粘贴截图')
    return
  }
  actionLoading.value = true
  try {
    await postRemarkFromEditor(activateEditorRef.value, activateContent.value)
    const res: any = await activateBug(bug.value.id)
    bug.value = res.data
    showActivate.value = false
    message.success('已激活')
    notifyListReload()
    await loadRemarks()
  } catch (e: any) {
    message.error(e.message || '激活失败')
  } finally {
    actionLoading.value = false
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
  remarkSaving.value = true
  try {
    await postRemarkFromEditor(remarkEditorRef.value, remarkContent.value, remarkFileList.value)
    message.success('备注已添加')
    remarkModalShow.value = false
    await loadRemarks()
  } catch (e: any) {
    message.error(e.message || '提交失败')
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

function onDelete() {
  dialog.warning({
    title: '确认删除',
    content: `确定删除缺陷「${bug.value.title}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteBug(bug.value.id)
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
