# YTV production release

Production uses `compose.production.yaml` with the existing `yueli-core-db` PostgreSQL provider. The Web binds only `127.0.0.1:13013`; API stays on its private network. Configure the existing `tv.yuelili.com` HTTPS proxy only after both candidates pass.

Set explicit `YTV_API_IMAGE` and `YTV_WEB_IMAGE` in the release `.env`. Keep private `api.env` and `web.env` mode 0600. Do not use the development fixture or copy a local database. API configuration includes a dedicated database role, stable site key, HTTPS Identity issuer/JWKS and validated source/playback/image allowlists. Web configuration includes the matching HTTPS callback and stable seal secret. Shared Identity is not recreated by this Compose.

Leave `YTV_BOOTSTRAP_ADMIN_SUBS` unset for first deployment. The first signed-in user who explicitly clicks **初始化站点并成为管理员** at `/manage/setup` becomes the administrator through Foundation's atomic, persistent claim. Login alone does not grant privileges. Existing bootstrap deployments remain supported; later upgrades never reopen initial claim.

For an existing database, compare applied migration bytes and back up before applying only new migrations. Do not rerun initial database/client provisioning. Save the paired release manifest, dependency receipt, artifact digests, candidate and live checks, and run `verify_release.py` with the current Workspace baseline. Preserve the previous image IDs and proxy configuration for rollback; do not automatically roll back a database.

Current deployment evidence: [2026-10-23 movie upgrade](../docs/deployment-20261023-movie.md).
