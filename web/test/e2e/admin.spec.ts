import { expect, test, type Page } from "@playwright/test";

const email = process.env.YTV_E2E_EMAIL || "test@example.com";
const password = process.env.YTV_E2E_PASSWORD || "Yueli-local-development-2026";

async function login(page: Page) {
  await page.goto("/manage");
  await expect(page).toHaveURL(/\/login(?:\?|$)/u);
  await page.waitForFunction(() => {
    const input = document.querySelector('form input[name="email"]');
    const form = input?.closest("form") as (HTMLFormElement & { __vueParentComponent?: unknown }) | null;
    return Boolean(form?.__vueParentComponent);
  });
  await page.getByLabel("邮箱").fill(email);
  await page.getByLabel("密码", { exact: true }).fill(password);
  await page.getByRole("button", { name: "登录", exact: true }).click();
  await page.waitForURL(/\/manage(?:\?|$)/u);
}

test("bootstrap administrator reaches the shared shell and real traffic", async ({ page }) => {
  await login(page);
  await expect(page.locator("[data-admin-shell]")).toBeVisible();
  await expect(page.getByRole("heading", { name: "控制台" })).toBeVisible();
  await expect(page.getByText("累计浏览", { exact: true })).toBeVisible();
  await expect(page.getByText("累计播放启动", { exact: true })).toBeVisible();
  const cookies = await page.context().cookies();
  expect(cookies.some((cookie) => /^ys_ytv-main_[0-9a-f]{12}$/u.test(cookie.name))).toBe(true);
});

test("a public query appears in administrator search analytics", async ({ page }) => {
  const query = `霜线验收${Date.now()}`;
  await page.goto(`/search?q=${encodeURIComponent(query)}`, { waitUntil: "networkidle" });
  await expect(page.getByText("没有找到可播放结果", { exact: true })).toBeVisible();
  await login(page);
  await page.goto("/manage");
  await expect(page.getByRole("heading", { name: "搜索洞察" })).toBeVisible();
  await expect(page.getByText(query, { exact: true })).toBeVisible();
  await expect(page.getByText("近 14 天搜索", { exact: true })).toBeVisible();
});

test("signed-in source preference survives reload", async ({ page }) => {
  await login(page);
  await page.goto("/", { waitUntil: "networkidle" });
  await page.getByRole("button", { name: "来源设置" }).click();
  const drawer = page.getByRole("dialog", { name: "来源设置" });
  await expect(drawer.getByText("已同步到账号", { exact: true })).toBeVisible();
  const switches = drawer.getByRole("switch");
  if (await switches.count() > 1) {
    const first = switches.first();
    const before = await first.getAttribute("aria-checked");
    const saved = page.waitForResponse((response) => response.request().method() === "PUT" && response.url().endsWith("/api/v1/preferences/sources"));
    await first.click();
    const saveResponse = await saved;
    expect(saveResponse.ok()).toBe(true);
    await page.reload({ waitUntil: "networkidle" });
    await page.getByRole("button", { name: "来源设置" }).click();
    await expect(page.getByRole("dialog", { name: "来源设置" }).getByRole("switch").first()).toHaveAttribute("aria-checked", before === "true" ? "false" : "true");
  }
});

test("administrator probes an installed source without a manual save action", async ({ page }) => {
  await login(page);
  await page.goto("/manage/sources");
  await expect(page.getByRole("heading", { name: "数据源" })).toBeVisible();
  await expect(page.getByLabel("本地演示源 名称")).toHaveValue("本地演示源");
  await page.getByRole("button", { name: "本地演示源 的操作" }).click();
  const probeButton = page.getByRole("menuitem", { name: "检测 本地演示源" });
  await probeButton.click();
  await expect(page.getByText("可用", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "保存 本地演示源" })).toHaveCount(0);
});

test("administrator can add sources from the bundled Kazumi catalog", async ({ page }) => {
  await login(page);
  await page.goto("/manage/sources", { waitUntil: "networkidle" });
  await expect(page.getByLabel("7sefun 名称")).toHaveValue("7sefun");
  await expect(page.getByText("Kazumi 1.3 · kazumi-7sefun", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "添加数据源" }).click();
  const dialog = page.getByRole("dialog", { name: "添加数据源" });
  await expect(dialog.getByText("全部来源已添加", { exact: true })).toBeVisible();
  await expect(dialog.getByText(/目录中的 \d+ 个来源都已在当前数据源列表中/u)).toBeVisible();
});

test("signed-in playback progress is merged into account history", async ({ page }) => {
  await login(page);
  await page.goto("/watch/fixture/ring-sea/0", { waitUntil: "networkidle" });
  const video = page.locator("[data-player-state] video");
  await video.evaluate(async (element: HTMLVideoElement) => {
    element.muted = true;
    await element.play();
  });
  await expect.poll(() => video.evaluate((element: HTMLVideoElement) => element.currentTime)).toBeGreaterThan(0);
  const saved = page.waitForResponse((response) => response.request().method() === "PUT" && response.url().includes("/api/v1/progress/"));
  await video.evaluate((element: HTMLVideoElement) => element.pause());
  expect((await saved).ok()).toBe(true);
  await page.goto("/history");
  await expect(page.getByRole("heading", { name: "播放历史" })).toBeVisible();
  await expect(page.getByText("环海之门", { exact: true })).toBeVisible();
  const seeded = await page.evaluate(async () => {
    const response = await fetch("/api/v1/progress/fixture/ring-sea/0", {
      method: "PUT",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        titleName: "环海之门", episodeTitle: "第 1 集",
        positionSeconds: 10, durationSeconds: 100, completed: false,
        updatedAt: new Date(Date.now() + 1000).toISOString()
      })
    });
    return response.ok;
  });
  expect(seeded).toBe(true);
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "继续观看" })).toBeVisible();
});

test("administrator persists the public site profile", async ({ page }) => {
  await login(page);
  await page.goto("/manage/settings");
  await expect(page.getByRole("heading", { name: "站点设置" })).toBeVisible();
  const name = page.getByLabel("站点名称");
  await expect(name).toHaveValue(/YTV/u);
  const saveButton = page.getByRole("button", { name: "保存公开资料" });
  await expect.poll(() => saveButton.evaluate((element: HTMLButtonElement & { __vueParentComponent?: unknown }) => Boolean(element.__vueParentComponent))).toBe(true);
  await saveButton.click();
  await expect(page.getByRole("button", { name: "已保存" })).toBeVisible();
});
