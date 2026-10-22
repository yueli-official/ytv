import type { TitleResponse } from "../types/ytv";

export function titleLines(title?: TitleResponse | null) {
  const lines = new Map<string, { key: string; name: string; episodes: TitleResponse["episodes"] }>();
  for (const episode of title?.episodes || []) {
    const key = episode.lineKey || "0";
    if (!lines.has(key)) lines.set(key, { key, name: episode.lineName || "默认线路", episodes: [] });
    lines.get(key)!.episodes.push(episode);
  }
  return [...lines.values()];
}
export function isSingleFeature(title?: TitleResponse | null) {
  // A film may include a theatrical cut and a special edition on the same line.
  // Category is stronger evidence than the number of playback entries.
  const category = title?.title.category?.trim() || "";
  if (/连续剧|电视剧|国产剧|欧美剧|美剧|英剧|日剧|韩剧|港剧|台剧|泰剧|海外剧|短剧|综艺|动漫/u.test(category)) return false;
  if (/电影|动作片|科幻片|喜剧片|剧情片|爱情片|恐怖片|惊悚片|战争片|动画片|奇幻片|冒险片|犯罪片|纪录片|记录片/u.test(category)) return true;
  const lines = titleLines(title);
  return lines.length > 0 && lines.every((line) => line.episodes.length === 1);
}
export function watchPath(source: string, id: string, episode: string) {
  return `/watch/${encodeURIComponent(source)}/${encodeURIComponent(id)}/${encodeURIComponent(episode)}`;
}
export function searchReturnPath(value: unknown) {
  return typeof value === "string" && (value === "/search" || value.startsWith("/search?")) ? value : undefined;
}
