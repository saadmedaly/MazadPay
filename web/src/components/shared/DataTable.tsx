import {
  useReactTable,
  getCoreRowModel,
  flexRender,
  type ColumnDef,
} from '@tanstack/react-table'
import { ChevronRight, ChevronLeft, Inbox } from 'lucide-react'
import { LoadingSpinner } from './LoadingSpinner'
import { EmptyState } from './EmptyState'

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  isLoading?: boolean
  total?: number
  page?: number
  onPageChange?: (page: number) => void
  emptyTitle?: string
  emptyDescription?: string
  emptyMessage?: string
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
  const totalPages = Math.ceil(total / 25)

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: totalPages,
  })

  return (
    <div className="admin-card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-right text-sm" style={{ borderCollapse: 'separate', borderSpacing: 0 }}>
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    className="px-5 py-3.5 text-[10px] font-black uppercase tracking-[0.1em] whitespace-nowrap"
                    style={{
                      color: '#2D4560',
                      background: 'rgba(255,255,255,0.02)',
                      borderBottom: '1px solid rgba(255,255,255,0.05)',
                    }}
                  >
                    {flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {isLoading ? (
              <tr>
                <td colSpan={columns.length} className="px-5 py-16 text-center">
                  <LoadingSpinner label="جاري تحميل البيانات..." />
                </td>
              </tr>
            ) : data.length === 0 ? (
              <tr>
                <td colSpan={columns.length}>
                  <EmptyState title={emptyTitle} description={emptyDescription} icon={Inbox} />
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  className="table-row-hover"
                >
                  {row.getVisibleCells().map((cell) => (
                    <td
                      key={cell.id}
                      className="px-5 py-3.5 whitespace-nowrap"
                      style={{ color: '#8BA3C0' }}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {total > 0 && (
        <div
          className="flex items-center justify-between px-5 py-3"
          style={{
            borderTop: '1px solid rgba(255,255,255,0.04)',
            background: 'rgba(255,255,255,0.01)',
          }}
        >
          <p className="text-[11px] font-semibold" style={{ color: '#2D4560' }}>
            صفحة{' '}
            <span style={{ color: '#60A5FA' }}>{page}</span>
            {' '}من{' '}
            <span style={{ color: '#60A5FA' }}>{totalPages}</span>
            {total > 0 && (
              <span style={{ color: '#1A2E45' }}> · {total} عنصر</span>
            )}
          </p>
          <div className="flex gap-1.5">
            {[
              { disabled: page <= 1, onClick: () => onPageChange?.(page - 1), Icon: ChevronRight },
              { disabled: page >= totalPages, onClick: () => onPageChange?.(page + 1), Icon: ChevronLeft },
            ].map(({ disabled, onClick, Icon }, i) => (
              <button
                key={i}
                disabled={disabled}
                onClick={onClick}
                className="w-8 h-8 flex items-center justify-center rounded-lg transition-all duration-150 disabled:opacity-20"
                style={{
                  border: '1px solid rgba(255,255,255,0.07)',
                  background: 'rgba(255,255,255,0.04)',
                  color: '#4A6080',
                }}
                onMouseEnter={e => {
                  if (!disabled) {
                    (e.currentTarget as HTMLElement).style.background = 'rgba(59,130,246,0.12)'
                    ;(e.currentTarget as HTMLElement).style.color = '#60A5FA'
                    ;(e.currentTarget as HTMLElement).style.borderColor = 'rgba(59,130,246,0.3)'
                  }
                }}
                onMouseLeave={e => {
                  (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.04)'
                  ;(e.currentTarget as HTMLElement).style.color = '#4A6080'
                  ;(e.currentTarget as HTMLElement).style.borderColor = 'rgba(255,255,255,0.07)'
                }}
              >
                <Icon className="w-3.5 h-3.5" />
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
