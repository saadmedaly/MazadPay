import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les notifications
class NotificationApi {
  final ApiService _apiService = ApiService();
  
  /// Récupérer les notifications de l'utilisateur
  Future<ApiResponse<Map<String, dynamic>>> getNotifications({
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/notifications',
        queryParameters: {
          'page': page,
          'limit': limit,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Marquer une notification comme lue
  Future<ApiResponse<Map<String, dynamic>>> markNotificationAsRead(String id) async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/notifications/$id/read',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Marquer toutes les notifications comme lues
  Future<ApiResponse<Map<String, dynamic>>> markAllNotificationsAsRead() async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/notifications/read-all',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Sauvegarder le token FCM pour les notifications push
  Future<ApiResponse<Map<String, dynamic>>> saveToken({
    required String fcmToken,
    String? deviceId,
    String? platform, // 'web', 'android', 'ios'
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/notifications/token',
        data: {
          'fcm_token': fcmToken,
          if (deviceId != null) 'device_id': deviceId,
          if (platform != null) 'platform': platform,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
