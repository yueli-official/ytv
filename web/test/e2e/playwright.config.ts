import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.YTV_BASE_URL || "http://127.0.0.1:3013";

export default defineConfig({
  testDir: ".",
  outputDir: "results",
  timeout: 45_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure"
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"], viewport: { width: 1440, height: 1000 } } },
    { name: "mobile", use: { ...devices["Pixel 7"], viewport: { width: 390, height: 844 } } },
    {
      name: "desktop-dark",
      testMatch: /public\.spec\.ts/u,
      use: { ...devices["Desktop Chrome"], viewport: { width: 1440, height: 1000 }, colorScheme: "dark" }
    }
  ]
});
