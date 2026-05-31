import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les bannières
class BannerApi {
  final ApiService _apiService = ApiService();
  
  /// Lister toutes les bannières actives
  Future<ApiResponse<List<dynamic>>> getBanners() async {
    try {
      final response = await _apiService.get<dynamic>('/banners');

      List<dynamic> bannerList = [];
      if (response is List) {
        // Direct array response
        bannerList = response;
      } else if (response is Map<String, dynamic>) {
        // Wrapped response: { success, data: [...] }
        final data = response['data'];
        if (data is List) {
          bannerList = data;
        } else if (response['banners'] is List) {
          bannerList = response['banners'] as List;
        }
      }

      return ApiResponse.success(bannerList);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
