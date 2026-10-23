import { ytvFailureMessage } from "~/utils/ytv-failure";
import { rememberSearch } from "~/utils/search-preferences";
export function useSearchView() {
const route = useRoute();
const api = useYtvApi();
const sourceSettingsOpen = useState("ytv-source-settings-open", () => false);
const query = computed(() => String(route.query.q || "").trim());
const page = computed(() => Math.max(1, Number(route.query.page) || 1));
const sourceKeys = computed(() => String(route.query.sources || "").split(",").map((value) => value.trim()).filter(Boolean));
const { data, pending, error, refresh } = useAsyncData(
  () => `ytv-search-${query.value}-${page.value}-${sourceKeys.value.join("-")}`,
  () => query.value ? api.search(query.value, page.value, sourceKeys.value) : Promise.resolve(null),
  { watch: [query, page, sourceKeys] }
);
// Keep SSR and hydration deterministic; minimum-duration timers start after mount.
const mounted = ref(false);
onMounted(() => { mounted.value = true; });
const delayedLoading = useMinimumLoading(computed(() => mounted.value && pending.value));
const showLoading = computed(() => mounted.value ? delayedLoading.value : pending.value);
const verificationSources = computed(() => data.value?.sources.filter(source => source.outcome === "verification_required") || []);
const pageLink = (value: number) => ({ path: "/search", query: { q: query.value, ...(value > 1 ? { page: value } : {}), ...(sourceKeys.value.length ? { sources: sourceKeys.value.join(",") } : {}) } });
watch(query, (value) => {
  if (import.meta.client && value) rememberSearch(value);
}, { immediate: true });
async function retry() {
  await refresh();
}
useSeoMeta({ title: () => query.value ? `搜索：${query.value}` : "搜索" });
return reactive({ ytvFailureMessage, rememberSearch, route, api, sourceSettingsOpen, query, page, sourceKeys, data, pending, error, refresh, showLoading, verificationSources, pageLink, retry });
}
