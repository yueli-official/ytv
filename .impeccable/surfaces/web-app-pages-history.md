---
version: 1
slug: "web-app-pages-history"
primary_target: "web/app/pages/history.vue"
related_targets: ["web/app/components/HomeContinueWatching.client.vue", "web/app/composables/useProgressPosters.ts"]
---

# Playback history surface

- Scope and mode: public `/history` plus homepage Continue Watching; dense Operate mode.
- Audience and job: anonymous and signed-in viewers recognize a title by poster, see their latest episode and progress, and resume
  directly without reading dashboard-style status cards.
- Identity: `sourceKey + externalId` owns one visible card. When several episodes exist, the newest `updatedAt` wins.
- Poster enrichment: after mount, resolve each unique Source Title Locator through the title API, at most four concurrent and 24
  titles per history view. Failure never blocks the record or resume action.
- History grid: 2→3→4→5→6 portrait cards, 2:3 wells, percent badge, episode overlay and bottom-edge progress; title, source and
  recent time are the only facts below the poster. The whole card resumes playback.
- Continue Watching: only unfinished titles, maximum six, in a 15–16rem horizontal card with a narrow poster column and compact
  facts. The whole section remains at or below 180px high and links to the full history grid.
- Media integrity: never substitute a shared demo poster. Use the resolved real poster or the neutral photo placeholder with a
  fixed aspect ratio, lazy loading, async decoding and no-referrer policy.
- Responsive behavior: horizontal scrolling stays inside the Continue Watching rail; the history route itself never overflows.

## Direction contract

THESIS: A viewing record is recognized first as a film poster and only second as progress metadata.

OWN-WORLD: Cold-blue dark portal, portrait media, tiny factual overlays and one sky progress signal.

STORY: Recognize the title, confirm the latest episode and progress, then continue with one click.

FIRST VIEWPORT: Compact page context leads directly into the first row of poster history, not large dashboard cards.

FORM: Progress remains local-first and account merge remains asynchronous; poster enrichment is optional presentation data.

FINISH: Playwright proves de-duplication, real poster enrichment, whole-card resume, 390px overflow and failure fallback.
