<script setup lang="ts">
import type { SearchGroup } from "~/types/ytv";

defineProps<{ groups: SearchGroup[] }>();

function hideBrokenImage(event: Event) {
  (event.currentTarget as HTMLImageElement).hidden = true;
}
</script>

<template>
  <ul class="ytv-source-grid" tabindex="0" aria-label="来源更新列表，可横向滚动" role="list" data-poster-grid>
    <li v-for="group in groups" :key="group.key" class="min-w-0">
      <NuxtLink v-if="group.candidates[0]" :to="`/title/${encodeURIComponent(group.candidates[0].sourceKey)}/${encodeURIComponent(group.candidates[0].externalId)}`" class="ytv-source-card">
        <div class="ytv-poster-image">
          <UIcon name="i-tabler-photo-off" class="absolute size-6 text-white/30" />
          <img v-if="group.candidates[0].imageUrl" :src="group.candidates[0].imageUrl" :alt="`${group.title} 海报`" class="absolute inset-0 size-full object-cover" loading="lazy" referrerpolicy="no-referrer" @error="hideBrokenImage">
        </div>
        <div class="min-w-0 space-y-2"><h3 class="line-clamp-2 text-base font-semibold leading-6">{{ group.title }}</h3><p class="truncate text-base text-white/70">{{ group.candidates[0].sourceName }}</p></div>
      </NuxtLink>
    </li>
  </ul>
</template>
