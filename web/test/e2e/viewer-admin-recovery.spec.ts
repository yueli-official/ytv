import { expect, test, type Page } from "@playwright/test";

async function login(page: Page) {
  await page.goto("/manage", { waitUntil: "networkidle" });
  await page.getByLabel("邮箱").fill(process.env.YTV_E2E_EMAIL || "test@example.com");
  await page.getByLabel("密码", { exact: true }).fill(process.env.YTV_E2E_PASSWORD || "Yueli-local-development-2026");
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL(/\/manage(?:\?|$)/u);
}

test("account progress resumes on a browser with no local history", async ({ page }) => {
  await login(page);
  const saved = await page.request.put("/api/v1/progress/fixture/frost-line/0", { data: {
    titleName: "霜线", episodeTitle: "正片", positionSeconds: 25, durationSeconds: 60,
    completed: false, updatedAt: new Date().toISOString()
  } });
  expect(saved.ok()).toBe(true);
  await page.evaluate(() => localStorage.clear());
  await page.route("**/api/v1/playback/fixture/frost-line/0", async route => {
    const response = await route.fetch();
    await route.fulfill({ response, json: { ...await response.json(), engine: "direct", mediaType: "video/mp4", url: "/demo/feature.mp4" } });
  });
  await page.goto("/title/fixture/frost-line", { waitUntil: "networkidle" });
  await page.getByRole("link", { name: "继续观看", exact: true }).click();
  await expect.poll(() => page.locator("[data-player-state] video").evaluate((video: HTMLVideoElement) => Math.round(video.currentTime))).toBe(25);
});

test("local history stays usable when account history is unavailable", async ({ page }, info) => {
  await login(page);
  await page.evaluate(() => localStorage.setItem("ytv-progress-v2:fixture:ring-sea:0", JSON.stringify({
    sourceKey: "fixture", externalId: "ring-sea", episodeKey: "0", titleName: "环海之门", episodeTitle: "第 1 集",
    position: 30, duration: 120, completed: false, updatedAt: Date.now()
  })));
  await page.route("**/api/v1/progress?*", route => route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ status: 503, code: "ytv.unavailable", title: "Unavailable" }) }));
  await page.goto("/", { waitUntil: "networkidle" });
  await page.getByRole("link", { name: "播放历史", exact: true }).last().click();
  await expect(page.getByText("账号历史暂时不可用", { exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "环海之门" })).toBeVisible();
  await expect(page.locator('[data-history-records] a[href="/watch/fixture/ring-sea/0"]').first()).toBeVisible();
  await page.screenshot({ path: info.outputPath("history-fallback.png"), fullPage: true, animations: "disabled" });
});

test("admin source catalog failure offers retry instead of false empty success", async ({ page }, info) => {
  await login(page);
  let fail = true;
  await page.route("**/api/v1/admin/source-catalog", route => fail ? route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ status: 503, code: "ytv.unavailable", title: "Unavailable" }) }) : route.continue());
  await page.goto("/manage", { waitUntil: "networkidle" });
  // Client navigation keeps the fault at the browser's API boundary.
  if ((page.viewportSize()?.width || 0) < 1024) {
    await page.getByRole("button", { name: "打开侧边栏", exact: true }).click();
  }
  await page.getByRole("link", { name: "数据源", exact: true }).filter({ visible: true }).click();
  await page.getByRole("button", { name: "添加数据源", exact: true }).click();
  const dialog = page.getByRole("dialog", { name: "添加数据源" });
  await expect(dialog.getByText("来源目录暂时不可用", { exact: true })).toBeVisible();
  await expect(dialog.getByText("全部来源已添加", { exact: true })).toHaveCount(0);
  await expect(dialog.getByRole("button", { name: "添加所选" })).toBeDisabled();
  await page.screenshot({ path: info.outputPath("catalog-error.png"), fullPage: true, animations: "disabled" });
  fail = false;
  const retry = page.waitForResponse(r => r.url().endsWith("/api/v1/admin/source-catalog") && r.ok());
  await dialog.getByRole("button", { name: "重试", exact: true }).click();
  await retry;
  await expect(dialog.getByText("来源目录暂时不可用", { exact: true })).toHaveCount(0);
});

test("admin source edits persist and console survives refresh", async ({ page }, info) => {
  await login(page);
  await page.reload({ waitUntil: "networkidle" });
  await expect(page.getByRole("img", { name: /^独立访客趋势/u })).toBeVisible();
  await page.screenshot({ path: info.outputPath("console.png"), fullPage: true, animations: "disabled" });
  await page.goto("/manage/sources", { waitUntil: "networkidle" });
  const response = await page.request.get("/api/v1/admin/sources");
  const source = (await response.json()).items.find((item: { key: string }) => item.key === "fixture");
  const input = page.getByLabel(`${source.name} 优先级`, { exact: true });
  try {
    const saved = page.waitForResponse(r => r.url().endsWith("/api/v1/admin/sources/fixture") && r.request().method() === "PATCH");
    await input.fill(String(source.priority + 1));
    expect((await saved).ok()).toBe(true);
    await page.reload({ waitUntil: "networkidle" });
    await expect(input).toHaveValue(String(source.priority + 1));
    await page.screenshot({ path: info.outputPath("sources.png"), fullPage: true, animations: "disabled" });
    await page.goto("/manage/settings", { waitUntil: "networkidle" });
    await page.screenshot({ path: info.outputPath("settings.png"), fullPage: true, animations: "disabled" });
  } finally {
    const listed = await page.request.get("/api/v1/admin/sources");
    const current = (await listed.json()).items.find((item: { key: string }) => item.key === "fixture");
    const restored = await page.request.patch("/api/v1/admin/sources/fixture", { data: { name: source.name, enabled: source.enabled, homeVisible: source.homeVisible, priority: source.priority, revision: current.revision } });
    expect(restored.ok()).toBe(true);
  }
});
