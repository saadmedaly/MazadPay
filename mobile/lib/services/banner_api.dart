import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les bannières
class BannerApi {
  final ApiService _apiService = ApiService();
  
  /// Lister toutes les bannières actives
  Future<ApiResponse<List<dynamic>>> getBanners() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/banners',
      );
      
      final List<dynamic> bannerList = response?['data'] ?? response?['banners'] ?? [];
      return ApiResponse.success(bannerList);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
