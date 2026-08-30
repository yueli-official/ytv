import { beforeEach, describe, expect, it, vi } from "vitest";
import { clearSearchHistory, readSearchHistory, readSelectedSources, rememberSearch, writeSelectedSources } from "./search-preferences";

const values = new Map<string, string>();

beforeEach(() => {
  values.clear();
  vi.stubGlobal("localStorage", {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
    removeItem: (key: string) => values.delete(key)
  });
});

describe("search preferences", () => {
  it("keeps eight normalized recent unique searches", () => {
    rememberSearch("  环海   之门 ");
    rememberSearch("环海 之门");
    for (let index = 0; index < 9; index++) rememberSearch(`查询 ${index}`);
    expect(readSearchHistory()).toHaveLength(8);
    expect(readSearchHistory()[0]).toBe("查询 8");
    clearSearchHistory();
    expect(readSearchHistory()).toEqual([]);
  });

  it("stores disabled source keys so newly enabled sources stay selected", () => {
    writeSelectedSources(["a"], ["a", "b"]);
    expect(readSelectedSources(["a", "b", "c"])).toEqual(["a", "c"]);
  });
});
