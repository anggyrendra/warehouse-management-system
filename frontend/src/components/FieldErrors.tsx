import type { FieldError } from '@/types';

interface FieldErrorsProps {
  /** Field-specific errors from the API (undefined → empty). */
  errors?: FieldError[];
}

/**
 * FieldErrors renders a list of API field-level validation errors below a form.
 * It maps error fields to a readable bullet list.
 */
export function FieldErrors({ errors }: FieldErrorsProps) {
  if (!errors || errors.length === 0) return null;

  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-3">
      <ul className="list-inside list-disc space-y-1 text-sm text-red-700">
        {errors.map((e, i) => (
          <li key={i}>
            <span className="font-medium">{e.field}</span>: {e.reason}
          </li>
        ))}
      </ul>
    </div>
  );
}
