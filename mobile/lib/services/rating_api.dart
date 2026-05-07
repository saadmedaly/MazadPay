import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les évaluations de l'application
class RatingApi {
  final ApiService _apiService = ApiService();

  /// Soumettre une évaluation de l'application
  Future<ApiResponse<Map<String, dynamic>>> createAppRating({
    required int rating,
    String? comment,
    String title = "App Rating",
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/app/ratings',
        data: {
          'title': title,
          'rating': rating,
          'comment': comment ?? '',
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
