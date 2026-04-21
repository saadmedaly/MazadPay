import React from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  className?: string
}

export function Pagination({ currentPage, totalPages, onPageChange, className }: PaginationProps) {
  const pages = Array.from({ length: totalPages }, (_, i) => i + 1)
  const maxVisiblePages = 5
  const halfVisible = Math.floor(maxVisiblePages / 2)
  
  let startPage = Math.max(1, currentPage - halfVisible)
  let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1)
  
  if (endPage - startPage < maxVisiblePages - 1) {
    startPage = Math.max(1, endPage - maxVisiblePages + 1)
  }

  const visiblePages = pages.slice(startPage - 1, endPage)

  return (
    <div className={cn("flex items-center justify-center gap-2", className)}>
      {/* Previous Button */}
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage === 1}
        className={cn(
          "p-2 rounded-lg border border-surface-border transition-all",
          currentPage === 1 
            ? "opacity-50 cursor-not-allowed" 
            : "text-surface-muted hover:text-white hover:border-surface-muted"
        )}
      >
        <ChevronRight className="w-4 h-4" />
      </button>

      {/* Page Numbers */}
      {startPage > 1 && (
        <button
          onClick={() => onPageChange(startPage - 1)}
          className="px-3 py-2 text-sm font-bold text-surface-muted hover:text-white transition-all"
        >
          ...
        </button>
      )}

      {visiblePages.map((page) => (
        <button
          key={page}
          onClick={() => onPageChange(page)}
          className={cn(
            "px-3 py-2 text-sm font-bold rounded-lg border transition-all",
            currentPage === page
              ? "bg-mazad-primary text-white border-mazad-primary"
              : "text-surface-muted border-surface-border hover:text-white hover:border-surface-muted"
          )}
        >
          {page}
        </button>
      ))}

      {endPage < totalPages && (
        <button
          onClick={() => onPageChange(endPage + 1)}
          className="px-3 py-2 text-sm font-bold text-surface-muted hover:text-white transition-all"
        >
          ...
        </button>
      )}

      {/* Next Button */}
      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        className={cn(
          "p-2 rounded-lg border border-surface-border transition-all",
          currentPage === totalPages 
            ? "opacity-50 cursor-not-allowed" 
            : "text-surface-muted hover:text-white hover:border-surface-muted"
        )}
      >
        <ChevronLeft className="w-4 h-4" />
      </button>

      {/* Page Info */}
      <span className="text-xs text-surface-muted">
        الصفحة {currentPage} من {totalPages}
      </span>
    </div>
  )
}
