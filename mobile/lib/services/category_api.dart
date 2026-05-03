import 'dart:convert';
import 'package:mezadpay/models/models.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les catégories, locations et pays
class CategoryApi {
  final ApiService _apiService = ApiService();

  /// Lister toutes les catégories
  Future<ApiResponse<List<dynamic>>> getCategories({int? limit}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (limit != null) queryParams['limit'] = limit;

      final response = await _apiService.get<dynamic>(
        '/categories',
        queryParameters: queryParams.isNotEmpty ? queryParams : null,
      );

      print('=== CATEGORY API DEBUG ===');
      print('Response type: ${response.runtimeType}');

      if (response == null) {
        return ApiResponse.error('Null response from server');
      }

      // Directly create ApiResponse without converting the Map
      if (response is Map) {
        final success = (response['success'] as bool?) ?? false;
        final data = response['data'];
        final errorData = response['error'];
        final message = response['message'] as String?;
        
        ApiError? error;
        if (errorData != null && errorData is Map) {
          error = ApiError.fromJson(errorData as Map<String, dynamic>);
        }
        
        return ApiResponse<List<dynamic>>(
          success: success,
          data: data as List<dynamic>?,
          error: error,
          message: message,
        );
      } else {
        return ApiResponse.error('Invalid response format: expected Map, got ${response.runtimeType}');
      }
    } catch (e) {
      print('Category API error: $e');
      return ApiResponse.error(e.toString());
    }
  }

  /// Lister toutes les locations (villes)
  Future<ApiResponse<List<dynamic>>> getLocations({String? countryId}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (countryId != null) queryParams['country_id'] = countryId;

      final response = await _apiService.get<dynamic>(
        '/locations',
        queryParameters: queryParams.isNotEmpty ? queryParams : null,
      );

      if (response == null) {
        return ApiResponse.error('Null response from server');
      }

      // Directly create ApiResponse similar to getCategories
      if (response is Map) {
        final success = (response['success'] as bool?) ?? false;
        final data = response['data'];
        final errorData = response['error'];
        final message = response['message'] as String?;
        
        ApiError? error;
        if (errorData != null && errorData is Map) {
          error = ApiError.fromJson(errorData as Map<String, dynamic>);
        }
        
        return ApiResponse<List<dynamic>>(
          success: success,
          data: data as List<dynamic>?,
          error: error,
          message: message,
        );
      } else {
        return ApiResponse.error('Invalid response format: expected Map, got ${response.runtimeType}');
      }
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Lister les locations par pays
  Future<ApiResponse<Map<String, dynamic>>> getLocationsByCountry(String countryId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/locations/$countryId',
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Lister tous les pays
  Future<ApiResponse<Map<String, dynamic>>> getCountries() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/countries');

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer les raisons de signalement
  Future<ApiResponse<Map<String, dynamic>>> getReportReasons() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/report-reasons');

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Lister les sous-catégories d'une catégorie
  Future<ApiResponse<Map<String, dynamic>>> getSubCategories(String categoryId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/categories/$categoryId/sub-categories',
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
