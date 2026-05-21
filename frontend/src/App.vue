<template>
  <div class="shell">
    <header class="topbar">
      <RouterLink class="brand" to="/">
        <SiteLogo />
        <span>xiaoli博客</span>
      </RouterLink>

      <nav class="nav">
        <RouterLink to="/">首页</RouterLink>
        <RouterLink v-if="user?.role === 'admin'" to="/admin">后台</RouterLink>

        <div
          v-if="user"
          class="user-menu"
          @mouseenter="userMenuOpen = true"
          @mouseleave="userMenuOpen = false"
        >
          <button class="avatar-trigger" type="button" @click="openProfileEditor">
            <img class="topbar-avatar" :src="user.avatarUrl || defaultAvatar" :alt="user.name" />
          </button>

          <div v-if="userMenuOpen" class="user-popover">
            <img class="user-popover-avatar" :src="user.avatarUrl || defaultAvatar" :alt="user.name" />
            <div class="user-popover-body">
              <strong>{{ user.name }}</strong>
              <p>{{ user.email }}</p>
              <span>{{ user.role === 'admin' ? '管理员' : '普通用户' }}</span>
              <div class="user-popover-actions">
                <button type="button" class="ghost" @click="openProfileEditor">编辑资料</button>
                <button type="button" class="ghost danger" @click="logout">退出登录</button>
              </div>
            </div>
          </div>
        </div>

        <RouterLink v-else to="/auth">登录</RouterLink>
      </nav>
    </header>

    <main class="content" :class="{ 'article-content': isArticlePage }">
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
import SiteLogo from './components/SiteLogo.vue'
import { api } from './api'
import type { User } from './types'

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
const isArticlePage = computed(() => String(route.path || '').startsWith('/article/'))
const userMenuOpen = ref(false)
const profileOpen = ref(false)
const profileName = ref('')
const profileError = ref('')
const profileNotice = ref('')
const profileAvatarFile = ref<File | null>(null)
const profileAvatarPreview = ref('')
let objectUrl = ''

const syncUser = async () => {
  const token = localStorage.getItem('blog_token')
  if (!token) {
    user.value = null
    return
  }
  try {
    const res = await api.me()
    user.value = res.user
  } catch {
    api.clearToken()
    user.value = null
  }
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

function openProfileEditor() {
  if (!user.value) return
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

onMounted(async () => {
  await syncUser()
  window.addEventListener('blog-auth-changed', syncUser)
  window.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  window.removeEventListener('blog-auth-changed', syncUser)
  window.removeEventListener('keydown', handleEscape)
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl)
  }
})

function logout() {
  api.clearToken()
  user.value = null
  window.location.href = '/'
}
</script>
