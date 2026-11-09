import { expect, test } from "@playwright/test";
import { scan } from "../helpers/axe.js";
import { pages } from "../helpers/pages.js";

const preferences = [
  { name: "a dark colour scheme", media: { colorScheme: "dark" as const } },
  { name: "more contrast", media: { contrast: "more" as const } },
];

for (const preference of preferences) {
  test.describe(`when the reader prefers ${preference.name}`, () => {
    test.beforeEach(async ({ page }) => {
      await page.emulateMedia(preference.media);
    });

    for (const page of pages) {
      test.describe(`and it is the ${page.name} page`, () => {
        test.beforeEach(async ({ page: browser }) => {
          await browser.goto(page.path);
        });

        test("it should pass the axe colour-contrast rule", async ({ page: browser }, testInfo) => {
          const results = await scan(browser, testInfo, ["color-contrast"]);
          expect(results.violations).toEqual([]);
        });
      });
    }
  });
}
