# Repository instructions
AGENTS version: 1.0.2

通用约束遵循用户级根 AGENTS.md；本文件只保留本仓库职责和入口。

- 生产发布先读取[配套发布合同](../workspace/docs/deployment-policy.md)，核对线上实际实现；清单与验收覆盖前后端，以实际运行实现和生产回执选择发布入口，不把历史旧站状态当成当前事实。

- When adjacent `../workspace` exists and its `repos.lock.yaml` includes this repository, read and follow
  `../workspace/docs/multi-project-development.md` before cross-repository work or any local process lifecycle action. Product
  sessions treat Workspace contracts and `.doctor/` as read-only, use the Workspace CLI for generated state, use a separate Git
  worktree for concurrent writes to this repository, and never own shared Provider processes.
- The historical cross-repository rebuild is recorded at `../workspace/flightdeck/work/2026-08-30-ytv-rebuild/index.md`; keep product
  implementation in this repository and Workspace orchestration contracts in the adjacent Workspace repository.
- Keep the legacy checkout `E:\projects\yozya\tv` read-only. It is product and visual evidence, not the implementation target.
- Identity is the only user identity authority. Use Identity Nuxt and Foundation Authorization; do not copy product cookies,
  refresh rotation, initial administrator claim or role persistence.
- YTV public and administration pages use SPA rendering (`ssr: false`), as approved on 2026-11-14. Keep Nitro authentication,
  BFF and poster handlers; this product choice is an exception to the shared administration SSR default. Preserve existing UI.
- Source definitions, endpoints and scraping behavior are code/deployment allowlist contracts. Administration may install or remove
  an approved definition, rename its presentation, change enabled/home-visible/priority policy and probe it, but must not accept
  arbitrary endpoint or playback URLs.
- Do not upload, transcode, proxy or persist third-party video streams. Asset is only for YTV-owned brand and operational media,
  using the formal Asset SDK and consumer media contract.
- Administration uses the Foundation admin shell, page, collection, settings and feedback modules. Do not copy equivalent local
  Sidebar, Topbar, PageHeader, Collection, Toast or Overlay implementations.
- Local backend listeners bind to loopback. Only a Workspace Environment Contract may expose the Nuxt frontend beyond loopback.
- Validate public and administration surfaces with repository CLI Playwright through the real Workspace composition, covering
  desktop/mobile, light/dark, source degradation, playback recovery, login and administration refresh.
- Interface icons use Tabler through Iconify. Do not introduce a second icon family.

## 共用入口

- 共用规则、技能和知识的归属见 [Workspace 内容归属](../workspace/docs/knowledge-placement.md)；产品实现与特有操作留本仓库，不复制根流程。
- 代码部署使用 [Workspace 部署技能](../workspace/.agents/skills/yueli-deploy/SKILL.md)和[发布合同](../workspace/docs/deployment-policy.md)，服务参数与更新命令仍由本仓库拥有。
- 使用共享 UI 前，先读实际前端目录中已安装的 `node_modules/@yueli/ui/AGENTS.md`，按该包执行 inspect、search/docs；不拿邻仓新源码冒充已安装接口。

## Flightdeck

长期工作使用已安装的 Flightdeck 技能，从 `flightdeck/deck.md` 恢复。`/flightdeck/` 保持 Git 忽略，维护工作台不授予提交、推送或发布权限。
