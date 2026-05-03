import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les demandes (auction requests, banner requests)
class RequestApi {
  final ApiService _apiService = ApiService();
  
  /// Créer une demande d'enchère
  Future<ApiResponse<Map<String, dynamic>>> createAuctionRequest({
    required String title,
    required String description,
    required double startingPrice,
    required String categoryId,
    required String location,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/requests/auctions',
        data: {
          'title': title,
          'description': description,
          'starting_price': startingPrice,
          'category_id': categoryId,
          'location': location,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Créer une demande de bannière
  Future<ApiResponse<Map<String, dynamic>>> createBannerRequest({
    required String titleAr,
    required String titleFr,
    required String titleEn,
    required String descriptionAr,
    required String descriptionFr,
    required String descriptionEn,
    required String imageUrl,
    required String linkUrl,
    required DateTime startsAt,
    required DateTime endsAt,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/requests/banners',
        data: {
          'title_ar': titleAr,
          'title_fr': titleFr,
          'title_en': titleEn,
          'description_ar': descriptionAr,
          'description_fr': descriptionFr,
          'description_en': descriptionEn,
          'image_url': imageUrl,
          'link_url': linkUrl,
          'starts_at': startsAt.toIso8601String(),
          'ends_at': endsAt.toIso8601String(),
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer mes demandes d'enchères
  Future<ApiResponse<Map<String, dynamic>>> getMyAuctionRequests({String? status}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (status != null) queryParams['status'] = status;
      
      final response = await _apiService.get<Map<String, dynamic>>(
        '/requests/auctions/my',
        queryParameters: queryParams,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer mes demandes de bannières
  Future<ApiResponse<Map<String, dynamic>>> getMyBannerRequests({String? status}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (status != null) queryParams['status'] = status;
      
      final response = await _apiService.get<Map<String, dynamic>>(
        '/requests/banners/my',
        queryParameters: queryParams,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
