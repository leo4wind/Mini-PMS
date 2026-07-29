<template>
  <n-drawer
    :show="show"
    :width="width"
    placement="right"
    :trap-focus="false"
    :block-scroll="false"
    @update:show="emit('update:show', $event)"
  >
    <n-drawer-content :title="title" closable :native-scrollbar="false">
      <slot />
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
defineProps<{
  show: boolean
  title: string
  width: number
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()
</script>
