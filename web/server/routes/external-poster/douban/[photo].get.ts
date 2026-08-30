const allowedPhoto = /^p\d{6,20}\.(?:jpg|webp)$/u;
const maximumPosterBytes = 2 * 1024 * 1024;

export default defineEventHandler(async (event) => {
  const photo = getRouterParam(event, "photo") || "";
  if (!allowedPhoto.test(photo)) {
    throw createError({ statusCode: 404, statusMessage: "Poster not found" });
  }

  const response = await fetch(`https://img1.doubanio.com/view/photo/s_ratio_poster/public/${photo}`, {
    headers: {
      accept: "image/avif,image/webp,image/jpeg,image/*;q=0.8",
      referer: "https://movie.douban.com/",
      "user-agent": "Mozilla/5.0 (compatible; YTV/1.0)"
    },
    redirect: "error",
    signal: AbortSignal.timeout(8_000)
  }).catch(() => null);

  const contentType = response?.headers.get("content-type")?.split(";", 1)[0]?.trim().toLowerCase();
  const declaredLength = Number(response?.headers.get("content-length") || 0);
  if (!response?.ok || !contentType || !["image/jpeg", "image/webp"].includes(contentType) || declaredLength > maximumPosterBytes) {
    throw createError({ statusCode: 502, statusMessage: "Poster unavailable" });
  }

  const body = new Uint8Array(await response.arrayBuffer());
  if (body.byteLength === 0 || body.byteLength > maximumPosterBytes) {
    throw createError({ statusCode: 502, statusMessage: "Poster unavailable" });
  }

  setHeader(event, "content-type", contentType);
  setHeader(event, "content-length", body.byteLength);
  setHeader(event, "cache-control", "public, max-age=1800, stale-while-revalidate=86400");
  setHeader(event, "x-content-type-options", "nosniff");
  return body;
});
