import type { PlaybackManifest } from "../types/ytv";
import { completedPlayback, restorablePosition } from "./progress";
import { progressStorageKey, type LocalPlaybackProgress } from "./local-progress";

export interface PlaybackSelection {
  sourceKey: string;
  externalId: string;
  episodeKey: string;
  titleName: string;
  episodeTitle: string;
}
export type PlaybackState = "idle" | "resolving" | "loading" | "ready" | "playing" | "paused" | "buffering" | "ended" | "error";
export interface PlaybackSnapshot { state: PlaybackState; message: string }
export interface PlaybackMedia extends EventTarget {
  currentTime: number;
  duration: number;
  seekable: TimeRanges;
  paused: boolean;
  ended: boolean;
  play(): Promise<void>;
  pause(): void;
}
export interface PlaybackMount { media: PlaybackMedia; destroy(): void }
interface Options {
 describeError?(cause: unknown): string;
  resolve(selection: PlaybackSelection): Promise<PlaybackManifest>;
  loadProgress?(selection: PlaybackSelection): Promise<LocalPlaybackProgress | undefined>;
  mount(manifest: PlaybackManifest): PlaybackMount;
  storage(): Pick<Storage, "getItem" | "setItem">;
  onState(snapshot: PlaybackSnapshot): void;
  onProgress(record: LocalPlaybackProgress): void;
  onStarted(selection: PlaybackSelection): void;
  onEnded(): void;
}

/** Resolution, recovery and progress without leaking player-library objects to pages. */
export function createPlaybackSession(options: Options) {
  let generation = 0;
  let disposed = false;
  let selection: PlaybackSelection | null = null;
  let active: PlaybackMount | null = null;
  let removeListeners = () => {};
  let timer: ReturnType<typeof setInterval> | undefined;
  let memoryProgress: LocalPlaybackProgress | null = null;
  let metadataReady = false;
  let restorePending = false;
  async function accountProgress(next: PlaybackSelection) {
    if (!options.loadProgress) return undefined;
    let timeout: ReturnType<typeof setTimeout> | undefined;
    try {
      return await Promise.race([
        options.loadProgress(next),
        new Promise<undefined>(resolve => { timeout = setTimeout(() => resolve(undefined), 2500); })
      ]);
    } catch { return undefined; }
    finally { clearTimeout(timeout); }
  }
  function state(value: PlaybackState, message = "") {
    if (!disposed) options.onState({ state: value, message });
  }
  function save() {
    const media = active?.media;
    if (!selection || !media || !metadataReady || restorePending || !Number.isFinite(media.duration) || media.duration <= 0 || !Number.isFinite(media.currentTime)) return;
    const record: LocalPlaybackProgress = { ...selection, position: Math.floor(media.currentTime), duration: Math.floor(media.duration),
      updatedAt: Date.now(), completed: completedPlayback(media.currentTime, media.duration) };
    memoryProgress = record;
    try { options.storage().setItem(progressStorageKey(selection.sourceKey, selection.externalId, selection.episodeKey), JSON.stringify(record)); } catch { /* Full or disabled storage must not interrupt playback. */ }
    options.onProgress(record);
  }
  function release() {
    save();
    removeListeners();
    removeListeners = () => {};
    clearInterval(timer);
    timer = undefined;
    metadataReady = false;
    restorePending = false;
    active?.destroy();
    active = null;
  }
  async function open(next: PlaybackSelection) {
    if (disposed) return;
    const request = ++generation;
    release(); // Save against the old locator before replacing it.
    const same = selection && selection.sourceKey === next.sourceKey && selection.externalId === next.externalId && selection.episodeKey === next.episodeKey;
    if (!same) memoryProgress = null;
    selection = { ...next };
    state("resolving", "正在连接播放线路");
    try {
      const [manifest, remoteRecord] = await Promise.all([options.resolve({ ...next }), accountProgress({ ...next })]);
      if (disposed || request !== generation) return;
      active = options.mount(manifest);
      const media = active.media;
      let started = false;
      let restored = false;
      let savedRecord = memoryProgress;
      if (!savedRecord) {
        try { savedRecord = JSON.parse(options.storage().getItem(progressStorageKey(next.sourceKey, next.externalId, next.episodeKey)) || "null"); } catch { /* Invalid records are ignored. */ }
      }
      if (remoteRecord && remoteRecord.sourceKey === next.sourceKey && remoteRecord.externalId === next.externalId && remoteRecord.episodeKey === next.episodeKey
        && (!savedRecord || remoteRecord.updatedAt > savedRecord.updatedAt)) savedRecord = remoteRecord;
      restorePending = !!savedRecord && !savedRecord.completed && Number.isFinite(savedRecord.position) && savedRecord.position > 5;
      const listeners: Array<[string, EventListener]> = [];
      const listen = (name: string, callback: () => void) => {
        const handler = () => { if (!disposed && request === generation) callback(); };
        media.addEventListener(name, handler);
        listeners.push([name, handler]);
      };
      const restore = () => {
        if (restored) return;
        const target = restorablePosition(savedRecord, media.duration, media.seekable);
        if (target > 0) { media.currentTime = target; restored = true; restorePending = false; }
      };
      listen("loadedmetadata", () => { metadataReady = true; restore(); state("ready"); });
      listen("canplay", () => { restore(); state(media.paused ? "ready" : "playing"); });
      listen("playing", () => { restore(); restorePending = false; restored = true; state("playing"); if (!started) { started = true; options.onStarted({ ...next }); } });
      listen("pause", () => { if (!media.ended) { state("paused"); save(); } });
      listen("waiting", () => state("buffering", "正在缓冲"));
      listen("seeked", () => { state(media.paused ? "paused" : "playing"); save(); });
      listen("ended", () => { state("ended"); save(); options.onEnded(); });
      listen("error", () => { save(); state("error", "这条线路暂时无法播放，可以重新连接或选择其他线路。"); });
      removeListeners = () => listeners.forEach(([name, handler]) => media.removeEventListener(name, handler));
      timer = setInterval(save, 10_000);
      state("loading", "正在载入影片");
    } catch (cause) {
      if (!disposed && request === generation) state("error", options.describeError?.(cause) || "没有取得可播放的地址，请重试或选择其他线路。");
    }
  }
  return {
    open,
    async retry() { if (selection) await open(selection); },
    async play() {
      const request = generation;
      try { await active?.media.play(); } catch { if (request === generation) state("ready", "请再次点击播放"); }
    },
    pause() { active?.media.pause(); },
    seek(seconds: number) { if (active && Number.isFinite(seconds)) active.media.currentTime = Math.max(0, Math.min(seconds, active.media.duration || 0)); },
    save,
    dispose() { if (disposed) return; generation++; release(); disposed = true; }
  };
}
