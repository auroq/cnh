import { expect, test } from "@playwright/test";
import { scan } from "../helpers/axe.js";
import { pages } from "../helpers/pages.js";

test.describe("when axe scans the page against WCAG 2.2 AA", () => {
  for (const page of pages) {
    test.describe(`and it is the ${page.name} page`, () => {
      test.beforeEach(async ({ page: browser }) => {
        await browser.goto(page.path);
      });

      test("it should report no violations", async ({ page: browser }, testInfo) => {
        const results = await scan(browser, testInfo);
        expect(results.violations).toEqual([]);
      });
    });
  }
});
