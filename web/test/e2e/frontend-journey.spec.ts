import { expect, test } from "@playwright/test";

test("search is directly reachable and history keeps the input and source scope in sync", async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem("ytv-search-history-v1", JSON.stringify(["霜线", "环海之门"])));
  await page.goto("/search?q=霜线&sources=fixture", { waitUntil: "networkidle" });
  await expect(page.locator("header").getByRole("link", { name: "搜索", exact: true })).toHaveCount(0);
  await page.locator("header").getByLabel("搜索想看的影视").focus();
  await page.getByRole("link", { name: "环海之门", exact: true }).click();
  await expect(page.locator("main").getByLabel("搜索想看的影视")).toHaveValue("环海之门");
  expect(new URL(page.url()).searchParams.get("sources")).toBe("fixture");
  await page.goBack();
  await expect(page.locator("main").getByLabel("搜索想看的影视")).toHaveValue("霜线");
  if ((page.viewportSize()?.width || 0) < 640) {
    await page.getByRole("button", { name: "打开导航", exact: true }).click();
    const drawer = page.getByRole("dialog", { name: "导航", exact: true });
    await drawer.getByLabel("搜索想看的影视").fill("环海之门");
    await drawer.getByRole("button", { name: "搜索 环海之门", exact: true }).click();
    await expect(drawer).toBeHidden();
    await expect(page.locator("main").getByLabel("搜索想看的影视")).toHaveValue("环海之门");
  }
});

test("a long synopsis stays below playback and the viewing journey returns to the original search", async ({ page }, info) => {
  await page.route("**/api/v1/titles/fixture/frost-line", async route => {
    const response = await route.fetch();
    const body = await response.json();
    body.title.summary = "一名摄影师重返多年未归的小镇，在冬日山林中追寻一封迟到的来信。沿途的人与往事让他重新理解家的含义。".repeat(16);
    await route.fulfill({ response, json: body });
  });
  await page.goto("/search?q=霜线&sources=fixture&page=1", { waitUntil: "networkidle" });
  const original = new URL(page.url()).pathname + new URL(page.url()).search;
  await page.locator('article a[href^="/title/fixture/frost-line"]').first().click();
  await expect(page.getByRole("link", { name: "返回搜索结果", exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "播放正片", exact: true })).toBeVisible();
  const expand = page.getByRole("button", { name: "展开简介", exact: true });
  await expect(expand).toHaveAttribute("aria-expanded", "false");
  await page.screenshot({ path: info.outputPath("detail.png"), fullPage: true, animations: "disabled" });
  await expand.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("button", { name: "收起简介", exact: true })).toHaveAttribute("aria-expanded", "true");
  await page.getByRole("button", { name: "收起简介", exact: true }).click();
  await page.getByRole("link", { name: "播放正片", exact: true }).click();
  await expect(page.locator("[data-player-state]")).toHaveAttribute("data-player-state", "ready");
  await page.getByRole("link", { name: "影片详情", exact: true }).click();
  await page.getByRole("link", { name: "返回搜索结果", exact: true }).click();
  await expect(page).toHaveURL(new URL(original, page.url()).href);
  await expect(page.locator("main").getByLabel("搜索想看的影视")).toHaveValue("霜线");
});

test("an empty later search page retains a route back to results", async ({ page }, info) => {
  await page.route("**/api/v1/search?**", async route => {
    const response = await route.fetch();
    const body = await response.json();
    if (new URL(route.request().url()).searchParams.get("page") === "2") body.groups = [];
    await route.fulfill({ response, json: body });
  });
  await page.goto("/search?q=霜线&sources=fixture", { waitUntil: "networkidle" });
  await page.getByRole("link", { name: "下一页", exact: true }).click();
  await expect(page.getByRole("heading", { name: "这一页没有更多结果" })).toBeVisible();
  await page.screenshot({ path: info.outputPath("empty-page.png"), fullPage: true, animations: "disabled" });
  await page.getByRole("link", { name: "上一页", exact: true }).click();
  await expect(page.getByRole("heading", { name: "霜线", exact: true })).toBeVisible();
});

test("the detail page resumes the last line without overriding a new line choice", async ({ page }, info) => {
  await page.addInitScript(() => localStorage.setItem("ytv-progress-v2:fixture:frost-line:line1-0", JSON.stringify({
    titleName: "霜线", episodeTitle: "正片", position: 125, duration: 900, completed: false, updatedAt: Date.now()
  })));
  await page.goto("/title/fixture/frost-line", { waitUntil: "networkidle" });
  await expect(page.getByRole("button", { name: "备用线路", exact: true })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("link", { name: "继续观看", exact: true })).toHaveAttribute("href", "/watch/fixture/frost-line/line1-0");
  await expect(page.getByText("正片 · 2:05", { exact: true }).filter({ visible: true })).toBeVisible();
  await page.screenshot({ path: info.outputPath("resume.png"), fullPage: true, animations: "disabled" });
  await page.getByRole("button", { name: "主线路", exact: true }).click();
  await expect(page.getByRole("link", { name: "播放正片", exact: true })).toHaveAttribute("href", "/watch/fixture/frost-line/0");
});

test("all matching sources remain reachable including long labels", async ({ page }, info) => {
  await page.route("**/api/v1/search?**", async route => {
    const response = await route.fetch();
    const body = await response.json();
    const first = body.groups[0]?.candidates[0];
    if (first) body.groups[0].candidates = Array.from({ length: 6 }, (_, index) => ({ ...first, sourceKey: `source-${index}`, externalId: `film-${index}`, sourceName: index === 5 ? "第六来源与较长的完整显示名称" : `来源 ${index + 1}` }));
    await route.fulfill({ response, json: body });
  });
  await page.goto("/search?sources=fixture", { waitUntil: "networkidle" });
  await page.locator("main").getByLabel("搜索想看的影视").fill("霜线");
  await page.locator("main").getByLabel("搜索想看的影视").press("Enter");
  const disclosure = page.locator("article summary").first();
  await expect(disclosure).toHaveText("选择来源 · 6 个");
  await disclosure.focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("link", { name: "第六来源与较长的完整显示名称", exact: true })).toBeVisible();
  await page.screenshot({ path: info.outputPath("sources-expanded.png"), fullPage: true, animations: "disabled" });
});
