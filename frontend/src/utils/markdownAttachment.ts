import { useAuthStore } from '@/stores/auth'

const ATTACHMENT_PREFIX = 'attachment:'
const PREVIEW_PATH_RE = /(?:^|\/)api\/v1\/attachments\/(\d+)\/preview(?:\?.*)?$/i

export function attachmentPreviewPath(id: number | string) {
  return `/api/v1/attachments/${id}/preview`
}

export function parseAttachmentId(src: string): string | null {
  if (!src) return null
  if (src.startsWith(ATTACHMENT_PREFIX)) {
    return src.slice(ATTACHMENT_PREFIX.length).split(/[?#]/)[0] || null
  }
  const m = src.replace(/^https?:\/\/[^/]+/i, '').match(PREVIEW_PATH_RE)
  return m?.[1] || null
}

export function withAuthPreviewUrl(src: string): string {
  const id = parseAttachmentId(src)
  if (!id) return src
  const auth = useAuthStore()
  const path = attachmentPreviewPath(id)
  return auth.token ? `${path}?token=${encodeURIComponent(auth.token)}` : path
}

/** 持久化到描述：去掉 token，统一成 /api/v1/attachments/:id/preview */
export function toStoredPreviewUrl(src: string): string {
  const id = parseAttachmentId(src)
  if (id) return attachmentPreviewPath(id)
  return src
}

export function rewriteHtmlMediaSrcs(html: string, mapSrc: (src: string) => string): string {
  if (!html) return html
  const doc = new DOMParser().parseFromString(`<div>${html}</div>`, 'text/html')
  const root = doc.body.firstElementChild
  if (!root) return html
  root.querySelectorAll('img, video, source').forEach((el) => {
    const src = el.getAttribute('src')
    if (src) el.setAttribute('src', mapSrc(src))
  })
  return root.innerHTML
}

export function htmlForEditorDisplay(html: string): string {
  return rewriteHtmlMediaSrcs(html, withAuthPreviewUrl)
}

export function htmlForStorage(html: string): string {
  return rewriteHtmlMediaSrcs(html, (src) => {
    if (src.startsWith('blob:')) return src
    return toStoredPreviewUrl(src)
  })
}
