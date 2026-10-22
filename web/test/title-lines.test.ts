import { describe, expect, it } from "vitest";
import { isSingleFeature } from "../app/utils/title-lines";
import type { TitleResponse } from "../app/types/ytv";
const title = (category: string, names: string[]) => ({
  title: { category }, episodes: names.map((name, index) => ({ key: String(index), title: name }))
}) as TitleResponse;
describe("film and series presentation", () => {
  it("keeps alternate film editions under viewing options without a next episode", () => {
    expect(isSingleFeature(title("科幻片", ["HD中字", "特别版"]))).toBe(true);
    expect(isSingleFeature(title("剧情片", ["导演剪辑版", "公映版"]))).toBe(true);
  });
  it("keeps a series with only its first episode in series presentation", () => {
    expect(isSingleFeature(title("国产剧", ["第1集"]))).toBe(false);
    expect(isSingleFeature(title("日韩动漫", ["第1集"]))).toBe(false);
  });
  it("retains the entry-count fallback for sources with no category", () => {
    expect(isSingleFeature(title("", ["正片"]))).toBe(true);
    expect(isSingleFeature(title("", ["第1集", "第2集"]))).toBe(false);
    expect(isSingleFeature(null)).toBe(false);
  });
});
