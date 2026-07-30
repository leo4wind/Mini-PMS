import { useAuthStore } from '@/stores/auth'
import { API_BASE } from '@/config'

export async function fetchAttachmentBlob(path: string): Promise<Blob> {
  const auth = useAuthStore()
  const res = await fetch(`${API_BASE}${path}`, {
    headers: auth.token ? { Authorization: `Bearer ${auth.token}` } : {},
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new Error(text || '文件请求失败')
  }
  return res.blob()
}

export async function downloadAttachment(id: number | string, filename: string) {
  const blob = await fetchAttachmentBlob(`/attachments/${id}/download`)
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

export async function previewAttachmentUrl(id: number | string): Promise<string> {
  const blob = await fetchAttachmentBlob(`/attachments/${id}/preview`)
  return URL.createObjectURL(blob)
}
