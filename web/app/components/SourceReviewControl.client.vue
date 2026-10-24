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
const adsOptions = [{ value: "few", label: "广告少" }, { value: "many", label: "广告多" }];
const qualityOptions = [{ value: "hd", label: "高清" }, { value: "normal", label: "一般" }, { value: "blurry", label: "模糊" }];
async function load() {
  loading.value = true; loaded.value = false; error.value = ""; saved.value = "";
  try { const result = await api.sourceReview(props.sourceKey); ads.value = result.review.ads; quality.value = result.review.quality; loaded.value = true; }
  catch (e) { error.value = ytvFailureMessage(e, "读取你的评价失败，请重试。"); }
  finally { loading.value = false; }
}
onMounted(() => { void load(); });
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
  <section class="source-vote" aria-label="评价当前来源">
    <div class="source-vote-intro">
      <span class="source-vote-symbol" aria-hidden="true"><UIcon name="i-tabler-message-star" class="size-5" /></span>
      <div><h3>这个来源看起来怎么样？</h3><p>{{ sourceName }}<span class="mx-2">·</span>{{ total ? `${total} 人已评价` : '等你投下第一票' }}</p></div>
    </div>
    <p class="source-vote-hint">{{ loggedIn ? '说说实际观看感受，帮大家选好来源。' : '无需登录，说说实际观看感受，帮大家选好来源。' }}</p>
    <p v-if="loading" role="status" class="text-sm text-muted">正在读取你的评价…</p>
    <form v-else-if="loaded" class="source-vote-form" @submit.prevent="submit()">
      <div class="source-vote-groups">
        <fieldset :disabled="saving" @change="saved = ''"><legend>画质怎么样</legend><div class="source-vote-options"><label v-for="option in qualityOptions" :key="option.value" class="source-vote-option"><input v-model="quality" type="radio" :name="`quality-${sourceKey}`" :value="option.value" /><span>{{ option.label }}</span></label><label class="source-vote-option"><input v-model="quality" type="radio" :name="`quality-${sourceKey}`" value="" /><span>暂不评价</span></label></div></fieldset>
        <fieldset :disabled="saving" @change="saved = ''"><legend>广告多不多</legend><div class="source-vote-options"><label v-for="option in adsOptions" :key="option.value" class="source-vote-option"><input v-model="ads" type="radio" :name="`ads-${sourceKey}`" :value="option.value" /><span>{{ option.label }}</span></label><label class="source-vote-option"><input v-model="ads" type="radio" :name="`ads-${sourceKey}`" value="" /><span>暂不评价</span></label></div></fieldset>
      </div>
      <div class="source-vote-actions"><UButton type="submit" label="提交评价" icon="i-tabler-check" class="ytv-primary" :disabled="!ads && !quality" :loading="saving" /><UButton v-if="ads || quality" label="撤回评价" size="sm" variant="ghost" color="neutral" :disabled="saving" @click="submit(true)" /><span class="text-xs text-muted">每个来源计一票，可修改或撤回</span></div>
    </form>
    <p v-if="saved" role="status" class="text-sm text-primary">{{ saved }}</p>
    <div v-if="error" class="space-y-2"><p role="alert" class="text-sm text-error">{{ error }}</p><UButton v-if="!loaded" label="重试" size="sm" variant="outline" @click="load" /></div>
  </section>
</template>

<style scoped>
.source-vote{display:grid;gap:16px;padding:22px;border:1px solid #d9b86c38;border-radius:12px;background:linear-gradient(115deg,#d9b86c0b,transparent 60%),var(--ytv-panel)}
.source-vote-intro{display:flex;align-items:center;gap:12px}
.source-vote-symbol{display:grid;place-items:center;width:38px;height:38px;border-radius:10px;background:var(--ytv-selected);color:var(--ui-primary);flex-shrink:0}
.source-vote-intro h3{font-size:17px;font-weight:600;color:var(--ui-text-highlighted)}
.source-vote-intro p,.source-vote-hint{font-size:13px;color:var(--ui-text-muted)}
.source-vote-form{display:grid;gap:20px}
.source-vote-groups{display:flex;flex-wrap:wrap;gap:22px 48px}
.source-vote-groups legend{font-size:13px;font-weight:500;margin-bottom:10px}
.source-vote-options{display:flex;flex-wrap:wrap;gap:8px}
.source-vote-option{position:relative;cursor:pointer}
.source-vote-option input{position:absolute;width:1px;height:1px;opacity:0}
.source-vote-option span{display:block;padding:7px 14px;border:1px solid var(--ui-border);border-radius:7px;background:var(--ui-bg);font-size:13px;color:var(--ui-text-muted)}
.source-vote-option input:checked:not([value=""])+span{border-color:var(--ui-primary);background:var(--ytv-selected);color:var(--ui-primary)}
.source-vote-option input:focus-visible+span{outline:2px solid var(--ui-primary);outline-offset:3px}
.source-vote-option:hover span{border-color:var(--ui-primary)}
.source-vote-option input:disabled+span{opacity:.6;cursor:wait}
.source-vote-actions{display:flex;align-items:center;flex-wrap:wrap;gap:12px}
@media(max-width:767px){.source-vote{padding:16px;gap:12px}.source-vote-intro h3{font-size:15px}.source-vote-groups{gap:16px}.source-vote-option span{padding:7px 11px}.source-vote-actions{gap:8px}.source-vote-actions>span{flex-basis:100%}}
</style>
