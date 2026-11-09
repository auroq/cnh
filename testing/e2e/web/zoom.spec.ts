import { expect, test } from "@playwright/test";
import { pages } from "../helpers/pages.js";

const layouts = [
  { name: "200% zoom", viewport: { width: 640, height: 900 }, deviceScaleFactor: 2 },
  { name: "320px wide", viewport: { width: 320, height: 900 }, deviceScaleFactor: 1 },
];

for (const layout of layouts) {
  test.describe(`when the page is viewed at ${layout.name}`, () => {
    test.use({ viewport: layout.viewport, deviceScaleFactor: layout.deviceScaleFactor });

    for (const page of pages) {
      test.describe(`and it is the ${page.name} page`, () => {
        test.beforeEach(async ({ page: browser }) => {
          await browser.goto(page.path);
        });

        test("it should not scroll horizontally", async ({ page: browser }) => {
          const overflow = await browser.evaluate(
            () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
          );
          expect(overflow).toBeLessThanOrEqual(0);
        });

        test("it should keep the Ask button inside the viewport width", async ({ page: browser }) => {
          const box = await browser.getByRole("button", { name: "Ask" }).boundingBox();
          expect(box && box.x >= 0 && box.x + box.width <= layout.viewport.width).toBe(true);
        });
      });
    }
  });
}
