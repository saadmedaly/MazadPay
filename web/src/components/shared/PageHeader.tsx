interface Props {
  title: string
  subtitle?: string
  children?: React.ReactNode  // action buttons
}

export function PageHeader({ title, subtitle, children }: Props) {
  return (
    <div className="flex items-start justify-between mb-8">
      <div>
        <h1 className="text-3xl font-display font-bold text-white tracking-tight">{title}</h1>
        {subtitle && <p className="text-sm text-surface-muted mt-1.5 font-medium">{subtitle}</p>}
      </div>
      {children && <div className="flex items-center gap-2">{children}</div>}
    </div>
  )
}
