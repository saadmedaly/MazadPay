# 📋 Rapport Complet des Endpoints Backend - MazadPay

**Date :** 22 Avril 2026  
**Version API :** v1  
**Base URL :** `http://localhost:8082/v1/api`

---

## 📊 Résumé

| Catégorie | Nombre d'Endpoints |
|-----------|-------------------|
| Authentification | 8 |
| Enchères (Auctions) | 14 |
| Utilisateurs (Users) | 15 |
| Admin | 32 |
| Bannières (Banners) | 7 |
| Contenu (Content) | 8 |
| Notifications | 8 |
| Demandes (Requests) | 13 |
| WebSocket | 2 |
| **TOTAL** | **107** |

---

## 🔐 Module Authentification

**Base Path :** `/auth`

| Méthode | Endpoint | Auth | Rate Limit | Paramètres | Description |
|---------|----------|------|------------|-----------|-------------|
| POST | `/auth/register` | ❌ | ✅ Phone | `phone`, `pin`, `full_name`, `email`, `city` | Inscription utilisateur |
| POST | `/auth/login` | ❌ | ✅ Phone | `phone`, `pin` | Connexion utilisateur |
| POST | `/auth/otp/send` | ❌ | ✅ Phone | `phone`, `purpose` | Envoyer OTP (register/reset_password) |
| POST | `/auth/otp/verify` | ❌ | ✅ Phone | `phone`, `code`, `purpose` | Vérifier OTP |
| POST | `/auth/reset-password` | ❌ | ✅ Phone | `phone`, `new_pin`, `otp_code` | Réinitialiser mot de passe |
| POST | `/auth/register-admin` | ❌ | ✅ Phone | `phone`, `pin`, `full_name`, `email`, `invitation_code` | Inscription admin avec invitation |
| POST | `/auth/logout` | ✅ JWT | ❌ | - | Déconnexion |
| PUT | `/auth/change-password` | ✅ JWT | ❌ | `old_pin`, `new_pin` | Changer mot de passe |

**Middlewares :**
- `RateLimitByPhone` : Limite par numéro de téléphone
- `JWT` : Authentification JWT requise

---

## 🏆 Module Enchères (Auctions)

**Base Path :** `/auctions`

| Méthode | Endpoint | Auth | Paramètres | Description |
|---------|----------|------|-----------|-------------|
| GET | `/categories` | ❌ | Query: `?limit=50` | Lister les catégories |
| GET | `/locations` | ❌ | Query: `?country_id=uuid` | Lister les locations |
| GET | `/countries` | ❌ | - | Lister les pays |
| GET | `/locations/:countryId` | ❌ | Path: `countryId` | Locations par pays |
| GET | `/auctions` | ❌ | Query: `?page=1&limit=20&category_id=&location_id=&status=` | Lister les enchères |
| GET | `/auctions/:id` | ❌ | Path: `id` | Détails d'une enchère |
| POST | `/auctions/:id/view` | ❌ | Path: `id` | Incrémenter vues |
| GET | `/report-reasons` | ❌ | - | Raisons de signalement |
| GET | `/auctions/:id/bids` | ❌ | Path: `id` | Historique des enchères |
| POST | `/auctions` | ✅ JWT | Body: `title`, `description`, `starting_price`, `category_id`, `location`, `end_time`, `images[]` | Créer une enchère |
| POST | `/auctions/:id/report` | ✅ JWT | Path: `id`, Body: `reason`, `description` | Signaler une enchère |
| POST | `/auctions/:id/images` | ✅ JWT | Path: `id`, Body: `images[]` | Ajouter images |
| POST | `/auctions/:id/buy-now` | ✅ JWT | Path: `id` | Acheter maintenant |
| POST | `/auctions/:id/cancel` | ✅ JWT | Path: `id` | Annuler enchère |
| POST | `/auctions/:id/relist` | ✅ JWT | Path: `id` | Remettre en vente |
| POST | `/auctions/:id/extend` | ✅ JWT | Path: `id`, Body: `duration` | Prolonger enchère |
| POST | `/auctions/:id/bids` | ✅ JWT | Path: `id`, Body: `amount` | Placer une enchère |
| GET | `/auctions/:id/seller-contact` | ✅ JWT | Path: `id` | Contact vendeur |

---

## 👤 Module Utilisateurs (Users)

**Base Path :** `/users`

| Méthode | Endpoint | Auth | Paramètres | Description |
|---------|----------|------|-----------|-------------|
| GET | `/users/me` | ✅ JWT | - | Profil utilisateur |
| PUT | `/users/me` | ✅ JWT | Body: `full_name`, `email`, `phone` | Mettre à jour profil |
| POST | `/users/me/avatar` | ✅ JWT | Body: `avatar_url` ou multipart | Mettre à jour avatar |
| PUT | `/users/me/language` | ✅ JWT | Body: `language` (ar/fr/en) | Changer langue |
| PUT | `/users/me/notification-prefs` | ✅ JWT | Body: `preferences` | Préférences notifications |
| GET | `/users/me/favorites` | ✅ JWT | - | Favoris |
| POST | `/users/me/favorites/:auction_id` | ✅ JWT | Path: `auction_id` | Ajouter favori |
| DELETE | `/users/me/favorites/:auction_id` | ✅ JWT | Path: `auction_id` | Supprimer favori |
| GET | `/users/me/auctions` | ✅ JWT | Query: `?status=` | Mes enchères |
| GET | `/users/me/bids` | ✅ JWT | - | Mes enchères placées |
| GET | `/users/me/winnings` | ✅ JWT | - | Mes gains |
| GET | `/users/wallet` | ✅ JWT | - | Solde wallet |
| POST | `/users/wallet/deposit` | ✅ JWT | Body: `amount`, `payment_method` | Dépôt |
| POST | `/users/wallet/transactions/:id/receipt` | ✅ JWT | Path: `id`, Body: `receipt_image` | Upload reçu |
| POST | `/users/wallet/withdraw` | ✅ JWT | Body: `amount`, `bank_details` | Retrait |
| GET | `/users/wallet/transactions` | ✅ JWT | Query: `?page=&limit=` | Historique transactions |
| GET | `/users/wallet/transactions/:id` | ✅ JWT | Path: `id` | Détails transaction |
| GET | `/users/kyc` | ✅ JWT | - | Statut KYC |
| POST | `/users/kyc` | ✅ JWT | Body: `id_card_front`, `id_card_back`, `selfie` | Soumettre KYC |

---

## 👑 Module Admin

**Base Path :** `/admin`

**Middlewares :** `JWT` + `AdminOnly` (sauf indication contraire)

| Méthode | Endpoint | Auth | Paramètres | Description |
|---------|----------|------|-----------|-------------|
| **Dashboard** |
| GET | `/admin/dashboard/stats` | ✅ JWT+Admin | - | Statistiques dashboard |
| GET | `/admin/dashboard/revenue-chart` | ✅ JWT+Admin | Query: `?days=30` | Graphique revenus |
| GET | `/admin/dashboard/activity` | ✅ JWT+Admin | Query: `?limit=20` | Feed d'activité |
| **Utilisateurs** |
| GET | `/admin/users` | ✅ JWT+Admin | Query: `?page=&limit=&search=` | Lister utilisateurs |
| GET | `/admin/users/:id` | ✅ JWT+Admin | Path: `id` | Détails utilisateur |
| GET | `/admin/users/:id/auctions` | ✅ JWT+Admin | Path: `id` | Enchères utilisateur |
| GET | `/admin/users/:id/transactions` | ✅ JWT+Admin | Path: `id` | Transactions utilisateur |
| POST | `/admin/invitations` | ✅ JWT+Admin | Body: `role` (admin) | Générer invitation |
| PUT | `/admin/users/:id/block` | ✅ JWT+Admin | Path: `id`, Body: `reason` | Bloquer utilisateur |
| DELETE | `/admin/users/:id` | ✅ JWT+SuperAdmin | Path: `id` | Supprimer utilisateur |
| **Enchères** |
| GET | `/admin/auctions` | ✅ JWT+Admin | Query: `?status=&category_id=` | Lister enchères |
| PUT | `/admin/auctions/:id/validate` | ✅ JWT+Admin | Path: `id`, Body: `status` | Valider enchère |
| PUT | `/admin/auctions/:id` | ✅ JWT+Admin | Path: `id`, Body: `auction data` | Mettre à jour enchère |
| DELETE | `/admin/auctions/:id` | ✅ JWT+Admin | Path: `id` | Supprimer enchère |
| **Transactions** |
| GET | `/admin/transactions` | ✅ JWT+Admin | Query: `?status=&user_id=` | Lister transactions |
| PUT | `/admin/transactions/:id/validate` | ✅ JWT+Admin | Path: `id`, Body: `status` | Valider transaction |
| **Signalements** |
| GET | `/admin/reports` | ✅ JWT+Admin | Query: `?status=` | Lister signalements |
| PUT | `/admin/reports/:id/review` | ✅ JWT+Admin | Path: `id`, Body: `action` | Revoir signalement |
| **KYC** |
| GET | `/admin/kyc` | ✅ JWT+Admin | Query: `?status=` | Lister KYC |
| PUT | `/admin/kyc/:user_id` | ✅ JWT+Admin | Path: `user_id`, Body: `status` | Revoir KYC |
| **Catégories** |
| POST | `/admin/categories` | ✅ JWT+Admin | Body: `name`, `description` | Créer catégorie |
| PUT | `/admin/categories/:id` | ✅ JWT+Admin | Path: `id`, Body: `name`, `description` | Mettre à jour catégorie |
| DELETE | `/admin/categories/:id` | ✅ JWT+Admin | Path: `id` | Supprimer catégorie |
| **Locations** |
| POST | `/admin/locations` | ✅ JWT+Admin | Body: `name`, `country_id` | Créer location |
| PUT | `/admin/locations/:id` | ✅ JWT+Admin | Path: `id`, Body: `name` | Mettre à jour location |
| DELETE | `/admin/locations/:id` | ✅ JWT+Admin | Path: `id` | Supprimer location |
| **Pays** |
| GET | `/admin/countries` | ✅ JWT+Admin | - | Lister pays |
| POST | `/admin/countries` | ✅ JWT+Admin | Body: `name`, `code`, `flag` | Créer pays |
| PUT | `/admin/countries/:id` | ✅ JWT+Admin | Path: `id`, Body: `name`, `code` | Mettre à jour pays |
| DELETE | `/admin/countries/:id` | ✅ JWT+Admin | Path: `id` | Supprimer pays |
| **Téléphones bloqués** |
| GET | `/admin/blocked-phones` | ✅ JWT+Admin | - | Lister téléphones bloqués |
| POST | `/admin/blocked-phones` | ✅ JWT+Admin | Body: `phone`, `reason` | Bloquer téléphone |
| DELETE | `/admin/blocked-phones/:phone` | ✅ JWT+Admin | Path: `phone` | Débloquer téléphone |
| **Settings** |
| GET | `/admin/settings` | ✅ JWT+Admin | - | Lister settings |
| PUT | `/admin/settings/:key` | ✅ JWT+Admin | Path: `key`, Body: `value` | Mettre à jour setting |

---

## 🖼️ Module Bannières (Banners)

| Méthode | Endpoint | Auth | Paramètres | Description |
|---------|----------|------|-----------|-------------|
| GET | `/banners` | ❌ | Query: `?active=true` | Lister bannières publiques |
| POST | `/banners/request` | ✅ JWT | Body: `title`, `description`, `image_url`, `link_url` | Demande de création |
| POST | `/banners` | ✅ JWT+Admin | Body: `title`, `description`, `image_url`, `link_url`, `is_active` | Créer bannière (admin) |
| GET | `/admin/banners` | ✅ JWT+Admin | - | Lister bannières admin |
| GET | `/admin/banners/all` | ✅ JWT+Admin | - | Toutes les bannières (debug) |
| POST | `/admin/banners` | ✅ JWT+Admin | Body: banner data | Créer bannière |
| PUT | `/admin/banners/:id/toggle` | ✅ JWT+Admin | Path: `id` | Activer/désactiver |
| PUT | `/admin/banners/:id` | ✅ JWT+Admin | Path: `id`, Body: banner data | Mettre à jour |
| DELETE | `/admin/banners/:id` | ✅ JWT+Admin | Path: `id` | Supprimer |

---

## 📄 Module Contenu (Content)

| Méthode | Endpoint | Auth | Paramètres | Description |
|---------|----------|------|-----------|-------------|
| GET | `/faq` | ❌ | Query: `?lang=ar` | FAQ publique |
| GET | `/tutorials` | ❌ | Query: `?lang=ar` | Tutoriels publics |
| GET | `/about` | ❌ | - | À propos |
| GET | `/privacy-policy` | ❌ | - | Politique de confidentialité |
| GET | `/admin/faq` | ✅ JWT+Admin | - | Lister FAQ admin |
| POST | `/admin/faq` | ✅ JWT+Admin | Body: `question`, `answer`, `lang` | Créer FAQ |
| PUT | `/admin/faq/:id` | ✅ JWT+Admin | Path: `id`, Body: FAQ data | Mettre à jour FAQ |
| DELETE | `/admin/faq/:id` | ✅ JWT+Admin | Path: `id` | Supprimer FAQ |
| GET | `/admin/tutorials` | ✅ JWT+Admin | - | Lister tutoriels admin |
| POST | `/admin/tutorials` | ✅ JWT+Admin | Body: `title`, `content`, `lang` | Créer tutoriel |
| PUT | `/admin/tutorials/:id` | ✅ JWT+Admin | Path: `id`, Body: tutorial data | Mettre à jour |
| DELETE | `/admin/tutorials/:id` | ✅ JWT+Admin | Path: `id` | Supprimer |

---

## 🔔 Module Notifications

| Méthode | Endpoint | Auth | Paramètres | Description |
|---------|----------|------|-----------|-------------|
| POST | `/notifications/token` | ✅ JWT | Body: `fcm_token`, `device_id`, `platform` | Sauvegarder token FCM |
| POST | `/notifications/push-tokens` | ✅ JWT | Body: `fcm_token`, `device_id`, `platform` | Sauvegarder token (alias) |
| GET | `/notifications` | ✅ JWT | Query: `?limit=20` | Lister notifications utilisateur |
| PUT | `/notifications/:id/read` | ✅ JWT | Path: `id` | Marquer comme lue |
| PUT | `/notifications/read-all` | ✅ JWT | - | Tout marquer comme lu |
| GET | `/admin/notifications` | ✅ JWT | Query: `?status=&limit=50` | Lister notifications admin |
| POST | `/admin/notifications/send` | ✅ JWT | Body: `user_id`, `title`, `body`, `type`, `data` | Envoyer notification |
| POST | `/admin/notifications/broadcast` | ✅ JWT | Body: `title`, `body`, `type`, `data` | Broadcast notification |
| PUT | `/admin/notifications/:id/read` | ✅ JWT | Path: `id` | Marquer comme lue (admin) |
| PUT | `/admin/notifications/read-all` | ✅ JWT | - | Tout marquer comme lu (admin) |
| DELETE | `/admin/notifications/:id` | ✅ JWT | Path: `id` | Supprimer notification |
| GET | `/admin/notifications/templates` | ✅ JWT | Query: `?lang=ar` | Templates de notifications |

---

## 📝 Module Demandes (Requests)

**Rate Limit :** 5 requêtes/heure par utilisateur

| Méthode | Endpoint | Auth | Rate Limit | Paramètres | Description |
|---------|----------|------|------------|-----------|-------------|
| **Demandes utilisateurs** |
| POST | `/requests/auctions` | ✅ JWT | ✅ (5/heure) | Body: `title`, `description`, `starting_price`, `category_id`, `location` | Demande enchère |
| POST | `/requests/banners` | ✅ JWT | ✅ (5/heure) | Body: `title`, `description`, `image_url`, `link_url` | Demande bannière |
| GET | `/requests/auctions/my` | ✅ JWT | ❌ | Query: `?status=` | Mes demandes enchères |
| GET | `/requests/banners/my` | ✅ JWT | ❌ | Query: `?status=` | Mes demandes bannières |
| **Gestion admin** |
| GET | `/admin/requests/auctions` | ✅ JWT+Admin | ❌ | Query: `?status=` | Lister demandes enchères |
| GET | `/admin/requests/auctions/:id` | ✅ JWT+Admin | ❌ | Path: `id` | Détails demande enchère |
| PUT | `/admin/requests/auctions/:id/review` | ✅ JWT+Admin | ❌ | Path: `id`, Body: `action` (approve/reject) | Revoir demande |
| DELETE | `/admin/requests/auctions/:id` | ✅ JWT+Admin | ❌ | Path: `id` | Supprimer demande |
| POST | `/admin/requests/auctions/bulk/review` | ✅ JWT+Admin | ❌ | Body: `ids[]`, `action` | Revoir en masse |
| POST | `/admin/requests/auctions/bulk/delete` | ✅ JWT+Admin | ❌ | Body: `ids[]` | Supprimer en masse |
| GET | `/admin/requests/banners` | ✅ JWT+Admin | ❌ | Query: `?status=` | Lister demandes bannières |
| GET | `/admin/requests/banners/:id` | ✅ JWT+Admin | ❌ | Path: `id` | Détails demande bannière |
| PUT | `/admin/requests/banners/:id/review` | ✅ JWT+Admin | ❌ | Path: `id`, Body: `action` | Revoir demande |
| DELETE | `/admin/requests/banners/:id` | ✅ JWT+Admin | ❌ | Path: `id` | Supprimer demande |
| POST | `/admin/requests/banners/bulk/review` | ✅ JWT+Admin | ❌ | Body: `ids[]`, `action` | Revoir en masse |
| POST | `/admin/requests/banners/bulk/delete` | ✅ JWT+Admin | ❌ | Body: `ids[]` | Supprimer en masse |
| **Audit logs** |
| GET | `/admin/requests/audit/logs` | ✅ JWT+Admin | ❌ | Query: `?user_id=&entity_type=&limit=` | Logs d'audit |
| GET | `/admin/requests/audit/logs/:entity_type/:entity_id` | ✅ JWT+Admin | ❌ | Path: `entity_type`, `entity_id` | Logs par entité |

---

## 🌐 WebSocket Endpoints

| Endpoint | Auth | Description |
|----------|------|-------------|
| `GET /ws/auction/:id` | ❌ | WebSocket enchère en temps réel |
| `GET /ws/admin` | ✅ JWT (query param `?token=`) | WebSocket admin en temps réel |

---

## 🔑 Codes d'authentification

| Code | Description |
|------|-------------|
| `Bearer <token>` | Header Authorization pour JWT |
| `?token=<jwt>` | Query param pour WebSocket admin |

---

## 📊 Codes de statut HTTP

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 429 | Too Many Requests (Rate Limit) |
| 500 | Internal Server Error |

---

## 🔒 Rôles utilisateurs

| Rôle | Description |
|------|-------------|
| `user` | Utilisateur standard |
| `admin` | Administrateur |
| `super_admin` | Super administrateur (accès complet) |

---

## 📝 Notes importantes

1. **Rate Limiting :**
   - Auth endpoints : limité par numéro de téléphone
   - Request endpoints : 5 requêtes/heure par utilisateur
   - Autres endpoints : pas de rate limiting explicite

2. **Pagination :**
   - La plupart des endpoints list utilisent `?page=1&limit=20`
   - Certains utilisent seulement `?limit=50`

3. **Filtrage :**
   - Les endpoints de liste acceptent souvent des query params pour filtrer (status, category_id, etc.)

4. **UUID :**
   - Tous les IDs dans les paths sont des UUIDs (ex: `/users/92d8f678-6f50-4489-b86f-cd6a92d6a33b`)

5. **Langues supportées :**
   - `ar` (Arabe - défaut)
   - `fr` (Français)
   - `en` (Anglais)

---

**Fin du rapport**
