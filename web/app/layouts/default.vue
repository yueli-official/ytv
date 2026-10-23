<script setup lang="ts">
import type { AccountMenuAction } from "@yueli/ui/account-menu/pattern";

const drawerOpen = ref(false);
const route = useRoute();
watch(() => route.fullPath, () => { drawerOpen.value = false; });
const sourceSettingsOpen = useState("ytv-source-settings-open", () => false);
const api = useYtvApi();
const { data: siteMeta } = await useAsyncData("ytv-site-meta", () => api.meta());
const siteName = computed(() => siteMeta.value?.name || "YTV");
const accountActions: readonly AccountMenuAction[] = [
  { label: "播放历史", icon: "i-tabler-history", to: "/history" },
  { label: "管理后台", icon: "i-tabler-shield-cog", to: "/manage" }
];

const navigation = [
  { label: "发现", to: "/", icon: "i-tabler-compass" },
  { label: "播放历史", to: "/history", icon: "i-tabler-history" }
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
    <a href="#page-content" class="ytv-skip">跳到内容</a>
    <main id="main-content" class="ytv-main">
      <header class="ytv-header">
        <div class="ytv-container ytv-topbar">
          <NuxtLink to="/" class="ytv-wordmark" :aria-label="`${siteName} 首页`">{{ siteName }}</NuxtLink>
          <nav class="ytv-nav" aria-label="主导航">
            <NuxtLink v-for="item in navigation" :key="item.to" :to="item.to" :aria-current="route.path === item.to ? 'page' : undefined" class="ytv-nav-link" :class="{ active: route.path === item.to }">{{ item.label }}</NuxtLink>
          </nav>
          <div class="ytv-top-search"><SearchCommand compact history-popover :show-submit="route.path === '/search'" :initial-value="route.path === '/search' ? String(route.query.q || '') : ''" :initial-sources="route.path === '/search' ? String(route.query.sources || '').split(',').filter(Boolean) : undefined" /></div>
          <div class="ytv-top-actions">
            <button type="button" aria-label="来源设置" class="ytv-source-trigger" @click="openSourceSettings"><UIcon name="i-tabler-adjustments-horizontal" class="size-5" /><span>来源设置</span></button>
            <ConsumerAccountControl :context-actions="accountActions" />
            <UButton icon="i-tabler-menu-2" color="neutral" variant="ghost" square aria-label="打开导航" :aria-expanded="drawerOpen" class="ytv-mobile-menu size-10" @click="openDrawer" />
          </div>
        </div>
      </header>
      <div id="page-content" tabindex="-1"><slot /></div>
    </main>
    <footer class="ytv-footer border-t border-white/8">
      <div class="ytv-container flex flex-wrap items-center justify-between gap-4 py-8 text-xs text-white/45">
        <p><span class="font-semibold tracking-wide text-white/65">{{ siteName }}</span><span class="mx-3 text-white/15">/</span>电影与故事，在这里相遇。</p>
        <div class="flex items-center gap-5"><NuxtLink to="/history" class="text-inherit no-underline hover:text-white">播放历史</NuxtLink><button class="hover:text-white" @click="openSourceSettings">选择来源</button></div>
      </div>
    </footer>
    <USlideover :open="sourceSettingsOpen" side="right" title="来源设置" description="选择搜索和热门入口使用的数据源。" :ui="{ content: 'ytv-public w-full max-w-md bg-[#141e28] text-white', body: 'p-4 sm:p-5' }" @update:open="updateSourceSettingsOpen">
      <template #body><ViewerSourceSettings /></template>
    </USlideover>

    <USlideover v-model:open="drawerOpen" side="right" title="导航" :ui="{ content: 'ytv-public max-w-80 bg-[#141e28] text-white', body: 'p-4' }">
      <template #body>
        <div class="space-y-4">
          <SearchCommand compact :initial-value="route.path === '/search' ? String(route.query.q || '') : ''" :initial-sources="route.path === '/search' ? String(route.query.sources || '').split(',').filter(Boolean) : undefined" />
          <nav class="grid gap-1" aria-label="移动端主导航">
            <UButton v-for="item in navigation" :key="item.to" :to="item.to" :label="item.label" color="neutral" variant="ghost" block class="min-h-11 justify-start" @click="closeDrawer" />
          </nav>
        </div>
      </template>
    </USlideover>
  </div>
</template>
