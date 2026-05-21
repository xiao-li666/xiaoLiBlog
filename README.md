# 个人博客

## 技术栈
- 前端：Vue 3 + Vite
- 后端：Go + Gin + GORM
- 数据库：MySQL
- 缓存：Redis（可选）

## 运行
1. 本地 MySQL 已在 `.env` 中配置，Redis 目前留空即可。注意 MySQL 密码如果以 `@` 结尾，DSN 里要写成双 `@`，例如 `root:密码@@tcp(...)`。
2. 如需一键起依赖，可选运行 `docker compose up -d`。
3. 安装并启动后端：`cd backend && go mod tidy && go run .`。
4. 安装并启动前端：`cd frontend && npm install && npm run dev`。
5. 打开 `http://localhost:5173`。

## 默认规则
- 第一位注册用户自动成为站长。
- 访客可以浏览、搜索、按分类查看文章。
- 登录用户可以评论、点赞、收藏。
- 只有站长可以进入后台发布文章、管理文章和分类。

## 文档
- 接口文档：`docs/API.md`
- 数据库设计：`docs/DATABASE.md`
- 开发流程：`docs/DEVELOPMENT.md`
- MySQL 建表脚本：`database/schema.sql`

## 后端结构
- `backend/main.go`：极薄入口
- `backend/internal/app`：启动编排
- `backend/internal/config`：环境配置
- `backend/internal/database`：数据库连接、迁移、初始化
- `backend/internal/model`：GORM 模型
- `backend/internal/handler`：Gin 控制器与业务处理
- `backend/internal/router`：路由与中间件挂载
