import { useState, useRef, useCallback } from 'react'
import { Upload, X, Image as ImageIcon, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { ImagePreview } from './ImagePreview'

interface ImageUploadProps {
  images: File[]
  previewUrls: string[]
  onImagesChange: (images: File[], previewUrls: string[]) => void
  maxImages?: number
  maxSizeMB?: number
  className?: string
  disabled?: boolean
  uploading?: boolean
}

export function ImageUpload({
  images,
  previewUrls,
  onImagesChange,
  maxImages = 10,
  maxSizeMB = 10,
  className,
  disabled = false,
  uploading = false,
}: ImageUploadProps) {
  const [dragActive, setDragActive] = useState(false)
  const [previewFullscreen, setPreviewFullscreen] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const validateFile = (file: File): string | null => {
    // Check file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
    if (!allowedTypes.includes(file.type)) {
      return `Le fichier ${file.name} n'est pas une image valide (JPEG, PNG, WebP uniquement)`
    }

    // Check file size
    const maxSize = maxSizeMB * 1024 * 1024
    if (file.size > maxSize) {
      return `Le fichier ${file.name} dépasse la taille maximale de ${maxSizeMB}MB`
    }

    return null
  }

  const handleFiles = useCallback(
    (files: FileList | null) => {
      if (!files || disabled || uploading) return

      const newImages: File[] = []
      const newPreviewUrls: string[] = []
      const errors: string[] = []

      Array.from(files).forEach((file) => {
        // Check if adding this file would exceed max images
        if (images.length + newImages.length >= maxImages) {
          errors.push(`Maximum ${maxImages} images autorisées`)
          return
        }

        // Validate file
        const error = validateFile(file)
        if (error) {
          errors.push(error)
          return
        }

        // Create preview URL
        const previewUrl = URL.createObjectURL(file)
        newImages.push(file)
        newPreviewUrls.push(previewUrl)
      })

      if (errors.length > 0) {
         console.error('Image upload errors:', errors)
        alert(errors.join('\n'))
      }

      if (newImages.length > 0) {
        onImagesChange([...images, ...newImages], [...previewUrls, ...newPreviewUrls])
      }
    },
    [images, previewUrls, onImagesChange, maxImages, maxSizeMB, disabled, uploading]
  )

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()

    if (disabled || uploading) return

    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }, [disabled, uploading])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      e.stopPropagation()
      setDragActive(false)

      if (disabled || uploading) return

      handleFiles(e.dataTransfer.files)
    },
    [handleFiles, disabled, uploading]
  )

  const removeImage = useCallback(
    (index: number) => {
      if (disabled || uploading) return

      // Revoke object URL to prevent memory leaks
      URL.revokeObjectURL(previewUrls[index])

      const newImages = images.filter((_, i) => i !== index)
      const newPreviewUrls = previewUrls.filter((_, i) => i !== index)
      onImagesChange(newImages, newPreviewUrls)
    },
    [images, previewUrls, onImagesChange, disabled, uploading]
  )

  const openFilePicker = () => {
    if (disabled || uploading) return
    inputRef.current?.click()
  }

  return (
    <div className={cn('space-y-4', className)}>
      {/* Upload Zone */}
      <div
        onClick={openFilePicker}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
        className={cn(
          'relative border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-all',
          'hover:border-mazad-primary hover:bg-mazad-primary/5',
          dragActive && 'border-mazad-primary bg-mazad-primary/10',
          disabled && 'opacity-50 cursor-not-allowed',
          uploading && 'opacity-50 cursor-not-allowed'
        )}
      >
        <input
          ref={inputRef}
          type="file"
          multiple
          accept="image/jpeg,image/jpg,image/png,image/webp"
          onChange={(e) => handleFiles(e.target.files)}
          className="hidden"
          disabled={disabled || uploading}
        />

        {uploading ? (
          <div className="flex flex-col items-center gap-3">
            <Loader2 className="w-10 h-10 text-mazad-primary animate-spin" />
            <p className="text-surface-muted">Téléchargement en cours...</p>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-3">
            <div className="p-4 rounded-full bg-mazad-primary/10">
              <Upload className="w-8 h-8 text-mazad-primary" />
            </div>
            <div>
              <p className="font-medium text-white">Glissez-déposez des images ici</p>
              <p className="text-sm text-surface-muted mt-1">ou cliquez pour sélectionner</p>
            </div>
            <div className="text-xs text-surface-muted mt-2">
              <p>Maximum {maxImages} images • Max {maxSizeMB}MB par image</p>
              <p>JPEG, PNG, WebP uniquement</p>
            </div>
          </div>
        )}
      </div>

      {/* Preview Grid */}
      {previewUrls.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {previewUrls.map((url, index) => (
            <div
              key={`${url}-${index}`}
              className="relative aspect-square group"
            >
              <ImagePreview
                src={url}
                alt={`Preview ${index + 1}`}
                className="w-full h-full"
                onClick={() => setPreviewFullscreen(url)}
              />

              {/* Remove button */}
              {!disabled && !uploading && (
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    removeImage(index)
                  }}
                  className="absolute -top-2 -right-2 p-1.5 rounded-full bg-red-500 text-white shadow-lg opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-600"
                >
                  <X className="w-4 h-4" />
                </button>
              )}

              {/* Image number */}
              <div className="absolute top-2 left-2 px-2 py-1 rounded-md bg-black/60 text-white text-xs font-mono">
                {index + 1}
              </div>
            </div>
          ))}

          {/* Add more button */}
          {!disabled && !uploading && images.length < maxImages && (
            <button
              onClick={openFilePicker}
              className="aspect-square rounded-xl border-2 border-dashed border-surface-border flex flex-col items-center justify-center gap-2 hover:border-mazad-primary hover:bg-mazad-primary/5 transition-all"
            >
              <ImageIcon className="w-8 h-8 text-surface-muted" />
              <span className="text-xs text-surface-muted">Ajouter</span>
            </button>
          )}
        </div>
      )}

      {/* Fullscreen Preview */}
      {previewFullscreen && (
        <ImagePreview
          src={previewFullscreen}
          alt="Fullscreen preview"
          fullScreen
          onClose={() => setPreviewFullscreen(null)}
        />
      )}

      {/* Image count */}
      <p className="text-sm text-surface-muted text-center">
        {images.length} / {maxImages} images
      </p>
    </div>
  )
}
