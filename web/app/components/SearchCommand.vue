<script setup lang="ts">
import { clearSearchHistory, readSearchHistory, readSelectedSources, rememberSearch, writeSelectedSources } from "~/utils/search-preferences";

const props = withDefaults(defineProps<{
  compact?: boolean;
  initialValue?: string;
  initialSources?: string[];
  showHistory?: boolean;
  showSources?: boolean;
}>(), {
  compact: false,
  initialValue: "",
  showHistory: true,
  showSources: true
});

const api = useYtvApi();
const draft = ref(props.initialValue);
const composing = ref(false);
const history = ref<string[]>([]);
const selectedSources = useState<string[]>("ytv-selected-source-keys", () => []);
const hydrated = ref(false);
const selectionInitialized = ref(false);
const { data: sourceData } = await useAsyncData("ytv-public-sources", () => api.publicSources());
const availableKeys = computed(() => (sourceData.value?.sources || []).map((source) => source.key));
const searchEnabled = computed(() => draft.value.trim() !== "" && !composing.value && (!props.showSources || selectedSources.value.length > 0));
const routeSources = computed(() => selectedSources.value.join(","));

if (availableKeys.value.length && selectedSources.value.length === 0) {
  selectedSources.value = props.initialSources?.length
    ? props.initialSources.filter((key) => availableKeys.value.includes(key))
    : [...availableKeys.value];
}

function syncAvailableSources(keys: string[]) {
  if (!hydrated.value || keys.length === 0) return;
  const available = new Set(keys);
  if (props.initialSources === undefined) {
    selectedSources.value = readSelectedSources(keys);
  } else if (!selectionInitialized.value) {
    selectedSources.value = props.initialSources.length ? props.initialSources.filter((key) => available.has(key)) : [...keys];
  } else {
    selectedSources.value = selectedSources.value.filter((key) => available.has(key));
  }
  selectionInitialized.value = true;
}

onMounted(() => {
  history.value = readSearchHistory();
  hydrated.value = true;
  syncAvailableSources(availableKeys.value);
});

watch(availableKeys, (keys) => {
  syncAvailableSources(keys);
}, { flush: "post" });

async function submit() {
  if (!searchEnabled.value) return;
  const query = draft.value.trim();
  history.value = rememberSearch(query);
  await navigateTo({ path: "/search", query: { q: query, ...(routeSources.value ? { sources: routeSources.value } : {}) } });
}

function toggleSource(key: string) {
	if (selectedSources.value.includes(key) && selectedSources.value.length === 1) return;
  selectedSources.value = selectedSources.value.includes(key)
    ? selectedSources.value.filter((item) => item !== key)
    : [...selectedSources.value, key];
  writeSelectedSources(selectedSources.value, availableKeys.value);
}

function clearHistory() {
  clearSearchHistory();
  history.value = [];
}

function historyTo(query: string) {
  return { path: "/search", query: { q: query, ...(routeSources.value ? { sources: routeSources.value } : {}) } };
}
</script>

<template>
  <div class="w-full space-y-3">
    <form role="search" action="/search" method="get" @submit.prevent="submit">
      <input v-if="routeSources" type="hidden" name="sources" :value="routeSources">
      <UFieldGroup class="w-full">
        <UInput
          v-model="draft"
          name="q"
          icon="i-tabler-search"
          autocomplete="off"
          :size="compact ? 'lg' : 'xl'"
          placeholder="搜索想看的影视"
          aria-label="搜索想看的影视"
          class="min-w-0 flex-1"
          :ui="{
            root: 'w-full',
            base: compact
              ? 'min-h-11 bg-default/95 text-highlighted'
              : 'min-h-16 bg-black/70 text-white placeholder:text-white/60 sm:min-h-18 sm:text-lg'
          }"
          @compositionstart="composing = true"
          @compositionend="composing = false"
        />
        <UButton
          type="submit"
          icon="i-tabler-arrow-right"
          :size="compact ? 'lg' : 'xl'"
          :aria-label="draft.trim() ? `搜索 ${draft.trim()}` : '开始搜索'"
          :disabled="hydrated ? !searchEnabled : false"
          class="min-w-12 justify-center sm:min-w-16"
        />
      </UFieldGroup>
    </form>

    <section v-if="showHistory && history.length" aria-labelledby="search-history-heading">
      <div class="mb-2 flex items-center justify-between gap-3">
        <h2 id="search-history-heading" class="text-xs font-semibold text-white/60">搜索历史</h2>
        <UButton label="清空" color="neutral" variant="link" size="xs" @click="clearHistory" />
      </div>
      <div class="flex gap-2 overflow-x-auto pb-1">
        <NuxtLink v-for="item in history" :key="item" :to="historyTo(item)" class="shrink-0 rounded-lg bg-[#111a24] px-3 py-2 text-xs text-white/80 no-underline hover:bg-[#162333] hover:text-white">{{ item }}</NuxtLink>
      </div>
    </section>

    <section v-if="showSources && sourceData?.sources.length" aria-labelledby="search-sources-heading">
      <div class="mb-2 flex items-center justify-between gap-3">
        <h2 id="search-sources-heading" class="text-xs font-semibold text-white/60">搜索来源</h2>
        <span class="text-xs text-white/55">已选 {{ selectedSources.length }} / {{ sourceData.sources.length }}</span>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          v-for="source in sourceData.sources"
          :key="source.key"
          :label="source.name"
          :aria-pressed="selectedSources.includes(source.key)"
          :color="selectedSources.includes(source.key) ? 'primary' : 'neutral'"
          :variant="selectedSources.includes(source.key) ? 'soft' : 'outline'"
          size="sm"
          @click="toggleSource(source.key)"
        />
      </div>
      <p v-if="hydrated && selectedSources.length === 1" class="mt-2 text-xs text-white/55">至少保留一个搜索来源。</p>
    </section>
  </div>
</template>
