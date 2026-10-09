export type RowStatus = "created" | "updated" | "unchanged" | "skipped" | "rejected";

export interface ImportRow {
  line: number;
  sku: string;
  name: string;
  status: RowStatus;
  messages: string[];
}

export interface ImportReport {
  summary: Record<RowStatus | "total", number>;
  rows: ImportRow[];
}

export interface ImportRepository {
  upload(file: File): Promise<ImportReport>;
}

export const rowStatuses: RowStatus[] = ["created", "updated", "unchanged", "skipped", "rejected"];

export const maxFileBytes = 5 * 1024 * 1024;

export function checkFile(file: File): string | null {
  if (!file.name.toLowerCase().endsWith(".csv")) {
    return "Pick a .csv file";
  }
  if (file.size === 0) {
    return "The file is empty";
  }
  if (file.size > maxFileBytes) {
    return "The file is larger than 5 MB";
  }
  return null;
}

export type ReportFilter = RowStatus | "all" | "attention";

export const reportPageSize = 50;

export function needsAttention(row: ImportRow): boolean {
  return row.status === "rejected" || row.status === "skipped" || row.messages.length > 0;
}

export function rowsFor(report: ImportReport, filter: ReportFilter): ImportRow[] {
  switch (filter) {
    case "all":
      return report.rows;
    case "attention":
      return report.rows.filter(needsAttention);
    default:
      return report.rows.filter((r) => r.status === filter);
  }
}

export function defaultFilter(report: ImportReport): ReportFilter {
  return report.rows.some(needsAttention) ? "attention" : "all";
}

export function pageOf<T>(items: T[], page: number, size = reportPageSize): T[] {
  return items.slice((page - 1) * size, page * size);
}
