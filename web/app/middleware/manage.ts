export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === "/manage/setup") return;
  const api = useYtvApi();
  try {
    const setup = await api.setup();
    if (!setup.claimed) return navigateTo("/manage/setup");
    await api.adminSession();
  } catch {
    return navigateTo("/");
  }
});
