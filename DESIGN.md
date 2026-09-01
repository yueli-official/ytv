---
version: alpha
name: YTV
description: A cold-blue dark film portal for source-backed discovery, complete episode navigation, and playback.
colors:
  primary-soft: "oklch(82.8% 0.111 230.318)"
  primary: "#38bdf8"
  primary-action: "oklch(50% 0.134 242.749)"
  primary-action-hover: "oklch(58.8% 0.158 241.966)"
  navigation-ink: "#06090d"
  page-ink: "#070b10"
  title-canvas: "#080b10"
  poster-well: "#0c1219"
  search-summary: "#0e151d"
  module-panel: "#0f1720"
  card-slate: "#111a24"
  detail-card: "#121821"
  episode-panel: "#10151d"
  player-canvas: "#05080d"
  player-panel: "#10151c"
  black: "#000000"
  white: "#ffffff"
  muted-white: "rgb(255 255 255 / 0.60)"
  faint-border: "rgb(255 255 255 / 0.10)"
  score-amber: "oklch(82.8% 0.189 84.429)"
  status-warning: "oklch(85.2% 0.199 91.936)"
  status-error: "oklch(70.4% 0.191 22.216)"
typography:
  brand-mark:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 900
    lineHeight: 1.3333
    letterSpacing: "-0.04em"
  display:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1.875rem"
    fontWeight: 900
    lineHeight: 1.2
    letterSpacing: "-0.035em"
  display-wide:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "2.25rem"
    fontWeight: 900
    lineHeight: 1.1111
    letterSpacing: "-0.035em"
  page-title:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 700
    lineHeight: 1.3333
    letterSpacing: "-0.03em"
  page-title-wide:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1.875rem"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "-0.03em"
  headline:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 700
    lineHeight: 1.4
    letterSpacing: "-0.025em"
  headline-wide:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 700
    lineHeight: 1.3333
    letterSpacing: "-0.025em"
  card-title:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 600
    lineHeight: 1.4286
  card-title-wide:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 600
    lineHeight: 1.5
  body:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.4286
  control:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.4286
  meta:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1.3333
  episode-row:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "0.8125rem"
    fontWeight: 400
    lineHeight: 1.5385
  eyebrow:
    fontFamily: "DM Sans, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 600
    lineHeight: 1.3333
    letterSpacing: "0.22em"
rounded:
  sm: "0.5rem"
  md: "0.75rem"
  full: "9999px"
spacing:
  micro: "0.25rem"
  xs: "0.5rem"
  sm: "0.75rem"
  md: "1rem"
  compact: "1.25rem"
  lg: "1.5rem"
  page: "1.75rem"
  xl: "2rem"
  discovery-max: "90rem"
  shell-max: "96rem"
  detail-max: "72rem"
  playlist-min: "20rem"
  playlist-max: "22rem"
components:
  site-navigation:
    backgroundColor: "{colors.navigation-ink}"
    textColor: "{colors.white}"
    height: "4rem"
    padding: "{spacing.md}"
  button-primary:
    backgroundColor: "{colors.primary-action}"
    textColor: "{colors.white}"
    typography: "{typography.control}"
    rounded: "{rounded.sm}"
    padding: "0.5rem 0.75rem"
  button-primary-hover:
    backgroundColor: "{colors.primary-action-hover}"
  search-compact:
    backgroundColor: "{colors.black}"
    textColor: "{colors.white}"
    typography: "{typography.body}"
    rounded: "{rounded.sm}"
    padding: "0.5rem 0.75rem"
    height: "2.75rem"
  search-panel:
    backgroundColor: "{colors.module-panel}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"
  search-history-chip:
    backgroundColor: "{colors.card-slate}"
    textColor: "{colors.white}"
    typography: "{typography.meta}"
    rounded: "{rounded.sm}"
    padding: "0.5rem 0.75rem"
  source-toggle:
    backgroundColor: "{colors.poster-well}"
    textColor: "{colors.white}"
    typography: "{typography.meta}"
    rounded: "{rounded.sm}"
    padding: "0.375rem 0.625rem"
  source-poster-card:
    backgroundColor: "{colors.card-slate}"
    textColor: "{colors.white}"
    rounded: "{rounded.sm}"
    padding: "0.625rem 0.75rem"
  douban-hot-module:
    backgroundColor: "{colors.module-panel}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"
  continue-watching-card:
    backgroundColor: "{colors.module-panel}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    width: "15rem"
    height: "7rem"
  progress-poster-card:
    backgroundColor: "{colors.poster-well}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    aspectRatio: "2 / 3"
  search-result-card:
    backgroundColor: "{colors.card-slate}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    padding: "{spacing.sm}"
    height: "10rem"
  title-facts-card:
    backgroundColor: "{colors.detail-card}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    padding: "{spacing.compact}"
  title-episode-browser:
    backgroundColor: "{colors.episode-panel}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"
  player-primary-control:
    backgroundColor: "rgb(0 0 0 / 0.75)"
    textColor: "{colors.white}"
    rounded: "{rounded.full}"
    size: "4rem"
  player-playlist:
    backgroundColor: "{colors.player-panel}"
    textColor: "{colors.white}"
    rounded: "{rounded.md}"
    width: "clamp(20rem, 24vw, 22rem)"
  player-episode-row:
    backgroundColor: "transparent"
    textColor: "{colors.muted-white}"
    typography: "{typography.episode-row}"
    height: "2.75rem"
    padding: "0.5rem 0.75rem"
---

# Design System: YTV

## Overview

**Creative North Star: "The Cold-Blue Portal"**

The Cold-Blue Portal is a dense, dark-locked film interface built for finding, selecting, and watching rather than promotion. Near-black canvas, cold slate panels, portrait posters, compact facts, and one sky/cyan interaction signal make the public product feel like a working source portal.

Discovery tasks are explicit siblings. Search owns viewer-local history, while source choice lives in the navigation settings drawer; Douban owns its own filters and paging, and the source-update grid owns its own paging. Title facts, ranged episode browsing, and complete in-player episode navigation each receive a distinct task region. Administration stays in the shared Yueli console family and inherits its semantic light/dark shell instead of the public dark lock.

**Key Characteristics:**

- Dark-locked public browse, search, title, history, and playback surfaces.
- White YTV wordmark with one small sky/cyan dot.
- Search history capped at eight; viewer source exclusions stay local for visitors and synchronize for signed-in users.
- Independent source-update and Douban poster modules with separate batch actions.
- Compact real-poster Continue Watching and a de-duplicated 2→6-column playback-history grid.
- Real source posters in a two-to-six-column 2:3 grid and compact horizontal search results.
- Stable 50-episode ranges on title and watch routes.
- App-owned central play/replay plus a 20–22rem single-line desktop playlist.
- Shared semantic admin console with dashboard, sources, search insights, and public profile settings only.

## Colors

The public palette is a cold-blue night ladder: true and near blacks for canvas, tightly stepped blue-slates for structure, and sky/cyan for interaction.

### Primary

- **Signal Sky** (`#38bdf8`): The 6px brand dot, focus rings, selected edges, app-owned playback control, player progress, and Plyr's main control color.
- **Action Sky** (`oklch(50% 0.134 242.749)`; hover `oklch(58.8% 0.158 241.966)`): Solid primary buttons, selected filters, and the current episode.
- **Soft Sky** (`oklch(82.8% 0.111 230.318)`): Eyebrows, compact labels, icon treatments, and active episode text.

### Neutral

- **Navigation Ink** (`#06090d`): Sticky public navigation.
- **Page Ink** (`#070b10`): Browse, search, and history canvas.
- **Title Canvas** (`#080b10`): Title-detail route behind its task cards.
- **Poster Well** (`#0c1219`): Mobile drawer and absent-poster wells.
- **Search Summary** (`#0e151d`): Search counts and empty-state regions.
- **Module Panel** (`#0f1720`): Separate search and external-discovery modules.
- **Card Slate** (`#111a24`): Source cards, history chips, and compact search results.
- **Detail Card** (`#121821`): Poster support and factual title summary.
- **Episode Panel** (`#10151d`): Title-page episode browser.
- **Player Canvas** (`#05080d`): Watch-page canvas around true-black video.
- **Player Panel** (`#10151c`): Previous/next bar and playlist.
- **White Hierarchy** (`#ffffff`): Brand, titles, and controls use full white; supporting facts step down through controlled white opacity.

### Status

- **Score Amber** (`oklch(82.8% 0.189 84.429)`): Source-provided scores only.
- **Warning Yellow** (`oklch(85.2% 0.199 91.936)`): Partial source availability and recoverable warning state.
- **Error Red** (`oklch(70.4% 0.191 22.216)`): Failed sources, episode remarks, and explicit recovery state.

### Named Rules

**The One Sky Signal Rule.** Sky/cyan is the sole public interaction color; it marks action and state without becoming a large surface.

**The White Mark Rule.** The YTV name stays white and gains identity from one 6px sky dot; never apply a gradient or multicolor fill to the public wordmark.

**The Split Appearance Rule.** Public routes stay dark-locked, while the shared administration shell retains its semantic light/dark appearance control.

## Typography

**Display Font:** DM Sans (with system-ui and sans-serif fallbacks)

**Body Font:** DM Sans (with system-ui and sans-serif fallbacks)

**Character:** One assured workhorse sans keeps Chinese media labels compact and predictable. Hierarchy comes from weight, size, tabular numbers, and restrained negative tracking rather than a separate entertainment display face.

### Hierarchy

- **Brand Mark** (900, 1.5rem, 1.3333, -0.04em): White YTV text beside the sky dot.
- **Display** (900, 1.875rem / 2.25rem from 640px, 1.2 / 1.1111, -0.035em): Title-detail names.
- **Page Title** (700, 1.5rem / 1.875rem from 640px, 1.3333 / 1.2, -0.03em): Dedicated route titles.
- **Headline** (700, 1.25rem / 1.5rem from 640px, 1.4 / 1.3333, -0.025em): Search, source-update, and external-discovery headings.
- **Card Title** (600, 0.875rem / 1rem from 640px, 1.4286 / 1.5): Source-poster and compact result titles.
- **Body** (400, 0.875rem, 1.4286): Search values, summaries, and supporting copy.
- **Control** (500, 0.875rem, 1.4286): Navigation, search, paging, playback, and episode actions.
- **Meta** (400, 0.75rem, 1.3333): Attribution, range, year, category, source, and progress facts.
- **Episode Row** (400, 0.8125rem, 1.5385): One-line watch-playlist number and real title.
- **Eyebrow** (600, 0.75rem, 1.3333, 0.22em, uppercase): Sparse YTV Search label.

### Named Rules

**The Workhorse Only Rule.** Use the same sans family throughout YTV; let compact facts, poster density, charts, and row structure provide character.

## Layout

Public browse, search, and playback use a fluid frame capped at 90rem; global navigation may extend to 96rem, while title details cap at 72rem. Gutters step from 1rem to 1.5rem at 640px and 2rem at 1024px. A 4px base rhythm supports 44px controls, compact cards, and 1.75rem between major browse regions.

The source and playback-history poster grids use two columns by default, then three at 640px, four at 768px, five at 1024px, and six at 1280px. Search results use one compact horizontal card column by default, two at 768px, and four at 1280px. The dedicated search route keeps its form, source controls, result summary, results and paging on one 90rem axis.

The homepage sequence is search, optional compact resume rail, Douban hot, then source updates. Continue Watching is a 15–16rem horizontal card carousel and never a large 16:9 feature row. The search panel contains only the field and local history; viewer source choice lives in the navigation settings drawer. Continue Watching and Douban hide native scrollbar chrome and use side-centered carousel controls. Source updates and Douban each own an independent header, filter or paging state, batch action, failure state, and content region.

Title details stack on mobile and become a 15rem poster column plus flexible facts at 768px. Its separate episode browser shows 50 episodes per selected range and keeps the compact 4→6→8→10→12 grid.

Playback stacks video, previous/next, sticky range header, and one-column episode rows on narrow screens. From 1024px it becomes flexible video plus a `clamp(20rem, 24vw, 22rem)` sidebar. The sidebar aligns with video and the previous/next bar, fixes its header, and scrolls only its episode body.

Administration uses the shared fixed sidebar, 64px context topbar, full working canvas, and responsive gutters. Metrics collapse 4→2→1, trend columns stack below 1280px, and source collection rows reflow without horizontal page scrolling.

### Named Rules

**The Independent Modules Rule.** Search, source updates, and Douban never share paging, filtering, batch actions, or recovery state.

**The Density Ladder Rule.** Preserve the 2→3→4→5→6 source-poster progression and 1→2→4 search-result progression.

**The Complete Player Rule.** The watch route contains every episode in stable 50-episode ranges; desktop reserves 20–22rem for the list and mobile uses document scroll.

## Elevation & Depth

Public UI is border-led and mostly flat. Tonal steps and 10% white hairlines separate ordinary surfaces. Only primary media receives lift: the title poster uses the shipped extra-large shadow and the video stage uses the shipped 2×-large shadow. Administration inherits Foundation's semantic card and overlay depth.

### Shadow Vocabulary

- **Title Poster Lift** (`0 20px 25px -5px rgb(0 0 0 / 0.10), 0 8px 10px -6px rgb(0 0 0 / 0.10)`): Standalone title poster only.
- **Video Stage Lift** (`0 25px 50px -12px rgb(0 0 0 / 0.25)`): Dominant playback surface only.
- **Admin Semantic Card:** Use the shared Foundation card shadow token rather than a YTV-local imitation.

### Named Rules

**The Media-Only Lift Rule.** Public shadows lift the title poster and video stage; ordinary public components use tone and border, while admin uses shared semantic depth.

## Shapes

Compact buttons, source toggles, history chips, and icon wells use the 0.5rem curve. Public panels, poster cards, result cards, title facts, episode regions, playlist, and video stage use 0.75rem. Scores, status dots, and the sky brand dot are fully round.

The recurring media silhouette is a bordered 2:3 poster. Search results preserve a portrait poster column inside a horizontal fact card, while the player uses a bordered 16:9 rectangle. Playlist entries are rows with a 2px leading state edge, not independent rounded cards.

### Named Rules

**The Source Poster Rule.** Homepage poster cards always represent real administrator-selected source entries with a 2:3 image well and compact facts.

**The Playlist Row Rule.** Watch episodes form one continuous divided list; the current row gains a sky leading edge and tint without changing geometry.

## Components

### Navigation

- **Public:** Sticky 4rem navigation ink bar, white YTV plus 6px sky dot, Playback History, search/account actions, and no appearance switch. The mark is the only home action; do not duplicate it with “浏览”.
- **Admin:** Shared `AdminConsoleLayout` owns the commercial sidebar, 64px context topbar, mobile drawer, account position, and appearance control.

### Buttons and Focus

- **Primary:** Action sky with white text and no gradient; selected filters and current episode use the same family.
- **Neutral:** Dark or transparent fill with a thin semantic border.
- **State:** Visible 2px sky focus, short color transitions, and at least 44px for touch-critical controls.

### Search Panel

- **Contents:** Search field, local search history, and viewer source toggles only.
- **History:** Most-recent unique queries first, maximum eight, clearable, and passed back through the current selected sources.
- **ViewerSourcePreference:** A public-header settings action opens a right Slideover listing globally enabled definitions. Keep at least one source selected and persist only disabled keys so newly enabled sources enter by default; visitors use LocalStorage and signed-in users synchronize the same preference to their account.
- **Submit:** Include selected source keys in the search route; empty query, IME composition, or zero selected sources prevents submission.

### Source Updates

- **Header:** Factual module title, administrator-selected visible source badges, and “换一批资源”.
- **Content:** Real source poster grid with its own loading, partial-failure, error, empty, and paging state.
- **Separation:** Its batch action changes only source discovery paging.

### Douban Hot Module

- **Structure:** Independent bordered panel below source updates with its own title, attribution, “换一批热门”, film/series choice, tags, error, and compact horizontal 2:3 poster rail.
- **Action:** Each poster, title and provider score form one search shortcut using the viewer's selected source keys.
- **Delivery:** The API projects only validated Douban poster filenames. A same-origin fixed provider route supplies the required
  provider headers, rejects redirects, non-images and bodies over 2 MiB, and never accepts a URL parameter.
- **Overflow:** Only the poster rail scrolls horizontally; it remains one row rather than a second multi-row source wall.

### Continue Watching and Playback History

- **Continue:** Maximum six unfinished titles, de-duplicated by Source Title Locator, in 15–16rem × 7rem horizontal cards with a
  narrow real-poster column, latest episode, remaining time and bottom-edge progress.
- **History:** Latest episode per title in the same 2→3→4→5→6 portrait grid as source updates. Percent and episode overlay the
  poster; title, source and recent time remain below; the whole card resumes playback.
- **Enrichment:** Resolve real poster/source metadata from the existing title API after mount, four at a time and at most 24.
  Disabled, missing or failed sources fall back to the stable neutral poster well without blocking progress.

### Source Poster Card

- **Frame:** Flat bordered 2:3 real source poster over a compact facts body.
- **Facts:** Score and episode remark may overlay the image; title, year, category, and source identity remain factual.
- **State:** Hover scales only imagery to 1.035 over 300ms; focus uses a 2px sky outline with a 4px offset.

### Search Result Card

- **Frame:** 0.75rem horizontal card, fixed 7rem poster well, and 10rem minimum height.
- **Facts:** Title, year, category, episode state, primary source, and up to three alternate sources remain together.
- **Responsive:** One column on mobile, two at 768px, and four at 1280px.

### Title and Episode Browser

- **Facts:** Mobile stacks the 2:3 poster above a factual card; desktop aligns a 15rem poster beside it.
- **Episodes:** Independent panel below facts, explicit 50-episode ranges, range replacement rather than append, and 44px buttons in the 4→12 grid.

### Player Primary Control

- **Ownership:** The app owns the central play/replay overlay; Plyr remains the accessible transport adapter.
- **State:** Ready and paused show Play, ended shows Replay, playing hides the control, and loading/error replace it with explicit status.
- **Size:** 64px on mobile and 72px from 640px; white icon, translucent black circle, subtle border, and sky hover/focus.

### Player Playlist

- **Desktop:** `clamp(20rem, 24vw, 22rem)` sidebar beside the video; fixed range header and locally scrolling body.
- **Mobile:** Sticky range header followed by a one-column episode list in document scroll.
- **Range:** 50 episodes, automatically selecting the range containing the current route.
- **Row:** Minimum 44px, 13px text, fixed number column, truncated real title, divided edges, sky current marker, and no route away to find the remaining episodes.

### Administration

- **Routes:** Dashboard, Sources, and Site Settings only; search metrics, trend and top queries live on Dashboard. The former Search Insights route redirects to Dashboard.
- **Source Lifecycle:** “添加数据源” installs only an uninstalled deployment-approved definition. Delete removes it from runtime
  search/home participation but retains the approved definition for later reinstall. Endpoint, adapter and host fields never enter the UI.
- **Source Policy:** Display name, `enabled`, `homeVisible`, and `priority` auto-save; switches persist immediately and text/number
  fields settle for 350ms. Pending, “已自动保存”, and failure feedback stay beside the edited source; there is no per-row Save button.
- **Source Probing:** Per-row Probe remains available and “检测全部” probes every installed source with bounded concurrency, then
  shows row results and an available/total summary.
- **Source Collection:** Foundation `CollectionPanel` owns search, enabled/home-visible filters, paging, selection and bulk actions.
  Per-row Probe/Delete are low-frequency actions inside an `AdminRowActions` overflow menu. Desktop header and rows share the exact
  `minmax(18rem,1fr) 7rem 7rem 7rem 3rem` template; mobile removes the tabular header and uses explicit field labels without overflow.
- **Source Tools:** A compact header menu owns Probe All, enable available sources and disable unavailable sources. Availability tools probe first, then apply only the matching policy and report the changed count in place.
- **Initial Policy:** `zy360`, `mdzy`, and `zuid` start enabled; `qgikun` starts disabled; only `zy360` starts visible on the homepage.
- **Approved Candidates:** `yinghua`, `qgsuoni`, and `qgyouzhiziyuanku` start uninstalled and disabled, so the Add dialog has
  meaningful choices without making an unprobed source participate in public search.
- **Site Settings:** One Public Profile section with editable name and description; the page-level save action owns pending, success, and error feedback.
- **Shared Modules:** Use `ManagePage`, semantic metric cards and charts, `CollectionPanel`, `SettingSection`, and `ActionFeedbackButton`; do not reproduce them locally.

### Motion

- **Poster Hover:** Scale only source imagery to 1.035 over 300ms and remove the transition under reduced motion.
- **Control State:** Short color transitions only; the central player control may scale to 1.05 on hover.

## Do's and Don'ts

### Do:

- Do keep public routes dark-locked and administration on the shared semantic light/dark console.
- Do keep search history unique, local, clearable, and capped at eight.
- Do persist disabled viewer source keys, preserve at least one selected source, and respect globally enabled definitions.
- Do give source updates and Douban independent headers, actions, filters, paging, and recovery.
- Do keep search form, controls, result summary, grid and paging on the same 90rem axis.
- Do show real resolved posters in Continue Watching and playback history and de-duplicate history by title.
- Do use stable 50-episode ranges on title and watch routes.
- Do keep the app-owned play/replay control unmistakable in ready, paused, and ended states.
- Do use one-line 44px playlist rows and keep every episode reachable on the watch route.
- Do restrict source creation to deployment-approved definitions while allowing install/remove, display-name and policy management.
- Do auto-save source changes with explicit row feedback and keep one-click probing available for all installed sources.
- Do keep the initial policy explicit: zy360/mdzy/zuid enabled, qgikun disabled, and only zy360 home-visible.

### Don't:

- Don't place Douban filters, hot entries, or batch action inside the search panel.
- Don't add a redundant “浏览” link beside the YTV home mark.
- Don't use large 16:9 cards, full-width action buttons or a shared demo image for playback records.
- Don't let one batch action change both source updates and external discovery.
- Don't render watch episodes as a multi-column button grid or send viewers away to find the full list.
- Don't merge title facts, episode browsing, and playback into one oversized card.
- Don't add a placeholder access-control page or unsupported user/role workflows to the admin console.
- Don't add deployment endpoints, adapters, or secrets to source settings.
- Don't require a redundant per-row Save button for independently editable source fields.
