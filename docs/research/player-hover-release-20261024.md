# 播放器悬停灰罩修复与发布（2026-10-24）

## 原因

线上仍使用 `ytv-web:guest-reviews-20261024-03`。后续两次来源配置发布复用了该前端，未包含本地去遮罩修改。
Video.js Neutral 皮肤的 `video-controls-backdrop` 覆盖整个视频，默认渐变黑色透明度为 20%～50%，控件出现时把全画面压暗。

## 修改

使用 NeutralVideoSkinElement 的模板扩展入口注册 `ytv-video-skin`，在皮肤容器上设置 `--media-controls-gradient: none`，保留原控件和播放能力。仅在外层设置同名变量会被皮肤内部主题覆盖，因此必须在真正拥有该变量的容器设置。

## 发布与证据

- 正式版本：`ytv-hover-20261024-09`；Web 镜像 `yueli/ytv-web:hover-20261024-09`。
- API 保持 `playback-policy-20261024-06`；数据库 v11 clean；优选 20 源及推荐配置保持不变。
- 基于已发布归档仅替换播放器组件、皮肤定义、Nuxt custom element 配置；未混入其他工作区修改。
- API 127 个源文件与原制品清单逐项相同；旧清单哈希按原 JSON 字节校验，不能混用另一种 JSON 序列化后的哈希。
- 正式类型检查、构建、当前 Workspace 依赖基线与配套发布门禁通过。
- 候选与正式域名 106 个 JavaScript 文件 HTTP SHA256 全部匹配；API 就绪、匿名权限检查通过。
- CLI Playwright 在桌面 dark / 手机尺寸 light 下用真实《食神》媒体验证播放、暂停、静音、进度操作、进入及退出全屏，无 pageerror。需等待媒体 readyState 就绪再点击，不以初始按钮出现代替媒体已就绪。
- 运行时诊断：背景图 `none`，背景色 `rgba(0, 0, 0, 0)`。这是样式诊断，非自动视觉判分。
- 实际浏览器画面人工复看：候选播放中控件显示、暂停控件显示，以及正式站暂停控件显示均无整屏灰罩；控件、字幕可见。手机仅模拟尺寸功能验收，未宣称实体手机或手机视觉验收。
- 其他容器未变化；候选容器和服务端临时构建目录已清理。

本机证据：`D:/coding/yueli-official/task-artifacts/ytv/hover-20261024/`，包含新旧运行时样式、候选/正式功能结果、HTTP 哈希、构建日志与发布清单。服务器备份为 `/projects/yuelili.com/backups/ytv-hover-20261024-09`。
