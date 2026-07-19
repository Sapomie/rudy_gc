import assert from "node:assert/strict";
import { test } from "vitest";
import { paginationItems } from "./models";

test("pagination model preserves links, current item, disabled item and ellipsis", () => {
  const items = paginationItems({
    Links: [
      { Label: "Prev", Href: "/list?p=1", Disabled: true },
      { Label: "2", Href: "/list?p=2", Active: true },
      { Ellipsis: true },
      { Label: "Next", Href: "/list?p=3" }
    ]
  });
  assert.equal(items[0].href, undefined);
  assert.equal(items[1].active, true);
  assert.equal(items[2].label, "...");
  assert.equal(items[2].disabled, true);
  assert.equal(items[3].href, "/list?p=3");
});
