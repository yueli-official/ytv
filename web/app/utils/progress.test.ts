import { describe, expect, it } from "vitest";
import { completedPlayback, restorablePosition, type PlaybackProgressRecord } from "./progress";

function ranges(...values: Array<[number, number]>): TimeRanges {
  return {
    length: values.length,
    start: (index) => values[index]?.[0] || 0,
    end: (index) => values[index]?.[1] || 0
  };
}

describe("playback progress", () => {
  it("marks the last thirty seconds or final ten percent complete", () => {
    expect(completedPlayback(571, 600)).toBe(true);
    expect(completedPlayback(91, 100)).toBe(true);
    expect(completedPlayback(60, 100)).toBe(false);
  });

  it("restores only inside a seekable range", () => {
    const record: PlaybackProgressRecord = { position: 42, duration: 100, updatedAt: 1, completed: false };
    expect(restorablePosition(record, 100, ranges([0, 90]))).toBe(42);
    expect(restorablePosition(record, 100, ranges([50, 90]))).toBe(0);
    expect(restorablePosition({ ...record, completed: true }, 100, ranges([0, 90]))).toBe(0);
  });
});
