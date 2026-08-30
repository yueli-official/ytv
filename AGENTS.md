# Repository instructions

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
