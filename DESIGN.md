# YTV interface design

## Public cinema

The public site is a dark movie discovery experience. Real film artwork is the visual focus; navigation, source diagnostics and supporting metadata remain secondary. Public routes lock dark appearance; the shared administration retains its light/dark controls.

### Visual language

- Canvas `#0b0e14`, panels `#131821`, 8% white dividers and clear white text hierarchy.
- Pale sky `#7dd3fc` is the sole interaction accent, with dark text for solid actions. Amber marks actual source scores.
- A compact play-shaped sky mark accompanies the white YTV wordmark. Tabler is the only icon family.
- The public frame caps at 1280px with 32px desktop and 16px mobile outer gutters. Heading, poster rows, detail and player share that axis.
- Section titles are 17–19px; ordinary cards 12–14px. The selected featured film may use a larger title. Small text and disabled states must be checked on actual backgrounds.
- Ordinary surfaces use subtle borders and 6–12px corners. Only featured artwork, title posters and video receive depth. Respect reduced motion.

### Discovery

The sequence is search introduction, compact featured film, optional Continue Watching, attributed Douban hot rail, then actual source updates.

The featured film comes from the real trending response, with title, original poster and supplied score. Four manual buttons change the selection; no autoplay. The action searches the viewer's selected sources and does not promise availability. Wide screens include a three-entry hot shortlist. The shortlist is omitted on narrower screens because its entries remain in the main rail. Do not fabricate editorial descriptions, release dates or availability.

The search field and local search history belong to the homepage introduction. Search, hot and source update paging stay independent. Source preferences remain in the header drawer; low-priority provider badges do not compete with the posters. Partial results use compact local feedback with retry; complete errors remain explicit.

Both poster collections use the same 2:3 image ratio. The source grid is 3 columns on phones, then 4/5/6 at wider sizes; the hot rail uses corresponding card widths and horizontal navigation. Title and score sit below hot posters so the artwork remains readable. Do not use artificial visual assertions as design signoff: inspect screenshots at 375px and desktop, plus layout boundaries.

### Search and history

Search has a bounded form, clear query/result count, and a low-priority partial-source status. Results are horizontal portrait-and-facts cards in 1/2/3 columns. Preserve all alternate source links behind a keyboard-accessible disclosure and retain query, page and source scope on the journey to playback and back.

Continue Watching stays a compact horizontal portrait rail. The history route preserves local/account progress recovery and real metadata enrichment. Empty and unavailable states retain a route back to discovery. Source selection preserves at least one source, and newly enabled definitions are selected by default.

### Detail and playback

The title introduction groups the real poster, name, metadata, primary playback action and expandable synopsis in a restrained panel. Desktop uses a 200px poster; smaller screens use a compact poster beside facts, followed by synopsis. Keep long titles and primary actions legible on narrow screens.

The playback section is separate. Explicit film categories remain films even when a line contains multiple versions; explicit series categories remain series even with only one available episode. Unknown categories use the existing entry-count fallback. Films show versions and lines, without next-episode controls. Series preserve the complete playlist in stable 50-entry ranges.

The watch route keeps the 16:9 player primary, with a compact desktop sidebar and stacked mobile list. Video.js provides transport controls; temporary 2× hold and 10-second double-tap seek remain available, along with keyboard, progress, volume and rate preferences. Help and synopsis are expandable. Do not proxy or persist third-party video streams.

### Interaction and ownership

Use Foundation/Nuxt UI controls and their documented public slots. Product CSS is restricted to public layout/brand classes and must not override administration. Avoid private component DOM selectors, duplicated shared modules and important overrides. Buttons/links retain visible keyboard focus; carousels support both touch scrolling and desktop side controls.

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
