# 2026-11-24 影视移动导航部署

手机底栏提供「首页」「搜索」「我的」。[「我的」](https://tv.yuelili.com/me)是独立页面，站群九宫格留在右上角。播放和影片沉浸页隐藏底栏。「我的」提供播放历史和来源设置。

| 服务 | 版本 | 运行镜像 ID |
| --- | --- | --- |
| Web | `mobile-navigation-20261124-1` | `sha256:540616f69f73979a1214ecb950d6395b52b00839bf9b5490029446cdd605714b` |
| API | `site-launcher-20261123-1` | `sha256:da80b98c84442668e0380ed8d81d28a8aa7697574b89e765a3b098cdec562ab1` |

候选通过手机与桌面浏览器检查。公网复查了九宫格、手机导航和「我的」页面，62 项脚本及样式摘要与制品一致。本次切换 `ytv-web`，复用原 API，数据库无需迁移。

当前回执为 `/projects/yuelili.com/ytv/deployment-release.json`，制品和验证证据在 `/projects/yuelili.com/.deploy/ytv-mobile-navigation-20261124-1/`。恢复配置和前版回执位于 `/projects/yuelili.com/backups/ytv-mobile-navigation-20261124-1/`，前版 Web 镜像为 `yueli/ytv-web:site-launcher-20261123-1`。恢复时使用备份的 Web 镜像配置重建 `ytv-web`，核对公网后恢复对应回执。
