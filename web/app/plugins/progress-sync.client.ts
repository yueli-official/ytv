import { loadLocalProgress, migrateLegacyProgress } from "~/utils/local-progress";

export default defineNuxtPlugin(async () => {
  migrateLegacyProgress(localStorage);
  const { loggedIn } = useAuth();
  if (!loggedIn.value) return;
  const api = useYtvApi();
  const cutoff = Date.now() - 30 * 24 * 60 * 60 * 1000;
  for (const record of loadLocalProgress(localStorage).filter((item) => item.updatedAt >= cutoff)) {
    await api.saveProgress(record.sourceKey, record.externalId, record.episodeKey, {
      titleName: record.titleName,
      episodeTitle: record.episodeTitle,
      positionSeconds: record.position,
      durationSeconds: record.duration,
      completed: record.completed,
      updatedAt: new Date(record.updatedAt).toISOString()
    }).catch(() => undefined);
  }
});
