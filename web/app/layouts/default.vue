<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

const drawerOpen = ref(false);
const sourceSettingsOpen = ref(false);
const api = useYtvApi();
const { data: siteMeta } = await useAsyncData("ytv-site-meta", () => api.meta());
const siteName = computed(() => siteMeta.value?.name || "YTV");
const accountActions: readonly AccountMenuAction[] = [
  { label: "播放历史", icon: "i-tabler-history", to: "/history" },
  { label: "管理后台", icon: "i-tabler-shield-cog", to: "/manage" }
];

const navigation = [
  { label: "播放历史", to: "/history" }
];

function openDrawer() {
  drawerOpen.value = true;
}

function closeDrawer() {
  drawerOpen.value = false;
}
function openSourceSettings() { sourceSettingsOpen.value = true; }
function updateSourceSettingsOpen(value: boolean) { if (!value) sourceSettingsOpen.value = false; }
</script>

<template>
  <div class="min-h-dvh bg-[#070b10] text-white">
    <header class="sticky top-0 z-30 border-b border-white/10 bg-[#06090d]/95 text-white backdrop-blur-md">
      <div class="mx-auto flex min-h-16 max-w-[96rem] items-center gap-3 px-4 sm:px-6 lg:px-8">
        <NuxtLink to="/" class="inline-flex items-center gap-1.5 text-2xl font-black tracking-[-0.04em] text-white no-underline" :aria-label="`${siteName} 首页`"><span>{{ siteName }}</span><span class="size-1.5 rounded-full bg-sky-400" aria-hidden="true" /></NuxtLink>
        <nav class="hidden items-center gap-1 sm:flex" aria-label="主导航">
          <UButton v-for="item in navigation" :key="item.to" :to="item.to" :label="item.label" color="neutral" variant="ghost" />
        </nav>
        <UButton to="/search" label="搜索" icon="i-tabler-search" color="neutral" variant="ghost" class="ms-auto hidden min-h-11 md:inline-flex" />
        <button type="button" aria-label="来源设置" class="grid size-11 shrink-0 place-items-center rounded-lg text-white/75 transition hover:bg-white/10 hover:text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-300" @click="openSourceSettings"><UIcon name="i-tabler-adjustments-horizontal" class="size-5" /></button>
        <ConsumerAccountControl :context-actions="accountActions" />
        <UButton
          icon="i-tabler-menu-2"
          color="neutral"
          variant="ghost"
          square
          aria-label="打开导航"
          :aria-expanded="drawerOpen"
          class="size-11 sm:hidden"
          @click="openDrawer"
        />
      </div>
    </header>
    <main id="main-content"><slot /></main>
    <USlideover :open="sourceSettingsOpen" side="right" title="来源设置" description="选择搜索和热门入口使用的数据源。" :ui="{ content: 'w-full max-w-md bg-[#0c1219] text-white', body: 'p-4 sm:p-5' }" @update:open="updateSourceSettingsOpen">
      <template #body><ViewerSourceSettings /></template>
    </USlideover>

    <USlideover v-model:open="drawerOpen" side="right" title="导航" :ui="{ content: 'max-w-80 bg-[#0c1219] text-white', body: 'p-4' }">
      <template #body>
        <div class="space-y-4">
          <SearchCommand compact />
          <nav class="grid gap-1" aria-label="移动端主导航">
            <UButton v-for="item in navigation" :key="item.to" :to="item.to" :label="item.label" color="neutral" variant="ghost" block class="min-h-11 justify-start" @click="closeDrawer" />
          </nav>
        </div>
      </template>
    </USlideover>
  </div>
</template>
