import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import { ItemListPage } from '@/pages/ItemListPage';
import { ItemFormPage } from '@/pages/ItemFormPage';

// Single QueryClient instance with sensible defaults for a dashboard app.
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 30 * 1000,
    },
  },
});

/**
 * App is the root component. It sets up the React Router, TanStack Query
 * provider, toast notifications, and the error boundary.
 */
export function App() {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            {/* Dashboard / item list */}
            <Route path="/" element={<ItemListPage />} />
            {/* Create item */}
            <Route path="/items/new" element={<ItemFormPage />} />
            {/* Edit item */}
            <Route path="/items/:id/edit" element={<ItemFormPage />} />
            {/* Fallback redirect */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
        {/* Toast notifications for success/error actions */}
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              borderRadius: '8px',
              background: '#fff',
              color: '#1e293b',
              border: '1px solid #e2e8f0',
              boxShadow: '0 4px 12px rgba(0,0,0,0.08)',
            },
            success: { iconTheme: { primary: '#16a34a', secondary: '#fff' } },
            error: { iconTheme: { primary: '#dc2626', secondary: '#fff' } },
          }}
        />
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
