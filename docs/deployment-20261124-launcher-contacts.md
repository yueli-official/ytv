# 2026-11-24 站群入口与联系方式部署

页头九宫格放在账号入口左侧，已有主题按钮时放在主题按钮右侧。共享菜单保留站点官方图标，在底部提供 QQ 群、GitHub 和哔哩哔哩入口。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `launcher-contacts-20261124-1` | `sha256:901d2a4e4faad557fa97d9f0497206e51ec0e5fefbbbd38542cb7765af0ad728` |
| API | `site-launcher-20261123-1` | `sha256:da80b98c84442668e0380ed8d81d28a8aa7697574b89e765a3b098cdec562ab1` |

[正式站点](https://tv.yuelili.com/)的桌面和手机浏览器检查通过，62 项静态资源摘要与发布制品一致。共享 UI 使用 `0.10.4-preview.20261124.contacts.1`。API 复用原镜像，数据库无需迁移。

生效回执位于 `/projects/yuelili.com/ytv/deployment-release.json`，制品与核验记录位于 `/projects/yuelili.com/.deploy/ytv-launcher-contacts-20261124-1/`。恢复配置及旧回执位于 `/projects/yuelili.com/backups/ytv-launcher-contacts-20261124-1/`，上一版 Web 镜像为 `yueli/ytv-web:brand-icons-20261124-1`。
