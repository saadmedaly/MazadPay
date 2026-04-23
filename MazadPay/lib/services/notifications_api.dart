import 'package:mezadpay/models/api_response.dart';
import 'api_service.dart';

/// Service API pour les notifications
class NotificationsApi {
  final ApiService _apiService = ApiService();

  /// Récupérer les notifications de l'utilisateur
  Future<ApiResponse<Map<String, dynamic>>> getNotifications() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/notifications',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Marquer toutes les notifications comme lues
  Future<ApiResponse<Map<String, dynamic>>> markAllAsRead() async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/notifications/read-all',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
