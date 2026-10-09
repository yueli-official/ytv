# 2026-10-09 正式发布准备

用户明确纳入 旧版迁移 范围，Distribution 排除。本文件记录准备证据，不表示已经上线。

域名候选：tv.yuelili.com（Nav/YTV 保留原入口；其余新域名等待用户确认）。服务器正式产品目录拟为 `/projects/yuelili.com/ytv`。当前 checkout `17bc209ab4b9de2982e79004328a7497782a7722`；应用制品来源以下方独立 SHA 为准，后续文档或迁移工具提交不会自动改变既有应用制品。

后端源 `7607889f0626134abdf4c34579a9b5a417dc255d`：GOWORK=off 的 go test ./...、Linux/amd64/CGO=0 固定构建通过；命令 ytv。迁移文件 18 项 SHA 记录在 manifest.json，数据库集成覆盖以测试配置/日志为准，不能把未设置数据库的单测当生产验证。

前端：固定源 17bc209ab4b9de2982e79004328a7497782a7722；冻结安装、前端测试、类型、生产构建、已安装内容 verified 均通过。Web tar SHA-256 0b5292940e58a1ba5e01788170e45a65ea0d6d647fd598c09a77ed601673f578。UI 为 .6 固定 SHA 06407c95454b215c61b5cf3ef432476de6677d3a78542b07fe706a7c52106c0a，sourceOverrides=[]。

迁移入口：api/internal/postgres/migrations。只在新建/明确备份的目标数据库预检并应用，不运行 devseed，不复制本地账号/业务数据，不修改旧 checksum。现有生产 Provider 按精确能力绑定，禁止直接套用 standalone compose 重建中央服务。

所需注册：Identity Web client、真实 source allowlist；Chromium 只为受支持来源解析 URL。

管理员/所有权：现有启动入口仍要求 Bootstrap subject，必须补齐共享首位管理员认领。

网络与密钥：Web/API 仅发布宿主机回环端口。所有 browser issuer/callback 使用正式 HTTPS Origin；Seal Secret、服务客户端 secret  由私有部署配置独立生成并备份，不使用开发默认值。证书需覆盖所选域名，现有证书不能作为新域名证明。

旧版文件完整备份 `/projects/yuelili.com/backups/ui-components-20261009-3/ytv-legacy`，逐文件 SHA 核对。10 个旧来源均保留为禁用的部署候选；1 个 HTTP endpoint、1 个旧 detail override 不能静默放宽新版合同。播放/图片主机允许列表和目录协议仍需逐源元数据验证，不获取、代理或持久化第三方视频流。候选来源目录不表示已验证可播放。

完成门禁：固定源码/依赖与制品可追溯 → 域名/TLS → Provider 注册及绑定检查 → 数据库迁移预检/初始化 → 候选健康/API → 精确切换 → 正式域名 CLI Playwright 手机/桌面、明暗主题、实际交互与登录门禁。初始化认领由用户显式完成；线上管理写入不以匿名测试代替。

现有部署资料：没有可直接用于生产的独立 Compose 合同，需要补齐。。

本机证据 `C:/Users/yl/.codex/task-artifacts/sites-ui-upgrade/20261009-3/ytv` 与 `ytv-api`；原始依赖对照在相邻 20261009-2/first-release-original-dependencies.json。源码快照/依赖为可再生过程产物，收尾删除；固定归档、二进制、Web 包、哈希、日志和迁移准备保留用于后续发布。
