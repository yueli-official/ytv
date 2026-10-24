# 2026-10-23 电影站升级上线

## 超分辨率修复发布（2026-10-25）

已部署 `ytv-super-resolution-20261025-12`，源码提交 `46a9fd8`（包含 `f3da9c0`）。统一控件为“超分辨率：开启 / 关闭”，失败为“超分辨率失败”；默认开启、浏览器保存偏好与已批准的失败保留原画行为不变。

用户提供的电影天堂 `dytt/84586/3` 实测为1920×1080，旧CNNx2M确实运行，但只有放大，且小窗口固定生成4K再缩小。现改为 Anime4K-WebGPU Mode A，先修复线条，再按显示尺寸及屏幕像素密度决定放大；最终呈现使用线性采样。对照[上游 Kazumi 着色器配置](https://github.com/Predidit/Kazumi/blob/main/lib/utils/constants.dart)确认其包含 Restore 阶段，网页采用相近阶段，不宣称与mpv渲染结果相同。同一人物画面600秒处人工检查眼睑、头发边缘更清楚。另修复全屏皮肤区域与普通host尺寸不同，画布现在读取实际区域并在进出全屏时重建。

冻结依赖、批准依赖基线、类型检查、生产构建和配套发布门禁通过。候选和正式站的108个JS HTTP摘要均匹配。CLI Playwright覆盖桌面1440/DPR1、手机视口390/DPR3、真实该集600秒后连续GPU帧、关闭销毁画布、刷新保存关闭、再次开启和进出全屏，无pageerror。正式站8秒观测约23.6/23.9处理帧每秒，GPU完成等待均值约18.4/18.3ms；这是本机同一GPU的两种视口，不是实体手机基准。桌面画布由1009×568调整为全屏1440×810，手机模拟1159×652调整为1170×658。实际候选/正式桌面及手机截图已人工检查；没有用自动截图比较代替视觉判断。未测实体手机、Safari、所有片源或整集持续性能。

只替换Web，镜像 `yueli/ytv-web:super-resolution-20261025-12`，image ID `sha256:b357ef63eadea28717d9b0ec0f883f7787bafb65fafbe6e22705403933589f60`。API源码、镜像、二进制摘要及容器ID核验与上轮一致，未重启。两端healthy/restart0；20个来源、评价和私有运行配置不变，v11 clean，无迁移。备份已验证可读。

配套清单 `/projects/yuelili.com/ytv/deployment-release.json`，备份 `/projects/yuelili.com/backups/ytv-super-resolution-20261025-12`；证据 `D:/coding/yueli-official/task-artifacts/ytv/super-resolution-20261025-12/production-release.json`、`production-browser.json`、`production-http.json`及`production-mobile.png`。候选容器、远端组装目录、SSH13626隧道与临时比较服务已停止/清理；本次E盘依赖和编译缓存回收状态见同目录`cleanup.json`，共享开发服务占用的原生硬链接不强删，不停共享服务。回滚只恢复本次备份.env并重建ytv-web、核验旧Web和原API；不恢复数据库。代码和发布记录均本地提交，未push。

## 搜索合并、缓存与分页发布（2026-10-25）

已发布 `ytv-search-cache-20261025-11` 到 https://tv.yuelili.com。Web/API 均从完整提交 `65766bafd1a0a53d5fbb8412cbd6e8fe66196f71` 构建；此前搜索修复提交 `c8ab8a9`，其余代码、研究文档、测试与依赖已一并提交。未 Git push。正式共享前端依赖已同步 Workspace 当前批准基线 `32339ab7339bbcccbcecbb920cd556e5d689b309b906e2ee6979ba83b3de78ad`。

同名结果合并并沿用已有图片和简介；浏览器缓存最近20次查询和分页，只有“强制刷新搜索”主动失效本查询缓存；首屏未读取缓存时显示加载，不误报无结果。JSON来源改为搜索取得 ID、再批量读取详情，保留搜索分页；修复优质资源库详情接口忽略关键词导致全站5376页混入搜索。画质增强失败提示和允许跨域取帧的直链加载修正同时纳入，仍不代理媒体。

Go全量测试、Linux构建、冻结依赖校验、类型检查、23项前端测试、生产构建和双端门禁通过。候选与正式均完成108个JS HTTP摘要核对；桌面暗色、手机保存浅色偏好的真实搜索、2页翻页、缓存刷新零请求、强制刷新、无错误空状态、历史、登录门禁及非凡真实1920×1080播放通过；实际桌面/手机截图人工复核。候选初次自动跳转正式域名，已修正候选 OIDC origin 后重验；手机补验修正临时脚本编码并通过键盘暂停，不以强制点击规避控制栏自动隐藏。未做实体手机/GPU性能或所有片源稳定性承诺。

数据库v11 clean，原迁移字节逐项保留，无新增迁移；dump已用pg_restore --list验证可读。启用来源仍20个，用户评价和私有配置未改。生产API镜像ID `sha256:058a8c3b2f562603c487905f2727e388aa20644bfc151ebc484418843a57a3c5`，Web镜像ID `sha256:b4554094bd525bc41e9cf257603e33a27d4d461f7e72a05d59b98b2e06dc31e2`，两端healthy/restart0。

配套清单 `/projects/yuelili.com/ytv/deployment-release.json`；备份 `/projects/yuelili.com/backups/ytv-search-cache-20261025-11`；本机制品与证据 `D:/coding/yueli-official/task-artifacts/ytv/search-cache-20261025-11/`。候选容器、远端镜像组装目录和SSH隧道已清理；本任务前端依赖与编译缓存按E盘任务约定回收。观察到另一个Gallery发布替换其容器，已在清单记录，本发布只切换YTV双端。回滚仅恢复备份.env并重建ytv-api/ytv-web，核验旧配套镜像；不自动恢复数据库。

## 播放地址限制修复发布（2026-10-24）

已发布 `ytv-playback-policy-20261024-06`。以线上 `source-expansion-20261024-05` 源码归档为基准，仅替换播放 resolver 和对应已有测试，移除媒体域名/IP/HTTP授权限制；保留地址格式与HTTP/HTTPS协议校验。其余本地未发布改动未纳入。没有commit/push。

API全套测试与Linux构建通过；使用Workspace当前导出的基线通过配套发布门禁。Web源码逐文件摘要与旧制品一致，复用原镜像和容器。候选/正式均验证8条原409解析恢复、桌面暗色非凡和手机亮色电影天堂实际播放及180秒拖动、106份JS哈希。8条正式媒体另外完成24张抽帧人工复查；[审核表](research/online-source-five-audit-20261024.md)已更新。

数据库v11 clean，备份已验证可读，没有迁移、来源配置或用户评价改动。仅切换YTV API，其他原有容器ID不变。生产清单 `/projects/yuelili.com/ytv/deployment-release.json`，备份 `/projects/yuelili.com/backups/ytv-playback-policy-20261024-06`，本机证据 `D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-policy-20261024-06`。候选容器、镜像组装目录和SSH隧道已清理。

## 手机播放器适配发布（2026-10-23）

用户授权部署后，已上线 `ytv-player-20261023-14`：https://tv.yuelili.com。Web 更新为 `yueli/ytv-web:player-20261023-14`；API 复用 `yueli/ytv-api:playback-20261023-12`，源码摘要、二进制摘要、运行镜像和容器 ID 均核验一致，没有重启 API 或迁移数据库。数据库 v9 clean，发布前备份可读。没有 commit/push。

手机全屏加入 landscape 方向锁请求、退出解锁，保留原生浏览器回退；横屏继续使用手机布局并保留同一视频元素，辅助操作压缩为一排。正式冻结依赖、Workspace 当前基线核验、类型检查、生产构建及双端发布门禁通过。归档 manifest/基线与工作区差异仅换行符，分别对齐当前 manifest 和 Workspace 原始导出字节，未更改版本或批准内容。

候选与正式 HTTPS 的 105 个 JS 摘要匹配。CLI Playwright 覆盖 1440px 桌面、390px 手机、暗色/保存浅色偏好、搜索历史与清空、Account 登录表单跳转；真实 7sefun 31667-1-1 均进入 playing。手机触摸模拟验证全屏/退出方向调用与旋转保留视频节点，实际竖屏、横屏全屏和桌面截图已亲自检查。首次候选脚本在 ready 前点击导致等待超时，修正脚本等待 ready 后通过；未修改产品代码绕过问题。实体手机/iOS Safari 的旋转未验证。

生产清单 `/projects/yuelili.com/ytv/deployment-release.json`；备份 `/projects/yuelili.com/backups/ytv-player-20261023-14`。Web image ID `sha256:9cdd226edf7a8767d431553ad32e74f4feca4fccffac794b7d622244bdd7b52c`，两端 healthy。证据 `D:/coding/yueli-official/task-artifacts/ytv/deploy-player-20261023-14`。候选预选端口被占用，未动占用服务，改为 Docker 分配回环 32772；候选和远端组装目录已清理，SSH 隧道已停止。

部署期间观察到独立 Paste 发布替换了 paste-paste-web-1，已核验其自身清单与新镜像一致并记录；本次脚本只切换 YTV Web，保留其他任务状态。回滚仅恢复本次备份 .env 后重建 ytv-web，并核验旧配套清单；不恢复数据库。


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


## 2026-10-23 播放修复与搜索历史发布完成

发布 `ytv-playback-20261023-12` 已部署到 https://tv.yuelili.com。来源为 HEAD `6b154e8` 加发布清单记录的 7 个未提交文件；本轮没有 commit/push。已批准 Kazumi 解析器信任其 HTTP/HTTPS 媒体结果，不再逐个登记 CDN。另修复播放页 Referer：上游在站内请求下返回不可播放内容，使用页面 no-referrer 后实际播放恢复。搜索历史改为纵向列表，公开站细滚动条。

Go 全量测试、正式依赖门禁、类型检查、生产构建通过。候选和正式站都经 CLI Playwright 运行，1440 桌面/390 手机视口，暗色和保存浅色偏好，确切 7sefun 31667-1-1 进入 playing，截图亲自复核；历史长文本无横向溢出，Escape/清空正常；登录到 Account 表单正常。105 个 JS 哈希与构建一致，双端镜像健康，API 二进制摘要匹配；没有验证实体手机/Safari、整集播放或所有片源。

数据库保持 v9，无迁移；备份验证可读。生产清单 `/projects/yuelili.com/ytv/deployment-release.json`，备份 `/projects/yuelili.com/backups/ytv-playback-20261023-12`。候选容器和远端镜像组装目录已清理，无关容器未变化。完整证据在 `D:/coding/yueli-official/task-artifacts/ytv/deploy-playback-20261023-12`。本机构建中断后重启恢复，正式构建最终成功。
