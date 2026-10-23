import { expect, test } from "@playwright/test";

test("blocked sources remain distinct from empty results and preserve successful candidates", async ({page}, info) => {
 let partial = false;
 await page.route("**/api/v1/search?**", async route => {
  const response = await route.fetch(); const body = await response.json();
  if (!partial) body.groups = [];
  body.complete = false; body.attempted = 2; body.succeeded = partial ? 1 : 0; body.failed = partial ? 1 : 2;
  body.sources = [{key:"blocked",name:"图片验证来源",outcome:"verification_required",verificationKind:"image",itemCount:0},partial ? {key:"fixture",name:"本地演示源",outcome:"success",itemCount:1} : {key:"down",name:"故障来源",outcome:"unavailable",itemCount:0}];
  await route.fulfill({response,json:body});
 });
 await page.goto('/search?sources=fixture',{waitUntil:'networkidle'});
 await page.getByLabel('搜索想看的影视').fill('霜线');
 await page.getByLabel('搜索想看的影视').press('Enter');
 await expect(page.getByRole('heading',{name:'部分来源需要验证',exact:true})).toBeVisible();
 await expect(page.getByText('需要验证码',{exact:true})).toBeVisible();
 await expect(page.getByRole('heading',{name:'没有找到相关影片',exact:true})).toHaveCount(0);
 await page.screenshot({path:info.outputPath('verification-search.png'),fullPage:true});
 partial=true;
 await page.getByRole('button',{name:'重新搜索',exact:true}).click();
 await expect(page.getByRole('heading',{name:'霜线',exact:true})).toBeVisible();
 await page.locator('summary').filter({hasText:'查看来源状态'}).click();
 await expect(page.getByText('需要验证码',{exact:true})).toBeVisible();
});

test("verification cause reaches detail and playback with a way back", async ({page},info)=>{
 const failure={type:'https://errors.yueli.dev/problems/ytv.source_verification_required',status:409,code:'ytv.source_verification_required',params:{},violations:[],traceId:'verification-test'};
 await page.route('**/api/v1/titles/fixture/frost-line',route=>route.fulfill({status:409,contentType:'application/problem+json',json:failure}));
 await page.goto('/search?sources=fixture',{waitUntil:'networkidle'});
 await page.getByLabel('搜索想看的影视').fill('霜线');
 await page.getByLabel('搜索想看的影视').press('Enter');
 await page.getByRole('link',{name:'查看 霜线',exact:true}).click();
 await expect(page.getByText('此来源要求完成验证，本站暂不能代为完成，请选择其他来源。',{exact:true})).toBeVisible();
 await expect(page.getByRole('link',{name:'返回搜索',exact:true})).toHaveAttribute('href',/\/search\?.*q=/);
 await page.screenshot({path:info.outputPath('verification-detail.png'),fullPage:true});
 await page.unroute('**/api/v1/titles/fixture/frost-line');
 await page.route('**/api/v1/playback/**',route=>route.fulfill({status:409,contentType:'application/problem+json',json:failure}));
 await page.goto('/watch/fixture/frost-line/0',{waitUntil:'networkidle'});
 await expect(page.getByText('此来源要求完成验证，本站暂不能代为完成，请选择其他来源。',{exact:true})).toBeVisible();
 await expect(page.getByRole('link',{name:'换个来源',exact:true})).toBeVisible();
 await page.screenshot({path:info.outputPath('verification-player.png'),fullPage:true});
});
