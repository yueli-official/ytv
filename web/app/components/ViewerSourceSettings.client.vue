<script setup lang="ts">
const { sources, selected, ready, saving, loggedIn, toggle, selectAll } = useViewerSourcePreferences();
const filter = ref("全部");
const filters = computed(() => ["全部", "推荐"]);
const visibleSources = computed(() => sources.value.filter(source => filter.value === "全部" || source.recommended).toSorted((a,b) => Number(b.recommended) - Number(a.recommended)));
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between gap-3">
      <p class="text-xs text-white/55">{{ !ready ? "正在读取设置" : loggedIn ? "已同步到账号" : "保存在此浏览器" }}</p>
      <UButton label="全选" color="neutral" variant="ghost" size="xs" :disabled="selected.length === sources.length" @click="selectAll" />
    </div>
    <div v-if="ready" class="flex flex-wrap gap-1.5" aria-label="筛选来源">
      <UButton v-for="tag in filters" :key="tag" :label="tag" size="xs" :color="filter === tag ? 'primary' : 'neutral'" :variant="filter === tag ? 'soft' : 'ghost'" :aria-pressed="filter === tag" @click="() => { filter = tag; }" />
    </div>
    <p v-if="ready && !visibleSources.length" class="text-sm text-white/55" role="status">暂无推荐来源。</p>
    <div v-if="ready && visibleSources.length" class="divide-y divide-white/10 overflow-hidden rounded-xl border border-white/10">
      <div v-for="source in visibleSources" :key="source.key" class="flex min-h-14 items-center justify-between gap-4 px-4 py-3">
        <div class="min-w-0"><span class="break-words text-sm font-medium text-white">{{ source.name }}</span><SourceBadges :recommended="source.recommended" :reviews="source.reviews" /></div>
        <USwitch :model-value="selected.includes(source.key)" :disabled="selected.includes(source.key) && selected.length === 1" :aria-label="`使用来源 ${source.name}`" @update:model-value="toggle(source.key)" />
      </div>
    </div>
    <div v-if="!ready" class="space-y-3"><USkeleton v-for="item in 5" :key="item" class="h-14 rounded-xl bg-white/10" /></div>
    <p v-if="selected.length === 1" class="text-xs text-white/50">至少保留一个来源。</p>
    <p v-if="saving" class="text-xs text-[var(--ui-primary)]" role="status">正在保存…</p>
  </div>
</template>
