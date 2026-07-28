<template>
  <n-spin :show="loading">
    <n-space vertical>
      <n-page-header :title="`分配菜单 - ${roleName}`" @back="$router.push('/system/roles')" />
      <n-card>
        <n-tree
          block-line
          checkable
          cascade
          :data="treeData"
          key-field="key"
          label-field="label"
          children-field="children"
          :checked-keys="checkedKeys"
          @update:checked-keys="onCheck"
        />
        <template #footer>
          <n-space justify="end">
            <n-button @click="$router.push('/system/roles')">取消</n-button>
            <n-button type="primary" :loading="saving" @click="save">保存</n-button>
          </n-space>
        </template>
      </n-card>
      <n-alert type="warning" title="提示">
        经理角色保存时会校验：必须保留系统管理相关菜单，避免无法进入系统。
      </n-alert>
    </n-space>
  </n-spin>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage, type TreeOption } from 'naive-ui'
import * as api from '@/api'

const route = useRoute()
const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const roleName = ref('')
const treeData = ref<TreeOption[]>([])
const checkedKeys = ref<(string | number)[]>([])
const allKeys = ref<number[]>([])

function mapTree(nodes: any[]): TreeOption[] {
  return (nodes || []).map((n) => {
    allKeys.value.push(n.id)
    const opt: TreeOption = {
      key: n.id,
      label: `${n.name} (${n.code}) [${n.type}]`,
    }
    if (n.children?.length) opt.children = mapTree(n.children)
    return opt
  })
}

function onCheck(keys: (string | number)[]) {
  checkedKeys.value = keys
}

async function load() {
  loading.value = true
  allKeys.value = []
  try {
    const roleId = route.params.id as string
    const [rolesRes, treeRes, checkedRes]: any[] = await Promise.all([
      api.listRoles(),
      api.menuTree(),
      api.getRoleMenus(roleId),
    ])
    const role = (rolesRes.data || []).find((r: any) => String(r.id) === String(roleId))
    roleName.value = role ? `${role.name}(${role.code})` : `#${roleId}`
    treeData.value = mapTree(treeRes.data || [])
    checkedKeys.value = checkedRes.data?.menuIds || []
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const ids = checkedKeys.value.map((k) => Number(k)).filter((n) => !Number.isNaN(n))
    await api.assignRoleMenus(route.params.id as string, ids)
    message.success('菜单权限已保存')
  } catch (e: any) {
    message.error(e.message)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
