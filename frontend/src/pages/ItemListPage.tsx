import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import type { Item } from '@/types';
import { Header } from '@/components/Header';
import { ItemTable } from '@/components/ItemTable';
import { ItemTableSkeleton } from '@/components/Skeleton';
import { Pagination } from '@/components/Pagination';
import { useDebounce } from '@/hooks/useDebounce';
import { useCategories, useDeleteItem, useItems } from '@/hooks/useItems';

const LIMIT = 10;

/**
 * ItemListPage is the main dashboard. It shows a table of items with
 * debounced search (client-side over the current page), a category filter
 * (server-side), and pagination (server-side).
 */
export function ItemListPage() {
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState('');
  const [page, setPage] = useState(1);

  const debouncedSearch = useDebounce(search, 300);

  // Server-side query params (search is applied client-side; see note in service).
  const { data, isLoading, isError, error } = useItems({
    category: category || undefined,
    page,
    limit: LIMIT,
  });

  const { data: categories } = useCategories();
  const deleteMutation = useDeleteItem();

  const items = data?.data ?? [];
  const meta = data?.meta;

  // Client-side search filtering on the fetched page. This keeps the UI
  // responsive while the backend list endpoint only supports category filter.
  const filteredItems = useMemo(() => {
    if (!debouncedSearch.trim()) return items;
    const q = debouncedSearch.toLowerCase();
    return items.filter(
      (it) => it.name.toLowerCase().includes(q) || it.sku.toLowerCase().includes(q),
    );
  }, [items, debouncedSearch]);

  async function handleDelete(item: Item) {
    if (!confirm(`Delete item "${item.name}" (${item.sku})? This is a soft delete.`)) return;
    try {
      await deleteMutation.mutateAsync(item.id);
      toast.success('Item deleted successfully');
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete item';
      toast.error(message);
    }
  }

  // Reset to page 1 when category changes
  function handleCategoryChange(value: string) {
    setCategory(value);
    setPage(1);
  }

  return (
    <div className="min-h-screen">
      <Header
        title="Warehouse Dashboard"
        subtitle="Manage warehouse items and inventory"
        actions={
          <Link to="/items/new" className="btn-primary">
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            Add Item
          </Link>
        }
      />

      <main className="mx-auto max-w-6xl space-y-4 px-4 py-6 sm:px-6">
        {/* Search + Filter bar */}
        <div className="card flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative flex-1 sm:max-w-xs">
            <svg
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
              fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-4.35-4.35M11 19a8 8 0 100-16 8 8 0 000 16z" />
            </svg>
            <input
              type="text"
              className="input pl-9"
              placeholder="Search by name or SKU..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <div className="flex items-center gap-2">
            <label htmlFor="category-filter" className="text-sm text-slate-600">Category:</label>
            <select
              id="category-filter"
              className="input min-w-[150px]"
              value={category}
              onChange={(e) => handleCategoryChange(e.target.value)}
            >
              <option value="">All categories</option>
              {categories?.map((cat) => (
                <option key={cat} value={cat}>{cat}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Error banner */}
        {isError && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            Failed to load items: {error instanceof Error ? error.message : 'Unknown error'}
          </div>
        )}

        {/* Table */}
        {isLoading ? (
          <ItemTableSkeleton />
        ) : (
          <ItemTable
            items={filteredItems}
            onDelete={handleDelete}
            isDeleting={deleteMutation.isPending}
          />
        )}

        {/* Pagination */}
        {meta && !isLoading && debouncedSearch.trim() === '' && (
          <div className="card">
            <Pagination
              page={meta.page}
              limit={meta.limit}
              total={meta.total}
              onPageChange={setPage}
            />
          </div>
        )}
      </main>
    </div>
  );
}
