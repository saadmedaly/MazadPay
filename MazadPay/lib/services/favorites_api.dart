import 'package:mezadpay/models/api_response.dart';
import 'api_service.dart';

/// Service API pour les favoris
class FavoritesApi {
  final ApiService _apiService = ApiService();

  /// Récupérer les favoris de l'utilisateur
  Future<ApiResponse<Map<String, dynamic>>> getFavorites() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/favorites',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Ajouter une enchère aux favoris
  Future<ApiResponse<Map<String, dynamic>>> addFavorite(String auctionId) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/favorites',
        data: {'auction_id': auctionId},
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Supprimer une enchère des favoris
  Future<ApiResponse<Map<String, dynamic>>> removeFavorite(String auctionId) async {
    try {
      final response = await _apiService.delete<Map<String, dynamic>>(
        '/favorites/$auctionId',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
