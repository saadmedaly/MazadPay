import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les sponsors
class SponsorApi {
  final ApiService _apiService = ApiService();
  
  /// Lister tous les sponsors actifs
  Future<ApiResponse<List<dynamic>>> getSponsors() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/sponsors',
      );
      
      final data = response?['data'];
      List<dynamic> sponsorList = [];
      if (data is Map && data.containsKey('sponsors')) {
        sponsorList = data['sponsors'] ?? [];
      } else if (response?.containsKey('sponsors') == true) {
        sponsorList = response?['sponsors'] ?? [];
      } else if (data is List) {
        sponsorList = data;
      }
      return ApiResponse.success(sponsorList);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
