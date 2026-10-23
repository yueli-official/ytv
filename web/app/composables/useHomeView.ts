
export function useHomeView() {
const api = useYtvApi();
const sourcePage = ref(1);
const trendingPage = ref(1);
const trendingType = ref("movie");
const trendingTag = ref("热门");
const failedFeaturePosters = ref(new Set<string>());
const failedTrendingPosters = ref(new Set<string>());
const tags = ["热门", "美剧", "英剧", "日剧", "韩剧", "港剧", "综艺", "纪录片"];
const viewerSourceKeys = useState<string[]>("ytv-selected-source-keys", () => []);
const { rail: trendingRail, canPrevious: canPreviousTrending, canNext: canNextTrending, previous: previousTrending, next: nextTrending } = useHorizontalRail();
const { data: sourceData, pending: sourcePending, error: sourceError, refresh: refreshSources } = useAsyncData(
  () => `ytv-discover-${sourcePage.value}`,
  () => api.discover(sourcePage.value),
  { watch: [sourcePage] }
);
const { data: trendingData, pending: trendingPending, error: trendingError, refresh: refreshTrending } = useAsyncData(
  () => `ytv-trending-${trendingType.value}-${trendingTag.value}-${trendingPage.value}`,
  () => api.trending(trendingType.value, trendingTag.value, trendingPage.value, 12),
  { watch: [trendingType, trendingTag, trendingPage] }
);
const featureArtwork = ref<HTMLImageElement>();
const activeFeatureIndex = ref(0);
const features = computed(() => {
  const items = trendingData.value?.items || [];
  const selected = items.filter(item => Number(item.score) >= 7).slice(0, 3);
  return [...selected, ...items.filter(item => !selected.includes(item))].slice(0, 3);
});
const activeFeature = computed(() => features.value[activeFeatureIndex.value] || features.value[0]);
const remainingTrending = computed(() => trendingData.value?.items || []);
watch(features, () => { activeFeatureIndex.value = 0; });
onMounted(() => {
  const image = featureArtwork.value;
  if (image?.complete && !image.naturalWidth && image.dataset.id) failFeaturePoster(image.dataset.id);
});
function selectType(value: string) {
  trendingType.value = value;
  trendingTag.value = "热门";
  trendingPage.value = 1;
}
function selectTag(value: string) {
  trendingTag.value = value;
  trendingPage.value = 1;
}
function changeSourceBatch() {
  sourcePage.value += 1;
}
function changeTrendingBatch() {
  trendingPage.value += 1;
}
async function searchTrending(title: string) {
  await navigateTo({ path: "/search", query: { q: title, ...(viewerSourceKeys.value.length ? { sources: viewerSourceKeys.value.join(",") } : {}) } });
}
function trendingPoster(item: { id: string; coverUrl: string }) {
  return item.coverUrl && !failedTrendingPosters.value.has(item.id) ? item.coverUrl : "";
}
function failTrendingPoster(id: string) {
  failedTrendingPosters.value = new Set([...failedTrendingPosters.value, id]);
}
function failFeaturePoster(id: string) {
  if (failedFeaturePosters.value.has(id)) failTrendingPoster(id);
  else failedFeaturePosters.value = new Set([...failedFeaturePosters.value, id]);
}
useSeoMeta({
  title: "发现好电影",
  description: "搜索想看的电影，浏览热门影片，接着上次的进度继续观看。"
});
return reactive({ api, sourcePage, trendingPage, trendingType, trendingTag, failedFeaturePosters, failedTrendingPosters, tags, viewerSourceKeys, trendingRail, canPreviousTrending, canNextTrending, previousTrending, nextTrending, sourceData, sourcePending, sourceError, refreshSources, trendingData, trendingPending, trendingError, refreshTrending, featureArtwork, activeFeatureIndex, features, activeFeature, remainingTrending, selectType, selectTag, changeSourceBatch, changeTrendingBatch, searchTrending, trendingPoster, failTrendingPoster, failFeaturePoster });
}
