<script setup lang="ts">
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

watch(data, (value) => {
  if (!value?.profile) return;
  form.name = value.profile.name;
  form.description = value.profile.description;
  form.revision = value.profile.revision;
}, { immediate: true });

async function saveProfile(_event?: FormSubmitEvent<z.output<typeof schema>>) {
  saveError.value = "";
  try {
    await saveFeedback.run(async () => {
      const result = await api.updateSiteProfile(form);
      form.revision = result.profile.revision;
    });
  } catch {
    saveError.value = "保存失败；站点资料可能已被其他管理员更新，请刷新后重试。";
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
        <UFormField label="站点名称" name="name"><UInput v-model="form.name" class="w-full" maxlength="80" /></UFormField>
        <UFormField label="站点简介" name="description"><UTextarea v-model="form.description" class="w-full" :rows="4" maxlength="300" /></UFormField>
        <p v-if="saveError" class="text-sm text-error" role="alert">{{ saveError }}</p>
      </UForm>
    </SettingSection>
  </ManagePage>
</template>
