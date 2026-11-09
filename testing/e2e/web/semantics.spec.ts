import { expect, test } from "@playwright/test";
import { pages } from "../helpers/pages.js";

for (const page of pages) {
  test.describe(`when the ${page.name} page is read by assistive technology`, () => {
    test.beforeEach(async ({ page: browser }) => {
      await browser.goto(page.path);
    });

    test("it should have exactly one h1", async ({ page: browser }) => {
      await expect(browser.locator("h1")).toHaveCount(1);
    });

    test("it should declare its language", async ({ page: browser }) => {
      await expect(browser.locator("html")).toHaveAttribute("lang", /^[a-z]{2}/);
    });

    test.describe("and it lists the quick picks", () => {
      test("it should give every quick pick an accessible name equal to its visible text", async ({ page: browser }) => {
        const picks = browser.getByRole("group", { name: "Or pick one" }).getByRole("link");
        const names = await picks.evaluateAll((links) =>
          links.map((link) => ({
            visible: (link.textContent ?? "").trim(),
            accessible: (link.getAttribute("aria-label") ?? link.textContent ?? "").trim(),
          })),
        );
        expect(names.map((n) => n.accessible)).toEqual(names.map((n) => n.visible));
      });
    });
  });
}
