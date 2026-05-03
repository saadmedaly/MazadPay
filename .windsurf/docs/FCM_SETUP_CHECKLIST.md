# FCM Configuration Checklist - MazadPay

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  Mobile (Flutter)│────▶│   Backend    │────▶│    FCM      │
│  Android/iOS     │     │   Go/Fiber   │     │   Server    │
└─────────────────┘     └──────────────┘     └──────┬──────┘
                                                      │
                         ┌────────────────────────────┘
                         │
                         ▼
                ┌─────────────────┐
                │  React Web Admin│
                │  (No FCM needed) │
                └─────────────────┘
```

---

## Current Status

### ✅ Backend (Go/Fiber) - PARTIALLY CONFIGURED

**Files configured:**
- `backend/internal/config/config.go` - Firebase config structure ✅
- `backend/internal/services/notification_service.go` - FCM service implementation ✅
- `backend/internal/services/notification_localizations.go` - Multilingual support ✅
- `backend/.env` - Environment variable `FIREBASE_SERVICE_ACCOUNT_PATH` ✅

**Files present:**
- `backend/configs/firebase-service-account.json` - Service account key ✅

**Implementation:**
- `SendPush()` - Send notifications to single user ✅
- `SendLocalizedPush()` - Send with AR/FR/EN support ✅
- `NotifyAdminsLocalized()` - Send to all admins ✅
- `NotifyAuctionEndingSoon()` - Scheduler notifications ✅

**What's working:**
- Backend can send FCM notifications if service account is valid
- Multilingual notifications (Arabic, French, English)
- Deep linking with payload data

**To verify:**
```bash
# Check if service account file exists and is valid
cat backend/configs/firebase-service-account.json | jq .

# Test backend notification endpoint (when running)
curl -X POST http://localhost:8082/v1/api/notifications/test \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### ⚠️ Mobile (Flutter) - MISSING CONFIG FILES

**What's configured:**
- `lib/services/fcm_service.dart` - FCM service with lazy initialization ✅
- `lib/widgets/notification_handler.dart` - Deep linking handler ✅
- `lib/pages/notifications_page.dart` - Notification history UI ✅
- `lib/main.dart` - Firebase initialization ✅

**What's MISSING (Critical):**

#### Android
1. **`android/app/google-services.json`** - NOT PRESENT ❌
   - Download from Firebase Console → Project Settings → Your App → google-services.json
   - Place in `android/app/`

2. **`android/build.gradle` modifications** - NOT DONE ❌
   ```gradle
   // Add to android/build.gradle (project level)
   buildscript {
       dependencies {
           classpath 'com.google.gms:google-services:4.4.0'
       }
   }
   ```

3. **`android/app/build.gradle` modifications** - NOT DONE ❌
   ```gradle
   // Add at the bottom of android/app/build.gradle
   plugins {
       id 'com.android.application'
       id 'kotlin-android'
       id 'dev.flutter.flutter-gradle-plugin'
       id 'com.google.gms.google-services'  // Add this
   }
   ```

#### iOS
1. **`ios/Runner/GoogleService-Info.plist`** - NOT PRESENT ❌
   - Download from Firebase Console → Project Settings → Your App → GoogleService-Info.plist
   - Place in `ios/Runner/`

2. **`ios/Runner/AppDelegate.swift` modifications** - NOT DONE ❌
   ```swift
   import FirebaseCore  // Add this import
   
   @UIApplicationMain
   class AppDelegate: UIResponder, UIApplicationDelegate {
       func application(_ application: UIApplication,
                       didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
           FirebaseApp.configure()  // Add this line
           GeneratedPluginRegistrant.register(with: self)
           return true
       }
   }
   ```

**Dependencies already in pubspec.yaml:**
```yaml
dependencies:
  firebase_core: ^2.24.2
  firebase_messaging: ^14.7.10
  flutter_local_notifications: ^16.3.2
```

---

### ✅ Web (React) - NOT NEEDED

**Status:** No FCM configuration required ✅

The React web interface is for admin purposes only and doesn't need push notifications. FCM is mobile-only in your architecture.

---

## Setup Instructions

### Step 1: Create Firebase Project (if not done)

1. Go to [Firebase Console](https://console.firebase.google.com)
2. Create new project or select existing
3. Add Android app:
   - Package name: `com.mezadpay.app.mezad_pay`
   - Download `google-services.json`
4. Add iOS app:
   - Bundle ID: `com.mezadpay.app.mezadPay`
   - Download `GoogleService-Info.plist`

### Step 2: Mobile Configuration

#### Android
```bash
# Place file
cp ~/Downloads/google-services.json MazadPay/android/app/

# Modify android/build.gradle
cat >> android/build.gradle << 'EOF'
buildscript {
    dependencies {
        classpath 'com.google.gms:google-services:4.4.0'
    }
}
EOF

# Modify android/app/build.gradle - add plugin
echo "apply plugin: 'com.google.gms.google-services'" >> android/app/build.gradle
```

#### iOS
```bash
# Place file
cp ~/Downloads/GoogleService-Info.plist MazadPay/ios/Runner/

# Install pods
cd MazadPay/ios && pod install
```

### Step 3: Backend Service Account

Already configured. Verify:
```bash
ls -la backend/configs/firebase-service-account.json
```

If missing, generate from Firebase Console:
- Project Settings → Service Accounts → Generate New Private Key

### Step 4: Test FCM

```bash
# 1. Start backend
cd backend && go run ./cmd/server/main.go

# 2. Run Flutter on device (not web)
cd MazadPay && flutter run -d <your_device_id>

# 3. Place a bid or trigger auction approval to test notification
```

---

## Dependency Matrix

| Component | Depends On | Files Needed |
|-----------|-----------|--------------|
| **Backend** | Firebase Service Account | `configs/firebase-service-account.json` |
| **Mobile (Flutter)** | Backend API + Firebase Config | `google-services.json`, `GoogleService-Info.plist` |
| **Web (React)** | Backend API only | None (no FCM) |

---

## Quick Checklist

- [ ] Firebase project created
- [ ] Android app registered in Firebase
- [ ] iOS app registered in Firebase
- [ ] `google-services.json` downloaded and placed in `android/app/`
- [ ] `GoogleService-Info.plist` downloaded and placed in `ios/Runner/`
- [ ] Android Gradle files modified
- [ ] iOS AppDelegate.swift modified
- [ ] iOS pods installed (`cd ios && pod install`)
- [ ] Backend service account key valid
- [ ] Backend running and connected to Firebase
- [ ] Test notification sent successfully

---

## Testing FCM

### Backend Test
```bash
# Trigger a notification by creating an auction
# Or use the test endpoint when implemented
curl -X POST http://localhost:8082/v1/api/admin/notifications/test \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"userId": "USER_ID", "title": "Test", "body": "Hello"}'
```

### Mobile Test
1. Login to app
2. Create an auction request
3. Approve it from admin (or use admin API)
4. Check if notification appears

---

## Troubleshooting

### "No Firebase App '[DEFAULT]' has been created"
- Missing `google-services.json` or `GoogleService-Info.plist`
- Files not properly configured

### "Invalid service account"
- Backend service account file is invalid or expired
- Generate new key from Firebase Console

### "Registration token not found"
- User hasn't granted notification permissions
- Device not connected to internet during registration

### Notifications not received on web
- **Expected** - FCM is disabled on web in your configuration
- Use mobile device for testing
