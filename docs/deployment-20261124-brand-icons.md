# 2026-11-24 影视品牌图标部署

影视站采用新版金色图标和公开页面金黄色。共享九宫格显示各站的官方图标及原色。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `brand-icons-20261124-1` | `sha256:f21bed859c73da5cc4a8b12e0c28214fc6e68c1cc597aeece244acc47dfe1004` |
| API | `site-launcher-20261123-1` | `sha256:da80b98c84442668e0380ed8d81d28a8aa7697574b89e765a3b098cdec562ab1` |

候选和[生产域名](https://tv.yuelili.com/)通过桌面与手机浏览器检查，63 项脚本和样式摘要与 Web 制品一致。SVG、ICO 和 iPhone 主屏幕图标的公网响应与源码逐字节一致。管理图表使用测试数据验证。API 复用现有镜像，数据库无需迁移。

服务器回执位于 `/projects/yuelili.com/ytv/deployment-release.json`。制品和验证证据位于 `/projects/yuelili.com/.deploy/ytv-brand-icons-20261124-1/`。恢复配置和上一版回执位于 `/projects/yuelili.com/backups/ytv-brand-icons-20261124-1/`，前版 Web 镜像为 `yueli/ytv-web:traffic-charts-20261124-1`。
