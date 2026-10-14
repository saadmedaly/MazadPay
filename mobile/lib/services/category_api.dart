import 'dart:async';
import 'package:flutter/foundation.dart';
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
      debugPrint('Category API error: $e');
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
  ///
  /// On failure, the returned ApiResponse's ApiError carries a safe category
  /// via `code` (a DioExceptionType-derived value such as
  /// 'connectionTimeout'/'connectionError'/'badCertificate', or
  /// 'invalid_shape'/'null_response' for a non-network parsing failure) plus
  /// `details['httpStatus']` when the failure was an HTTP error response.
  /// Never includes exception messages that could carry response bodies --
  /// only the safe category/status.
  Future<ApiResponse<List<dynamic>>> getCountries() async {
    try {
      // /countries is a verified public, pre-login backend route
      // (routes.go:229, no jwtMiddleware/OptionalJWT) and must never wait on
      // secure-storage token/userId reads (see AuthInterceptor.onRequest).
      // Wrapped in a total 35s Future.timeout as defense in depth: Dio's own
      // 30s connect/send/receive timeouts do not cover time spent inside
      // interceptors before handler.next(), so without this the whole call
      // could otherwise hang indefinitely if any stage ahead of Dio's timed
      // phases ever stalls. On timeout, the caller (LoginPage/
      // PhonePasswordPage/etc.) receives the same failure-shaped
      // ApiResponse.error(...) as any other network failure and follows the
      // existing MR-fallback path -- the UI must never stay stuck loading.
      final response = await _apiService
          .get<dynamic>('/countries', skipAuth: true)
          .timeout(const Duration(seconds: 35));

      if (response == null) {
        return ApiResponse.error('Null response from server', code: 'null_response');
      }

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
        return ApiResponse.error(
          'Invalid response format: expected Map, got ${response.runtimeType}',
          code: 'invalid_shape',
        );
      }
    } on TimeoutException {
      // The total-call timeout above fired -- distinct from a DioException,
      // since it means no exception from the network layer had propagated
      // within 35s (the "stuck on loading" real-device symptom).
      return ApiResponse.error('Countries request timed out', code: 'totalCallTimeout');
    } on ApiException catch (e) {
      return ApiResponse.error(
        e.code ?? 'unknown',
        code: e.code ?? 'unknown',
        details: e.httpStatus != null ? {'httpStatus': e.httpStatus} : null,
      );
    } catch (e) {
      return ApiResponse.error(e.runtimeType.toString(), code: 'parse_error');
    }
  }

  /// Récupérer les raisons de signalement
  Future<ApiResponse<List<dynamic>>> getReportReasons() async {
    try {
      final response = await _apiService.get<dynamic>('/report-reasons');

      if (response == null) {
        return ApiResponse.error('Null response from server');
      }

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
        return ApiResponse.error('Invalid response format');
      }
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
