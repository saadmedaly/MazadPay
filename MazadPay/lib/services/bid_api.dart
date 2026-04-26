import 'package:mezadpay/models/models.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les enchères (bids)
class BidApi {
  final ApiService _apiService = ApiService();

  /// Placer une enchère sur une auction
  Future<ApiResponse<Map<String, dynamic>>> placeBid({
    required String auctionId,
    required double amount,
    bool? isAutoBid,
    double? maxAutoBidAmount,
  }) async {
    try {
      final data = <String, dynamic>{
        'amount': amount,
      };

      if (isAutoBid != null) data['is_auto_bid'] = isAutoBid;
      if (maxAutoBidAmount != null) data['max_auto_bid_amount'] = maxAutoBidAmount;

      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/bids',
        data: data,
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer l'historique des enchères d'une auction
  Future<ApiResponse<Map<String, dynamic>>> getBidHistory(String auctionId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$auctionId/bids',
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer les enchères de l'utilisateur connecté
  Future<ApiResponse<Map<String, dynamic>>> getMyBids() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/me/bids');

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer les gains de l'utilisateur (enchères remportées)
  Future<ApiResponse<Map<String, dynamic>>> getMyWinnings() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/me/winnings');

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Acheter maintenant (buy now)
  Future<ApiResponse<Map<String, dynamic>>> buyNow(String auctionId) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/buy-now',
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
