---
version: 1
slug: "web-app-pages-watch"
primary_target: "web/app/pages/watch/[source]/[id]/[episode].vue"
related_targets: ["web/app/components/VideoPlayer.client.vue", "web/app/pages/title/[source]/[id].vue"]
---

# Playback and episode navigation surface

- Scope and mode: `/watch/**` plus the title-page episode handoff; focused Operate mode.
- Audience and job: a viewer must immediately know how to start or resume playback, identify the current episode and navigate any
  episode without leaving the player.
- Primary path: content context → explicit play → video controls → previous/next → range selection → one-line episode list.
- Player state: an app-owned 72px desktop / 64px mobile central play button is visible for ready and paused, hidden while playing,
  and replaced by explicit loading, replay or recovery states. Plyr remains the accessible transport/control adapter.
- Desktop topology: flexible video plus a `clamp(20rem,24vw,22rem)` playlist from 1024px. The playlist aligns with video plus the
  previous/next bar, owns a fixed header and a locally scrolling body, and has no “view all” route away.
- Episode model: 50 episodes per stable range. The range containing the current route is selected automatically; changing range
  replaces only list content. Every episode remains directly reachable on the watch route.
- Episode row: one row per episode, at least 44px high, 13px text, fixed number column, truncated real title, current-row sky marker,
  and no independent rounded button cards.
- Mobile topology: video → previous/next → sticky range header → one-column episode list in document scroll; no 2/3-column grid.
- Title page: may keep a compact 50-episode grid for pre-play browsing, but uses the same ranges and never serves as the player's
  hidden “all episodes” dependency.
- Content boundary: borrow the mature fixed-player/ranged-list topology visible in the user's iQIYI references, not its brand,
  advertising, membership, story arcs or proprietary assets.

## Direction contract

THESIS: Watching never requires leaving the player to find another episode, and pausing never hides the primary way to resume.

OWN-WORLD: True-black video stage, cold-blue playlist, one-line factual navigation and a single sky active signal.

STORY: Confirm the episode, press the unmistakable play control, then move through the complete series from the same screen.

FIRST VIEWPORT: Video dominates while a compact one-line playlist exposes current range and episode at a glance.

FORM: HTMLVideoElement remains truth, hls.js resolves adaptive HLS, Plyr owns transport chrome, and the app owns player state/recovery overlays.

FINISH: real HLS, fixture long series, ready/paused/playing/error, keyboard, touch, 390/1024/1440 and no-overflow screenshots all pass.
