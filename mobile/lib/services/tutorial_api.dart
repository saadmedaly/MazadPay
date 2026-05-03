import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

class TutorialApi {
  final ApiService _apiService = ApiService();

  Future<ApiResponse<List<dynamic>>> getTutorials() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/tutorials',
      );

      final List<dynamic> tutorialList = response?['data'] ?? response?['tutorials'] ?? response?['items'] ?? response ?? [];
      return ApiResponse.success(tutorialList);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
