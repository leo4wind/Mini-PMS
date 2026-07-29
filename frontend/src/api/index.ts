import http from './http'

export const login = (account: string, password: string) =>
  http.post('/auth/login', { account, password })

export const fetchMe = () => http.get('/auth/me')

export const logout = () => http.post('/auth/logout')

export const changePassword = (oldPassword: string, newPassword: string) =>
  http.put('/auth/password', { oldPassword, newPassword })

export const listProducts = (params: Record<string, unknown>) =>
  http.get('/products', { params })

export const createProduct = (data: Record<string, unknown>) =>
  http.post('/products', data)

export const getProduct = (id: number | string) => http.get(`/products/${id}`)

export const updateProduct = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/products/${id}`, data)

export const deleteProduct = (id: number | string) => http.delete(`/products/${id}`)

export const listProductProjects = (productId: number | string) =>
  http.get(`/products/${productId}/projects`)

// projects
export const listProjects = (params: Record<string, unknown>) =>
  http.get('/projects', { params })
export const createProject = (data: Record<string, unknown>) => http.post('/projects', data)
export const getProject = (id: number | string) => http.get(`/projects/${id}`)
export const updateProject = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/projects/${id}`, data)
export const deleteProject = (id: number | string) => http.delete(`/projects/${id}`)
export const listProjectSprints = (projectId: number | string) =>
  http.get(`/projects/${projectId}/sprints`)
export const listProjectStories = (projectId: number | string) =>
  http.get(`/projects/${projectId}/stories`)

// stories
export const listStories = (params: Record<string, unknown>) =>
  http.get('/stories', { params })
export const createStory = (data: Record<string, unknown>) => http.post('/stories', data)
export const getStory = (id: number | string) => http.get(`/stories/${id}`)
export const updateStory = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/stories/${id}`, data)
export const deleteStory = (id: number | string) => http.delete(`/stories/${id}`)
export const uploadStoryAttachment = (id: number | string, formData: FormData) =>
  http.post(`/stories/${id}/attachments`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
export const listStoryRemarks = (storyId: number | string) =>
  http.get(`/stories/${storyId}/remarks`)
export const createStoryRemarkDraft = (storyId: number | string) =>
  http.post(`/stories/${storyId}/remarks`)
export const finalizeStoryRemark = (
  storyId: number | string,
  remarkId: number | string,
  data: { content: string },
) => http.post(`/stories/${storyId}/remarks/${remarkId}/finalize`, data)
export const discardStoryRemarkDraft = (storyId: number | string, remarkId: number | string) =>
  http.delete(`/stories/${storyId}/remarks/${remarkId}`)
export const uploadStoryRemarkAttachment = (
  storyId: number | string,
  remarkId: number | string,
  formData: FormData,
) =>
  http.post(`/stories/${storyId}/remarks/${remarkId}/attachments`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

// sprints
export const listSprints = (params: Record<string, unknown>) =>
  http.get('/sprints', { params })
export const createSprint = (data: Record<string, unknown>) => http.post('/sprints', data)
export const getSprint = (id: number | string) => http.get(`/sprints/${id}`)
export const updateSprint = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/sprints/${id}`, data)
export const deleteSprint = (id: number | string) => http.delete(`/sprints/${id}`)
export const listSprintStories = (sprintId: number | string) =>
  http.get(`/sprints/${sprintId}/stories`)
export const linkSprintStories = (sprintId: number | string, storyIds: number[]) =>
  http.post(`/sprints/${sprintId}/stories`, { storyIds })
export const unlinkSprintStory = (sprintId: number | string, storyId: number | string) =>
  http.delete(`/sprints/${sprintId}/stories/${storyId}`)
export const listSprintStoryCandidates = (sprintId: number | string, params: Record<string, unknown>) =>
  http.get(`/sprints/${sprintId}/story-candidates`, { params })

// bugs
export const listBugs = (params: Record<string, unknown>) => http.get('/bugs', { params })
export const createBug = (data: Record<string, unknown>) => http.post('/bugs', data)
export const getBug = (id: number | string) => http.get(`/bugs/${id}`)
export const updateBug = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/bugs/${id}`, data)
export const resolveBug = (id: number | string, resolution: string) =>
  http.post(`/bugs/${id}/resolve`, { resolution })
export const closeBug = (id: number | string) => http.post(`/bugs/${id}/close`)
export const activateBug = (id: number | string) => http.post(`/bugs/${id}/activate`)
export const deleteBug = (id: number | string) => http.delete(`/bugs/${id}`)
export const uploadBugAttachment = (id: number | string, formData: FormData) =>
  http.post(`/bugs/${id}/attachments`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

// attachments
export const deleteAttachment = (id: number | string) => http.delete(`/attachments/${id}`)

// dashboard
export const getDashboardSummary = () => http.get('/dashboard/summary')

// system - users
export const listUsers = (params: Record<string, unknown>) => http.get('/users', { params })
export const getUser = (id: number | string) => http.get(`/users/${id}`)
export const createUser = (data: Record<string, unknown>) => http.post('/users', data)
export const updateUser = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/users/${id}`, data)
export const disableUser = (id: number | string) => http.post(`/users/${id}/disable`)
export const enableUser = (id: number | string) => http.post(`/users/${id}/enable`)
export const assignUserRoles = (id: number | string, roleIds: number[]) =>
  http.put(`/users/${id}/roles`, { roleIds })

// system - roles
export const listRoles = () => http.get('/roles')
export const updateRole = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/roles/${id}`, data)
export const getRoleMenus = (id: number | string) => http.get(`/roles/${id}/menus`)
export const assignRoleMenus = (id: number | string, menuIds: number[]) =>
  http.put(`/roles/${id}/menus`, { menuIds })

// system - menus
export const menuTree = () => http.get('/menus/tree')
export const createMenu = (data: Record<string, unknown>) => http.post('/menus', data)
export const updateMenu = (id: number | string, data: Record<string, unknown>) =>
  http.put(`/menus/${id}`, data)
export const deleteMenu = (id: number | string) => http.delete(`/menus/${id}`)
