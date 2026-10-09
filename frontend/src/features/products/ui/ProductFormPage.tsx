import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { Alert, errorMessage } from "../../../shared/ui/components";
import { ValidationFailure, type FieldErrors } from "../../../shared/validation";
import { useProduct, useSaveProduct } from "../application/products";
import { emptyForm, toDraft, toFormValues, type ProductFormValues } from "../domain/product";
import { ProductForm } from "./ProductForm";

export function ProductFormPage() {
  const params = useParams();
  const id = params.id ? Number(params.id) : null;
  const product = useProduct(id);

  if (id !== null && product.isPending) {
    return <p className="muted">Loading product...</p>;
  }
  if (product.error) {
    return <Alert>{errorMessage(product.error)}</Alert>;
  }

  const initial = product.data ? toFormValues(product.data) : emptyForm;
  return <ProductEditor key={id ?? "new"} id={id} initial={initial} />;
}

function ProductEditor({ id, initial }: { id: number | null; initial: ProductFormValues }) {
  const navigate = useNavigate();
  const save = useSaveProduct(id);
  const [values, setValues] = useState(initial);
  const [errors, setErrors] = useState<FieldErrors>({});

  function submit() {
    const { draft, errors } = toDraft(values);
    setErrors(errors);
    if (!draft) {
      return;
    }
    save.mutate(draft, {
      onSuccess: () => navigate("/"),
      onError: (error) => {
        if (error instanceof ValidationFailure) {
          setErrors(error.fields);
        }
      },
    });
  }

  return (
    <section className="narrow">
      <h1>{id === null ? "New product" : "Edit product"}</h1>
      {save.error && !(save.error instanceof ValidationFailure) && <Alert>{errorMessage(save.error)}</Alert>}
      <ProductForm
        values={values}
        errors={errors}
        submitting={save.isPending}
        submitLabel={id === null ? "Create product" : "Save changes"}
        onChange={(field, value) => setValues((v) => ({ ...v, [field]: value }))}
        onSubmit={submit}
        onCancel={() => navigate("/")}
      />
    </section>
  );
}
