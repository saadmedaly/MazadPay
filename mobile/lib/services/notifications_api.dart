import 'package:mezadpay/models/api_response.dart';
import 'api_service.dart';

/// Service API pour les notifications
class NotificationsApi {
  final ApiService _apiService = ApiService();

  /// Récupérer les notifications de l'utilisateur
  ///
  /// Live Staging root cause (client feedback, round 3): GET /notifications
  /// (notification_handler.go, after 30b0fec) returns a success envelope
  /// whose "data" field is a BARE ARRAY, not an object -- but this used to
  /// call `ApiResponse.fromJson` typed for a Map. `ApiResponse.fromJson`
  /// assigns the raw `json['data']` (a List at runtime) into a field typed
  /// for a Map; Dart's sound null safety inserts an implicit downcast at
  /// that assignment, which THROWS a TypeError for a List where a Map was
  /// expected. That throw was caught by this method's own try/catch and
  /// silently turned into an error response (success: false) -- so
  /// notifications_page.dart's success check was never entered, even though
  /// the backend had already correctly returned the real, persisted
  /// notification list (verified live: DB row existed, GET returned 200).
  /// Before 30b0fec this was masked by success defaulting to false for a
  /// different reason (missing "success" key) -- fixing that envelope bug
  /// is what exposed this pre-existing type mismatch. Declaring a List
  /// return type here (matching the real "data" shape, same pattern already
  /// used by getMyAuctions()/getBidHistory() elsewhere in this file) removes
  /// the mismatched cast entirely.
  Future<ApiResponse<List<dynamic>>> getNotifications() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/notifications',
      );
      final success = response?['success'] as bool? ?? false;
      final data = response?['data'] as List<dynamic>? ?? [];
      final message = response?['message'] as String?;
      return ApiResponse<List<dynamic>>(
        success: success,
        data: data,
        message: message,
      );
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
