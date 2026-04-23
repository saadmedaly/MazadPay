# Analyse Architecture MazadPay - Backend API
## Distinction Endpoints Publics vs Admin

**Date:** 22 Avril 2026  
**Projet:** MazadPay - Application de mauxidère numérique

---

## 🎯 PRINCIPE DIRECTEUR

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ARCHITECTURE CLAIRE                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  📱 FLUTTER MOBILE (@MazadPay)       🌐 REACT WEB ADMIN (@web)              │
│  ═════════════════════════════       ═══════════════════════════            │
│                                                                             │
│  Users (Bidders/Sellers)             Super Admin / Moderators                 │
│         ↓                                    ↓                              │
│  Endpoints PUBLICS                   Endpoints ADMIN                        │
│  /api/v1/* (Authentifié JWT)         /api/v1/admin/* (JWT + Role)           │
│         ↓                                    ↓                              │
│         └──────────────────┬─────────────────┘                              │
│                            ↓                                                │
│                     🔧 BACKEND GO (@backend)                                │
│                    ═══════════════════════                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. ENDPOINTS PUBLICS (Flutter Mobile)

### 1.1 AUTHENTICATION (`/auth/*`)
```http
# Authentification utilisateur
POST /api/v1/auth/login                    ✅ Existe
POST /api/v1/auth/register                 ✅ Existe
POST /api/v1/auth/verify-otp               ✅ Existe
POST /api/v1/auth/resend-otp               ✅ Existe
POST /api/v1/auth/forgot-password          ✅ Existe
POST /api/v1/auth/reset-password           ✅ Existe
POST /api/v1/auth/logout                     ✅ Existe
POST /api/v1/auth/refresh-token            ✅ Existe
```

### 1.2 USER PROFILE (`/users/*`)
```http
# Profil utilisateur (Authentifié)
GET    /api/v1/users/me                    ✅ Existe
PUT    /api/v1/users/me                    ✅ Existe
POST   /api/v1/users/me/avatar             ✅ Existe
PUT    /api/v1/users/me/password           ✅ Existe
PUT    /api/v1/users/me/preferences        ✅ Existe
GET    /api/v1/users/me/stats             ❌ MANQUANT - Stats personnelles
GET    /api/v1/users/me/activity           ❌ MANQUANT - Historique activité
```

### 1.3 FAVORITES (`/favorites/*`)
```http
# Gestion favoris (Authentifié)
GET    /api/v1/favorites                   ✅ Existe (via user_api.dart)
POST   /api/v1/favorites/:auctionId        ✅ Existe
DELETE /api/v1/favorites/:auctionId        ✅ Existe
```

### 1.4 AUCTIONS - PUBLIC (`/auctions/*`)
```http
# Enchères publiques (Partiellement public)
GET    /api/v1/auctions                    ✅ Existe - Liste publique
GET    /api/v1/auctions/:id              ✅ Existe - Détails publique
GET    /api/v1/auctions/:id/bids         ✅ Existe - Historique bids
GET    /api/v1/auctions/:id/seller-contact  ✅ Existe
POST   /api/v1/auctions/:id/view         ✅ Existe - Incrémenter vues

# Création d'enchère (Authentifié)
POST   /api/v1/auctions                    ✅ Existe
PUT    /api/v1/auctions/:id              ❌ MANQUANT - Modifier mon enchère
DELETE /api/v1/auctions/:id              ❌ MANQUANT - Supprimer mon enchère

# Enchérir (Authentifié)
POST   /api/v1/auctions/:id/bids         ✅ Existe
GET    /api/v1/auctions/:id/bid-status    ❌ MANQUANT - Statut de ma bid
```

### 1.5 MY AUCTIONS & BIDS (`/my/*`) - NOUVEAUX ENDPOINTS PROPOSÉS
```http
# Mes enchères et bids regroupés
GET    /api/v1/my/auctions               ❌ MANQUANT (actuel: /users/me/auctions)
GET    /api/v1/my/auctions/active        ❌ MANQUANT
GET    /api/v1/my/auctions/ended         ❌ MANQUANT
GET    /api/v1/my/auctions/pending        ❌ MANQUANT - En attente validation

GET    /api/v1/my/bids                   ❌ MANQUANT (actuel: /users/me/bids)
GET    /api/v1/my/bids/active            ❌ MANQUANT
GET    /api/v1/my/bids/won               ❌ MANQUANT (actuel: /users/me/winnings)
GET    /api/v1/my/bids/lost              ❌ MANQUANT

GET    /api/v1/my/watchlist             ❌ MANQUANT - Enchères suivies
```

### 1.6 WALLET (`/wallet/*`)
```http
# Portefeuille (Authentifié)
GET    /api/v1/wallet/balance            ✅ Existe (via /users/wallet)
GET    /api/v1/wallet/transactions       ✅ Existe
GET    /api/v1/wallet/transactions/:id   ✅ Existe

POST   /api/v1/wallet/deposit           ✅ Existe
POST   /api/v1/wallet/withdraw           ✅ Existe
POST   /api/v1/wallet/deposit/:id/receipt  ✅ Existe - Upload reçu

# Méthodes de paiement
GET    /api/v1/wallet/payment-methods    ❌ MANQUANT
POST   /api/v1/wallet/payment-methods    ❌ MANQUANT
DELETE /api/v1/wallet/payment-methods/:id  ❌ MANQUANT
```

### 1.7 NOTIFICATIONS (`/notifications/*`)
```http
# Notifications (Authentifié)
GET    /api/v1/notifications             ✅ Existe
PUT    /api/v1/notifications/:id/read    ✅ Existe
PUT    /api/v1/notifications/read-all    ✅ Existe
GET    /api/v1/notifications/settings    ✅ Existe
PUT    /api/v1/notifications/settings    ✅ Existe

# Préférences de notification
GET    /api/v1/notifications/preferences  ❌ MANQUANT - Détail par type
PUT    /api/v1/notifications/preferences/:type  ❌ MANQUANT
```

### 1.8 CATEGORIES & LOCATIONS - PUBLIC
```http
# Données de référence publiques
GET    /api/v1/categories                ✅ Existe
GET    /api/v1/categories/:id           ❌ MANQUANT - Détails catégorie
GET    /api/v1/categories/:id/auctions  ❌ MANQUANT - Enchères par catégorie

GET    /api/v1/locations                ✅ Existe
GET    /api/v1/locations/:id           ❌ MANQUANT
GET    /api/v1/locations/:id/auctions  ❌ MANQUANT - Enchères par location

GET    /api/v1/countries               ✅ Existe
```

### 1.9 REQUESTS - UTILISATEUR (`/requests/*`)
```http
# Demandes de l'utilisateur (Authentifié)
GET    /api/v1/my/requests              ❌ MANQUANT - Mes demandes enchères
GET    /api/v1/my/requests/:id          ❌ MANQUANT
POST   /api/v1/requests/auctions        ✅ Existe
POST   /api/v1/requests/banners         ✅ Existe - Seulement pour users avec privilèges
```

### 1.10 MESSAGING - NOUVEAU MODULE PROPOSÉ
```http
# Système de messagerie entre utilisateurs
GET    /api/v1/conversations            ❌ MANQUANT
POST   /api/v1/conversations            ❌ MANQUANT - Démarrer conversation
GET    /api/v1/conversations/:id        ❌ MANQUANT
POST   /api/v1/conversations/:id/messages  ❌ MANQUANT
PUT    /api/v1/conversations/:id/read    ❌ MANQUANT
DELETE /api/v1/conversations/:id        ❌ MANQUANT

# Contact vendeur pour une enchère spécifique
POST   /api/v1/auctions/:id/contact     ❌ MANQUANT
```

### 1.11 RATINGS & REVIEWS - NOUVEAU MODULE PROPOSÉ
```http
# Système d'évaluation
POST   /api/v1/auctions/:id/rate        ❌ MANQUANT - Noter une transaction
GET    /api/v1/users/:id/ratings        ❌ MANQUANT - Voir ratings vendeur
GET    /api/v1/my/ratings               ❌ MANQUANT - Mes ratings
```

### 1.12 KYC - UTILISATEUR (`/kyc/*`)
```http
# KYC pour utilisateurs
GET    /api/v1/kyc/status               ✅ Existe (via /users/kyc)
POST   /api/v1/kyc/submit               ✅ Existe
GET    /api/v1/kyc/documents            ❌ MANQUANT - Mes documents
DELETE /api/v1/kyc/documents/:id        ❌ MANQUANT
```

---

## 2. ENDPOINTS ADMIN (React Web Admin)

### 2.1 AUTHENTIFICATION ADMIN
```http
# Login spécifique admin
POST   /api/v1/admin/auth/login         ✅ Existe
POST   /api/v1/admin/auth/logout        ✅ Existe
GET    /api/v1/admin/auth/me            ✅ Existe
POST   /api/v1/admin/auth/refresh       ✅ Existe
```

### 2.2 DASHBOARD STATS
```http
# Statistiques globales pour dashboard
GET    /api/v1/admin/dashboard/stats              ✅ Existe
GET    /api/v1/admin/dashboard/revenue            ✅ Existe
GET    /api/v1/admin/dashboard/auctions           ✅ Existe
GET    /api/v1/admin/dashboard/users              ✅ Existe
GET    /api/v1/admin/dashboard/activity           ❌ MANQUANT - Activité récente
GET    /api/v1/admin/dashboard/realtime         ❌ MANQUANT - WebSocket stats
```

### 2.3 USERS MANAGEMENT
```http
# Gestion utilisateurs complète
GET    /api/v1/admin/users                      ✅ Existe
GET    /api/v1/admin/users/:id                  ✅ Existe
PUT    /api/v1/admin/users/:id/status           ✅ Existe (ban/unban)
DELETE /api/v1/admin/users/:id                  ✅ Existe
GET    /api/v1/admin/users/:id/auctions         ❌ MANQUANT - Enchères user
GET    /api/v1/admin/users/:id/transactions     ❌ MANQUANT - Transactions user
GET    /api/v1/admin/users/:id/activity         ❌ MANQUANT - Logs activité

# KYC Management
GET    /api/v1/admin/kyc/pending                ✅ Existe
GET    /api/v1/admin/kyc/:id                    ✅ Existe
PUT    /api/v1/admin/kyc/:id/approve            ✅ Existe
PUT    /api/v1/admin/kyc/:id/reject             ✅ Existe
GET    /api/v1/admin/kyc/stats                  ❌ MANQUANT
```

### 2.4 AUCTIONS MODERATION
```http
# Modération enchères
GET    /api/v1/admin/auctions                   ✅ Existe
GET    /api/v1/admin/auctions/pending           ✅ Existe - En attente
GET    /api/v1/admin/auctions/reported          ✅ Existe - Signalées
GET    /api/v1/admin/auctions/featured          ❌ MANQUANT - Mises en avant

PUT    /api/v1/admin/auctions/:id/approve       ✅ Existe
PUT    /api/v1/admin/auctions/:id/reject        ✅ Existe
PUT    /api/v1/admin/auctions/:id/feature       ❌ MANQUANT
PUT    /api/v1/admin/auctions/:id/unfeature     ❌ MANQUANT
DELETE /api/v1/admin/auctions/:id               ✅ Existe

# Batch operations
POST   /api/v1/admin/auctions/bulk-approve      ❌ MANQUANT
POST   /api/v1/admin/auctions/bulk-reject       ❌ MANQUANT
POST   /api/v1/admin/auctions/bulk-delete       ❌ MANQUANT
```

### 2.5 TRANSACTIONS MANAGEMENT
```http
# Gestion transactions
GET    /api/v1/admin/transactions               ✅ Existe
GET    /api/v1/admin/transactions/pending       ❌ MANQUANT - En attente
GET    /api/v1/admin/transactions/:id           ❌ MANQUANT - Détails

# Approbation manuelle
PUT    /api/v1/admin/transactions/:id/approve    ❌ MANQUANT
PUT    /api/v1/admin/transactions/:id/reject    ❌ MANQUANT
PUT    /api/v1/admin/transactions/:id/refund    ❌ MANQUANT

# Withdrawals
GET    /api/v1/admin/withdrawals/pending        ❌ MANQUANT
PUT    /api/v1/admin/withdrawals/:id/process    ❌ MANQUANT
```

### 2.6 CATEGORIES & LOCATIONS CRUD
```http
# Gestion catégories
GET    /api/v1/admin/categories                 ❌ MANQUANT
POST   /api/v1/admin/categories                 ❌ MANQUANT
PUT    /api/v1/admin/categories/:id             ❌ MANQUANT
DELETE /api/v1/admin/categories/:id             ❌ MANQUANT
PUT    /api/v1/admin/categories/:id/order       ❌ MANQUANT - Réordonner

# Gestion locations
GET    /api/v1/admin/locations                  ❌ MANQUANT
POST   /api/v1/admin/locations                  ❌ MANQUANT
PUT    /api/v1/admin/locations/:id              ❌ MANQUANT
DELETE /api/v1/admin/locations/:id              ❌ MANQUANT
```

### 2.7 BANNERS MANAGEMENT
```http
# Gestion bannières
GET    /api/v1/admin/banners                    ❌ MANQUANT
POST   /api/v1/admin/banners                    ❌ MANQUANT
PUT    /api/v1/admin/banners/:id                ❌ MANQUANT
DELETE /api/v1/admin/banners/:id                ❌ MANQUANT
PUT    /api/v1/admin/banners/:id/toggle         ❌ MANQUANT - Activer/Désactiver

# Banner requests
GET    /api/v1/admin/banners/requests           ❌ MANQUANT
PUT    /api/v1/admin/banners/requests/:id/approve  ❌ MANQUANT
PUT    /api/v1/admin/banners/requests/:id/reject   ❌ MANQUANT
```

### 2.8 NOTIFICATIONS BROADCAST
```http
# Notifications admin
POST   /api/v1/admin/notifications/broadcast    ❌ MANQUANT - Envoi masse
POST   /api/v1/admin/notifications/user         ❌ MANQUANT - Envoi ciblé
GET    /api/v1/admin/notifications/history      ❌ MANQUANT - Historique envois
```

### 2.9 REPORTS & MODERATION
```http
# Signalements
GET    /api/v1/admin/reports                    ✅ Existe
GET    /api/v1/admin/reports/:id                ❌ MANQUANT
PUT    /api/v1/admin/reports/:id/resolve        ✅ Existe
GET    /api/v1/admin/reports/stats               ❌ MANQUANT

# Blocked phones
GET    /api/v1/admin/blocked-phones             ✅ Existe
POST   /api/v1/admin/blocked-phones             ✅ Existe
DELETE /api/v1/admin/blocked-phones/:id         ✅ Existe
```

### 2.10 CONTENT MANAGEMENT (CMS)
```http
# FAQ / Tutoriels / Pages statiques
GET    /api/v1/admin/faq                        ❌ MANQUANT
POST   /api/v1/admin/faq                        ❌ MANQUANT
PUT    /api/v1/admin/faq/:id                    ❌ MANQUANT
DELETE /api/v1/admin/faq/:id                    ❌ MANQUANT

GET    /api/v1/admin/tutorials                  ❌ MANQUANT
POST   /api/v1/admin/tutorials                  ❌ MANQUANT
PUT    /api/v1/admin/tutorials/:id              ❌ MANQUANT
DELETE /api/v1/admin/tutorials/:id              ❌ MANQUANT

# Terms & Privacy
GET    /api/v1/admin/content/terms             ❌ MANQUANT
PUT    /api/v1/admin/content/terms              ❌ MANQUANT
GET    /api/v1/admin/content/privacy           ❌ MANQUANT
PUT    /api/v1/admin/content/privacy           ❌ MANQUANT
```

### 2.11 SYSTEM SETTINGS
```http
# Configuration système
GET    /api/v1/admin/settings                   ✅ Existe
PUT    /api/v1/admin/settings                   ✅ Existe
GET    /api/v1/admin/settings/fees              ❌ MANQUANT - Commission/frais
PUT    /api/v1/admin/settings/fees              ❌ MANQUANT

# Auction settings
GET    /api/v1/admin/settings/auction-rules       ❌ MANQUANT
PUT    /api/v1/admin/settings/auction-rules     ❌ MANQUANT
```

### 2.12 AUDIT LOGS
```http
# Logs d'audit
GET    /api/v1/admin/audit-logs                 ✅ Existe
GET    /api/v1/admin/audit-logs/:id            ❌ MANQUANT
GET    /api/v1/admin/audit-logs/user/:id        ❌ MANQUANT
```

### 2.13 ADMIN INVITATIONS
```http
# Gestion des admins
GET    /api/v1/admin/invitations                ✅ Existe
POST   /api/v1/admin/invitations                ✅ Existe
PUT    /api/v1/admin/invitations/:id/revoke      ❌ MANQUANT
GET    /api/v1/admin/admins                     ✅ Existe - Liste admins
PUT    /api/v1/admin/admins/:id/role             ❌ MANQUANT - Changer rôle
DELETE /api/v1/admin/admins/:id                 ❌ MANQUANT - Retirer admin
```

---

## 3. ENDPOINTS MANQUANTS - PRIORITÉS

### 🔴 CRITIQUE (Cette semaine)

#### Pour Mobile Flutter:
```
1. GET /api/v1/my/auctions              - Consolidation endpoint
2. GET /api/v1/my/bids                 - Consolidation endpoint
3. PUT /api/v1/auctions/:id            - Modifier son enchère
4. DELETE /api/v1/auctions/:id          - Supprimer son enchère
5. GET /api/v1/wallet/payment-methods   - Gestion méthodes paiement
```

#### Pour Admin Web:
```
1. POST /api/v1/admin/categories        - CRUD complet catégories
2. PUT /api/v1/admin/categories/:id
3. DELETE /api/v1/admin/categories/:id
4. POST /api/v1/admin/locations         - CRUD complet locations
5. PUT /api/v1/admin/locations/:id
6. DELETE /api/v1/admin/locations/:id
```

### 🟡 IMPORTANT (Cette semaine)

#### Pour Mobile:
```
1. GET /api/v1/my/auctions/pending      - Enchères en attente
2. GET /api/v1/my/bids/won             - Bids gagnés (remplace winnings)
3. GET /api/v1/my/bids/lost             - Bids perdus
4. GET /api/v1/auctions/:id/bid-status  - Statut bid en cours
5. POST /api/v1/auctions/:id/contact    - Contacter vendeur
```

#### Pour Admin:
```
1. GET /api/v1/admin/banners            - Gestion bannières
2. POST /api/v1/admin/banners
3. PUT /api/v1/admin/banners/:id
4. DELETE /api/v1/admin/banners/:id
5. POST /api/v1/admin/notifications/broadcast
6. GET /api/v1/admin/transactions/pending
7. PUT /api/v1/admin/transactions/:id/approve
```

### 🟢 MOYEN (Semaine prochaine)

#### Pour Mobile:
```
1. GET /api/v1/my/watchlist             - Liste de suivi
2. GET /api/v1/conversations            - Messagerie
3. POST /api/v1/conversations
4. POST /api/v1/auctions/:id/rate       - Rating
```

#### Pour Admin:
```
1. GET /api/v1/admin/faq               - CMS FAQ
2. POST /api/v1/admin/faq
3. GET /api/v1/admin/tutorials           - CMS Tutoriels
4. POST /api/v1/admin/tutorials
5. GET /api/v1/admin/settings/fees       - Config frais
```

---

## 4. STRUCTURE DES RÉPONSES API

### Format Standard (Déjà implémenté)
```json
{
  "success": true,
  "data": { ... },
  "message": "Opération réussie",
  "error": null
}
```

### Pagination (Standardiser)
```json
{
  "success": true,
  "data": {
    "items": [ ... ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "totalPages": 8
    }
  }
}
```

### Erreurs (Standardiser)
```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Solde insuffisant",
    "details": { ... }
  }
}
```

---

## 5. MIDDLEWARE & SÉCURITÉ

### Authentification JWT
```go
// Middleware existant
authMiddleware := middleware.Auth(jwtSecret)

// À implémenter
adminMiddleware := middleware.AdminOnly()      // Role >= Admin
superAdminMiddleware := middleware.SuperAdminOnly()  // Role = SuperAdmin
moderatorMiddleware := middleware.ModeratorOrAbove() // Role >= Moderator
```

### Rate Limiting
```go
// Par endpoint
/api/v1/auth/*          : 5 req/minute
/api/v1/auctions/*       : 100 req/minute  
/api/v1/admin/*          : 200 req/minute
```

---

## 6. RÉSUMÉ PAR INTERFACE

### 📱 Flutter Mobile (@MazadPay)
**Besoins:**
- Endpoints simples et rapides
- Support offline avec cache
- Notifications push
- Optimisés pour mobile (pagination, lazy loading)

**Endpoints prioritaires:** 23 nouveaux à créer

### 🌐 React Web (@web)  
**Besoins:**
- Dashboards avec stats
- Tables avec filtres avancés
- Export données
- Modération en temps réel

**Endpoints prioritaires:** 35 nouveaux à créer

### 🔧 Backend Go (@backend)
**Structure recommandée:**
```
internal/
├── handlers/
│   ├── public/          # Endpoints mobile
│   │   ├── user_handler.go
│   │   ├── auction_handler.go
│   │   ├── wallet_handler.go
│   │   └── ...
│   └── admin/             # Endpoints web admin
│       ├── admin_user_handler.go
│       ├── admin_auction_handler.go
│       ├── admin_stats_handler.go
│       └── ...
├── middleware/
│   ├── auth.go
│   ├── admin_auth.go      # JWT + Role checking
│   └── rate_limit.go
└── routes/
    ├── public_routes.go
    └── admin_routes.go
```

---

## 7. PROCHAINES ACTIONS

### Immédiat (Backend)
1. Créer namespace `/admin/*` séparé
2. Implémenter middleware RBAC
3. Ajouter CRUD catégories/locations (admin)
4. Ajouter endpoints `/my/*` (mobile)
5. Standardiser format pagination

### Court terme
1. WebSocket pour temps réel (bids, notifications)
2. Système de messagerie
3. Module CMS complet
4. Analytics avancés

---

**Document mis à jour:** 22 Avril 2026
