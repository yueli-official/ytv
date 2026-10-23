import { test, expect } from '@playwright/test';

for (const identity of ['guest', 'account']) test(`${identity} source reviews count viewers, replace votes and withdraw`, async ({ page, browser }, info) => {
  test.setTimeout(150_000);
  if (identity === 'account') {
  await page.goto('/manage');
  await page.waitForFunction(() => Boolean((document.querySelector('form input[name=email]')?.closest('form') as unknown as { __vueParentComponent?: unknown })?.__vueParentComponent));
  await page.getByLabel('邮箱').fill(process.env.YTV_E2E_EMAIL || 'test@example.com');
  await page.getByLabel('密码', { exact: true }).fill(process.env.YTV_E2E_PASSWORD || 'Yueli-local-development-2026');
  await page.getByRole('button', { name: '登录', exact: true }).click();
  await page.waitForURL(/\/manage(?:\?|$)/u);
  } else { await page.goto('/'); }
  const sources = (await (await page.request.get('/api/v1/sources')).json()).items;
  const source = sources.find((s: { key: string }) => s.key === 'zuid') || sources[0];
  const endpoint = '/api/v1/me/source-reviews/' + source.key;
  const original = (await (await page.request.get(endpoint)).json()).review;
  const put = (ads: string, quality: string) => page.evaluate(async ({ endpoint, ads, quality }) => {
    const r = await fetch(endpoint, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ads, quality }) });
    return { status: r.status, body: await r.json() };
  }, { endpoint, ads, quality });
  const summary = async () => (await (await page.request.get('/api/v1/sources')).json()).items.find((s: { key: string }) => s.key === source.key).reviews;
  const guest = await browser.newContext();
  try {
    const other = await guest.request.get(new URL(endpoint, page.url()).href);
    expect(other.status()).toBe(200);
    expect((await other.json()).review).toMatchObject({ ads: '', quality: '' });
    expect((await put('', '')).status).toBe(200);
    const baseline = await summary();
    await page.route('**/api/v1/search?*', async route => route.fulfill({ json: {
      query: '来源评价验收', page: 1, attempted: 1, succeeded: 1, failed: 0, complete: true,
      groups: [{ key: 'review-test', title: '来源评价验收', year: '2026', category: '测试', candidates: [{ sourceKey: source.key, sourceName: source.name, externalId: 'review-test', name: '来源评价验收', remarks: '界面测试条目', episodeCount: 1 }] }],
      sources: [{ key: source.key, name: source.name, outcome: 'success', durationMs: 10, itemCount: 1 }]
    } }));
    await page.goto('/search');
    const input = page.getByRole('textbox', { name: '搜索想看的影视', exact: true });
    await expect.poll(() => input.evaluate(el => Boolean((el as unknown as { __vueParentComponent?: { isMounted?: boolean } }).__vueParentComponent?.isMounted))).toBe(true);
    await input.fill('来源评价验收');
    await input.press('Enter');
    const card = page.locator('.ytv-result-card').first();
    await expect(card).toBeVisible();
    const details = card.locator('details');
    if (!(await details.evaluate(el => el.open))) await details.locator('summary').click();
    await expect(card.getByRole('button', { name: `评价 ${source.name}`, exact: true })).toHaveCount(0);
    const watchPath = `/watch/${source.key}/review-test/0`;
    await page.route(`**/api/v1/titles/${source.key}/review-test`, route => route.fulfill({ json: {
      title: { sourceKey: source.key, sourceName: source.name, externalId: 'review-test', name: '来源评价验收', year: '2026', category: '测试', episodeCount: 1 },
      episodes: [{ key: '0', title: '正片', lineKey: '0', lineName: '测试线路' }]
    } }));
    const openWatch = async () => {
      await page.goto(watchPath);
      const retry = page.getByRole('button', { name: '重试', exact: true });
      if (await retry.isVisible()) {
        await expect.poll(() => retry.evaluate(el => Boolean((el as unknown as { __vueParentComponent?: { isMounted?: boolean } }).__vueParentComponent?.isMounted))).toBe(true);
        await retry.click();
      }
      await expect(page.getByRole('region', { name: '当前来源评价' })).toBeVisible();
    };
    await openWatch();
    await page.getByRole('button', { name: `评价 ${source.name}`, exact: true }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByRole('radio', { name: '广告少', exact: true }).check();
    await dialog.getByRole('radio', { name: '高清', exact: true }).check();
    await dialog.getByRole('button', { name: '保存评价', exact: true }).click();
    await expect(dialog.getByText('已保存评价', { exact: true })).toBeVisible();
    await expect.poll(summary).toEqual({ ...baseline, total: baseline.total + 1, adsFew: baseline.adsFew + 1, hd: baseline.hd + 1 });
    await page.screenshot({ path: info.outputPath('review-editor.png'), animations: 'disabled' });
    await page.keyboard.press('Escape');
    await expect(dialog).not.toBeVisible();
    await page.screenshot({ path: info.outputPath('review-watch.png'), animations: 'disabled' });
    await page.goto('/search');
    await expect.poll(() => input.evaluate(el => Boolean((el as unknown as { __vueParentComponent?: { isMounted?: boolean } }).__vueParentComponent?.isMounted))).toBe(true);
    await input.fill('来源评价验收');
    await input.press('Enter');
    await expect(card).toBeVisible();
    if (!(await details.evaluate(el => el.open))) await details.locator('summary').click();
    await expect(card.getByRole('button', { name: `评价 ${source.name}`, exact: true })).toHaveCount(0);
    await expect(card.getByLabel(`广告少，${baseline.adsFew + 1} 人`, { exact: true })).toBeVisible();
    await expect(card.getByLabel(`高清，${baseline.hd + 1} 人`, { exact: true })).toBeVisible();
    await page.screenshot({ path: info.outputPath('review-badges.png'), animations: 'disabled' });
    for (const theme of ['light', 'dark'] as const) {
      await page.emulateMedia({ colorScheme: theme });
      await page.screenshot({ path: info.outputPath('review-badges-'+theme+'.png'), animations: 'disabled' });
    }
    await Promise.all([put('few', 'hd'), put('few', 'hd'), put('few', 'hd')]);
    expect((await summary()).total).toBe(baseline.total + 1);
    expect((await put('many', 'blurry')).status).toBe(200);
    expect(await summary()).toEqual({ ...baseline, total: baseline.total + 1, adsMany: baseline.adsMany + 1, blurry: baseline.blurry + 1 });
    expect((await put('made-up', 'hd')).status).toBe(400);
    expect((await (await page.request.get(endpoint)).json()).review).toMatchObject({ ads: 'many', quality: 'blurry' });
    await openWatch();
    await page.getByRole('button', { name: `评价 ${source.name}`, exact: true }).click();
    await expect(dialog.getByRole('radio', { name: '广告多', exact: true })).toBeChecked();
    await dialog.getByRole('button', { name: '撤回评价', exact: true }).click();
    await expect(dialog.getByText('已撤回评价', { exact: true })).toBeVisible();
    await expect.poll(summary).toEqual(baseline);
    expect((await page.request.get('/api/v1/me/source-reviews/no-such-source')).status()).toBe(404);
  } finally {
    expect((await put(original.ads, original.quality)).status).toBe(200);
    await guest.close();
  }
});

