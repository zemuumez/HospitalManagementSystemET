import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
const read = (name: string) =>
  JSON.parse(readFileSync(new URL(name, import.meta.url), "utf8"));
test("every source-derived screen, field, column and CMS label has Amharic coverage", () => {
  const dictionary = read("am.json");
  const labels = new Set<string>();
  for (const screen of read("legacy-catalog.json"))
    for (const label of [
      screen.title,
      screen.group,
      ...screen.columns,
      ...screen.fields.map((f: { label: string }) => f.label),
    ])
      labels.add(label);
  for (const fields of Object.values(read("front-cms-fields.json")) as {
    label: string;
  }[][])
    for (const f of fields) labels.add(f.label);
  for (const f of read("prescription-fields.json")) labels.add(f.label);
  const missing = [...labels].filter(
    (label) => !dictionary[label] || !/[\u1200-\u137F]/.test(dictionary[label]),
  );
  assert.deepEqual(
    missing,
    [],
    "Missing Amharic labels: " + missing.join(", "),
  );
});
