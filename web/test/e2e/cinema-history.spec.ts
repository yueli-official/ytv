import { expect, test } from "@playwright/test";

test("source status distinguishes empty results from timeout without losing usable results", async ({page}, info) => {
 await page.route('**/api/v1/search?**', async route=>{
  const response=await route.fetch();const body=await response.json();
  await route.fulfill({response,json:{...body,complete:false,attempted:3,succeeded:2,failed:1,sources:[
   {key:'fixture',name:'演示来源',outcome:'success',itemCount:1,durationMs:20},
   {key:'empty',name:'无匹配来源',outcome:'success',itemCount:0,durationMs:20},
   {key:'timeout',name:'超时来源',outcome:'timeout',itemCount:0,durationMs:6000}
  ]}});
 });
 await page.goto('/search?sources=fixture',{waitUntil:'networkidle'});
 await page.getByLabel('搜索想看的影视').fill('霜线');await page.getByLabel('搜索想看的影视').press('Enter');
 await page.getByText('查看来源状态 · 2/3 正常响应',{exact:true}).click();
 const states=page.getByRole('list',{name:'本次搜索来源状态'});
 await expect(states.getByText('没有匹配结果',{exact:true})).toBeVisible();
 await expect(states.getByText('请求超时',{exact:true})).toBeVisible();
 await expect(page.getByRole('heading',{name:'霜线',exact:true})).toBeVisible();
 await page.screenshot({path:info.outputPath('source-status.png'),fullPage:true});
});

test("history groups recent records, filters loaded titles and preserves source identity", async ({ page }, info) => {
  await page.addInitScript(() => {
    const today = new Date().setHours(12, 0, 0, 0);
    const yesterday = new Date(today); yesterday.setDate(yesterday.getDate() - 1);
    const entries = [
      { source: "fixture", id: "ring-sea", title: "环海之门", time: today },
      { source: "fixture", id: "frost-line", title: "霜线", time: yesterday.getTime() },
      { source: "other", id: "ring-sea", title: "环海之门", time: today - 3 * 86400000 }
    ];
    for (const item of entries) localStorage.setItem(`ytv-progress-v2:${item.source}:${item.id}:0`, JSON.stringify({ titleName: item.title, episodeTitle: "正片", position: 90, duration: 900, completed: false, updatedAt: item.time }));
  });
  await page.goto("/history", { waitUntil: "networkidle" });
  await expect(page.locator("[data-history-records] li")).toHaveCount(3);
  for (const name of ["今天", "昨天", "更早"]) await expect(page.getByRole("heading", { name, exact: true })).toBeVisible();
  await page.getByLabel("筛选已加载的记录").fill("环海");
  await expect(page.locator("[data-history-records] li")).toHaveCount(2);
  await expect(page.locator('[data-history-records] a[href="/watch/other/ring-sea/0"]').first()).toBeVisible();
  await page.getByLabel("筛选已加载的记录").fill("没有这个片名");
  await expect(page.getByRole("heading", { name: "没有匹配的记录" })).toBeVisible();
  await page.getByRole("button", { name: "清除筛选" }).click();
  await expect(page.locator("[data-history-records] li")).toHaveCount(3);
  await page.screenshot({ path: info.outputPath("history-grouped.png"), fullPage: true });
});

test("partial search retains available source candidates and can retry", async ({ page }) => {
  let partial = true;
  await page.route("**/api/v1/search?**", async route => {
    const response = await route.fetch();
    const body = await response.json();
    await route.fulfill({ response, json: { ...body, complete: !partial } });
  });
  await page.goto("/search?sources=fixture", { waitUntil: "networkidle" });
  await page.getByLabel("搜索想看的影视").fill("霜线");
  await page.getByLabel("搜索想看的影视").press("Enter");
  await expect(page.getByText("部分来源暂时不可用", { exact: false })).toBeVisible();
  await expect(page.locator('article a[href^="/title/fixture/frost-line"]').first()).toBeVisible();
  partial = false;
  await page.getByRole("button", { name: "重新搜索", exact: true }).click();
  await expect(page.getByText("部分来源暂时不可用", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "重新搜索", exact: true })).toBeEnabled();
  await expect(page.getByRole("heading", { name: "霜线", exact: true })).toBeVisible();
});
