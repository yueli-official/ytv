---
version: 1
slug: "web-app-pages-manage"
primary_target: "web/app/layouts/manage.vue"
related_targets: ["web/app/pages/manage/index.vue", "web/app/pages/manage/sources.vue", "web/app/pages/manage/searches.vue", "web/app/pages/manage/settings.vue"]
---

# Administration surface

- Scope and mode: `/manage/**`; dense Operate mode.
- Audience and job: a YTV administrator must understand real traffic and search behavior, maintain the installed approved-source set and
  update the public profile without learning a new console grammar.
- Structural authority: Workspace `new-site-admin-baseline.md`, `admin-design-system.md`, `admin-console-ui.md`,
  `admin-color-system.md`, and `nuxt-ui-admin-conventions.md`; Blog and BVideo are rendered evidence, not code to copy.
- Shared topology: `YAdminConsoleLayout` owns the fixed commercial Sidebar, 64px context Topbar, mobile drawer, account position,
  Canvas, responsive gutters and SSR shell. Each route owns one `YManagePage` with its icon, title and sole primary action.
- Product adapter: YTV supplies sky/slate semantic theme values, navigation, permissions, traffic/search/source/profile data and
  mutations. It does not implement local Sidebar, PageHeader, Collection, Toast or Overlay equivalents.
- Reading path: Sidebar navigation → Topbar context → Canvas PageHeader/primary action → tabs or time range → one task surface →
  inline result or recovery.
- Page kinds: Dashboard for `/manage`, including search insights; `/manage/searches` redirects there. Collection owns sources; Settings owns site profile. Identity,
  Foundation Authorization and `adminSession` remain server-side access boundaries without a placeholder governance page.
- Density and rhythm: full-width working Canvas, `space-y-5`, four summary metrics at wide widths, asymmetric primary/secondary
  trend regions, compact lists, and no decorative or explanatory filler.
- Form contract: `SettingSection` provides the full-width body; Nuxt UI field component roots use `w-full`. Site profile saving is
  the single `YManagePage` action via `ActionFeedbackButton`; high-frequency source policy fields auto-save after switch activation
  or a 350ms settled input value and expose pending/success/error beside the source.
- Source lifecycle: Add lists only deployment-approved definitions not currently installed; inline name, participation, homepage
  visibility and priority are editable; multi-select supports batch policy, probing and removal, while Add supports selecting multiple
  candidates. “检测全部” probes installed sources concurrently and projects per-row availability plus one summary.
- Source tools: the page action area keeps Add as primary and groups Probe All, enable available and disable unavailable in one compact menu.
- Collection behavior: `CollectionPanel` owns enabled/home-visible filters, selection and bulk workflow; `AdminRowActions` overflow owns
  per-row Probe/Delete. Wide header and rows share `minmax(18rem,1fr) 7rem 7rem 7rem 3rem`; below `xl`, the header collapses and rows
  expose explicit labels without horizontal overflow.
- Appearance: admin supports the shared light/dark appearance control. Shell, Canvas, Card, Inset and Overlay remain semantically
  distinct in both modes; public dark-lock rules never leak into the admin subtree.
- Responsive behavior: 390/768/1024/1440px; metrics collapse 4→2→1, trend columns stack below `xl`, collection rows reflow without
  horizontal page scroll, and every interactive target remains at least 44px where the shared module requires it.

## Direction contract

THESIS: YTV administration is one member of the shared Yueli console family, differentiated by data and a sky signal—not by a new shell.

OWN-WORLD: Foundation commercial console topology, semantic sky/slate theme, clear shell/canvas/card depth, and factual operational density.

STORY: Understand what viewers are doing, inspect where discovery fails, install and tune approved sources, then adjust the public profile.

FIRST VIEWPORT: A fixed shared Sidebar and 64px context Topbar lead to a 30px Canvas title, four factual metrics and the highest-value operational trend.

FORM: Deep shared Modules first; YTV contributes only route, data, permission and mutation adapters.

FINISH: desktop/mobile and light/dark Playwright must prove hierarchy, field width, shared CSS defaults, SSR shell, overlays and zero overflow.
