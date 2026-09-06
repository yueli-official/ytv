import { expect, test, type Page } from "@playwright/test";

async function login(page: Page) {
  await page.goto("/manage", { waitUntil: "domcontentloaded" });
  await expect(page).toHaveURL(/\/login(?:\?|$)/u);
  await page.waitForFunction(() => Boolean((document.querySelector('form input[name="email"]')?.closest("form") as HTMLFormElement & { __vueParentComponent?: unknown })?.__vueParentComponent));
  await page.getByLabel("邮箱").fill(process.env.YTV_E2E_EMAIL || "test@example.com");
  await page.getByLabel("密码", { exact: true }).fill(process.env.YTV_E2E_PASSWORD || "Yueli-local-development-2026");
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL(/\/manage(?:\?|$)/u);
}

test("public results use collections and safe Problems", async ({ page }) => {
  const sources = await page.request.get("/api/v1/sources");
  expect(sources.status()).toBe(200);
  const collection = await sources.json();
  expect(Array.isArray(collection.items)).toBe(true);
  expect(collection).not.toHaveProperty("sources");
  for (const [path, status, code] of [
    ["/api/v1/search?q=", 400, "common.validation_failed"],
    ["/api/v1/missing-contract-route", 404, "ytv.not_found"],
  ] as const) {
    const response = await page.request.get(path);
    expect(response.status()).toBe(status);
    expect(response.headers()["content-type"]).toContain("application/problem+json");
    const failure = await response.json();
    expect(failure).toMatchObject({ status, code });
    expect(failure.traceId).toBeTruthy();
    expect(response.headers()["x-trace-id"]).toBe(failure.traceId);
    expect(failure).not.toHaveProperty("message");
  }
  await page.goto("/search?q=环海&sources=fixture", { waitUntil: "networkidle" });
  await expect(page.getByText("环海", { exact: false }).first()).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBe(true);
});

test("administrator sees a real revision conflict and request trace", async ({ page }) => {
  test.setTimeout(120_000);
  await login(page);
  await page.goto("/manage/settings", { waitUntil: "networkidle" });
  const originalResponse = await page.request.get("/api/v1/admin/site-profile");
  expect(originalResponse.status()).toBe(200);
  const original = (await originalResponse.json()).profile;
  const update = await page.request.put("/api/v1/admin/site-profile", { data: { name: original.name, description: original.description, revision: original.revision } });
  expect(update.status()).toBe(200);
  const current = (await update.json()).profile;
  try {
    await page.getByLabel("站点名称").fill("待保存的站点名称");
    const response = page.waitForResponse(item => item.url().endsWith("/api/v1/admin/site-profile") && item.request().method() === "PUT");
    await page.getByRole("button", { name: "保存公开资料" }).click();
    expect((await response).status()).toBe(409);
    await expect(page.getByText("设置已被更新，请刷新后再修改。", { exact: true })).toBeVisible();
    await page.getByText("技术详情", { exact: true }).click();
    await expect(page.locator("details code")).toContainText("ytv.revision_conflict");
  } finally {
    const restored = await page.request.put("/api/v1/admin/site-profile", { data: { name: original.name, description: original.description, revision: current.revision } });
    expect(restored.status()).toBe(200);
  }
});

test("source installation returns 201 and its resource location", async ({ page }) => {
  test.setTimeout(120_000);
  await login(page);
  const listed = await page.request.get("/api/v1/admin/sources");
  expect(listed.status()).toBe(200);
  const source = (await listed.json()).items.find((item: { key: string }) => item.key === "fixture");
  expect(source).toBeTruthy();
  const removed = await page.request.delete("/api/v1/admin/sources/fixture", { data: { revision: source.revision } });
  expect(removed.status()).toBe(200);
  const available = (await removed.json()).source;
  try {
    const installed = await page.request.post("/api/v1/admin/sources", { data: { sourceKey: "fixture", revision: available.revision } });
    expect(installed.status()).toBe(201);
    expect(installed.headers().location).toBe("/api/v1/admin/sources/fixture");
    const result = (await installed.json()).source;
    const conflict = await page.request.post("/api/v1/admin/sources", { data: { sourceKey: "fixture", revision: result.revision } });
    expect(conflict.status()).toBe(409);
    expect((await conflict.json()).code).toBe("ytv.source_already_installed");
  } finally {
    const catalog = await page.request.get("/api/v1/admin/source-catalog");
    const missing = (await catalog.json()).items.find((item: { key: string }) => item.key === "fixture");
    if (missing) {
      const restored = await page.request.post("/api/v1/admin/sources", { data: { sourceKey: "fixture", revision: missing.revision } });
      expect(restored.status()).toBe(201);
    }
  }
});
