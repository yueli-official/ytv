import {expect,test} from '@playwright/test';

test('mobile journey keeps source selection, history and a single player across layouts',async({page},info)=>{
 const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message));
 await page.addInitScript(()=>localStorage.setItem('ytv-progress-v2:fixture:ring-sea:0',JSON.stringify({titleName:'环海之门',episodeTitle:'第 1 集',position:12,duration:120,completed:false,updatedAt:Date.now()})));
 await page.goto('/',{waitUntil:'networkidle'});
 await expect(page.locator('[data-mobile-page=home]')).toBeVisible();
 await page.screenshot({path:info.outputPath('home.png'),fullPage:true});
 await page.getByRole('navigation',{name:'移动端主导航'}).getByRole('link',{name:'搜索',exact:true}).click();
 await page.getByLabel('搜索想看的影视').fill('霜线');await page.getByLabel('搜索想看的影视').press('Enter');
 await expect(page.getByRole('heading',{name:'霜线',exact:true})).toBeVisible();
 await page.screenshot({path:info.outputPath('search.png'),fullPage:true});
 await page.getByText('选择来源 · 1 个',{exact:true}).click();
 await page.getByRole('link',{name:'本地演示源',exact:true}).click();
 await expect(page.locator('[data-mobile-page=title]')).toBeVisible();
 await page.screenshot({path:info.outputPath('title.png'),fullPage:true});
 await page.getByRole('link',{name:'播放正片',exact:true}).click();
 await expect(page.locator('[data-player-state]')).toHaveCount(1);
 const player=page.locator('[data-player-state]');await player.evaluate(el=>el.setAttribute('data-same-player','yes'));
 await page.getByRole('button',{name:'选集',exact:true}).click();
 await expect(page.getByRole('dialog')).toBeVisible();
 await page.screenshot({path:info.outputPath('watch-sheet.png'),fullPage:true});
 await page.keyboard.press('Escape');
 await expect(page.getByRole('dialog')).toHaveCount(0);
 await page.setViewportSize({width:1440,height:1000});
 await expect(page.locator('[data-same-player=yes]')).toHaveCount(1);
 await page.setViewportSize({width:390,height:844});
 await expect(page.locator('[data-same-player=yes]')).toHaveCount(1);
 await page.screenshot({path:info.outputPath('watch.png'),fullPage:true});
 await page.goto('/history',{waitUntil:'networkidle'});
 await expect(page.locator('[data-mobile-page=history]')).toBeVisible();
 await page.getByRole('button',{name:'筛选',exact:true}).click();
 await page.getByLabel('筛选已加载的记录').fill('环海');
 await expect(page.getByRole('heading',{name:'环海之门',exact:true})).toBeVisible();
 await page.screenshot({path:info.outputPath('history.png'),fullPage:true});
 expect(errors).toEqual([]);
});

test('search state survives width changes without issuing another search',async({page},info)=>{
 let requests=0;
 page.on('request',r=>{if(r.url().includes('/api/v1/search?'))requests++});
 await page.goto('/search?sources=fixture',{waitUntil:'networkidle'});
 await page.getByLabel('搜索想看的影视').fill('环海');await page.getByLabel('搜索想看的影视').press('Enter');
 await expect(page.getByRole('heading',{name:'环海之门',exact:true})).toBeVisible();
 const before=requests;
 for(const width of [768,767,844,390]){await page.setViewportSize({width,height:width===844?390:844});await expect(page.getByLabel('搜索想看的影视')).toHaveValue('环海')}
 expect(requests).toBe(before);
 await page.screenshot({path:info.outputPath('search-resized.png'),fullPage:true});
});

test('mobile long-series selection closes the sheet and settings stay reachable',async({page},info)=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('/watch/fixture/long-voyage/61',{waitUntil:'networkidle'});
 await page.getByRole('button',{name:'选集',exact:true}).click();
 const sheet=page.getByRole('dialog');
 await expect(sheet.getByLabel('剧集范围')).toBeVisible();
 await expect(sheet.locator('[data-player-episode-row]')).toHaveCount(50);
 await page.screenshot({path:info.outputPath('long-series.png'),fullPage:true});
 const target=sheet.locator('[data-player-episode-row]').first();
 const href=await target.getAttribute('href');await target.click();
 await expect(sheet).toHaveCount(0);
 expect(new URL(page.url()).pathname).toBe(new URL(href!,page.url()).pathname);
 await expect(page.locator('[data-player-state]')).toHaveCount(1);
 await page.goto('/search',{waitUntil:'networkidle'});
 await page.getByRole('button',{name:'来源设置',exact:true}).click();
 await expect(page.getByRole('dialog',{name:'来源设置'})).toBeVisible();
 await page.screenshot({path:info.outputPath('source-settings.png'),fullPage:true});
 await page.keyboard.press('Escape');
 await expect(page.getByRole('dialog')).toHaveCount(0);
 expect(errors).toEqual([]);
});
