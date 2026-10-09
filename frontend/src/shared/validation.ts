export type FieldErrors = Record<string, string>;

export class ValidationFailure extends Error {
  readonly fields: FieldErrors;

  constructor(fields: FieldErrors, message = "Some fields are not valid") {
    super(message);
    this.name = "ValidationFailure";
    this.fields = fields;
  }
}
