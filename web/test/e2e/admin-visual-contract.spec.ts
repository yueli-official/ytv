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

test("settings uses the shared canvas page hierarchy", async ({ page }) => {
  await login(page);
  await page.goto("/manage/settings");

  const pageTitle = page.locator("main h1").filter({ hasText: /^站点设置$/u });
  await expect(pageTitle).toBeVisible();
});

test("shared console shell renders without hydration or component warnings", async ({ page }) => {
  const problems: string[] = [];
  page.on("console", (message) => {
    if (message.type() === "error" || message.type() === "warning") problems.push(message.text());
  });
  page.on("pageerror", (error) => problems.push(error.message));

  await login(page);
  await page.goto("/manage", { waitUntil: "networkidle" });
  await expect(page.locator("[data-admin-console]")).toBeVisible();
  await expect(page.locator("[data-admin-console-canvas]")).toBeVisible();
  await expect(page.getByRole("img", { name: /^独立访客趋势/u })).toBeVisible();
  await page.getByRole("tab", { name: "浏览量", exact: true }).click();
  await expect(page.getByRole("img", { name: /^浏览量趋势/u })).toBeVisible();
  await page.getByRole("tab", { name: "7 天" }).click();
  await expect(page.getByText("近 7 天访客日", { exact: true })).toBeVisible();
  await expect(page.getByRole("img", { name: /^浏览量趋势/u })).toBeVisible();

  if ((page.viewportSize()?.width || 0) >= 1024) {
    const sidebar = page.locator("[data-admin-sidebar-appearance]");
    await expect(sidebar).toBeVisible();
  }

  expect(problems).toEqual([]);
});

test("administration exposes only real editable YTV workflows", async ({ page }) => {
  await login(page);
  await page.goto("/manage/settings");
  await expect(page.getByRole("tab", { name: "发现内容" })).toHaveCount(0);
  await expect(page.getByRole("tab", { name: "播放偏好" })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "权限治理" })).toHaveCount(0);
  await page.goto("/manage/sources");
  await expect(page.getByRole("switch", { name: /在首页展示/u }).first()).toBeVisible();
});

test("source collection supports CRUD auto-save and bulk probing in aligned columns", async ({ page }) => {
  await login(page);
  await page.goto("/manage/sources");

  await expect(page.getByRole("button", { name: "添加数据源" })).toBeVisible();
  await expect(page.getByRole("button", { name: "来源工具" })).toBeVisible();
  await expect(page.getByRole("button", { name: /保存 .*资源/u })).toHaveCount(0);

  await page.getByRole("checkbox", { name: /选择 .*资源/u }).first().check();
  const bulkRegion = page.getByRole("region", { name: "来源批量操作" });
  await expect(bulkRegion).toBeVisible();
  await expect(bulkRegion.getByRole("combobox", { name: "批量操作" })).toBeVisible();
  await expect(bulkRegion.getByRole("button", { name: "应用" })).toBeDisabled();
  await bulkRegion.getByRole("button", { name: "清除选择" }).click();

  await page.getByRole("button", { name: "筛选" }).click();
  await expect(page.getByText("参与搜索", { exact: true }).last()).toBeVisible();
  await expect(page.getByText("首页展示", { exact: true }).last()).toBeVisible();
  await page.keyboard.press("Escape");

  await page.getByRole("button", { name: "添加数据源" }).click();
  const addDialog = page.getByRole("dialog", { name: "添加数据源" });
  if (await addDialog.getByRole("checkbox", { name: "选择全部可添加来源" }).count()) {
    await addDialog.getByRole("checkbox", { name: "选择全部可添加来源" }).check();
    await expect(addDialog.getByRole("button", { name: /添加所选（\d+）/u })).toBeEnabled();
  }
  await addDialog.getByRole("button", { name: "取消" }).click();

  const header = page.locator("[data-source-columns]");
  const firstRow = page.locator("[data-source-row]").first();
  await expect(firstRow).toBeVisible();
  const desktop = (page.viewportSize()?.width || 0) >= 1280;
  if (desktop) {
    await expect(header).toBeVisible();
  } else {
    await expect(header).toBeHidden();
    await expect(firstRow.getByText("参与搜索", { exact: true })).toBeVisible();
    await expect(firstRow.getByText("首页展示", { exact: true })).toBeVisible();
    const dimensions = await page.evaluate(() => ({ page: document.documentElement.scrollWidth, viewport: window.innerWidth }));
    expect(dimensions.page).toBe(dimensions.viewport);
  }

  await page.getByRole("button", { name: "来源工具" }).click();
  await expect(page.getByRole("menuitem", { name: "启用可用来源" })).toBeVisible();
  await expect(page.getByRole("menuitem", { name: "停用不可用来源" })).toBeVisible();
  const probeAllButton = page.getByRole("menuitem", { name: "检测全部" });
  const probed = page.waitForResponse((response) => response.request().method() === "POST" && response.url().endsWith("/api/v1/admin/sources/probe-all"));
  await probeAllButton.click();
  expect((await probed).ok()).toBe(true);
  await expect(page.getByRole("status").filter({ hasText: /个来源可用/u })).toBeVisible();

  if (desktop) {
    const priority = firstRow.getByLabel(/优先级/u);
    await expect.poll(() => priority.evaluate((element: HTMLInputElement & { __vueParentComponent?: unknown }) => Boolean(element.__vueParentComponent))).toBe(true);
    const original = Number(await priority.inputValue());
    const saved = page.waitForResponse((response) => response.request().method() === "PATCH" && response.url().includes("/api/v1/admin/sources/"));
    await priority.fill(String(original + 1));
    await priority.press("Tab");
    expect((await saved).ok()).toBe(true);
    await expect(firstRow.getByText("已自动保存", { exact: true })).toBeVisible();
  }
});
