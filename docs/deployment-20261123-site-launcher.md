# 2026-11-23 影视九宫格部署

[影视](https://tv.yuelili.com/)已更新至 `site-launcher-20261123-1`。页头已采用共享九宫格。九个入口以浅色底和彩色图标呈现。站点 Logo 和 favicon 可独立配置，公开接口 `GET /api/v1/site/branding` 返回当前地址。

- 桌面和手机公网浏览器均通过九宫格验证，44 项脚本与样式的摘要与制品一致。两端运行镜像已核对，重启次数为 0。
- 执行了 `0012_site_branding`。数据库迁移为 12、clean。
- API 镜像 `sha256:da80b98c84442668e0380ed8d81d28a8aa7697574b89e765a3b098cdec562ab1`，Web 镜像 `sha256:14db684e1e1ecb3a139cc868b87daf05862dfed8e6895256c15eea5020451b86`。
- 源码快照 `5ce553b8afe844337fb83ccc59d6f47c9c21176210f0e454c952cb38eef21816`。双端验收回执见 `/projects/yuelili.com/ytv/deployment-release.json`，恢复配置见 `/projects/yuelili.com/backups/ytv-site-launcher-20261123-1`。
