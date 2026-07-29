<template>
  <div v-if="!content" class="md-empty">-</div>
  <div v-else-if="isHtml" class="desc-html" v-html="safeHtml" />
  <MdPreview v-else :model-value="content" language="zh-CN" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/preview.css'
import { htmlForEditorDisplay } from '@/utils/markdownAttachment'

const props = defineProps<{
  modelValue?: string | null
}>()

const content = computed(() => (props.modelValue || '').trim())
const isHtml = computed(() => /<[a-z][\s\S]*>/i.test(content.value))
const safeHtml = computed(() => sanitizeDescriptionHtml(htmlForEditorDisplay(content.value)))

function sanitizeDescriptionHtml(html: string): string {
  const doc = new DOMParser().parseFromString(`<div>${html}</div>`, 'text/html')
  const root = doc.body.firstElementChild
  if (!root) return ''

  for (const el of Array.from(root.querySelectorAll('*'))) {
    for (const attr of Array.from(el.attributes)) {
      const name = attr.name.toLowerCase()
      if (name.startsWith('on') || name === 'srcdoc') el.removeAttribute(attr.name)
    }
    if (el.tagName === 'SCRIPT' || el.tagName === 'IFRAME' || el.tagName === 'OBJECT') {
      el.remove()
      continue
    }
    if (el.tagName === 'IMG' || el.tagName === 'VIDEO' || el.tagName === 'SOURCE') {
      const src = el.getAttribute('src') || ''
      const ok =
        src.startsWith('data:image/') ||
        src.startsWith('data:video/') ||
        src.startsWith('https://') ||
        src.startsWith('http://') ||
        src.startsWith('/') ||
        src.startsWith('blob:')
      if (!ok) el.remove()
      else if (el.tagName === 'VIDEO') {
        el.setAttribute('controls', 'true')
      }
    }
    if (el.tagName === 'A') {
      const href = el.getAttribute('href') || ''
      if (href.toLowerCase().startsWith('javascript:')) el.removeAttribute('href')
    }
  }
  return root.innerHTML
}
</script>

<style scoped>
.md-empty {
  color: var(--n-text-color-3, #999);
}
.desc-html :deep(img),
.desc-html :deep(video) {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 8px 0;
}
.desc-html :deep(p) {
  margin: 0.4em 0;
}
</style>
