# Repository instructions

- 生产发布先读取[配套发布合同](../docs/docs/deployment-policy.md)，核对线上实际实现；清单与验收覆盖前后端，当前旧站和未上线新实现使用各自的发布入口。

- When adjacent `../workspace` exists and its `repos.lock.yaml` includes this repository, read and follow
  `../workspace/docs/multi-project-development.md` before cross-repository work or any local process lifecycle action. Product
  sessions treat Workspace contracts and `.doctor/` as read-only, use the Workspace CLI for generated state, use a separate Git
  worktree for concurrent writes to this repository, and never own shared Provider processes.
- The active cross-repository rebuild is owned by `../workspace/flightdeck/work/2026-08-30-ytv-rebuild/index.md`; keep product
  implementation in this repository and Workspace orchestration contracts in the adjacent Workspace repository.
- Keep the legacy checkout `E:\projects\yozya\tv` read-only. It is product and visual evidence, not the implementation target.
- Identity is the only user identity authority. Use Identity Nuxt and Foundation Authorization; do not copy product cookies,
  refresh rotation, initial administrator claim or role persistence.
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

## Flightdeck 工作台维护

开始或恢复仓库工作时，先读 `flightdeck/deck.md` 与 `flightdeck/knowledge/index.md`；
按相关目录摘要逐层进入所需知识，不预加载整个知识树。每个自有文档文件夹都保留 `index.md`，
说明范围、直属子文件用途与何时阅读，并链接子目录的索引；外部只读参考库保持原布局。
摘要不记录进度、当前结论、验证结果、日期或数量。Work 的 Goal、Status、Current、Next
及执行入口仍由 Work 页负责；Plan 管阶段顺序，Slice 管局部细节，Context 管稳定背景。
创建 Work 时在 References 引用适用 Knowledge 并说明用途；立即需要的资料放进 Next，
最多三个直接链接，不添加无关依赖。deck 保持 Knowledge 与 Work 索引入口。
新增、移动、重命名或删除文档时同步维护所在索引和受影响入链；移动前搜索全部工作台 Markdown，
包括 Work，而不只检查 Knowledge。正文、进度或验证结论更新不向父索引传播；只有用途范围改变才改摘要。
Focus 是 Open Work 的恢复优先级，可有多个；未聚焦的 Open Work 放 Other open work。
日常保存保留 Focus；明确搁置才移出。交付待确认仍为 Open，用户确认整个 Work 完成后才 Finished；
用户取消时为 Stopped。终态 Work 留在原位，历史摘要放 `flightdeck/work/index.md`，未知确认日期不推测。
新建工作台或完成升级验证后，在 deck 标题下写入实际加载的完整 Flightdeck 版本；普通保存与新建 Work 不刷新版本。
`flightdeck/` 仅作本地记录：`.gitignore` 包含 `/flightdeck/`；已有跟踪记录保留本地文件并从 Git 索引移除。
只维护工作台不代表执行 Work 下一步、提交、推送或发布。
