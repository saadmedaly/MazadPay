import 'package:mezadpay/models/api_response.dart';
import 'api_service.dart';

/// Service API pour les favoris
class FavoritesApi {
  final ApiService _apiService = ApiService();

  /// Récupérer les favoris de l'utilisateur.
  ///
  /// Backend (GET /users/me/favorites) renvoie {"success":true,"data":[...]}
  /// où "data" est un tableau JSON brut de mazads (auctions) déjà enrichis
  /// via JOIN -- jamais {"favorites": [...]}. Voir UserHandler.ListFavorites /
  /// favoriteRepo.ListByUserID côté backend (client feedback: Bug C).
  /// `ApiResponse` is typed with `List<dynamic>` so `data` matches the array
  /// FavoritesService actually receives, instead of the previous
  /// `Map<String, dynamic>` typing, which crashed the moment anything tried
  /// to read `data['favorites']` on what was really a List.
  Future<ApiResponse<List<dynamic>>> getFavorites() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/me/favorites',
      );
      return ApiResponse<List<dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Ajouter une enchère aux favoris
  Future<ApiResponse<Map<String, dynamic>>> addFavorite(String auctionId) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/users/me/favorites/$auctionId',
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
        '/users/me/favorites/$auctionId',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
