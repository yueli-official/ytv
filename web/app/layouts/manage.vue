<script setup lang="ts">
import { AdminConsoleLayout, type AdminNavigationItem, type AdminShellMessages } from "@yueli/ui/admin";

const route = useRoute();
useHead({ bodyAttrs: { class: "ytv-manage-active" } });
const messages: AdminShellMessages = {
  skipToContent: "跳到主要内容",
  search: "搜索控制台",
  searchPlaceholder: "搜索管理页面",
  currentLocation: "当前位置"
};

const navigation = computed<readonly AdminNavigationItem[]>(() => [
  { label: "控制台", icon: "i-tabler-layout-dashboard", to: "/manage", active: route.path === "/manage" },
  { label: "数据源", icon: "i-tabler-plug-connected", to: "/manage/sources", active: route.path.startsWith("/manage/sources") },
  { label: "站点设置", icon: "i-tabler-adjustments-horizontal", to: "/manage/settings", active: route.path.startsWith("/manage/settings") }
]);

const currentLabel = computed(() => navigation.value.find((item) => item.active)?.label || "控制台");
const { loggedIn: headerLoggedIn } = useAuth();
</script>

<template>
  <AdminConsoleLayout
    class="yueli-admin-branded"
    :navigation="navigation"
    :search-groups="[]"
    :messages="messages"
    storage-key="ytv-admin"
    main-id="ytv-admin-main"
    brand-label="YTV"
    brand-icon="i-tabler-device-tv"
    brand-to="/"
    context-label="YTV"
    :current-label="currentLabel"
    back-to-top-label="返回顶部"
    data-ytv-admin-shell
  >
    <template #topbar-right>
      <ConsumerManageAccountControl :trigger-mode="headerLoggedIn ? 'collapsed' : 'inline'" home-to="" show-appearance />
    </template>
    <slot />
  </AdminConsoleLayout>
</template>
