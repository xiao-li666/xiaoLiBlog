# Docker 部署步骤

这份文档适用于当前博客项目，部署后对外访问端口是 `5555`。

架构如下：

- MySQL 8.4
- Redis 7
- Go 后端容器
- Vue + Nginx 前端容器
- 前端 Nginx 统一转发 `/api` 和 `/uploads`

## 1. 服务器准备

```bash
ssh root@你的服务器IP
apt update
apt install -y git curl ufw
```

安装 Docker 和 Compose：

```bash
curl -fsSL https://get.docker.com | sh
docker --version
docker compose version
```

放行端口：

```bash
ufw allow OpenSSH
ufw allow 5555
ufw allow 443
ufw enable
ufw status
```

如果你暂时不用 HTTPS，`443` 可以先不管；真正对外访问会走 `5555`。

## 2. 拉取项目

```bash
cd /opt
git clone 你的仓库地址 personal-blog
cd /opt/personal-blog
```

## 3. 创建生产环境变量

在项目根目录创建 `.env.prod`：

```env
MYSQL_ROOT_PASSWORD=换成强密码
MYSQL_DATABASE=personal_blog
MYSQL_USER=blog_user
MYSQL_PASSWORD=换成强密码

JWT_SECRET=换成很长的随机字符串
FRONTEND_ORIGIN=http://你的域名或IP:5555

SMTP_HOST=smtp.qq.com
SMTP_PORT=465
SMTP_USERNAME=你的邮箱
SMTP_PASSWORD=邮箱授权码
SMTP_FROM=你的邮箱

DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
DEEPSEEK_TIMEOUT_SECONDS=30
```

说明：

- 不需要手工写 `MYSQL_DSN`，`docker-compose.prod.yml` 会自动拼接。
- `REDIS_ADDR` 也不需要写，容器里默认连 `redis:6379`。
- 第一次注册的用户会自动成为站长。

## 4. 构建并启动

```bash
cd /opt/personal-blog
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
```

## 5. 检查状态

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml ps
docker compose --env-file .env.prod -f docker-compose.prod.yml logs --tail=200
docker compose --env-file .env.prod -f docker-compose.prod.yml logs -f backend
```

## 6. 访问方式

- 前台：`http://你的域名或IP:5555`
- 后端健康检查：`http://你的域名或IP:5555/api/health`

上传文件会落在后端挂载卷里，不会因为容器重建丢失。

## 7. 常用维护命令

重启：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml restart
```

重新构建：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
```

停止：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml down
```

查看日志：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml logs -f
```

## 8. 更新部署

```bash
cd /opt/personal-blog
git pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
```

## 9. HTTPS

当前方案默认先跑 `5555`。如果后面要上 HTTPS，可以：

- 在云服务器安全组放行 443
- 给前置 Nginx 加证书
- 或者再套一层外部反向代理

如果你要，我可以继续给你补一版“`5555 + HTTPS`”的完整部署方案。
