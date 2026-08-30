import type { FeaturedTitle, MediaTitle } from "~/types/ytv";

export const featuredTitle: FeaturedTitle = {
  id: "ring-sea",
  sourceKey: "fixture",
  title: "环海之门",
  category: "科幻 / 冒险",
  year: 2026,
  image: "/demo/hero-gate.png",
  episodeLabel: "演示内容",
  summary: "用于本地验收的合成条目。正式页面只展示当前获准来源返回的事实。",
  sources: [
    { key: "fixture-a", name: "演示源 A", quality: "1080p", state: "available", selected: true },
    { key: "fixture-b", name: "演示源 B", quality: "4K", state: "available" },
    { key: "fixture-c", name: "演示源 C", quality: "720p", state: "degraded" },
    { key: "fixture-d", name: "演示源 D", quality: "1080p", state: "unavailable" }
  ]
};

export const continueWatching: MediaTitle[] = [
  { id: "orbit-beyond", sourceKey: "fixture", title: "轨道之外", category: "科幻", year: 2026, image: "/demo/card-orbit.png", episodeLabel: "第 3 集", progress: 58, remaining: "剩余 18 分钟" },
  { id: "frost-line", sourceKey: "fixture", title: "霜线", category: "悬疑", year: 2025, image: "/demo/card-frost.png", episodeLabel: "第 2 集", progress: 32, remaining: "剩余 31 分钟" },
  { id: "night-line", sourceKey: "fixture", title: "夜行线", category: "剧情", year: 2026, image: "/demo/card-city.png", episodeLabel: "第 5 集", progress: 71, remaining: "剩余 12 分钟" },
  { id: "desert-archive", sourceKey: "fixture", title: "砂原档案", category: "纪录", year: 2024, image: "/demo/card-desert.png", episodeLabel: "第 1 集", progress: 24, remaining: "剩余 39 分钟" }
];

export const popularTitles: MediaTitle[] = [
  { id: "abyss-archive", sourceKey: "fixture", title: "深水档案", category: "悬疑", year: 2025, image: "/demo/card-abyss.png" },
  ...continueWatching.map(({ progress: _progress, remaining: _remaining, ...item }) => item)
];
