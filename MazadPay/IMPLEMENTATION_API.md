# 📱 Implémentation API Centralisée - MazadPay Mobile

**Date :** 22 Avril 2026  
**Framework :** Flutter  
**Backend API :** `http://localhost:8082/v1/api`

---

## 📊 Résumé de l'implémentation

J'ai créé une structure de services API centralisée suivant les best practices pour remplacer les données mockées par de vrais appels API tout en conservant la même UI.

### ✅ Fichiers créés

| Fichier | Description |
|---------|-------------|
| `lib/services/api_service.dart` | Service API centralisé avec Dio |
| `lib/services/auth_service.dart` | Gestion du JWT token avec flutter_secure_storage |
| `lib/services/interceptors/auth_interceptor.dart` | Interceptor pour authentification JWT |
| `lib/services/interceptors/error_interceptor.dart` | Interceptor pour gestion des erreurs |
| `lib/models/api_response.dart` | Modèles de réponse API standardisés |
| `lib/services/auth_api.dart` | Service API pour l'authentification |
| `lib/services/auction_api.dart` | Service API pour les enchères |
| `lib/services/wallet_api.dart` | Service API pour le wallet |
| `lib/services/notification_api.dart` | Service API pour les notifications |
| `lib/services/user_api.dart` | Service API pour les utilisateurs |
| `lib/services/request_api.dart` | Service API pour les demandes |
| `lib/providers/auction_provider_api.dart` | Provider enchères avec API |
| `lib/providers/favorites_provider_api.dart` | Provider favoris avec API |

---

## 🔧 Installation des dépendances

Ajouté à `pubspec.yaml` :
```yaml
# API & HTTP
dio: ^5.4.0

# Secure Storage
flutter_secure_storage: ^9.0.0
```

**Installer les dépendances :**
```bash
flutter pub get
```

---

## 🏗️ Architecture

### 1. **ApiService** (Service centralisé)
- Gère toutes les requêtes HTTP avec Dio
- Inclut timeout, headers par défaut
- Intercepteurs automatiques pour auth et erreurs
- Méthodes génériques : `get()`, `post()`, `put()`, `delete()`, `upload()`

### 2. **AuthService** (Gestion JWT)
- Stocke le token JWT de manière sécurisée
- Gère le refresh token
- Gère l'ID utilisateur
- Méthodes : `saveToken()`, `getToken()`, `logout()`, `isLoggedIn()`

### 3. **Interceptors**
- **AuthInterceptor** : Ajoute automatiquement le token JWT à chaque requête
- **ErrorInterceptor** : Gère centralisée des erreurs avec messages personnalisés

### 4. **Services API spécifiques**
Chaque module a son propre service :
- `AuthApi` : login, register, OTP, reset password
- `AuctionApi` : enchères, catégories, locations
- `WalletApi` : wallet, transactions, dépôts, retraits
- `NotificationApi` : notifications push, FCM
- `UserApi` : profil, favoris, enchères, KYC
- `RequestApi` : demandes d'enchères/bannières

---

## 📝 Utilisation

### Exemple 1 : Connexion utilisateur

```dart
import 'package:mezadpay/services/auth_api.dart';

final authApi = AuthApi();

final response = await authApi.login(
  phone: '32323232',
  pin: '3232',
);

if (response.success) {
  final token = response.data!['token'];
  // Le token est automatiquement sauvegardé par AuthService
  // Rediriger vers la page d'accueil
} else {
  // Afficher l'erreur
  print(response.error?.message);
}
```

### Exemple 2 : Lister les enchères

```dart
import 'package:mezadpay/services/auction_api.dart';

final auctionApi = AuctionApi();

final response = await auctionApi.getAuctions(
  page: 1,
  limit: 20,
  categoryId: 'uuid',
  status: 'active',
);

if (response.success) {
  final auctions = response.data!['auctions'];
  // Afficher la liste
}
```

### Exemple 3 : Placer une enchère

```dart
import 'package:mezadpay/services/auction_api.dart';

final auctionApi = AuctionApi();

final response = await auctionApi.placeBid(
  auctionId: 'uuid',
  amount: 15000.0,
);

if (response.success) {
  // Succès
} else {
  // Afficher l'erreur
}
```

### Exemple 4 : Gérer les favoris

```dart
import 'package:mezadpay/providers/favorites_provider_api.dart';

// Dans un Widget ConsumerWidget
final favorites = ref.watch(favoritesProviderApi);

// Ajouter aux favoris
await ref.read(favoritesProviderApi.notifier).addToFavorites('auctionId');

// Retirer des favoris
await ref.read(favoritesProviderApi.notifier).removeFromFavorites('auctionId');

// Vérifier si favori
final isFav = ref.read(favoritesProviderApi.notifier).isFavorite('auctionId');
```

---

## 🔌 Utilisation des Providers API

### Remplacer les providers existants

**Avant (avec mock data) :**
```dart
import 'package:mezadpay/providers/auction_provider.dart';

final auction = ref.watch(auctionNotifierProvider(auctionId));
```

**Après (avec API) :**
```dart
import 'package:mezadpay/providers/auction_provider_api.dart';

final auction = ref.watch(auctionNotifierApiProvider(auctionId));
```

### Remplacer le provider de favoris

**Avant :**
```dart
import 'package:mezadpay/providers/favorites_provider.dart';

final favorites = ref.watch(favoritesProvider);
```

**Après :**
```dart
import 'package:mezadpay/providers/favorites_provider_api.dart';

final favorites = ref.watch(favoritesProviderApi);
```

---

## ⚙️ Configuration

### Changer l'URL de l'API

Dans `lib/services/api_service.dart`, modifier la baseUrl :

```dart
final Dio _dio = Dio(BaseOptions(
  baseUrl: 'http://YOUR_API_URL/v1/api',  // Modifier ici
  connectTimeout: const Duration(seconds: 30),
  // ...
));
```

### Désactiver le logger en production

Commenter ou supprimer le LogInterceptor :

```dart
// _dio.interceptors.add(
//   LogInterceptor(
//     requestBody: true,
//     responseBody: true,
//   ),
// );
```

---

## 🚨 Gestion des erreurs

Les exceptions personnalisées sont définies dans `ApiService` :

```dart
try {
  final response = await authApi.login(phone: '32323232', pin: '3232');
} on UnauthorizedException {
  // Token invalide ou expiré
} on RateLimitException {
  // Trop de requêtes
} on ServerException {
  // Erreur serveur
} on ApiException catch (e) {
  // Autre erreur
  print(e.message);
}
```

---

## 📝 Étapes suivantes pour l'intégration

### 1. Installer les dépendances
```bash
flutter pub get
```

### 2. Générer les fichiers Riverpod
```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

### 3. Mettre à jour les imports dans les pages
Remplacer les imports des providers existants par les nouveaux providers API.

### 4. Tester l'intégration
- Tester la connexion
- Tester la liste des enchères
- Tester les favoris
- Vérifier que la UI reste inchangée

---

## 🔄 Migration progressive

Vous pouvez migrer progressivement :

1. **Phase 1** : Utiliser les services API pour l'authentification uniquement
2. **Phase 2** : Migrer les enchères
3. **Phase 3** : Migrer les favoris
4. **Phase 4** : Migrer le wallet
5. **Phase 5** : Migrer le reste

Les providers mockés peuvent coexister avec les providers API pendant la transition.

---

## 📋 Checklist de migration

- [ ] Installer les dépendances (`flutter pub get`)
- [ ] Générer les fichiers Riverpod (`build_runner`)
- [ ] Mettre à jour `login_page.dart` pour utiliser `AuthApi`
- [ ] Mettre à jour `auction_details_page.dart` pour utiliser `AuctionNotifierApi`
- [ ] Mettre à jour `favorites_page.dart` pour utiliser `FavoritesProviderApi`
- [ ] Mettre à jour `home_page.dart` pour utiliser `AuctionApi`
- [ ] Tester l'authentification
- [ ] Tester les enchères
- [ ] Tester les favoris
- [ ] Tester le wallet
- [ ] Désactiver le logger en production

---

## 🎯 Avantages de cette implémentation

1. **Centralisation** : Toutes les requêtes passent par ApiService
2. **Sécurité** : JWT token stocké de manière sécurisée
3. **Auth automatique** : Token ajouté automatiquement à chaque requête
4. **Erreurs gérées** : Messages d'erreur personnalisés
5. **Maintenabilité** : Code structuré et facile à maintenir
6. **Testabilité** : Services faciles à tester
7. **UI inchangée** : L'interface utilisateur reste la même
8. **Scalabilité** : Facile d'ajouter de nouveaux endpoints

---

**Fin de la documentation**
