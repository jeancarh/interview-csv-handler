import type { components } from "../../../shared/api/schema";
import { request } from "../../../shared/http/client";
import type { ImportReport, ImportRepository } from "../domain/importReport";

type ImportReportDto = components["schemas"]["ImportReport"];

export const httpImportRepository: ImportRepository = {
  async upload(file) {
    const form = new FormData();
    form.append("file", file);
    const { data } = await request<ImportReportDto>("/api/v1/products/import", { method: "POST", body: form });
    const report: ImportReport = data;
    return report;
  },
};
