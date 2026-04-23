# 📋 Plan d'Implémentation API - MazadPay Mobile

**Date :** 22 Avril 2026  
**Objectif :** Remplacer les données mockées par de vrais appels API tout en conservant l'UI

---

## 🎯 Vue d'ensemble

### Services API créés ✅
- `ApiService` - Service HTTP centralisé
- `AuthService` - Gestion JWT token
- `AuthInterceptor` - Ajout auto token
- `ErrorInterceptor` - Gestion erreurs
- `AuthApi` - Authentification
- `AuctionApi` - Enchères
- `WalletApi` - Wallet
- `NotificationApi` - Notifications
- `UserApi` - Utilisateurs
- `RequestApi` - Demandes

### Providers API créés ✅
- `AuctionNotifierApi` - Enchères avec API
- `FavoritesNotifierApi` - Favoris avec API

---

## 📝 Plan d'implémentation par Module

### Module 1 : Authentification ⚡ Priorité HAUTE

#### 1.1 `login_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `AuthApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auth_api.dart';

class _LoginPageState extends State<LoginPage> {
  final AuthApi _authApi = AuthApi();
  bool _isLoading = false;

  Future<void> _login() async {
    setState(() => _isLoading = true);
    
    final response = await _authApi.login(
      phone: _phoneController.text,
      pin: _passwordController.text,
    );
    
    setState(() => _isLoading = false);
    
    if (response.success) {
      // Rediriger vers HomePage
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (context) => const HomePage()),
      );
    } else {
      // Afficher erreur
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(response.error?.message ?? 'Erreur de connexion')),
      );
    }
  }
}
```

**Ajouter UI :** Indicateur de chargement pendant la requête

---

#### 1.2 `phone_registration_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `AuthApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auth_api.dart';

class _PhoneRegistrationPageState extends State<PhoneRegistrationPage> {
  final AuthApi _authApi = AuthApi();
  
  Future<void> _sendOTP() async {
    final response = await _authApi.sendOTP(
      phone: _phoneController.text,
      purpose: 'register',
    );
    
    if (response.success) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => OtpEntryPage(phone: _phoneController.text),
        ),
      );
    } else {
      // Afficher erreur
    }
  }
}
```

---

#### 1.3 `otp_entry_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `AuthApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auth_api.dart';

class _OtpEntryPageState extends State<OtpEntryPage> {
  final AuthApi _authApi = AuthApi();
  
  Future<void> _verifyOTP() async {
    final response = await _authApi.verifyOTP(
      phone: widget.phone,
      code: _otpController.text,
      purpose: 'register',
    );
    
    if (response.success) {
      // Rediriger vers set_password
    } else {
      // Afficher erreur
    }
  }
}
```

---

#### 1.4 `set_password_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `AuthApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auth_api.dart';

Future<void> _setPassword() async {
  final response = await _authApi.register(
    phone: widget.phone,
    pin: _pinController.text,
    fullName: widget.fullName,
    email: widget.email,
  );
  
  if (response.success) {
    // Rediriger vers login
  }
}
```

---

### Module 2 : Enchères (Auctions) ⚡ Priorité HAUTE

#### 2.1 `home_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `AuctionApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auction_api.dart';

class _HomePageState extends State<HomePage> {
  final AuctionApi _auctionApi = AuctionApi();
  List<Map<String, dynamic>> _auctions = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadAuctions();
  }

  Future<void> _loadAuctions() async {
    final response = await _auctionApi.getAuctions(
      page: 1,
      limit: 10,
      status: 'active',
    );
    
    setState(() {
      _isLoading = false;
      if (response.success) {
        _auctions = List.from(response.data!['auctions'] ?? []);
      }
    });
  }
}
```

---

#### 2.2 `all_auctions_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `AuctionApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auction_api.dart';

class _AllAuctionsPageState extends State<AllAuctionsPage> {
  final AuctionApi _auctionApi = AuctionApi();
  List<Map<String, dynamic>> _auctions = [];
  
  Future<void> _loadAuctions() async {
    final response = await _auctionApi.getAuctions(
      page: 1,
      limit: 20,
      categoryId: _selectedCategoryId,
      status: 'active',
    );
    
    if (response.success) {
      setState(() {
        _auctions = List.from(response.data!['auctions'] ?? []);
      });
    }
  }
}
```

---

#### 2.3 `auction_details_page.dart`
**État actuel :** Provider mocké (`auctionNotifierProvider`)  
**Service à utiliser :** `AuctionNotifierApi`

**Modifications :**
```dart
// Remplacer l'import
import 'package:mezadpay/providers/auction_provider_api.dart';

class AuctionDetailsPage extends ConsumerStatefulWidget {
  final String auctionId;
  const AuctionDetailsPage({super.key, required this.auctionId});

  @override
  ConsumerState<AuctionDetailsPage> createState() => _AuctionDetailsPageState();
}

class _AuctionDetailsPageState extends ConsumerState<AuctionDetailsPage> {
  // Remplacer auctionNotifierProvider par auctionNotifierApiProvider
  final auction = ref.watch(auctionNotifierApiProvider(widget.auctionId));
  
  Future<void> _placeBid(double amount) async {
    final success = await ref.read(auctionNotifierApiProvider(widget.auctionId).notifier).placeBid(amount);
    
    if (success) {
      // Succès
    } else {
      // Afficher erreur
    }
  }
}
```

---

#### 2.4 `auction_history_page.dart`
**État actuel :** Provider mocké  
**Service à utiliser :** `AuctionApi`

**Modifications :**
```dart
import 'package:mezadpay/services/auction_api.dart';

Future<void> _loadHistory() async {
  final response = await _auctionApi.getBidHistory(widget.auctionId);
  
  if (response.success) {
    setState(() {
      _bidHistory = List.from(response.data!['bids'] ?? []);
    });
  }
}
```

---

### Module 3 : Favoris ⚡ Priorité HAUTE

#### 3.1 `favorites_page.dart`
**État actuel :** Provider mocké (`favoritesProvider`)  
**Service à utiliser :** `FavoritesProviderApi`

**Modifications :**
```dart
// Remplacer l'import
import 'package:mezadpay/providers/favorites_provider_api.dart';

class FavoritesPage extends ConsumerWidget {
  // Remplacer favoritesProvider par favoritesProviderApi
  final favoriteIds = ref.watch(favoritesProviderApi);
  
  // Le reste du code reste identique
}
```

**Modifier `auction_details_page.dart` pour le bouton favori :**
```dart
// Remplacer favoritesProvider par favoritesProviderApi
IconButton(
  onPressed: () {
    ref.read(favoritesProviderApi.notifier).toggleFavorite(widget.auctionId);
  },
  icon: Icon(
    ref.watch(favoritesProviderApi).contains(widget.auctionId) 
      ? Icons.favorite 
      : Icons.favorite_border,
  ),
)
```

---

### Module 4 : Compte Utilisateur ⚡ Priorité MOYENNE

#### 4.1 `account_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `UserApi`, `WalletApi`

**Modifications :**
```dart
import 'package:mezadpay/services/user_api.dart';
import 'package:mezadpay/services/wallet_api.dart';

class _AccountPageState extends State<AccountPage> {
  final UserApi _userApi = UserApi();
  final WalletApi _walletApi = WalletApi();
  Map<String, dynamic>? _userProfile;
  Map<String, dynamic>? _walletInfo;
  
  @override
  void initState() {
    super.initState();
    _loadUserData();
  }
  
  Future<void> _loadUserData() async {
    final userResponse = await _userApi.getProfile();
    final walletResponse = await _walletApi.getWalletBalance();
    
    if (userResponse.success) {
      setState(() => _userProfile = userResponse.data);
    }
    
    if (walletResponse.success) {
      setState(() => _walletInfo = walletResponse.data);
    }
  }
}
```

---

#### 4.2 `account_profile_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `UserApi`

**Modifications :**
```dart
import 'package:mezadpay/services/user_api.dart';

Future<void> _updateProfile() async {
  final response = await _userApi.updateProfile(
    fullName: _nameController.text,
    email: _emailController.text,
  );
  
  if (response.success) {
    // Afficher succès
  }
}
```

---

### Module 5 : Wallet & Paiements ⚡ Priorité MOYENNE

#### 5.1 `deposit_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `WalletApi`

**Modifications :**
```dart
import 'package:mezadpay/services/wallet_api.dart';

Future<void> _deposit() async {
  final response = await _walletApi.deposit(
    amount: double.parse(_amountController.text),
    paymentMethod: _selectedMethodId,
  );
  
  if (response.success) {
    // Rediriger vers payment_success_page
  } else {
    // Afficher erreur
  }
}
```

---

#### 5.2 `withdraw_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `WalletApi`

**Modifications :**
```dart
import 'package:mezadpay/services/wallet_api.dart';

Future<void> _withdraw() async {
  final response = await _walletApi.withdraw(
    amount: double.parse(_amountController.text),
    bankDetails: _bankDetails,
  );
  
  if (response.success) {
    // Succès
  }
}
```

---

#### 5.3 `payment_details_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `WalletApi`

**Modifications :**
```dart
import 'package:mezadpay/services/wallet_api.dart';

Future<void> _uploadReceipt() async {
  final response = await _walletApi.uploadReceipt(
    transactionId: widget.transactionId,
    receiptImagePath: _receiptPath,
  );
  
  if (response.success) {
    // Succès
  }
}
```

---

### Module 6 : Notifications ⚡ Priorité MOYENNE

#### 6.1 `notifications_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `NotificationApi`

**Modifications :**
```dart
import 'package:mezadpay/services/notification_api.dart';

class _NotificationsPageState extends State<NotificationsPage> {
  final NotificationApi _notifApi = NotificationApi();
  List<Map<String, dynamic>> _notifications = [];
  
  @override
  void initState() {
    super.initState();
    _loadNotifications();
  }
  
  Future<void> _loadNotifications() async {
    final response = await _notifApi.getNotifications(limit: 20);
    
    if (response.success) {
      setState(() {
        _notifications = List.from(response.data!['notifications'] ?? []);
      });
    }
  }
  
  Future<void> _markAllAsRead() async {
    await _notifApi.markAllAsRead();
    _loadNotifications();
  }
}
```

**Ajouter FCM token dans `main.dart` ou `splash_page.dart` :**
```dart
import 'package:mezadpay/services/notification_api.dart';

Future<void> _saveFCMToken(String token) async {
  final notifApi = NotificationApi();
  await notifApi.saveFCMToken(
    fcmToken: token,
    deviceId: await _getDeviceId(),
    platform: Platform.isAndroid ? 'android' : 'ios',
  );
}
```

---

### Module 7 : Création d'Annonces ⚡ Priorité MOYENNE

#### 7.1 `create_ad_form_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `RequestApi`

**Modifications :**
```dart
import 'package:mezadpay/services/request_api.dart';

Future<void> _submitAd() async {
  final response = await _requestApi.createAuctionRequest(
    title: _nameController.text,
    description: _descriptionController.text,
    startingPrice: double.parse(_priceController.text),
    categoryId: _selectedCategoryId,
    location: _selectedCity,
  );
  
  if (response.success) {
    // Rediriger vers ad_success_page
  } else {
    // Afficher erreur
  }
}
```

---

### Module 8 : Autres Pages ⚡ Priorité BASSE

#### 8.1 `language_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `UserApi`

**Modifications :**
```dart
import 'package:mezadpay/services/user_api.dart';

Future<void> _changeLanguage(String language) async {
  final response = await _userApi.updateLanguage(language: language);
  
  if (response.success) {
    // Mettre à jour la locale
  }
}
```

---

#### 8.2 `my_auctions_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `UserApi`

**Modifications :**
```dart
import 'package:mezadpay/services/user_api.dart';

Future<void> _loadMyAuctions() async {
  final response = await _userApi.getMyAuctions(status: 'active');
  
  if (response.success) {
    setState(() {
      _myAuctions = List.from(response.data ?? []);
    });
  }
}
```

---

#### 8.3 `my_winnings_page.dart`
**État actuel :** Données mockées  
**Service à utiliser :** `UserApi`

**Modifications :**
```dart
import 'package:mezadpay/services/user_api.dart';

Future<void> _loadMyWinnings() async {
  final response = await _userApi.getMyWinnings();
  
  if (response.success) {
    setState(() {
      _myWinnings = List.from(response.data ?? []);
    });
  }
}
```

---

## 🔧 Étapes d'implémentation

### Phase 1 : Préparation
1. ✅ Installer les dépendances : `flutter pub get`
2. ✅ Configurer le fichier `.env`
3. ⏳ Générer les fichiers Riverpod : `flutter pub run build_runner build --delete-conflicting-outputs`

### Phase 2 : Authentification (Priorité HAUTE)
4. ⏳ Implémenter `login_page.dart`
5. ⏳ Implémenter `phone_registration_page.dart`
6. ⏳ Implémenter `otp_entry_page.dart`
7. ⏳ Implémenter `set_password_page.dart`

### Phase 3 : Enchères (Priorité HAUTE)
8. ⏳ Implémenter `home_page.dart`
9. ⏳ Implémenter `all_auctions_page.dart`
10. ⏳ Implémenter `auction_details_page.dart`
11. ⏳ Implémenter `auction_history_page.dart`

### Phase 4 : Favoris (Priorité HAUTE)
12. ⏳ Implémenter `favorites_page.dart`
13. ⏳ Mettre à jour le bouton favori dans `auction_details_page.dart`

### Phase 5 : Compte (Priorité MOYENNE)
14. ⏳ Implémenter `account_page.dart`
15. ⏳ Implémenter `account_profile_page.dart`
16. ⏳ Implémenter `language_page.dart`

### Phase 6 : Wallet (Priorité MOYENNE)
17. ⏳ Implémenter `deposit_page.dart`
18. ⏳ Implémenter `withdraw_page.dart`
19. ⏳ Implémenter `payment_details_page.dart`

### Phase 7 : Notifications (Priorité MOYENNE)
20. ⏳ Implémenter `notifications_page.dart`
21. ⏳ Implémenter FCM token dans `main.dart`

### Phase 8 : Autres (Priorité BASSE)
22. ⏳ Implémenter `create_ad_form_page.dart`
23. ⏳ Implémenter `my_auctions_page.dart`
24. ⏳ Implémenter `my_winnings_page.dart`

### Phase 9 : Tests
25. ⏳ Tester l'authentification
26. ⏳ Tester les enchères
27. ⏳ Tester les favoris
28. ⏳ Tester le wallet
29. ⏳ Tester les notifications

---

## 📊 Checklist de validation

### Pour chaque page :
- [ ] Importer le service API approprié
- [ ] Remplacer les données mockées par les appels API
- [ ] Ajouter un indicateur de chargement
- [ ] Gérer les erreurs avec un message utilisateur
- [ ] Tester le succès
- [ ] Tester l'échec (backend arrêté, mauvais token, etc.)

### Tests globaux :
- [ ] Authentification complète (register → OTP → login)
- [ ] Navigation entre les pages après login
- [ ] Token automatiquement ajouté aux requêtes
- [ ] Déconnexion supprime le token
- [ ] Erreur 401 redirige vers login
- [ ] Rate limit géré correctement

---

## 🎯 Points d'attention

1. **Loading states** : Toujours afficher un indicateur pendant les requêtes
2. **Error handling** : Afficher des messages clairs à l'utilisateur
3. **Token management** : Le token est automatiquement ajouté par l'interceptor
4. **UI inchangée** : Ne pas modifier l'UI, seulement remplacer les données
5. **Fallback** : Garder les données mockées en cas d'erreur si nécessaire
6. **Pagination** : Implémenter l'infinite scroll pour les listes longues
7. **Refresh** : Ajouter pull-to-refresh sur les listes
8. **Offline** : Gérer le mode hors-ligne (optionnel)

---

**Fin du plan d'implémentation**
