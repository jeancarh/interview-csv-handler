import { describe, expect, it } from "vitest";
import { defaultFilter, pageOf, rowsFor, type ImportReport, type ImportRow } from "./importReport";

function row(line: number, status: ImportRow["status"], messages: string[] = []): ImportRow {
  return { line, sku: `SKU-${line}`, name: `Row ${line}`, status, messages };
}

const report: ImportReport = {
  summary: { total: 4, created: 2, updated: 0, unchanged: 0, skipped: 1, rejected: 1 },
  rows: [row(2, "created"), row(3, "created", ["removed currency symbol from price"]), row(4, "skipped"), row(5, "rejected")],
};

describe("import report", () => {
  it("keeps only rows someone should look at", () => {
    expect(rowsFor(report, "attention").map((r) => r.line)).toEqual([3, 4, 5]);
  });

  it("opens on the attention view only when there is something to see", () => {
    expect(defaultFilter(report)).toBe("attention");
    expect(defaultFilter({ ...report, rows: [row(2, "created")] })).toBe("all");
  });

  it("pages large reports", () => {
    const many = Array.from({ length: 120 }, (_, i) => i);
    expect(pageOf(many, 1)).toHaveLength(50);
    expect(pageOf(many, 3)).toEqual(many.slice(100));
  });
});
