const cookieSecure =
  process.env.NUXT_COOKIE_SECURE === undefined
    ? process.env.NODE_ENV === "production"
    : process.env.NUXT_COOKIE_SECURE === "true";

export default defineNuxtConfig({
  vue: { compilerOptions: { isCustomElement: (tag) => ["video-player", "video-neutral-skin", "hlsjs-video", "media-i18n"].includes(tag) } },
  devtools: { enabled: process.env.NUXT_DEVTOOLS === "true" },
  extends: ["@yueli/identity-nuxt"],
  modules: ["@nuxt/ui", "@yueli/ui", "@yueli/nuxt-runtime"],
  yueliRuntime: {
    defaultTarget: "platform",
    targets: {
      platform: {
        path: "/",
        ssr: { cookies: [], headers: ["accept-language", "user-agent"] }
      }
    }
  },
  icon: {
    provider: "none",
    fallbackToApi: false,
    serverBundle: { collections: ["tabler"] },
    clientBundle: {
      scan: {
        globInclude: ["app/**/*.{vue,ts}", "node_modules/@yueli/**/*.{vue,js,mjs,ts}"],
        globExclude: ["test/**", "tests/**", ".*"]
      },
      sizeLimitKb: 256
    }
  },
  css: ["~/assets/css/main.css", "~/assets/css/mobile.css"],
  colorMode: { preference: "dark", fallback: "dark", storageKey: "ytv-color-mode" },
  vite: { resolve: { dedupe: ["vue", "vue-router"] } },
  app: {
    head: {
      link: [
        { rel: "icon", type: "image/x-icon", href: "/favicon.ico?v=20260929" },
        { rel: "icon", type: "image/svg+xml", sizes: "any", href: "/favicon.svg?v=20260929" },
        { rel: "apple-touch-icon", sizes: "180x180", href: "/apple-touch-icon.png?v=20260929" },
      ],
      htmlAttrs: { lang: "zh-CN" },
      titleTemplate: "%s · YTV",
      meta: [
        { name: "theme-color", content: "#0d141b" },
        { property: "og:type", content: "website" },
        { name: "twitter:card", content: "summary_large_image" }
      ]
    }
  },
  fonts: {
    providers: {
      google: false,
      googleicons: false,
      bunny: false,
      fontshare: false,
      fontsource: false
    }
  },
  nitro: {
    esbuild: { options: { target: "es2022", exclude: /node_modules(?!.*(?:@yueli\+|@yueli[\\/]))/ } }
  },
  buildDir: process.env.NUXT_BUILD_DIR || ".nuxt",
  devServer: {
    host: "127.0.0.1",
    port: Number(process.env.NUXT_DEV_PORT || "3013")
  },
  runtimeConfig: {
    downstreamBase: process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8097",
    identityBase: process.env.NUXT_IDENTITY_BASE || "http://127.0.0.1:8081",
    cookieSecure,
    authCookieSecure: cookieSecure,
    guestCookieSecure: cookieSecure,
    guestSessionTtlSeconds: 30 * 24 * 60 * 60,
    ytvAudience: "ytv-api",
    sealSecret: process.env.NUXT_SEAL_SECRET || "dev-ytv-seal-secret-change-me-0123456789abcdef",
    public: {
      oidcIssuer: process.env.NUXT_PUBLIC_OIDC_ISSUER || "http://localhost:3000",
      oidcClientId: process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "ytv-main-web",
      oidcRedirectUri: process.env.NUXT_PUBLIC_OIDC_REDIRECT_URI || "http://localhost:3013/auth/callback",
      oidcPostLogoutRedirectUri: process.env.NUXT_PUBLIC_OIDC_POST_LOGOUT_REDIRECT_URI || "http://localhost:3013/",
      oidcScopes: process.env.NUXT_PUBLIC_OIDC_SCOPES || "openid profile email roles offline_access",
      accountUrl: process.env.NUXT_PUBLIC_ACCOUNT_URL || "http://localhost:3000",
      publicBase: process.env.NUXT_PUBLIC_YTV_BASE || "http://localhost:3013",
      siteBrand: "YTV"
    }
  }
});
