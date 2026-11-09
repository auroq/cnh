import { expect, test, type Page } from "@playwright/test";
import { scan } from "../helpers/axe.js";

const chips = [
  "Step-free only",
  "Avoid crowds and noise",
  "Quiet spaces matter",
  "All-gender restrooms",
  "Captions or ASL",
  "I use a wheelchair or scooter",
  "Dietary needs, unspecified",
];

const heading = /What Wide Door knows about you/;

async function openPanel(page: Page) {
  const panel = page.locator("details.profile");
  if (!(await panel.getAttribute("open"))) await panel.locator("summary").click();
  return panel;
}

test.describe("when I open the profile panel with step-free and wheels in the query", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/profile?prefs=step-free,wheels");
  });

  test("it should say how many things it knows in one line", async ({ page }) => {
    await expect(page.getByRole("heading", { name: heading })).toHaveText("What Wide Door knows about you: 2 things");
  });

  test("it should name the chips that are on as removable", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Remove Step-free only" })).toBeVisible();
  });

  test("it should name the chips that are off as addable", async ({ page }) => {
    await expect(page.getByRole("button", { name: "Add Quiet spaces matter" })).toBeVisible();
  });

  test("it should reach every chip with the keyboard", async ({ page }) => {
    await page.getByRole("button", { name: "Remove Step-free only" }).focus();
    const reached: string[] = [];
    for (let i = 0; i < chips.length; i++) {
      reached.push(await page.evaluate(() => (document.activeElement as HTMLElement)?.innerText.trim() ?? ""));
      await page.keyboard.press("Tab");
    }
    expect(reached.map((t) => t.replace(/^✓\s*/, ""))).toEqual(chips);
  });

  test("it should give every chip a visible focus ring", async ({ page }) => {
    const chip = page.getByRole("button", { name: "Add Quiet spaces matter" });
    await chip.focus();
    const outline = await chip.evaluate((el) => getComputedStyle(el).outlineStyle);
    expect(outline).not.toBe("none");
  });

  test("it should make every chip at least 44 pixels tall and wide", async ({ page }) => {
    const boxes = await page.locator(".chips button").evaluateAll((els) => els.map((el) => el.getBoundingClientRect()));
    expect(boxes.every((b) => b.width >= 44 && b.height >= 44)).toBe(true);
  });

  test("it should pass the axe scan", async ({ page }, testInfo) => {
    const results = await scan(page, testInfo);
    expect(results.violations).toEqual([]);
  });
});

test.describe("when JavaScript is off", () => {
  test.use({ javaScriptEnabled: false });

  test.describe("and I remove a chip with its button", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/profile?prefs=step-free,wheels");
      await page.getByRole("button", { name: "Remove I use a wheelchair or scooter" }).click();
      await page.waitForURL(/\/profile\?/);
    });

    test("it should no longer list the chip as on", async ({ page }) => {
      await expect(page.getByRole("button", { name: "Add I use a wheelchair or scooter" })).toBeVisible();
    });

    test("it should count one thing", async ({ page }) => {
      await expect(page.getByRole("heading", { name: heading })).toHaveText("What Wide Door knows about you: 1 thing");
    });

    test("it should say what changed in the live region", async ({ page }) => {
      await expect(page.locator("[aria-live='polite']")).toContainText("Removed I use a wheelchair or scooter");
    });

    test("it should keep the chip out of the next page opened with only the cookie", async ({ page }) => {
      await page.goto("/");
      await openPanel(page);
      await expect(page.getByRole("button", { name: "Remove Step-free only" })).toBeVisible();
    });
  });

  test.describe("and I ask to be forgotten", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/profile?prefs=step-free");
      await page.getByRole("button", { name: "Forget me" }).click();
      await page.waitForURL(/forgot=1/);
    });

    test("it should know nothing", async ({ page }) => {
      await expect(page.getByRole("heading", { name: heading })).toHaveText("What Wide Door knows about you: nothing yet");
    });

    test("it should say so in the live region", async ({ page }) => {
      await expect(page.locator("[aria-live='polite']")).toContainText("Forgotten");
    });

    test("it should stay forgotten on the next page", async ({ page }) => {
      await page.goto("/");
      await expect(page.getByRole("heading", { name: heading })).toHaveText("What Wide Door knows about you: nothing yet");
    });
  });
});

test.describe("when I open the home page with a profile", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/?prefs=step-free");
  });

  test("it should keep the panel closed so the chips are not read out on every page", async ({ page }) => {
    await expect(page.locator("details.profile")).not.toHaveAttribute("open", "");
  });

  test("it should keep the panel after the answer region in the tab order", async ({ page }) => {
    const order = await page.evaluate(() => {
      const answer = document.getElementById("answer")!;
      const panel = document.querySelector("details.profile")!;
      return answer.compareDocumentPosition(panel) & Node.DOCUMENT_POSITION_FOLLOWING ? "after" : "before";
    });
    expect(order).toBe("after");
  });
});
