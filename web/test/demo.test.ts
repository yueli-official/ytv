import { describe, expect, it } from "vitest";
import { continueWatching, featuredTitle } from "../app/data/demo";

describe("YTV deterministic demo catalog", () => {
  it("keeps source states and progress explicit", () => {
    expect(featuredTitle.sources.some((source) => source.state === "degraded")).toBe(true);
    expect(featuredTitle.sources.filter((source) => source.selected)).toHaveLength(1);
    expect(continueWatching.every((item) => (item.progress || 0) > 0)).toBe(true);
  });
});
