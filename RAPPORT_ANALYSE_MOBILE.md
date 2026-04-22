# 📱 Rapport d'Analyse Interface Mobile - MazadPay

**Date :** 22 Avril 2026  
**Framework :** Flutter  
**Base URL API :** `http://localhost:8082/v1/api`

---

## 📊 Résumé

| Catégorie | Nombre de Pages |
|-----------|----------------|
| Authentification | 5 |
| Enchères (Auctions) | 5 |
| Compte Utilisateur | 5 |
| Wallet/Paiements | 4 |
| Autres | 17 |
| **TOTAL** | **36** |

---

## 🔐 Module Authentification

### 1. **login_page.dart**
**Fonctionnalité :** Page de connexion avec téléphone et PIN

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Connexion | POST | `/auth/login` | Authentifier l'utilisateur avec phone + pin |

**Paramètres :**
```json
{
  "phone": "string",
  "pin": "string"
}
```

**Note :** L'application utilise actuellement des données mockées. À connecter au backend.

---

### 2. **phone_registration_page.dart**
**Fonctionnalité :** Inscription avec numéro de téléphone

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Inscription | POST | `/auth/register` | Créer un compte utilisateur |
| Envoyer OTP | POST | `/auth/otp/send` | Envoyer code OTP par SMS |
| Vérifier OTP | POST | `/auth/otp/verify` | Vérifier le code OTP |

**Paramètres (Register) :**
```json
{
  "phone": "string",
  "pin": "string",
  "full_name": "string",
  "email": "string",
  "city": "string"
}
```

**Paramètres (OTP Send) :**
```json
{
  "phone": "string",
  "purpose": "register"
}
```

**Paramètres (OTP Verify) :**
```json
{
  "phone": "string",
  "code": "string",
  "purpose": "register"
}
```

---

### 3. **otp_entry_page.dart**
**Fonctionnalité :** Saisie du code OTP

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Vérifier OTP | POST | `/auth/otp/verify` | Valider le code OTP |

**Paramètres :**
```json
{
  "phone": "string",
  "code": "string",
  "purpose": "register"
}
```

---

### 4. **set_password_page.dart**
**Fonctionnalité :** Définition du mot de passe (PIN)

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Changer PIN | PUT | `/auth/change-password` | Mettre à jour le PIN |

**Paramètres :**
```json
{
  "old_pin": "string",
  "new_pin": "string"
}
```

---

### 5. **new_password_page.dart**
**Fonctionnalité :** Réinitialisation du mot de passe

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Reset Password | POST | `/auth/reset-password` | Réinitialiser le PIN |

**Paramètres :**
```json
{
  "phone": "string",
  "new_pin": "string",
  "otp_code": "string"
}
```

---

## 🏆 Module Enchères (Auctions)

### 6. **home_page.dart**
**Fonctionnalité :** Page d'accueil avec liste d'enchères

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Lister enchères | GET | `/auctions` | Récupérer la liste des enchères |
| Lister catégories | GET | `/categories` | Récupérer les catégories |
| Lister locations | GET | `/locations` | Récupérer les villes |
| Lister pays | GET | `/countries` | Récupérer les pays |

**Paramètres (Auctions) :**
```
?page=1&limit=20&category_id=&location_id=&status=
```

---

### 7. **all_auctions_page.dart**
**Fonctionnalité :** Page toutes les enchères avec filtres

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Lister enchères | GET | `/auctions` | Récupérer enchères avec filtres |
| Lister catégories | GET | `/categories` | Récupérer catégories pour filtres |

**Paramètres :**
```
?page=1&limit=20&category_id={uuid}&location_id={uuid}&status=active
```

---

### 8. **auction_details_page.dart**
**Fonctionnalité :** Détails d'une enchère

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Détails enchère | GET | `/auctions/:id` | Récupérer détails enchère |
| Historique enchères | GET | `/auctions/:id/bids` | Récupérer historique des enchères |
| Incrémenter vues | POST | `/auctions/:id/view` | Incrémenter compteur vues |
| Placer enchère | POST | `/auctions/:id/bids` | Placer une enchère |
| Contact vendeur | GET | `/auctions/:id/seller-contact` | Obtenir contact vendeur |

**Paramètres (Place Bid) :**
```json
{
  "amount": "number"
}
```

**Note :** L'app utilise un provider avec des données mockées. À connecter à l'API.

---

### 9. **auction_history_page.dart**
**Fonctionnalité :** Historique des enchères d'une enchère

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Historique enchères | GET | `/auctions/:id/bids` | Récupérer historique |

---

### 10. **auction_winner_page.dart**
**Fonctionnalité :** Page gagnant d'enchère

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mes gains | GET | `/users/me/winnings` | Récupérer enchères gagnées |
| Détails enchère | GET | `/auctions/:id` | Détails de l'enchère gagnée |

---

## 👤 Module Compte Utilisateur

### 11. **account_page.dart**
**Fonctionnalité :** Page compte utilisateur principal

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Profil utilisateur | GET | `/users/me` | Récupérer profil |
| Solde wallet | GET | `/users/wallet` | Récupérer solde |

---

### 12. **account_profile_page.dart**
**Fonctionnalité :** Page profil utilisateur

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mettre à jour profil | PUT | `/users/me` | Mettre à jour infos |
| Mettre à jour avatar | POST | `/users/me/avatar` | Mettre à jour photo |
| Changer langue | PUT | `/users/me/language` | Changer langue |
| Préférences notifications | PUT | `/users/me/notification-prefs` | Mettre à jour préférences |

**Paramètres (Update Profile) :**
```json
{
  "full_name": "string",
  "email": "string",
  "phone": "string"
}
```

---

### 13. **create_profile_page.dart**
**Fonctionnalité :** Création de profil

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mettre à jour profil | PUT | `/users/me` | Compléter profil |

---

### 14. **my_auctions_page.dart**
**Fonctionnalité :** Mes enchères

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mes enchères | GET | `/users/me/auctions` | Récupérer enchères créées |

**Paramètres :**
```
?status=active|ended|cancelled
```

---

### 15. **my_winnings_page.dart**
**Fonctionnalité :** Mes gains

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mes gains | GET | `/users/me/winnings` | Récupérer enchères gagnées |

---

## ❤️ Module Favoris

### 16. **favorites_page.dart**
**Fonctionnalité :** Page favoris

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Lister favoris | GET | `/users/me/favorites` | Récupérer enchères favorites |
| Ajouter favori | POST | `/users/me/favorites/:auction_id` | Ajouter aux favoris |
| Supprimer favori | DELETE | `/users/me/favorites/:auction_id` | Supprimer des favoris |

**Note :** L'app utilise un provider local. À connecter à l'API.

---

## 💰 Module Wallet & Paiements

### 17. **deposit_page.dart**
**Fonctionnalité :** Page de dépôt

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Dépôt | POST | `/users/wallet/deposit` | Initier un dépôt |

**Paramètres :**
```json
{
  "amount": "number",
  "payment_method": "string"
}
```

---

### 18. **payment_details_page.dart**
**Fonctionnalité :** Détails de paiement

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Upload reçu | POST | `/users/wallet/transactions/:id/receipt` | Upload reçu de paiement |
| Détails transaction | GET | `/users/wallet/transactions/:id` | Détails transaction |

---

### 19. **payment_success_page.dart**
**Fonctionnalité :** Page succès paiement

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Historique transactions | GET | `/users/wallet/transactions` | Voir transactions |

---

### 20. **withdraw_page.dart**
**Fonctionnalité :** Page de retrait

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Retrait | POST | `/users/wallet/withdraw` | Initier un retrait |

**Paramètres :**
```json
{
  "amount": "number",
  "bank_details": "object"
}
```

---

## 📝 Module Création d'Annonces

### 21. **create_ad_start_page.dart**
**Fonctionnalité :** Page de démarrage création annonce

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Lister catégories | GET | `/categories` | Récupérer catégories |
| Lister locations | GET | `/locations` | Récupérer villes |

---

### 22. **create_ad_form_page.dart**
**Fonctionnalité :** Formulaire de création d'annonce

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Créer demande enchère | POST | `/requests/auctions` | Soumettre demande d'enchère |

**Paramètres :**
```json
{
  "title": "string",
  "description": "string",
  "starting_price": "number",
  "category_id": "uuid",
  "location": "string"
}
```

**Note :** Rate limité à 5 requêtes/heure par utilisateur.

---

### 23. **ad_success_page.dart**
**Fonctionnalité :** Page succès après création annonce

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mes demandes | GET | `/requests/auctions/my` | Voir mes demandes |

---

## 🔔 Module Notifications

### 24. **notifications_page.dart**
**Fonctionnalité :** Page notifications

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Lister notifications | GET | `/notifications` | Récupérer notifications |
| Marquer comme lue | PUT | `/notifications/:id/read` | Marquer notification lue |
| Tout marquer comme lu | PUT | `/notifications/read-all` | Tout marquer lu |
| Sauvegarder token FCM | POST | `/notifications/token` | Sauvegarder token push |

**Paramètres (List Notifications) :**
```
?limit=20
```

**Paramètres (Save Token) :**
```json
{
  "fcm_token": "string",
  "device_id": "string",
  "platform": "string"
}
```

---

## 🌐 Module Autres Pages

### 25. **splash_page.dart**
**Fonctionnalité :** Écran de démarrage

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Vérifier token | GET | `/users/me` | Vérifier si token valide (optionnel) |

---

### 26. **onboarding_page.dart**
**Fonctionnalité :** Page d'onboarding

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Aucun | - | - | Page statique |

---

### 27. **language_page.dart**
**Fonctionnalité :** Sélection de langue

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Changer langue | PUT | `/users/me/language` | Mettre à jour langue |

**Paramètres :**
```json
{
  "language": "ar|fr|en"
}
```

---

### 28. **about_mazad_pay_page.dart**
**Fonctionnalité :** À propos de MazadPay

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| À propos | GET | `/about` | Contenu À propos |

---

### 29. **privacy_policy_page.dart**
**Fonctionnalité :** Politique de confidentialité

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Politique | GET | `/privacy-policy` | Contenu politique |

---

### 30. **terms_page.dart**
**Fonctionnalité :** Conditions d'utilisation

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| FAQ | GET | `/faq` | Questions fréquentes |
| Tutoriels | GET | `/tutorials` | Tutoriels |

---

### 31. **services_page.dart**
**Fonctionnalité :** Page services

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Aucun | - | - | Page statique |

---

### 32. **support_page.dart**
**Fonctionnalité :** Page support

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Aucun | - | - | Page statique (ou créer endpoint support) |

---

### 33. **start_bidding_page.dart**
**Fonctionnalité :** Page de démarrage d'enchère

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Tutoriels | GET | `/tutorials` | Comment enchérir |

---

### 34. **how_to_bid_page.dart**
**Fonctionnalité :** Comment enchérir

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Tutoriels | GET | `/tutorials` | Guide enchères |

---

### 35. **delivery_details_page.dart**
**Fonctionnalité :** Détails de livraison

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Détails transaction | GET | `/users/wallet/transactions/:id` | Détails livraison |

---

### 36. **auction_winner_page.dart**
**Fonctionnalité :** Page gagnant (déjà listée)

**Endpoints Backend :**
| Action | Méthode | Endpoint | Description |
|--------|---------|----------|-------------|
| Mes gains | GET | `/users/me/winnings` | Enchères gagnées |

---

## 🔌 WebSocket Endpoints

### Enchères en temps réel
| Endpoint | Description |
|----------|-------------|
| `WS /ws/auction/:id` | WebSocket pour enchères en temps réel |

---

## 📝 Notes Importantes

### 1. **État actuel de l'application**
- L'application utilise des **données mockées** (dummy data) dans les providers
- Les services API ne sont pas encore implémentés
- Toutes les données sont statiques/hardcodées

### 2. **À implémenter**
1. **Service HTTP** : Créer une classe pour les requêtes API (Dio ou http package)
2. **Gestion des tokens** : Stocker et utiliser le JWT token
3. **Gestion d'erreurs** : Gérer les erreurs API (401, 404, 500, etc.)
4. **Loading states** : Ajouter des indicateurs de chargement
5. **Refresh tokens** : Implémenter le rafraîchissement des tokens

### 3. **Rate Limiting**
- **Auth endpoints** : Limité par numéro de téléphone
- **Request endpoints** : 5 requêtes/heure par utilisateur
- Gérer les erreurs 429 avec retry ou attente

### 4. **Pagination**
- La plupart des endpoints utilisent : `?page=1&limit=20`
- Implémenter l'infinite scroll ou pagination classique

### 5. **UUID**
- Tous les IDs dans les paths sont des UUIDs
- S'assurer de passer des UUIDs valides dans les requêtes

### 6. **Langues**
- Langues supportées : `ar` (Arabe), `fr` (Français), `en` (Anglais)
- Envoyer `Accept-Language` header ou query param

### 7. **Fichiers**
- Upload d'images : Utiliser multipart/form-data
- Endpoint pour upload : À vérifier dans le backend

---

## 🔧 Recommandations d'Implémentation

### 1. Créer un service API centralisé

```dart
// lib/services/api_service.dart
class ApiService {
  final String baseUrl = 'http://localhost:8082/v1/api';
  final Dio _dio = Dio();
  
  Future<LoginResponse> login(String phone, String pin) async {
    // Implementation
  }
  
  Future<List<Auction>> getAuctions({int page = 1, int limit = 20}) async {
    // Implementation
  }
  
  // ... other methods
}
```

### 2. Gestion du JWT Token

```dart
// lib/services/auth_service.dart
class AuthService {
  Future<void> saveToken(String token) async {
    await secureStorage.write(key: 'jwt_token', value: token);
  }
  
  Future<String?> getToken() async {
    return await secureStorage.read(key: 'jwt_token');
  }
  
  Future<void> logout() async {
    await secureStorage.delete(key: 'jwt_token');
  }
}
```

### 3. Interceptor pour ajouter le token

```dart
_dio.interceptors.add(InterceptorsWrapper(
  onRequest: (options, handler) async {
    final token = await AuthService().getToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    return handler.next(options);
  },
));
```

---

**Fin du rapport**
