# MazadPay Admin Interface - Rapport Complet des Pages UI

## Résumé Exécutif

| Module | Page UI | CRUD UI | Backend Endpoint | Status |
|--------|-------|--------|---------------|--------|
| Dashboard | ✅ | R | ✅ | ✅ Complet |
| Auctions | ✅ | CRUD | ✅ | ✅ Complet |
| Users | ✅ | R | ✅ | ✅ Complet |
| Transactions | ✅ | R | ✅ | ✅ Complet |
| Categories | ✅ | CRUD | ✅ | ✅ Complet |
| Locations | ✅ | CRUD | ✅ | ✅ Complet |
| Reports | ✅ | R | ✅ | ✅ Complet |
| KYC | ✅ | CRUD | ✅ | ✅ Complet |
| Banners | ✅ | CRUD | ✅ | ✅ Complet |
| Settings | ✅ | CRUD | ✅ | ✅ Complet |
| Blocked Phones | ✅ | CRUD | ✅ | ✅ Complet |
| FAQ | ✅ | R | ✅ | ⚠️ Partiel |
| Tutorials | ✅ | R | ✅ | ⚠️ Partiel |
| Notifications | ✅ | R | ✅ | ✅ Complet |
| Profile | ✅ | RU | ✅ | ✅ Complet |
| Admin Invite | ✅ | C | ✅ | ✅ Complet |

---

## Détail par Page

### 1. AuctionsPage (المزادات)

**Page:** `src/pages/AuctionsPage.tsx`

#### Formulaire (Create/Edit)
| Champ | Type | Requis | Backend Field | Description |
|-------|------|-------|---------------|-------------|
| الفئة (Category) | Select | ✅ | `category_id` | Catégorie depuis `/api/categories` |
| الموقع (Location) | Select | ✅ | `location_id` | Localisation depuis `/api/locations` |
| Titre (AR) | Text | ✅ | `title_ar` | Titre en arabe |
| Description (AR) | Textarea | ❌ | `description_ar` | Description en arabe |
| Titre (FR) | Text | ❌ | `title_fr` | Titre en français |
| Description (FR) | Textarea | ❌ | `description_fr` | Description en français |
| Titre (EN) | Text | ❌ | `title_en` | Titre en anglais |
| Description (EN) | Textarea | ❌ | `description_en` | Description en anglais |
| Prix d'ouverture | Number | ✅ | `start_price` | Prix de départ |
| Prix Achat Immédiat | Number | ❌ | `buy_now_price` | Prix "buy now" |
| Min Increment | Number | ❌ | `min_increment` | Increment minimum |
| Assurance | Number | ❌ | `insurance_amount` | Montant garantie |
| Date/Heure Début | Datetime | ❌ | `start_time` | Début du lot |
| Date/Heure Fin | Datetime | ✅ | `end_time` | Fin du lot |
| Images | Multiple URLs | ❌ | `images[]` | Tableau d'URLs |

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|---------------|
| الصورة | image | `image_urls[0]` |
| العنوان | text | `title_ar` / `title_fr` |
| السعر | price | `start_price` |
| المزايدات | number | `bid_count` |
| ينتهي في | date | `end_time` |
| الحالة | badge | `status` |

**Endpoints:**
- List: `GET /api/admin/auctions`
- Create: `POST /api/auctions`
- Update: `PUT /api/admin/auctions/:id`
- Delete: `DELETE /api/admin/auctions/:id`
- Validate: `PUT /api/admin/auctions/:id/validate`

---

### 2. UsersPage (المستخدمين)

**Page:** `src/pages/UsersPage.tsx`

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|---------------|
| المستخدم | avatar+name | `full_name`, `id` |
| الهاتف | text | `phone` |
| الدور | badge | `role` |
| التوثيق | badge | `is_verified` |
| الحالة | badge | `is_blocked` |
| تاريخ التسجيل | date | `created_at` |

**Endpoints:**
- List: `GET /api/admin/users`
- Block: `PUT /api/admin/users/:id/block`
- Invitation: `POST /api/admin/invitations`

---

### 3. TransactionsPage (المعاملات)

**Page:** `src/pages/TransactionsPage.tsx`

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|---------------|
| المستخدم | text | `user_id` |
| المبلغ | price | `amount` |
| الطريقة | badge | `gateway` |
| الحالة | badge | `status` |
| التاريخ | date | `created_at` |

**Endpoints:**
- List: `GET /api/admin/transactions`
- Validate: `PUT /api/admin/transactions/:id/validate`

---

### 4. CategoriesPage (الفئات)

**Page:** `src/pages/CategoriesPage.tsx`

#### Formulaire (Create/Edit)
| Champ | Type | Requis | Backend Field |
|-------|------|-------|-------------|
| الاسم (عربي) | Text | ✅ | `name_ar` |
| Nom (Français) | Text | ✅ | `name_fr` |
| Name (English) | Text | ❌ | `name_en` |
| Icon | Text | ❌ | `icon_name` |
| Ordre d'affichage | Number | ❌ | `display_order` |

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| ID | number | `id` |
| الاسم (عربي) | text | `name_ar` |
| Nom (Français) | text | `name_fr` |
| Name (English) | text | `name_en` |
| الترتيب | number | `display_order` |

**Endpoints:**
- List: `GET /api/categories`
- Create: `POST /api/admin/categories`
- Update: `PUT /api/admin/categories/:id`
- Delete: `DELETE /api/admin/categories/:id`

---

### 5. LocationsPage (المواقع)

**Page:** `src/pages/LocationsPage.tsx`

#### Formulaire (Create/Edit)
| Champ | Type | Requis | Backend Field |
|-------|------|-------|-------------|
| المدينة (عربي) | Text | ✅ | `city_name_ar` |
| Ville (Français) | Text | ✅ | `city_name_fr` |
| المنطقة (عربي) | Text | ❌ | `area_name_ar` |
| Zone (Français) | Text | ❌ | `area_name_fr` |

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|---------------|
| ID | number | `id` |
| المدينة / Nom | text | `city_name_ar`, `city_name_fr` |
| المنطقة / Zone | text | `area_name_ar`, `area_name_fr` |

**Endpoints:**
- List: `GET /api/locations`
- Create: `POST /api/admin/locations`
- Update: `PUT /api/admin/locations/:id`
- Delete: `DELETE /api/admin/locations/:id`

---

### 6. BannersPage (الإعلانات)

**Page:** `src/pages/BannersPage.tsx`

#### Formulaire (Create/Edit)
| Champ | Type | Requis | Backend Field |
|-------|------|-------|-------------|
| العنوان (عربي) | Text | ✅ | `title_ar` |
| Titre (Français) | Text | ✅ | `title_fr` |
| رابط الصورة | URL | ✅ | `image_url` |
| رابط الويب | URL | ❌ | `link_url` |
| نشط | Toggle | ❌ | `is_active` |
| الترتيب | Number | ❌ | `display_order` |

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| الصورة | image | `image_url` |
| العنوان | text | `title_ar` / `title_fr` |
| نشط | toggle | `is_active` |
| الترتيب | number | `display_order` |

**Endpoints:**
- List: `GET /api/admin/banners`
- Create: `POST /api/admin/banners`
- Update: `PUT /api/admin/banners/:id` **(NOUVEAU)**
- Toggle: `PUT /api/admin/banners/:id/toggle`
- Delete: `DELETE /api/admin/banners/:id`

---

### 7. SettingsPage (الإعدادات)

**Page:** `src/pages/SettingsPage.tsx`

#### Groupes de Paramètres

**إعدادات عامة:**
| Setting | Type | Backend Key |
|---------|------|-------------|
| وضع الصيانة | Toggle | `maintenance_mode` |
| فتح التسجيل | Toggle | `registration_open` |

**إعدادات المزادات:**
| Setting | Type | Backend Key |
|---------|------|-------------|
| مدة المزاد القصوى | Number | `max_auction_duration_hours` |
| مبلغ التأمين الافتراضي | Number | `default_insurance_amount` |
| الحد الأدنى للزيادة | Number | `min_bid_increment` |

**معلومات الاتصال:**
| Setting | Type | Backend Key |
|---------|------|-------------|
| الهاتف | Text | `contact_phone` |
| الإيميل | Text | `contact_email` |
| العنوان | Text | `contact_address` |

**Endpoints:**
- List: `GET /api/admin/settings`
- Update: `PUT /api/admin/settings/:key`

---

### 8. BlockedPhonesPage (أرقام محظورة)

**Page:** `src/pages/BlockedPhonesPage.tsx`

#### Formulaire (Block)
| Champ | Type | Backend Field |
|-------|------|-------------|
| الرقم | Text | `phone` |
| السبب | Text | `reason` |

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| الرقم | text | `phone` |
| السبب | text | `reason` |
| تاريخ الحظر | date | `blocked_at` |

**Endpoints:**
- List: `GET /api/admin/blocked-phones`
- Block: `POST /api/admin/blocked-phones`
- Unblock: `DELETE /api/admin/blocked-phones/:phone`

---

### 9. ReportsPage (البلاغات)

**Page:** `src/pages/ReportsPage.tsx`

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| معرف المزاد | text | `auction_id` |
| السبب | text | `reason` |
| المبلغ | text | `reporter_id` |
| الحالة | badge | `status` |
| التاريخ | date | `created_at` |

**Endpoints:**
- List: `GET /api/admin/reports`
- Review: `PUT /api/admin/reports/:id/review`

---

### 10. KYCPage (توثيق الحسابات)

**Page:** `src/pages/KYCPage.tsx`

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| المستخدم | text | `user_id` |
| نوع التوثيق | badge | `document_type` |
| الحالة | badge | `status` |
| تاريخ الإرسال | date | `submitted_at` |

**Endpoints:**
- List: `GET /api/admin/kyc`
- Review: `PUT /api/admin/kyc/:user_id`

---

### 11. FAQPage (الأسئلة الشائعة)

**Page:** `src/pages/FAQPage.tsx`

**Status:** ⚠️ Affichage uniquement (pas de CRUD dans UI)

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| السؤال | text | `question` |
| الإجابة | text | `answer` |

**Endpoints:**
- List: `GET /api/faq`
- Admin List: `GET /api/admin/faq` **(NOUVEAU)**
- Create: `POST /api/admin/faq`
- Update: `PUT /api/admin/faq/:id`
- Delete: `DELETE /api/admin/faq/:id`

---

### 12. TutorialsPage (شروحات الفيديو)

**Page:** `src/pages/TutorialsPage.tsx`

**Status:** ⚠️ Affichage uniquement (pas de CRUD dans UI)

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| العنوان | text | `title` |
| الفيديو | video | `video_url` |

**Endpoints:**
- List: `GET /api/tutorials`
- Admin List: `GET /api/admin/tutorials` **(NOUVEAU)**
- Create: `POST /api/admin/tutorials`
- Update: `PUT /api/admin/tutorials/:id`
- Delete: `DELETE /api/admin/tutorials/:id`

---

### 13. NotificationsPage (الإشعارات)

**Page:** `src/pages/NotificationsPage.tsx`

**Status:** ⚠️ Affichage uniquement dans UI (CRUD backend ajouté)

#### Table (Liste)
| Colonne | Source | Backend Field |
|--------|--------|-------------|
| العنوان | text | `title` |
| الرسالة | text | `body` |
| النوع | badge | `type` |
| المقروء | status | `is_read` |
| التاريخ | date | `created_at` |

**Endpoints:**
- List (user): `GET /api/notifications`
- Admin List: `GET /api/admin/notifications` **(NOUVEAU)**
- Send: `POST /api/admin/notifications/send` **(NOUVEAU)**
- Delete: `DELETE /api/admin/notifications/:id` **(NOUVEAU)**

---

### 14. ProfilePage (الملف الشخصي)

**Page:** `src/pages/ProfilePage.tsx`

#### Formulaire
| Champ | Type | Backend Field |
|-------|------|-------------|
| الاسم الكامل | Text | `full_name` |
| اللغة المفضلة | Select | `language` |
| الإشعار بالبريد | Toggle | `email_notifications` |
| الإشعار_push | Toggle | `push_notifications` |

**Endpoints:**
- Get: `GET /api/users/me`
- Update: `PUT /api/users/me`
- Language: `PUT /api/users/me/language`
- Notification Prefs: `PUT /api/users/me/notification-prefs`

---

### 15. AdminInvitePage (دعوة مشرف)

**Page:** `src/pages/AdminInvitePage.tsx`

#### Fonctionnalité
-Générer un lien d'invitation admin
- Copier le lien

**Endpoints:**
- Generate: `POST /api/admin/invitations`

---

## Hooks Utilisés

| Hook | Page | Usage |
|------|-----|-------|
| `useAuctions` | AuctionsPage | Liste + CRUD |
| `useCategories` | CategoriesPage | Liste |
| `useLocations` | LocationsPage | Liste |
| `useUsers` | UsersPage | Liste + Block |
| `useTransactions` | TransactionsPage | Liste |
| `useReports` | ReportsPage | Liste |
| `useKYC` | KYCPage | Liste + Review |
| `useSettings` | SettingsPage | Liste + Update |
| `useBlockedPhones` | BlockedPhonesPage | Liste + Block/Unblock |
| `useGenerateInvitation` | AdminInvitePage | Generate |

---

## API Endpoints Récapitulatifs

### Admin (CRUD)
```
GET    /api/admin/auctions
POST   /api/admin/auctions
PUT    /api/admin/auctions/:id
DELETE /api/admin/auctions/:id
PUT    /api/admin/auctions/:id/validate

GET    /api/admin/users
PUT    /api/admin/users/:id/block

GET    /api/admin/transactions
PUT    /api/admin/transactions/:id/validate

GET    /api/admin/categories
POST   /api/admin/categories
PUT    /api/admin/categories/:id
DELETE /api/admin/categories/:id

GET    /api/admin/locations
POST   /api/admin/locations
PUT    /api/admin/locations/:id
DELETE /api/admin/locations/:id

GET    /api/admin/banners
POST   /api/admin/banners
PUT    /api/admin/banners/:id
PUT    /api/admin/banners/:id/toggle
DELETE /api/admin/banners/:id

GET    /api/admin/settings
PUT    /api/admin/settings/:key

GET    /api/admin/blocked-phones
POST   /api/admin/blocked-phones
DELETE /api/admin/blocked-phones/:phone

GET    /api/admin/reports
PUT    /api/admin/reports/:id/review

GET    /api/admin/kyc
PUT    /api/admin/kyc/:user_id

GET    /api/admin/faq
POST   /api/admin/faq
PUT    /api/admin/faq/:id
DELETE /api/admin/faq/:id

GET    /api/admin/tutorials
POST   /api/admin/tutorials
PUT    /api/admin/tutorials/:id
DELETE /api/admin/tutorials/:id

GET    /api/admin/notifications          (NOUVEAU)
POST   /api/admin/notifications/send    (NOUVEAU)
DELETE /api/admin/notifications/:id    (NOUVEAU)

POST   /api/admin/invitations
```

---

## Statut des Implémentations

| Page | Status UI | Status Backend | Commentaire |
|------|----------|---------------|------------|
| Dashboard | ✅ Complet | ✅ | Stats, charts |
| Auctions | ✅ Complet | ✅ | Full CRUD + languages + images |
| Users | ✅ Complet | ✅ | List + Block |
| Transactions | ✅ Complet | ✅ | List + Validate |
| Categories | ✅ Complet | ✅ | Full CRUD |
| Locations | ✅ Complet | ✅ | Full CRUD |
| Reports | ✅ Complet | ✅ | List + Review |
| KYC | ✅ Complet | ✅ | Full CRUD |
| Banners | ✅ Complet | ✅ | Full CRUD (edit ajouté) |
| Settings | ✅ Complet | ✅ | Full CRUD |
| Blocked Phones | ✅ Complet | ✅ | Full CRUD |
| FAQ | ⚠️ Affichage | ✅ | CRUD backend ajouté |
| Tutorials | ⚠️ Affichage | ✅ | CRUD backend ajouté |
| Notifications | ⚠️ Affichage | ✅ | CRUD backend ajouté |
| Profile | ✅ Complet | ✅ | Full CRUD |
| Admin Invite | ✅ Complet | ✅ | Generate link |