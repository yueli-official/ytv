import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

test("anonymous source preference stays in the browser", async ({ page }) => {
  await page.goto("/", { waitUntil: "networkidle" });
  const settingsButton = page.getByRole("button", { name: "来源设置" });
  await settingsButton.click();
  const drawer = page.getByRole("dialog", { name: "来源设置" });
  await expect(drawer).toBeVisible();
  await expect(drawer.getByText("保存在此浏览器", { exact: true })).toBeVisible();
  const switches = drawer.getByRole("switch");
  if (await switches.count() > 1) {
    const first = switches.first();
    const before = await first.getAttribute("aria-checked");
    await first.click();
    await page.reload({ waitUntil: "networkidle" });
    await page.getByRole("button", { name: "来源设置" }).click();
    await expect(page.getByRole("dialog", { name: "来源设置" }).getByRole("switch").first()).toHaveAttribute("aria-checked", before === "true" ? "false" : "true");
  }
});

test("browse-first homepage exposes recovery without horizontal page overflow", async ({ page }, testInfo) => {
  const errors: string[] = [];
  page.on("console", (message) => {
    if (message.type() === "error") errors.push(message.text());
  });
  page.on("pageerror", (error) => errors.push(error.message));

  await page.goto("/", { waitUntil: "networkidle" });
  await expect(page.getByRole("link", { name: "YTV 首页" })).toBeVisible();
  await expect(page.locator("#main-content").getByRole("search").getByLabel("搜索想看的影视")).toBeVisible();
  await expect(page.getByRole("heading", { name: "大家在看" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "片库新上架" })).toBeVisible();
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
  await expect(page.locator("#main-content").getByRole("heading", { name: "霜线", level: 2, exact: true })).toBeVisible();
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
  await expect(player.getByRole("heading", { name: "播放暂时中断" })).toBeVisible();
  await expect(player.getByRole("button", { name: "重新连接" })).toBeVisible();
});

test("watch keeps long-series navigation and a primary play action inside the player", async ({ page }) => {
  await page.goto("/watch/fixture/long-voyage/61", { waitUntil: "networkidle" });
  await expect(page.locator("video-neutral-skin").getByRole("button", { name: "播放", exact: true }).first()).toBeVisible();
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
  await expect(page.getByRole("heading", { name: "搜索来源" })).toHaveCount(0);
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
  const hot = page.getByRole("heading", { name: "大家在看" });
  await expect(hot).toBeVisible();
  const sourceUpdates = page.getByRole("heading", { name: "片库新上架" });
  const hotBox = await hot.boundingBox();
  const sourceBox = await sourceUpdates.boundingBox();
  expect(hotBox?.y || 0).toBeLessThan(sourceBox?.y || 0);
  const searchPanel = page.locator(".ytv-page-heading");
  await expect(searchPanel.getByRole("heading", { name: "大家在看" })).toHaveCount(0);
});

test("homepage carousel rails use side controls without native scrollbars", async ({ page }, testInfo) => {
  await page.addInitScript(() => {
    for (let index = 0; index < 6; index += 1) {
      localStorage.setItem(`ytv-progress-v2:fixture:carousel-${index}:0`, JSON.stringify({
        titleName: `继续观看 ${index + 1}`,
        episodeTitle: "第 1 集",
        position: 60,
        duration: 600,
        updatedAt: Date.now() - index,
        completed: false
      }));
    }
  });
  await page.goto("/", { waitUntil: "networkidle" });

  const continueRail = page.locator("[data-continue-watching] .continue-rail");
  const doubanRail = page.locator("[data-douban-grid]");
  await expect(continueRail).toBeVisible();
  await expect(doubanRail).toBeVisible();
  expect(await continueRail.evaluate((element) => getComputedStyle(element).scrollbarWidth)).toBe("none");
  expect(await doubanRail.evaluate((element) => getComputedStyle(element).scrollbarWidth)).toBe("none");

  if (testInfo.project.name !== "mobile") {
    const continueNext = page.getByRole("button", { name: "下一组继续观看" });
    const doubanNext = page.getByRole("button", { name: "下一组豆瓣热门" });
    await expect(continueNext).toBeVisible();
    await expect(doubanNext).toBeVisible();
    const continueBox = await continueRail.boundingBox();
    const continueButtonBox = await continueNext.boundingBox();
    expect(continueButtonBox?.x || 0).toBeGreaterThan((continueBox?.x || 0) + (continueBox?.width || 0) - 60);
    await continueNext.click();
    await expect.poll(() => continueRail.evaluate((element) => element.scrollLeft)).toBeGreaterThan(0);
    await expect(page.getByRole("button", { name: "上一组继续观看" })).toBeVisible();
  }
  await page.screenshot({ path: testInfo.outputPath("carousel-rails.png"), fullPage: true });
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
