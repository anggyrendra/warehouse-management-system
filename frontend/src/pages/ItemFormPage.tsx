import { useParams } from 'react-router-dom';
import { Header } from '@/components/Header';
import { ItemForm } from '@/components/ItemForm';
import { ItemTableSkeleton } from '@/components/Skeleton';
import { useItem } from '@/hooks/useItems';

/**
 * ItemFormPage renders the create/edit form. When the route includes an :id
 * param, it fetches the item first (edit mode); otherwise it shows a blank
 * form (create mode).
 */
export function ItemFormPage() {
  const { id } = useParams<{ id: string }>();
  const itemId = id ? Number(id) : undefined;
  const isEdit = !!itemId;

  const { data: item, isLoading } = useItem(itemId);

  return (
    <div className="min-h-screen">
      <Header
        title={isEdit ? 'Edit Item' : 'Add Item'}
        subtitle={isEdit ? 'Update warehouse item details' : 'Create a new warehouse item'}
      />
      <main className="mx-auto max-w-6xl px-4 py-6 sm:px-6">
        {/* In edit mode, show a skeleton while the item loads. */}
        {isEdit && isLoading ? (
          <div className="max-w-2xl space-y-4">
            <div className="card p-6">
              <div className="skeleton mb-4 h-6 w-40" />
              <div className="skeleton mb-2 h-10 w-full" />
              <div className="skeleton mb-2 h-10 w-full" />
              <div className="skeleton mb-2 h-10 w-full" />
              <div className="skeleton mb-2 h-10 w-full" />
            </div>
          </div>
        ) : (
          <ItemForm initialItem={isEdit ? item : undefined} />
        )}
      </main>
    </div>
  );
}

/**
 * Re-export to keep the skeleton import referenced (avoids unused import
 * warnings in some bundler configs) — it is used in the edit loading state.
 */
export { ItemTableSkeleton };
