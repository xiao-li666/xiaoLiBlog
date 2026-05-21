# 个人博客 API 接口文档

## 基础信息

- Base URL：`http://localhost:8080/api`
- 数据格式：`Content-Type: application/json`
- 认证方式：JWT Bearer Token
- 认证请求头：`Authorization: Bearer <token>`
- 角色：
  - 访客：浏览文章、分类、搜索、查看评论
  - 登录用户：评论、点赞、收藏、查看个人互动记录
  - 站长：文章管理、分类管理、评论管理、用户查看

## 通用响应

成功响应按接口返回对应对象，失败响应统一格式：

```json
{
  "error": "错误原因"
}
```

常见状态码：

- `200`：请求成功
- `201`：创建成功
- `400`：参数错误
- `401`：未登录或 Token 无效
- `403`：权限不足
- `404`：资源不存在
- `500`：服务端错误

## 数据模型

### User

```json
{
  "id": 1,
  "name": "站长",
  "email": "admin@example.com",
  "role": "admin",
  "createdAt": "2026-05-21T10:00:00+08:00"
}
```

### Category

```json
{
  "id": 1,
  "name": "Go",
  "slug": "go"
}
```

### Article

```json
{
  "id": 1,
  "title": "第一篇博客",
  "slug": "first-post",
  "summary": "文章摘要",
  "coverUrl": "https://example.com/cover.jpg",
  "content": "# Markdown 正文",
  "tags": ["Go", "Vue"],
  "status": "published",
  "publishedAt": "2026-05-21T10:00:00+08:00",
  "categoryId": 1,
  "category": {},
  "author": {},
  "likesCount": 12,
  "favoritesCount": 5
}
```

### Comment

```json
{
  "id": 1,
  "articleId": 1,
  "body": "评论内容",
  "status": "published",
  "createdAt": "2026-05-21T10:00:00+08:00",
  "user": {}
}
```

## 健康检查

### GET `/health`

检查后端服务是否可用。

响应：

```json
{
  "ok": true
}
```

## 认证接口

### POST `/auth/register`

注册用户。第一位注册用户会自动成为站长。

请求：

```json
{
  "name": "站长",
  "email": "admin@example.com",
  "password": "123456"
}
```

响应：

```json
{
  "token": "jwt-token",
  "user": {}
}
```

### POST `/auth/login`

用户登录。

请求：

```json
{
  "email": "admin@example.com",
  "password": "123456"
}
```

响应：

```json
{
  "token": "jwt-token",
  "user": {}
}
```

### GET `/auth/me`

获取当前登录用户信息。

权限：登录用户

响应：

```json
{
  "user": {}
}
```

## 文章接口

### GET `/articles`

文章列表，默认只返回已发布文章。支持模糊搜索和分类筛选。

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `q` | string | 否 | 搜索标题、摘要、正文、标签 |
| `category` | string | 否 | 分类 `slug` 或分类名 |
| `page` | number | 否 | 页码，默认 `1` |
| `pageSize` | number | 否 | 每页数量，默认 `10`，最大 `50` |

响应：

```json
{
  "items": [],
  "page": 1,
  "pageSize": 10,
  "total": 20,
  "pages": 2
}
```

### GET `/articles/popular`

热门文章列表，按点赞数、收藏数、发布时间排序，最多返回 6 篇。

响应：

```json
{
  "items": []
}
```

### GET `/articles/:slug`

文章详情。已发布文章所有人可见，草稿仅站长可预览。

路径参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `slug` | string | 文章唯一标识 |

响应：

```json
{
  "article": {},
  "comments": [],
  "related": []
}
```

### GET `/articles/id/:id/comments`

获取文章公开评论。

响应：

```json
{
  "items": []
}
```

### POST `/articles/id/:id/comments`

发布评论。

权限：登录用户

请求：

```json
{
  "body": "这篇文章很有帮助"
}
```

响应：

```json
{
  "comment": {}
}
```

### POST `/articles/id/:id/reactions/:type`

切换点赞或收藏状态。重复调用会取消。

权限：登录用户

路径参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `id` | number | 文章 ID |
| `type` | string | `like` 或 `favorite` |

响应：

```json
{
  "on": true
}
```

### GET `/me/library`

获取当前用户点赞和收藏过的文章 ID。

权限：登录用户

响应：

```json
{
  "likes": [1, 2],
  "favorites": [3, 4]
}
```

## 分类接口

### GET `/categories`

公开分类列表。

响应：

```json
{
  "items": []
}
```

## 站长后台接口

后台接口路径统一以 `/admin` 开头，全部需要站长权限。

### GET `/admin/articles`

后台文章列表，包含草稿和已发布文章。

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `q` | string | 否 | 搜索标题、摘要、正文、标签 |
| `status` | string | 否 | `draft` 或 `published` |

响应：

```json
{
  "items": []
}
```

### GET `/admin/stats`

后台概览统计。

响应：

```json
{
  "articles": {
    "total": 12,
    "published": 9,
    "draft": 3
  },
  "categories": 4,
  "comments": 31,
  "users": 8,
  "reactions": {
    "likes": 42,
    "favorites": 16
  }
}
```

### POST `/admin/uploads`

上传文章相关文件。支持图片封面和 Markdown 文件。

请求方式：`multipart/form-data`

字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | file | 是 | 图片或 `.md` 文件 |
| `kind` | string | 否 | `image` 或 `markdown`，不传时按文件扩展名判断 |

图片响应：

```json
{
  "kind": "image",
  "url": "http://127.0.0.1:8080/uploads/cover.png",
  "filename": "cover.png"
}
```

Markdown 响应：

```json
{
  "kind": "markdown",
  "title": "article-title",
  "content": "# Markdown 内容",
  "filename": "article-title.md"
}
```

### POST `/admin/articles`

创建文章。

说明：`categoryId` 传 `0` 或不传表示未分类，数据库会存为 `NULL`。

请求：

```json
{
  "title": "文章标题",
  "summary": "文章摘要",
  "coverUrl": "https://example.com/cover.jpg",
  "content": "# Markdown 正文",
  "tags": "Go,Vue,MySQL",
  "status": "published",
  "categoryId": 1
}
```

响应：

```json
{
  "article": {}
}
```

### PUT `/admin/articles/:id`

更新文章。

请求字段同创建文章。`categoryId` 传 `0` 会清空文章分类。

响应：

```json
{
  "ok": true
}
```

### DELETE `/admin/articles/:id`

删除文章，同时删除文章评论和互动记录。

响应：

```json
{
  "ok": true
}
```

### GET `/admin/comments`

后台评论列表。

响应：

```json
{
  "items": []
}
```

### PATCH `/admin/comments/:id`

显示或隐藏评论。

请求：

```json
{
  "status": "hidden"
}
```

响应：

```json
{
  "ok": true
}
```

### GET `/admin/users`

用户列表。

响应：

```json
{
  "items": []
}
```

### GET `/admin/categories`

后台分类列表。

响应：

```json
{
  "items": []
}
```

### POST `/admin/categories`

创建分类。

请求：

```json
{
  "name": "数据库"
}
```

响应：

```json
{
  "category": {}
}
```

### PUT `/admin/categories/:id`

更新分类。

请求：

```json
{
  "name": "后端"
}
```

响应：

```json
{
  "ok": true
}
```

### DELETE `/admin/categories/:id`

删除分类。分类下存在文章时会拒绝删除。

响应：

```json
{
  "ok": true
}
```
