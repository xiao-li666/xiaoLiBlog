<template>
  <section class="auth-page">
    <canvas ref="meteorCanvas" class="auth-canvas"></canvas>

    <div class="auth-card">
      <div class="auth-card-title">欢迎回来</div>

      <div class="auth-tabs">
        <button type="button" :class="{ active: mode === 'login' }" @click="switchMode('login')">
          <LogInIcon :size="16" />
          <span>登录</span>
        </button>
        <button type="button" :class="{ active: mode === 'register' }" @click="switchMode('register')">
          <UserPlusIcon :size="16" />
          <span>注册</span>
        </button>
      </div>

      <form class="auth-form" @submit.prevent="submit">
        <div v-if="mode === 'register'" class="auth-field-row">
          <span class="auth-prefix">用户名</span>
          <input v-model.trim="name" class="auth-input" type="text" autocomplete="name" placeholder="请输入昵称" />
        </div>

        <div class="auth-field-row">
          <span class="auth-prefix">邮箱</span>
          <input v-model.trim="email" class="auth-input" type="email" autocomplete="email" placeholder="name@example.com" />
        </div>

        <div class="auth-field-row">
          <span class="auth-prefix">密码</span>
          <input
            v-model="password"
            class="auth-input"
            type="password"
            :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
            placeholder="至少 6 位"
          />
        </div>

        <div v-if="mode === 'register'" class="auth-field-row">
          <span class="auth-prefix">确认密码</span>
          <input v-model="confirmPassword" class="auth-input" type="password" autocomplete="new-password" placeholder="再次输入密码" />
        </div>

        <div class="auth-field-row auth-code-row">
          <span class="auth-prefix">验证码</span>
          <input
            v-model.trim="verificationCode"
            class="auth-input"
            type="text"
            inputmode="numeric"
            autocomplete="one-time-code"
            placeholder="输入验证码"
          />
          <button type="button" class="auth-code-btn" :disabled="codeSending || codeCountdown > 0 || !email.trim()" @click="sendCode">
            {{ codeButtonText }}
          </button>
        </div>

        <button class="auth-btn" :disabled="submitting">
          <span v-if="submitting" class="auth-spinner"></span>
          <span>{{ submitting ? '处理中...' : mode === 'login' ? '登录' : '注册' }}</span>
        </button>

        <p class="auth-message success" v-if="message">{{ message }}</p>
        <p class="auth-message error" v-if="error">{{ error }}</p>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LogInIcon, UserPlusIcon } from 'lucide-vue-next'
import { api } from '../api'

type AuthMode = 'login' | 'register'

const router = useRouter()
const route = useRoute()
const mode = ref<AuthMode>('login')
const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const verificationCode = ref('')
const error = ref('')
const message = ref('')
const submitting = ref(false)
const codeSending = ref(false)
const codeCountdown = ref(0)
const codeButtonText = computed(() => {
  if (codeSending.value) return '发送中'
  if (codeCountdown.value > 0) return `${codeCountdown.value}s`
  return '发送验证码'
})

const meteorCanvas = ref<HTMLCanvasElement | null>(null)
let animId = 0
let pageVisible = true
let lastSpawn = 0
let codeTimer: ReturnType<typeof window.setInterval> | null = null

const W = () => window.innerWidth
const H = () => window.innerHeight

class Meteor {
  x = 0
  y = 0
  angle = 0
  speed = 0
  tailLen = 0
  alpha = 0
  warmShift = 0
  life = 0

  constructor() {
    this.x = Math.random() * W() * 0.9
    this.y = -(20 + Math.random() * 140)
    this.angle = Math.PI / 4 + (Math.random() - 0.5) * 0.8
    this.speed = 10 + Math.random() * 14
    this.tailLen = 80 + Math.random() * 120
    this.alpha = 0.5 + Math.random() * 0.5
    this.warmShift = Math.random() * 0.12
  }

  update() {
    this.life += 1
    this.x += Math.cos(this.angle) * this.speed
    this.y += Math.sin(this.angle) * this.speed
    return this.y > H() + 250 || this.x > W() + 250 || this.x < -250 || this.life > 180
  }

  draw(ctx: CanvasRenderingContext2D) {
    const endX = this.x - Math.cos(this.angle) * this.tailLen
    const endY = this.y - Math.sin(this.angle) * this.tailLen
    const gradient = ctx.createLinearGradient(this.x, this.y, endX, endY)
    gradient.addColorStop(0, `rgba(${255 - this.warmShift * 80},${255 - this.warmShift * 40},255,${this.alpha})`)
    gradient.addColorStop(1, 'rgba(255,255,255,0)')

    ctx.save()
    ctx.globalAlpha = this.alpha
    ctx.beginPath()
    ctx.moveTo(this.x, this.y)
    ctx.lineTo(endX, endY)
    ctx.strokeStyle = gradient
    ctx.lineWidth = 1.6
    ctx.lineCap = 'round'
    ctx.stroke()
    ctx.restore()

    ctx.save()
    ctx.beginPath()
    ctx.arc(this.x, this.y, 2.2, 0, Math.PI * 2)
    ctx.fillStyle = '#fff'
    ctx.shadowColor = '#fff'
    ctx.shadowBlur = 6 + Math.random() * 4
    ctx.fill()
    ctx.restore()
  }
}

const stars: Array<{ x: number; y: number; r: number; alpha: number; ts: number; to: number }> = []
const meteors: Meteor[] = []

function buildStars() {
  stars.length = 0
  for (let i = 0; i < 200; i += 1) {
    stars.push({
      x: Math.random() * W(),
      y: Math.random() * H(),
      r: 0.4 + Math.random() * 1.4,
      alpha: 0.15 + Math.random() * 0.55,
      ts: 0.5 + Math.random() * 2.5,
      to: Math.random() * Math.PI * 2,
    })
  }
}

function spawnMeteors() {
  const count = 2 + Math.floor(Math.random() * 2)
  for (let i = 0; i < count; i += 1) {
    meteors.push(new Meteor())
  }
}

function handleVisibility() {
  pageVisible = !document.hidden
}

function handleResize() {
  const canvas = meteorCanvas.value
  if (!canvas) return
  const dpr = window.devicePixelRatio || 1
  canvas.width = W() * dpr
  canvas.height = H() * dpr
  canvas.style.width = `${W()}px`
  canvas.style.height = `${H()}px`
  const ctx = canvas.getContext('2d')
  if (ctx) {
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.scale(dpr, dpr)
  }
  stars.forEach((star) => {
    star.x = Math.random() * W()
    star.y = Math.random() * H()
  })
}

function startAnimation() {
  const canvas = meteorCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  function frame(timestamp: number) {
    if (!pageVisible) {
      animId = requestAnimationFrame(frame)
      return
    }
    const time = timestamp * 0.001
    ctx!.fillStyle = '#000'
    ctx!.fillRect(0, 0, W(), H())

    for (const star of stars) {
      const twinkle = Math.sin(time * star.ts + star.to) * 0.3 + 0.7
      ctx!.beginPath()
      ctx!.arc(star.x, star.y, star.r, 0, Math.PI * 2)
      ctx!.fillStyle = `rgba(255,255,255,${star.alpha * twinkle})`
      ctx!.fill()
    }

    if (timestamp - lastSpawn > 1800 + Math.random() * 600) {
      spawnMeteors()
      lastSpawn = timestamp
    }

    for (let i = meteors.length - 1; i >= 0; i -= 1) {
      if (meteors[i].update()) {
        meteors.splice(i, 1)
      } else {
        meteors[i].draw(ctx!)
      }
    }
    animId = requestAnimationFrame(frame)
  }
  animId = requestAnimationFrame(frame)
}

function switchMode(nextMode: AuthMode) {
  mode.value = nextMode
  error.value = ''
  message.value = ''
  verificationCode.value = ''
  confirmPassword.value = ''
  resetCodeCountdown()
}

function validateEmail() {
  if (!email.value) return '请输入邮箱'
  if (!/^\S+@\S+\.\S+$/.test(email.value)) return '邮箱格式不正确'
  return ''
}

async function sendCode() {
  error.value = ''
  message.value = ''
  const emailError = validateEmail()
  if (emailError) {
    error.value = emailError
    return
  }
  codeSending.value = true
  try {
    await api.requestVerificationCode({ email: email.value, purpose: mode.value })
    message.value = '验证码已发送，请查看邮箱'
    startCodeCountdown()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    codeSending.value = false
  }
}

function validate() {
  if (mode.value === 'register' && !name.value) return '请输入用户名'
  const emailError = validateEmail()
  if (emailError) return emailError
  if (password.value.length < 6) return '密码至少需要 6 位'
  if (mode.value === 'register' && password.value !== confirmPassword.value) return '两次密码不一致'
  if (!verificationCode.value) return '请输入验证码'
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
        ? await api.login({ email: email.value, password: password.value, verificationCode: verificationCode.value })
        : await api.register({
            name: name.value,
            email: email.value,
            password: password.value,
            confirmPassword: confirmPassword.value,
            verificationCode: verificationCode.value,
          })
    api.setToken(res.token)
    const redirect = typeof route.query.redirect === 'string' && route.query.redirect.trim() ? route.query.redirect : '/'
    await router.push(redirect)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}

function startCodeCountdown() {
  resetCodeCountdown()
  codeCountdown.value = 60
  codeTimer = window.setInterval(() => {
    codeCountdown.value -= 1
    if (codeCountdown.value <= 0) resetCodeCountdown()
  }, 1000)
}

function resetCodeCountdown() {
  if (codeTimer) {
    window.clearInterval(codeTimer)
    codeTimer = null
  }
  codeCountdown.value = 0
}

onMounted(() => {
  buildStars()
  handleResize()
  startAnimation()
  window.addEventListener('resize', handleResize)
  document.addEventListener('visibilitychange', handleVisibility)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(animId)
  resetCodeCountdown()
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('visibilitychange', handleVisibility)
})
</script>
