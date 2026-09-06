# 交付证据

2026-09-06，基线 eaab973；修改范围为当前 api/、web/ 和基础合同 CI。旧根 Go/Vue frontend、来源定义、媒体 allowlist、播放器实现与页面设计保持不变。

## 实现

- Foundation Go v0.4.1 与当前发布的 HTTP/Nuxt/UI 精确包引用；Identity Nuxt 0.2.0。
- 15 个声明式产品错误，26 个 GoFrame 路由推导的 OpenAPI/operations；Go、TS、i18n 生成物。
- 来源目录与批量 Probe 使用 items；安装返回 201 和 Location。搜索聚合保持 resource，Trending/近期进度保持有界 collection，不伪造未知 total。
- 校验错误、来源不可用、播放地址拒绝、版本冲突与安装状态独立映射，包装保留 errors.Is 原因；未知存储失败保持 internal。
- 前后台安全本地文案、表单字段反馈、技术详情 trace、BFF Problem；共享 Cookie 与恢复逻辑未复制。
- 新增 CI OpenAPI freshness、目录与引用验证、生成消费者 check、基线 compatibility diff、Go 与 Web 门禁。Project v1 尚未在固定 Go v0.4.1 发布，完整 Project 推导检查当前使用 Foundation 本地 CLI，不复制产品 checker。

## 通过的检查

- Go 全包测试；HTTP 映射回归、安装控制器/路由改动的定向测试；go vet ./...。
- Foundation local Project -check：26 operations current；本次基线无旧合同文件，主动迁移说明见 api/contracts/http-result/README.md。未来 CI 基线有文件时执行 compatibility diff。
- Web Vitest 5 个文件 9 项测试；Nuxt typecheck；生产 build 完成（5.91 MB / 1.27 MB gzip）。
- CLI Playwright 16 项：桌面/手机公共 collection、400/404 Problem、真实 OIDC 登录、409 版本冲突与 trace（4）；安装 201/Location 及已安装冲突（1）；桌面/手机/深色首页、实际 HLS 播放、失败恢复（9）；桌面/手机共享后台壳和零 hydration/component 警告（2）。
- 老后台用例曾硬编码两张趋势图；现有页面实际已有访问、访客、搜索三张。修正断言后定向复验通过，无未解决测试失败。人工查看了移动首页截图。

## 运行与限制

isolated Session 20260906T144333Z-48716：Web http://192.168.5.7:3013，API loopback 8097，Identity loopback 8781，Account 3700。独立测试数据库 doctor_ytv_upgrade_main / doctor_ytv_upgrade_identity；固定 fixture 来源，未改变既有真实源数据库。运行由 Workspace CLI 创建，不停止其他站点或共享 Provider。

本地日志/截图在忽略的 web/test/e2e/results 下：http-upgrade、http-upgrade-installation、http-upgrade-regression、http-upgrade-admin 与 http-upgrade-*.log。

YTV 类型检查暴露 Identity SDK 外部请求被 Nitro InternalApi 条件类型污染，已在 Identity 提交 4bef95f 修复。guest-session 回归 1/1 与 YTV 完整 typecheck 均通过。该修复尚未进入正式 SDK tarball：这里的类型/构建与浏览器结果使用 Workspace local overlay，不能当作无 overlay 的正式包验证。新 CI 未远程运行；正式 SDK、Foundation Project/BFF 候选发布和无 overlay 组合归后续发布阶段。

用户要求站点基础升级完成后再逐站人工验收，因此本 Work 的完成表示本地实现与自动化门禁完成，不表示人工验收、第三方来源权利确认或生产部署已完成。既有 YTV 重建 Work 的其他正式化目标不由本 Work 接管。

## 队列交接

Workspace P6 其余站点已有各自迁移记录。本会话补完 Shop 6bf5444、Licensing e23886c 与 YTV，并提交 Asset 私有附件修复 c713a7b、Identity SDK 类型修复 4bef95f。顶层 Workspace Flightdeck 与合同依规定保持只读；本地迁移状态以产品 Work 为准，不把仍列 Shop/Licensing 的旧顶层指针当作待重做任务。
