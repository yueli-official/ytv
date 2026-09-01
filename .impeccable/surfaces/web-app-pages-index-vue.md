---
version: 1
slug: "web-app-pages-index-vue"
primary_target: "web/app/pages/index.vue"
related_targets: ["web/app/pages/search.vue", "web/app/components/SearchCommand.vue"]
---

# Home and discovery surface

- Scope and mode: public `/` and `/search`; dense Operate mode.
- Audience and job: viewers search multiple enabled internet catalogs, revisit recent queries, browse admin-selected source updates,
  and use attributed Douban discovery without confusing those tasks.
- Primary order: compact search panel → Continue Watching when present → Douban hot → source updates.
- Width contract: `/search` form, source controls, result summary, result grid and paging all use the same 90rem content axis.
- Navigation contract: the YTV mark owns home navigation; do not add a second “浏览” link.
- Search panel: search field and local search history only. Viewer source choice belongs exclusively to the navigation settings
  drawer; the panel does not repeat those controls or own source paging or Douban filters.
- Search history: local-first, most recent unique queries first, maximum eight, clearable, and never synchronized without a
  separate product decision.
- Viewer source choice: list only globally enabled Source Definitions; default to all enabled, persist locally, submit selected keys
  with the search route and keep partial-failure reporting explicit.
- Source updates: only admin-selected `homeVisible` sources participate. “换一批资源” belongs in this module and advances only
  source discovery paging.
- Continue Watching: at most six de-duplicated unfinished titles in a 15–16rem horizontal rail, each with resolved source poster,
  episode, remaining time and bottom-edge progress. It must stay under 180px high including its heading.
- Douban hot: a separate attributed module below source updates. Its movie/series/tag controls and “换一批热门” affect only
  Douban data; items use a compact 2:3 poster rail with score/new badges and remain search shortcuts.
- Poster delivery: Douban cover paths are provider-validated before projection and use the fixed-ID same-origin relay. The relay
  never accepts arbitrary URLs; failed individual covers retain the neutral poster well.
- Visual world: white YTV plus one sky dot, near-black/cold-blue surfaces, real source posters, no public purple or fabricated media.
- Responsive behavior: 2→3→4→5→6 source posters, local horizontal overflow only for compact history/hot strips, no page overflow.

## Direction contract

THESIS: Discovery tasks are explicit siblings; search, source paging and external hot discovery never masquerade as one control panel.

OWN-WORLD: Cold-blue dark portal, factual source identity, dense portrait posters and one sky interaction signal.

STORY: Search chosen sources or revisit compact poster history, scan Douban's attributed poster rail, then browse operator-selected source updates.

FIRST VIEWPORT: A slim navigation bar leads to a compact multi-source search panel; real source updates begin immediately after any resume state.

FORM: Source and history preferences remain viewer-local; deployment endpoints and adapters remain code/deployment allowlist facts.

FINISH: Playwright proves history, source toggles, independent batch actions, partial failure, 390px overflow and keyboard operation.
