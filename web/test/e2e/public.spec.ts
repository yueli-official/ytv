import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

test("browse-first homepage exposes recovery without horizontal page overflow", async ({ page }, testInfo) => {
  const errors: string[] = [];
  page.on("console", (message) => {
    if (message.type() === "error") errors.push(message.text());
  });
  page.on("pageerror", (error) => errors.push(error.message));

  await page.goto("/", { waitUntil: "networkidle" });
  await expect(page.getByRole("link", { name: "YTV 首页" })).toBeVisible();
  await expect(page.locator("#main-content").getByRole("search").getByLabel("搜索想看的影视")).toBeVisible();
  await expect(page.getByRole("heading", { name: "豆瓣热门" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "来源更新" })).toBeVisible();
  expect(await page.locator("[data-poster-grid] li").count()).toBeGreaterThanOrEqual(6);

  const dimensions = await page.evaluate(() => ({ width: document.documentElement.scrollWidth, viewport: window.innerWidth }));
  expect(dimensions.width).toBe(dimensions.viewport);
  expect(errors).toEqual([]);

  const accessibility = await new AxeBuilder({ page }).analyze();
  expect(accessibility.violations.filter((item) => ["serious", "critical"].includes(item.impact || ""))).toEqual([]);
  await page.screenshot({ path: testInfo.outputPath("home.png"), fullPage: true });
});

test("search query is route state", async ({ page }) => {
  await page.goto("/", { waitUntil: "networkidle" });
  const search = page.locator("#main-content").getByRole("search").getByLabel("搜索想看的影视");
  await search.fill("霜线");
  await search.press("Enter");
  await expect(page).toHaveURL(/\/search\?q=%E9%9C%9C%E7%BA%BF/u);
  await expect(page.locator("#main-content").getByRole("heading", { name: "搜索影视", level: 1 })).toBeVisible();
  await expect(page.locator("#main-content").getByRole("heading", { name: "霜线", level: 2 })).toBeVisible();
});

test("same-origin fixture manifest reaches the HLS player", async ({ page }) => {
  await page.goto("/watch/fixture/ring-sea/0", { waitUntil: "networkidle" });
  const player = page.locator("[data-player-state]");
  await expect(player).toBeVisible();
  const video = player.locator("video");
  await expect(video).toHaveCount(1);
  await video.evaluate(async (element: HTMLVideoElement) => {
    element.muted = true;
    await element.play();
  });
  await expect(player).toHaveAttribute("data-player-state", "playing");
  await expect.poll(() => video.evaluate((element: HTMLVideoElement) => element.currentTime)).toBeGreaterThan(0);
});

test("fatal manifest failure stops loading and offers recovery", async ({ page }) => {
  await page.route("**/demo/hls/index.m3u8", (route) => route.abort("failed"));
  await page.goto("/watch/fixture/ring-sea/0", { waitUntil: "domcontentloaded" });
  const player = page.locator("[data-player-state]");
  await expect(player).toHaveAttribute("data-player-state", "error", { timeout: 30_000 });
  await expect(player.getByRole("heading", { name: "播放失败" })).toBeVisible();
  await expect(player.getByRole("button", { name: "重试当前来源" })).toBeVisible();
});

test("watch keeps long-series navigation and a primary play action inside the player", async ({ page }) => {
  await page.goto("/watch/fixture/long-voyage/61", { waitUntil: "networkidle" });
  await expect(page.locator("[data-primary-play]")).toBeVisible();
  await expect(page.getByLabel("剧集范围")).toBeVisible();
  const rows = page.locator("[data-player-episode-row]");
  await expect(rows).toHaveCount(50);
  const first = await rows.nth(0).boundingBox();
  const second = await rows.nth(1).boundingBox();
  expect(first?.x).toBe(second?.x);
  expect((second?.y || 0) - (first?.y || 0)).toBeGreaterThanOrEqual(40);
  await expect(page.getByText("查看全部剧集", { exact: true })).toHaveCount(0);
});

test("homepage separates search history, source refresh and Douban discovery", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("button", { name: /本地演示源/u })).toHaveAttribute("aria-pressed", "true");
  const search = page.locator("#main-content").getByLabel("搜索想看的影视");
  await search.fill("环海之门");
  await search.press("Enter");
  await page.waitForURL(/\/search\?/u);
  await page.waitForFunction(() => {
    const value = JSON.parse(localStorage.getItem("ytv-search-history-v1") || "[]");
    return Array.isArray(value) && value.includes("环海之门");
  });
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "搜索历史" })).toBeVisible();
  await expect(page.getByRole("link", { name: "环海之门", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "换一批资源" })).toBeVisible();
  const hot = page.getByRole("heading", { name: "豆瓣热门" });
  await expect(hot).toBeVisible();
  const searchPanel = page.getByRole("heading", { name: "影视搜索" }).locator("xpath=ancestor::section[1]");
  await expect(searchPanel.getByRole("heading", { name: "豆瓣热门" })).toHaveCount(0);
});

test("portal aligns search and uses compact poster playback records", async ({ page }) => {
  await page.addInitScript(() => {
    localStorage.setItem("ytv-progress-v2:fixture:ring-sea:0", JSON.stringify({
      titleName: "环海之门",
      episodeTitle: "第 1 集",
      position: 120,
      duration: 1200,
      updatedAt: Date.now(),
      completed: false
    }));
  });

  await page.goto("/search?q=霜线&sources=fixture", { waitUntil: "networkidle" });
  await expect(page.locator("header").getByRole("link", { name: "浏览", exact: true })).toHaveCount(0);
  const searchBox = await page.locator("[data-search-workspace]").getByRole("search").boundingBox();
  const summary = await page.locator("[data-search-summary]").boundingBox();
  expect(Math.abs((searchBox?.x || 0) - (summary?.x || 0))).toBeLessThanOrEqual(1);
  expect(Math.abs((searchBox?.width || 0) - (summary?.width || 0))).toBeLessThanOrEqual(1);

  await page.goto("/", { waitUntil: "networkidle" });
  const continueWatching = page.locator("[data-continue-watching]");
  await expect(continueWatching).toBeVisible();
  expect((await continueWatching.boundingBox())?.height || 0).toBeLessThanOrEqual(180);
  expect(await page.locator("[data-douban-grid] img").count()).toBeGreaterThanOrEqual(6);

  await page.goto("/history", { waitUntil: "networkidle" });
  await expect(page.locator("[data-history-poster-grid]")).toBeVisible();
  await expect(page.getByAltText("环海之门 海报")).toBeVisible();
  expect(await page.locator("[data-history-poster-grid] > li").count()).toBe(1);
});
