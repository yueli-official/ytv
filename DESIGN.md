# YTV interface design

## Public cinema — approved 2026-10-23

The five-page project adaptation is the current visual direction: charcoal blue neutrals, white text, pink primary actions, pale-blue selections and progress, and amber attributed scores. The public pages stay dark; Foundation owns account and administration appearance.

### Shared frame

Use a compact horizontal YTV wordmark, discovery/history links, header search, source settings and shared account control on desktop. Content uses a 92% desktop container capped at 1600px. Below 768px, use independent mobile views with discovery/search/history bottom navigation; detail and watch use back navigation without the bottom bar. Use Tabler icons, visible keyboard focus, real Nuxt UI controls and documented public slots. Public styles must not affect administration.

### Discovery

A full-bleed cinematic hero immediately below the header uses the real portrait intact, with a blurred copy behind its supplied title, attributed score and pink “查找片源” action. The three manual selectors have keyboard support and no autoplay. Large posters fall back to their original cover. No invented synopsis, availability or backdrop; a source-search action never promises playback.

Continue Watching uses three horizontal poster-and-progress cards on desktop. The attributed Douban rail retains upstream order, independent movie/TV filters and paging; featured items may also appear here so the rail remains the actual trending list. Source updates use one horizontal row of four visible cards with real source identity and their own paging. Empty history reserves no content row. Local search history opens beneath the focused header search input and closes on navigation.

### Search

The shared header form tracks query and source scope. Results show current-page counts, never a fabricated total. Wide portrait-and-facts cards group matching candidates without merging title identity. Up to three candidate sources are initially expanded; longer lists use a keyboard disclosure. Candidate links go to details and preserve the original query, source scope and page. Partial source failure stays local and retryable.

### Detail and playback

Details use an unboxed portrait-and-content composition over a full-width blurred backdrop, real metadata, primary start/resume action and expandable synopsis. A separate section identifies source and line, with stable 50-episode ranges. A lone line is text; multiple lines remain selectable. Films retain version/line semantics without series next-episode controls.

The watch page keeps a dominant 16:9 Video.js viewport beside a single-column episode list. Its title/actions and collapsed help span the full content width beneath both columns. Retain previous/next, wide mode, source recovery, gestures, keyboard controls and progress synchronization. A decoded paused frame stays visible; generated design images never substitute for media. Do not proxy or persist third-party video.

### History

Deduplicate by source title locator, preserving the latest episode. Show portrait-and-text rows grouped by today/yesterday/earlier, source identity, time and an explicit continue action. The filter searches only the loaded recent records. Show progress bars only with known duration; otherwise show elapsed time. Account failure preserves local history. There is no unsupported deletion or remote full-history search.

### Verification

Mobile implementation was authorized after the separate five-page concept. Keep actions compact and content-sized; large generated-reference typography is not a mandate to enlarge ordinary controls. Review actual desktop and mobile screenshots separately; concept approval does not constitute acceptance of the implementation.

Use the real Workspace composition and CLI Playwright for desktop/mobile rendering and interaction, including recovery and long-series behavior. Inspect actual screenshots personally for layout and legibility; automated assertions only verify their stated functional behavior. Respect reduced motion and keep source failure fallbacks legible.

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


### Shared state, independent mobile views

Routes own a single data controller through useHomeView/useSearchView/useTitleView/useHistoryView; components/public/desktop and mobile compose that same state. Keep API, source scope, history, errors and routing shared. Do not mount two request-owning page trees behind CSS visibility.

Mobile prioritizes one continue-watching card, three visible discovery posters, collapsed source candidates, a compact detail introduction and history rows. Auxiliary metadata and notices are disclosed on demand. The watch page retains one player DOM instance across viewport changes; its shared playlist appears beside it on desktop and in a modal bottom sheet on mobile. Selecting an episode closes the sheet.

usePublicLayout owns the 768px switch. SSR starts from the same desktop state as hydration; mobile switches after mount. Search loading timers start after mount so SSR output and hydration agree. New mobile composition styles live in mobile.css rather than accumulating desktop overrides.
