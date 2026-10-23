<script setup lang="ts">
import type { SearchGroup } from "~/types/ytv";

const props = defineProps<{ groups: SearchGroup[]; returnTo?: string; compact?: boolean }>();
function titleLink(source: string, id: string) {
  return { path: `/title/${encodeURIComponent(source)}/${encodeURIComponent(id)}`, query: props.returnTo ? { from: props.returnTo } : {} };
}

function sourceCount(group: SearchGroup) { return new Set(group.candidates.map(item => item.sourceKey)).size; }
function candidateLabel(group: SearchGroup, index: number) {
 const item = group.candidates[index]!;
 const siblings = group.candidates.filter(candidate => candidate.sourceKey === item.sourceKey);
 return siblings.length > 1 ? `${item.sourceName} · 条目 ${siblings.indexOf(item) + 1}` : item.sourceName;
}
const failedPosters = ref(new Set<string>());
function posterFor(group: SearchGroup) {
  return group.candidates.find(item => item.imageUrl && !failedPosters.value.has(item.imageUrl))?.imageUrl;
}
function failPoster(url: string) { failedPosters.value = new Set([...failedPosters.value, url]); }
function summaryFor(group: SearchGroup) {
  return group.candidates.map(item => item.summary?.trim()).find(value => value && value !== group.title && !['内详', '暂无', '暂无简介'].includes(value));
}
</script>

<template>
  <div class="grid gap-5" :class="{'compact-results':compact}">
    <article v-for="group in groups" :key="group.key" class="ytv-result-card">
      <NuxtLink v-if="group.candidates[0]" :to="titleLink(group.candidates[0].sourceKey, group.candidates[0].externalId)" :aria-label="`查看 ${group.title}`" class="ytv-result-poster">
        <UIcon name="i-tabler-photo-off" class="absolute left-1/2 top-1/2 size-7 -translate-x-1/2 -translate-y-1/2 text-white/25" />
        <img v-if="posterFor(group)" :key="posterFor(group)" :src="posterFor(group)" :alt="`${group.title} 海报`" class="absolute inset-0 size-full object-contain" loading="lazy" referrerpolicy="no-referrer" @error="failPoster(posterFor(group)!)">
      </NuxtLink>
      <div v-if="group.candidates[0]" class="min-w-0 space-y-3">
        <NuxtLink :to="titleLink(group.candidates[0].sourceKey, group.candidates[0].externalId)" class="text-white no-underline"><h2 class="ytv-result-title break-words text-sm font-semibold sm:text-lg leading-7">{{ group.title }}</h2></NuxtLink>
        <p class="text-sm leading-6 text-white/70">{{ [group.year, group.category, `${sourceCount(group)} 个来源`].filter(Boolean).join(' · ') }}</p>
        <p v-if="summaryFor(group)" class="hidden max-w-3xl text-sm leading-6 text-white/60 sm:line-clamp-2">{{ summaryFor(group) }}</p>
      </div>
      <details v-if="group.candidates.length" :open="!compact && group.candidates.length <= 3" :class="['ytv-result-sources', { 'ytv-result-sources-open': !compact && group.candidates.length <= 3 }]">
        <summary class="mb-3 cursor-pointer text-sm text-primary">选择来源 · {{ sourceCount(group) }} 个</summary>
        <div class="grid gap-2 xl:grid-cols-3" aria-label="影片来源">
          <NuxtLink v-for="(candidate, candidateIndex) in group.candidates" :key="`${candidate.sourceKey}:${candidate.externalId}`" :aria-label="candidateLabel(group, candidateIndex)" :to="titleLink(candidate.sourceKey, candidate.externalId)" class="flex min-h-14 items-center gap-3 rounded-lg border border-white/10 bg-white/[0.02] px-3 py-2 text-inherit no-underline hover:border-primary/40 hover:bg-white/5">
            <UIcon name="i-tabler-file-description" class="size-5 shrink-0 text-white/80" />
            <span class="min-w-0 flex-1"><strong class="block break-words text-sm font-semibold sm:text-base">{{ candidateLabel(group, candidateIndex) }}</strong><span class="mt-1 block break-words text-xs sm:text-sm text-white/75">{{ candidate.remarks || (candidate.episodeCount === 1 ? '正片' : candidate.episodeCount > 1 ? `${candidate.episodeCount} 集` : '查看播放选项') }}</span></span>
            <UIcon name="i-tabler-chevron-right" class="size-4 shrink-0 text-white/65" />
          </NuxtLink>
        </div>
      </details>
    </article>
  </div>
</template>

<style scoped>
.compact-results .ytv-result-card{grid-template-columns:76px minmax(0,1fr);gap:6px 14px;padding:0 0 16px;border:0;border-bottom:1px solid var(--ytv-line);border-radius:0;background:none}
.compact-results .ytv-result-poster{width:76px;aspect-ratio:2/3;grid-row:span 2}
.compact-results .ytv-result-title{font-size:15px;line-height:1.5}
.compact-results .ytv-result-sources{grid-column:2}
.compact-results .ytv-result-sources summary{font-size:12px;margin:6px 0}
.compact-results .ytv-result-sources[open]{grid-column:1/-1}
</style>
