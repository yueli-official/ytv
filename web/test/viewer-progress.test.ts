import { describe, expect, it } from "vitest";
import { mergeViewerProgress } from "../app/utils/viewer-progress";

describe("viewer history reconciliation", () => {
  it("keeps unsynced local episodes and honors newer remote completion", () => {
    const local = { sourceKey: "source", externalId: "film", episodeKey: "0", titleName: "film", episodeTitle: "正片", position: 30, duration: 60, completed: false, updatedAt: 1000 };
    const remote = { ...local, positionSeconds: 60, durationSeconds: 60, completed: true, revision: 1, updatedAt: new Date(2000).toISOString() };
    expect(mergeViewerProgress([local, { ...local, episodeKey: "line1-0" }], [remote])).toEqual([
      { ...local, position: 60, completed: true, updatedAt: 2000 }, { ...local, episodeKey: "line1-0" }
    ]);
  });
});
