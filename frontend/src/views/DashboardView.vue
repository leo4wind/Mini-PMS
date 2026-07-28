<template>
  <n-space vertical size="large">
    <n-h2 style="margin: 0">工作台</n-h2>
    <n-grid :cols="3" :x-gap="16" :y-gap="16">
      <n-gi>
        <n-card title="欢迎">
          {{ auth.user?.realname || auth.user?.account }}，已登录 MiniPMS。
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="快捷入口">
          <n-space>
            <n-button v-if="auth.has('product.create')" type="primary" @click="$router.push('/products/new')">
              新建产品
            </n-button>
            <n-button v-if="auth.has('story.create')" @click="$router.push('/stories/new')">新建需求</n-button>
            <n-button v-if="auth.has('bug.create')" @click="$router.push('/bugs/new')">新建缺陷</n-button>
            <n-button v-if="auth.has('product.list')" @click="$router.push('/products')">产品列表</n-button>
          </n-space>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="角色">
          <n-space>
            <n-tag v-for="r in auth.roles" :key="r.id" type="info">{{ r.name }}</n-tag>
          </n-space>
        </n-card>
      </n-gi>

      <n-gi v-if="auth.has('story.list')">
        <n-card title="我的需求">
          <template #header-extra>
            <n-button text type="primary" @click="$router.push('/stories?assignedTo=me')">查看全部</n-button>
          </template>
          <n-spin :show="loading">
            <n-empty v-if="!summary?.myStories?.length" description="暂无待办需求" size="small" />
            <n-list v-else hoverable clickable>
              <n-list-item v-for="s in summary.myStories" :key="s.id" @click="$router.push(`/stories/${s.id}`)">
                <n-thing :title="s.title">
                  <template #description>
                    <n-space size="small">
                      <n-tag size="small">{{ storyStatusMap[s.status] || s.status }}</n-tag>
                      <span v-if="s.productId">产品 #{{ s.productId }}</span>
                    </n-space>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
          </n-spin>
        </n-card>
      </n-gi>

      <n-gi v-if="auth.has('bug.list')">
        <n-card title="我的缺陷">
          <template #header-extra>
            <n-button text type="primary" @click="goMyBugs">查看全部</n-button>
          </template>
          <n-spin :show="loading">
            <n-empty v-if="!summary?.myBugs?.length" description="暂无待办缺陷" size="small" />
            <n-list v-else hoverable clickable>
              <n-list-item v-for="b in summary.myBugs" :key="b.id" @click="$router.push(`/bugs/${b.id}`)">
                <n-thing :title="b.title">
                  <template #description>
                    <n-space size="small">
                      <n-tag size="small">{{ bugStatusMap[b.status] || b.status }}</n-tag>
                      <span v-if="b.productId">产品 #{{ b.productId }}</span>
                    </n-space>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
          </n-spin>
        </n-card>
      </n-gi>

      <n-gi v-if="auth.has('sprint.list')">
        <n-card title="进行中迭代">
          <template #header-extra>
            <n-button text type="primary" @click="$router.push('/sprints?status=doing')">查看全部</n-button>
          </template>
          <n-spin :show="loading">
            <n-empty v-if="!summary?.doingSprints?.length" description="暂无进行中迭代" size="small" />
            <n-list v-else hoverable clickable>
              <n-list-item v-for="s in summary.doingSprints" :key="s.id" @click="$router.push(`/sprints/${s.id}`)">
                <n-thing :title="s.name">
                  <template #description>
                    <span v-if="s.projectId">项目 #{{ s.projectId }}</span>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
          </n-spin>
        </n-card>
      </n-gi>
    </n-grid>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { getDashboardSummary } from '@/api'
import { storyStatusMap, bugStatusMap } from '@/constants/labels'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const message = useMessage()
const loading = ref(false)
const summary = ref<any>(null)

function goMyBugs() {
  if (auth.user?.id) {
    router.push(`/bugs?assignedTo=${auth.user.id}`)
  } else {
    router.push('/bugs')
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const res: any = await getDashboardSummary()
    summary.value = res.data
  } catch (e: any) {
    message.error(e.message)
  } finally {
    loading.value = false
  }
})
</script>
