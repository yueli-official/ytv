<script setup lang="ts">
import type { SearchGroup } from "~/types/ytv";

const props = defineProps<{ groups: SearchGroup[]; returnTo?: string }>();
function titleLink(source: string, id: string) {
  return { path: `/title/${encodeURIComponent(source)}/${encodeURIComponent(id)}`, query: props.returnTo ? { from: props.returnTo } : {} };
}

function hideBrokenImage(event: Event) {
  (event.currentTarget as HTMLImageElement).hidden = true;
}
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
    <article v-for="group in groups" :key="group.key" class="min-w-0 overflow-hidden rounded-xl border border-sky-400/15 bg-[#111a24]">
      <NuxtLink v-if="group.candidates[0]" :to="titleLink(group.candidates[0].sourceKey, group.candidates[0].externalId)" class="flex min-h-40 text-inherit no-underline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-primary">
        <div class="relative w-28 shrink-0 overflow-hidden bg-[#0c1219]">
          <UIcon name="i-tabler-photo-off" class="absolute left-1/2 top-1/2 size-7 -translate-x-1/2 -translate-y-1/2 text-white/15" />
          <img v-if="group.candidates[0].imageUrl" :src="group.candidates[0].imageUrl" :alt="`${group.title} 海报`" class="absolute inset-0 size-full object-cover" loading="lazy" referrerpolicy="no-referrer" @error="hideBrokenImage">
          <UBadge v-if="group.year" :label="group.year" color="neutral" variant="solid" size="xs" class="absolute right-1.5 top-1.5" />
        </div>
        <div class="flex min-w-0 flex-1 flex-col p-3">
          <h2 class="line-clamp-2 text-sm font-bold leading-5 text-white">{{ group.title }}</h2>
          <div class="mt-2 flex flex-wrap gap-1.5">
            <UBadge v-if="group.category" :label="group.category" color="primary" variant="subtle" size="xs" />
          </div>
          <div class="mt-auto space-y-1.5 pt-3 text-xs text-white/60">
            <p class="flex items-center gap-1.5">
              <UIcon name="i-tabler-clock" />
              {{ group.candidates[0].remarks || (group.candidates[0].episodeCount === 1 ? '正片' : group.candidates[0].episodeCount > 1 ? `${group.candidates[0].episodeCount} 集` : "查看播放选项") }}
            </p>
            <p class="flex items-center gap-1.5"><UIcon name="i-tabler-server" />{{ group.candidates[0].sourceName }}</p>
          </div>
        </div>
      </NuxtLink>
      <details v-if="group.candidates.length > 1" class="border-t border-white/10 px-3 py-2">
        <summary class="cursor-pointer py-1 text-xs text-sky-300 focus-visible:outline-2 focus-visible:outline-sky-300">选择来源 · {{ group.candidates.length }} 个</summary>
        <div class="mt-2 flex flex-wrap gap-2" aria-label="影片来源">
          <UButton v-for="candidate in group.candidates" :key="`${candidate.sourceKey}:${candidate.externalId}`" :label="candidate.sourceName" :to="titleLink(candidate.sourceKey, candidate.externalId)" color="neutral" variant="soft" size="sm" class="max-w-full whitespace-normal text-start" />
        </div>
      </details>
    </article>
  </div>
</template>
