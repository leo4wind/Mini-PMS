<template>
  <n-modal
    :show="show"
    preset="card"
    :title="title"
    :style="{ width: 'min(960px, 92vw)' }"
    :bordered="false"
    :mask-closable="false"
    :trap-focus="false"
    display-directive="show"
    @update:show="emit('update:show', $event)"
  >
    <div class="entity-modal-body">
      <slot />
    </div>
  </n-modal>
</template>

<script setup lang="ts">
defineProps<{
  show: boolean
  title: string
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()
</script>

<style scoped>
.entity-modal-body {
  max-height: min(78vh, 820px);
  overflow: auto;
  padding-right: 4px;
}
</style>
