import type { AdminNotificationResponse, AdminStats, Article, Category, Comment, ExternalLink, NotificationType, TagStat, User } from './types'

export const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api'

type ApiResponse<T> = T

function token() {
  return localStorage.getItem('blog_token') || ''
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers || {})
  if (!(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  const auth = token()
  if (auth) headers.set('Authorization', `Bearer ${auth}`)
  const res = await fetch(`${API_BASE}${path}`, { ...init, headers })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || '请求失败')
  return data as T
}

export const api = {
  setToken(v: string) {
    localStorage.setItem('blog_token', v)
    window.dispatchEvent(new CustomEvent('blog-auth-changed'))
  },
  clearToken() {
    localStorage.removeItem('blog_token')
    window.dispatchEvent(new CustomEvent('blog-auth-changed'))
  },
  health() {
    return request<{ ok: boolean }>('/health')
  },
  requestVerificationCode(payload: { email: string; purpose: 'login' | 'register' }) {
    return request<{ ok: boolean; expiresIn: number }>('/auth/verification-code', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  register(payload: { name: string; email: string; password: string; confirmPassword: string; verificationCode: string }) {
    return request<{ token: string; user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  login(payload: { email: string; password: string; verificationCode: string }) {
    return request<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  me() {
    return request<{ user: User }>('/auth/me')
  },
  updateMe(payload: { name: string; avatarUrl?: string }) {
    return request<{ user: User }>('/auth/me', {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
  },
  uploadAvatar(file: File) {
    const body = new FormData()
    body.append('file', file)
    return request<{ kind: string; url: string; filename: string }>('/auth/me/avatar', {
      method: 'POST',
      body,
    })
  },
  articles(params: { q?: string; category?: string; page?: number; pageSize?: number } = {}) {
    const query = new URLSearchParams()
    if (params.q) query.set('q', params.q)
    if (params.category) query.set('category', params.category)
    if (params.page) query.set('page', String(params.page))
    if (params.pageSize) query.set('pageSize', String(params.pageSize))
    return request<{ items: Article[]; total: number; pages: number }>(`/articles?${query.toString()}`)
  },
  popularArticles() {
    return request<{ items: Article[] }>('/articles/popular')
  },
  article(slug: string) {
    return request<{ article: Article; comments: Comment[]; related: Article[] }>(`/articles/${slug}`)
  },
  categories() {
    return request<{ items: Category[] }>('/categories')
  },
  tags() {
    return request<{ items: TagStat[] }>('/tags')
  },
  externalLinks() {
    return request<{ items: ExternalLink[] }>('/external-links')
  },
  comments(articleId: number) {
    return request<{ items: Comment[] }>(`/articles/id/${articleId}/comments`)
  },
  addComment(articleId: number, body: string) {
    return request<{ comment: Comment }>(`/articles/id/${articleId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    })
  },
  toggleReaction(articleId: number, type: 'like' | 'favorite') {
    return request<{ on: boolean }>(`/articles/id/${articleId}/reactions/${type}`, { method: 'POST' })
  },
  library() {
    return request<{ likes: number[]; favorites: number[] }>('/me/library')
  },
  adminArticles(params: { q?: string; status?: string } = {}) {
    const query = new URLSearchParams()
    if (params.q) query.set('q', params.q)
    if (params.status) query.set('status', params.status)
    return request<{ items: Article[] }>(`/admin/articles?${query.toString()}`)
  },
  adminStats() {
    return request<AdminStats>('/admin/stats')
  },
  adminNotifications() {
    return request<AdminNotificationResponse>('/admin/notifications')
  },
  deleteNotifications() {
    return request<{ ok: boolean }>('/admin/notifications', {
      method: 'DELETE',
    }).catch(async (error) => {
      const message = error instanceof Error ? error.message : ''
      if (!message.includes('404')) throw error
      return request<{ ok: boolean }>('/admin/notifications/clear', {
        method: 'POST',
      })
    })
  },
  markNotificationsRead(types: NotificationType[] = [], articleIds: number[] = []) {
    return request<{ ok: boolean }>('/admin/notifications/read', {
      method: 'PATCH',
      body: JSON.stringify({ types, articleIds }),
    })
  },
  uploadAdminFile(file: File, kind: 'image' | 'markdown') {
    const body = new FormData()
    body.append('file', file)
    body.append('kind', kind)
    return request<{ kind: string; url?: string; content?: string; title?: string; filename: string }>('/admin/uploads', {
      method: 'POST',
      body,
    })
  },
  saveArticle(payload: {
    id?: number
    title: string
    summary?: string
    content: string
    tags?: string
    status?: 'draft' | 'published'
    categoryId?: number
  }) {
    const method = payload.id ? 'PUT' : 'POST'
    const path = payload.id ? `/admin/articles/${payload.id}` : '/admin/articles'
    return request<{ ok?: boolean; article?: Article }>(path, {
      method,
      body: JSON.stringify(payload),
    })
  },
  generateArticleSummary(payload: { title?: string; content: string }) {
    return request<{ summary: string }>('/admin/articles/summary', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  deleteArticle(id: number) {
    return request<{ ok: boolean }>(`/admin/articles/${id}`, { method: 'DELETE' })
  },
  adminCategories() {
    return request<{ items: Category[] }>('/admin/categories')
  },
  adminExternalLinks() {
    return request<{ items: ExternalLink[] }>('/admin/external-links')
  },
  adminComments() {
    return request<{ items: Comment[] }>('/admin/comments')
  },
  adminUsers() {
    return request<{ items: User[] }>('/admin/users')
  },
  saveExternalLink(payload: {
    id?: number
    name: string
    platform: string
    description?: string
    linkUrl?: string
    qrCodeUrl?: string
    sortOrder?: number
    isActive?: boolean
  }) {
    const method = payload.id ? 'PUT' : 'POST'
    const path = payload.id ? `/admin/external-links/${payload.id}` : '/admin/external-links'
    return request<{ item?: ExternalLink; ok?: boolean }>(path, {
      method,
      body: JSON.stringify(payload),
    })
  },
  deleteExternalLink(id: number) {
    return request<{ ok: boolean }>(`/admin/external-links/${id}`, { method: 'DELETE' })
  },
  saveCategory(payload: { id?: number; name: string }) {
    const method = payload.id ? 'PUT' : 'POST'
    const path = payload.id ? `/admin/categories/${payload.id}` : '/admin/categories'
    return request(path, {
      method,
      body: JSON.stringify({ name: payload.name }),
    })
  },
  moderateComment(id: number, status: 'published' | 'hidden') {
    return request(`/admin/comments/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    })
  },
  deleteCategory(id: number) {
    return request(`/admin/categories/${id}`, { method: 'DELETE' })
  },
}

export type Api = typeof api
