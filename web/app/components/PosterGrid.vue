<script setup lang="ts">
import type { SearchGroup } from "~/types/ytv";

defineProps<{ groups: SearchGroup[] }>();

function score(value?: string) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed.toFixed(1) : "";
}

function hideBrokenImage(event: Event) {
  (event.currentTarget as HTMLImageElement).hidden = true;
}
</script>

<template>
  <ul class="grid grid-cols-2 gap-x-3 gap-y-6 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6" role="list" data-poster-grid>
    <li v-for="group in groups" :key="group.key" class="min-w-0">
      <NuxtLink
        v-if="group.candidates[0]"
        :to="`/title/${group.candidates[0].sourceKey}/${group.candidates[0].externalId}`"
        class="group block text-inherit no-underline focus-visible:rounded-xl focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-sky-300"
      >
        <div class="relative aspect-[2/3] overflow-hidden rounded-xl border border-white/10 bg-[#111a24] transition group-hover:border-sky-300/40">
          <UIcon name="i-tabler-photo-off" class="absolute left-1/2 top-1/2 size-9 -translate-x-1/2 -translate-y-1/2 text-white/20" />
          <img
            v-if="group.candidates[0].imageUrl"
            :src="group.candidates[0].imageUrl"
            :alt="`${group.title} 海报`"
            class="absolute inset-0 size-full object-cover transition-transform duration-300 group-hover:scale-[1.035] motion-reduce:transition-none"
            loading="lazy"
            referrerpolicy="no-referrer"
            @error="hideBrokenImage"
          >
          <span v-if="score(group.candidates[0].score)" class="absolute right-2 top-2 rounded-full bg-black/80 px-2 py-1 text-xs font-bold text-amber-400 shadow-sm">
            ★ {{ score(group.candidates[0].score) }}
          </span>
          <span v-if="group.candidates[0].remarks" class="absolute inset-x-0 bottom-0 truncate bg-gradient-to-t from-black/90 to-transparent px-3 pb-2 pt-8 text-xs text-white/85">
            {{ group.candidates[0].remarks }}
          </span>
        </div>
        <div class="space-y-1 py-2.5">
          <h3 class="truncate text-sm font-semibold text-white group-hover:text-sky-200">{{ group.title }}</h3>
          <p class="truncate text-xs text-white/60">{{ group.year || '年份未知' }} · {{ group.category || group.candidates[0].sourceName }}</p>
        </div>
      </NuxtLink>
    </li>
  </ul>
</template>
