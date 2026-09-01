<script setup lang="ts">
import type { SearchGroup } from "~/types/ytv";

defineProps<{ groups: SearchGroup[] }>();

function hideBrokenImage(event: Event) {
  (event.currentTarget as HTMLImageElement).hidden = true;
}
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
    <article v-for="group in groups" :key="group.key" class="min-w-0 overflow-hidden rounded-xl border border-sky-400/15 bg-[#111a24]">
      <NuxtLink v-if="group.candidates[0]" :to="`/title/${group.candidates[0].sourceKey}/${group.candidates[0].externalId}`" class="flex min-h-40 text-inherit no-underline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-primary">
        <div class="relative w-28 shrink-0 overflow-hidden bg-[#0c1219]">
          <UIcon name="i-tabler-photo-off" class="absolute left-1/2 top-1/2 size-7 -translate-x-1/2 -translate-y-1/2 text-white/15" />
          <img v-if="group.candidates[0].imageUrl" :src="group.candidates[0].imageUrl" :alt="`${group.title} 海报`" class="absolute inset-0 size-full object-cover" loading="lazy" referrerpolicy="no-referrer" @error="hideBrokenImage">
          <UBadge v-if="group.year" :label="group.year" color="neutral" variant="solid" size="xs" class="absolute right-1.5 top-1.5" />
        </div>
        <div class="flex min-w-0 flex-1 flex-col p-3">
          <h2 class="line-clamp-2 text-sm font-bold leading-5 text-white">{{ group.title }}</h2>
          <div class="mt-2 flex flex-wrap gap-1.5">
            <UBadge v-if="group.category" :label="group.category" color="primary" variant="subtle" size="xs" />
            <UBadge v-if="group.candidates[0].remarks" :label="group.candidates[0].remarks" color="error" variant="subtle" size="xs" />
          </div>
          <div class="mt-auto space-y-1.5 pt-3 text-xs text-white/60">
            <p class="flex items-center gap-1.5">
              <UIcon name="i-tabler-clock" />
              {{ group.candidates[0].remarks || (group.candidates[0].episodeCount > 0 ? `${group.candidates[0].episodeCount} 集` : "进入详情查看选集") }}
            </p>
            <p class="flex items-center gap-1.5"><UIcon name="i-tabler-server" />{{ group.candidates[0].sourceName }}</p>
          </div>
        </div>
      </NuxtLink>
      <div v-if="group.candidates.length > 1" class="flex gap-1.5 overflow-x-auto border-t border-white/5 px-3 py-2" aria-label="其他来源">
        <UButton v-for="candidate in group.candidates.slice(1, 4)" :key="`${candidate.sourceKey}:${candidate.externalId}`" :label="candidate.sourceName" :to="`/title/${candidate.sourceKey}/${candidate.externalId}`" color="neutral" variant="ghost" size="xs" />
      </div>
    </article>
  </div>
</template>
