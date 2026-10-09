import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createPortContext } from "../../../shared/di/createPortContext";
import type { ImportRepository } from "../domain/importReport";

export const [ImportRepositoryProvider, useImportRepository] = createPortContext<ImportRepository>("ImportRepository");

export function useImportProducts() {
  const repo = useImportRepository();
  const client = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => repo.upload(file),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["products"] });
      client.invalidateQueries({ queryKey: ["categories"] });
    },
  });
}
