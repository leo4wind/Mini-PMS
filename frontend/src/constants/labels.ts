export const storyTypeMap: Record<string, string> = {
  planning: '规划',
  story: '可交付',
}

export const storyStatusMap: Record<string, string> = {
  draft: '草稿',
  active: '激活',
  closed: '关闭',
}

export const sprintStatusMap: Record<string, string> = {
  wait: '未开始',
  doing: '进行中',
  done: '已完成',
  closed: '已关闭',
}

export const bugStatusMap: Record<string, string> = {
  active: '激活',
  resolved: '已解决',
  closed: '已关闭',
}

export const bugResolutionMap: Record<string, string> = {
  fixed: '已修复',
  duplicate: '重复',
  willnotfix: '不予修复',
  external: '外部原因',
  bydesign: '设计如此',
  notrepro: '无法重现',
}

export const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp']

export function isImageExt(ext: string) {
  return imageExts.includes(ext.toLowerCase())
}

export function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}
