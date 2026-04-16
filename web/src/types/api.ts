export interface APIResponse<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

export interface PaginatedResponse<T> {
  success: boolean
  data: T[]
  meta: {
    total: number
    page: number
    per_page: number
  }
}

export interface AdminUser {
  id: string
  phone: string
  full_name: string | null
  email: string | null
  profile_pic_url: string | null
  city: string | null
  language_pref: string
  is_active: boolean
  role: 'user' | 'admin' | 'driver'
  is_verified: boolean
  last_login_at: string | null
  created_at: string
}

export interface Auction {
  id: string
  seller_id: string
  category_id: number
  location_id: number | null
  title: string
  description: string | null
  start_price: string
  current_price: string
  min_increment: string
  insurance_amount: string
  start_time: string
  end_time: string
  images?: string[]
  category?: string
  city?: string
  status: 'pending' | 'active' | 'ended' | 'canceled' | 'rejected'
  lot_number: string | null
  views: number
  bidder_count: number
  winner_id: string | null
  item_details: Record<string, unknown>
  created_at: string
  rejection_reason?: string
}

export interface Transaction {
  id: string
  user_id: string
  auction_id: string | null
  type: 'deposit' | 'withdraw' | 'bid_hold' | 'bid_refund' | 'payment'
  amount: string
  gateway: string | null
  status: 'pending' | 'pending_review' | 'completed' | 'failed' | 'refunded'
  receipt_url: string | null
  admin_notes: string | null
  reviewed_by: string | null
  reviewed_at: string | null
  created_at: string
}

export interface DashboardStats {
  active_auctions: number
  pending_auctions: number
  pending_transactions: number
  pending_reports: number
  total_users: number
  verified_users: number
  today_deposits: number
  week_deposits: number
  today_bids: number
}
