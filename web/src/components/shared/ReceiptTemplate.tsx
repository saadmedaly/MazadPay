import { forwardRef } from 'react'
import type { Transaction } from '@/types/api'
import { GATEWAY_LABELS } from '@/lib/constants'
import { formatPrice, shortID } from '@/lib/formatters'

interface ReceiptTemplateProps {
  txn: Transaction
  beneficiary?: string | null
}

// Format spécifique au reçu PDF : "22 يوليو 2026 - 11:44" (séparateur "-" entre date et
// heure, plus lisible sur un reçu imprimé que la virgule utilisée par formatDate()
// ailleurs dans l'admin, qu'on ne modifie pas pour ne pas affecter le reste de l'UI).
function formatReceiptDate(date: string): string {
  const d = new Date(date)
  if (isNaN(d.getTime())) return '—'
  const datePart = new Intl.DateTimeFormat('ar', { day: 'numeric', month: 'long', year: 'numeric' }).format(d)
  const timePart = new Intl.DateTimeFormat('ar', { hour: '2-digit', minute: '2-digit', hour12: false }).format(d)
  return `${datePart} - ${timePart}`
}

const TYPE_LABELS: Record<string, string> = {
  deposit: 'إيداع رصيد',
  withdraw: 'استرجاع مبلغ التأمين',
  bid_hold: 'تجميد تأمين مزايدة',
  bid_refund: 'استرجاع تأمين مزايدة',
  payment: 'دفع',
}

const STATUS_LABELS: Record<string, string> = {
  completed: 'مكتملة',
  pending: 'قيد الانتظار',
  pending_review: 'قيد المراجعة',
  failed: 'فاشلة',
  refunded: 'مُرجعة',
}

// Gabarit visuel imprimé en PDF via html2canvas + jsPDF (voir receiptPdf.ts). Rendu
// caché de la page (position absolue hors écran) — jamais affiché directement à
// l'utilisateur, seulement capturé en image puis converti en PDF. Ce choix (HTML/CSS
// plutôt que dessin direct jsPDF) est nécessaire car jsPDF ne supporte pas nativement
// le rendu RTL/arabe correctement, alors que le rendu HTML du navigateur le fait
// parfaitement.
export const ReceiptTemplate = forwardRef<HTMLDivElement, ReceiptTemplateProps>(
  ({ txn, beneficiary }, ref) => {
    const isSuccess = txn.status === 'completed'

    return (
      <div
        ref={ref}
        dir="rtl"
        style={{
          width: '600px',
          padding: '40px',
          backgroundColor: '#ffffff',
          fontFamily: 'Tajawal, Cairo, Arial, sans-serif',
          color: '#111827',
        }}
      >
        {/* En-tête MazadPay */}
        <div style={{ textAlign: 'center', marginBottom: '24px' }}>
          <div style={{ fontSize: '28px', fontWeight: 800, color: '#0084FF' }}>
            MazadPay <span style={{ color: '#111827' }}>مزاد باي</span>
          </div>
          <div style={{ fontSize: '14px', color: '#6b7280', marginTop: '4px' }}>
            إيصال دفع إلكتروني
          </div>
        </div>

        {/* Carte principale */}
        <div
          style={{
            border: '1px solid #e5e7eb',
            borderRadius: '16px',
            padding: '32px',
            textAlign: 'center',
          }}
        >
          {/* Badge succès */}
          <div
            style={{
              width: '64px',
              height: '64px',
              borderRadius: '50%',
              backgroundColor: isSuccess ? '#d1fae5' : '#fee2e2',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 16px',
              fontSize: '32px',
              color: isSuccess ? '#059669' : '#dc2626',
            }}
          >
            {isSuccess ? '✓' : '!'}
          </div>

          <div style={{ fontSize: '20px', fontWeight: 700, marginBottom: '4px' }}>
            {isSuccess ? 'العملية ناجحة' : (STATUS_LABELS[txn.status] ?? txn.status)}
          </div>
          <div style={{ fontSize: '14px', color: '#6b7280', marginBottom: '24px' }}>
            {TYPE_LABELS[txn.type] ?? txn.type}
          </div>

          <div style={{ fontSize: '32px', fontWeight: 800, color: '#0084FF', marginBottom: '24px' }}>
            {formatPrice(parseFloat(txn.amount))}
          </div>

          {/* Détails */}
          {/* Priorité d'affichage demandée : full_name, sinon phone, sinon shortID —
              jamais de "—" tant que user_id existe (audit correction affichage nom). */}
          <div style={{ textAlign: 'right', borderTop: '1px dashed #d1d5db', paddingTop: '20px' }}>
            {[
              ['اسم المستخدم', txn.user_full_name || txn.user_phone || shortID(txn.user_id)],
              ['رقم الهاتف', txn.user_phone || 'غير متوفر'],
              ...(beneficiary ? [['المستفيد', beneficiary]] : []),
              ['بوابة/طريقة الدفع', txn.gateway ? (GATEWAY_LABELS[txn.gateway] ?? txn.gateway) : '—'],
              ['معرف المعاملة', shortID(txn.id)],
              ['التاريخ والوقت', formatReceiptDate(txn.created_at)],
            ].map(([label, value]) => (
              <div
                key={label}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  padding: '8px 0',
                  fontSize: '14px',
                  borderBottom: '1px solid #f3f4f6',
                }}
              >
                <span style={{ color: '#6b7280' }}>{label}</span>
                <span style={{ fontWeight: 700 }}>{value}</span>
              </div>
            ))}
          </div>
        </div>

        <div style={{ textAlign: 'center', marginTop: '20px', fontSize: '12px', color: '#9ca3af' }}>
          هذا الإيصال صادر من منصة MazadPay
        </div>
      </div>
    )
  }
)

ReceiptTemplate.displayName = 'ReceiptTemplate'
