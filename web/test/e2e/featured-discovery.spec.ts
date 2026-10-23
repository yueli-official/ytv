import { expect, test } from "@playwright/test";

test("featured selection supports keyboard and keeps the attributed trending rail available", async ({ page }) => {
  await page.goto("/", { waitUntil: "networkidle" });
  const selectors = page.getByRole("group", { name: "切换热门推荐" }).getByRole("button");
  await expect(selectors).toHaveCount(3);
  await selectors.nth(1).focus();
  await page.keyboard.press("Enter");
  await expect(selectors.nth(1)).toHaveAttribute("aria-pressed", "true");
  const title = await selectors.nth(1).locator("strong").innerText();
  await expect(page.locator(".ytv-feature h2")).toHaveText(title);
  await expect(page.locator("[data-douban-grid] h3", { hasText: title })).toHaveCount(1);
  await page.getByRole("link", { name: "查找片源", exact: true }).click();
  await expect(page.locator("main").getByLabel("搜索想看的影视")).toHaveValue(title);
});

test("an unavailable large poster falls back to the original cover", async ({ page }) => {
  await page.route("**/external-poster/douban/*?size=feature", route => route.abort());
  await page.goto("/", { waitUntil: "networkidle" });
  const art = page.locator(".ytv-feature-art");
  await expect(art).toHaveAttribute("src", /\/external-poster\/douban\/p\d+\.(jpg|webp)$/u);
  await expect.poll(() => art.evaluate((image: HTMLImageElement) => image.complete && image.naturalWidth > 0)).toBe(true);
  await expect(page.getByRole("link", { name: "查找片源", exact: true })).toBeVisible();
});
