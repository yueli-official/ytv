import { completedPlayback, type PlaybackProgressRecord } from "./progress";

const currentPrefix = "ytv-progress-v2:";
const previousPrefix = "ytv-progress-v1:";
const legacyKey = "video_play_history";
const migrationMarker = "ytv-legacy-progress-migrated-v1";

export interface LocalPlaybackProgress extends PlaybackProgressRecord {
  sourceKey: string;
  externalId: string;
  episodeKey: string;
  titleName: string;
  episodeTitle: string;
}

export function progressStorageKey(sourceKey: string, externalId: string, episodeKey: string) {
  return `${currentPrefix}${encodeURIComponent(sourceKey)}:${encodeURIComponent(externalId)}:${encodeURIComponent(episodeKey)}`;
}

export function parseProgressStorageKey(key: string) {
  if (key.startsWith(currentPrefix)) {
    const fields = key.slice(currentPrefix.length).split(":");
    if (fields.length !== 3) return null;
    try {
      return { sourceKey: decodeURIComponent(fields[0] || ""), externalId: decodeURIComponent(fields[1] || ""), episodeKey: decodeURIComponent(fields[2] || "") };
    } catch {
      return null;
    }
  }
  if (!key.startsWith(previousPrefix)) return null;
  const fields = key.slice(previousPrefix.length).split(":");
  if (fields.length < 3) return null;
  return {
    sourceKey: fields[0] || "",
    externalId: fields.slice(1, -1).join(":"),
    episodeKey: fields.at(-1) || ""
  };
}

export function parseLegacyProgress(raw: string | null): LocalPlaybackProgress[] {
  if (!raw) return [];
  try {
    const values = JSON.parse(raw) as unknown;
    if (!Array.isArray(values)) return [];
    const result: LocalPlaybackProgress[] = [];
    for (const value of values.slice(0, 100)) {
      if (!value || typeof value !== "object") continue;
      const item = value as Record<string, unknown>;
      const sourceKey = String(item.sourceKey || "").trim();
      const externalId = String(item.vod_id ?? "").trim();
      const episodeKey = String(item.episode_index ?? "").trim();
      const position = Math.max(0, Math.floor(Number(item.progress)));
      const duration = Math.floor(Number(item.duration));
      const updatedAt = Math.floor(Number(item.lastPlayTime));
      if (!sourceKey || !externalId || !episodeKey || !Number.isFinite(position) || !Number.isFinite(duration) || duration <= 0 || !Number.isFinite(updatedAt) || updatedAt <= 0) continue;
      result.push({
        sourceKey, externalId, episodeKey,
        titleName: String(item.name || "").trim().slice(0, 200),
        episodeTitle: String(item.episode_title || "").trim().slice(0, 200),
        position, duration, updatedAt, completed: completedPlayback(position, duration)
      });
    }
    return result;
  } catch {
    return [];
  }
}

export function migrateLegacyProgress(storage: Storage) {
  if (storage.getItem(migrationMarker)) return 0;
  const records = parseLegacyProgress(storage.getItem(legacyKey));
  for (const record of records) {
    const key = progressStorageKey(record.sourceKey, record.externalId, record.episodeKey);
    if (!storage.getItem(key)) storage.setItem(key, JSON.stringify(record));
  }
  storage.setItem(migrationMarker, new Date().toISOString());
  return records.length;
}

export function loadLocalProgress(storage: Storage): LocalPlaybackProgress[] {
  const result: LocalPlaybackProgress[] = [];
  for (let index = 0; index < storage.length; index += 1) {
    const key = storage.key(index);
    if (!key) continue;
    const locator = parseProgressStorageKey(key);
    if (!locator?.sourceKey || !locator.externalId || !locator.episodeKey) continue;
    try {
      const value = JSON.parse(storage.getItem(key) || "") as Partial<LocalPlaybackProgress>;
      if (!Number.isFinite(value.position) || !Number.isFinite(value.duration) || Number(value.duration) <= 0 || !Number.isFinite(value.updatedAt)) continue;
      result.push({
        ...locator,
        titleName: String(value.titleName || ""), episodeTitle: String(value.episodeTitle || ""),
        position: Math.max(0, Math.floor(Number(value.position))), duration: Math.floor(Number(value.duration)),
        updatedAt: Math.floor(Number(value.updatedAt)), completed: Boolean(value.completed)
      });
    } catch {
      continue;
    }
  }
  return result.sort((left, right) => right.updatedAt - left.updatedAt).slice(0, 100);
}
