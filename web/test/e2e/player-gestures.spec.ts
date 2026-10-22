import { expect, test, type Page } from '@playwright/test';

async function openFilm(page: Page) {
 await page.addInitScript(() => localStorage.setItem('plyr', JSON.stringify({speed:1.5,volume:0.4,muted:true})));
 await page.route('**/api/v1/playback/fixture/frost-line/0', async route => {
  const response = await route.fetch();
  await route.fulfill({response,json:{...await response.json(),engine:'direct',mediaType:'video/mp4',url:'/demo/feature.mp4'}});
 });
 await page.goto('/watch/fixture/frost-line/0',{waitUntil:'networkidle'});
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','ready');
 await page.locator('[data-primary-play]').click();
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','playing');
 return page.locator('video');
}

test('holding the picture uses temporary 2x and release restores the saved speed', async ({page},info)=>{
 test.setTimeout(60000);
 const video = await openFilm(page);
 const box = (await page.locator('video-neutral-skin').boundingBox())!;
 const x=box.x+box.width/2,y=box.y+box.height*0.4;
 const touch=info.project.name==='mobile'?await page.context().newCDPSession(page):null;
 if(touch) await touch.send('Input.dispatchTouchEvent',{type:'touchStart',touchPoints:[{x,y}]});
 else {await page.mouse.move(x,y);await page.mouse.down();}
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(2);
 await expect(page.locator('[data-hold-speed]')).toBeVisible();
 expect(await page.evaluate(()=>JSON.parse(localStorage.getItem('plyr')!).speed)).toBe(1.5);
 await page.screenshot({path:info.outputPath('hold-speed.png'),fullPage:true});
 if(touch) await touch.send('Input.dispatchTouchEvent',{type:'touchEnd',touchPoints:[]}); else await page.mouse.up();
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(1.5);
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','playing');
 await expect(page.locator('[data-hold-speed]')).toHaveCount(0);
 if(touch){
  await touch.send('Input.dispatchTouchEvent',{type:'touchStart',touchPoints:[{x,y}]});
  await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(2);
  await touch.send('Input.dispatchTouchEvent',{type:'touchCancel',touchPoints:[]});
 } else {
  await page.mouse.down();
  await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(2);
  await page.mouse.move(box.x-5,y);await page.mouse.up();
 }
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(1.5);
 expect(await page.evaluate(()=>JSON.parse(localStorage.getItem('plyr')!).speed)).toBe(1.5);
});

test('keyboard and double-tap skip ten seconds without a second seek handler',async({page},info)=>{
 const video=await openFilm(page);
 await video.evaluate((v:HTMLVideoElement)=>{v.pause();v.currentTime=20;});
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.currentTime)).toBe(20);
 const container=page.locator('media-container');
 await container.focus();
 await page.keyboard.press('ArrowRight');
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.currentTime)).toBe(30);
 await page.keyboard.press('ArrowLeft');
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.currentTime)).toBe(20);
 const box=(await page.locator('video-neutral-skin').boundingBox())!;
 const y=box.y+box.height*0.4;
 for(const [fraction,expected] of [[0.15,10],[0.85,20]]){
  const x=box.x+box.width*fraction!;
  if(info.project.name==='mobile'){await page.touchscreen.tap(x,y);await page.touchscreen.tap(x,y);}
  else await page.mouse.dblclick(x,y);
  await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.currentTime)).toBe(expected);
 }
 await expect(page.locator('[data-hold-speed]')).toHaveCount(0);
});

test('fullscreen hold feedback stays visible and holding a control never accelerates',async({page},info)=>{
 test.skip(info.project.name==='mobile','Native phone fullscreen needs device verification.');
 const video=await openFilm(page);
 const settings=page.locator('video-neutral-skin').getByRole('button',{name:'设置',exact:true});
 const settingsBox=(await settings.boundingBox())!;
 await page.mouse.move(settingsBox.x+settingsBox.width/2,settingsBox.y+settingsBox.height/2);
 await page.mouse.down();
 // Exceed the gesture threshold while deliberately holding an interactive control.
 await page.waitForTimeout(550);
 expect(await video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(1.5);
 await page.mouse.up();await page.keyboard.press('Escape');
 await page.locator('media-container').focus();await page.keyboard.press('f');
 await expect.poll(()=>page.evaluate(()=>Boolean(document.fullscreenElement))).toBe(true);
 const box=(await page.locator('media-container').boundingBox())!;
 await page.mouse.move(box.x+box.width/2,box.y+box.height*0.4);await page.mouse.down();
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(2);
 await expect(page.locator('[data-hold-speed]')).toBeVisible();
 await page.screenshot({path:info.outputPath('fullscreen-hold.png')});
 await page.mouse.up();
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(1.5);
 await page.keyboard.press('f');
 await expect.poll(()=>page.evaluate(()=>Boolean(document.fullscreenElement))).toBe(false);
});

test('HLS pause during a hold restores speed and clears the acceleration state',async({page})=>{
 await page.goto('/watch/fixture/frost-line/0',{waitUntil:'networkidle'});
 const video=page.locator('video');
 await expect(page.locator('[data-player-state]')).toHaveAttribute('data-player-state','ready');
 await video.evaluate((v:HTMLVideoElement)=>{v.muted=true;v.playbackRate=1.25;});
 await page.locator('[data-primary-play]').click();
 const box=(await page.locator('video-neutral-skin').boundingBox())!;
 await page.mouse.move(box.x+box.width/2,box.y+box.height*0.4);await page.mouse.down();
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(2);
 await video.evaluate((v:HTMLVideoElement)=>v.pause());
 await expect.poll(()=>video.evaluate((v:HTMLVideoElement)=>v.playbackRate)).toBe(1.25);
 await expect(page.locator('[data-hold-speed]')).toHaveCount(0);
 await page.mouse.up();
 expect(await page.evaluate(()=>JSON.parse(localStorage.getItem('plyr')!).speed)).toBe(1.25);
});
