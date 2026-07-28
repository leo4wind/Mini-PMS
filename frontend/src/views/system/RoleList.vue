<template>
  <n-space vertical>
    <n-h2 style="margin: 0">角色管理</n-h2>
    <n-data-table :columns="columns" :data="list" :loading="loading" />

    <n-modal v-model:show="showEdit" preset="card" title="编辑角色" style="width: 420px">
      <n-form>
        <n-form-item label="编码">
          <n-input :value="form.code" disabled />
        </n-form-item>
        <n-form-item label="名称">
          <n-input v-model:value="form.name" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="form.remark" type="textarea" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showEdit = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="saveEdit">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import * as api from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const message = useMessage()
const list = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const showEdit = ref(false)
const form = reactive({ id: 0, code: '', name: '', remark: '' })

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '编码', key: 'code' },
  { title: '名称', key: 'name' },
  {
    title: '内置',
    key: 'builtin',
    width: 80,
    render(row) {
      return h(NTag, { type: row.builtin ? 'warning' : 'default', size: 'small' }, { default: () => (row.builtin ? '是' : '否') })
    },
  },
  { title: '备注', key: 'remark' },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          auth.has('role.edit')
            ? h(NButton, { text: true, type: 'primary', onClick: () => openEdit(row) }, { default: () => '编辑' })
            : null,
          auth.has('role.assignMenu')
            ? h(NButton, { text: true, onClick: () => router.push(`/system/roles/${row.id}/menus`) }, { default: () => '分配菜单' })
            : null,
        ],
      })
    },
  },
]

async function load() {
  loading.value = true
  try {
    const res: any = await api.listRoles()
    list.value = res.data || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

function openEdit(row: any) {
  form.id = row.id
  form.code = row.code
  form.name = row.name
  form.remark = row.remark || ''
  showEdit.value = true
}

async function saveEdit() {
  saving.value = true
  try {
    await api.updateRole(form.id, { name: form.name, remark: form.remark })
    message.success('已保存')
    showEdit.value = false
    load()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
