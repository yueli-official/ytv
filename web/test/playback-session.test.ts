import { afterEach, describe, expect, it, vi } from "vitest";
import { createPlaybackSession, type PlaybackMedia, type PlaybackSelection, type PlaybackSnapshot } from "../app/utils/playback-session";
import type { LocalPlaybackProgress } from "../app/utils/local-progress";
import type { PlaybackManifest } from "../app/types/ytv";

class Media extends EventTarget implements PlaybackMedia {
  currentTime = 0;
  duration = 600;
  paused = true;
  ended = false;
  seekable = { length: 1, start: () => 0, end: () => 600 };
  async play() { this.paused = false; this.dispatchEvent(new Event("playing")); }
  pause() { this.paused = true; this.dispatchEvent(new Event("pause")); }
}
const choice = (id: string): PlaybackSelection => ({ sourceKey: "source", externalId: id, episodeKey: "0", titleName: id, episodeTitle: "正片" });
const manifest = (url: string): PlaybackManifest => ({ sourceName: "source", episodeKey: "0", url, engine: "direct", timeline: "source:title" });
function setup(resolve = vi.fn(async () => manifest("first.mp4"))) {
  const media: Media[] = [];
  const states: PlaybackSnapshot[] = [];
  const storage = new Map<string, string>();
  const onProgress = vi.fn();
  const destroy = vi.fn();
  const onStarted = vi.fn();
  const options = { resolve, loadProgress: undefined as undefined | ((selection: PlaybackSelection) => Promise<LocalPlaybackProgress | undefined>), mount: vi.fn(() => { const value = new Media(); media.push(value); return { media: value, destroy }; }),
    storage: () => ({ getItem: (key: string) => storage.get(key) || null, setItem: (key: string, value: string) => { storage.set(key, value); } }),
    onState: (state: PlaybackSnapshot) => states.push(state), onProgress, onStarted, onEnded: vi.fn() };
  return { session: createPlaybackSession(options), options, media, states, storage, onProgress, destroy, onStarted };
}
afterEach(() => vi.useRealTimers());
describe("YTV playback session", () => {
  it("restores account progress on a fresh browser and keeps a newer local position", async () => {
    const value = setup();
    value.options.loadProgress = async () => ({ ...choice("film"), position: 120, duration: 600, updatedAt: 100, completed: false });
    await value.session.open(choice("film"));
    value.media[0]!.dispatchEvent(new Event("loadedmetadata"));
    expect(value.media[0]!.currentTime).toBe(120);
    value.media[0]!.currentTime = 180;
    await value.session.retry();
    value.media[1]!.dispatchEvent(new Event("loadedmetadata"));
    expect(value.media[1]!.currentTime).toBe(180);
    value.session.dispose();
  });
  it("a stalled account request cannot block playback or seek it later", async () => {
    vi.useFakeTimers();
    const value = setup();
    let finish!: (value: LocalPlaybackProgress) => void;
    value.options.loadProgress = () => new Promise(resolve => { finish = resolve; });
    const opening = value.session.open(choice("film"));
    await vi.advanceTimersByTimeAsync(2500); await opening;
    value.media[0]!.dispatchEvent(new Event("loadedmetadata"));
    await value.session.play();
    finish({ ...choice("film"), position: 120, duration: 600, updatedAt: Date.now(), completed: false });
    await Promise.resolve();
    expect(value.media[0]!.currentTime).toBe(0);
    expect(value.states.at(-1)?.state).toBe("playing");
    value.session.dispose();
    expect(vi.getTimerCount()).toBe(0);
  });
  it("ignores an old resolution after switching films", async () => {
    let finish!: (value: PlaybackManifest) => void;
    const pending = new Promise<PlaybackManifest>((resolve) => { finish = resolve; });
    const resolve = vi.fn().mockReturnValueOnce(pending).mockResolvedValueOnce(manifest("new.mp4"));
    const value = setup(resolve);
    const old = value.session.open(choice("old"));
    await value.session.open(choice("new"));
    finish(manifest("old.mp4")); await old;
    expect(value.options.mount).toHaveBeenCalledTimes(1);
    expect(value.options.mount).toHaveBeenCalledWith(manifest("new.mp4"));
    value.session.dispose();
  });
  it("retry re-resolves the URL and restores the current position", async () => {
    const value = setup(vi.fn().mockResolvedValueOnce(manifest("expired.mp4")).mockResolvedValueOnce(manifest("fresh.mp4")));
    await value.session.open(choice("film"));
    value.media[0]!.dispatchEvent(new Event("loadedmetadata")); value.media[0]!.currentTime = 120;
    value.media[0]!.dispatchEvent(new Event("error"));
    expect(value.states.at(-1)?.state).toBe("error");
    await value.session.retry(); value.media[1]!.dispatchEvent(new Event("loadedmetadata"));
    expect(value.options.resolve).toHaveBeenCalledTimes(2);
    expect(value.options.mount).toHaveBeenLastCalledWith(manifest("fresh.mp4"));
    expect(value.media[1]!.currentTime).toBe(120);
    value.session.dispose();
  });
  it("saves old progress under the old film and detaches its events", async () => {
    const value = setup(); await value.session.open(choice("old"));
    value.media[0]!.dispatchEvent(new Event("loadedmetadata")); value.media[0]!.currentTime = 75;
    await value.session.open(choice("new"));
    expect(value.onProgress).toHaveBeenLastCalledWith(expect.objectContaining({ externalId: "old", position: 75 }));
    const count = value.states.length;
    value.media[0]!.dispatchEvent(new Event("error"));
    expect(value.states).toHaveLength(count);
    value.session.dispose();
  });
  it("keeps playing with unavailable storage and releases its timer once", async () => {
    vi.useFakeTimers(); const value = setup();
    value.options.storage = () => { throw new Error("disabled"); };
    await value.session.open(choice("film")); value.media[0]!.dispatchEvent(new Event("loadedmetadata"));
    await value.session.play(); await value.session.play();
    expect(value.onStarted).toHaveBeenCalledTimes(1);
    expect(() => vi.advanceTimersByTime(10_000)).not.toThrow();
    value.session.dispose(); value.session.dispose();
    expect(value.destroy).toHaveBeenCalledTimes(1); expect(vi.getTimerCount()).toBe(0);
  });
  it("does not mount after disposal while resolving", async () => {
    let finish!: (value: PlaybackManifest) => void;
    const value = setup(vi.fn(() => new Promise<PlaybackManifest>((resolve) => { finish = resolve; })));
    const opening = value.session.open(choice("film")); value.session.dispose(); finish(manifest("late.mp4")); await opening;
    expect(value.options.mount).not.toHaveBeenCalled();
  });
  it("preserves saved progress if metadata arrives before the seekable range and playback fails", async () => {
    const value = setup();
    const key = "ytv-progress-v2:source:film:0";
    value.storage.set(key, JSON.stringify({ position: 120, duration: 600, completed: false, updatedAt: Date.now() }));
    await value.session.open(choice("film"));
    value.media[0]!.seekable.length = 0;
    value.media[0]!.dispatchEvent(new Event("loadedmetadata"));
    value.media[0]!.dispatchEvent(new Event("error"));
    await value.session.retry();
    value.media[1]!.dispatchEvent(new Event("loadedmetadata"));
    expect(value.media[1]!.currentTime).toBe(120);
    value.session.dispose();
  });
});
