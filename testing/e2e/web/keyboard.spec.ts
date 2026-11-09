import { expect, test, type Locator, type Page } from "@playwright/test";
import { question } from "../helpers/pages.js";

const stops: { name: string; find: (page: Page) => Locator }[] = [
  { name: "the skip link", find: (p) => p.getByRole("link", { name: "Skip to the answer" }) },
  { name: "the home link", find: (p) => p.getByRole("link", { name: "Wide Door" }) },
  { name: "the question field", find: (p) => p.getByRole("searchbox", { name: "What do you need?" }) },
  { name: "the location field", find: (p) => p.getByRole("textbox", { name: /Where are you\?/ }) },
  { name: "the Ask button", find: (p) => p.getByRole("button", { name: "Ask" }) },
];

async function pressTab(page: Page, times: number) {
  for (let i = 0; i < times; i++) await page.keyboard.press("Tab");
}

test.describe("when I use the keyboard only", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
  });

  test.describe("and I press Tab from the top of the home page", () => {
    stops.forEach((stop, i) => {
      test(`it should reach ${stop.name} on Tab ${i + 1}`, async ({ page }) => {
        await pressTab(page, i + 1);
        await expect(stop.find(page)).toBeFocused();
      });
    });

    test("it should then reach each quick pick in order", async ({ page }) => {
      const picks = page.getByRole("group", { name: "Or pick one" }).getByRole("link");
      const expected = await picks.allTextContents();
      await pressTab(page, stops.length);
      const reached: string[] = [];
      for (let i = 0; i < expected.length; i++) {
        await page.keyboard.press("Tab");
        reached.push(await page.evaluate(() => (document.activeElement?.textContent ?? "").trim()));
      }
      expect(reached).toEqual(expected.map((t) => t.trim()));
    });
  });

  test.describe("and I type a question and press Enter on the Ask button", () => {
    test.beforeEach(async ({ page }) => {
      await page.getByRole("searchbox", { name: "What do you need?" }).fill(question.q);
      await page.getByRole("textbox", { name: /Where are you\?/ }).fill(question.from);
      await page.getByRole("button", { name: "Ask" }).focus();
      await page.keyboard.press("Enter");
      await page.waitForURL(/\/ask\?/);
    });

    test("it should load the answer page", async ({ page }) => {
      await expect(page).toHaveURL(/\/ask\?q=/);
    });

    test("it should put the summary in the live region", async ({ page, request }) => {
      const response = await request.get(`/api/ask?${new URLSearchParams(question)}`);
      const { summary } = (await response.json()) as { summary: string };
      await expect(page.locator("[aria-live='polite']")).toContainText(summary);
    });

    test.describe("and I follow the skip link", () => {
      test("it should land on a heading first", async ({ page }) => {
        const target = await page.getByRole("link", { name: "Skip to the answer" }).getAttribute("href");
        await expect(page.locator(`${target} > *`).first()).toHaveRole("heading");
      });

      test("it should name that heading Answer", async ({ page }) => {
        const target = await page.getByRole("link", { name: "Skip to the answer" }).getAttribute("href");
        await expect(page.locator(`${target} > *`).first()).toHaveText("Answer");
      });
    });
  });
});
