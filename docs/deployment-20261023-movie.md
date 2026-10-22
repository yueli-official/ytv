# 2026-10-23 电影站升级上线

新版 Web 与 API 已切换到 https://tv.yuelili.com 。产品源码快照为 `3931114356f039e5b605d7b19e193a506e2ad189`（以 release 清单中的完整 commit 为准）；此前包含 `402be7d` 电影播放升级和 `e31af47` 首次管理员认领。发布通过直接上传制品完成，没有 Git push。

## 运行版本

- API：`yueli/ytv-api:movie-20261023-01`，image ID `sha256:23df461305cd07c15e9d0b09f6bb691336abe6538bf68846d3ba65b88d83279c`。
- Web：`yueli/ytv-web:movie-20261023-02`，image ID `sha256:a0ed36c783a2538f0c7e0a1c43deb80da56ad364ec3f7b76434071d292c2ba79`。
- 生产目录 `/projects/yuelili.com/ytv`，Compose 项目 `ytv-movie`，服务 `ytv-api` / `ytv-web`。Web 仅监听宿主机 `127.0.0.1:13013`；API 仅在容器网络内访问。
- 新建独立 `ytv` 数据库与非超级用户，应用既有 1–9 迁移；没有新增持久化版本，也没有覆盖旧站数据文件。迁移原始文件摘要保存在发布清单。
- Identity 只新增 `ytv-main-web` 公共 OIDC client，正式回调 `/auth/callback`，audiences 为 `ytv-api, identity-api`。共享 Provider 未重启，其他客户端和账号未修改。

## 管理员与来源

管理员通过 [初始化入口](https://tv.yuelili.com/manage/setup) 登录后显式点击认领。首次成功后入口永久关闭；普通登录不会授予管理员。部署未代替用户认领，也未预设管理员账号。

原有 10 个来源配置完整备份。9 个 HTTPS 定义迁入，5 个目录探测可用并启用，4 个禁用。Omo 的 HTTP 内网地址保留在旧配置中，不放宽新版允许列表。播放与图片主机采用真实元数据和现有 YTV 来源配置的允许主机族；未代理或保存第三方视频。

来源可用性会波动。此次“流浪地球”搜索 3 个来源成功、2 个失败，页面保留已获得的结果并允许重试/换源；最大资源的真实影片播放成功。没有把目录成功等同于全部影片、所有来源均可播放。

## 验收与证据

Go 全量测试、19 个前端单测、冻结依赖/typecheck/Nitro build 通过；HTTP Project 合同生成检查覆盖 28 个操作。首次认领在独立本地数据库完成真实 OIDC 登录、显式点击、刷新和 API/Web 实际重启后的持久性验证，匿名请求拒绝，重复认领及他人认领冲突测试通过。

候选和正式域名均验证 API/Web 健康、搜索、目录、详情、匿名后台门禁与首次认领 401；104 个前端 JS 经 HTTP 获取后与本地制品摘要一致。正式域名 CLI Playwright 验证桌面和手机页面、海报解码、搜索→详情→播放导航、真实媒体 `playing`、跳转 Account 登录表单；无 pageerror。已亲自查看实际截图并修正保存浅色偏好时的公开深色界面对比度。正式两容器 healthy、RestartCount=0。

生产管理员写操作留给站点所有者认领后执行；本次没有生产管理员会话。iOS Safari、投屏、所有上游源和字幕/音轨媒体不属于本次生产验收。电影版本与剧集标签仍按来源条目结构呈现，尚不是跨源统一的影片身份目录。

发布清单：`/projects/yuelili.com/ytv/deployment-release.json`。完整发布资料：`/projects/yuelili.com/.deploy/ytv-movie-20261023-01/`。本机截图、清单和验证输出：`D:/coding/yueli-official/task-artifacts/ytv/deploy-20261023-01/`。源码包 SHA256 `b0b1df3b40b0159710d72dfc849e86d8029261cd6f6b2332c117fc8a3b65f136`；Web 包 SHA256 `58e865c514a4477c1e358f79df6efe6466c39f5d4f26f128b37bf5fc402bf313`；API 二进制 SHA256 `4207767820f161b33e70ff0e9bfa9db156772a184f247f776ca659a3de4aa33a`。

## 回滚与收尾

备份位于 `/projects/yuelili.com/backups/ytv-movie-20261023-01/`，包括旧配置、旧运行信息、Identity 注册前 dump、新 YTV 初始 dump、代理配置和新运行配置，归档与数据库目录可读性均已核验。

旧容器 `ytv` 已停止，旧镜像 `sha256:2ddf8d7b45cb67d84d79ea202eb7948085b2e95c70a416057445f8be709c826d` 与 `docker-compose.legacy-20261023.yml` 保留。若回滚，先启动该旧容器并确认原 9003 入口，再仅恢复本站 `tv-proxy.conf`、执行 nginx 配置检查和 reload；确认旧站可用后停止新版两个服务。不自动恢复或删除数据库。

服务器临时镜像构建目录已清理；本地临时 Consumer 已由 Workspace 停止，SSH 旁路隧道已关闭，共享 Identity Provider 保留。其他工作的未提交测试清理和 vendor 包未纳入本轮提交。
