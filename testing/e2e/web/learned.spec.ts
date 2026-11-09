import { expect, test } from "@playwright/test";

const learnedSentence = "Added Step-free only to what Wide Door knows about you, because you asked for an elevator";

test.describe("when I ask for the nearest elevator with no profile", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/ask?q=nearest+elevator");
  });

  test("it should say what it learned in the live region", async ({ page }) => {
    await expect(page.locator("[aria-live='polite']")).toContainText(learnedSentence);
  });

  test("it should name the learned chip's remove button with where it came from", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Remove Step-free only, from your question" })).toBeVisible();
  });

  test("it should offer to keep the chip as my own choice", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Keep Step-free only" })).toBeVisible();
  });

  test.describe("and I remove it with the keyboard only", () => {
    test.beforeEach(async ({ page }) => {
        await page.getByRole("button", { name: "Remove Step-free only, from your question" }).focus();
      await page.keyboard.press("Enter");
      await page.waitForURL(/removed=step-free/);
    });

    test("it should no longer list the chip as on", async ({ page }) => {
      await expect(page.getByRole("button", { name: "Add Step-free only" })).toBeVisible();
    });

    test("it should not add the chip back when I ask again", async ({ page }) => {
      await page.goto("/ask?q=nearest+elevator");
      await expect(page.getByRole("button", { name: "Add Step-free only" })).toBeVisible();
    });
  });
});
