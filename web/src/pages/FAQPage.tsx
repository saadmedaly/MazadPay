import { useState } from 'react'
import { Plus, HelpCircle, Trash2, ChevronDown, ChevronUp, Edit2 } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { Input } from '@/components/ui/input'
import { useFAQs, useCreateFAQ, useUpdateFAQ, useDeleteFAQ } from '@/hooks/useContent'
import type { FAQItem } from '@/types/api'

export function FAQPage() {
  const [editingFAQ, setEditingFAQ] = useState<FAQItem | null>(null)
  const [isAdding, setIsAdding] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [expandId, setExpandId] = useState<number | null>(null)

  const [newFAQ, setNewFAQ] = useState({
    question_ar: '',
    answer_ar: '',
    display_order: 1
  })

  const { data: faqs, isLoading } = useFAQs()
  const createFAQ = useCreateFAQ()
  const updateFAQ = useUpdateFAQ()
  const deleteFAQ = useDeleteFAQ()

  const handleCreate = () => {
    if (!newFAQ.question_ar || !newFAQ.answer_ar) return
    createFAQ.mutate(newFAQ, {
      onSuccess: () => {
        setIsAdding(false)
        setNewFAQ({ question_ar: '', answer_ar: '', display_order: (faqs?.length ?? 0) + 1 })
      }
    })
  }

  const handleUpdate = () => {
    if (!editingFAQ?.question_ar || !editingFAQ?.answer_ar) return
    updateFAQ.mutate(editingFAQ, {
      onSuccess: () => setEditingFAQ(null)
    })
  }

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader 
        title="الأسئلة الشائعة" 
        subtitle="إدارة محتوى قسم المساعدة والدعم"
        icon={HelpCircle}
        action={{
          label: 'إضافة سؤال جديد',
          icon: Plus,
          onClick: () => setIsAdding(true)
        }}
      />

      <div className="grid gap-4">
        {isLoading ? (
          <div className="admin-card p-12 text-center animate-pulse">
            <div className="w-12 h-12 bg-surface-border rounded-full mx-auto mb-4" />
            <div className="h-4 bg-surface-border w-32 mx-auto rounded" />
          </div>
        ) : faqs?.length === 0 ? (
          <div className="admin-card p-20 text-center">
            <HelpCircle className="w-12 h-12 text-surface-muted mx-auto mb-4" />
            <h3 className="text-white font-bold mb-2">لا يوجد أسئلة حالياً</h3>
            <p className="text-surface-muted text-sm">ابدأ بإضافة أول سؤال لمساعدة المستخدمين.</p>
          </div>
        ) : (
          faqs?.map((faq) => (
            <div key={faq.id} className="admin-card overflow-hidden group">
              <div className="p-4 flex items-center justify-between gap-4">
                <button 
                  onClick={() => setExpandId(expandId === faq.id ? null : faq.id)}
                  className="flex-1 flex items-center gap-4 text-right"
                >
                  <div className="w-8 h-8 rounded-lg bg-mazad-primary/10 border border-mazad-primary/20 flex items-center justify-center text-mazad-primary text-xs font-bold">
                    {faq.display_order}
                  </div>
                  <h4 className="text-white font-bold text-sm">{faq.question_ar}</h4>
                </button>
                <div className="flex items-center gap-2">
                  <button 
                    onClick={() => setEditingFAQ(faq)}
                    className="p-2 rounded-lg text-mazad-primary/60 hover:text-mazad-primary hover:bg-mazad-primary/10 transition-colors"
                  >
                    <Edit2 className="w-4 h-4" />
                  </button>
                  <button 
                    onClick={() => setDeleteId(faq.id)}
                    className="p-2 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                  <button 
                    onClick={() => setExpandId(expandId === faq.id ? null : faq.id)}
                    className="p-2 rounded-lg text-surface-muted hover:text-white transition-colors"
                  >
                    {expandId === faq.id ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                  </button>
                </div>
              </div>
              {expandId === faq.id && (
                <div className="p-4 pt-0 border-t border-surface-border/50 animate-slide-down">
                  <p className="text-surface-muted text-sm leading-relaxed pr-12">{faq.answer_ar}</p>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {/* Add Modal */}
      <ConfirmDialog
        open={isAdding}
        onOpenChange={setIsAdding}
        title="إضافة سؤال جديد"
        description={
          <div className="space-y-4 pt-4 text-right" dir="rtl">
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">السؤال (بالعربية)</label>
              <Input 
                value={newFAQ.question_ar}
                onChange={(e) => setNewFAQ({...newFAQ, question_ar: e.target.value})}
                placeholder="اكتب السؤال هنا..."
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs  font-bold block">الإجابة (بالعربية)</label>
              <textarea
                value={newFAQ.answer_ar}
                onChange={(e) => setNewFAQ({...newFAQ, answer_ar: e.target.value})}
                className="w-full  text-white border bg-surface-border rounded-xl p-3 text-sm   focus:outline-none focus:border-mazad-primary min-h-[150px]"
                placeholder="اكتب الإجابة المفصلة هنا..."
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">ترتيب العرض</label>
              <Input 
                type="number"
                value={newFAQ.display_order}
                onChange={(e) => setNewFAQ({...newFAQ, display_order: parseInt(e.target.value)})}
              />
            </div>
          </div>
        }
        confirmLabel="حفظ السؤال"
        loading={createFAQ.isPending}
        onConfirm={handleCreate}
      />

      {/* Edit Modal */}
      <ConfirmDialog
        open={!!editingFAQ}
        onOpenChange={(v) => !v && setEditingFAQ(null)}
        title="تعديل السؤال"
        description={
          editingFAQ && (
            <div className="space-y-4 pt-4 text-right" dir="rtl">
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">السؤال (بالعربية)</label>
                <Input 
                  value={editingFAQ.question_ar}
                  onChange={(e) => setEditingFAQ({...editingFAQ, question_ar: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs   text-surface-muted font-bold block">الإجابة (بالعربية)</label>
                <textarea
                  value={editingFAQ.answer_ar ?? ''}
                  onChange={(e) => setEditingFAQ({...editingFAQ, answer_ar: e.target.value})}
                  className="w-full bg-surface-border border border-surface-border rounded-xl p-3 text-sm text-white focus:outline-none focus:border-mazad-primary min-h-[150px]"
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">ترتيب العرض</label>
                <Input 
                  type="number"
                  value={editingFAQ.display_order}
                  onChange={(e) => setEditingFAQ({...editingFAQ, display_order: parseInt(e.target.value)})}
                />
              </div>
            </div>
          )
        }
        confirmLabel="تحديث"
        loading={updateFAQ.isPending}
        onConfirm={handleUpdate}
      />

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="حذف السؤال"
        description="هل أنت متأكد من رغبتك في حذف هذا السؤال؟ لا يمكن التراجع عن هذه الخطوة."
        variant="danger"
        confirmLabel="حذف الآن"
        loading={deleteFAQ.isPending}
        onConfirm={() => {
          if (deleteId) deleteFAQ.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
        }}
      />
    </div>
  )
}
