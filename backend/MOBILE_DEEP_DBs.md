# Rapport d'Analyse: Interface Mobile vs Base de Données
**Date:** 2025-04-23  
**Projet:** MazadPay V2  
**Objectif:** Analyse approfondie de l'interface mobile et identification des champs/tables manquants ou incomplets dans la base de données

---

## Table des Matières
1. [Vue d'ensemble](#vue-densemble)
2. [Analyse des Pages Mobile](#analyse-des-pages-mobile)
3. [Comparaison Schema BD vs Interface Mobile](#comparaison-schema-bd-vs-interface-mobile)
4. [Champs/Tables Manquants](#champstables-manquants)
5. [Recommandations](#recommandations)

---

## Vue d'ensemble

### Pages Mobile Analysées
- **Authentification:** LoginPage, CreateProfilePage, PhoneRegistrationPage, OtpEntryPage, NewPasswordPage, SetPasswordPage
- **Compte:** AccountPage, AccountProfilePage
- **Enchères:** HomePage, AllAuctionsPage, AuctionDetailsPage, AuctionHistoryPage, AuctionWinnerPage, CreateAdFormPage, CreateAdStartPage, AdSuccessPage
- **Mes Activités:** MyAuctionsPage, MyWinningsPage, FavoritesPage
- **Portefeuille:** DepositPage, WithdrawPage, PaymentDetailsPage, PaymentSuccessPage
- **Services:** ServicesPage, DeliveryDetailsPage
- **Autres:** NotificationsPage, SupportPage, AboutMazadPayPage, PrivacyPolicyPage, TermsPage, LanguagePage, HowToBidPage, StartBiddingPage, OnboardingPage, SplashPage

### Tables Base de Données Existantes
- users, otp_verifications, categories, locations, countries, auctions, bids, auction_images, user_favorites
- wallets, transactions, wallet_holds, auction_payments
- notifications, service_requests, delivery_timeline
- faq_items, banners, app_ratings, tutorials, reports, kyc_verifications
- push_tokens, blocked_phones, password_reset_attempts
- auction_requests, banner_requests
- audit_logs, system_settings, admin_invitations

---

## Analyse des Pages Mobile

### 1. Authentification & Inscription

#### LoginPage
**Champs utilisés:**
- phone (numéro de téléphone)
- password/pin (code PIN de 4 chiffres)
- country_code (+222 pour Mauritanie)

**API Calls:**
- POST /auth/login { phone, pin }

#### CreateProfilePage
**Champs utilisés:**
- full_name (nom complet)

**API Calls:**
- POST /auth/register { phone, pin, full_name, email?, city? }

#### PhoneRegistrationPage
**Champs utilisés:**
- phone

**API Calls:**
- POST /auth/otp/send { phone, purpose: 'register' }

#### OtpEntryPage
**Champs utilisés:**
- otp_code (code OTP)

**API Calls:**
- POST /auth/otp/verify { phone, code, purpose }

#### NewPasswordPage
**Champs utilisés:**
- phone, new_pin, otp_code

**API Calls:**
- POST /auth/reset-password { phone, new_pin, otp_code }

### 2. Profil Utilisateur

#### AccountProfilePage
**Champs affichés:**
- profile_pic_url
- full_name
- phone
- email
- city
- language_pref
- notifications_enabled

**Actions disponibles:**
- Modifier photo de profil
- Modifier mot de passe
- Changer langue
- Activer/désactiver notifications

**API Calls manquants:**
- PUT /users/profile (mise à jour profil)
- PUT /users/avatar (upload photo)
- PUT /users/settings (préférences)

### 3. Enchères

#### AuctionDetailsPage
**Champs affichés:**
- id, title (title_ar/title_fr/title_en)
- description (description_ar/description_fr/description_en)
- images (List<String>)
- start_price, current_price, min_increment
- end_time (ends_at)
- bidder_count, views
- lot_number
- phone_contact
- is_user_highest_bidder
- manufacturer, fuel_type, transmission, year, mileage, model (détails véhicule)
- status
- is_favorite

**API Calls:**
- GET /auctions/{id}
- POST /auctions/{id}/view (incrémenter vues)
- POST /auctions/{id}/bids (placer enchère)

#### CreateAdFormPage
**Champs requis:**
- title
- description
- phone
- starting_price
- category (main category)
- sub_category
- location (city)
- images (List<String>, max 5)

**API Calls:**
- POST /auctions { title, description, starting_price, category, sub_category, location, images, phone }

**Problème:** L'API envoie `category` et `sub_category` comme strings, mais la BD utilise `category_id` (INT) et n'a pas de champ `sub_category`.

#### AuctionHistoryPage
**Champs affichés:**
- BidEntry: bidderName, phoneNumber, amount, timestamp, isWinner
- Statistiques: nombre d'enchérisseurs, nombre d'enchères

**API Calls:**
- GET /auctions/{id}/bids

**Problème:** La table `bids` n'a pas de champ `bidder_name` ou `phone_number`. Ces infos doivent être jointes avec la table `users`.

#### AuctionWinnerPage
**Champs affichés:**
- winning_amount
- auction_image
- payment_deadline
- winner_info

**API Calls:**
- GET /auctions/{id}/winner (endpoint manquant)

### 4. Portefeuille

#### DepositPage
**Champs utilisés:**
- payment_method (masrvi, bankily, sedad, click)
- amount

**API Calls:**
- POST /users/wallet/deposit { amount, payment_method }

#### WithdrawPage
**Champs utilisés:**
- amount
- method (bank_transfer, mobile_money)
- bank_details

**API Calls:**
- POST /users/wallet/withdraw { amount, method, bank_details }

#### PaymentDetailsPage
**Champs utilisés:**
- merchant_name
- merchant_phone
- amount
- order_reference
- status
- receipt_image

**API Calls:**
- POST /users/wallet/transactions/{id}/receipt (upload reçu)

**Problème:** La table `transactions` n'a pas de champ `receipt_image` séparé. Il y a `receipt_url` mais pas de champ pour l'image uploadée avant validation.

### 5. Services & Livraison

#### DeliveryDetailsPage
**Champs affichés:**
- tracking_number
- delivery_status
- pickup_location
- delivery_location
- driver_info (name, phone, photo)
- delivery_timeline (steps avec timestamps)

**API Calls:**
- GET /service-requests/{id}
- GET /service-requests/{id}/timeline

**Problème:** La table `service_requests` a `driver_id` mais pas de champs détaillés pour le driver. La table `delivery_timeline` existe mais manque certains champs potentiels.

### 6. Notifications

#### NotificationsPage
**Champs affichés:**
- type (bid, win, payment, system)
- title
- message/body
- created_at
- is_read
- reference_id
- reference_type

**API Calls:**
- GET /notifications
- PUT /notifications/read-all

### 7. Favoris

#### FavoritesPage
**Champs utilisés:**
- auction_id
- user_id
- created_at

**API Calls:**
- GET /favorites
- POST /favorites/{auction_id}
- DELETE /favorites/{auction_id}

**Table existante:** `user_favorites` - ✅ Complète

---

## Comparaison Schema BD vs Interface Mobile

### Tables et Champs Manquants ou Incomplets

#### 1. Table `users`
**Champs existants:**
- id, phone, password_hash, full_name, email, profile_pic_url, city, language_pref, notifications_enabled, terms_accepted_at, is_active, role, is_verified, blocked_until, last_login_at, created_at, updated_at

**Champs manquants pour l'interface mobile:**
- `country_code` - Le mobile supporte plusieurs pays (+222, +221, +212, +216) mais la BD ne stocke que le phone
- `date_of_birth` - Potentiellement nécessaire pour KYC
- `address` - Pour la livraison
- `postal_code` - Pour la livraison
- `gender` - Pour personnalisation
- `profile_completed` - Flag pour savoir si le profil est complet
- `kyc_status` - Statut de vérification KYC (pending, approved, rejected)

**Recommandation:**
```sql
ALTER TABLE users ADD COLUMN country_code VARCHAR(5);
ALTER TABLE users ADD COLUMN date_of_birth DATE;
ALTER TABLE users ADD COLUMN address TEXT;
ALTER TABLE users ADD COLUMN postal_code VARCHAR(20);
ALTER TABLE users ADD COLUMN gender VARCHAR(10);
ALTER TABLE users ADD COLUMN profile_completed BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN kyc_status VARCHAR(20) DEFAULT 'none';
```

#### 2. Table `auctions`
**Champs existants:**
- id, seller_id, category_id, location_id, title_ar, title_fr, title_en, description_ar, description_fr, description_en, start_price, current_price, min_increment, insurance_amount, reserve_price, start_time, end_time, status, lot_number, views, bidder_count, winner_id, winning_bid_id, payment_deadline, is_featured, featured_until, rejection_reason, phone_contact, item_details (JSONB), buy_now_price, version, created_at

**Champs manquants ou problèmes:**
- `sub_category_id` - Le mobile envoie `sub_category` mais la BD n'a que `category_id` avec `parent_id`
- `condition` - État de l'item (neuf, utilisé, etc.)
- `brand` - Marque (utilisé dans les détails véhicule)
- `is_verified` - Vérification de l'annonce par admin
- `boosted_until` - Pour les annonces boostées
- `video_url` - URL vidéo de l'annonce

**Problème de structure:** Le mobile envoie `category` et `sub_category` comme strings, mais la BD utilise des IDs. Il faut soit:
1. Ajouter `sub_category_id` FK vers `categories(id)`
2. Ou mapper les strings vers IDs côté backend

**Recommandation:**
```sql
ALTER TABLE auctions ADD COLUMN sub_category_id INT REFERENCES categories(id);
ALTER TABLE auctions ADD COLUMN condition VARCHAR(20) DEFAULT 'used';
ALTER TABLE auctions ADD COLUMN brand VARCHAR(100);
ALTER TABLE auctions ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE auctions ADD COLUMN boosted_until TIMESTAMP WITH TIME ZONE;
ALTER TABLE auctions ADD COLUMN video_url TEXT;
```

#### 3. Table `bids`
**Champs existants:**
- id, auction_id, user_id, amount, previous_price, is_winning, created_at

**Champs manquants:**
- `bidder_name` - Pour l'historique des enchères sans join avec users
- `bidder_phone` - Pour l'historique des enchères (affiché dans AuctionHistoryPage)
- `is_anonymous` - Option pour enchérisseur anonyme
- `bid_time_remaining` - Temps restant quand l'enchère a été placée

**Note:** Les champs `bidder_name` et `bidder_phone` peuvent être obtenus via JOIN avec `users`, mais pour optimiser les performances de l'historique, ils pourraient être dénormalisés.

**Recommandation:**
```sql
ALTER TABLE bids ADD COLUMN bidder_name VARCHAR(100);
ALTER TABLE bids ADD COLUMN bidder_phone VARCHAR(20);
ALTER TABLE bids ADD COLUMN is_anonymous BOOLEAN DEFAULT FALSE;
```

#### 4. Table `transactions`
**Champs existants:**
- id, user_id, auction_id, type, amount, gateway, status, reference, receipt_url, admin_notes, reviewed_by, reviewed_at, created_at, wallet_hold_id

**Champs manquants:**
- `receipt_image_temp` - Image uploadée avant validation (différente de receipt_url validé)
- `payment_method` - Méthode spécifique (masrvi, bankily, sedad, click, bank_transfer, mobile_money)
- `fee_amount` - Montant des frais de transaction
- `net_amount` - Montant net après frais
- `description` - Description personnalisée
- `failure_reason` - Raison de l'échec si status = failed

**Problème:** Le mobile upload une image de reçu avant validation, mais la BD n'a qu'un champ `receipt_url` pour le reçu validé.

**Recommandation:**
```sql
ALTER TABLE transactions ADD COLUMN receipt_image_temp TEXT;
ALTER TABLE transactions ADD COLUMN payment_method VARCHAR(50);
ALTER TABLE transactions ADD COLUMN fee_amount DECIMAL(15,2) DEFAULT 0.00;
ALTER TABLE transactions ADD COLUMN net_amount DECIMAL(15,2);
ALTER TABLE transactions ADD COLUMN description TEXT;
ALTER TABLE transactions ADD COLUMN failure_reason TEXT;
```

#### 5. Table `service_requests`
**Champs existants:**
- id, user_id, service_type, pickup_location, delivery_location, status, tracking_number, estimated_price, actual_price, notes, driver_id, completed_at, created_at

**Champs manquants:**
- `pickup_address` - Adresse détaillée de récupération
- `delivery_address` - Adresse détaillée de livraison
- `pickup_contact_name` - Nom contact récupération
- `pickup_contact_phone` - Téléphone contact récupération
- `delivery_contact_name` - Nom contact livraison
- `delivery_contact_phone` - Téléphone contact livraison
- `pickup_time` - Heure souhaitée de récupération
- `delivery_time` - Heure souhaitée de livraison
- `item_description` - Description des items à livrer
- `item_images` - Images des items
- `weight` - Poids approximatif
- `distance` - Distance estimée
- `duration` - Durée estimée

**Recommandation:**
```sql
ALTER TABLE service_requests ADD COLUMN pickup_address TEXT;
ALTER TABLE service_requests ADD COLUMN delivery_address TEXT;
ALTER TABLE service_requests ADD COLUMN pickup_contact_name VARCHAR(100);
ALTER TABLE service_requests ADD COLUMN pickup_contact_phone VARCHAR(20);
ALTER TABLE service_requests ADD COLUMN delivery_contact_name VARCHAR(100);
ALTER TABLE service_requests ADD COLUMN delivery_contact_phone VARCHAR(20);
ALTER TABLE service_requests ADD COLUMN pickup_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE service_requests ADD COLUMN delivery_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE service_requests ADD COLUMN item_description TEXT;
ALTER TABLE service_requests ADD COLUMN item_images JSONB;
ALTER TABLE service_requests ADD COLUMN weight DECIMAL(10,2);
ALTER TABLE service_requests ADD COLUMN distance DECIMAL(10,2);
ALTER TABLE service_requests ADD COLUMN duration INT;
```

#### 6. Table `delivery_timeline`
**Champs existants:**
- id, request_id, step_name, description, completed_at, created_at

**Champs manquants:**
- `step_order` - Ordre du step pour tri
- `icon` - Icône pour l'affichage mobile
- `location_lat` - Latitude pour tracking GPS
- `location_lng` - Longitude pour tracking GPS
- `performed_by` - Qui a effectué l'action (driver, system, user)

**Recommandation:**
```sql
ALTER TABLE delivery_timeline ADD COLUMN step_order INT;
ALTER TABLE delivery_timeline ADD COLUMN icon VARCHAR(50);
ALTER TABLE delivery_timeline ADD COLUMN location_lat DECIMAL(10,8);
ALTER TABLE delivery_timeline ADD COLUMN location_lng DECIMAL(11,8);
ALTER TABLE delivery_timeline ADD COLUMN performed_by VARCHAR(20);
```

#### 7. Table `notifications`
**Champs existants:**
- id, user_id, type, title, body, is_read, reference_id, reference_type, data (JSONB), created_at

**Champs manquants:**
- `priority` - Priorité (low, normal, high, urgent)
- `action_url` - URL de deep link pour l'action
- `action_label` - Label du bouton d'action
- `expires_at` - Date d'expiration
- `image_url` - Image attachée à la notification

**Recommandation:**
```sql
ALTER TABLE notifications ADD COLUMN priority VARCHAR(10) DEFAULT 'normal';
ALTER TABLE notifications ADD COLUMN action_url TEXT;
ALTER TABLE notifications ADD COLUMN action_label VARCHAR(50);
ALTER TABLE notifications ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE notifications ADD COLUMN image_url TEXT;
```

#### 8. Table `categories`
**Champs existants:**
- id, name_ar, name_fr, parent_id, icon_name, display_order

**Champs manquants:**
- `name_en` - Nom en anglais (partiellement implémenté dans migration 006)
- `is_active` - Catégorie active/inactive
- `image_url` - Image de la catégorie
- `has_subcategories` - Flag pour optimiser les requêtes

**Note:** La migration 006 ajoute `name_en` mais pas les autres champs.

**Recommandation:**
```sql
ALTER TABLE categories ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE categories ADD COLUMN image_url TEXT;
ALTER TABLE categories ADD COLUMN has_subcategories BOOLEAN DEFAULT FALSE;
```

#### 9. Tables Manquantes

##### Table `auction_car_details` (pour les détails véhicule)
Le mobile affiche des détails spécifiques aux véhicules (manufacturer, fuel_type, transmission, year, mileage, model) qui sont stockés dans `item_details` JSONB. Une table dédiée serait plus propre.

```sql
CREATE TABLE auction_car_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    year INT,
    mileage INT,
    fuel_type VARCHAR(50),
    transmission VARCHAR(50),
    color VARCHAR(50),
    engine_size VARCHAR(20),
    vin VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

##### Table `payment_methods`
Pour gérer les méthodes de paiement disponibles (masrvi, bankily, sedad, click, etc.)

```sql
CREATE TABLE payment_methods (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name_ar VARCHAR(100) NOT NULL,
    name_fr VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    logo_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    country_id INT REFERENCES countries(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

##### Table `delivery_drivers`
Pour stocker les informations détaillées des chauffeurs de livraison

```sql
CREATE TABLE delivery_drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    vehicle_type VARCHAR(50),
    vehicle_plate VARCHAR(20),
    vehicle_color VARCHAR(50),
    license_number VARCHAR(50),
    rating DECIMAL(3,2),
    total_deliveries INT DEFAULT 0,
    is_available BOOLEAN DEFAULT TRUE,
    current_location_lat DECIMAL(10,8),
    current_location_lng DECIMAL(11,8),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

##### Table `auction_boosts`
Pour gérer les annonces boostées/feature

```sql
CREATE TABLE auction_boosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    boost_type VARCHAR(20) NOT NULL, -- 'featured', 'urgent', 'top'
    start_at TIMESTAMP WITH TIME ZONE NOT NULL,
    end_at TIMESTAMP WITH TIME ZONE NOT NULL,
    amount DECIMAL(15,2),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

##### Table `user_settings`
Pour les préférences utilisateur détaillées

```sql
CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(3) DEFAULT 'MRU',
    theme VARCHAR(10) DEFAULT 'auto', -- 'light', 'dark', 'auto'
    language VARCHAR(5) DEFAULT 'ar',
    notifications_email BOOLEAN DEFAULT TRUE,
    notifications_push BOOLEAN DEFAULT TRUE,
    notifications_sms BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

##### Table `bid_auto_bids`
Pour les enchères automatiques (proxy bidding)

```sql
CREATE TABLE bid_auto_bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    max_amount DECIMAL(15,2) NOT NULL,
    current_bid_amount DECIMAL(15,2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## Champs/Tables Manquants

### Résumé par Priorité

#### 🔴 Priorité Critique (Bloquant pour l'interface mobile)

1. **Table `users` - Champs manquants:**
   - `country_code` - Essentiel pour le support multi-pays
   - `address`, `postal_code` - Nécessaire pour la livraison

2. **Table `auctions` - Structure incorrecte:**
   - `sub_category_id` - Le mobile envoie sub_category mais la BD ne le supporte pas
   - Mapping strings vs IDs pour categories

3. **Table `transactions` - Champs manquants:**
   - `receipt_image_temp` - Pour l'upload de reçu avant validation
   - `payment_method` - Pour tracker la méthode utilisée

4. **Table `service_requests` - Champs manquants:**
   - Adresses détaillées pickup/delivery
   - Contacts pickup/delivery
   - Description et images des items

#### 🟡 Priorité Haute (Fonctionnalités importantes)

5. **Table `bids` - Optimisation:**
   - `bidder_name`, `bidder_phone` - Pour éviter les JOINs dans l'historique

6. **Table `delivery_timeline` - Champs manquants:**
   - `step_order`, `icon` - Pour l'affichage mobile
   - Coordonnées GPS pour tracking

7. **Table `notifications` - Améliorations:**
   - `priority`, `action_url` - Pour une meilleure UX

8. **Table `categories` - Champs manquants:**
   - `is_active`, `image_url` - Pour gestion complète

#### 🟢 Priorité Moyenne (Améliorations futures)

9. **Nouvelles tables:**
   - `auction_car_details` - Pour les détails véhicule structurés
   - `payment_methods` - Pour gérer les méthodes de paiement
   - `delivery_drivers` - Pour les infos chauffeurs
   - `auction_boosts` - Pour les annonces boostées
   - `user_settings` - Pour les préférences utilisateur
   - `bid_auto_bids` - Pour les enchères automatiques

---

## Recommandations

### 1. Actions Immédiates (Sprint 1)

#### Migration SQL Prioritaire
```sql
-- Migration: 000023_mobile_critical_fields.up.sql

-- Ajouter country_code à users
ALTER TABLE users ADD COLUMN country_code VARCHAR(5) DEFAULT '+222';
CREATE INDEX idx_users_country_code ON users(country_code);

-- Ajouter address et postal_code à users
ALTER TABLE users ADD COLUMN address TEXT;
ALTER TABLE users ADD COLUMN postal_code VARCHAR(20);

-- Ajouter sub_category_id à auctions
ALTER TABLE auctions ADD COLUMN sub_category_id INT REFERENCES categories(id);
CREATE INDEX idx_auctions_sub_category ON auctions(sub_category_id);

-- Ajouter receipt_image_temp à transactions
ALTER TABLE transactions ADD COLUMN receipt_image_temp TEXT;

-- Ajouter payment_method à transactions
ALTER TABLE transactions ADD COLUMN payment_method VARCHAR(50);
CREATE INDEX idx_transactions_payment_method ON transactions(payment_method);

-- Compléter service_requests
ALTER TABLE service_requests ADD COLUMN pickup_address TEXT;
ALTER TABLE service_requests ADD COLUMN delivery_address TEXT;
ALTER TABLE service_requests ADD COLUMN pickup_contact_name VARCHAR(100);
ALTER TABLE service_requests ADD COLUMN pickup_contact_phone VARCHAR(20);
ALTER TABLE service_requests ADD COLUMN delivery_contact_name VARCHAR(100);
ALTER TABLE service_requests ADD COLUMN delivery_contact_phone VARCHAR(20);
ALTER TABLE service_requests ADD COLUMN item_description TEXT;
ALTER TABLE service_requests ADD COLUMN item_images JSONB;

-- Optimiser bids pour l'historique
ALTER TABLE bids ADD COLUMN bidder_name VARCHAR(100);
ALTER TABLE bids ADD COLUMN bidder_phone VARCHAR(20);

-- Améliorer delivery_timeline
ALTER TABLE delivery_timeline ADD COLUMN step_order INT;
ALTER TABLE delivery_timeline ADD COLUMN icon VARCHAR(50);

-- Améliorer notifications
ALTER TABLE notifications ADD COLUMN priority VARCHAR(10) DEFAULT 'normal';
ALTER TABLE notifications ADD COLUMN action_url TEXT;

-- Compléter categories
ALTER TABLE categories ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE categories ADD COLUMN image_url TEXT;
```

### 2. Actions Backend (Sprint 1-2)

1. **Modifier l'endpoint POST /auctions:**
   - Accepter `category` et `sub_category` comme strings
   - Mapper vers `category_id` et `sub_category_id` côté backend
   - Ou créer un endpoint pour obtenir les mappings

2. **Créer l'endpoint GET /auctions/{id}/winner:**
   - Retourner les détails du gagnant et le montant gagnant

3. **Créer l'endpoint PUT /users/profile:**
   - Permettre la mise à jour du profil utilisateur
   - Gérer l'upload de photo de profil

4. **Créer l'endpoint PUT /users/settings:**
   - Permettre la mise à jour des préférences utilisateur

5. **Modifier l'endpoint POST /users/wallet/deposit:**
   - Ajouter le champ `payment_method`
   - Gérer l'image de reçu temporaire

### 3. Actions Backend (Sprint 2-3)

1. **Créer les tables supplémentaires:**
   - `auction_car_details`
   - `payment_methods`
   - `delivery_drivers`
   - `auction_boosts`
   - `user_settings`

2. **Implémenter les endpoints CRUD pour ces nouvelles tables**

3. **Optimiser les requêtes de l'historique des enchères:**
   - Utiliser les champs dénormalisés `bidder_name` et `bidder_phone`
   - Ajouter des indexes pour les performances

### 4. Actions Mobile (Optionnel)

1. **Ajouter la gestion des erreurs:**
   - Gérer les cas où `sub_category` n'existe pas
   - Afficher des messages d'erreur clairs

2. **Ajouter la validation côté mobile:**
   - Valider les formats de téléphone selon le pays
   - Valider les montants minimaux/maximaux

3. **Optimiser les appels API:**
   - Utiliser le cache pour les données statiques (categories, locations)
   - Implémenter le lazy loading pour les listes

---

## Conclusion

L'analyse a révélé plusieurs disparités entre l'interface mobile et la base de données existante. Les problèmes critiques concernent principalement:

1. **Structure des catégories:** Le mobile utilise des strings pour les catégories/sous-catégories alors que la BD utilise des IDs avec des clés étrangères
2. **Informations de localisation:** Manque de country_code et d'adresses détaillées pour le support multi-pays et la livraison
3. **Gestion des reçus:** Manque de champ pour l'image de reçu temporaire avant validation
4. **Détails de livraison:** La table service_requests manque de champs essentiels pour une expérience de livraison complète

Les recommandations proposées visent à combler ces lacunes de manière progressive, en priorisant les éléments bloquants pour l'interface mobile actuelle, tout en préparant la base de données pour les fonctionnalités futures.

---

**Document généré automatiquement par Cascade**  
**Version:** 1.0  
**Dernière mise à jour:** 2025-04-23
