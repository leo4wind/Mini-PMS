<template>
  <n-space vertical>
    <n-space justify="space-between">
      <n-h2 style="margin: 0">用户管理</n-h2>
      <n-button v-if="auth.has('user.create')" type="primary" @click="openCreate">新建用户</n-button>
    </n-space>
    <n-space>
      <n-select v-model:value="status" :options="statusOptions" clearable placeholder="状态" style="width: 140px" />
      <n-input v-model:value="keyword" placeholder="账号/姓名" clearable style="width: 200px" />
      <n-button @click="load">查询</n-button>
    </n-space>
    <n-data-table :columns="columns" :data="list" :loading="loading" :pagination="pagination" remote @update:page="onPage" />

    <n-modal v-model:show="showForm" preset="card" :title="editing ? '编辑用户' : '新建用户'" style="width: 480px">
      <n-form>
        <n-form-item label="账号" v-if="!editing">
          <n-input v-model:value="form.account" />
        </n-form-item>
        <n-form-item label="密码" v-if="!editing">
          <n-input v-model:value="form.password" type="password" show-password-on="click" placeholder="默认 123456" />
        </n-form-item>
        <n-form-item label="姓名">
          <n-input v-model:value="form.realname" />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="form.email" />
        </n-form-item>
        <n-form-item v-if="!editing" label="角色">
          <n-select v-model:value="form.roleIds" multiple :options="roleOptions" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showForm = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="saveForm">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showRoles" preset="card" title="分配角色" style="width: 420px">
      <n-select v-model:value="assignRoleIds" multiple :options="roleOptions" />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showRoles = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="saveRoles">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { NButton, NSpace, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import * as api from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const message = useMessage()
const list = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const status = ref<string | null>(null)
const keyword = ref('')
const pagination = reactive({ page: 1, pageSize: 20, itemCount: 0 })
const roleOptions = ref<{ label: string; value: number }[]>([])

const showForm = ref(false)
const editing = ref(false)
const editId = ref<number | null>(null)
const form = reactive({ account: '', password: '', realname: '', email: '', roleIds: [] as number[] })

const showRoles = ref(false)
const assignUserId = ref<number | null>(null)
const assignRoleIds = ref<number[]>([])

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'disabled' },
]

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '账号', key: 'account' },
  { title: '姓名', key: 'realname' },
  { title: '邮箱', key: 'email' },
  {
    title: '角色',
    key: 'roles',
    render(row) {
      return h(NSpace, { size: 4 }, {
        default: () => (row.roles || []).map((r: any) => h(NTag, { size: 'small', type: 'info' }, { default: () => r.name })),
      })
    },
  },
  { title: '状态', key: 'status', width: 90 },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          auth.has('user.edit')
            ? h(NButton, { text: true, type: 'primary', onClick: () => openEdit(row) }, { default: () => '编辑' })
            : null,
          auth.has('user.disable')
            ? h(
                NButton,
                { text: true, onClick: () => toggleStatus(row) },
                { default: () => (row.status === 'active' ? '停用' : '启用') },
              )
            : null,
          auth.has('user.assignRole')
            ? h(NButton, { text: true, onClick: () => openAssign(row) }, { default: () => '分配角色' })
            : null,
        ],
      })
    },
  },
]

async function loadRoles() {
  if (!auth.has('role.list') && !auth.has('user.assignRole') && !auth.has('user.create')) return
  try {
    const res: any = await api.listRoles()
    roleOptions.value = (res.data || []).map((r: any) => ({ label: `${r.name}(${r.code})`, value: r.id }))
  } catch {
    // role.list 无权限时，创建用户仍可能需要角色；尝试用 listRoles 失败则忽略
  }
}

async function load() {
  loading.value = true
  try {
    const res: any = await api.listUsers({
      page: pagination.page,
      pageSize: pagination.pageSize,
      status: status.value || undefined,
      keyword: keyword.value || undefined,
    })
    list.value = res.data.list || []
    pagination.itemCount = res.data.total || 0
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

function onPage(p: number) {
  pagination.page = p
  load()
}

function openCreate() {
  editing.value = false
  editId.value = null
  form.account = ''
  form.password = '123456'
  form.realname = ''
  form.email = ''
  form.roleIds = []
  showForm.value = true
}

function openEdit(row: any) {
  editing.value = true
  editId.value = row.id
  form.realname = row.realname
  form.email = row.email || ''
  showForm.value = true
}

function openAssign(row: any) {
  assignUserId.value = row.id
  assignRoleIds.value = (row.roles || []).map((r: any) => r.id)
  showRoles.value = true
}

async function saveForm() {
  saving.value = true
  try {
    if (editing.value && editId.value) {
      await api.updateUser(editId.value, { realname: form.realname, email: form.email || null })
    } else {
      await api.createUser({
        account: form.account,
        password: form.password,
        realname: form.realname,
        email: form.email || null,
        roleIds: form.roleIds,
      })
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

async function saveRoles() {
  if (!assignUserId.value) return
  saving.value = true
  try {
    await api.assignUserRoles(assignUserId.value, assignRoleIds.value)
    message.success('角色已更新')
    showRoles.value = false
    load()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    saving.value = false
  }
}

async function toggleStatus(row: any) {
  try {
    if (row.status === 'active') await api.disableUser(row.id)
    else await api.enableUser(row.id)
    message.success('已更新')
    load()
  } catch (e: any) {
    message.error(e.message)
  }
}

onMounted(async () => {
  await loadRoles()
  // 分配角色需要角色列表：若无 role.list，用一个不鉴权不了的接口——种子里 manager 都有。
  // 额外：让有 user.assignRole 的也能拉角色。后端 listRoles 目前要 role.list。
  // 为方便分配，若失败则再试一次（manager 有权限）。
  if (!roleOptions.value.length) {
    try {
      const res: any = await api.listRoles()
      roleOptions.value = (res.data || []).map((r: any) => ({ label: `${r.name}(${r.code})`, value: r.id }))
    } catch {}
  }
  load()
})
</script>
