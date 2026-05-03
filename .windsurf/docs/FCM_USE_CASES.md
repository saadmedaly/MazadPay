# FCM Notification System - Use Cases Documentation

## Overview
This document describes all Firebase Cloud Messaging (FCM) notification use cases implemented in the MazadPay application.

---

## Implemented Use Cases

### 1. Auction Pending (User → Admin)

**Trigger:** When a user submits a new auction for approval

**Recipients:** All admins and superadmins

**Timing:** Immediate

**Data Payload:**
```json
{
  "type": "auction_pending",
  "auctionId": "uuid",
  "sellerId": "uuid"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | مزاد جديد في الانتظار | {userName} أنشأ مزاد: {auctionTitle} |
| French | Nouvelle enchère en attente | {userName} a créé une enchère: {auctionTitle} |
| English | New auction pending | {userName} created an auction: {auctionTitle} |

**Deep Link:** Navigate to admin review page

**Implementation:**
- `backend/internal/services/auction_service.go` - `Create()` method
- Sends notification via `notifSvc.NotifyAdminsLocalized()`

---

### 2. Auction Approved (Admin → User)

**Trigger:** When an admin approves a pending auction request

**Recipients:** The auction creator

**Timing:** Immediate (after approval)

**Data Payload:**
```json
{
  "type": "auction_approved",
  "request_id": "uuid",
  "auction_id": "uuid"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | تمت الموافقة على المزاد! | مزادك "{auctionTitle}" أصبح متاحًا الآن |
| French | Enchère approuvée ! | Votre enchère "{auctionTitle}" est maintenant en ligne |
| English | Auction approved! | Your auction "{auctionTitle}" is now live |

**Deep Link:** Navigate to auction details page

**Implementation:**
- `backend/internal/services/request_service.go` - `ReviewAuctionRequest()` method
- Uses user's preferred language from profile

---

### 3. Auction Rejected (Admin → User)

**Trigger:** When an admin rejects a pending auction request

**Recipients:** The auction creator

**Timing:** Immediate (after rejection)

**Data Payload:**
```json
{
  "type": "auction_rejected",
  "request_id": "uuid",
  "reason": "optional rejection reason"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | تم رفض المزاد | السبب: {reason} |
| French | Enchère refusée | Raison: {reason} |
| English | Auction rejected | Reason: {reason} |

**Deep Link:** Navigate to request history page

**Implementation:**
- `backend/internal/services/request_service.go` - `ReviewAuctionRequest()` method

---

### 4. Auction Ending Soon (System → Seller & Bidders)

**Trigger:** Background scheduler detects auction ending in 5 minutes

**Recipients:** 
- Auction seller
- Active bidders

**Timing:** Automated check every minute, sends at 5 minutes before end

**Data Payload:**
```json
{
  "type": "auction_ending_soon",
  "auctionId": "uuid"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | ⚡ الفرصة الأخيرة! | "{auctionTitle}" ينتهي في 5 دقائق |
| French | ⚡ Dernière chance ! | "{auctionTitle}" se termine dans 5 minutes |
| English | ⚡ Last chance! | "{auctionTitle}" ends in 5 minutes |

**Deep Link:** Navigate to auction details page with bidding interface

**Implementation:**
- `backend/internal/services/auction_scheduler.go`
- Runs every minute via `AuctionScheduler`
- Method: `checkEndingSoon()`

---

### 5. Auction Ended (System → Seller & Winner)

**Trigger:** Background scheduler detects auction has ended

**Recipients:**
- Auction seller
- Auction winner (if any)

**Timing:** Automated check every minute

**Data Payload:**
```json
{
  "type": "auction_ended",
  "auctionId": "uuid",
  "finalPrice": "decimal"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | انتهى المزاد | تم بيع "{auctionTitle}" بـ {finalPrice} MRU |
| French | Enchère terminée | "{auctionTitle}" vendu pour {finalPrice} MRU |
| English | Auction ended | "{auctionTitle}" sold for {finalPrice} MRU |

**Deep Link:** Navigate to auction details or payment page

**Implementation:**
- `backend/internal/services/auction_scheduler.go`
- Method: `checkEndedAuctions()`

---

### 6. Auction Won (System → Winner)

**Trigger:** When a bidder wins an auction (highest bid)

**Recipients:** The winning bidder

**Timing:** Immediate after auction ends

**Data Payload:**
```json
{
  "type": "auction_won",
  "auctionId": "uuid",
  "finalPrice": "decimal"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | تهانينا! لقد فزت | مزاد "{auctionTitle}" - {finalPrice} MRU |
| French | Félicitations ! Vous avez gagné | Enchère "{auctionTitle}" - {finalPrice} MRU |
| English | Congratulations! You won | Auction "{auctionTitle}" - {finalPrice} MRU |

**Deep Link:** Navigate to payment/confirmation page

**Implementation:**
- `backend/internal/services/auction_scheduler.go`
- Called within `checkEndedAuctions()`

---

### 7. Payment Received (System → Seller)

**Trigger:** When payment is confirmed for an auction

**Recipients:** The auction seller

**Timing:** Immediate after payment confirmation

**Data Payload:**
```json
{
  "type": "payment_received",
  "auctionId": "uuid",
  "amount": "decimal"
}
```

**Localized Messages:**
| Language | Title | Body |
|----------|-------|------|
| Arabic | 💰 تم استلام الدفع | {amount} MRU لمزاد "{auctionTitle}" |
| French | 💰 Paiement reçu | {amount} MRU pour "{auctionTitle}" |
| English | 💰 Payment received | {amount} MRU for "{auctionTitle}" |

**Deep Link:** Navigate to wallet/transaction page

**Implementation:**
- TODO: Add to payment service

---

### 8. Banner Approved/Rejected (Admin → User)

**Trigger:** When an admin approves or rejects a banner request

**Recipients:** The banner request creator

**Timing:** Immediate

**Data Payload:**
```json
{
  "type": "banner_approved" | "banner_rejected",
  "request_id": "uuid",
  "banner_id": "int (if approved)"
}
```

**Localized Messages:**
| Language | Title (Approved) | Body |
|----------|------------------|------|
| Arabic | تم قبول طلب الإعلان | تم قبول طلبك لإضافة الإعلان {bannerTitle} |
| French | Publicité approuvée | Votre demande de publicité {bannerTitle} a été acceptée |
| English | Banner approved | Your banner request {bannerTitle} has been approved |

**Deep Link:** Navigate to banner status page

**Implementation:**
- `backend/internal/services/request_service.go`
- Methods: `ReviewBannerRequest()` for both approved and rejected

---

## Technical Implementation

### Flutter Frontend

**Files:**
- `lib/services/fcm_service.dart` - Core FCM service
- `lib/widgets/notification_handler.dart` - Deep linking handler
- `lib/pages/notifications_page.dart` - Notification history UI

**Key Features:**
1. FCM token registration on app start
2. Foreground notification display via `flutter_local_notifications`
3. Background message handling
4. Deep linking from notification taps
5. Notification history with filtering
6. Multi-language support

### Go Backend

**Files:**
- `internal/services/notification_service.go` - Core notification service
- `internal/services/notification_localizations.go` - Translation strings
- `internal/services/auction_scheduler.go` - Background job scheduler
- `internal/repository/auction_repo.go` - Auction queries

**Key Features:**
1. Multi-language notification sending
2. Admin broadcast capabilities
3. Automated scheduler for ending auctions
4. User language preference support

---

## Notification Flow

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│   Event Trigger │────▶│   Backend    │────▶│    FCM      │
│  (Auction, etc) │     │   Service    │     │   Server    │
└─────────────────┘     └──────────────┘     └──────┬──────┘
                                                    │
                       ┌────────────────────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  User Device(s) │
              │  ├─ Foreground  │
              │  ├─ Background  │
              │  └─ Terminated  │
              └─────────────────┘
```

---

## Configuration

### Firebase Setup
1. Add Firebase project to Google Cloud Console
2. Download `google-services.json` (Android) and `GoogleService-Info.plist` (iOS)
3. Place files in appropriate platform directories
4. Set up service account key for backend (`backend/config/serviceAccountKey.json`)

### Environment Variables
```bash
# Backend
FIREBASE_SERVICE_ACCOUNT_PATH=/path/to/serviceAccountKey.json

# Flutter - no additional env vars needed
```

---

## Testing

### Manual Testing Steps
1. Create auction → Check admin receives notification
2. Approve auction → Check user receives approval notification
3. Wait for auction to end → Check ending soon and ended notifications
4. Tap notification → Verify deep linking works

### Unit Tests
- TODO: Add notification service tests
- TODO: Add scheduler tests

---

## Future Enhancements

1. **Rich Notifications** - Add images to notifications
2. **Notification Preferences** - Allow users to customize notification settings
3. **Batch Notifications** - Group multiple notifications
4. **Web Push** - Add browser notifications for web version
5. **Analytics** - Track notification open rates and engagement
