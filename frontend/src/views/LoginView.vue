<template>
  <div class="login-wrap">
    <n-card title="MiniPMS 登录" style="width: 380px">
      <n-form @submit.prevent="onSubmit">
        <n-form-item label="账号">
          <n-input v-model:value="account" placeholder="admin" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input v-model:value="password" type="password" show-password-on="click" placeholder="password" />
        </n-form-item>
        <n-button type="primary" attr-type="submit" block :loading="loading">登录</n-button>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const account = ref('admin')
const password = ref('password')
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const message = useMessage()

async function onSubmit() {
  loading.value = true
  try {
    await auth.login(account.value, password.value)
    message.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.replace(redirect)
  } catch (e: any) {
    message.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #eef2f7, #f8fafc);
}
</style>
