import type { PlaybackProgress } from "../types/ytv";
import { loadLocalProgress, migrateLegacyProgress, type LocalPlaybackProgress } from "./local-progress";

export function browserProgress(): LocalPlaybackProgress[] {
  try {
    migrateLegacyProgress(localStorage);
    return loadLocalProgress(localStorage);
  } catch { return []; }
}

/** Prefer the latest observation of each episode, including completion on another device. */
export function mergeViewerProgress(local: LocalPlaybackProgress[], account: PlaybackProgress[] = []): LocalPlaybackProgress[] {
  const records = [...local, ...account.map(item => ({
    sourceKey: item.sourceKey, externalId: item.externalId, episodeKey: item.episodeKey,
    titleName: item.titleName, episodeTitle: item.episodeTitle, position: item.positionSeconds,
    duration: item.durationSeconds, completed: item.completed, updatedAt: new Date(item.updatedAt).getTime()
  }))];
  const seen = new Set<string>();
  return records.sort((a, b) => b.updatedAt - a.updatedAt).filter(item => {
    const key = JSON.stringify([item.sourceKey, item.externalId, item.episodeKey]);
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}
