import { useState, type FormEvent } from "react";
import { Field } from "../../../shared/ui/components";
import { formatCents } from "../../../shared/money/money";

interface CheckoutFormProps {
  totalCents: number;
  submitting: boolean;
  cardError?: string;
  onPay: (cardNumber: string) => void;
}

export function CheckoutForm({ totalCents, submitting, cardError, onPay }: CheckoutFormProps) {
  const [cardNumber, setCardNumber] = useState("");

  function submit(e: FormEvent) {
    e.preventDefault();
    onPay(cardNumber);
  }

  return (
    <form className="form checkout" onSubmit={submit} noValidate>
      <h2>Payment</h2>
      <p className="muted small">
        This is a fake payment. Use 4242 4242 4242 4242 to approve or 4000 0000 0000 0002 to see a decline.
      </p>
      <Field
        label="Card number"
        name="cardNumber"
        inputMode="numeric"
        autoComplete="off"
        placeholder="4242 4242 4242 4242"
        value={cardNumber}
        error={cardError}
        onChange={(e) => setCardNumber(e.target.value)}
      />
      <button type="submit" className="primary" disabled={submitting}>
        {submitting ? "Processing..." : `Pay ${formatCents(totalCents)}`}
      </button>
    </form>
  );
}
