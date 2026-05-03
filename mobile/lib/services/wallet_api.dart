import 'package:dio/dio.dart';
import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour le wallet
class WalletApi {
  final ApiService _apiService = ApiService();
  
  /// Récupérer le solde du wallet
  Future<ApiResponse<Map<String, dynamic>>> getWalletBalance() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/wallet');

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Alias pour getWalletBalance (utilisé par les pages)
  Future<ApiResponse<Map<String, dynamic>>> getBalance() async {
    return getWalletBalance();
  }
  
  /// Effectuer un dépôt
  Future<ApiResponse<Map<String, dynamic>>> deposit({
    required double amount,
    String? paymentMethod,
    String? method,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/users/wallet/deposit',
        data: {
          'amount': amount,
          'payment_method': paymentMethod ?? method,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Effectuer un retrait
  Future<ApiResponse<Map<String, dynamic>>> withdraw({
    required double amount,
    Map<String, dynamic>? bankDetails,
    String? method,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/users/wallet/withdraw',
        data: {
          'amount': amount,
          if (bankDetails != null) 'bank_details': bankDetails,
          if (method != null) 'method': method,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer l'historique des transactions
  Future<ApiResponse<Map<String, dynamic>>> getTransactions({
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/wallet/transactions',
        queryParameters: {
          'page': page,
          'limit': limit,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer les détails d'une transaction
  Future<ApiResponse<Map<String, dynamic>>> getTransactionDetails(String id) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/wallet/transactions/$id',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Uploader un reçu de paiement
  Future<ApiResponse<Map<String, dynamic>>> uploadReceipt({
    required String transactionId,
    required String receiptImagePath,
  }) async {
    try {
      // Note: Ceci nécessite l'implémentation du multipart upload
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/users/wallet/transactions/$transactionId/receipt',
        data: FormData.fromMap({
          'receipt': await MultipartFile.fromFile(receiptImagePath),
        }),
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
