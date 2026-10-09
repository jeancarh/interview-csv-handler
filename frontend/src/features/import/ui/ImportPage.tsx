import { useState, type FormEvent } from "react";
import { Alert, errorMessage } from "../../../shared/ui/components";
import { useImportProducts } from "../application/importProducts";
import { checkFile, defaultFilter, type ReportFilter } from "../domain/importReport";
import { ImportReportView } from "./ImportReportView";

export function ImportPage() {
  const upload = useImportProducts();
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const [filter, setFilter] = useState<ReportFilter>("attention");
  const [page, setPage] = useState(1);

  function submit(e: FormEvent) {
    e.preventDefault();
    if (!file) {
      setFileError("Pick a file first");
      return;
    }
    const problem = checkFile(file);
    setFileError(problem);
    if (!problem) {
      upload.mutate(file, {
        onSuccess: (report) => {
          setFilter(defaultFilter(report));
          setPage(1);
        },
      });
    }
  }

  return (
    <section>
      <h1>Import products</h1>
      <p className="muted">
        Upload a CSV with the columns <code>name, sku, description, category, price, stock, weight_kg</code>. Existing
        SKUs are updated (stock is left as is), new ones are created, and every row gets a status below.
      </p>

      <form className="upload" onSubmit={submit}>
        <input
          type="file"
          accept=".csv,text/csv"
          aria-label="CSV file"
          onChange={(e) => {
            setFile(e.target.files?.[0] ?? null);
            setFileError(null);
          }}
        />
        <button type="submit" className="primary" disabled={upload.isPending}>
          {upload.isPending ? "Importing..." : "Import"}
        </button>
      </form>

      {fileError && <Alert>{fileError}</Alert>}
      {upload.error && <Alert>{errorMessage(upload.error)}</Alert>}
      {upload.data && (
        <ImportReportView
          report={upload.data}
          filter={filter}
          page={page}
          onFilterChange={(next) => {
            setFilter(next);
            setPage(1);
          }}
          onPageChange={setPage}
        />
      )}
    </section>
  );
}
