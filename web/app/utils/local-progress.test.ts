import { describe, expect, it } from "vitest";
import { parseLegacyProgress, parseProgressStorageKey, progressStorageKey } from "./local-progress";

describe("local progress migration", () => {
  it("maps the legacy video_play_history contract without deleting source authority", () => {
    const records = parseLegacyProgress(JSON.stringify([{
      vod_id: 42, episode_index: 3, name: "环海之门", episode_title: "第四集",
      sourceKey: "legacy-source", lastPlayTime: 1788040000000, progress: 60, duration: 100
    }]));
    expect(records).toEqual([expect.objectContaining({
      sourceKey: "legacy-source", externalId: "42", episodeKey: "3", titleName: "环海之门", completed: false
    })]);
  });

  it("round-trips encoded locator segments", () => {
    const key = progressStorageKey("json:one", "id/42", "正片:1");
    expect(parseProgressStorageKey(key)).toEqual({ sourceKey: "json:one", externalId: "id/42", episodeKey: "正片:1" });
  });
});
