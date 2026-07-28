<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">菜单管理</n-h2>
      <n-button v-if="auth.has('menu.create')" type="primary" @click="openCreate(null)">新建顶级菜单</n-button>
    </n-space>
    <n-data-table
      :columns="columns"
      :data="tableData"
      :loading="loading"
      :row-key="(row: any) => row.id"
      default-expand-all
    />

    <n-modal v-model:show="showForm" preset="card" :title="editing ? '编辑菜单' : '新建菜单'" style="width: 520px">
      <n-form label-placement="left" label-width="80">
        <n-form-item label="父级">
          <n-tree-select
            v-model:value="form.parentId"
            :options="parentOptions"
            clearable
            key-field="key"
            label-field="label"
            children-field="children"
            placeholder="空=顶级"
          />
        </n-form-item>
        <n-form-item label="编码" required>
          <n-input v-model:value="form.code" :disabled="editing" />
        </n-form-item>
        <n-form-item label="名称" required>
          <n-input v-model:value="form.name" />
        </n-form-item>
        <n-form-item label="类型" required>
          <n-select v-model:value="form.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item label="路径">
          <n-input v-model:value="form.path" placeholder="menu 类型填写路由" />
        </n-form-item>
        <n-form-item label="图标">
          <n-input v-model:value="form.icon" />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="form.sort" />
        </n-form-item>
        <n-form-item label="状态">
          <n-select v-model:value="form.status" :options="statusOptions" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showForm = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { NButton, NSpace, useMessage, type TreeSelectOption } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import * as api from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const rawTree = ref<any[]>([])
const showForm = ref(false)
const editing = ref(false)
const editId = ref<number | null>(null)

const form = reactive({
  parentId: null as number | null,
  code: '',
  name: '',
  type: 'menu',
  path: '' as string | null,
  icon: '' as string | null,
  sort: 0,
  status: 'enabled',
})

const typeOptions = [
  { label: '目录 dir', value: 'dir' },
  { label: '菜单 menu', value: 'menu' },
  { label: '按钮 button', value: 'button' },
]
const statusOptions = [
  { label: '启用', value: 'enabled' },
  { label: '停用', value: 'disabled' },
]

function flatten(nodes: any[], level = 0): any[] {
  const out: any[] = []
  for (const n of nodes || []) {
    out.push({ ...n, level })
    if (n.children?.length) out.push(...flatten(n.children, level + 1))
  }
  return out
}

const tableData = computed(() => {
  // tree table via children
  return rawTree.value
})

function toParentOptions(nodes: any[], disableId?: number | null): TreeSelectOption[] {
  return (nodes || [])
    .filter((n) => n.type !== 'button')
    .map((n) => ({
      key: n.id,
      label: n.name,
      disabled: disableId === n.id,
      children: n.children?.length ? toParentOptions(n.children, disableId) : undefined,
    }))
}

const parentOptions = computed(() => toParentOptions(rawTree.value, editId.value))

const columns: DataTableColumns<any> = [
  {
    title: '名称',
    key: 'name',
    render(row) {
      return `${'　'.repeat(row._level || 0)}${row.name}`
    },
  },
  { title: '编码', key: 'code' },
  { title: '类型', key: 'type', width: 90 },
  { title: '路径', key: 'path' },
  { title: '排序', key: 'sort', width: 70 },
  { title: '状态', key: 'status', width: 90 },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          auth.has('menu.create')
            ? h(NButton, { text: true, onClick: () => openCreate(row.id) }, { default: () => '加子级' })
            : null,
          auth.has('menu.edit')
            ? h(NButton, { text: true, type: 'primary', onClick: () => openEdit(row) }, { default: () => '编辑' })
            : null,
          auth.has('menu.delete')
            ? h(NButton, { text: true, type: 'error', onClick: () => onDelete(row) }, { default: () => '删除' })
            : null,
        ],
      })
    },
  },
]

function withLevel(nodes: any[], level = 0): any[] {
  return (nodes || []).map((n) => ({
    ...n,
    _level: level,
    children: n.children?.length ? withLevel(n.children, level + 1) : undefined,
  }))
}

async function load() {
  loading.value = true
  try {
    const res: any = await api.menuTree()
    rawTree.value = withLevel(res.data || [])
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

function openCreate(parentId: number | null) {
  editing.value = false
  editId.value = null
  form.parentId = parentId
  form.code = ''
  form.name = ''
  form.type = parentId ? 'menu' : 'dir'
  form.path = ''
  form.icon = ''
  form.sort = 0
  form.status = 'enabled'
  showForm.value = true
}

function openEdit(row: any) {
  editing.value = true
  editId.value = row.id
  form.parentId = row.parentId ?? null
  form.code = row.code
  form.name = row.name
  form.type = row.type
  form.path = row.path || ''
  form.icon = row.icon || ''
  form.sort = row.sort || 0
  form.status = row.status
  showForm.value = true
}

async function save() {
  saving.value = true
  try {
    const payload = {
      parentId: form.parentId,
      code: form.code,
      name: form.name,
      type: form.type,
      path: form.path || null,
      icon: form.icon || null,
      sort: form.sort,
      status: form.status,
    }
    if (editing.value && editId.value) {
      await api.updateMenu(editId.value, payload)
    } else {
      await api.createMenu(payload)
    }
    message.success('已保存')
    showForm.value = false
    load()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    saving.value = false
  }
}

async function onDelete(row: any) {
  try {
    await api.deleteMenu(row.id)
    message.success('已删除')
    load()
  } catch (e: any) {
    message.error(e.message)
  }
}

onMounted(load)
</script>
