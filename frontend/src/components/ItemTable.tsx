import { Link } from 'react-router-dom';
import type { Item } from '@/types';

interface ItemTableProps {
  items: Item[];
  onDelete: (item: Item) => void;
  isDeleting: boolean;
}

/**
 * ItemTable renders the warehouse items in a responsive table.
 * Each row has edit and delete actions. The table scrolls horizontally on
 * narrow screens.
 */
export function ItemTable({ items, onDelete, isDeleting }: ItemTableProps) {
  if (items.length === 0) {
    return (
      <div className="card p-12 text-center">
        <p className="text-slate-500">No items found.</p>
        <p className="mt-1 text-sm text-slate-400">
          Try adjusting your search or filters, or create a new item.
        </p>
      </div>
    );
  }

  return (
    <div className="card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3 font-semibold">SKU</th>
              <th className="px-4 py-3 font-semibold">Name</th>
              <th className="px-4 py-3 font-semibold">Category</th>
              <th className="px-4 py-3 font-semibold">Unit</th>
              <th className="px-4 py-3 text-right font-semibold">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {items.map((item) => (
              <tr key={item.id} className="hover:bg-slate-50">
                <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-slate-700">
                  {item.sku}
                </td>
                <td className="px-4 py-3 font-medium text-slate-900">{item.name}</td>
                <td className="px-4 py-3">
                  <span className="inline-flex rounded-full bg-brand-50 px-2.5 py-0.5 text-xs font-medium text-brand-700">
                    {item.category}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-600">{item.unit}</td>
                <td className="px-4 py-3">
                  <div className="flex justify-end gap-2">
                    <Link to={`/items/${item.id}/edit`} className="btn-secondary !px-3 !py-1.5 text-xs">
                      Edit
                    </Link>
                    <button
                      className="btn-danger !px-3 !py-1.5 text-xs"
                      onClick={() => onDelete(item)}
                      disabled={isDeleting}
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
