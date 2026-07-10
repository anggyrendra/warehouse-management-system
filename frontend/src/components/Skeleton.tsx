/**
 * Skeleton rows shown while item data is loading. Mimics the shape of the
 * real table so there is no layout shift when data arrives.
 */
export function ItemTableSkeleton({ rows = 6 }: { rows?: number }) {
  return (
    <div className="card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3">SKU</th>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Category</th>
              <th className="px-4 py-3">Unit</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {Array.from({ length: rows }).map((_, i) => (
              <tr key={i}>
                <td className="px-4 py-3"><div className="skeleton h-4 w-20" /></td>
                <td className="px-4 py-3"><div className="skeleton h-4 w-40" /></td>
                <td className="px-4 py-3"><div className="skeleton h-4 w-24" /></td>
                <td className="px-4 py-3"><div className="skeleton h-4 w-16" /></td>
                <td className="px-4 py-3"><div className="skeleton ml-auto h-4 w-24" /></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
