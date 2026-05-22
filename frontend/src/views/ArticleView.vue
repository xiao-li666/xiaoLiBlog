<template>
  <section class="article-page" :class="{ 'without-toc': !toc.length }" v-if="article">
    <aside class="article-toc" v-if="toc.length">
      <section class="sidebar-card toc-card">
        <div class="sidebar-title">文章目录</div>
        <nav class="toc-nav" aria-label="文章目录">
          <a
            v-for="item in toc"
            :key="item.id"
            :href="`#${item.id}`"
            :class="['toc-link', `depth-${item.depth}`, { active: activeHeading === item.id }]"
            @click="activeHeading = item.id"
          >
            {{ item.text }}
          </a>
        </nav>
      </section>
    </aside>

    <main class="article-main">
      <article class="article-reader">
        <header class="article-header">
          <div class="article-meta">
            <span>{{ article.category?.name || '未分类' }}</span>
            <span>·</span>
            <span>{{ article.author?.name || '站长' }}</span>
            <span v-if="article.publishedAt">·</span>
            <span v-if="article.publishedAt">{{ formatDate(article.publishedAt) }}</span>
          </div>
          <h1>{{ article.title }}</h1>
          <p class="article-summary" v-if="article.summary">{{ article.summary }}</p>
          <div class="chip-row" v-if="article.tags?.length">
            <span v-for="tag in article.tags" :key="tag" class="chip">{{ tag }}</span>
          </div>
        </header>

        <img v-if="article.coverUrl" :src="article.coverUrl" class="article-cover" />

        <div class="markdown-body" v-html="html"></div>
      </article>
    </main>

    <aside class="article-aside">
      <section class="sidebar-card">
        <div class="sidebar-title">互动</div>
        <div class="article-actions">
          <button :class="['reaction-button', { animating: reactionPulse.like }]" @click="react('like')">
            <ThumbsUpIcon :size="16" /> 点赞 {{ article.likesCount }}
          </button>
          <button :class="['reaction-button ghost', 'favorite', { animating: reactionPulse.favorite }]" @click="react('favorite')">
            <BookmarkIcon :size="16" /> 收藏 {{ article.favoritesCount }}
          </button>
        </div>
      </section>

      <section class="article-panel article-comment-panel">
        <div class="panel-head">
          <h3><MessageSquareIcon :size="16" /> 评论</h3>
          <span class="muted">{{ comments.length }} 条</span>
        </div>
        <textarea v-model="comment" placeholder="写下你的评论"></textarea>
        <div class="actions-row">
          <button @click="sendComment"><SendIcon :size="14" /> 发表评论</button>
        </div>

        <div class="comment-list">
          <div class="comment-card" v-for="item in comments" :key="item.id">
            <div class="comment-avatar">
              <img :src="item.user.avatarUrl || defaultAvatar" :alt="item.user.name" />
            </div>
            <div>
              <strong>{{ item.user.name }}</strong>
              <p>{{ item.body }}</p>
            </div>
          </div>
        </div>
      </section>

      <section class="sidebar-card" v-if="related.length">
        <div class="sidebar-title"><LinkIcon :size="14" /> 相关文章</div>
        <div class="rank-list">
          <RouterLink v-for="item in related" :key="item.id" class="rank-item" :to="`/article/${item.slug}`">
            <span class="rank-title">{{ item.title }}</span>
          </RouterLink>
        </div>
      </section>
    </aside>
  </section>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Marked, type Token, type Tokens } from 'marked'
import { BookmarkIcon, LinkIcon, MessageSquareIcon, SendIcon, ThumbsUpIcon } from 'lucide-vue-next'
import { api } from '../api'
import type { Article, Comment } from '../types'
import { highlightCodeToHtml, normalizeCodeLanguage } from '../utils/codeHighlight'

const defaultAvatar =
  'data:image/svg+xml;charset=UTF-8,' +
  encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96">
      <rect width="96" height="96" rx="24" fill="#30363d"/>
      <circle cx="48" cy="38" r="17" fill="#8b949e"/>
      <path d="M20 80c4-17 19-26 28-26s24 9 28 26" fill="#8b949e"/>
    </svg>
  `)

type TocItem = {
  id: string
  text: string
  depth: number
}

let tocDraft: TocItem[] = []
let headingIds = new WeakMap<Tokens.Heading, string>()
let headingCounts = new Map<string, number>()

const markdown = new Marked({
  gfm: true,
  breaks: false,
  walkTokens(token) {
    if (token.type !== 'heading') return
    const heading = token as Tokens.Heading
    const text = tokenText(heading.tokens) || heading.text
    const id = uniqueHeadingId(text)
    headingIds.set(heading, id)
    if (heading.depth <= 4) {
      tocDraft.push({ id, text, depth: heading.depth })
    }
  },
  renderer: {
    heading(token: Tokens.Heading) {
      const id = headingIds.get(token) || uniqueHeadingId(token.text)
      const text = tokenText(token.tokens) || token.text
      return `<h${token.depth} id="${escapeAttr(id)}" class="markdown-heading">${escapeHtml(text)}</h${token.depth}>`
    },
    code(token: Tokens.Code) {
      const language = normalizeCodeLanguage(token.lang || '')
      const label = language || 'text'
      return [
        '<figure class="code-block">',
        `<figcaption><span>${escapeHtml(label)}</span></figcaption>`,
        `<pre><code class="language-${escapeAttr(language || 'text')}">${highlightCodeToHtml(token.text, language)}</code></pre>`,
        '</figure>',
      ].join('')
    },
  },
})

const route = useRoute()
const article = ref<Article | null>(null)
const related = ref<Article[]>([])
const comments = ref<Comment[]>([])
const comment = ref('')
const html = ref('')
const toc = ref<TocItem[]>([])
const activeHeading = ref('')
const reactionPulse = reactive({
  like: false,
  favorite: false,
})

let headingObserver: IntersectionObserver | null = null
const router = useRouter()

function renderMarkdown(source: string) {
  tocDraft = []
  headingIds = new WeakMap<Tokens.Heading, string>()
  headingCounts = new Map<string, number>()
  const rendered = markdown.parse(source || '', { async: false })
  return {
    html: rendered,
    toc: [...tocDraft],
  }
}

function tokenText(tokens: Token[] = []): string {
  return tokens.map((token) => {
    if ('tokens' in token && Array.isArray(token.tokens)) return tokenText(token.tokens)
    if ('text' in token && typeof token.text === 'string') return token.text
    return ''
  }).join('')
}

function uniqueHeadingId(value: string) {
  const base = value
    .trim()
    .toLowerCase()
    .replace(/[^\w\u4e00-\u9fa5\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '') || 'section'
  const count = headingCounts.get(base) || 0
  headingCounts.set(base, count + 1)
  return count ? `${base}-${count + 1}` : base
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function escapeAttr(value: string) {
  return escapeHtml(value)
}

function setupHeadingObserver() {
  teardownHeadingObserver()
  activeHeading.value = toc.value[0]?.id || ''
  const headings = toc.value
    .map((item) => document.getElementById(item.id))
    .filter((item): item is HTMLElement => Boolean(item))
  if (!headings.length) return

  headingObserver = new IntersectionObserver((entries) => {
    const visible = entries
      .filter((entry) => entry.isIntersecting)
      .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0]
    if (visible?.target.id) {
      activeHeading.value = visible.target.id
    }
  }, { rootMargin: '-18% 0px -68% 0px', threshold: [0, 1] })

  headings.forEach((heading) => headingObserver?.observe(heading))
}

function teardownHeadingObserver() {
  headingObserver?.disconnect()
  headingObserver = null
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(value))
}

async function load() {
  const res = await api.article(String(route.params.slug))
  article.value = res.article
  const rendered = renderMarkdown(res.article.content || '')
  html.value = rendered.html
  toc.value = rendered.toc
  comments.value = res.comments
  related.value = res.related
  await nextTick()
  setupHeadingObserver()
}

async function sendComment() {
  if (!article.value || !comment.value.trim()) return
  if (!hasLogin()) {
    promptLogin()
    return
  }
  try {
    await api.addComment(article.value.id, comment.value)
    comment.value = ''
    await load()
  } catch (e) {
    handleActionError(e)
  }
}

async function react(type: 'like' | 'favorite') {
  if (!article.value) return
  if (!hasLogin()) {
    promptLogin()
    return
  }
  try {
    await api.toggleReaction(article.value.id, type)
    pulseReaction(type)
    await load()
  } catch (e) {
    handleActionError(e)
  }
}

function hasLogin() {
  return !!localStorage.getItem('blog_token')
}

function promptLogin() {
  window.alert('请先进行登录')
  void router.push({ path: '/auth', query: { redirect: route.fullPath } })
}

function handleActionError(error: unknown) {
  const message = error instanceof Error ? error.message : ''
  if (message.includes('401') || message.includes('missing token') || message.includes('invalid token')) {
    promptLogin()
  }
}

function pulseReaction(type: 'like' | 'favorite') {
  reactionPulse[type] = true
  window.setTimeout(() => {
    reactionPulse[type] = false
  }, 220)
}

onMounted(load)
onBeforeUnmount(teardownHeadingObserver)
watch(() => route.params.slug, load)
</script>
