<template>
  <section class="home-page">
    <section class="hero hero-slim">
      <div class="hero-copy">
        <p class="eyebrow">xiaoli博客</p>
        <h1>记录、分享、沉淀</h1>
      </div>
      <div class="hero-search">
        <div class="search-bar">
          <input v-model="q" placeholder="搜索标题、摘要、正文、标签" @keyup.enter="loadArticles" />
          <button @click="loadArticles"><SearchIcon :size="16" /> 搜索</button>
        </div>
      </div>
    </section>

    <section class="blog-layout">
      <aside class="sidebar sidebar-left">
        <div class="sidebar-card sidebar-category-card">
          <div class="sidebar-title">
            <LayersIcon :size="15" />
            <span>分类导航</span>
          </div>
          <nav class="category-nav">
            <button
              class="category-nav-item"
              :class="{ active: category === '' && !selectedTag }"
              @click="selectCategory('')"
            >
              <span class="category-nav-icon"><LayoutListIcon :size="16" /></span>
              <span class="category-nav-label">全部文章</span>
              <span class="category-nav-count">{{ totalArticles }}</span>
            </button>
            <button
              v-for="item in categories"
              :key="item.id"
              class="category-nav-item"
              :class="{ active: category === item.slug && !selectedTag }"
              @click="selectCategory(item.slug)"
            >
              <span class="category-nav-icon"><HashIcon :size="16" /></span>
              <span class="category-nav-label">{{ item.name }}</span>
              <span class="category-nav-count">{{ item.articleCount ?? 0 }}</span>
            </button>
          </nav>
        </div>

        <div class="sidebar-card sidebar-tag-card" v-if="tags.length > 0">
          <div class="sidebar-title">
            <TagIcon :size="15" />
            <span>热门标签</span>
          </div>
          <div class="tag-cloud">
            <button
              v-for="tag in tags"
              :key="tag.name"
              class="tag-chip"
              :class="{ active: selectedTag === tag.name }"
              @click="selectTag(tag.name)"
            >
              <span>{{ tag.name }}</span>
              <span class="tag-chip-count">{{ tag.articleCount }}</span>
            </button>
          </div>
        </div>

        <div class="sidebar-card sidebar-platform-card" v-if="externalLinks.length > 0">
          <div class="sidebar-title">
            <QrCodeIcon :size="15" />
            <span>我的平台</span>
          </div>
          <div class="platform-list">
            <article v-for="item in externalLinks" :key="item.id" class="platform-card">
              <div class="platform-qr" v-if="item.qrCodeUrl">
                <img :src="item.qrCodeUrl" :alt="`${item.platform} 二维码`" />
              </div>
              <div class="platform-card-body">
                <span class="platform-name">{{ item.platform }}</span>
                <strong>{{ item.name }}</strong>
                <p v-if="item.description">{{ item.description }}</p>
                <a v-if="item.linkUrl" :href="item.linkUrl" target="_blank" rel="noreferrer">
                  <Link2Icon :size="13" />
                  访问
                </a>
              </div>
            </article>
          </div>
        </div>
      </aside>

      <main class="feed">
        <div class="feed-inner">
          <div class="section-head">
            <div>
              <p class="eyebrow">最新文章</p>
              <h2>{{ activeLabel }}</h2>
            </div>
            <div class="meta">{{ articles.length }} 篇文章</div>
          </div>

          <article class="post-card post-card-hero" v-if="articles[0]">
            <div class="post-card-shine"></div>
            <div class="post-body">
              <div class="post-meta">
                <span class="post-meta-tag"><FolderOpenIcon :size="13" /> {{ articles[0].category?.name || '未分类' }}</span>
                <span class="post-meta-sep">·</span>
                <span class="post-meta-tag"><UserRoundIcon :size="13" /> {{ articles[0].author?.name || '站长' }}</span>
                <span class="post-meta-sep">·</span>
                <span class="post-meta-tag"><CalendarDaysIcon :size="13" /> {{ formatDate(articles[0].publishedAt) }}</span>
              </div>
              <h3><RouterLink :to="`/article/${articles[0].slug}`">{{ articles[0].title }}</RouterLink></h3>
              <p>{{ articleSummary(articles[0]) }}</p>
              <div class="post-footer">
                <div class="chip-row">
                  <span v-for="tag in articles[0].tags" :key="tag" class="chip">{{ tag }}</span>
                </div>
                <div class="post-stats">
                  <span class="post-stat"><HeartIcon :size="14" /> <strong>{{ articles[0].likesCount }}</strong></span>
                  <span class="post-stat"><BookmarkIcon :size="14" /> <strong>{{ articles[0].favoritesCount }}</strong></span>
                  <RouterLink :to="`/article/${articles[0].slug}`" class="post-read-link">阅读全文 <ArrowRightIcon :size="14" /></RouterLink>
                </div>
              </div>
            </div>
          </article>

          <div class="article-list">
            <article class="post-card post-card-row" v-for="item in articles.slice(1)" :key="item.id">
              <div class="post-body">
                <div class="post-meta">
                  <span class="post-meta-tag"><FolderOpenIcon :size="13" /> {{ item.category?.name || '未分类' }}</span>
                  <span class="post-meta-sep">·</span>
                  <span class="post-meta-tag"><CalendarDaysIcon :size="13" /> {{ formatDate(item.publishedAt) }}</span>
                </div>
                <h3><RouterLink :to="`/article/${item.slug}`">{{ item.title }}</RouterLink></h3>
                <p>{{ articleSummary(item) }}</p>
                <div class="post-footer">
                  <div class="chip-row">
                    <span v-for="tag in item.tags" :key="tag" class="chip">{{ tag }}</span>
                  </div>
                  <div class="post-stats">
                    <span class="post-stat"><HeartIcon :size="14" /> {{ item.likesCount }}</span>
                    <span class="post-stat"><BookmarkIcon :size="14" /> {{ item.favoritesCount }}</span>
                    <RouterLink :to="`/article/${item.slug}`" class="post-read-link">阅读全文 <ArrowRightIcon :size="14" /></RouterLink>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </div>
      </main>

      <aside class="sidebar sidebar-right">
        <div class="sidebar-card">
          <div class="sidebar-title"><TrendingUpIcon :size="14" /> 热门文章</div>
          <div class="rank-list">
            <RouterLink v-for="item in popular" :key="item.id" class="rank-item" :to="`/article/${item.slug}`">
              <span class="rank-title">{{ item.title }}</span>
              <span class="rank-meta">
                <span class="rank-stat"><HeartIcon :size="12" /> {{ item.likesCount }}</span>
                <span class="rank-stat"><BookmarkIcon :size="12" /> {{ item.favoritesCount }}</span>
              </span>
            </RouterLink>
          </div>
        </div>

        <div class="sidebar-card">
          <div class="sidebar-title"><BookOpenIcon :size="14" /> 推荐阅读</div>
          <div class="recommend-item" v-for="item in popular.slice(0, 3)" :key="item.slug + '-rec'">
            <RouterLink :to="`/article/${item.slug}`">{{ item.title }}</RouterLink>
            <p>{{ articleSummary(item) }}</p>
          </div>
        </div>
      </aside>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRightIcon, BookmarkIcon, BookOpenIcon, CalendarDaysIcon, FolderOpenIcon, HashIcon, HeartIcon, LayersIcon, LayoutListIcon, Link2Icon, QrCodeIcon, SearchIcon, TagIcon, TrendingUpIcon, UserRoundIcon } from 'lucide-vue-next'
import { api } from '../api'
import type { Article, Category, ExternalLink, TagStat } from '../types'

const q = ref('')
const category = ref('')
const selectedTag = ref('')
const articles = ref<Article[]>([])
const totalArticles = ref(0)
const popular = ref<Article[]>([])
const categories = ref<Category[]>([])
const tags = ref<TagStat[]>([])
const externalLinks = ref<ExternalLink[]>([])

function articleSummary(item: Article) {
  return item.summary || item.content.slice(0, 140).replace(/\s+/g, ' ').trim()
}

function formatDate(iso: string | undefined) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
}

async function loadArticles() {
  const query = [q.value, selectedTag.value].filter(Boolean).join(' ') || undefined
  const res = await api.articles({ q: query, category: category.value || undefined, pageSize: 50 })
  articles.value = res.items
  if (!query && !category.value) {
    totalArticles.value = res.total
  }
}

function selectCategory(slug: string) {
  category.value = slug
  selectedTag.value = ''
  loadArticles()
}

function selectTag(tag: string) {
  if (selectedTag.value === tag) {
    selectedTag.value = ''
  } else {
    selectedTag.value = tag
    category.value = ''
  }
  loadArticles()
}

const activeLabel = computed(() => {
  if (selectedTag.value) return `标签: ${selectedTag.value}`
  if (!category.value) return '全部分类'
  return categories.value.find(item => item.slug === category.value)?.name || '全部分类'
})

onMounted(async () => {
  const [categoriesRes, tagsRes, popularRes, externalLinksRes] = await Promise.all([
    api.categories(),
    api.tags(),
    api.popularArticles(),
    api.externalLinks(),
  ])
  categories.value = categoriesRes.items
  tags.value = tagsRes.items
  popular.value = popularRes.items
  externalLinks.value = externalLinksRes.items
  await loadArticles()
})
</script>
