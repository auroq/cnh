import { defineConfig, devices } from "@playwright/test";
import path from "node:path";

const baseURL = (process.env.WIDEDOOR_URL ?? "").trim().replace(/\/+$/, "");
if (!baseURL) {
  throw new Error(
    "WIDEDOOR_URL is unset; set it to the Wide Door base URL, such as https://widedoor.cnh.auroq.com, to run the e2e tests",
  );
}

process.env.PLAYWRIGHT_BROWSERS_PATH ??= path.join(import.meta.dirname, ".cache", "ms-playwright");

export default defineConfig({
  testDir: "./web",
  outputDir: "./test-results",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: [
    ["list"],
    ["html", { outputFolder: "./playwright-report", open: "never" }],
  ],
  use: {
    baseURL,
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
