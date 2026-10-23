import { test, expect, type Page } from '@playwright/test';

async function login(page: Page) {
  await page.goto('/manage');
  await page.waitForFunction(() => Boolean((document.querySelector('form input[name=email]')?.closest('form') as unknown as { __vueParentComponent?: unknown })?.__vueParentComponent));
  await page.getByLabel('邮箱').fill(process.env.YTV_E2E_EMAIL || 'test@example.com');
  await page.getByLabel('密码', { exact: true }).fill(process.env.YTV_E2E_PASSWORD || 'Yueli-local-development-2026');
  await page.getByRole('button', { name: '登录', exact: true }).click();
  await page.waitForURL(/\/manage(?:\?|$)/u);
}

test('source recommendation persists and reaches viewer choices', async ({ page }, info) => {
  test.setTimeout(120_000);
  await login(page);
  await page.goto('/manage/sources');
  const original = (await (await page.request.get('/api/v1/admin/sources')).json()).items[0];
  const recommendControl = page.getByRole('switch', { name: `推荐 ${original.name}`, exact: true });
  await expect(recommendControl).toBeVisible({ timeout: 30_000 });
  try {
    await page.getByRole('switch', { name: `推荐 ${original.name}`, exact: true }).click();
    await expect.poll(async () => (await (await page.request.get('/api/v1/admin/sources')).json()).items.find((s: { key: string }) => s.key === original.key).recommended).toBe(!original.recommended);
    for (let i=0;i<3;i++) {
      const response = await page.reload();
      expect(await response!.text()).toContain('data-admin-shell');
      await expect(recommendControl).toBeVisible();
    }
    await page.screenshot({ path: info.outputPath('source-admin.png'), fullPage: true });
    await page.goto('/search');
    const opener = page.getByRole('button', { name: '来源设置', exact: true }).first();
    await expect.poll(() => opener.evaluate(el => Boolean((el as unknown as { __vueParentComponent?: { isMounted?: boolean } }).__vueParentComponent?.isMounted))).toBe(true);
    await opener.click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText(original.name, { exact: true })).toBeVisible();
    await dialog.getByRole('button', { name: '推荐', exact: true }).click();
    await expect(dialog.getByText(original.name, { exact: true })).toHaveCount(original.recommended ? 0 : 1);
    await page.screenshot({ path: info.outputPath('source-viewer.png') });
    await dialog.getByRole('button', { name: '全部', exact: true }).click();
  } finally {
    const latest = (await (await page.request.get('/api/v1/admin/sources')).json()).items.find((s: { key: string }) => s.key === original.key);
    const status = await page.evaluate(async ({ original, revision }) => {
      const result = await fetch('/api/v1/admin/sources/'+encodeURIComponent(original.key), { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...original, revision }) });
      return result.status;
    }, { original, revision: latest.revision });
    expect(status).toBe(200);
  }
});
