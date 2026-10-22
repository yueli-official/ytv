import { browserProgress } from "~/utils/viewer-progress";

export default defineNuxtPlugin((nuxtApp) => {
  const records = browserProgress();
  const { loggedIn } = useAuth();
  if (!loggedIn.value) return;
  const api = useYtvApi();
  const cutoff = Date.now() - 30 * 24 * 60 * 60 * 1000;
  nuxtApp.hook("app:mounted", () => { void sync(); });
  async function sync() {
    for (const record of records.filter((item) => item.updatedAt >= cutoff)) {
      await api.saveProgress(record.sourceKey, record.externalId, record.episodeKey, {
        titleName: record.titleName,
        episodeTitle: record.episodeTitle,
        positionSeconds: record.position,
        durationSeconds: record.duration,
        completed: record.completed,
        updatedAt: new Date(record.updatedAt).toISOString()
      }).catch(() => undefined);
    }
  }
});
