# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- Visitors who want to find a film, series, variety show, documentary or animation across several approved external sources and
  start watching without first creating an account.
- Returning viewers who use local history anonymously or sign in to continue the same episode and position across devices.
- Site administrators who monitor real usage, maintain the installed, enabled and prioritized approved source set, inspect search gaps, update site
  presentation and govern access.

## Product Purpose

YTV gives viewers one stable place to search multiple approved external catalogs, understand which sources and episodes are
currently playable, recover from a failing source and continue watching. Success means source volatility is absorbed by YTV's
discovery and playback modules instead of becoming repeated manual work for the viewer.

## Positioning

YTV is a source-resilient viewing workspace rather than a generic movie database, editorial streaming catalog or video community.
Its distinctive mechanism is the explicit chain from one query to grouped source candidates, a stable source title locator, a
resolved episode and a recoverable playback session, while YTV itself does not host the video stream.

## Operating Context

- Visitors arrive to search directly, revisit local search history, choose which globally enabled sources participate in their
  search, use a small set of external discovery tags or continue an unfinished episode.
- YTV's catalog data comes from approved internet sources rather than an owned editorial library. The homepage may show an
  explicitly attributed external trending list and current source updates; neither is presented as a YTV recommendation.
- Search fans out across enabled Source Definitions under one deadline and may return partial results when a source is degraded.
- A viewer can inspect a title, choose an episode, start playback and switch to another approved source when recovery is needed.
- Anonymous progress remains local-first. A signed-in viewer's progress synchronizes asynchronously without blocking playback.
- Administrators use the shared Yueli console to inspect traffic, playback starts, search terms, zero-result searches and source
  health, and to change only Source Settings rather than source endpoints or scraping logic.

## Capabilities and Constraints

- The implementation stack is Nuxt 4, Nuxt UI 4, Tailwind CSS 4, GoFrame 2 and PostgreSQL, using versions locked by the adjacent
  Workspace at implementation time.
- Public search and playback are anonymous. Identity adds account controls and cross-device history; it does not gate ordinary
  discovery or playback.
- Source Definitions, endpoints, adapter kinds, scraping selectors and credentials are code/deployment allowlist facts. The product
  UI may install or remove an approved definition, rename its presentation, enable or disable it, prioritize it, choose homepage
  visibility and probe one or all installed sources. It never accepts an endpoint, playback host or image host.
- The approved source catalog may include definitions compiled at startup from an immutable KazumiRules commit selected by deployment.
  Catalog failure or one invalid rule does not block YTV or existing sources; imported definitions remain uninstalled and disabled until an
  administrator explicitly adds them, and their compatibility notices stay visible in the source collection.
- Only sources that the operator has the right to use may enter the allowlist. Technical reachability is not evidence of permission.
- The current local approved catalog contains the four installed definitions plus three initially uninstalled candidates validated
  from the user-supplied configuration: 樱花资源、索尼资源 and 优质资源库. 卧龙、旺旺 and the Omo placeholder remain excluded
  because they did not return a usable HTTPS catalog response.
- YTV does not upload, transcode, proxy or persist third-party video streams. Playback manifests are resolved on demand and full
  signed media URLs are not stored in product data, logs or telemetry. Douban poster delivery is a separate, protocol-specific
  read-through route: it accepts only a validated poster file ID, constructs one fixed provider path, rejects redirects and
  non-image or oversized responses, and never accepts an arbitrary URL.
- Similar search results may be grouped for one presentation, but automated grouping never becomes permanent title identity.
  `sourceKey + externalID` remains the authoritative Source Title Locator.
- The administration MVP contains Dashboard, Sources and Site Settings. Search insights belong to the Dashboard rather than a separate route. Identity and Foundation Authorization
  still protect every management route, but YTV does not expose a placeholder governance page without application, role or user
  workflows. A film CMS, editorial recommendations, community comments, product-authored ratings, payments and subtitle uploads
  are outside the current scope.
- Foundation Traffic records page views and playback starts in separate instances. Search analytics are best-effort and must not
  turn a valid search into an error response.
- YTV-owned logos, Open Graph artwork and operational media use Asset `mediaKey` contracts. External stream and poster handling
  does not bypass Asset ownership or create an arbitrary-fetch proxy.
- The legacy checkout `E:\projects\yozya\tv` is read-only evidence. Product implementation belongs in this repository; Workspace
  owns only local orchestration contracts.

## Brand Commitments

- The product name is `YTV`, with a white wordmark and compact horizontal navigation.
- The user approved a new five-page cinematic direction on 2026-10-23 and explicitly authorized implementation. This supersedes the earlier portal, champagne and mint-gallery visual attempts. Current visual rules live in `DESIGN.md`.
- Public browse, search, detail, history and playback remain dark. Pink identifies primary actions; pale blue identifies selections and progress; amber is reserved for actual attributed scores. Shared account and administration retain their own appearance controls.
- Real external artwork and factual metadata lead the design. A blurred copy of a supplied portrait may create depth, but no fictional backdrop, cast, availability, ratings or recommendations are invented.
- Homepage sections separate featured discovery, optional Continue Watching, attributed Douban trending and actual source updates. Trending and source paging remain independent. Source preferences live in the header drawer.
- Search groups preserve every candidate source and the return query/page/scope. Source title locators remain authoritative even for matching names.
- History uses chronological rows, real poster enrichment and explicit continue actions. Filtering applies only to loaded records. Unknown duration never produces a misleading percentage.
- Detail separates playback actions and synopsis from episodes. Long series use stable 50-entry ranges. Playback uses a dominant 16:9 viewport and a single-column playlist, with visible controls and at least 44px episode targets.
- Use Foundation/Nuxt UI controls, Tabler icons, accessible focus and motion preferences. Familiar category conventions do not authorize copying third-party branding or implementation.

## Evidence on Hand

- Legacy product and implementation: `E:\projects\yozya\tv`, commit
  `04f50d6d43f40e4eca1e366ea039c27dec2a8c3a`.
- Legacy visual evidence: `preview/home.png`, `preview/search.png`, `preview/player.png` and `preview/history.png`.
- Confirmed implementation plan: `../workspace/flightdeck/work/2026-08-29-ytv-site-group-upgrade/references/research-and-upgrade-plan.md`.
- Player research from hls.js, Plyr, MDN and W3C primary sources:
  `../workspace/flightdeck/work/2026-08-29-ytv-site-group-upgrade/references/player-research.md`.
- Existing Yueli products provide verified Identity, Traffic, Asset and administration integration patterns, but their product
  visual worlds and domain models are not YTV templates.
- Netflix's official 2025 TV experience announcement is the craft reference for visible top shortcuts, front-loaded choice
  information, responsive discovery and a clean modern homepage:
  `https://about.netflix.com/en/news/unveiling-our-innovative-new-tv-experience`.
- No user counts, availability guarantee, customer claims, licensed title catalog or historical traffic dataset has been supplied;
  future UI and documentation must not fabricate them.
- The legacy README declares MIT but the repository contains no tracked `LICENSE`; repository licensing remains an explicit release
  decision rather than an inferred fact.

## Product Principles

- Absorb source failure; do not export it as viewer confusion.
- Keep source identity explicit and automated title grouping reversible.
- Let anonymous viewing work immediately, then make signing in earn its place through durable history.
- Reuse platform capabilities and keep YTV-specific complexity behind small, testable interfaces.
- Count only clearly named events and retain the minimum data needed to operate the product.
- Keep viewer search history local and bounded. Anonymous disabled-source preferences stay in LocalStorage; signed-in preferences synchronize to the YTV account and migrate the browser choice on first use. A viewer preference cannot re-enable an administrator-disabled Source Definition.

## Accessibility & Inclusion

- Public and administration surfaces support keyboard navigation, visible focus, semantic landmarks, reduced motion and WCAG AA
  contrast in each surface's supported appearance mode.
- Player actions remain available as visible controls; gestures are enhancements, not the only way to seek, change episodes or
  recover playback.
- Captions, audio tracks, autoplay rejection, buffering and fatal errors expose understandable state without high-frequency screen
  reader noise.
- CLI Playwright acceptance covers 390, 768, 1024 and 1440 pixel viewports, Chromium/Firefox/WebKit where applicable, and Axe
  serious/critical findings at zero.

## Movie playback upgrade

Film viewing distinguishes a content item from alternative delivery lines. Single-item lines use a feature-film action; series retain episode navigation. The site uses Video.js 10 through a YTV-owned playback session for fresh URL resolution, progress, retry and cleanup. Integration details and current boundaries are documented in [the playback contract](docs/research/playback-integration.md).
