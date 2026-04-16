export const ROUTES = {
  LOGIN:        '/login',
  DASHBOARD:    '/',
  AUCTIONS:     '/auctions',
  AUCTION:      '/auctions/:id',
  TRANSACTIONS: '/transactions',
  TRANSACTION:  '/transactions/:id',
  USERS:        '/users',
  USER:         '/users/:id',
  REPORTS:      '/reports',
  BANNERS:      '/banners',
} as const

export const STATUS_COLORS = {
  // Enchères
  pending:        'bg-amber-500/15 text-amber-400 border-amber-500/30',
  active:         'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  ended:          'bg-slate-500/15 text-slate-400 border-slate-500/30',
  rejected:       'bg-red-500/15 text-red-400 border-red-500/30',
  canceled:       'bg-red-500/15 text-red-400 border-red-500/30',
  // Transactions
  pending_review: 'bg-orange-500/15 text-orange-400 border-orange-500/30',
  completed:      'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  failed:         'bg-red-500/15 text-red-400 border-red-500/30',
  refunded:       'bg-blue-500/15 text-blue-400 border-blue-500/30',
} as const

export const STATUS_LABELS: Record<string, string> = {
  pending:        'قيد الانتظار',
  active:         'نشط',
  ended:          'منتهي',
  rejected:       'مرفوض',
  canceled:       'ملغي',
  pending_review: 'في انتظار المراجعة',
  completed:      'مكتمل',
  failed:         'فاشل',
  refunded:       'مسترجع',
}

export const GATEWAY_LABELS: Record<string, string> = {
  Bankily:  'بانكيلي',
  Masrivi:  'مصرفي',
  Sedad:    'سداد',
  Click:    'كليك',
  Bank:     'تحويل بنكي',
}
