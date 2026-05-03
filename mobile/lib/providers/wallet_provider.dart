import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mezadpay/models/models.dart';
import 'package:mezadpay/services/wallet_api.dart';

part 'wallet_provider.g.dart';

/// État du wallet
class WalletState {
  final Wallet? wallet;
  final List<Transaction> transactions;
  final bool isLoading;
  final String? error;
  final bool isLoadingTransactions;
  final int currentPage;
  final bool hasMoreTransactions;

  WalletState({
    this.wallet,
    this.transactions = const [],
    this.isLoading = false,
    this.error,
    this.isLoadingTransactions = false,
    this.currentPage = 1,
    this.hasMoreTransactions = true,
  });

  WalletState copyWith({
    Wallet? wallet,
    List<Transaction>? transactions,
    bool? isLoading,
    String? error,
    bool? isLoadingTransactions,
    int? currentPage,
    bool? hasMoreTransactions,
  }) {
    return WalletState(
      wallet: wallet ?? this.wallet,
      transactions: transactions ?? this.transactions,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      isLoadingTransactions: isLoadingTransactions ?? this.isLoadingTransactions,
      currentPage: currentPage ?? this.currentPage,
      hasMoreTransactions: hasMoreTransactions ?? this.hasMoreTransactions,
    );
  }

  /// Solde disponible
  double get availableBalance => wallet?.availableBalance ?? 0.0;

  /// Solde total
  double get totalBalance => wallet?.balance ?? 0.0;

  /// Montant gelé
  double get frozenAmount => wallet?.frozenAmount ?? 0.0;
}

/// Provider pour la gestion du wallet
@riverpod
class WalletNotifier extends _$WalletNotifier {
  final WalletApi _walletApi = WalletApi();

  @override
  WalletState build() {
    return WalletState();
  }

  /// Charger le solde du wallet
  Future<void> loadWallet() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _walletApi.getWalletBalance();

      if (response.success && response.data != null) {
        final wallet = Wallet.fromJson(response.data!);
        state = state.copyWith(
          isLoading: false,
          wallet: wallet,
          error: null,
        );
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de chargement du solde',
        );
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Charger l'historique des transactions
  Future<void> loadTransactions({
    int page = 1,
    int limit = 20,
    bool refresh = false,
  }) async {
    if (refresh) {
      state = state.copyWith(
        isLoadingTransactions: true,
        error: null,
        currentPage: 1,
        transactions: [],
      );
    } else {
      state = state.copyWith(isLoadingTransactions: true, error: null);
    }

    try {
      final response = await _walletApi.getTransactions(
        page: refresh ? 1 : page,
        limit: limit,
      );

      if (response.success && response.data != null) {
        final data = response.data!;
        final transactionsList = data['transactions'] as List<dynamic>? ?? [];
        final transactions = transactionsList
            .map((e) => Transaction.fromJson(e as Map<String, dynamic>))
            .toList();

        final total = data['total'] ?? 0;
        final currentPage = data['page'] ?? page;

        state = state.copyWith(
          isLoadingTransactions: false,
          transactions: refresh
              ? transactions
              : [...state.transactions, ...transactions],
          currentPage: currentPage,
          hasMoreTransactions: transactions.length >= limit,
          error: null,
        );
      } else {
        state = state.copyWith(
          isLoadingTransactions: false,
          error: response.error?.message ?? 'Erreur de chargement des transactions',
        );
      }
    } catch (e) {
      state = state.copyWith(
        isLoadingTransactions: false,
        error: e.toString(),
      );
    }
  }

  /// Charger plus de transactions (pagination)
  Future<void> loadMoreTransactions() async {
    if (state.isLoadingTransactions || !state.hasMoreTransactions) return;

    await loadTransactions(page: state.currentPage + 1);
  }

  /// Effectuer un dépôt
  Future<bool> deposit({
    required double amount,
    String? paymentMethod,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _walletApi.deposit(
        amount: amount,
        paymentMethod: paymentMethod,
      );

      if (response.success) {
        // Recharger le wallet après le dépôt
        await loadWallet();
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de dépôt',
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Effectuer un retrait
  Future<bool> withdraw({
    required double amount,
    Map<String, dynamic>? bankDetails,
    String? method,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _walletApi.withdraw(
        amount: amount,
        bankDetails: bankDetails,
        method: method,
      );

      if (response.success) {
        // Recharger le wallet après le retrait
        await loadWallet();
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de retrait',
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Uploader un reçu de paiement
  Future<bool> uploadReceipt({
    required String transactionId,
    required String receiptImagePath,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _walletApi.uploadReceipt(
        transactionId: transactionId,
        receiptImagePath: receiptImagePath,
      );

      state = state.copyWith(isLoading: false);
      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Récupérer les détails d'une transaction
  Future<Transaction?> getTransactionDetails(String id) async {
    try {
      final response = await _walletApi.getTransactionDetails(id);

      if (response.success && response.data != null) {
        return Transaction.fromJson(response.data!);
      }
      return null;
    } catch (e) {
      return null;
    }
  }

  /// Rafraîchir tout
  Future<void> refresh() async {
    await loadWallet();
    await loadTransactions(refresh: true);
  }

  /// Effacer les erreurs
  void clearError() {
    state = state.copyWith(error: null);
  }
}
