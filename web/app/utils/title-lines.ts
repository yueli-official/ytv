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
  const lines = titleLines(title);
  return lines.length > 0 && lines.every((line) => line.episodes.length === 1);
}
export function watchPath(source: string, id: string, episode: string) {
  return `/watch/${encodeURIComponent(source)}/${encodeURIComponent(id)}/${encodeURIComponent(episode)}`;
}
export function searchReturnPath(value: unknown) {
  return typeof value === "string" && (value === "/search" || value.startsWith("/search?")) ? value : undefined;
}
