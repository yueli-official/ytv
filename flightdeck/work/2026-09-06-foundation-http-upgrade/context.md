# 上下文

只修改官方组织 YTV 的 api/ 与 web/，旧根 Go/Vue frontend 和旧 checkout 保持不变。基础升级范围为 HTTP 合同与依赖，不接管 Source 权利登记、QoE/Audit 等独立重建事项。用户要求人工验收全部升级后逐站进行；保留必要的代码与 CLI Playwright 验证，不等待人工确认。

普通 API 为正式发布前主动迁移，不保留旧 wire fallback；播放 URL allowlist、Cookie 与共享授权所有权不变。Workspace 合同与顶层 Flightdeck 只读。
