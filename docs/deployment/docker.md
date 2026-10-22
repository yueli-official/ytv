# Docker 部署

生产 Compose 包含 PostgreSQL、一次性迁移任务、Go API 和 Nuxt Web 四个容器。只有 Web 发布宿主机端口。

Chromium 已安装在 API 镜像内，宿主机不需要安装 Chrome。它以非 root 用户运行，只负责为 WebView 类来源捕获媒体 URL；
普通来源不会使用浏览器。API 停止时 Chromium 子进程一并退出。

## 启动

```bash
cp .env.docker.example .env
# 替换所有 replace-with-* 和 example.com
docker compose build
docker compose up -d
docker compose ps
```

默认访问 http://localhost:9000。正式入口的 Origin 必须与 OIDC Redirect URI、Post Logout Redirect URI 和
NUXT_PUBLIC_YTV_BASE 完全一致。

## BrowserCapture

- 最大并发解析页面：2
- 单次解析超时：25 秒
- 共享内存：512 MiB
- API 非 root、丢弃 capabilities、禁止提权
- Chromium 的 no-sandbox 仅在容器内启用；进程隔离由 Docker 提供

只使用确定性来源时，可设置 YTV_BROWSER_CAPTURE_ENABLED=false。建议为 API 至少预留 1 CPU 和 768 MiB 内存。

## 数据和升级

PostgreSQL 数据保存在 ytv-postgres volume。migrate 容器使用 golang-migrate，只执行尚未应用的 up migrations；
迁移成功后 API 才会启动。

```bash
docker compose build
docker compose up -d
```

升级前备份数据库。不要修改已经执行过的迁移文件。

## Identity

Compose 不复制 Identity。YTV_JWKS_URL、YTV_JWKS_ISSUER、NUXT_IDENTITY_BASE 和所有 OIDC 公开地址必须指向已部署的
中央 Identity/Account。生产环境必须使用稳定的 NUXT_SEAL_SECRET，并保持 NUXT_COOKIE_SECURE=true。

首次上线可不设置 YTV_BOOTSTRAP_ADMIN_SUBS。登录后访问 /manage/setup，显式点击初始化即可成为首位管理员；已认领站点不会重新开放。正式共享服务器使用 [生产发布入口](../../deploy/README.md)。
