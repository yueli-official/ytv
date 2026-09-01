import { expect, test } from "@playwright/test";

test.skip(process.env.YTV_E2E_LIVE_KAZUMI !== "true", "set YTV_E2E_LIVE_KAZUMI=true for the live source smoke");

test("7sefun resolves into the native YTV player", async ({ page }) => {
  await page.goto("/watch/kazumi-7sefun/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RkZXRhaWwvMjkxOTMuaHRtbA/aHR0cHM6Ly93d3cuN3NlZnVuLnRvcC92b2RwbGF5LzI5MTkzLTEtMS5odG1s", { waitUntil: "domcontentloaded" });
  const player = page.locator("[data-player-state]");
  await expect(player.locator("iframe")).toHaveCount(0);
  const video = player.locator("video");
  await expect(video).toBeVisible();
  await expect.poll(() => video.evaluate((element: HTMLVideoElement) => element.currentSrc)).toMatch(/^https:\/\/[^/]+\.svp\.tencent-cloud\.com\/.*\.mp4\?/u);
});

const nativeCases = [
  ["kazumi-aafun", "aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjguaHRtbA", /^(?:dl\.playxf\.top|play\.xfvod\.pro)$/u],
  ["kazumi-akianime", "aHR0cHM6Ly93d3cuYWtpYW5pbWUuY2MvYmdtcGxheS9kUEREREUtMS0xLmh0bWw", /\.cmecloud\.cn$/u],
  ["kazumi-moonci", "aHR0cHM6Ly93d3cubW9vbmNpLmNvbS9hbmltZS8yMjguaHRtbA", /^(?:dl\.playxf\.top|play\.xfvod\.pro)$/u],
  ["kazumi-mxdm", "aHR0cHM6Ly93d3cuZGNjMy5jb20vZGV0YWlsLzI3NDIv", /\.play-cdn(?:15|19)\.com$/u],
  ["kazumi-xfdmneo", "aHR0cHM6Ly9kbTEueGZkbS5wcm8vYmFuZ3VtaS80NC5odG1s", /^play\.xfvod\.pro$/u]
] as const;

test("reviewed Kazumi adapters return native media manifests", async ({ request }) => {
  for (const [sourceKey, externalId, expectedHost] of nativeCases) {
    const titleResponse = await request.get(`/api/v1/titles/${sourceKey}/${externalId}`);
    expect(titleResponse.ok(), `${sourceKey} title`).toBe(true);
    const title = await titleResponse.json() as { episodes: Array<{ key: string }> };
    expect(title.episodes.length, `${sourceKey} episodes`).toBeGreaterThan(0);
    const playbackResponse = await request.get(`/api/v1/playback/${sourceKey}/${externalId}/${title.episodes[0]!.key}`, { timeout: 45_000 });
    expect(playbackResponse.ok(), `${sourceKey} playback`).toBe(true);
    const manifest = await playbackResponse.json() as { engine: string; url: string };
    expect(manifest.engine).toMatch(/^(?:hls|direct)$/u);
    expect(new URL(manifest.url).hostname).toMatch(expectedHost);
  }
});
