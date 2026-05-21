<template>
  <section class="admin-shell" v-if="ready">
    <aside class="admin-sidebar">
      <div class="admin-brand">
        <p class="eyebrow">Backend</p>
        <h2>xiaoli博客后台</h2>
        <p class="muted">文章、分类、评论、用户统一管理。</p>
      </div>

      <div class="admin-menu">
        <button
          v-for="item in adminMenu"
          :key="item.key"
          :class="['admin-menu-item', { active: activeSection === item.key }]"
          @click="setActiveSection(item.key)"
        >
          {{ item.label }}
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
          <p class="muted">{{ activeSectionCopy.desc }}</p>
        </div>
        <div class="admin-hero-actions">
          <button @click="refreshAll">刷新数据</button>
          <button class="ghost" @click="startCreateArticle">新建文章</button>
        </div>
      </header>

      <section class="stat-grid" v-if="activeSection === 'overview'">
        <div class="stat-card" v-for="item in statCards" :key="item.label">
          <span class="stat-label">{{ item.label }}</span>
          <strong class="stat-value">{{ item.value }}</strong>
          <span class="stat-desc">{{ item.desc }}</span>
        </div>
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
            <button @click="loadArticles">筛选</button>
          </div>
        </div>

        <div class="table-list">
          <div class="table-head">
            <span>标题</span>
            <span>分类</span>
            <span>状态</span>
            <span>热度</span>
            <span>操作</span>
          </div>
          <div v-for="item in articles" :key="item.id" class="table-row">
            <div>
              <strong>{{ item.title }}</strong>
              <p>{{ item.summary || articleSummary(item) }}</p>
            </div>
            <span>{{ item.category?.name || '未分类' }}</span>
            <span>{{ item.status }}</span>
            <span>{{ item.likesCount + item.favoritesCount }}</span>
            <div class="actions-row small">
              <button class="ghost" @click="editArticle(item)">编辑</button>
              <button class="ghost danger" @click="removeArticle(item.id)">删除</button>
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

        <div class="markdown-editor">
          <div class="markdown-toolbar">
            <button class="ghost" @click="insertMarkdown('## ', '', '二级标题')">H2</button>
            <button class="ghost" @click="insertMarkdown('**', '**', '加粗文本')">B</button>
            <button class="ghost" @click="insertMarkdown('[', '](https://example.com)', '链接文本')">链接</button>
            <button class="ghost" @click="insertMarkdown('> ', '', '引用内容')">引用</button>
            <button class="ghost" @click="insertMarkdown('- ', '', '列表项')">列表</button>
            <button class="ghost" @click="insertCodeBlock">代码</button>
          </div>

          <div class="markdown-editor-grid" ref="markdownEditorGrid" :style="markdownGridStyle" :class="{ dragging: isMarkdownDragging }">
            <label class="markdown-editor-pane">
              <span>Markdown 正文</span>
              <textarea
                ref="markdownTextarea"
                v-model="articleForm.content"
                rows="18"
                placeholder="# 标题&#10;&#10;在这里使用 Markdown 编写正文..."
              ></textarea>
            </label>

            <div
              class="markdown-resizer"
              role="separator"
              aria-orientation="vertical"
              aria-label="调整编辑区和预览区宽度"
              @pointerdown="startMarkdownResize"
            >
              <span></span>
            </div>

            <section class="markdown-preview-pane">
              <div class="preview-head">
                <span>实时预览</span>
                <span class="muted">{{ articleForm.content.length }} 字符</span>
              </div>
              <div class="markdown-body admin-markdown-preview" v-html="markdownPreview"></div>
            </section>
          </div>
        </div>

        <div class="actions-row">
          <button @click="saveArticle">保存文章</button>
          <button class="ghost" @click="resetArticleForm">清空</button>
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
              <button @click="saveCategory">新增</button>
            </div>
            <div class="small-list">
              <div v-for="cat in categories" :key="cat.id" class="small-row">
                <div>
                  <strong>{{ cat.name }}</strong>
                  <p>{{ cat.slug }}</p>
                </div>
                <button class="ghost" @click="removeCategory(cat.id)">删除</button>
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
                <div>
                  <strong>{{ user.name }}</strong>
                  <p>{{ user.email }}</p>
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
        </div>
        <div class="table-list">
          <div class="table-head">
            <span>用户</span>
            <span>内容</span>
            <span>状态</span>
            <span>操作</span>
          </div>
          <div v-for="item in comments" :key="item.id" class="table-row">
            <div>
              <strong>{{ item.user.name }}</strong>
              <p>{{ item.user.email }}</p>
            </div>
            <span>{{ item.body }}</span>
            <span>{{ item.status }}</span>
            <div class="actions-row small">
              <button class="ghost" @click="setCommentStatus(item.id, 'published')">显示</button>
              <button class="ghost danger" @click="setCommentStatus(item.id, 'hidden')">隐藏</button>
            </div>
          </div>
        </div>
      </section>
    </main>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { marked } from 'marked'
import { api } from '../api'
import type { AdminStats, Article, Category, Comment, User } from '../types'

type AdminSection = 'overview' | 'articles' | 'createArticle' | 'categories' | 'comments' | 'users'

const router = useRouter()
const ready = ref(false)
const me = ref<User | null>(null)
const stats = ref<AdminStats | null>(null)
const articles = ref<Article[]>([])
const categories = ref<Category[]>([])
const comments = ref<Comment[]>([])
const users = ref<User[]>([])
const categoryName = ref('')
const articleQuery = ref('')
const articleStatus = ref('')
const activeSection = ref<AdminSection>('overview')
const markdownTextarea = ref<HTMLTextAreaElement | null>(null)
const markdownEditorGrid = ref<HTMLElement | null>(null)
const markdownSplit = ref(55)
const isMarkdownDragging = ref(false)
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

const statCards = computed(() => [
  { label: '文章总数', value: stats.value?.articles.total ?? 0, desc: '含草稿与已发布' },
  { label: '已发布', value: stats.value?.articles.published ?? 0, desc: '对访客可见' },
  { label: '草稿', value: stats.value?.articles.draft ?? 0, desc: '待处理内容' },
  { label: '分类', value: stats.value?.categories ?? 0, desc: '内容结构' },
  { label: '评论', value: stats.value?.comments ?? 0, desc: '互动记录' },
  { label: '用户', value: stats.value?.users ?? 0, desc: '注册用户' },
])

const adminMenu: Array<{ key: AdminSection; label: string }> = [
  { key: 'overview', label: '概览' },
  { key: 'articles', label: '文章' },
  { key: 'createArticle', label: '新建文章' },
  { key: 'categories', label: '分类' },
  { key: 'comments', label: '评论' },
  { key: 'users', label: '用户' },
]

const activeSectionCopy = computed(() => {
  const map: Record<AdminSection, { title: string; desc: string }> = {
    overview: {
      title: '大气简约的博客管理台',
      desc: '统一处理文章发布、分类维护、评论审核和用户查看。',
    },
    articles: {
      title: '文章管理',
      desc: '查看、筛选、编辑和删除文章内容。',
    },
    createArticle: {
      title: articleForm.id ? '编辑文章' : '新建文章',
      desc: '填写文章信息、上传封面或 Markdown 文件，并保存为草稿或发布。',
    },
    categories: {
      title: '分类管理',
      desc: '维护文章分类结构，让内容组织更清晰。',
    },
    comments: {
      title: '评论管理',
      desc: '审核评论内容，控制前台展示状态。',
    },
    users: {
      title: '用户管理',
      desc: '查看已注册用户与其基础信息。',
    },
  }
  return map[activeSection.value]
})

const markdownPreview = computed(() => String(marked.parse(articleForm.content || '')))
const markdownGridStyle = computed(() => ({
  gridTemplateColumns: `${markdownSplit.value}fr 12px ${100 - markdownSplit.value}fr`,
}))

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
  void insertMarkdown('```js\n', '\n```', 'console.log("hello")')
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function updateMarkdownSplit(clientX: number) {
  const grid = markdownEditorGrid.value
  if (!grid) return
  const rect = grid.getBoundingClientRect()
  const ratio = ((clientX - rect.left) / rect.width) * 100
  markdownSplit.value = clamp(Math.round(ratio), 30, 70)
}

function stopMarkdownResize() {
  if (!isMarkdownDragging.value) return
  isMarkdownDragging.value = false
  document.removeEventListener('pointermove', handleMarkdownResizeMove)
  document.removeEventListener('pointerup', stopMarkdownResize)
  document.removeEventListener('pointercancel', stopMarkdownResize)
}

function handleMarkdownResizeMove(event: PointerEvent) {
  updateMarkdownSplit(event.clientX)
}

function startMarkdownResize(event: PointerEvent) {
  event.preventDefault()
  isMarkdownDragging.value = true
  updateMarkdownSplit(event.clientX)
  document.addEventListener('pointermove', handleMarkdownResizeMove)
  document.addEventListener('pointerup', stopMarkdownResize)
  document.addEventListener('pointercancel', stopMarkdownResize)
}

function articleSummary(item: Article) {
  return item.summary || item.content.slice(0, 120).replace(/\s+/g, ' ').trim()
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
}

async function loadArticles() {
  articles.value = (await api.adminArticles({ q: articleQuery.value || undefined, status: articleStatus.value || undefined })).items
}

async function saveArticle() {
  error.value = ''
  notice.value = ''
  try {
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
})
onBeforeUnmount(() => {
  window.removeEventListener('blog-auth-changed', syncMe)
})
onBeforeUnmount(stopMarkdownResize)
</script>
