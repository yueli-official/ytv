# YTV SPA 发布与内存对照（2026-11-14）

YTV 已部署 Web `spa-20261114-2`，公开页和管理页使用 SPA，保留现有页面样式、播放器及 Nitro 服务端认证、BFF、海报处理。Go API 沿用 `search-cache-20261025-11`，容器未重建；未迁移数据库或重启其他站点。用户确认 AGENTS.md、Readme.md 和发布校验器维护源注释一并纳入来源快照；本次没有 commit 或 push。

## 内存结果

| 同条件 HTML 访问负载 | 原 SSR | 最终 SPA |
| --- | ---: | ---: |
| 预热后容器工作内存 | 96.54 MiB | 40.71 MiB |
| 访问后空闲工作内存 | 142.49 MiB | 40.91 MiB |
| cgroup 记账峰值 | 146.96 MiB | 41.44 MiB |
| Node RSS 采样峰值 | 191.68 MiB | 90.21 MiB |
| HTML 响应中位数 | 218.37 ms | 5.29 ms |

访问后空闲工作内存降低 **101.59 MiB，约71.29%**。使用相同生产环境变量、原 Node 24.22.1 运行时、无新增内存上限；两边均注入同一内存采样器。路径为首页、历史、搜索、影片详情、播放页：10次预热，再15次并发2的GET，空闲20秒。工作内存取 `memory.current - inactive_file`，峰值为 cgroup 原始 `memory.peak`；RSS是采样值，口径不同，不能相加。

该负载测的是服务端 HTML。SPA 随后的 API、脚本和图片请求发生在浏览器，因此 HTML 延迟不能解释为完整页面显示耗时。正式站浏览器访问期间90秒、4次采样为55.52、58.23、59.82、60.55 MiB，Swap为0；cgroup峰值最高77.75 MiB。历史长运行旧容器曾约243 MiB并有Swap，负载和运行时间不同，不用它计算优化比例。短期结果不代表长期稳定上限。

## 实现与验证

- `ssr: false`，增加协调现有深色主题的轻量加载壳；内容加载仍由现有客户端业务完成。服务器运行 Nitro，尚未改成纯静态托管。
- 采用批准基线的共享 UI 0.10.0 / Identity Nuxt 0.2.4 正式制品。原本机源码覆盖未用来构建。原API源码与生产归档核对，仅换行差异；制品复用来源及二进制摘要通过校验。
- 首版候选发现 Nuxt UI 4.9.0 在 Nuxt 4.5.1 / Unhead 3 的纯CSR路径调用缺失的 `hooks.hookOnce`。首版未上线，产品依赖升级到4.10.0修正主路径，无临时兼容层。上游证据：[问题6687](https://github.com/nuxt/ui/issues/6687)、[修复6658](https://github.com/nuxt/ui/pull/6658)、[4.10.0发布](https://github.com/nuxt/ui/releases/tag/v4.10.0)。
- 冻结依赖安装、类型检查、9个既有测试文件23项测试、正式构建及前后端发布门禁通过；SPA服务端包约2.52 MB。
- 原版浏览器10项、最终候选28项、正式站26项检查通过。覆盖1440/390视口与明暗偏好、搜索、来源设置、详情深链刷新、本地历史、登录和管理门禁。公开页按既有设计保持深色主题。
- 电脑和手机视口真实播放 `ffzy/1629/line1-0`，视频1280×504、未暂停、播放时间实际推进。候选另有模拟503片源失败后恢复检查。没有代理、下载或转码视频。
- 最终候选8项及正式站8项加载完成截图检查通过，人工查看桌面首页、手机详情和实际播放画面。功能截图中的早期骨架状态不作为最终视觉验收。
- 正式首页/历史/搜索/详情/播放/管理路径返回SPA壳；健康和公共API返回200；119个JS/CSS资源摘要匹配。90秒观察中持续healthy、restart 0、无OOM；同一shim的FD稳定24、PID文件稳定1。

未使用真实管理员身份执行管理写入或验证已登录管理刷新；本次使用生产等价候选及真实正式域名/API/Identity，没有启动本地Workspace组合。手机是浏览器视口验收，未测实体手机、GPU或长时间播放。浏览器拦截业务写请求，不提交评价或进度。上述范围不能宣称完整管理员或长期压力验收。

## 制品、恢复与证据

- Web镜像：`yueli/ytv-web:spa-20261114-2`，ID `sha256:e422b3a396a2d5e3ae2aefff10da7014eaf0109253baac0c960b4f97d2c83613`。
- 来源归档SHA256：`70aacdcf71d23bac99e4334875227fd869488b577254ee15ac913a9198e520d6`。
- Web归档SHA256：`a43fce90da2ff5df95b304a0499ff4eb3b5f3e98ab08f444c46627e34178d60c`。
- 生产回执：`/projects/yuelili.com/ytv/deployment-release.json`；任务回执与归档：`/projects/yuelili.com/.deploy/ytv-spa-20261114-1/`。
- 本机归档、截图和日志：`D:/coding/yueli-official/task-artifacts/ytv/spa-20261114-1/`。机器可读结果见[验证摘要](deployment-20261114-spa-verification.json)。
- 恢复资料：`/projects/yuelili.com/backups/ytv-spa-20261114-1/`；旧Web镜像 `super-resolution-20261025-12` 保留。必要时先核对当前配置及回执，再恢复备份的 `.env` 中旧Web镜像，执行 `docker compose --env-file .env -f docker-compose.yml -p ytv-movie up -d --no-deps --no-build --pull never ytv-web`。本次无数据库变更，不恢复数据库。

发布仅切换YTV Web，环境变量、端口及内存配置保持。切换连同初步检查22.78秒，并非实测停机时长。候选、隧道、未发布首版镜像和可再生产物已回收；保留源码/运行归档、截图、必要日志和受限权限备份。发布后新增的本报告及AGENTS中SPA约束是文档记录，不改写已发布来源快照。
