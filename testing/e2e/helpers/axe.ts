import AxeBuilder from "@axe-core/playwright";
import type { Page, TestInfo } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";

export const wcag22aa = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

export async function scan(page: Page, testInfo: TestInfo, rules?: string[]) {
  const builder = new AxeBuilder({ page }).withTags(wcag22aa);
  if (rules) builder.withRules(rules);
  const results = await builder.analyze();

  const dir = path.join(testInfo.project.outputDir, "axe");
  fs.mkdirSync(dir, { recursive: true });
  const slug = testInfo.titlePath
    .slice(1)
    .join(" ")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
  const file = path.join(dir, `${slug}.json`);
  fs.writeFileSync(file, JSON.stringify(results, null, 2));
  await testInfo.attach("axe report", { path: file, contentType: "application/json" });

  return results;
}
