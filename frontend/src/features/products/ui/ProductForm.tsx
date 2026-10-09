import type { ChangeEvent, FormEvent } from "react";
import { Field } from "../../../shared/ui/components";
import type { FieldErrors } from "../../../shared/validation";
import type { ProductFormValues } from "../domain/product";

interface ProductFormProps {
  values: ProductFormValues;
  errors: FieldErrors;
  submitting: boolean;
  submitLabel: string;
  onChange: (field: keyof ProductFormValues, value: string) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

export function ProductForm({ values, errors, submitting, submitLabel, onChange, onSubmit, onCancel }: ProductFormProps) {
  function submit(e: FormEvent) {
    e.preventDefault();
    onSubmit();
  }

  const field = (name: keyof ProductFormValues) => ({
    name,
    value: values[name],
    error: errors[name],
    onChange: (e: ChangeEvent<HTMLInputElement>) => onChange(name, e.target.value),
  });

  return (
    <form className="form" onSubmit={submit} noValidate>
      <Field label="Name" required {...field("name")} />
      <Field label="SKU" required {...field("sku")} hint="Letters, digits, dashes or underscores" />
      <Field label="Category" required {...field("category")} />
      <div className="field">
        <label htmlFor="field-description" className="field-label">
          Description
        </label>
        <textarea
          id="field-description"
          name="description"
          rows={3}
          value={values.description}
          onChange={(e) => onChange("description", e.target.value)}
        />
      </div>
      <div className="row">
        <Field label="Price (USD)" inputMode="decimal" required {...field("price")} />
        <Field label="Stock" inputMode="numeric" required {...field("stock")} />
        <Field label="Weight (kg)" inputMode="decimal" {...field("weightKg")} hint="Leave empty for digital goods" />
      </div>
      <div className="actions">
        <button type="button" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className="primary" disabled={submitting}>
          {submitting ? "Saving..." : submitLabel}
        </button>
      </div>
    </form>
  );
}
