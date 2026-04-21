import React from 'react'
import { cn } from '@/lib/utils'

interface TabsProps {
  tabs: Array<{
    id: string
    label: string
    content: React.ReactNode
  }>
  activeTab: string
  onTabChange: (tabId: string) => void
  className?: string
}

export function Tabs({ tabs, activeTab, onTabChange, className }: TabsProps) {
  return (
    <div className={cn("w-full", className)}>
      {/* Tab Headers */}
      <div className="flex space-x-1 border-b border-surface-border">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={cn(
              "px-4 py-2 text-sm font-bold transition-all border-b-2",
              activeTab === tab.id
                ? "text-mazad-primary border-mazad-primary"
                : "text-surface-muted border-transparent hover:text-white hover:border-surface-muted"
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="mt-6">
        {tabs.find((tab) => tab.id === activeTab)?.content}
      </div>
    </div>
  )
}
