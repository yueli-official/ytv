<script setup lang="ts">
import { clearSearchHistory, readSearchHistory, rememberSearch } from "~/utils/search-preferences";

const props = withDefaults(defineProps<{
  compact?: boolean;
  initialValue?: string;
  initialSources?: string[];
  showHistory?: boolean;
}>(), {
  compact: false,
  initialValue: "",
  showHistory: true
});

const draft = ref(props.initialValue);
watch(() => props.initialValue, value => { draft.value = value; });
const composing = ref(false);
const history = ref<string[]>([]);
const hydrated = ref(false);
const { selected: selectedSources } = useViewerSourcePreferences();
const searchEnabled = computed(() => draft.value.trim() !== "" && !composing.value && (props.initialSources !== undefined || selectedSources.value.length > 0));
const routeSources = computed(() => (props.initialSources ?? selectedSources.value).join(","));
const historyHeading = useId();

onMounted(() => {
  history.value = readSearchHistory();
  hydrated.value = true;
});

async function submit() {
  if (!searchEnabled.value) return;
  const query = draft.value.trim();
  history.value = rememberSearch(query);
  await navigateTo({ path: "/search", query: { q: query, ...(routeSources.value ? { sources: routeSources.value } : {}) } });
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

    <section v-if="showHistory && history.length" :aria-labelledby="historyHeading">
      <div class="mb-2 flex items-center justify-between gap-3">
        <h2 :id="historyHeading" class="text-xs font-semibold text-white/60">搜索历史</h2>
        <UButton label="清空" color="neutral" variant="link" size="xs" @click="clearHistory" />
      </div>
      <div class="flex gap-2 overflow-x-auto pb-1">
        <NuxtLink v-for="item in history" :key="item" :to="historyTo(item)" class="shrink-0 rounded-lg bg-[#111a24] px-3 py-2 text-xs text-white/80 no-underline hover:bg-[#162333] hover:text-white">{{ item }}</NuxtLink>
      </div>
    </section>
  </div>
</template>
