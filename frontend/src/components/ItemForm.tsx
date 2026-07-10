import { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import type { Item, ItemInput } from '@/types';
import { ApiError } from '@/services/api';
import { useCreateItem, useUpdateItem } from '@/hooks/useItems';
import { FieldErrors } from './FieldErrors';

interface ItemFormProps {
  /** When provided, the form is in edit mode; otherwise create mode. */
  initialItem?: Item;
}

interface FormState {
  sku: string;
  name: string;
  category: string;
  unit: string;
}

interface FormErrors {
  sku?: string;
  name?: string;
  category?: string;
  unit?: string;
}

const EMPTY: FormState = { sku: '', name: '', category: '', unit: '' };

/**
 * ItemForm handles both create and edit. Client-side validation runs before
 * submit; the API may return additional field errors which are surfaced below.
 */
export function ItemForm({ initialItem }: ItemFormProps) {
  const isEdit = !!initialItem;
  const navigate = useNavigate();

  const [form, setForm] = useState<FormState>(
    initialItem
      ? {
          sku: initialItem.sku,
          name: initialItem.name,
          category: initialItem.category,
          unit: initialItem.unit,
        }
      : EMPTY,
  );
  const [clientErrors, setClientErrors] = useState<FormErrors>({});
  const [apiErrors, setApiErrors] = useState<{ field: string; reason: string }[] | undefined>();

  const createMutation = useCreateItem();
  const updateMutation = useUpdateItem();

  const isSubmitting = createMutation.isPending || updateMutation.isPending;

  function validate(): boolean {
    const errs: FormErrors = {};
    if (!form.sku.trim()) errs.sku = 'SKU is required';
    if (!form.name.trim()) errs.name = 'Name is required';
    if (!form.category.trim()) errs.category = 'Category is required';
    if (!form.unit.trim()) errs.unit = 'Unit is required';
    setClientErrors(errs);
    return Object.keys(errs).length === 0;
  }

  function handleChange(field: keyof FormState, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
    // Clear the field error as the user types
    if (clientErrors[field]) {
      setClientErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setApiErrors(undefined);
    if (!validate()) return;

    const input: ItemInput = {
      sku: form.sku.trim(),
      name: form.name.trim(),
      category: form.category.trim(),
      unit: form.unit.trim(),
    };

    try {
      if (isEdit && initialItem) {
        await updateMutation.mutateAsync({ id: initialItem.id, input });
        toast.success('Item updated successfully');
      } else {
        await createMutation.mutateAsync(input);
        toast.success('Item created successfully');
      }
      navigate('/');
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Something went wrong';
      toast.error(message);
      if (err instanceof ApiError) {
        setApiErrors(err.errors);
      }
    }
  }

  return (
    <form onSubmit={handleSubmit} className="card max-w-2xl space-y-5 p-6">
      <div>
        <h2 className="text-lg font-bold text-slate-900">
          {isEdit ? 'Edit Item' : 'Add New Item'}
        </h2>
        <p className="text-sm text-slate-500">
          {isEdit ? 'Update the details of this warehouse item.' : 'Create a new warehouse item.'}
        </p>
      </div>

      {/* SKU */}
      <div>
        <label className="label" htmlFor="sku">SKU <span className="text-red-500">*</span></label>
        <input
          id="sku"
          className="input"
          value={form.sku}
          onChange={(e) => handleChange('sku', e.target.value)}
          placeholder="e.g. SKU-001"
          disabled={isSubmitting}
        />
        {clientErrors.sku && <p className="mt-1 text-sm text-red-600">{clientErrors.sku}</p>}
      </div>

      {/* Name */}
      <div>
        <label className="label" htmlFor="name">Name <span className="text-red-500">*</span></label>
        <input
          id="name"
          className="input"
          value={form.name}
          onChange={(e) => handleChange('name', e.target.value)}
          placeholder="e.g. Widget A"
          disabled={isSubmitting}
        />
        {clientErrors.name && <p className="mt-1 text-sm text-red-600">{clientErrors.name}</p>}
      </div>

      {/* Category */}
      <div>
        <label className="label" htmlFor="category">Category <span className="text-red-500">*</span></label>
        <input
          id="category"
          className="input"
          value={form.category}
          onChange={(e) => handleChange('category', e.target.value)}
          placeholder="e.g. Electronics"
          disabled={isSubmitting}
          list="category-suggestions"
        />
        <datalist id="category-suggestions" />
        {clientErrors.category && <p className="mt-1 text-sm text-red-600">{clientErrors.category}</p>}
      </div>

      {/* Unit */}
      <div>
        <label className="label" htmlFor="unit">Unit <span className="text-red-500">*</span></label>
        <input
          id="unit"
          className="input"
          value={form.unit}
          onChange={(e) => handleChange('unit', e.target.value)}
          placeholder="e.g. pcs"
          disabled={isSubmitting}
        />
        {clientErrors.unit && <p className="mt-1 text-sm text-red-600">{clientErrors.unit}</p>}
      </div>

      {apiErrors && <FieldErrors errors={apiErrors} />}

      <div className="flex gap-3 pt-2">
        <button type="submit" className="btn-primary" disabled={isSubmitting}>
          {isSubmitting && (
            <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
            </svg>
          )}
          {isSubmitting ? 'Saving...' : isEdit ? 'Update Item' : 'Create Item'}
        </button>
        <button type="button" className="btn-secondary" onClick={() => navigate('/')} disabled={isSubmitting}>
          Cancel
        </button>
      </div>
    </form>
  );
}
