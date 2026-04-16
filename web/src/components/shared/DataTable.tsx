import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  type ColumnDef,
} from '@tanstack/react-table'
import { ChevronRight, ChevronLeft } from 'lucide-react'
import { LoadingSpinner } from './LoadingSpinner'
import { EmptyState } from './EmptyState'
import { cn } from '@/lib/utils'

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  isLoading?: boolean
  total?: number
  page?: number
  onPageChange?: (page: number) => void
  emptyTitle?: string
  emptyDescription?: string
}

export function DataTable<TData, TValue>({
  columns,
  data,
  isLoading,
  total = 0,
  page = 1,
  onPageChange,
  emptyTitle,
  emptyDescription,
}: DataTableProps<TData, TValue>) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: Math.ceil(total / 25), // Assuming 25 per page default
  })

  return (
    <div className="admin-card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-right border-collapse text-sm">
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id} className="border-b border-surface-border bg-surface-base/30">
                {headerGroup.headers.map((header) => (
                  <th key={header.id} className="px-6 py-4 text-[10px] font-bold text-surface-muted uppercase tracking-widest">
                    {flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="divide-y divide-surface-border">
            {isLoading ? (
              <tr>
                <td colSpan={columns.length} className="px-6 py-20 text-center">
                  <LoadingSpinner label="جاري تحميل البيانات..." />
                </td>
              </tr>
            ) : data.length === 0 ? (
              <tr>
                <td colSpan={columns.length}>
                  <EmptyState title={emptyTitle} description={emptyDescription} icon={undefined} />
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-surface-border/20 transition-colors group">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-6 py-4">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Container */}
      {total > 0 && (
        <div className="px-6 py-4 bg-surface-base/30 border-t border-surface-border flex items-center justify-between gap-4">
          <div className="text-xs text-surface-muted font-bold">
             عرض صفحة <strong>{page}</strong> من أصل <strong>{Math.ceil(total / 25)}</strong> صفحات
          </div>
          <div className="flex gap-2">
            <button
              disabled={page <= 1}
              onClick={() => onPageChange?.(page - 1)}
              className="p-2 rounded-lg border border-surface-border text-surface-muted hover:text-white
                         hover:bg-surface-border/50 disabled:opacity-20 transition-all"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
            <button
              disabled={page >= Math.ceil(total / 25)}
              onClick={() => onPageChange?.(page + 1)}
              className="p-2 rounded-lg border border-surface-border text-surface-muted hover:text-white
                         hover:bg-surface-border/50 disabled:opacity-20 transition-all"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
