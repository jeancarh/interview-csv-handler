import { Pagination } from "../../../shared/ui/components";
import {
  pageOf,
  reportPageSize,
  rowStatuses,
  rowsFor,
  type ImportReport,
  type ReportFilter,
} from "../domain/importReport";

interface ImportReportViewProps {
  report: ImportReport;
  filter: ReportFilter;
  page: number;
  onFilterChange: (filter: ReportFilter) => void;
  onPageChange: (page: number) => void;
}

export function ImportReportView({ report, filter, page, onFilterChange, onPageChange }: ImportReportViewProps) {
  const matching = rowsFor(report, filter);
  const rows = pageOf(matching, page);
  const attention = rowsFor(report, "attention").length;

  function chip(value: ReportFilter, label: string, count: number, tone = "") {
    const className = ["chip", tone, filter === value ? "active" : ""].filter(Boolean).join(" ");
    return (
      <button key={value} type="button" className={className} onClick={() => onFilterChange(value)}>
        {label} <strong>{count}</strong>
      </button>
    );
  }

  return (
    <div className="report">
      <div className="summary" role="group" aria-label="Filter rows by status">
        {chip("attention", "Needs attention", attention)}
        {chip("all", "All", report.summary.total)}
        {rowStatuses.map((status) => chip(status, status, report.summary[status], `badge-${status}`))}
      </div>

      {matching.length === 0 ? (
        <p className="empty">No rows in this group.</p>
      ) : (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th className="num">Line</th>
                  <th>SKU</th>
                  <th>Name</th>
                  <th>Status</th>
                  <th>Notes</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((r) => (
                  <tr key={r.line}>
                    <td className="num mono">{r.line}</td>
                    <td className="mono">{r.sku}</td>
                    <td>{r.name || <span className="muted">no name</span>}</td>
                    <td>
                      <span className={`badge badge-${r.status}`}>{r.status}</span>
                    </td>
                    <td className="small">{r.messages.join("; ")}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pagination page={page} pageSize={reportPageSize} total={matching.length} onChange={onPageChange} />
        </>
      )}
    </div>
  );
}
