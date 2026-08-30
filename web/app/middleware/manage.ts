export default defineNuxtRouteMiddleware(async () => {
  const api = useYtvApi();
  try {
    await api.adminSession();
  } catch {
    return navigateTo("/");
  }
});
