import { X, ZoomIn } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ImagePreviewProps {
  src: string | null | undefined
  alt?: string
  className?: string
  fullScreen?: boolean
  onClose?: () => void
  onClick?: () => void
}

export function ImagePreview({ src, alt, className, fullScreen, onClose, onClick }: ImagePreviewProps) {
  if (!src) return null

  if (fullScreen) {
    return (
      <div
        className="fixed inset-0 bg-black/95 flex items-center justify-center z-[100] cursor-zoom-out p-4 backdrop-blur-md animate-zoom-in"
        onClick={onClose}
      >
        <img
          src={src}
          alt={alt}
          className="max-w-full max-h-full object-contain rounded-xl shadow-2xl"
        />
        <button
          className="absolute top-6 right-6 p-3 rounded-full bg-white/10 hover:bg-white/20 text-white transition-colors"
          onClick={(e) => {
            e.stopPropagation()
            onClose?.()
          }}
        >
          <X className="w-6 h-6" />
        </button>
      </div>
    )
  }

  return (
    <div className={cn('relative group overflow-hidden rounded-xl border border-surface-border bg-black/20', className)}>
      <img
        src={src}
        alt={alt}
        onClick={onClick}
        className="w-full h-full object-contain cursor-zoom-in transition-transform group-hover:scale-[1.02]"
      />
      <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
        <button 
          onClick={onClick}
          className="bg-white text-black font-bold px-4 py-2 rounded-lg text-xs flex items-center gap-2 shadow-2xl hover:scale-105 transition-transform"
        >
          <ZoomIn className="w-4 h-4" />
          تكبير الصورة
        </button>
      </div>
    </div>
  )
}
