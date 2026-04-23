import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les enchères
class AuctionApi {
  final ApiService _apiService = ApiService();
  
  /// Lister toutes les enchères
  Future<ApiResponse<Map<String, dynamic>>> getAuctions({
    int page = 1,
    int limit = 20,
    String? categoryId,
    String? locationId,
    String? status,
  }) async {
    try {
      final queryParams = <String, dynamic>{
        'page': page,
        'limit': limit,
      };
      
      if (categoryId != null) queryParams['category_id'] = categoryId;
      if (locationId != null) queryParams['location_id'] = locationId;
      if (status != null) queryParams['status'] = status;
      
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions',
        queryParameters: queryParams,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer les détails d'une enchère
  Future<ApiResponse<Map<String, dynamic>>> getAuctionById(String id) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$id',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Alias pour getAuctionById (utilisé par les pages)
  Future<ApiResponse<Map<String, dynamic>>> getAuctionDetails(String id) async {
    return getAuctionById(id);
  }

  /// Créer une nouvelle enchère
  Future<ApiResponse<Map<String, dynamic>>> createAuction({
    required String title,
    required String description,
    required double startingPrice,
    required String category,
    required String subCategory,
    required String location,
    required List<String> images,
    required String phone,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions',
        data: {
          'title': title,
          'description': description,
          'starting_price': startingPrice,
          'category': category,
          'sub_category': subCategory,
          'location': location,
          'images': images,
          'phone': phone,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Incrémenter le compteur de vues
  Future<ApiResponse<Map<String, dynamic>>> incrementViews(String id) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$id/view',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Placer une enchère
  Future<ApiResponse<Map<String, dynamic>>> placeBid({
    required String auctionId,
    required double amount,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/bids',
        data: {
          'amount': amount,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer l'historique des enchères
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
  
  /// Obtenir le contact du vendeur
  Future<ApiResponse<Map<String, dynamic>>> getSellerContact(String auctionId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$auctionId/seller-contact',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les catégories
  Future<ApiResponse<Map<String, dynamic>>> getCategories() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/categories');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les locations
  Future<ApiResponse<Map<String, dynamic>>> getLocations({String? countryId}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (countryId != null) queryParams['country_id'] = countryId;
      
      final response = await _apiService.get<Map<String, dynamic>>(
        '/locations',
        queryParameters: queryParams,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les pays
  Future<ApiResponse<Map<String, dynamic>>> getCountries() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/countries');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
