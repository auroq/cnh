---
title: "Accessibility regressions"
description: "Review UI changes against the accessibility checklist"
type: check
when: "A PR changes an HTML, template, CSS, JavaScript or TypeScript file under web/ or cmd/"
actions: "Review against .gitar/review/a11y.md; request changes for any missing label, alt text, focus order, live region or contrast problem, quoting the element"
---

# Accessibility regressions

The UI is screen-reader-first.
Check every changed page or component against `.gitar/review/a11y.md`.
Quote the element at fault and say what a screen reader or keyboard user would experience.
