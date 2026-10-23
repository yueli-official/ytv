<script setup lang="ts">
import { ytvFailureMessage } from "~/utils/ytv-failure";
const props = defineProps<{ sourceKey: string; sourceName: string; total?: number }>();
const api = useYtvApi();
const { loggedIn } = useAuth();
const open = ref(false);
const loading = ref(false);
const saving = ref(false);
const error = ref("");
const saved = ref("");
const loaded = ref(false);
const ads = ref("");
const quality = ref("");
const adsOptions = [{ value: "few", label: "广告少" }, { value: "many", label: "广告多" }];
const qualityOptions = [{ value: "hd", label: "高清" }, { value: "normal", label: "一般" }, { value: "blurry", label: "模糊" }];
async function load() {
  loading.value = true; loaded.value = false; error.value = ""; saved.value = "";
  try { const result = await api.sourceReview(props.sourceKey); ads.value = result.review.ads; quality.value = result.review.quality; loaded.value = true; }
  catch (e) { error.value = ytvFailureMessage(e, "读取你的评价失败，请重试。"); }
  finally { loading.value = false; }
}
watch(open, value => { if (value) void load(); });
async function submit(withdraw = false) {
  saving.value = true; error.value = ""; saved.value = "";
  try {
    const result = await api.saveSourceReview(props.sourceKey, { ads: withdraw ? "" : ads.value, quality: withdraw ? "" : quality.value });
    ads.value = result.review.ads; quality.value = result.review.quality;
    saved.value = withdraw || (!ads.value && !quality.value) ? "已撤回评价" : "已保存评价";
    try { await refreshNuxtData("ytv-public-sources"); } catch { error.value = "评价已保存，统计刷新失败，请刷新页面查看。"; }
  } catch (e) { error.value = ytvFailureMessage(e, "保存评价失败，请重试。"); }
  finally { saving.value = false; }
}
</script>

<template>
  <UModal v-model:open="open" :title="`${sourceName} · 来源评价`" :description="loggedIn ? '按实际观看体验评价，同一账号每个来源计一票，可修改或撤回。' : '无需登录，评价保存在当前访客身份下，可修改或撤回。'">
    <UButton :label="total ? `评价来源 · ${total}` : '评价来源'" :aria-label="`评价 ${sourceName}`" size="sm" variant="outline" color="neutral" icon="i-tabler-message-star" />
    <template #body>
      <div class="space-y-5">
        <p class="text-xs text-muted">评价反映用户体验，不代表每部影片或每条线路的表现。</p>
        <p v-if="loading" role="status" class="text-sm">正在读取你的评价…</p>
        <form v-else-if="loaded" class="space-y-5" @submit.prevent="submit()">
          <fieldset :disabled="saving" @change="saved = ''"><legend class="mb-2 text-sm font-medium">广告体验</legend><div class="flex flex-wrap gap-3"><label v-for="option in adsOptions" :key="option.value" class="flex items-center gap-2 text-sm"><input v-model="ads" class="accent-[var(--ui-primary)]" type="radio" :name="`ads-${sourceKey}`" :value="option.value" />{{ option.label }}</label><label class="flex items-center gap-2 text-sm"><input v-model="ads" class="accent-[var(--ui-primary)]" type="radio" :name="`ads-${sourceKey}`" value="" />不评价</label></div></fieldset>
          <fieldset :disabled="saving" @change="saved = ''"><legend class="mb-2 text-sm font-medium">画质体验</legend><div class="flex flex-wrap gap-3"><label v-for="option in qualityOptions" :key="option.value" class="flex items-center gap-2 text-sm"><input v-model="quality" class="accent-[var(--ui-primary)]" type="radio" :name="`quality-${sourceKey}`" :value="option.value" />{{ option.label }}</label><label class="flex items-center gap-2 text-sm"><input v-model="quality" class="accent-[var(--ui-primary)]" type="radio" :name="`quality-${sourceKey}`" value="" />不评价</label></div></fieldset>
          <div class="flex gap-2"><UButton type="submit" label="保存评价" size="sm" :loading="saving" /><UButton label="撤回评价" size="sm" variant="ghost" color="neutral" :disabled="saving || (!ads && !quality)" @click="submit(true)" /></div>
        </form>
        <p v-if="saved" role="status" class="text-sm text-primary">{{ saved }}</p>
        <div v-if="error" class="space-y-2"><p role="alert" class="text-sm text-error">{{ error }}</p><UButton v-if="!loaded" label="重试" size="sm" variant="outline" @click="load" /></div>
      </div>
    </template>
  </UModal>
</template>

