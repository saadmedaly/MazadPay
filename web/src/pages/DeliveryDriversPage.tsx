import { Truck } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'

export function DeliveryDriversPage() {
  return (
    <div className="p-6 space-y-6 flex flex-col items-center justify-center min-h-[60vh] text-center" dir="rtl">
      <div className="w-24 h-24 rounded-full bg-mazad-primary/10 flex items-center justify-center mb-6">
        <Truck className="w-12 h-12 text-mazad-primary opacity-50" />
      </div>
      <PageHeader
        title="خدمة التوصيل"
        subtitle="إدارة سائقي التوصيل وعمليات النقل"
        icon={Truck}
      />
      <div className="mt-8 p-10 bg-surface-card border border-surface-border rounded-3xl shadow-xl max-w-lg">
        <h2 className="text-2xl font-bold text-white mb-4">الخدمة غير متوفرة حالياً</h2>
        <p className="text-surface-muted leading-relaxed mb-8">
          نحن نعمل على إطلاق خدمة التوصيل MazadDelivery في الإصدارات القادمة. 
          ترقبوا التحديثات الجديدة قريباً!
        </p>
        <div className="py-2 px-6 bg-mazad-primary/20 text-mazad-primary rounded-full inline-block font-bold">
          قريباً في V2.5
        </div>
      </div>
    </div>
  )
}

