# 2026-10-23 电影站升级上线

## 当前发布：手机独立布局与片源改进

用户授权提交和部署后，已发布源码 `784ed7440a80f03a13ed0f6b73fbd22a5d18393d` 到 https://tv.yuelili.com 。本轮提交包含 `07ee3ad` 手机拆分、`491752b` 正式依赖下的点击回调类型修正、`784ed74` 手机电影版本标签修正；没有 Git push。

Web/API 均更新为 `mobile-20261023-10`，因为线上旧 API 尚未包含既有搜索分组、解析和验证状态改进。Web image ID `sha256:2adb3109b118745cea0b0245c14da49374634d5565d3a75894cd2d12b196dc68`；API image ID `sha256:489d9e919a9817b15ab104f0f86213f6d7501de6aabf7b93a426debf3883efd7`。两容器 healthy、restart 0；无关容器 ID 保持不变。

从提交归档独立安装冻结依赖，批准基线与 Workspace 当前 frontendPackages 完全一致。正式依赖校验、Nuxt 类型检查、生产构建、23 项前端单测、Go 全量测试通过。候选和正式 HTTPS 的 105 个 JS 哈希与制品一致；CLI Playwright 验证桌面/手机（含保存浅色偏好）的五页、来源设置、搜索重试、详情版本、选集弹层和登录到 Account。真实来源最大资源的《流浪地球》进入 playing，无失败请求；已亲自查看实际截图。没有生产管理员账号，管理写操作、真机软键盘和 Safari 不计已验收，也不承诺所有上游来源可用。

数据库为 version 9、dirty=false，无新增迁移。原首次发布迁移包的 LF 字节摘要仍匹配已执行清单；Git 快照为 CRLF，归一比对内容相同。本次未执行 SQL、未改写数据库迁移记录，原始迁移包继续保留。公开验证码提交界面仍未接入，内部会话试点不代表完整验证码功能已开放。

发布清单 `/projects/yuelili.com/ytv/deployment-release.json`；制品与检查归档 `/projects/yuelili.com/.deploy/ytv-mobile-20261023-10/`；数据库与配置备份 `/projects/yuelili.com/backups/ytv-mobile-20261023-10/`，dump 已通过 pg_restore --list。源码 SHA256 `e9b8ae572dc3bb348d73b98042406810169dc2d4b4fa2150872697dcbc2a4c21`。本机证据 `D:/coding/yueli-official/task-artifacts/ytv/deploy-mobile-20261023-10/`。

回滚时使用备份中的 `.env`，仅重新创建 `ytv-api` 与 `ytv-web`，核验原镜像 ID、健康和正式 HTTP 后恢复旧配套清单；不自动恢复数据库。旧镜像和备份保留。本次候选容器、服务器镜像组装目录和本地 SSH 隧道已清理。

## 历史发布：电影发现页重新设计

用户要求继续改善正式站视觉后，已发布 `b195391de6f4018a1d443be0517abf31e1f3c93c`。当前 Web 为 `yueli/ytv-web:cinema-20261023-01`，image ID `sha256:61b2a2ead4ebdb76d3263742c56d18ee0fb92a4de7135226c3e77dc661194bce`。API 继续复用 `movie-20261023-01`，源码摘要、二进制摘要和运行镜像 ID 与原始清单一致；API 容器未重启，没有重新迁移数据库、初始化或认领管理员。

首页采用真实热门影片主视觉与手动切换，统一海报尺寸、字体和内容宽度；手机片库改为三列。搜索结果、详情介绍、历史、导航及观看区统一风格。电影的多个版本按版本选择呈现，不再因为存在特别版而显示“下一集”。来源仍可能部分失败，状态已压缩为不抢占结果的局部提示。

22 个单元测试、34 个不同桌面/手机浏览器场景最终通过；发现并修复继续观看记录撑宽整页、小字对比度问题，相关8项复验通过。冻结依赖、类型检查、生产构建与依赖收据通过。亲自检查本地断点及候选/正式域名的桌面、手机截图。正式 HTTPS 的103个 JS 哈希一致，海报解码、主视觉切换、搜索详情导航、真实影片播放和 Account 登录跳转通过；两个生产容器 healthy、restart 0。

当前配套清单仍位于 `/projects/yuelili.com/ytv/deployment-release.json`；本次归档 `/projects/yuelili.com/.deploy/ytv-cinema-20261023-01/`，源码 SHA256 `dfa2735665417cb98c9841b7855bebbe8af541c1085e071e4cf4962fc9b1c0b1`，Web包 SHA256 `15947fc00e80f59b66196ff11345c7146bb498537d4603cf9f5564d9d1e2aa6f`。本机截图与回执在 `D:/coding/yueli-official/task-artifacts/ytv/visual-redesign-20261023-01/`。

本次回滚备份 `/projects/yuelili.com/backups/ytv-cinema-20261023-01/` 保留旧 `.env`、私有运行配置、Compose与配套清单。需要撤销本次 UI 时，恢复该目录中的 `.env`，仅执行 `docker compose up -d --no-deps --wait ytv-web`，核验 HTTP 后恢复旧发布清单；API与数据库保持当前状态。旧Web镜像保留。

候选端口最初与其他服务冲突，改由Docker分配独立回环端口；没有停止或更改占用服务。候选容器、镜像组装目录、本地 Consumer 与 SSH 隧道均已清理；源码/正式输出/截图/备份保留。提交未 push。

## 首次整体升级的历史记录

新版 Web 与 API 已切换到 https://tv.yuelili.com 。产品源码快照为 `3931114356f039e5b605d7b19e193a506e2ad189`（以 release 清单中的完整 commit 为准）；此前包含 `402be7d` 电影播放升级和 `e31af47` 首次管理员认领。发布通过直接上传制品完成，没有 Git push。

### 首次部署版本

- API：`yueli/ytv-api:movie-20261023-01`，image ID `sha256:23df461305cd07c15e9d0b09f6bb691336abe6538bf68846d3ba65b88d83279c`。
- Web：`yueli/ytv-web:movie-20261023-02`，image ID `sha256:a0ed36c783a2538f0c7e0a1c43deb80da56ad364ec3f7b76434071d292c2ba79`。
- 生产目录 `/projects/yuelili.com/ytv`，Compose 项目 `ytv-movie`，服务 `ytv-api` / `ytv-web`。Web 仅监听宿主机 `127.0.0.1:13013`；API 仅在容器网络内访问。
- 新建独立 `ytv` 数据库与非超级用户，应用既有 1–9 迁移；没有新增持久化版本，也没有覆盖旧站数据文件。迁移原始文件摘要保存在发布清单。
- Identity 只新增 `ytv-main-web` 公共 OIDC client，正式回调 `/auth/callback`，audiences 为 `ytv-api, identity-api`。共享 Provider 未重启，其他客户端和账号未修改。

### 管理员与来源

管理员通过 [初始化入口](https://tv.yuelili.com/manage/setup) 登录后显式点击认领。首次成功后入口永久关闭；普通登录不会授予管理员。部署未代替用户认领，也未预设管理员账号。

原有 10 个来源配置完整备份。9 个 HTTPS 定义迁入，5 个目录探测可用并启用，4 个禁用。Omo 的 HTTP 内网地址保留在旧配置中，不放宽新版允许列表。播放与图片主机采用真实元数据和现有 YTV 来源配置的允许主机族；未代理或保存第三方视频。

来源可用性会波动。此次“流浪地球”搜索 3 个来源成功、2 个失败，页面保留已获得的结果并允许重试/换源；最大资源的真实影片播放成功。没有把目录成功等同于全部影片、所有来源均可播放。

### 首次验收与证据

Go 全量测试、19 个前端单测、冻结依赖/typecheck/Nitro build 通过；HTTP Project 合同生成检查覆盖 28 个操作。首次认领在独立本地数据库完成真实 OIDC 登录、显式点击、刷新和 API/Web 实际重启后的持久性验证，匿名请求拒绝，重复认领及他人认领冲突测试通过。

候选和正式域名均验证 API/Web 健康、搜索、目录、详情、匿名后台门禁与首次认领 401；104 个前端 JS 经 HTTP 获取后与本地制品摘要一致。正式域名 CLI Playwright 验证桌面和手机页面、海报解码、搜索→详情→播放导航、真实媒体 `playing`、跳转 Account 登录表单；无 pageerror。已亲自查看实际截图并修正保存浅色偏好时的公开深色界面对比度。正式两容器 healthy、RestartCount=0。

生产管理员写操作留给站点所有者认领后执行；本次没有生产管理员会话。iOS Safari、投屏、所有上游源和字幕/音轨媒体不属于本次生产验收。首次部署时电影版本与剧集标签按来源条目结构呈现；当前前台已修正电影版本判定，仍未建设跨源统一影片身份目录。

发布清单：`/projects/yuelili.com/ytv/deployment-release.json`。完整发布资料：`/projects/yuelili.com/.deploy/ytv-movie-20261023-01/`。本机截图、清单和验证输出：`D:/coding/yueli-official/task-artifacts/ytv/deploy-20261023-01/`。源码包 SHA256 `b0b1df3b40b0159710d72dfc849e86d8029261cd6f6b2332c117fc8a3b65f136`；Web 包 SHA256 `58e865c514a4477c1e358f79df6efe6466c39f5d4f26f128b37bf5fc402bf313`；API 二进制 SHA256 `4207767820f161b33e70ff0e9bfa9db156772a184f247f776ca659a3de4aa33a`。

### 整体旧站回滚与收尾

备份位于 `/projects/yuelili.com/backups/ytv-movie-20261023-01/`，包括旧配置、旧运行信息、Identity 注册前 dump、新 YTV 初始 dump、代理配置和新运行配置，归档与数据库目录可读性均已核验。

旧容器 `ytv` 已停止，旧镜像 `sha256:2ddf8d7b45cb67d84d79ea202eb7948085b2e95c70a416057445f8be709c826d` 与 `docker-compose.legacy-20261023.yml` 保留。若回滚，先启动该旧容器并确认原 9003 入口，再仅恢复本站 `tv-proxy.conf`、执行 nginx 配置检查和 reload；确认旧站可用后停止新版两个服务。不自动恢复或删除数据库。

服务器临时镜像构建目录已清理；本地临时 Consumer 已由 Workspace 停止，SSH 旁路隧道已关闭，共享 Identity Provider 保留。其他工作的未提交测试清理和 vendor 包未纳入本轮提交。
