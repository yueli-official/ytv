<script setup lang="ts">
import { MobileBottomNav, type MobileBottomNavItem } from "@yueli/ui/navigation/mobile-bottom-nav";

const route = useRoute();
const mobileNavigation = computed<MobileBottomNavItem[]>(() => [
  { id: "home", label: "首页", icon: "i-tabler-home", to: "/", active: route.path === "/" },
  { id: "browse", label: "搜索", icon: "i-tabler-search", to: "/search", active: route.path === "/search" },
  { id: "account", label: "我的", icon: "i-tabler-user", to: "/me", active: route.path === "/me" || route.path === "/history" },
]);
</script>
<template>
  <MobileBottomNav :items="mobileNavigation" label="移动主导航" />
</template>
