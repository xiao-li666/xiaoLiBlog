<template>
  <section class="home-page">
    <section class="hero hero-slim">
      <div class="hero-copy">
        <p class="eyebrow">xiaoli博客</p>
        <h1>记录、分享、沉淀</h1>
        <p class="hero-text">简约阅读、分类浏览、热门推荐和模糊搜索都放在一个清爽的页面里。</p>
      </div>
      <div class="hero-search">
        <div class="search-bar">
          <input v-model="q" placeholder="搜索标题、摘要、正文、标签" @keyup.enter="loadArticles" />
          <button @click="loadArticles">搜索</button>
        </div>
        <div class="search-hint">支持模糊匹配，快速定位文章。</div>
      </div>
    </section>

    <section class="blog-layout">
      <aside class="sidebar sidebar-left">
        <div class="sidebar-card">
          <div class="sidebar-title">分类</div>
          <button class="category-item" :class="{ active: category === '' }" @click="selectCategory('')">
            全部文章
          </button>
          <button
            v-for="item in categories"
            :key="item.id"
            class="category-item"
            :class="{ active: category === item.slug }"
            @click="selectCategory(item.slug)"
          >
            <span>{{ item.name }}</span>
          </button>
        </div>

        <div class="sidebar-card">
          <div class="sidebar-title">专题入口</div>
          <div class="tag-cloud">
            <span class="mini-chip">Vue</span>
            <span class="mini-chip">Go</span>
            <span class="mini-chip">MySQL</span>
            <span class="mini-chip">前端</span>
            <span class="mini-chip">后端</span>
          </div>
        </div>
      </aside>

      <main class="feed">
      <div class="section-head">
        <div>
          <p class="eyebrow">最新文章</p>
          <h2>{{ activeLabel }}</h2>
        </div>
        <div class="meta">{{ articles.length }} 篇文章</div>
      </div>

        <article class="post-card featured" v-if="articles[0]">
          <div class="post-cover" :style="coverStyle(articles[0])"></div>
          <div class="post-body">
            <div class="post-meta">
              <span>{{ articles[0].category?.name || '未分类' }}</span>
              <span>·</span>
              <span>{{ articles[0].author?.name || '站长' }}</span>
            </div>
            <h3><RouterLink :to="`/article/${articles[0].slug}`">{{ articles[0].title }}</RouterLink></h3>
            <p>{{ articleSummary(articles[0]) }}</p>
            <div class="chip-row">
              <span v-for="tag in articles[0].tags" :key="tag" class="chip">{{ tag }}</span>
            </div>
          </div>
        </article>

        <div class="article-list">
          <article class="post-card compact" v-for="item in articles.slice(1)" :key="item.id">
            <div class="post-body">
              <div class="post-meta">
                <span>{{ item.category?.name || '未分类' }}</span>
                <span>·</span>
                <span>{{ item.author?.name || '站长' }}</span>
                <span>·</span>
                <span class="meta">{{ item.likesCount + item.favoritesCount }} 热度</span>
              </div>
              <h3><RouterLink :to="`/article/${item.slug}`">{{ item.title }}</RouterLink></h3>
              <p>{{ articleSummary(item) }}</p>
              <div class="chip-row">
                <span v-for="tag in item.tags" :key="tag" class="chip">{{ tag }}</span>
              </div>
            </div>
            <div class="post-cover small" :style="coverStyle(item)"></div>
          </article>
        </div>
      </main>

      <aside class="sidebar sidebar-right">
        <div class="sidebar-card">
          <div class="sidebar-title">热门文章</div>
          <div class="rank-list">
            <RouterLink v-for="item in popular" :key="item.id" class="rank-item" :to="`/article/${item.slug}`">
              <span class="rank-title">{{ item.title }}</span>
              <span class="rank-meta">{{ item.likesCount + item.favoritesCount }}</span>
            </RouterLink>
          </div>
        </div>

        <div class="sidebar-card">
          <div class="sidebar-title">推荐阅读</div>
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
import { api } from '../api'
import type { Article, Category } from '../types'

const q = ref('')
const category = ref('')
const articles = ref<Article[]>([])
const popular = ref<Article[]>([])
const categories = ref<Category[]>([])

function articleSummary(item: Article) {
  return item.summary || item.content.slice(0, 140).replace(/\s+/g, ' ').trim()
}

function coverStyle(item: Article) {
  return item.coverUrl
    ? { backgroundImage: `url(${item.coverUrl})` }
    : { backgroundImage: 'linear-gradient(135deg, #f3f4f6, #e5e7eb)' }
}

async function loadArticles() {
  const res = await api.articles({ q: q.value || undefined, category: category.value || undefined })
  articles.value = res.items
}

function selectCategory(slug: string) {
  category.value = slug
  loadArticles()
}

const activeLabel = computed(() => {
  if (!category.value) return '全部分类'
  return categories.value.find(item => item.slug === category.value)?.name || '全部分类'
})

onMounted(async () => {
  categories.value = (await api.categories()).items
  popular.value = (await api.popularArticles()).items
  await loadArticles()
})
</script>
