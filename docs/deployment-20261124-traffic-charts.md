# 2026-11-24 影视统计图表部署

后台访问趋势采用共享图表，横轴显示简短日期，选中数据点时显示完整日期和数值。

| 服务 | 版本 | 运行镜像 |
| --- | --- | --- |
| Web | `traffic-charts-20261124-1` | `sha256:f495c312ed96184fb09a721c7932ca2c81d0244bcccd4838337e936f5d0394fe` |
| API | `site-launcher-20261123-1` | `sha256:da80b98c84442668e0380ed8d81d28a8aa7697574b89e765a3b098cdec562ab1` |

候选和公网后台用浏览器测试数据检查桌面与手机宽度、键盘选点和完整日期。公网加载的 60 个脚本及样式与发布制品哈希一致。此轮切换 `ytv-web`，复用原 API 和数据库。

生产回执位于 `/projects/yuelili.com/ytv/deployment-release.json`。发布制品与验证证据在 `/projects/yuelili.com/.deploy/ytv-traffic-charts-20261124-1/`。恢复配置与前版回执在 `/projects/yuelili.com/backups/ytv-traffic-charts-20261124-1/`，前版 Web 镜像为 `yueli/ytv-web:mobile-navigation-20261124-1`。
