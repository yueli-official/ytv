const historyKey = "ytv-search-history-v1";
const disabledSourceKey = "ytv-search-disabled-sources-v1";
const historyLimit = 8;

function readList(key: string): string[] | null {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return null;
    const value = JSON.parse(raw);
    if (!Array.isArray(value)) return [];
    return value.filter((item): item is string => typeof item === "string" && item.trim() !== "");
  } catch {
    return [];
  }
}

function writeList(key: string, values: string[]) {
  try { localStorage.setItem(key, JSON.stringify(values)); } catch { /* Search remains usable with unavailable storage. */ }
}

export function readSearchHistory() {
  return readList(historyKey) || [];
}

export function rememberSearch(query: string) {
  const normalized = query.trim().replace(/\s+/gu, " ");
  if (!normalized) return readSearchHistory();
  const next = [normalized, ...readSearchHistory().filter((item) => item.localeCompare(normalized, "zh-CN", { sensitivity: "base" }) !== 0)].slice(0, historyLimit);
  writeList(historyKey, next);
  return next;
}

export function clearSearchHistory() {
  try { localStorage.removeItem(historyKey); } catch { /* Clearing the visible list still works. */ }
}

export function readSelectedSources(availableKeys: string[]) {
	const disabled = new Set(readList(disabledSourceKey) || []);
	return availableKeys.filter((key) => !disabled.has(key));
}

export function writeSelectedSources(values: string[], availableKeys: string[]) {
	const selected = new Set(values);
	writeList(disabledSourceKey, availableKeys.filter((key) => !selected.has(key)));
}
