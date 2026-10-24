import { cachedSearch, cacheSearch, clearCachedSearch } from '~/utils/search-cache';
import { ytvFailureMessage } from "~/utils/ytv-failure";
import { rememberSearch } from "~/utils/search-preferences";
export function useSearchView() {
const route = useRoute();
const api = useYtvApi();
const sourceSettingsOpen = useState("ytv-source-settings-open", () => false);
const query = computed(() => String(route.query.q || "").trim());
const page = computed(() => Math.max(1, Math.floor(Number(route.query.page)) || 1));
const sourceKeys = computed(() => String(route.query.sources || "").split(",").map((value) => value.trim()).filter(Boolean));
const searchKey = computed(() => JSON.stringify([query.value, [...new Set(sourceKeys.value)].sort()]));
const { data, status, pending, error, refresh } = useAsyncData(
  () => `ytv-search-${searchKey.value}-${page.value}`,
  async () => {
    if (!query.value) return null;
    const key = searchKey.value, requestedPage = page.value;
    const cached = cachedSearch(key, requestedPage);
    if (cached) return cached;
    const result = await api.search(query.value, requestedPage, sourceKeys.value);
    cacheSearch(key, requestedPage, result);
    return result;
  },
  { server: false, immediate: false }
);
if (import.meta.client) onNuxtReady(() => { void refresh(); });
// Idle means the browser has not read its cache yet, not an empty result.
// Cache hits should render immediately, without a minimum loading duration.
const showLoading = computed(() => Boolean(query.value) && (status.value === 'idle' || pending.value));
const verificationSources = computed(() => data.value?.sources.filter(source => source.outcome === "verification_required") || []);
const pageLink = (value: number) => ({ path: "/search", query: { q: query.value, ...(value > 1 ? { page: value } : {}), ...(sourceKeys.value.length ? { sources: sourceKeys.value.join(",") } : {}) } });
watch(query, (value) => {
  if (import.meta.client && value) rememberSearch(value);
}, { immediate: true });
async function retry() {
  clearCachedSearch(searchKey.value);
  await refresh();
}
useSeoMeta({ title: () => query.value ? `搜索：${query.value}` : "搜索" });
return reactive({ ytvFailureMessage, rememberSearch, route, api, sourceSettingsOpen, query, page, sourceKeys, data, pending, error, refresh, showLoading, verificationSources, pageLink, retry });
}
