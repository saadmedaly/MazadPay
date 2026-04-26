# Guide d'Intégration API - MazadPay

## Architecture

L'application Flutter utilise une architecture **MVVM** avec **Riverpod** pour la gestion d'état et **Dio** pour les requêtes HTTP.

### Structure du projet

```
lib/
├── models/          # Modèles de données (mapping JSON ↔ Dart)
├── providers/       # Gestion d'état Riverpod
├── services/        # Services API (Dio)
├── pages/           # UI Screens
└── widgets/         # Composants réutilisables
```

## Modèles de données

Les modèles suivants sont disponibles dans `lib/models/`:

| Modèle | Fichier | Backend Go |
|--------|---------|------------|
| User | `user.dart` | `models/user.go` |
| Wallet | `wallet.dart` | `models/wallet.go` |
| Transaction | `wallet.dart` | `models/wallet.go` |
| Auction | `auction_model.dart` | `models/auction.go` |
| Bid | `bid.dart` | `models/bid.go` |
| Category | `category.dart` | `models/auction.go` |
| Location | `category.dart` | `models/auction.go` |
| Country | `category.dart` | `models/auction.go` |
| Notification | `notification.dart` | `models/notification.go` |

### Exemple d'utilisation

```dart
import 'package:mezadpay/models/models.dart';

// Parsing depuis JSON
final user = User.fromJson(jsonData);

// Conversion vers JSON
final json = user.toJson();

// Utilisation
print(user.fullName);
print(user.maskedPhone);
```

## Services API

Tous les services utilisent `ApiService` comme base avec Dio.

### Services disponibles

```dart
import 'package:mezadpay/services/services.dart';

// Authentification
final authApi = AuthApi();

// Utilisateur
final userApi = UserApi();

// Wallet
final walletApi = WalletApi();

// Enchères
final auctionApi = AuctionApi();
final bidApi = BidApi();

// Favoris
final favoritesApi = FavoritesApi();

// Notifications
final notificationApi = NotificationApi();

// Catégories & Locations
final categoryApi = CategoryApi();

// Demandes
final requestApi = RequestApi();
```

### Exemple d'appel API

```dart
final response = await authApi.login(
  phone: '+213XXXXXXXX',
  pin: '1234',
);

if (response.success) {
  final user = User.fromJson(response.data!['user']);
  final token = response.data!['token'];
} else {
  final error = response.error?.message;
}
```

## Providers Riverpod

Les providers gèrent l'état global de l'application.

### Providers disponibles

```dart
import 'package:mezadpay/providers/providers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Auth
final authState = ref.watch(authNotifierProvider);
final authController = ref.read(authNotifierProvider.notifier);

// User
final userState = ref.watch(userNotifierProvider);
final userController = ref.read(userNotifierProvider.notifier);

// Wallet
final walletState = ref.watch(walletNotifierProvider);
final walletController = ref.read(walletNotifierProvider.notifier);

// Notifications
final notificationState = ref.watch(notificationNotifierProvider);
final notificationController = ref.read(notificationNotifierProvider.notifier);

// Favoris
final favorites = ref.watch(favoritesProvider);

// Locale (langue)
final locale = ref.watch(localeNotifierProvider);
```

### Exemple d'utilisation dans un widget

```dart
class MyPage extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authNotifierProvider);
    final userState = ref.watch(userNotifierProvider);

    if (authState.isLoading) {
      return const CircularProgressIndicator();
    }

    return Scaffold(
      body: Text('Bienvenue ${userState.user?.fullName}'),
    );
  }
}
```

## Intégration des Pages Existantes

### Login Page

```dart
class LoginPage extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authNotifierProvider);

    ref.listen(authNotifierProvider, (previous, next) {
      if (next.isAuthenticated) {
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (_) => const HomePage()),
        );
      }
    });

    return Scaffold(
      body: Column(
        children: [
          if (authState.error != null)
            ErrorMessage(message: authState.error!),
          LoginForm(
            onSubmit: (phone, pin) async {
              final success = await ref
                  .read(authNotifierProvider.notifier)
                  .login(phone: phone, pin: pin);
              
              if (success) {
                // Charger le profil utilisateur
                await ref.read(userNotifierProvider.notifier).loadProfile();
              }
            },
          ),
          if (authState.isLoading)
            const CircularProgressIndicator(),
        ],
      ),
    );
  }
}
```

### Home Page avec enchères

```dart
class HomePage extends ConsumerStatefulWidget {
  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  @override
  void initState() {
    super.initState();
    // Charger les données au démarrage
    Future.microtask(() {
      ref.read(userNotifierProvider.notifier).loadProfile();
      ref.read(notificationNotifierProvider.notifier).loadNotifications();
    });
  }

  @override
  Widget build(BuildContext context) {
    final userState = ref.watch(userNotifierProvider);
    final notificationState = ref.watch(notificationNotifierProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('MazadPay'),
        actions: [
          // Badge de notifications
          NotificationBadge(
            count: notificationState.unreadCount,
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const NotificationsPage()),
            ),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(userNotifierProvider.notifier).refresh();
          await ref.read(notificationNotifierProvider.notifier).refresh();
        },
        child: ListView(
          children: [
            UserProfileHeader(user: userState.user),
            AuctionList(),
          ],
        ),
      ),
    );
  }
}
```

### Account Page avec Wallet

```dart
class AccountPage extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userState = ref.watch(userNotifierProvider);
    final walletState = ref.watch(walletNotifierProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Mon Compte')),
      body: ListView(
        children: [
          // Solde du wallet
          WalletCard(
            balance: walletState.availableBalance,
            frozenAmount: walletState.frozenAmount,
            onDeposit: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const DepositPage()),
            ),
            onWithdraw: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const WithdrawPage()),
            ),
          ),
          
          // Infos utilisateur
          UserInfoCard(user: userState.user),
          
          // Menu
          ListTile(
            leading: const Icon(Icons.favorite),
            title: const Text('Mes Favoris'),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const FavoritesPage()),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.gavel),
            title: const Text('Mes Enchères'),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const MyAuctionsPage()),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.history),
            title: const Text('Historique des transactions'),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const TransactionHistoryPage()),
            ),
          ),
          ListTile(
            leading: const Icon(Icons.logout),
            title: const Text('Déconnexion'),
            onTap: () async {
              await ref.read(authNotifierProvider.notifier).logout();
              Navigator.pushAndRemoveUntil(
                context,
                MaterialPageRoute(builder: (_) => const LoginPage()),
                (route) => false,
              );
            },
          ),
        ],
      ),
    );
  }
}
```

### Auction Details Page

```dart
class AuctionDetailsPage extends ConsumerStatefulWidget {
  final String auctionId;

  const AuctionDetailsPage({super.key, required this.auctionId});

  @override
  ConsumerState<AuctionDetailsPage> createState() => _AuctionDetailsPageState();
}

class _AuctionDetailsPageState extends ConsumerState<AuctionDetailsPage> {
  AuctionModel? _auction;
  bool _isLoading = true;
  final AuctionApi _auctionApi = AuctionApi();
  final BidApi _bidApi = BidApi();

  @override
  void initState() {
    super.initState();
    _loadAuction();
  }

  Future<void> _loadAuction() async {
    final response = await _auctionApi.getAuctionById(widget.auctionId);
    if (response.success && response.data != null) {
      setState(() {
        _auction = AuctionModel.fromJson(response.data!);
        _isLoading = false;
      });
    }
  }

  Future<void> _placeBid(double amount) async {
    final response = await _bidApi.placeBid(
      auctionId: widget.auctionId,
      amount: amount,
    );

    if (response.success) {
      await _loadAuction(); // Recharger les données
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Enchère placée avec succès')),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(response.error?.message ?? 'Erreur')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      appBar: AppBar(title: Text(_auction!.getTitle('ar'))),
      body: Column(
        children: [
          AuctionImageGallery(images: _auction!.images),
          AuctionInfo(auction: _auction!),
          BidHistory(auctionId: widget.auctionId),
          BidButton(
            minBid: _auction!.minBidAmount,
            onBid: _placeBid,
          ),
        ],
      ),
    );
  }
}
```

## Gestion des Erreurs

Toutes les erreurs API sont gérées via les classes d'exception dans `ApiService`:

```dart
try {
  final response = await api.someCall();
} on UnauthorizedException {
  // Rediriger vers login
  Navigator.pushReplacement(...);
} on RateLimitException {
  // Afficher message "Trop de requêtes"
  showSnackBar('Veuillez réessayer plus tard');
} on ServerException {
  // Erreur serveur
  showSnackBar('Erreur serveur, veuillez réessayer');
} on ApiException catch (e) {
  // Autre erreur
  showSnackBar(e.message);
}
```

## Génération du code Riverpod

Après modification des providers, générer le code:

```bash
cd MazadPay
flutter pub run build_runner build --delete-conflicting-outputs
```

Ou pour watch mode:

```bash
flutter pub run build_runner watch --delete-conflicting-outputs
```

## Configuration

Créer un fichier `.env` à la racine du projet Flutter:

```
API_BASE_URL=https://api.mezadpay.com/v1/api
```

Pour développement local:

```
API_BASE_URL=http://localhost:8082/v1/api
```

## Sécurité

- Les tokens JWT sont stockés dans `FlutterSecureStorage`
- Les requêtes utilisent HTTPS en production
- L'interceptor `AuthInterceptor` ajoute automatiquement le token aux requêtes
- L'interceptor `ErrorInterceptor` gère les erreurs 401 (token expiré)

## Internationalisation (i18n)

L'application supporte 3 langues:
- 🇸🇦 Arabe (`ar`) - par défaut
- 🇫🇷 Français (`fr`)
- 🇬🇧 Anglais (`en`)

Changer la langue:

```dart
await ref.read(userNotifierProvider.notifier).updateLanguage('fr');
```

## Pagination

Les listes paginées utilisent le pattern suivant:

```dart
// Premier chargement
await ref.read(walletNotifierProvider.notifier).loadTransactions(refresh: true);

// Charger plus (infinite scroll)
await ref.read(walletNotifierProvider.notifier).loadMoreTransactions();

// Vérifier s'il y a plus de données
final hasMore = ref.watch(walletNotifierProvider).hasMoreTransactions;
```
