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
  title_ar: string
  title_fr: string | null
  title_en: string | null
  description_ar: string | null
  description_fr: string | null
  description_en: string | null
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
  buy_now_price: string | null
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
  total_auctions: number
  total_bids: number
  total_revenue: number
  today_revenue: number
  week_deposits: number
  pending_kycs: number
}

export interface KYCVerification {
  user_id: string
  id_card_front_url: string | null
  id_card_back_url: string | null
  nni_number: string | null
  status: 'pending' | 'approved' | 'rejected'
  admin_notes: string | null
  reviewed_by: string | null
  reviewed_at: string | null
  created_at: string
  user?: AdminUser
}

export interface FAQItem {
  id: number
  question_ar: string
  question_fr: string | null
  answer_ar: string
  answer_fr: string | null
  display_order: number
}

export interface Tutorial {
  id: number
  title_ar: string
  title_fr: string | null
  video_url: string
  thumbnail_url: string | null
  category: string | null
  display_order: number
}

export interface Category {
  id: number
  name_ar: string
  name_fr: string
  name_en: string
  parent_id: number | null
  icon_name: string | null
  display_order: number
  children?: Category[]
}

export interface Location {
  id: number
  city_name_ar: string
  city_name_fr: string
  area_name_ar: string
  area_name_fr: string
}

export interface AppRating {
  id: string
  user_id: string
  rating: number
  comment: string | null
  created_at: string
  user?: AdminUser
}

export interface Notification {
  id: string
  user_id: string
  type: 'bid' | 'win' | 'payment' | 'system' | 'ad'
  title: string
  body: string | null
  is_read: boolean
  reference_id: string | null
  reference_type: string | null
  data: Record<string, unknown> | null
  created_at: string
  user?: AdminUser
}

export interface BlockedPhone {
  phone: string
  reason: string | null
  blocked_by: string | null
  blocked_at: string
  expires_at: string | null
}

export interface AuctionPayment {
  id: string
  auction_id: string
  winner_id: string
  transaction_id: string | null
  amount: string
  status: 'pending' | 'completed' | 'failed' | 'overdue'
  deadline: string
  paid_at: string | null
  created_at: string
  auction?: Auction
  winner?: AdminUser
}

export interface DeliveryTimeline {
  id: number
  request_id: string
  step_name: string
  description: string | null
  completed_at: string | null
  created_at: string
}

export interface ServiceRequest {
  id: string
  user_id: string
  service_type: 'delivery' | 'taxi' | 'intercity' | 'shipping' | 'other'
  pickup_location: string | null
  delivery_location: string | null
  status: 'pending' | 'assigned' | 'in_transit' | 'delivered' | 'canceled'
  tracking_number: string | null
  estimated_price: string | null
  actual_price: string | null
  notes: string | null
  driver_id: string | null
  completed_at: string | null
  created_at: string
  user?: AdminUser
  driver?: AdminUser
  timeline?: DeliveryTimeline[]
}

export interface Banner {
  id: number
  title_ar: string | null
  title_fr: string | null
  image_url: string
  target_url: string | null
  is_active: boolean
  starts_at: string | null
  ends_at: string | null
  display_order: number
  created_at: string
}

export interface SystemSetting {
  id: number
  key: string
  value: string
  type: string
}

export interface BlockedPhone {
  phone: string
  reason: string | null
  blocked_at: string
  expires_at: string | null
}

export interface ReportReason {
  id: string
  label_ar: string
  label_fr: string
  label_en: string
}
