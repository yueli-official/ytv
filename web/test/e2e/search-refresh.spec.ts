import { expect, test } from "@playwright/test";

test("search can be repeated without changing the query and duplicate source entries stay distinct", async ({ page }) => {
 let calls = 0;
 await page.route("**/api/v1/search?**", async route => {
  const response = await route.fetch();
  const body = await response.json();
  const candidate = body.groups[0]?.candidates[0];
  calls++;
  if (calls === 1) body.groups = [];
  else if (candidate) body.groups[0].candidates = [candidate, {...candidate, externalId: "alternate"}];
  await route.fulfill({ response, json: body });
 });
 await page.goto("/search?sources=fixture", {waitUntil:"networkidle"});
 await page.locator("header").getByLabel("搜索想看的影视").fill("霜线");
 await page.locator("header").getByLabel("搜索想看的影视").press("Enter");
 await expect(page.locator("header").getByRole("link", {name:"搜索",exact:true})).toHaveCount(0);
 await expect(page.getByRole("heading", {name:"没有找到相关影片"})).toBeVisible();
 const before = calls;
 await page.getByRole("button", {name:"重新搜索",exact:true}).click();
 await expect.poll(()=>calls).toBe(before+1);
 await expect(page.getByRole("heading", {name:"霜线",exact:true})).toBeVisible();
 await expect(page.getByRole("link", {name:"本地演示源 · 条目 1",exact:true})).toHaveAttribute("href", /frost-line/);
 await expect(page.getByRole("link", {name:"本地演示源 · 条目 2",exact:true})).toHaveAttribute("href", /alternate/);
 await expect(page).toHaveURL(/q=.*sources=fixture/);
});

test("playback reports the supported-address error instead of hiding the cause", async ({page})=>{
 await page.route("**/api/v1/playback/**",route=>route.fulfill({status:409,contentType:"application/problem+json",json:{type:"https://errors.yueli.dev/problems/ytv.playback_unsafe",status:409,code:"ytv.playback_unsafe",params:{},violations:[],traceId:"playback-test"}}));
 await page.goto("/watch/fixture/frost-line/0", {waitUntil:"networkidle"});
 await expect(page.locator("[data-player-state]")).toHaveAttribute("data-player-state","error");
 await expect(page.getByText("此来源返回的播放地址不受支持，请选择其他来源。",{exact:true})).toBeVisible();
 await expect(page.getByRole("link", {name:"换条线路",exact:true})).toHaveAttribute("href", "/watch/fixture/frost-line/line1-0");
 await page.getByRole("link", {name:"换个来源",exact:true}).click();
 await expect(page).toHaveURL(/\/search\?q=/);
 await expect(page.getByLabel("搜索想看的影视")).toHaveValue("霜线");
});

test("a missing first poster falls back to another matching source", async ({page}, info) => {
 await page.route("**/api/v1/search?**", async route => {
  const response = await route.fetch(); const body = await response.json();
  const first = body.groups[0].candidates[0];
  body.groups[0].candidates = [{...first, imageUrl:'/missing-poster-ux.jpg'}, {...first, sourceKey:'backup', sourceName:'备用来源', externalId:'backup'}];
  await route.fulfill({response,json:body});
 });
 await page.goto('/search?sources=fixture',{waitUntil:'networkidle'});
 await page.getByLabel('搜索想看的影视').fill('霜线');
 await page.getByLabel('搜索想看的影视').press('Enter');
 const poster=page.locator('article img').first();
 await expect(poster).not.toHaveAttribute('src','/missing-poster-ux.jpg');
 await expect.poll(()=>poster.evaluate((image:HTMLImageElement)=>image.naturalWidth)).toBeGreaterThan(0);
 await expect(page.getByRole('link',{name:'备用来源',exact:true})).toBeVisible();
 await page.screenshot({path:info.outputPath('poster-fallback.png'),fullPage:true});
});

test("switching lines retains a matching episode instead of restarting the series", async ({page}) => {
 await page.addInitScript(()=>localStorage.setItem('ytv-progress-v2:fixture:ring-sea:1',JSON.stringify({titleName:'环海之门',episodeTitle:'第 2 集',position:12,duration:120,completed:false,updatedAt:Date.now()})));
 await page.route('**/api/v1/titles/fixture/ring-sea',async route=>{
  const response=await route.fetch();const body=await response.json();
  const original=body.episodes.map((ep: {key:string;title:string})=>({...ep,lineKey:'0',lineName:'主线路'}));
  body.episodes=[...original,...original.map((ep: {key:string;title:string})=>({...ep,key:`backup-${ep.key}`,lineKey:'1',lineName:'备用线路'}))];
  await route.fulfill({response,json:body});
 });
 await page.goto('/history',{waitUntil:'networkidle'});
 await page.getByRole('link',{name:'继续观看',exact:true}).click();
 await page.getByRole('combobox',{name:'播放线路'}).click();
 await page.getByRole('option',{name:'备用线路',exact:true}).click();
 await expect(page).toHaveURL(/\/ring-sea\/backup-1$/);
});
