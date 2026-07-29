<template>
  <n-space vertical v-if="bug">
    <n-space justify="end" wrap>
      <n-button v-if="auth.has('bug.edit')" @click="$router.push(`/bugs/${bug.id}/edit`)">编辑</n-button>
      <n-button
        v-if="auth.has('bug.resolve') && bug.status === 'active'"
        type="primary"
        @click="showResolve = true"
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
        :loading="actionLoading"
        @click="doActivate"
      >
        激活
      </n-button>
      <n-button v-if="auth.has('bug.delete') && bug.status === 'active'" type="error" @click="onDelete">
        删除
      </n-button>
    </n-space>

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
      <n-descriptions-item label="重现步骤" :span="2">{{ bug.steps || '-' }}</n-descriptions-item>
    </n-descriptions>

    <AttachmentPanel
      object-type="bug"
      :object-id="bug.id"
      :attachments="bug.attachments || []"
      :can-upload="auth.has('bug.attach')"
      :can-delete="auth.has('bug.attach')"
      @refresh="load"
    />

    <n-modal v-model:show="showResolve" preset="card" title="解决缺陷" style="width: 420px">
      <n-form>
        <n-form-item label="解决方案" required>
          <n-select v-model:value="resolution" :options="resolutionOptions" placeholder="选择解决方案" />
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
  </n-space>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDialog, useMessage } from 'naive-ui'
import { getBug, resolveBug, closeBug, activateBug, deleteBug } from '@/api'
import AttachmentPanel from '@/components/AttachmentPanel.vue'
import { bugStatusMap, bugResolutionMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'
import { useNotifyListReload, useSyncDrawerTitle } from '@/composables/useRouteDrawer'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const notifyListReload = useNotifyListReload()

const bug = ref<any>(null)
const actionLoading = ref(false)
const showResolve = ref(false)
const resolution = ref<string | null>(null)

useSyncDrawerTitle(() => bug.value?.title, '缺陷详情')

const resolutionOptions = Object.entries(bugResolutionMap).map(([value, label]) => ({ label, value }))

const backTo = computed(() => {
  const pid = bug.value?.productId
  return pid ? `/bugs?productId=${pid}` : '/bugs'
})

async function load() {
  try {
    const res: any = await getBug(route.params.id as string)
    bug.value = res.data
  } catch (e: any) {
    message.error(e.message)
  }
}

async function doResolve() {
  if (!resolution.value) return
  actionLoading.value = true
  try {
    const res: any = await resolveBug(bug.value.id, resolution.value)
    bug.value = res.data
    showResolve.value = false
    resolution.value = null
    message.success('已标记为已解决')
    notifyListReload()
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

async function doActivate() {
  actionLoading.value = true
  try {
    const res: any = await activateBug(bug.value.id)
    bug.value = res.data
    message.success('已激活')
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
