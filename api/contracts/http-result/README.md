# YTV HTTP Result

`error-catalog.json` owns public product error codes, statuses and translations. `operation-errors.json` declares reachable errors for the route-derived operations. API request/response Go types own paths and success schemas. Do not edit generated Go, TS, OpenAPI or operations files.

From `api/`, generate/check the complete project using Foundation's local Project CLI: `httpcontract -project contracts/http-result/project.json [-check]`. Project v1 is not in the pinned v0.4.1 release yet. Until its release, CI separately verifies route-derived OpenAPI freshness, released catalog generation and compatibility; the local Project check additionally verifies operations derivation.

This pre-release migration intentionally changes source collections and probe batches from `sources`/`results` to `items`, installation success to 201 with Location, validation to `common.validation_failed`, and domain-specific unavailable/conflict errors. The web adapter retains its product-facing source view model without accepting old wire shapes.

Discovery search remains a resource aggregating per-source outcomes. Trending and recent progress are bounded collections: their upstream/service limits do not imply a known total, so the API does not invent total/page-size values. Successful empty collections use `[]`.

Playback allowlists, source definitions, signed-in identity and cookie handling remain owned by their existing modules. Product errors wrap original causes; unknown storage errors stay internal. Frontend feedback uses generated codes and safe local wording, including field errors and trace details.

Human review is deferred until site upgrades are complete, per user instruction. Automated checks and focused real-composition CLI Playwright are still required. Deployment, new Foundation/Identity releases and remote CI execution are separate from this local migration.
