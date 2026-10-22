<script setup lang="ts">
definePageMeta({ layout: false, middleware: "auth" });
const api = useYtvApi();
const pending = ref(false);
const message = ref("");
const { data, error, refresh } = await useAsyncData("ytv-admin-setup", () => api.setup());
if (data.value?.claimed) await navigateTo("/manage");
async function claim() {
  pending.value = true;
  message.value = "";
  try {
    await api.claim();
    await navigateTo("/manage");
  } catch {
    await refresh();
    message.value = data.value?.claimed ? "站点已完成初始化，请使用管理员账号进入后台。" : "初始化没有完成，请稍后重试。";
  } finally { pending.value = false; }
}
useSeoMeta({ title: "初始化管理权限" });
</script>

<template>
  <main class="grid min-h-dvh place-items-center bg-default p-4">
    <section class="grid w-full max-w-md justify-items-center gap-4 rounded-2xl border border-default bg-elevated p-6 text-center sm:p-8">
      <UIcon name="i-tabler-shield-check" class="size-10 text-primary" />
      <h1 class="text-xl font-bold text-highlighted">初始化 YTV 管理权限</h1>
      <p class="text-sm leading-6 text-muted">当前登录用户将成为本站的首位管理员。初始化仅需完成一次，之后由管理员管理片源和站点设置。</p>
      <UAlert v-if="error || message" color="error" variant="subtle" :description="message || '暂时无法读取站点状态，请稍后重试。'" />
      <UButton v-if="error" label="重新加载" @click="() => { refresh(); }" />
      <UButton v-else label="初始化站点并成为管理员" icon="i-tabler-key" size="lg" :disabled="!data?.canClaim" :loading="pending" @click="claim" />
      <NuxtLink to="/" class="text-sm text-muted hover:text-primary">返回首页</NuxtLink>
    </section>
  </main>
</template>
