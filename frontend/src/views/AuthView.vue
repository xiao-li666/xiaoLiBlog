<template>
  <section class="auth-page">
    <div class="auth-intro">
      <p class="eyebrow">账户中心</p>
      <h1>登录后继续阅读与写作</h1>
      <p>登录后可以发表评论、点赞和收藏；注册后自动进入首页，站长则可从导航进入后台。</p>
      <div class="auth-note">
        <span>简洁</span>
        <span>安全</span>
        <span>统一入口</span>
      </div>
    </div>

    <div class="auth-card">
      <div class="auth-tabs">
        <button type="button" :class="{ active: mode === 'login' }" @click="switchMode('login')">登录</button>
        <button type="button" :class="{ active: mode === 'register' }" @click="switchMode('register')">注册</button>
      </div>

      <form class="auth-form" @submit.prevent="submit">
        <label v-if="mode === 'register'">
          <span class="label-text">昵称</span>
          <div class="input-icon">
            <UserIcon :size="18" />
            <input v-model.trim="name" autocomplete="name" placeholder="请输入昵称" />
          </div>
        </label>

        <label>
          <span class="label-text">邮箱</span>
          <div class="input-icon">
            <MailIcon :size="18" />
            <input v-model.trim="email" type="email" autocomplete="email" placeholder="name@example.com" />
          </div>
        </label>

        <label>
          <span class="label-text">密码</span>
          <div class="input-icon">
            <LockIcon :size="18" />
            <input v-model="password" type="password" autocomplete="current-password" placeholder="至少 6 位" />
          </div>
        </label>

        <button class="auth-submit" :disabled="submitting">
          {{ submitting ? '提交中...' : mode === 'login' ? '登录' : '注册并进入首页' }}
        </button>

        <p class="auth-message success" v-if="message">{{ message }}</p>
        <p class="auth-message error" v-if="error">{{ error }}</p>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { LockIcon, MailIcon, UserIcon } from 'lucide-vue-next'
import { api } from '../api'

type AuthMode = 'login' | 'register'

const router = useRouter()
const mode = ref<AuthMode>('login')
const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const message = ref('')
const submitting = ref(false)

function switchMode(nextMode: AuthMode) {
  mode.value = nextMode
  error.value = ''
  message.value = ''
}

function validate() {
  if (mode.value === 'register' && !name.value) return '请输入昵称'
  if (!email.value) return '请输入邮箱'
  if (!/^\S+@\S+\.\S+$/.test(email.value)) return '邮箱格式不正确'
  if (password.value.length < 6) return '密码至少需要 6 位'
  return ''
}

async function submit() {
  error.value = ''
  message.value = ''
  const validationError = validate()
  if (validationError) {
    error.value = validationError
    return
  }

  submitting.value = true
  try {
    const res =
      mode.value === 'login'
        ? await api.login({ email: email.value, password: password.value })
        : await api.register({ name: name.value, email: email.value, password: password.value })

    api.setToken(res.token)
    message.value = mode.value === 'login' ? '登录成功' : '注册成功'
    await router.push('/')
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}
</script>
