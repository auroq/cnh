import { expect, test } from "@playwright/test";
import { scan } from "../helpers/axe.js";

const chips = ["step-free", "avoid-crowds", "quiet-matters", "all-gender", "captions-asl", "wheels", "dietary"];

const heading = /What Wide Door knows about you/;

const widths = [
  { name: "a phone", viewport: { width: 390, height: 844 } },
  { name: "a wide screen", viewport: { width: 1280, height: 800 } },
];

test.describe("when I open the profile page with step-free and wheels in the query", () => {
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
      reached.push(await page.evaluate(() => (document.activeElement as HTMLButtonElement)?.value ?? ""));
      await page.keyboard.press("Tab");
    }
    expect(reached).toEqual(chips);
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

for (const width of widths) {
  test.describe(`when I open the profile page on ${width.name}`, () => {
    test.use({ viewport: width.viewport });

    test.beforeEach(async ({ page }) => {
      await page.goto("/profile?prefs=step-free");
    });

    test("it should show the full profile", async ({ page }) => {
      await expect(page.getByRole("complementary", { name: heading })).toBeVisible();
    });

    test("it should put the profile before the form", async ({ page }) => {
      const order = await page.evaluate(() => {
        const aside = document.getElementById("profile")!;
        const main = document.getElementById("main")!;
        return aside.compareDocumentPosition(main) & Node.DOCUMENT_POSITION_FOLLOWING ? "before" : "after";
      });
      expect(order).toBe("before");
    });
  });
}

test.describe("when JavaScript is off", () => {
  test.use({ javaScriptEnabled: false });

  test.describe("and I remove a chip with its button", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/profile?prefs=step-free,wheels");
      await page.getByRole("button", { name: "Remove Wide doors and ramps" }).click();
      await page.waitForURL(/\/profile\?/);
    });

    test("it should no longer list the chip as on", async ({ page }) => {
      await expect(page.getByRole("button", { name: "Add Wide doors and ramps" })).toBeVisible();
    });

    test("it should count one thing", async ({ page }) => {
      await expect(page.getByRole("heading", { name: heading })).toHaveText("What Wide Door knows about you: 1 thing");
    });

    test("it should say what changed in the live region", async ({ page }) => {
      await expect(page.locator("[aria-live='polite']")).toContainText("Removed Wide doors and ramps");
    });

    test("it should keep the chip out of the next page opened with only the cookie", async ({ page }) => {
      await page.goto("/");
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
      await expect(page.locator(".profile-bar")).toHaveText(/What Wide Door knows about you: nothing yet\./);
    });
  });
});

test.describe("when I open the home page with a profile", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/?prefs=step-free");
  });

  test("it should sum the profile up in one line between the header and the form", async ({ page }) => {
    const order = await page.evaluate(() => {
      const header = document.querySelector("header")!;
      const bar = document.querySelector(".profile-bar")!;
      const main = document.getElementById("main")!;
      const after = (a: Element, b: Element) => (a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0;
      return after(header, bar) && after(bar, main);
    });
    expect(order).toBe(true);
  });

  test("it should keep the full profile after the answer region in reading order", async ({ page }) => {
    const order = await page.evaluate(() => {
      const answer = document.getElementById("answer")!;
      const aside = document.getElementById("profile")!;
      return answer.compareDocumentPosition(aside) & Node.DOCUMENT_POSITION_FOLLOWING ? "after" : "before";
    });
    expect(order).toBe("after");
  });

  test.describe("and the screen is a phone", () => {
    test.use({ viewport: { width: 390, height: 844 } });

    test("it should show the summary bar without scrolling", async ({ page }) => {
      const box = await page.locator(".profile-bar").boundingBox();
      expect(box && box.y >= 0 && box.y + box.height <= 844).toBe(true);
    });

    test("it should not show the full profile to anyone", async ({ page }) => {
      await expect(page.getByRole("complementary", { name: heading })).toHaveCount(0);
    });

    test("it should send the Edit link to the profile page", async ({ page }) => {
      await expect(page.getByRole("link", { name: "Edit your profile" })).toHaveAttribute("href", /^\/profile\?/);
    });

    test("it should reach the Edit link with the keyboard", async ({ page }) => {
      await page.getByRole("link", { name: "Wide Door" }).focus();
      await page.keyboard.press("Tab");
      await expect(page.getByRole("link", { name: "Edit your profile" })).toBeFocused();
    });

    test("it should make the Edit link at least 44 pixels tall and wide", async ({ page }) => {
      const box = await page.getByRole("link", { name: "Edit your profile" }).boundingBox();
      expect(box && box.width >= 44 && box.height >= 44).toBe(true);
    });

    test("it should pass the axe scan", async ({ page }, testInfo) => {
      const results = await scan(page, testInfo);
      expect(results.violations).toEqual([]);
    });
  });

  test.describe("and the screen is wide", () => {
    test.use({ viewport: { width: 1280, height: 800 } });

    test("it should show the full profile without interaction", async ({ page }) => {
      await expect(page.getByRole("complementary", { name: heading })).toBeVisible();
    });

    test("it should put the full profile beside the form", async ({ page }) => {
      const [aside, main] = await Promise.all([page.locator("#profile").boundingBox(), page.locator("#main").boundingBox()]);
      expect(aside && main && aside.x >= main.x + main.width).toBe(true);
    });

    test("it should point the Edit link at the full profile", async ({ page }) => {
      await expect(page.getByRole("link", { name: "Edit your profile" })).toHaveAttribute("href", "#profile");
    });

    test("it should reach the profile's controls after the answer with the keyboard", async ({ page }) => {
      await page.getByRole("button", { name: "Ask" }).focus();
      const picks = await page.getByRole("group", { name: "Or pick one" }).getByRole("link").count();
      for (let i = 0; i <= picks; i++) await page.keyboard.press("Tab");
      await expect(page.getByRole("button", { name: "Remove Step-free only" })).toBeFocused();
    });

    test("it should pass the axe scan", async ({ page }, testInfo) => {
      const results = await scan(page, testInfo);
      expect(results.violations).toEqual([]);
    });
  });
});
