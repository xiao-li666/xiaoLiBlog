<template>
  <section class="admin-shell" v-if="ready">
    <aside class="admin-sidebar">
      <div class="admin-brand">
        <p class="eyebrow">Backend</p>
        <h2>xiaoli博客后台</h2>
      </div>

      <div class="admin-menu">
        <button
          v-for="item in adminMenu"
          :key="item.key"
          :class="['admin-menu-item', { active: activeSection === item.key }]"
          @click="setActiveSection(item.key)"
        >
          <component :is="menuIcons[item.key]" :size="16" />
          <span>{{ item.label }}</span>
          <span v-if="notificationBadgeMap[item.key] > 0" class="admin-menu-badge">
            {{ notificationBadgeMap[item.key] > 99 ? '99+' : notificationBadgeMap[item.key] }}
          </span>
        </button>
      </div>

      <div class="admin-profile">
        <div class="profile-avatar" :class="{ 'has-image': !!me?.avatarUrl }">
          <img v-if="me?.avatarUrl" :src="me.avatarUrl" :alt="me.name" />
          <span v-else>{{ me?.name?.slice(0, 1) || 'A' }}</span>
        </div>
        <div>
          <strong>{{ me?.name }}</strong>
          <p>{{ me?.email }}</p>
        </div>
      </div>
    </aside>

    <main class="admin-main">
      <header class="admin-hero">
        <div>
          <p class="eyebrow">Admin Console</p>
          <h1>{{ activeSectionCopy.title }}</h1>

        </div>
        <div class="admin-hero-actions">
          <button @click="refreshAll"><RefreshCwIcon :size="15" /> 刷新数据</button>
          <button class="ghost" @click="startCreateArticle"><PlusCircleIcon :size="15" /> 新建文章</button>
        </div>
      </header>

      <section class="stat-grid" v-if="activeSection === 'overview'">
        <div class="stat-card" v-for="item in statCards" :key="item.label">
          <span class="stat-label">{{ item.label }}</span>
          <strong class="stat-value">{{ item.value }}</strong>
          <span class="stat-desc">{{ item.desc }}</span>
        </div>
      </section>

      <section class="dashboard-chart-grid" v-if="activeSection === 'overview'">
        <article class="dashboard-chart-card">
          <div class="chart-head">
            <div>
              <p class="eyebrow">Publish</p>
              <h3>文章发布趋势</h3>
            </div>
            <span class="pill muted">近 6 个月</span>
          </div>
          <div class="line-chart neon-chart" :class="{ empty: isChartEmpty(articlePublishTrend) }">
            <svg viewBox="0 0 320 140" role="img" aria-label="文章发布趋势">
              <defs>
                <linearGradient id="publishTrendArea" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#22d3ee" stop-opacity="0.28" />
                  <stop offset="100%" stop-color="#22d3ee" stop-opacity="0.02" />
                </linearGradient>
                <linearGradient id="publishTrendLine" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#22d3ee" />
                  <stop offset="100%" stop-color="#38bdf8" />
                </linearGradient>
              </defs>
              <polyline class="line-grid" points="0,110 320,110" />
              <polyline class="line-grid" points="0,70 320,70" />
              <polyline class="line-grid" points="0,30 320,30" />
              <polygon class="line-area publish" :points="lineAreaPoints(articlePublishTrend)" />
              <polyline class="line-path publish" :points="linePoints(articlePublishTrend)" />
              <circle
                v-for="point in linePointItems(articlePublishTrend)"
                :key="point.label"
                class="line-dot publish"
                :cx="point.x"
                :cy="point.y"
                r="3.5"
              />
            </svg>
            <div class="chart-axis">
              <span v-for="item in articlePublishTrend" :key="item.label">{{ item.label }}</span>
            </div>
            <div class="chart-footnote">发布趋势按月统计，折线越亮代表当月发布越活跃。</div>
          </div>
        </article>

        <article class="dashboard-chart-card">
          <div class="chart-head">
            <div>
              <p class="eyebrow">Status</p>
              <h3>文章状态分布</h3>
            </div>
            <span class="pill muted">{{ stats?.articles.total ?? 0 }} 篇</span>
          </div>
          <div class="donut-chart-wrap">
            <div class="donut-chart" :style="articleStatusDonutStyle">
              <span>{{ publishedPercent }}%</span>
            </div>
            <div class="legend-list">
              <div class="legend-item">
                <span class="legend-dot published"></span>
                <strong>已发布</strong>
                <em>{{ stats?.articles.published ?? 0 }}</em>
              </div>
              <div class="legend-item">
                <span class="legend-dot draft"></span>
                <strong>草稿</strong>
                <em>{{ stats?.articles.draft ?? 0 }}</em>
              </div>
            </div>
          </div>
        </article>

        <article class="dashboard-chart-card">
          <div class="chart-head">
            <div>
              <p class="eyebrow">Categories</p>
              <h3>分类文章分布</h3>
            </div>
            <span class="pill muted">{{ categories.length }} 类</span>
          </div>
          <div class="rank-bars">
            <div v-for="item in categoryDistribution" :key="item.label" class="rank-bar-row">
              <span class="rank-bar-label">{{ item.label }}</span>
              <div class="rank-bar-track">
                <span :style="{ width: barPercent(item.value, categoryDistribution) + '%' }"></span>
              </div>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
        </article>

        <article class="dashboard-chart-card">
          <div class="chart-head">
            <div>
              <p class="eyebrow">Comments</p>
              <h3>评论活跃趋势</h3>
            </div>
            <span class="pill muted">近 7 天</span>
          </div>
          <div class="bar-chart compact-bars" :class="{ empty: isChartEmpty(commentActivityTrend) }">
            <div v-for="item in commentActivityTrend" :key="item.label" class="bar-column">
              <span class="bar-value">{{ item.value }}</span>
              <div class="bar-track">
                <span class="bar-fill comment" :style="{ height: barPercent(item.value, commentActivityTrend) + '%' }"></span>
              </div>
              <span class="bar-label">{{ item.label }}</span>
            </div>
          </div>
        </article>

        <article class="dashboard-chart-card">
          <div class="chart-head">
            <div>
              <p class="eyebrow">Popular</p>
              <h3>热门文章排行</h3>
            </div>
            <span class="pill muted">互动量</span>
          </div>
          <div class="rank-bars">
            <div v-for="item in popularArticleChart" :key="item.label" class="rank-bar-row">
              <span class="rank-bar-label">{{ item.label }}</span>
              <div class="rank-bar-track hot">
                <span :style="{ width: barPercent(item.value, popularArticleChart) + '%' }"></span>
              </div>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
        </article>

        <article class="dashboard-chart-card">
          <div class="chart-head">
            <div>
              <p class="eyebrow">Users</p>
              <h3>用户增长趋势</h3>
            </div>
            <span class="pill muted">近 6 个月</span>
          </div>
          <div class="line-chart" :class="{ empty: isChartEmpty(userGrowthTrend) }">
            <svg viewBox="0 0 320 140" role="img" aria-label="用户增长趋势">
              <defs>
                <linearGradient id="userTrendArea" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.42" />
                  <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.02" />
                </linearGradient>
                <linearGradient id="userTrendLine" x1="0" y1="0" x2="1" y2="0">
                  <stop offset="0%" stop-color="#7dd3fc" />
                  <stop offset="50%" stop-color="#60a5fa" />
                  <stop offset="100%" stop-color="#a78bfa" />
                </linearGradient>
              </defs>
              <polyline class="line-grid" points="0,110 320,110" />
              <polyline class="line-grid" points="0,70 320,70" />
              <polyline class="line-grid" points="0,30 320,30" />
              <polygon class="line-area" :points="lineAreaPoints(userGrowthTrend)" />
              <polyline class="line-path" :points="linePoints(userGrowthTrend)" />
              <circle
                v-for="point in linePointItems(userGrowthTrend)"
                :key="point.label"
                class="line-dot"
                :cx="point.x"
                :cy="point.y"
                r="4"
              />
            </svg>
            <div class="chart-axis">
              <span v-for="item in userGrowthTrend" :key="item.label">{{ item.label }}</span>
            </div>
            <div class="chart-footnote">用户注册按月统计，趋势线展示近 6 个月的增量变化。</div>
          </div>
        </article>
      </section>

      <section class="admin-card" v-if="activeSection === 'articles'">
        <div class="card-head">
          <div>
            <p class="eyebrow">文章列表</p>
            <h3>内容管理</h3>
          </div>
          <div class="inline-form compact">
            <input v-model="articleQuery" placeholder="搜索文章" @keyup.enter="loadArticles" />
            <select v-model="articleStatus" @change="loadArticles">
              <option value="">全部状态</option>
              <option value="draft">草稿</option>
              <option value="published">已发布</option>
            </select>
            <button @click="loadArticles"><SearchIcon :size="14" /> 筛选</button>
          </div>
        </div>

        <div class="table-list">
          <div class="table-head">
            <span>标题</span>
            <span>分类</span>
            <span>状态</span>
            <span>点赞 / 收藏</span>
            <span>操作</span>
          </div>
          <div v-for="item in articles" :key="item.id" class="table-row">
            <div>
              <strong>{{ item.title }}</strong>
              <p>{{ item.summary || articleSummary(item) }}</p>
            </div>
            <span>{{ item.category?.name || '未分类' }}</span>
            <span
              :class="['article-status-icon', item.status]"
              :title="item.status === 'published' ? '已发布' : '草稿'"
              :aria-label="item.status === 'published' ? '已发布' : '草稿'"
            >
              <component :is="articleStatusIcon(item.status)" :size="16" />
            </span>
            <div class="reaction-stats">
              <span class="reaction-pill like">赞 {{ item.likesCount }}</span>
              <span class="reaction-pill favorite">藏 {{ item.favoritesCount }}</span>
            </div>
            <div class="actions-row small">
              <button class="ghost" @click="editArticle(item)"><PencilIcon :size="13" /> 编辑</button>
              <button class="ghost danger" @click="removeArticle(item.id)"><Trash2Icon :size="13" /> 删除</button>
            </div>
          </div>
        </div>
      </section>

      <section class="admin-card" v-if="activeSection === 'createArticle'">
        <div class="card-head">
          <div>
            <p class="eyebrow">新建文章</p>
            <h3>{{ articleForm.id ? '编辑文章' : '新建文章' }}</h3>
          </div>
          <span class="pill">{{ articleForm.status === 'published' ? '已发布' : '草稿' }}</span>
        </div>

        <div class="form-grid">
          <input v-model="articleForm.coverUrl" placeholder="封面图链接" />
          <input v-model="articleForm.title" placeholder="标题" />
          <input v-model="articleForm.summary" placeholder="摘要" />
          <input v-model="articleForm.tags" placeholder="标签，逗号分隔" />
          <select v-model.number="articleForm.categoryId">
            <option :value="0">未分类</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
          <select v-model="articleForm.status">
            <option value="draft">草稿</option>
            <option value="published">发布</option>
          </select>
        </div>

        <div class="upload-grid">
          <label class="upload-box">
            <span>上传封面图</span>
            <input type="file" accept="image/*" @change="uploadCoverImage" />
          </label>
          <label class="upload-box">
            <span>上传 Markdown</span>
            <input type="file" accept=".md,.markdown,text/markdown" @change="uploadMarkdownFile" />
          </label>
        </div>

        <ArticleTiptapEditor ref="articleEditorRef" :markdown="articleForm.content" />

        <div class="actions-row">
          <button @click="saveArticle"><SaveIcon :size="15" /> 保存文章</button>
          <button class="ghost" @click="resetArticleForm"><RotateCcwIcon :size="15" /> 清空</button>
        </div>
        <p class="feedback error" v-if="error">{{ error }}</p>
        <p class="feedback success" v-if="notice">{{ notice }}</p>
      </section>

      <section class="admin-grid" :class="{ 'single-panel': true }" v-if="activeSection === 'categories' || activeSection === 'users'">
        <div class="admin-card admin-sidecards">
          <div class="mini-card" v-if="activeSection === 'categories'">
            <div class="card-head">
              <div>
                <p class="eyebrow">分类管理</p>
                <h3>文章分类</h3>
              </div>
            </div>
            <div class="inline-form">
              <input v-model="categoryName" placeholder="分类名称" />
              <button @click="saveCategory"><PlusCircleIcon :size="14" /> 新增</button>
            </div>
            <div class="small-list">
              <div v-for="cat in categories" :key="cat.id" class="small-row">
                <div>
                  <strong>{{ cat.name }}</strong>
                  <p>{{ cat.slug }}</p>
                </div>
                <button class="ghost" @click="removeCategory(cat.id)"><Trash2Icon :size="13" /> 删除</button>
              </div>
            </div>
          </div>

          <div class="mini-card" v-if="activeSection === 'users'">
            <div class="card-head">
              <div>
                <p class="eyebrow">用户管理</p>
                <h3>注册用户</h3>
              </div>
            </div>
            <div class="small-list">
              <div v-for="user in users" :key="user.id" class="small-row">
                <div class="user-row-main">
                  <img class="user-row-avatar" :src="user.avatarUrl || defaultAvatar" :alt="user.name" />
                  <div class="user-row-text">
                    <strong>{{ user.name }}</strong>
                    <p>{{ user.email }}</p>
                  </div>
                </div>
                <span class="pill muted">{{ user.role }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="admin-card" v-if="activeSection === 'comments'">
        <div class="card-head">
          <div>
            <p class="eyebrow">评论审核</p>
            <h3>互动管理</h3>
          </div>
          <span class="pill muted">{{ commentGroups.length }} 篇文章</span>
        </div>
        <div class="comment-group-list">
          <details
            v-for="group in commentGroups"
            :key="group.articleId"
            class="comment-group-card"
            @toggle="onCommentGroupToggle(group, $event)"
          >
            <summary class="comment-group-head">
              <div>
                <p class="eyebrow">文章 #{{ group.articleId }}</p>
                <h4>
                  <RouterLink v-if="group.slug" :to="`/article/${group.slug}`">{{ group.title }}</RouterLink>
                  <span v-else>{{ group.title }}</span>
                </h4>
              </div>
              <div class="comment-group-head-side">
                <span v-if="commentUnreadCount(group.articleId) > 0" class="admin-menu-badge compact">
                  {{ commentUnreadCount(group.articleId) > 99 ? '99+' : commentUnreadCount(group.articleId) }}
                </span>
                <span class="pill muted">{{ group.comments.length }} 条</span>
                <ChevronDownIcon class="comment-group-chevron" :size="18" />
              </div>
            </summary>
            <div class="comment-group-body">
              <article v-for="item in group.comments" :key="item.id" class="comment-row">
                <div class="comment-avatar">
                  <img :src="item.user.avatarUrl || defaultAvatar" :alt="item.user.name" />
                </div>
                <div class="comment-row-main">
                  <div class="comment-row-meta">
                    <strong>{{ item.user.name }}</strong>
                    <span>{{ item.user.email }}</span>
                    <span>·</span>
                    <span>{{ item.status === 'published' ? '已显示' : '已隐藏' }}</span>
                  </div>
                  <p>{{ item.body }}</p>
                </div>
                <div class="actions-row small comment-row-actions">
                  <button class="ghost" @click="setCommentStatus(item.id, 'published')"><EyeIcon :size="13" /> 显示</button>
                  <button class="ghost danger" @click="setCommentStatus(item.id, 'hidden')"><EyeOffIcon :size="13" /> 隐藏</button>
                </div>
              </article>
            </div>
          </details>
        </div>
      </section>
    </main>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  BoldIcon, ChevronDownIcon, Code2Icon, CircleCheckIcon, EyeIcon, EyeOffIcon, FilePenLineIcon, FileTextIcon, FolderOpenIcon,
  Heading2Icon, LayoutDashboardIcon, LinkIcon, ListIcon,
  MessageSquareIcon, PencilIcon, PlusCircleIcon, QuoteIcon,
  RefreshCwIcon, RotateCcwIcon, SaveIcon, SearchIcon,
  Trash2Icon, UsersIcon,
} from 'lucide-vue-next'
import { api } from '../api'
import ArticleTiptapEditor from '../components/ArticleTiptapEditor.vue'
import type { AdminNotificationResponse, AdminStats, Article, Category, Comment, NotificationCounts, NotificationType, User } from '../types'

const defaultAvatar =
  'data:image/svg+xml;charset=UTF-8,' +
  encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect width="96" height="96" rx="24" fill="#30363d"/>
      <circle cx="48" cy="38" r="17" fill="#8b949e"/>
      <path d="M20 80c4-17 19-26 28-26s24 9 28 26" fill="#8b949e"/>
    </svg>
  `)

type AdminSection = 'overview' | 'articles' | 'createArticle' | 'categories' | 'comments' | 'users'

const router = useRouter()
const ready = ref(false)
const me = ref<User | null>(null)
const stats = ref<AdminStats | null>(null)
const articles = ref<Article[]>([])
const categories = ref<Category[]>([])
const comments = ref<Comment[]>([])
const users = ref<User[]>([])
const notifications = ref<AdminNotificationResponse | null>(null)
const categoryName = ref('')
const articleQuery = ref('')
const articleStatus = ref('')
const activeSection = ref<AdminSection>('overview')
const articleEditorRef = ref<{ getMarkdown?: () => string; setMarkdown?: (markdown: string) => void } | null>(null)
const markdownTextarea = ref<HTMLTextAreaElement | null>(null)
const markdownPreviewRef = ref<HTMLDivElement | null>(null)
const error = ref('')
const notice = ref('')
const articleForm = reactive({
  id: 0,
  coverUrl: '',
  title: '',
  summary: '',
  tags: '',
  content: '',
  status: 'draft' as 'draft' | 'published',
  categoryId: 0,
})

const notificationBadgeCounts = computed<NotificationCounts>(() => ({
  register: notifications.value?.counts.register ?? 0,
  comment: notifications.value?.counts.comment ?? 0,
  like: notifications.value?.counts.like ?? 0,
  favorite: notifications.value?.counts.favorite ?? 0,
}))

const notificationBadgeMap = computed<Record<AdminSection, number>>(() => ({
  overview: notifications.value?.unreadTotal ?? 0,
  articles: (notificationBadgeCounts.value.like ?? 0) + (notificationBadgeCounts.value.favorite ?? 0),
  createArticle: 0,
  categories: 0,
  comments: notificationBadgeCounts.value.comment ?? 0,
  users: notificationBadgeCounts.value.register ?? 0,
}))

const statCards = computed(() => [
  { label: '文章总数', value: stats.value?.articles.total ?? 0, desc: '含草稿与已发布' },
  { label: '已发布', value: stats.value?.articles.published ?? 0, desc: '对访客可见' },
  { label: '草稿', value: stats.value?.articles.draft ?? 0, desc: '待处理内容' },
  { label: '分类', value: stats.value?.categories ?? 0, desc: '内容结构' },
  { label: '评论', value: stats.value?.comments ?? 0, desc: '互动记录' },
  { label: '用户', value: stats.value?.users ?? 0, desc: '注册用户' },
])

type ChartItem = {
  label: string
  key?: string
  value: number
}

const articlePublishTrend = computed<ChartItem[]>(() => {
  const months = recentMonths(6)
  const counts = new Map(months.map((item) => [item.key, 0]))
  articles.value.forEach((article) => {
    if (article.status !== 'published' || !article.publishedAt) return
    const key = monthKey(new Date(article.publishedAt))
    if (counts.has(key)) counts.set(key, (counts.get(key) ?? 0) + 1)
  })
  return months.map((item) => ({ label: item.label, key: item.key, value: counts.get(item.key) ?? 0 }))
})

const publishedPercent = computed(() => {
  const total = stats.value?.articles.total ?? 0
  if (!total) return 0
  return Math.round(((stats.value?.articles.published ?? 0) / total) * 100)
})

const articleStatusDonutStyle = computed(() => ({
  '--published-angle': `${publishedPercent.value * 3.6}deg`,
}))

const categoryDistribution = computed<ChartItem[]>(() => {
  const counts = new Map<string, number>()
  categories.value.forEach((category) => counts.set(category.name, 0))
  articles.value.forEach((article) => {
    const name = article.category?.name || '未分类'
    counts.set(name, (counts.get(name) ?? 0) + 1)
  })
  const rows = Array.from(counts, ([label, value]) => ({ label, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 6)
  return rows.length ? rows : [{ label: '暂无分类', value: 0 }]
})

const commentActivityTrend = computed<ChartItem[]>(() => {
  const days = recentDays(7)
  const counts = new Map(days.map((item) => [item.key, 0]))
  comments.value.forEach((comment) => {
    if (!comment.createdAt) return
    const key = dateKey(new Date(comment.createdAt))
    if (counts.has(key)) counts.set(key, (counts.get(key) ?? 0) + 1)
  })
  return days.map((item) => ({ label: item.label, key: item.key, value: counts.get(item.key) ?? 0 }))
})

const popularArticleChart = computed<ChartItem[]>(() => {
  const rows = articles.value
    .map((article) => ({
      label: article.title,
      value: (article.likesCount ?? 0) + (article.favoritesCount ?? 0),
    }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 5)
  return rows.length ? rows : [{ label: '暂无文章', value: 0 }]
})

const commentGroups = computed(() => {
  const groups = new Map<number, { articleId: number; title: string; slug: string; comments: Comment[] }>()
  comments.value.forEach((item) => {
    const articleId = item.articleId
    const article = item.article
    const title = article?.title || `文章 #${articleId}`
    const slug = article?.slug || ''
    if (!groups.has(articleId)) {
      groups.set(articleId, { articleId, title, slug, comments: [] })
    }
    groups.get(articleId)!.comments.push(item)
  })
  return Array.from(groups.values()).sort((a, b) => b.comments.length - a.comments.length || a.title.localeCompare(b.title, 'zh-Hans-CN'))
})

const unreadCommentCountsByArticle = computed(() => {
  const counts = new Map<number, number>()
  notifications.value?.items.forEach((item) => {
    if (item.type !== 'comment' || item.isRead || !item.articleId) return
    counts.set(item.articleId, (counts.get(item.articleId) ?? 0) + 1)
  })
  return counts
})

const userGrowthTrend = computed<ChartItem[]>(() => {
  const months = recentMonths(6)
  const counts = new Map(months.map((item) => [item.key, 0]))
  users.value.forEach((user) => {
    if (!user.createdAt) return
    const key = monthKey(new Date(user.createdAt))
    if (counts.has(key)) counts.set(key, (counts.get(key) ?? 0) + 1)
  })
  return months.map((item) => ({ label: item.label, key: item.key, value: counts.get(item.key) ?? 0 }))
})

const menuIcons: Record<AdminSection, any> = {
  overview: LayoutDashboardIcon,
  articles: FileTextIcon,
  createArticle: PlusCircleIcon,
  categories: FolderOpenIcon,
  comments: MessageSquareIcon,
  users: UsersIcon,
}

const adminMenu: Array<{ key: AdminSection; label: string }> = [
  { key: 'overview', label: '概览' },
  { key: 'articles', label: '文章' },
  { key: 'createArticle', label: '新建文章' },
  { key: 'categories', label: '分类' },
  { key: 'comments', label: '评论' },
  { key: 'users', label: '用户' },
]

const activeSectionCopy = computed(() => {
  const map: Record<AdminSection, { title: string }> = {
    overview: { title: 'xiaoli博客管理台' },
    articles: { title: '文章管理' },
    createArticle: { title: articleForm.id ? '编辑文章' : '新建文章' },
    categories: { title: '分类管理' },
    comments: { title: '评论管理' },
    users: { title: '用户管理' },
  }
  return map[activeSection.value]
})

const markdownPreviewHtml = computed(() => renderMarkdownEditor(articleForm.content || ''))

function recentMonths(count: number): ChartItem[] {
  const now = new Date()
  return Array.from({ length: count }, (_, index) => {
    const date = new Date(now.getFullYear(), now.getMonth() - (count - 1 - index), 1)
    return {
      key: monthKey(date),
      label: `${date.getMonth() + 1}月`,
      value: 0,
    }
  })
}

function recentDays(count: number): ChartItem[] {
  const now = new Date()
  return Array.from({ length: count }, (_, index) => {
    const date = new Date(now)
    date.setDate(now.getDate() - (count - 1 - index))
    return {
      key: dateKey(date),
      label: `${date.getMonth() + 1}/${date.getDate()}`,
      value: 0,
    }
  })
}

function monthKey(date: Date) {
  if (Number.isNaN(date.getTime())) return ''
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
}

function dateKey(date: Date) {
  if (Number.isNaN(date.getTime())) return ''
  return `${monthKey(date)}-${String(date.getDate()).padStart(2, '0')}`
}

function chartMax(items: ChartItem[]) {
  return Math.max(1, ...items.map((item) => item.value))
}

function barPercent(value: number, items: ChartItem[]) {
  if (!value) return 0
  return Math.max(8, Math.round((value / chartMax(items)) * 100))
}

function isChartEmpty(items: ChartItem[]) {
  return items.every((item) => item.value === 0)
}

function linePointItems(items: ChartItem[]) {
  const max = chartMax(items)
  const step = items.length > 1 ? 320 / (items.length - 1) : 320
  return items.map((item, index) => ({
    label: item.label,
    x: Math.round(index * step),
    y: Math.round(118 - (item.value / max) * 92),
  }))
}

function linePoints(items: ChartItem[]) {
  return linePointItems(items).map((point) => `${point.x},${point.y}`).join(' ')
}

function lineAreaPoints(items: ChartItem[]) {
  const points = linePointItems(items)
  if (!points.length) return ''
  return `0,126 ${points.map((point) => `${point.x},${point.y}`).join(' ')} 320,126`
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function renderMarkdownEditor(source: string) {
  if (!source.trim()) {
    return '<div class="markdown-empty-state">在这里直接输入 Markdown，样式会实时呈现</div>'
  }

  const lines = source.replace(/\r\n/g, '\n').split('\n')
  const blocks: string[] = []
  let inCode = false

  for (const line of lines) {
    const fenceMatch = line.match(/^(\s*)```([\w+-]+)?\s*$/)
    if (fenceMatch) {
      if (!inCode) {
        inCode = true
        const lang = fenceMatch[2] || 'text'
        blocks.push(`<div class="md-line md-fence open" data-lang="${escapeHtml(lang)}"><span class="md-hidden">${escapeHtml(line)}</span></div>`)
      } else {
        inCode = false
        blocks.push(`<div class="md-line md-fence close"><span class="md-hidden">${escapeHtml(line)}</span></div>`)
      }
      continue
    }

    if (inCode) {
      blocks.push(`<div class="md-line md-code-line">${escapeHtml(line || ' ')}</div>`)
      continue
    }

    if (!line.trim()) {
      blocks.push('<div class="md-line md-empty-line">&nbsp;</div>')
      continue
    }

    const hrMatch = line.match(/^(\s*)([-*_])\2\2+(?:\s*)$/)
    if (hrMatch) {
      blocks.push(`<div class="md-line md-hr"><span class="md-hidden">${escapeHtml(line)}</span></div>`)
      continue
    }

    const headingMatch = line.match(/^(#{1,6})(\s+)(.*)$/)
    if (headingMatch) {
      const depth = headingMatch[1].length
      blocks.push(
        `<div class="md-line md-heading depth-${depth}"><span class="md-hidden">${escapeHtml(headingMatch[1])}</span>${escapeHtml(headingMatch[2])}${renderInlineMarkdown(headingMatch[3])}</div>`,
      )
      continue
    }

    const quoteMatch = line.match(/^(>\s+)(.*)$/)
    if (quoteMatch) {
      blocks.push(
        `<div class="md-line md-quote"><span class="md-hidden">${escapeHtml(quoteMatch[1])}</span>${renderInlineMarkdown(quoteMatch[2])}</div>`,
      )
      continue
    }

    const listMatch = line.match(/^([*+-]\s+)(.*)$/)
    if (listMatch) {
      blocks.push(
        `<div class="md-line md-list"><span class="md-hidden">${escapeHtml(listMatch[1])}</span>${renderInlineMarkdown(listMatch[2])}</div>`,
      )
      continue
    }

    blocks.push(`<div class="md-line">${renderInlineMarkdown(line)}</div>`)
  }

  return blocks.join('')
}

function renderInlineMarkdown(text: string) {
  let index = 0
  let html = ''

  while (index < text.length) {
    if (text.startsWith('**', index) || text.startsWith('__', index)) {
      const marker = text.slice(index, index + 2)
      const end = findClosing(text, marker, index + 2)
      if (end !== -1) {
        const inner = text.slice(index + 2, end)
        html += `<span class="md-strong"><span class="md-hidden">${escapeHtml(marker)}</span>${renderInlineMarkdown(inner)}<span class="md-hidden">${escapeHtml(marker)}</span></span>`
        index = end + 2
        continue
      }
    }

    const current = text[index]
    if (current === '*' || current === '_') {
      const end = findClosing(text, current, index + 1)
      if (end !== -1) {
        const inner = text.slice(index + 1, end)
        html += `<span class="md-em"><span class="md-hidden">${escapeHtml(current)}</span>${renderInlineMarkdown(inner)}<span class="md-hidden">${escapeHtml(current)}</span></span>`
        index = end + 1
        continue
      }
    }

    if (current === '`') {
      const end = text.indexOf('`', index + 1)
      if (end !== -1) {
        const inner = text.slice(index + 1, end)
        html += `<span class="md-code-span"><span class="md-hidden">\`</span>${escapeHtml(inner)}<span class="md-hidden">\`</span></span>`
        index = end + 1
        continue
      }
    }

    if (current === '[') {
      const closeBracket = text.indexOf(']', index + 1)
      if (closeBracket !== -1 && text[closeBracket + 1] === '(') {
        const closeParen = text.indexOf(')', closeBracket + 2)
        if (closeParen !== -1) {
          const label = text.slice(index + 1, closeBracket)
          const url = text.slice(closeBracket + 2, closeParen)
          html += [
            '<span class="md-link">',
            '<span class="md-hidden">[</span>',
            renderInlineMarkdown(label),
            '<span class="md-hidden">](',
            escapeHtml(url),
            ')</span>',
            '</span>',
          ].join('')
          index = closeParen + 1
          continue
        }
      }
    }

    html += escapeHtml(current)
    index += 1
  }

  return html
}

function findClosing(text: string, marker: string, start: number) {
  let index = start
  while (index < text.length) {
    const found = text.indexOf(marker, index)
    if (found === -1) return -1
    if (text[found - 1] !== '\\') return found
    index = found + marker.length
  }
  return -1
}

function setActiveSection(section: AdminSection) {
  activeSection.value = section
}

function startCreateArticle() {
  resetArticleForm()
  error.value = ''
  notice.value = ''
  activeSection.value = 'createArticle'
}

async function insertMarkdown(prefix: string, suffix = '', placeholder = '') {
  activeSection.value = 'createArticle'
  await nextTick()
  const textarea = markdownTextarea.value
  if (!textarea) return
  const value = articleForm.content || ''
  const start = textarea.selectionStart ?? value.length
  const end = textarea.selectionEnd ?? value.length
  const selected = value.slice(start, end) || placeholder
  const nextValue = `${value.slice(0, start)}${prefix}${selected}${suffix}${value.slice(end)}`
  articleForm.content = nextValue
  await nextTick()
  const cursor = start + prefix.length + selected.length + suffix.length
  textarea.focus()
  textarea.setSelectionRange(cursor, cursor)
}

function insertCodeBlock() {
  void insertMarkdown('```cpp\n', '\n```', 'int main() {\n  return 0;\n}')
}

function syncMarkdownScroll() {
  if (!markdownTextarea.value || !markdownPreviewRef.value) return
  markdownPreviewRef.value.scrollTop = markdownTextarea.value.scrollTop
  markdownPreviewRef.value.scrollLeft = markdownTextarea.value.scrollLeft
}

function handleMarkdownKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return
  const textarea = markdownTextarea.value
  if (!textarea) return
  const value = articleForm.content || ''
  const start = textarea.selectionStart ?? 0
  const end = textarea.selectionEnd ?? 0
  if (start !== end) return
  const lineStart = value.lastIndexOf('\n', start - 1) + 1
  const currentLine = value.slice(lineStart, start)
  const fenceMatch = currentLine.match(/^```([\w+-]+)?\s*$/)
  if (!fenceMatch) return
  event.preventDefault()
  const insert = '\n\n```'
  articleForm.content = `${value.slice(0, start)}${insert}${value.slice(end)}`
  void nextTick(() => {
    const caret = start + 1
    textarea.focus()
    textarea.setSelectionRange(caret, caret)
  })
}

function articleSummary(item: Article) {
  return item.summary || item.content.slice(0, 120).replace(/\s+/g, ' ').trim()
}

function articleStatusIcon(status: Article['status']) {
  return status === 'published' ? CircleCheckIcon : FilePenLineIcon
}

async function uploadCoverImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  notice.value = ''
  try {
    const res = await api.uploadAdminFile(file, 'image')
    if (res.url) {
      articleForm.coverUrl = res.url
      notice.value = '封面图已上传'
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    input.value = ''
  }
}

async function uploadMarkdownFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  notice.value = ''
  try {
    const res = await api.uploadAdminFile(file, 'markdown')
    if (res.content) {
      articleForm.content = res.content
      if (!articleForm.title && res.title) {
        articleForm.title = res.title
      }
      notice.value = 'Markdown 已导入到正文'
    }
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    input.value = ''
  }
}

async function load() {
  try {
    const meRes = await api.me()
    if (meRes.user.role !== 'admin') {
      router.push('/')
      return
    }
    me.value = meRes.user
    await refreshAll()
    ready.value = true
  } catch {
    router.push('/auth')
  }
}

async function syncMe() {
  try {
    const meRes = await api.me()
    me.value = meRes.user
  } catch {
    // keep current state
  }
}

async function refreshAll() {
  const [statsRes, categoriesRes, commentsRes, usersRes] = await Promise.all([
    api.adminStats(),
    api.adminCategories(),
    api.adminComments(),
    api.adminUsers(),
  ])
  stats.value = statsRes
  categories.value = categoriesRes.items
  comments.value = commentsRes.items
  users.value = usersRes.items
  await loadArticles()
  await loadNotifications()
}

async function syncNotifications() {
  await loadNotifications()
  if (activeSection.value === 'comments') {
    comments.value = (await api.adminComments()).items
  }
  if (activeSection.value === 'users') {
    users.value = (await api.adminUsers()).items
  }
}

async function loadArticles() {
  articles.value = (await api.adminArticles({ q: articleQuery.value || undefined, status: articleStatus.value || undefined })).items
}

async function loadNotifications() {
  try {
    notifications.value = await api.adminNotifications()
  } catch {
    notifications.value = null
  }
}

async function markNotificationsRead(types: NotificationType[] = [], articleIds: number[] = []) {
  try {
    await api.markNotificationsRead(types, articleIds)
    await loadNotifications()
    window.dispatchEvent(new CustomEvent('blog-notifications-changed'))
  } catch {
    // ignore
  }
}

function commentUnreadCount(articleId: number) {
  return unreadCommentCountsByArticle.value.get(articleId) ?? 0
}

function onCommentGroupToggle(group: { articleId: number }, event: Event) {
  const details = event.currentTarget as HTMLDetailsElement | null
  if (!details?.open || commentUnreadCount(group.articleId) === 0) return
  void markNotificationsRead(['comment'], [group.articleId])
}

async function saveArticle() {
  error.value = ''
  notice.value = ''
  try {
    const editorMarkdown = articleEditorRef.value?.getMarkdown()
    if (typeof editorMarkdown === 'string') {
      articleForm.content = editorMarkdown
    }
    await api.saveArticle(articleForm)
    notice.value = articleForm.id ? '文章已更新' : '文章已创建'
    resetArticleForm()
    await refreshAll()
    activeSection.value = 'articles'
  } catch (e) {
    error.value = (e as Error).message
  }
}

function editArticle(item: Article) {
  activeSection.value = 'createArticle'
  Object.assign(articleForm, {
    id: item.id,
    coverUrl: item.coverUrl || '',
    title: item.title,
    summary: item.summary || '',
    tags: Array.isArray(item.tags) ? item.tags.join(', ') : '',
    content: item.content || '',
    status: item.status,
    categoryId: item.categoryId || 0,
  })
  notice.value = '已载入文章到编辑器'
}

function resetArticleForm() {
  Object.assign(articleForm, {
    id: 0,
    coverUrl: '',
    title: '',
    summary: '',
    tags: '',
    content: '',
    status: 'draft',
    categoryId: 0,
  })
}

async function saveCategory() {
  if (!categoryName.value.trim()) return
  await api.saveCategory({ name: categoryName.value })
  categoryName.value = ''
  await refreshAll()
}

async function removeArticle(id: number) {
  await api.deleteArticle(id)
  await refreshAll()
}

async function removeCategory(id: number) {
  await api.deleteCategory(id)
  await refreshAll()
}

async function setCommentStatus(id: number, status: 'published' | 'hidden') {
  await api.moderateComment(id, status)
  await refreshAll()
}

onMounted(() => {
  load()
  window.addEventListener('blog-auth-changed', syncMe)
  window.addEventListener('blog-notifications-changed', syncNotifications)
})
onBeforeUnmount(() => {
  window.removeEventListener('blog-auth-changed', syncMe)
  window.removeEventListener('blog-notifications-changed', syncNotifications)
})
</script>
