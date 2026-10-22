<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

const drawerOpen = ref(false);
const route = useRoute();
watch(() => route.fullPath, () => { drawerOpen.value = false; });
const sourceSettingsOpen = ref(false);
const api = useYtvApi();
const { data: siteMeta } = await useAsyncData("ytv-site-meta", () => api.meta());
const siteName = computed(() => siteMeta.value?.name || "YTV");
const accountActions: readonly AccountMenuAction[] = [
  { label: "播放历史", icon: "i-tabler-history", to: "/history" },
  { label: "管理后台", icon: "i-tabler-shield-cog", to: "/manage" }
];

const navigation = [
  { label: "发现", to: "/" },
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
  <div class="ytv-public min-h-dvh text-white">
    <header class="ytv-header sticky top-0 z-30 text-white">
      <div class="ytv-container flex min-h-18 items-center gap-3 sm:gap-6">
        <NuxtLink to="/" class="ytv-wordmark" :aria-label="`${siteName} 首页`"><span class="ytv-brand-icon" aria-hidden="true"><UIcon name="i-tabler-player-play-filled" /></span><span>{{ siteName }}</span></NuxtLink>
        <nav class="hidden items-center gap-5 md:flex" aria-label="主导航">
          <NuxtLink v-for="item in navigation" :key="item.to" :to="item.to" class="ytv-nav-link" :class="{ active: route.path === item.to }">{{ item.label }}</NuxtLink>
        </nav>
        <UButton to="/search" aria-label="搜索" icon="i-tabler-search" color="neutral" variant="ghost" class="ms-auto min-h-11 min-w-11 shrink-0 justify-center" />
        <button type="button" aria-label="来源设置" class="grid size-11 shrink-0 place-items-center rounded-lg text-white/75 transition hover:bg-white/10 hover:text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-300" @click="openSourceSettings"><UIcon name="i-tabler-adjustments-horizontal" class="size-5" /></button>
        <ConsumerAccountControl :context-actions="accountActions" />
        <UButton
          icon="i-tabler-menu-2"
          color="neutral"
          variant="ghost"
          square
          aria-label="打开导航"
          :aria-expanded="drawerOpen"
          class="size-11 md:hidden"
          @click="openDrawer"
        />
      </div>
    </header>
    <main id="main-content"><slot /></main>
    <footer class="border-t border-white/8">
      <div class="ytv-container flex flex-wrap items-center justify-between gap-4 py-8 text-xs text-white/45">
        <p><span class="font-semibold tracking-wide text-white/65">{{ siteName }}</span><span class="mx-3 text-white/15">/</span>电影与故事，在这里相遇。</p>
        <div class="flex items-center gap-5"><NuxtLink to="/history" class="text-inherit no-underline hover:text-white">播放历史</NuxtLink><button class="hover:text-white" @click="openSourceSettings">选择来源</button></div>
      </div>
    </footer>
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
