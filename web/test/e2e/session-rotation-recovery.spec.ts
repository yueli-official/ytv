import { expect, test, type Page } from "@playwright/test";
import { seal, unseal, type Session } from "../../../../identity/nuxt/server/utils/oidc.ts";

const enabled = process.env.YTV_E2E_SESSION_ROTATION === "true";
const phase = process.env.YTV_E2E_SESSION_ROTATION_PHASE || "seed";
const statePath = process.env.YTV_E2E_SESSION_STATE || "test/e2e/.session-rotation-state.json";
const email = process.env.YTV_E2E_EMAIL || "test@example.com";
const password = process.env.YTV_E2E_PASSWORD || "Yueli-local-development-2026";
const sealSecret = process.env.YTV_E2E_SEAL_SECRET || "local-service-secret-2026-change-me-ytv-web-seal";

if (phase === "verify") test.use({ storageState: statePath });

async function login(page: Page) {
  await page.goto("/manage");
  await page.waitForURL(/\/login(?:\?|$)/u);
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

test("product session survives a committed refresh whose response was lost across Consumer restart", async ({ page, context }) => {
  test.skip(!enabled, "set YTV_E2E_SESSION_ROTATION=true for the live fault-injection regression");
  if (phase === "seed") {
    await login(page);
    const cookies = await context.cookies();
    const product = cookies.find((cookie) => /^ys_ytv-main_[0-9a-f]{12}$/u.test(cookie.name));
    expect(product).toBeTruthy();
    const session = unseal<Session>(product!.value, sealSecret);
    expect(session?.refresh).toBeTruthy();
    const response = await fetch("http://127.0.0.1:8081/oauth2/token", {
      method: "POST", headers: { "content-type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({ grant_type: "refresh_token", refresh_token: session!.refresh!, client_id: "ytv-main-web" })
    });
    expect(response.ok).toBe(true);
    await context.addCookies([{ ...product!, value: seal({ ...session!, exp: 0 }, sealSecret) }]);
    await context.storageState({ path: statePath });
    return;
  }
  await page.goto("/", { waitUntil: "networkidle" });
  await expect(page.getByRole("button", { name: /打开测试管理员的用户菜单/u })).toBeVisible();
  const response = await page.request.get("/auth/session");
  expect(response.ok()).toBe(true);
  expect((await response.json()).user?.userKey).toBe("TestA123");
});
