<script setup lang="ts">
import { ytvFailureMessage } from "~/utils/ytv-failure";
const props = defineProps<{ sourceKey: string; sourceName: string; total?: number }>();
const api = useYtvApi();
const { loggedIn } = useAuth();
const loading = ref(false);
const saving = ref(false);
const error = ref("");
const saved = ref("");
const loaded = ref(false);
const ads = ref("");
const quality = ref("");
const storedReview = ref({ ads: "", quality: "" });
const editing = ref(true);
const adsOptions = [{ value: "few", label: "广告少" }, { value: "many", label: "广告多" }];
const qualityOptions = [{ value: "hd", label: "高清" }, { value: "normal", label: "一般" }, { value: "blurry", label: "模糊" }];
const hasReview = computed(() => Boolean(storedReview.value.ads || storedReview.value.quality));
const canSubmit = computed(() => Boolean(ads.value || quality.value) && (ads.value !== storedReview.value.ads || quality.value !== storedReview.value.quality));
const storedLabels = computed(() => [qualityOptions.find(option => option.value === storedReview.value.quality)?.label, adsOptions.find(option => option.value === storedReview.value.ads)?.label].filter(Boolean));
function resetDraft() { ads.value = storedReview.value.ads; quality.value = storedReview.value.quality; }
function startEditing() { resetDraft(); editing.value = true; error.value = ""; saved.value = ""; }
function cancelEditing() { resetDraft(); editing.value = false; error.value = ""; saved.value = ""; }
function chooseQuality(value: string) { quality.value = quality.value === value ? "" : value; saved.value = ""; }
function chooseAds(value: string) { ads.value = ads.value === value ? "" : value; saved.value = ""; }
async function load() {
  loading.value = true; loaded.value = false; error.value = ""; saved.value = "";
  try { const result = await api.sourceReview(props.sourceKey); storedReview.value = { ads: result.review.ads, quality: result.review.quality }; resetDraft(); editing.value = !hasReview.value; loaded.value = true; }
  catch (e) { error.value = ytvFailureMessage(e, "读取你的评价失败，请重试。"); }
  finally { loading.value = false; }
}
onMounted(() => { void load(); });
async function submit() {
  if (saving.value || !canSubmit.value) return;
  saving.value = true; error.value = ""; saved.value = "";
  try {
    const result = await api.saveSourceReview(props.sourceKey, { ads: ads.value, quality: quality.value });
    storedReview.value = { ads: result.review.ads, quality: result.review.quality }; resetDraft(); editing.value = false;
    saved.value = "已保存评价";
    try { await refreshNuxtData("ytv-public-sources"); } catch { error.value = "评价已保存，统计刷新失败，请刷新页面查看。"; }
  } catch (e) { error.value = ytvFailureMessage(e, "保存评价失败，请重试。"); }
  finally { saving.value = false; }
}
</script>

<template>
  <section class="source-vote" aria-label="评价当前来源">
    <div class="source-vote-intro">
      <span class="source-vote-symbol" aria-hidden="true"><UIcon name="i-tabler-message-star" class="size-5" /></span>
      <div><h3>{{ loaded && !editing ? '你的来源评价' : '这个来源看起来怎么样？' }}</h3><p>{{ sourceName }}<span class="mx-2">·</span>{{ total ? `${total} 人已评价` : '等你投下第一票' }}</p></div>
    </div>
    <p v-if="loaded && editing" class="source-vote-hint">{{ loggedIn ? '' : '无需登录，' }}可只评价一项。再点已选项可取消选择。</p>
    <p v-if="loading" role="status" class="text-sm text-muted">正在读取你的评价…</p>
    <div v-else-if="loaded && !editing" class="source-vote-saved">
      <div class="source-vote-summary" aria-label="已保存的评价"><span v-for="label in storedLabels" :key="label">{{ label }}</span></div>
      <UButton type="button" label="重新评价" icon="i-tabler-edit" size="sm" variant="ghost" color="neutral" @click="startEditing" />
    </div>
    <form v-else-if="loaded" class="source-vote-form" @submit.prevent="submit">
      <div class="source-vote-groups">
        <fieldset :disabled="saving"><legend>画质怎么样</legend><div class="source-vote-options"><button v-for="option in qualityOptions" :key="option.value" type="button" class="source-vote-option" :aria-pressed="quality === option.value" @click="chooseQuality(option.value)">{{ option.label }}</button></div></fieldset>
        <fieldset :disabled="saving"><legend>广告多不多</legend><div class="source-vote-options"><button v-for="option in adsOptions" :key="option.value" type="button" class="source-vote-option" :aria-pressed="ads === option.value" @click="chooseAds(option.value)">{{ option.label }}</button></div></fieldset>
      </div>
      <div class="source-vote-actions"><UButton type="submit" :label="hasReview ? '更新评价' : '提交评价'" icon="i-tabler-check" class="ytv-primary" :disabled="!canSubmit" :loading="saving" /><UButton v-if="hasReview" type="button" label="取消" size="sm" variant="ghost" color="neutral" :disabled="saving" @click="cancelEditing" /></div>
    </form>
    <p v-if="saved" role="status" class="sr-only">{{ saved }}</p>
    <div v-if="error" class="space-y-2"><p role="alert" class="text-sm text-error">{{ error }}</p><UButton v-if="!loaded" label="重试" size="sm" variant="outline" @click="load" /></div>
  </section>
</template>

<style scoped>
.source-vote{display:grid;gap:16px;padding:22px;border:1px solid color-mix(in srgb,var(--ui-primary) 22%,transparent);border-radius:12px;background:linear-gradient(115deg,color-mix(in srgb,var(--ui-primary) 4%,transparent),transparent 60%),var(--ytv-panel)}
.source-vote-intro{display:flex;align-items:center;gap:12px}
.source-vote-symbol{display:grid;place-items:center;width:38px;height:38px;border-radius:10px;background:var(--ytv-selected);color:var(--ui-primary);flex-shrink:0}
.source-vote-intro h3{font-size:17px;font-weight:600;color:var(--ui-text-highlighted)}
.source-vote-intro p,.source-vote-hint{font-size:13px;color:var(--ui-text-muted)}
.source-vote-form{display:grid;gap:20px}
.source-vote-groups{display:flex;flex-wrap:wrap;gap:22px 48px}
.source-vote-groups legend{font-size:13px;font-weight:500;margin-bottom:10px}
.source-vote-options{display:flex;flex-wrap:wrap;gap:8px}
.source-vote-option{padding:7px 14px;min-height:36px;border:1px solid var(--ui-border);border-radius:7px;background:var(--ui-bg);font-size:13px;color:var(--ui-text-muted);cursor:pointer}
.source-vote-option[aria-pressed="true"]{border-color:var(--ui-primary);background:var(--ytv-selected);color:var(--ui-primary)}
.source-vote-option:focus-visible{outline:2px solid var(--ui-primary);outline-offset:3px}
.source-vote-option:hover{border-color:var(--ui-primary)}
.source-vote-option:disabled{opacity:.6;cursor:wait}
.source-vote-saved{display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:12px}
.source-vote-summary{display:flex;flex-wrap:wrap;gap:8px}
.source-vote-summary>span{padding:7px 12px;border:1px solid color-mix(in srgb,var(--ui-primary) 22%,transparent);border-radius:7px;background:var(--ytv-selected);color:var(--ui-primary);font-size:13px}
.source-vote-actions{display:flex;align-items:center;flex-wrap:wrap;gap:12px}
@media(max-width:767px){.source-vote{padding:16px;gap:12px}.source-vote-intro h3{font-size:15px}.source-vote-groups{gap:16px}.source-vote-option{padding:7px 11px}.source-vote-actions{gap:8px}}
</style>
