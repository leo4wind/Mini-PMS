/** App public path, e.g. `/mini-pms/` (from Vite `base`). */
export const APP_BASE = import.meta.env.BASE_URL || '/mini-pms/'

/** API prefix under the same public path: `/mini-pms/api/v1`. */
export const API_BASE = `${APP_BASE.replace(/\/$/, '')}/api/v1`

/** Strip public base from a path/fullPath so Vue Router gets an in-app path. */
export function stripAppBase(fullPath: string): string {
  if (!fullPath) return '/'
  const base = APP_BASE.replace(/\/$/, '')
  if (!base) return fullPath.startsWith('/') ? fullPath : `/${fullPath}`
  if (fullPath === base || fullPath === `${base}/`) return '/'
  if (fullPath.startsWith(`${base}/`)) {
    const rest = fullPath.slice(base.length)
    return rest.startsWith('/') ? rest : `/${rest}`
  }
  return fullPath.startsWith('/') ? fullPath : `/${fullPath}`
}
