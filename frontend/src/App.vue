<template>
  <div class="shell">
    <header class="topbar">
      <RouterLink class="brand" to="/">
        <SiteLogo />
        <span class="brand-wordmark" aria-label="xiaoli博客">
          <span class="brand-xiaoli">xiaoli</span>
          <span class="brand-blog">博客</span>
        </span>
      </RouterLink>

      <nav class="nav">
        <RouterLink to="/"><HomeIcon :size="16" /> 首页</RouterLink>
        <RouterLink v-if="user?.role === 'admin'" to="/admin"><SettingsIcon :size="16" /> 后台</RouterLink>

        <div
          v-if="user?.role === 'admin'"
          class="notification-menu"
        >
          <button class="icon-trigger" type="button" @click.stop="toggleNotificationMenu">
            <BellIcon :size="17" />
            <span v-if="unreadTotal > 0" class="nav-badge">{{ unreadTotal > 99 ? '99+' : unreadTotal }}</span>
          </button>

          <div
            v-if="notificationMenuOpen"
            class="notification-popover"
            @click.stop
          >
            <div class="notification-popover-head">
              <strong>消息通知</strong>
              <div class="notification-head-actions">
                <button type="button" class="ghost" @click.stop="markAllNotificationsRead">全部已读</button>
                <button type="button" class="ghost danger" :disabled="clearingNotifications" @click.stop="clearNotifications">
                  {{ clearingNotifications ? '清空中' : '清空' }}
                </button>
              </div>
            </div>
            <div class="notification-list">
              <article v-for="item in notifications" :key="item.id" :class="['notification-item', { unread: !item.isRead }]">
                <span class="notification-dot"></span>
                <div>
                  <strong>{{ item.title }}</strong>
                  <p>{{ item.content }}</p>
                  <span>{{ formatNotificationTime(item.createdAt) }}</span>
                </div>
              </article>
              <p v-if="!notifications.length" class="notification-empty">暂无通知</p>
            </div>
            <p v-if="notificationError" class="notification-error">{{ notificationError }}</p>
          </div>
        </div>

        <div
          v-if="user"
          class="user-menu"
          @mouseenter="openUserMenu"
          @mouseleave="scheduleCloseUserMenu"
        >
          <button class="avatar-trigger" type="button" @click="openProfileEditor">
            <img class="topbar-avatar" :src="user.avatarUrl || defaultAvatar" :alt="user.name" />
          </button>

          <div
            v-if="userMenuOpen"
            class="user-popover"
            @mouseenter="openUserMenu"
            @mouseleave="scheduleCloseUserMenu"
          >
            <img class="user-popover-avatar" :src="user.avatarUrl || defaultAvatar" :alt="user.name" />
            <div class="user-popover-body">
              <strong>{{ user.name }}</strong>
              <p>{{ user.email }}</p>
              <span>{{ user.role === 'admin' ? '管理员' : '普通用户' }}</span>
              <div class="user-popover-actions">
                <button type="button" class="ghost" @click="openProfileEditor"><UserPenIcon :size="14" /> 编辑资料</button>
                <button type="button" class="ghost danger" @click="logout"><LogOutIcon :size="14" /> 退出登录</button>
              </div>
            </div>
          </div>
        </div>

        <RouterLink v-else to="/auth"><LogInIcon :size="16" /> 登录</RouterLink>
      </nav>
    </header>

    <main
      class="content"
      :class="{ 'article-content': isArticlePage, 'home-content': isHomePage, 'admin-content': isAdminPage }"
    >
      <router-view />
    </main>

    <div v-if="profileOpen" class="profile-modal-backdrop" @click.self="closeProfileEditor">
      <section class="profile-modal" role="dialog" aria-modal="true" aria-label="编辑个人资料">
        <div class="profile-modal-head">
          <div>
            <p class="eyebrow">个人资料</p>
            <h2>编辑头像与昵称</h2>
          </div>
          <button class="ghost" type="button" @click="closeProfileEditor">关闭</button>
        </div>

        <div class="profile-modal-body">
          <div class="profile-avatar-card">
            <img class="profile-avatar-preview" :src="profileAvatarPreview || defaultAvatar" :alt="profileName" />
            <label class="profile-upload">
              <span>选择本地头像</span>
              <input type="file" accept="image/*" @change="handleAvatarChange" />
            </label>
          </div>

          <div class="profile-form">
            <label>
              <span>昵称</span>
              <input v-model="profileName" placeholder="输入昵称" />
            </label>
            <label>
              <span>邮箱</span>
              <input :value="user?.email || ''" disabled />
            </label>
            <p class="feedback error" v-if="profileError">{{ profileError }}</p>
            <p class="feedback success" v-if="profileNotice">{{ profileNotice }}</p>
            <div class="actions-row">
              <button type="button" @click="saveProfile">保存修改</button>
              <button class="ghost" type="button" @click="closeProfileEditor">取消</button>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { BellIcon, HomeIcon, LogInIcon, LogOutIcon, SettingsIcon, UserPenIcon } from 'lucide-vue-next'
import SiteLogo from './components/SiteLogo.vue'
import { API_BASE, api } from './api'
import type { Notification, User } from './types'

const defaultAvatar =
  'data:image/svg+xml;charset=UTF-8,' +
  encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect width="96" height="96" rx="24" fill="#d1d5db"/>
      <circle cx="48" cy="38" r="18" fill="#9ca3af"/>
      <path d="M20 80c4-17 19-26 28-26s24 9 28 26" fill="#9ca3af"/>
    </svg>
  `)

const user = ref<User | null>(null)
const route = useRoute()
const isHomePage = computed(() => route.path === '/')
const isArticlePage = computed(() => String(route.path || '').startsWith('/article/'))
const isAdminPage = computed(() => String(route.path || '').startsWith('/admin'))
const userMenuOpen = ref(false)
const notificationMenuOpen = ref(false)
const profileOpen = ref(false)
const profileName = ref('')
const profileError = ref('')
const profileNotice = ref('')
const profileAvatarFile = ref<File | null>(null)
const profileAvatarPreview = ref('')
const notifications = ref<Notification[]>([])
const unreadTotal = ref(0)
const clearingNotifications = ref(false)
const notificationError = ref('')
let objectUrl = ''
let userMenuCloseTimer: ReturnType<typeof window.setTimeout> | null = null
let notificationPollTimer: ReturnType<typeof window.setInterval> | null = null
let notificationSource: EventSource | null = null

const syncUser = async () => {
  const token = localStorage.getItem('blog_token')
  if (!token) {
    user.value = null
    notifications.value = []
    unreadTotal.value = 0
    stopNotificationPolling()
    stopNotificationStream()
    return
  }
  try {
    const res = await api.me()
    user.value = res.user
    if (res.user.role === 'admin') {
      await loadNotifications()
      startNotificationStream()
      startNotificationPolling()
    } else {
      notifications.value = []
      unreadTotal.value = 0
      stopNotificationPolling()
      stopNotificationStream()
    }
  } catch {
    api.clearToken()
    user.value = null
    notifications.value = []
    unreadTotal.value = 0
    stopNotificationPolling()
    stopNotificationStream()
  }
}

async function loadNotifications() {
  if (user.value?.role !== 'admin') return
  try {
    const res = await api.adminNotifications()
    notifications.value = res.items
    unreadTotal.value = res.unreadTotal
  } catch {
    // 保留当前列表，避免短暂网络错误把通知面板直接清空
  }
}

function startNotificationPolling() {
  if (notificationPollTimer) return
  notificationPollTimer = window.setInterval(() => {
    if (user.value?.role === 'admin') {
      void loadNotifications()
    }
  }, 30000)
}

function stopNotificationPolling() {
  if (notificationPollTimer) {
    window.clearInterval(notificationPollTimer)
    notificationPollTimer = null
  }
}

function startNotificationStream() {
  const token = localStorage.getItem('blog_token')
  if (!token || user.value?.role !== 'admin' || notificationSource) return
  const source = new EventSource(`${API_BASE}/admin/notifications/stream?token=${encodeURIComponent(token)}`)
  source.onmessage = (event) => {
    if (event.data === 'ping' || event.data === 'connected') return
    void loadNotifications().then(() => {
      window.dispatchEvent(new CustomEvent('blog-notifications-changed'))
    })
  }
  source.onerror = () => {
    stopNotificationStream()
    startNotificationPolling()
  }
  notificationSource = source
}

function stopNotificationStream() {
  if (!notificationSource) return
  notificationSource.close()
  notificationSource = null
}

function setPreviewFromFile(file: File | null) {
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
    objectUrl = ''
  }
  profileAvatarFile.value = file
  if (!file) {
    profileAvatarPreview.value = user.value?.avatarUrl || ''
    return
  }
  objectUrl = URL.createObjectURL(file)
  profileAvatarPreview.value = objectUrl
}

function openUserMenu() {
  if (userMenuCloseTimer) {
    window.clearTimeout(userMenuCloseTimer)
    userMenuCloseTimer = null
  }
  userMenuOpen.value = true
}

function scheduleCloseUserMenu() {
  if (userMenuCloseTimer) {
    window.clearTimeout(userMenuCloseTimer)
  }
  userMenuCloseTimer = window.setTimeout(() => {
    userMenuOpen.value = false
    userMenuCloseTimer = null
  }, 180)
}

function toggleNotificationMenu() {
  notificationMenuOpen.value = !notificationMenuOpen.value
  if (notificationMenuOpen.value) {
    notificationError.value = ''
    void loadNotifications()
  }
}

function openProfileEditor() {
  if (!user.value) return
  userMenuOpen.value = false
  profileName.value = user.value.name
  profileError.value = ''
  profileNotice.value = ''
  profileOpen.value = true
  setPreviewFromFile(null)
}

function closeProfileEditor() {
  profileOpen.value = false
  profileError.value = ''
  profileNotice.value = ''
  setPreviewFromFile(null)
  profileAvatarFile.value = null
}

function handleEscape(event: KeyboardEvent) {
  if (event.key === 'Escape' && profileOpen.value) {
    closeProfileEditor()
  }
}

function handleAvatarChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] || null
  setPreviewFromFile(file)
  input.value = ''
}

async function saveProfile() {
  if (!user.value) return
  const name = profileName.value.trim()
  if (!name) {
    profileError.value = '昵称不能为空'
    return
  }
  profileError.value = ''
  profileNotice.value = ''
  try {
    let avatarUrl = user.value.avatarUrl || ''
    if (profileAvatarFile.value) {
      const uploaded = await api.uploadAvatar(profileAvatarFile.value)
      avatarUrl = uploaded.url
    }
    const res = await api.updateMe({ name, avatarUrl })
    user.value = res.user
    profileNotice.value = '资料已更新'
    window.dispatchEvent(new CustomEvent('blog-auth-changed'))
    closeProfileEditor()
  } catch (e) {
    profileError.value = (e as Error).message
  }
}

function formatNotificationTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function markAllNotificationsRead() {
  try {
    await api.markNotificationsRead()
    await loadNotifications()
    window.dispatchEvent(new CustomEvent('blog-notifications-changed'))
  } catch {
    // ignore
  }
}

async function clearNotifications() {
  if (clearingNotifications.value) return
  clearingNotifications.value = true
  notificationError.value = ''
  const wasMenuOpen = notificationMenuOpen.value
  stopNotificationStream()
  stopNotificationPolling()
  try {
    await api.deleteNotifications()
    notifications.value = []
    unreadTotal.value = 0
    window.dispatchEvent(new CustomEvent('blog-notifications-changed'))
  } catch (e) {
    notificationError.value = `${(e as Error).message}，请确认后端已重启`
    await loadNotifications()
  } finally {
    notificationMenuOpen.value = wasMenuOpen
    if (user.value?.role === 'admin') {
      startNotificationStream()
      startNotificationPolling()
    }
    clearingNotifications.value = false
  }
}

onMounted(async () => {
  await syncUser()
  window.addEventListener('blog-auth-changed', syncUser)
  window.addEventListener('blog-notifications-changed', loadNotifications)
  window.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  window.removeEventListener('blog-auth-changed', syncUser)
  window.removeEventListener('blog-notifications-changed', loadNotifications)
  window.removeEventListener('keydown', handleEscape)
  if (userMenuCloseTimer) {
    window.clearTimeout(userMenuCloseTimer)
  }
  stopNotificationPolling()
  stopNotificationStream()
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
  }
})

function logout() {
  api.clearToken()
  user.value = null
  notificationMenuOpen.value = false
  notifications.value = []
  unreadTotal.value = 0
  stopNotificationPolling()
  stopNotificationStream()
  window.location.href = '/'
}
</script>
