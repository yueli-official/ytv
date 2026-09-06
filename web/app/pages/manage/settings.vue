<script setup lang="ts">
import { ytvFailureFeedback } from "~/utils/ytv-failure";
import type { FormSubmitEvent } from "@nuxt/ui";
import { z } from "zod";
import { ManagePage } from "@yueli/ui/admin";
import { ActionFeedbackButton } from "@yueli/ui/feedback/pattern";
import { SettingSection } from "@yueli/ui/settings/pattern";

definePageMeta({ layout: "manage", middleware: ["auth", "manage"] });
useSeoMeta({ title: "站点设置" });

const api = useYtvApi();
const { data, error, refresh } = await useAsyncData("ytv-admin-site-profile", () => api.adminSiteProfile());
const form = reactive({ name: "", description: "", revision: 0 });
const schema = z.object({
  name: z.string().trim().min(1, "请输入站点名称").max(80, "站点名称不能超过 80 个字符"),
  description: z.string().max(300, "站点简介不能超过 300 个字符")
});
const saveFeedback = useActionFeedback();
const saveError = ref("");
const fieldErrors = ref<Readonly<Record<string, readonly string[]>>>({});
const technical = ref<{ code: string; traceId?: string }>();

watch(data, (value) => {
  if (!value?.profile) return;
  form.name = value.profile.name;
  form.description = value.profile.description;
  form.revision = value.profile.revision;
}, { immediate: true });

async function saveProfile(_event?: FormSubmitEvent<z.output<typeof schema>>) {
  saveError.value = "";
  fieldErrors.value = {};
  technical.value = undefined;
  try {
    await saveFeedback.run(async () => {
      const result = await api.updateSiteProfile(form);
      form.revision = result.profile.revision;
    });
  } catch (error) {
    const feedback = ytvFailureFeedback(error, "站点资料没有保存，请稍后重试。", { "/name": "name", "/description": "description" });
    fieldErrors.value = feedback.fieldErrors;
    technical.value = feedback.technical;
    saveError.value = [feedback.message, ...feedback.summary].join(" ");
  }
}
</script>

<template>
  <ManagePage id="ytv-site-settings" title="站点设置" icon="i-tabler-adjustments-horizontal">
    <template #actions>
      <ActionFeedbackButton
        form="ytv-site-profile-form"
        type="submit"
        :status="saveFeedback.status.value"
        idle-label="保存公开资料"
        pending-label="保存中"
        success-label="已保存"
        error-label="保存失败"
        :disabled="!form.name.trim() || saveFeedback.isPending.value"
      />
    </template>

    <SettingSection title="公开资料">
      <UAlert v-if="error" class="mb-4" color="error" variant="subtle" title="站点资料暂时不可用"><template #actions><UButton label="重试" color="error" variant="soft" @click="() => refresh()" /></template></UAlert>
      <UForm id="ytv-site-profile-form" :schema="schema" :state="form" class="grid gap-4" @submit="saveProfile">
        <UFormField label="站点名称" name="name" :error="fieldErrors.name?.[0]"><UInput v-model="form.name" class="w-full" maxlength="80" /></UFormField>
        <UFormField label="站点简介" name="description" :error="fieldErrors.description?.[0]"><UTextarea v-model="form.description" class="w-full" :rows="4" maxlength="300" /></UFormField>
        <p v-if="saveError" class="text-sm text-error" role="alert">{{ saveError }}</p>
      <details v-if="technical" class="text-xs text-muted"><summary>技术详情</summary><code class="break-all select-all">{{ technical.code }} {{ technical.traceId }}</code></details>
      </UForm>
    </SettingSection>
  </ManagePage>
</template>
