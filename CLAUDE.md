# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概览

个人博客系统，Vue 3 + TypeScript 前端，Go + Gin + GORM 后端，MySQL 8.x 数据库，Redis 可选缓存。第一位注册用户自动成为站长（admin），访客可浏览/搜索，登录用户可评论/点赞/收藏，站长可进入后台管理文章、分类、评论、用户。

## 常用命令

```bash
# 依赖服务（可选，手动起 MySQL/Redis 也行）
docker compose up -d

# 后端
cd backend && go mod tidy && go run .

# 前端
cd frontend && npm install && npm run dev    # 开发服务器 :5173
cd frontend && npm run build                  # 生产构建
```

后端默认监听 `:8080`，前端开发服务器默认 `:5173`。环境变量在根目录 `.env` 文件中配置（参考 `.env.example`）。

## 架构

### 后端 (`backend/`)

```
main.go                  → 极薄入口，直接调用 app.Run()
internal/
  app/app.go             → 启动编排：加载配置 → 建上传目录 → 连 MySQL → AutoMigrate → seed 分类 → 可选连 Redis → 创建 Handler → 创建路由 → 启动服务
  config/config.go       → 从 .env 读取配置
  database/database.go   → GORM 连接、AutoMigrate、seed 4 个默认分类
  model/models.go        → 6 个模型：User, Category, Article, Comment, Reaction, Notification
  handler/handler.go     → 所有 handler 方法 + 中间件 + SSE notificationHub（~1223 行，是整个后端核心）
  router/router.go       → 路由定义 + CORS/AuthRequired/AdminRequired 中间件挂载
```

路由层级：
- **公开**：健康检查、注册、登录、文章列表/详情/热门、分类列表、评论查询
- **需登录**（`AuthRequired`）：个人信息、头像上传、发表评论、点赞/收藏、个人互动记录
- **需管理员**（`AdminRequired`）：统计数据、文件上传、文章/分类/评论/用户 CRUD、通知管理

JWT 认证流程：`Authorization: Bearer <token>` 头或 SSE 连接时 `?token=` 查询参数。中间件解析 JWT → 查 DB 获取用户 → `c.Set("user", user)`，后续 handler 通过 `currentUser(c)` 获取。

Redis 缓存：仅缓存文章列表（key 模式 `articles:list:*`，45 秒 TTL），文章/评论/互动变更时主动失效。`REDIS_ADDR` 留空则跳过 Redis。

SSE 通知：内存 pub/sub（`notificationHub`），站长在后台时通过 EventSource 实时接收通知（注册、评论、点赞、收藏事件）。

### 前端 (`frontend/src/`)

```
main.ts          → 创建 Vue app + Vue Router（4 个路由）
App.vue          → 根组件：顶栏导航、通知铃铛（管理员）、用户菜单（含个人信息编辑弹窗）、SSE 连接、localStorage token 管理
api.ts           → fetch 封装，自动注入 Authorization 头，25+ 方法覆盖所有接口
types.ts         → TypeScript 类型定义
styles.css       → 全局样式（~1933 行，暗色主题，CSS 自定义属性）
views/
  HomeView.vue   → 首页：英雄区、文章列表、分类筛选、搜索、热门文章侧边栏
  ArticleView.vue → 文章详情：三栏布局（正文 + 目录 + 侧边栏）、评论、点赞/收藏
  AuthView.vue   → 登录/注册切换
  AdminView.vue  → 后台管理：统计、文章 CRUD（含 Tiptap 富文本编辑器）、分类管理、评论管理、用户列表（~1126 行）
components/
  ArticleTiptapEditor.vue → Tiptap 富文本编辑器封装
  SiteLogo.vue   → SVG Logo
```

无 Pinia/Vuex 状态管理，状态在各组件内用 `ref()`/`reactive()` 管理。登录状态通过 `localStorage.getItem('blog_token')` 检测，跨组件通信用 `blog-auth-changed` 和 `blog-notifications-changed` 自定义事件。

### 数据库

6 张表：`users`, `categories`, `articles`, `comments`, `reactions`, `notifications`。详细设计见 `docs/DATABASE.md`。

关键设计决策：
- 标签存储为逗号分隔字符串（非独立表）
- 点赞和收藏共用 `reactions` 表，通过 `type` 字段区分（`like` / `favorite`），唯一约束 `(user_id, article_id, type)` 保证同一用户对同一文章每种互动只有一条
- `articles.likes_count` 和 `articles.favorites_count` 是冗余计数字段，用于热门排序，真实明细在 `reactions` 表
- 未分类文章的 `category_id` 为 NULL，接口层返回 `categoryId: 0`
- 软删除：`users` 支持 GORM 软删除，评论用 `status = hidden` 隐藏

### 无测试/无 Linter

当前项目没有配置任何测试框架（前后端均无）和 linter。`docs/DEVELOPMENT.md` 中的 `go test ./...` 目前只是占位说明，没有实际测试文件。
