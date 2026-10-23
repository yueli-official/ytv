<script setup lang="ts">
import type { SourceReviewSummary } from "~/types/ytv";
const props = defineProps<{ recommended?: boolean; reviews?: SourceReviewSummary }>();
const badges = computed(() => [
 {label:"广告少", count:props.reviews?.adsFew || 0}, {label:"广告多", count:props.reviews?.adsMany || 0},
 {label:"高清", count:props.reviews?.hd || 0}, {label:"画质一般", count:props.reviews?.normal || 0}, {label:"画质模糊", count:props.reviews?.blurry || 0}
].filter(item => item.count > 0));
</script>

<template>
  <span v-if="recommended || badges.length" class="source-badges">
    <span v-if="recommended" class="source-badge source-badge-recommended"><UIcon name="i-tabler-star-filled" class="size-3" />推荐</span>
    <span v-for="badge in badges" :key="badge.label" class="source-badge" :title="`${badge.count} 位用户评价：${badge.label}`" :aria-label="`${badge.label}，${badge.count} 人`">{{ badge.label }} <span>{{ badge.count }}</span></span>
  </span>
</template>

<style scoped>
.source-badges{display:flex;flex-wrap:wrap;gap:4px;margin-top:5px}
.source-badge{display:inline-flex;align-items:center;gap:3px;padding:2px 6px;border:1px solid var(--ytv-line);border-radius:4px;color:var(--ui-text-muted);font-size:11px;line-height:1.4;font-weight:400}
.source-badge-recommended{color:var(--ui-primary);border-color:currentColor}
</style>

