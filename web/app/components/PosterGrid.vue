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
  <ul class="ytv-poster-grid" role="list" data-poster-grid>
    <li v-for="group in groups" :key="group.key" class="min-w-0">
      <NuxtLink
        v-if="group.candidates[0]"
        :to="`/title/${group.candidates[0].sourceKey}/${group.candidates[0].externalId}`"
        class="ytv-poster-card group"
      >
        <div class="ytv-poster-image">
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
        <div class="space-y-1.5 pt-3">
          <h3 class="truncate text-sm font-semibold text-white group-hover:text-sky-200">{{ group.title }}</h3>
          <p class="truncate text-[11px] text-white/45">{{ group.year || '年份未知' }} · {{ group.category || group.candidates[0].sourceName }}</p>
        </div>
      </NuxtLink>
    </li>
  </ul>
</template>
