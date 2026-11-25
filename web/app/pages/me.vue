<script setup lang="ts">
useSeoMeta({ title: "我的" });
const { loggedIn } = useAuth();
const colorMode = useColorMode();
const appearanceOptions = [
  { label: "跟随系统", value: "system" },
  { label: "浅色", value: "light" },
  { label: "深色", value: "dark" },
];

const api = useYtvApi();
const { data: siteMeta } = await useAsyncData("ytv-site-meta", () => api.meta());
const { data: canManage } = await useAsyncData("ytv-account-management", async () =>
  loggedIn.value ? api.adminSession().then(session => session.allowed).catch(() => false) : false,
  { server: false, watch: [loggedIn] },
);
const actions = [
  { label: "播放历史", icon: "i-tabler-history", to: "/history" },
];
</script>

<template>
  <div class="mx-auto my-5 w-full max-w-xl space-y-4 px-4">
    <ConsumerAccountPanel :site-name="siteMeta?.name || 'YTV'" :can-manage="canManage === true" :context-actions="actions" manage-to="/manage" manage-label="管理后台" />
    <UButton v-if="!loggedIn" to="/history" icon="i-tabler-history" trailing-icon="i-tabler-chevron-right" color="neutral" variant="outline" block class="justify-between">播放历史</UButton>
    <section class="space-y-3 rounded-xl border border-default bg-default p-4" aria-labelledby="ytv-viewer-sources">
      <h2 id="ytv-viewer-sources" class="text-sm font-semibold">来源设置</h2>
      <ViewerSourceSettings />
    </section>
    <UFormField v-if="!loggedIn" label="外观" class="rounded-xl border border-default bg-default p-4">
      <USelect v-model="colorMode.preference" :items="appearanceOptions" class="w-full" />
    </UFormField>
  </div>
</template>
