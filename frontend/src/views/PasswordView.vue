<template>
  <n-card title="修改密码" style="max-width: 480px">
    <n-form @submit.prevent="onSubmit">
      <n-form-item label="旧密码">
        <n-input v-model:value="oldPassword" type="password" show-password-on="click" />
      </n-form-item>
      <n-form-item label="新密码">
        <n-input v-model:value="newPassword" type="password" show-password-on="click" />
      </n-form-item>
      <n-button type="primary" attr-type="submit" :loading="loading">保存</n-button>
    </n-form>
  </n-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import { changePassword } from '@/api'

const oldPassword = ref('')
const newPassword = ref('')
const loading = ref(false)
const message = useMessage()

async function onSubmit() {
  loading.value = true
  try {
    await changePassword(oldPassword.value, newPassword.value)
    message.success('密码已更新')
    oldPassword.value = ''
    newPassword.value = ''
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
}
</script>
