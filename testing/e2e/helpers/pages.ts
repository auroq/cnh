export const question = { q: "nearest all-gender restroom", from: "255 BC" };

export const pages = [
  { name: "home", path: "/" },
  { name: "answer", path: `/ask?${new URLSearchParams(question)}` },
] as const;

export type PageUnderTest = (typeof pages)[number];
