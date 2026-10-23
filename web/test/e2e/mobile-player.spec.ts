import { expect, test, type Page } from '@playwright/test';

const watchPath = process.env.YTV_PLAYER_TEST_PATH || '/watch/fixture/frost-line/0';
async function openPlayer(page: Page, rejectLock = false) {
  await page.addInitScript((reject) => {
    const calls: string[] = [];
    Object.assign(window, { orientationCalls: calls });
    Object.defineProperty(screen.orientation, 'lock', { configurable: true, value: async (value: string) => {
      calls.push(value);
      if (reject) throw new DOMException('Not supported', 'NotSupportedError');
    } });
    Object.defineProperty(screen.orientation, 'unlock', { configurable: true, value: () => calls.push('unlock') });
  }, rejectLock);
  await page.route('**/api/v1/playback/**', route => route.fulfill({ json: {
    sourceName: 'Playback fixture', episodeKey: '0', url: '/demo/feature.mp4',
    engine: 'direct', timeline: 'mobile-player-test', mediaType: 'video/mp4'
  } }));
  await page.goto(watchPath, { waitUntil: 'domcontentloaded' });
  await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state', 'ready', { timeout: 30_000 });
}
const calls = (page: Page) => page.evaluate(() => (window as unknown as { orientationCalls: string[] }).orientationCalls);

test('fullscreen requests landscape and releases the lock on exit', async ({ page }, info) => {
  await openPlayer(page);
  const mobile = info.project.name === 'mobile';
  if (mobile) await page.getByRole('button', { name: '全屏观看', exact: true }).click();
  else await page.locator('video-neutral-skin').getByRole('button', { name: '全屏', exact: true }).click();
  await expect.poll(() => calls(page)).toContain('landscape');
  await expect.poll(() => page.evaluate(() => Boolean(document.fullscreenElement))).toBe(true);
  await page.evaluate(() => document.exitFullscreen());
  await expect.poll(() => calls(page)).toContain('unlock');
  await page.locator('[data-primary-play]').click();
  await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state', 'playing');
});

test('unsupported orientation lock does not prevent fullscreen or playback', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', error => errors.push(error.message));
  await openPlayer(page, true);
  await page.locator('video-neutral-skin').getByRole('button', { name: '全屏', exact: true }).click();
  await expect.poll(() => page.evaluate(() => Boolean(document.fullscreenElement))).toBe(true);
  await page.locator('video-neutral-skin').getByRole('button', { name: '播放', exact: true }).click();
  await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state', 'playing');
  await page.evaluate(() => document.exitFullscreen());
  expect(errors).toEqual([]);
});

test('phone rotation retains the mobile layout and one running media element', async ({ page }, info) => {
  test.skip(info.project.name !== 'mobile', 'Touch phone layout');
  await openPlayer(page);
  await page.locator('[data-primary-play]').click();
  await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state', 'playing');
  const media = await page.locator('video').elementHandle();
  await page.setViewportSize({ width: 844, height: 390 });
  await expect(page.getByRole('button', { name: '全屏观看', exact: true })).toBeVisible();
  await expect(page.locator('.ytv-topbar')).toHaveCount(0);
  expect(await media!.evaluate(element => element.isConnected)).toBe(true);
  await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state', 'playing');
  await page.screenshot({ path: info.outputPath('phone-landscape.png') });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: info.outputPath('phone-portrait.png') });
});
