import { expect, test } from '@playwright/test';

test('film lines remain separate and selection disposes the previous player', async ({ page }) => {
 await page.goto('/title/fixture/frost-line', {waitUntil:'networkidle'});
 await expect(page.getByRole('link',{name:'播放正片',exact:true})).toBeVisible();
 await page.getByRole('button',{name:'备用线路',exact:true}).click();
 await page.getByRole('link',{name:'播放正片',exact:true}).click();
 await expect(page).toHaveURL(/\/line1-0$/);
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','ready');
 await page.locator('[data-primary-play]').click();
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','playing');
 await page.getByRole('combobox',{name:'播放线路'}).click();
 await page.getByRole('option',{name:'主线路',exact:true}).click();
 await expect(page).toHaveURL(/\/frost-line\/0$/);
 await expect(page.locator('video')).toHaveCount(1);
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','ready');
});

test('reconnect requests a fresh manifest and recovers from an unreachable address', async ({page})=>{
 let attempts=0;
 await page.route('**/api/v1/playback/fixture/frost-line/0',async route=>{
  attempts++;
  const response=await route.fetch();const manifest=await response.json();
  await route.fulfill({response,json:{...manifest,url:attempts===1?'/demo/missing.m3u8':manifest.url}});
 });
 await page.goto('/watch/fixture/frost-line/0',{waitUntil:'domcontentloaded'});
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','error',{timeout:30000});
 await page.getByRole('button',{name:'重新连接',exact:true}).click();
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','ready');
 expect(attempts).toBe(2);
 await page.locator('[data-primary-play]').click();
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','playing');
});

test('native MP4 uses the same controls and retains existing volume and speed preferences',async({page})=>{
 await page.addInitScript(()=>localStorage.setItem('plyr',JSON.stringify({volume:0.4,muted:true,speed:1.5})));
 await page.route('**/api/v1/playback/fixture/frost-line/0',async route=>{
  const response=await route.fetch();
  await route.fulfill({response,json:{...await response.json(),engine:'direct',mediaType:'video/mp4',url:'/demo/feature.mp4'}});
 });
 await page.goto('/watch/fixture/frost-line/0',{waitUntil:'networkidle'});
 const player=page.locator('[data-player-state]');
 await expect(player).toHaveAttribute('data-player-state','ready');
 expect(await player.locator('video').evaluate((media:HTMLVideoElement)=>[media.volume,media.muted,media.playbackRate])).toEqual([0.4,true,1.5]);
 await page.locator('[data-primary-play]').click();
 await expect(player).toHaveAttribute('data-player-state','playing');
 await expect.poll(()=>player.locator('video').evaluate((media:HTMLVideoElement)=>media.currentTime)).toBeGreaterThan(0);
});
