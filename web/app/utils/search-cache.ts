import type { DiscoveryResponse } from '~/types/ytv';
const storageKey = 'ytv-search-results-v2';
const limit = 20;
type Entry = { key: string; pages: Record<string, DiscoveryResponse> };
function read(): Entry[] {
  // v1 may contain catalog-wide results returned by a detail-only search.
  localStorage.removeItem('ytv-search-results-v1');
  return JSON.parse(localStorage.getItem(storageKey) || '[]');
}
function write(entries: Entry[]) {
  // Keep the newest searches; a large search may require evicting more than one.
  while (entries.length) {
    try { localStorage.setItem(storageKey, JSON.stringify(entries)); return; }
    catch (cause) {
      if (!(cause instanceof DOMException) || cause.name !== 'QuotaExceededError' || entries.length === 1) throw cause;
      entries.pop();
    }
  }
  localStorage.removeItem(storageKey);
}
export function cachedSearch(key: string, page: number) {
  const entries = read();
  const entry = entries.find(item => item.key === key);
  if (!entry?.pages[page]) return undefined;
  write([entry, ...entries.filter(item => item !== entry)]);
  return entry.pages[page];
}
export function cacheSearch(key: string, page: number, result: DiscoveryResponse) {
  const entries = read();
  const entry = entries.find(item => item.key === key) || { key, pages: {} };
  entry.pages[page] = result;
  write([entry, ...entries.filter(item => item.key !== key)].slice(0, limit));
}
export function clearCachedSearch(key: string) { write(read().filter(item => item.key !== key)); }
