import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:dio/dio.dart';
import 'package:mezadpay/services/interceptors/auth_interceptor.dart';
import 'package:mezadpay/services/interceptors/error_interceptor.dart';

/// Builds the Options passed to Dio for ApiService.get/post, merging
/// `skipAuth` into `extra` so AuthInterceptor.onRequest can read it.
/// Extracted as a pure function so the skipAuth wiring itself is directly
/// unit-testable without a live Dio instance -- see test/skip_auth_test.dart.
@visibleForTesting
Options buildRequestOptions({Options? options, required bool skipAuth}) {
  return (options ?? Options()).copyWith(
    extra: {...?options?.extra, if (skipAuth) 'skipAuth': true},
  );
}

/// Maps a DioExceptionType to the safe category string used as
/// ApiException.code. Extracted as a pure top-level function (rather than
/// staying inline in the private `ApiService._handleError`) so it is
/// directly unit-testable without needing a live Dio instance/network call
/// -- see test/api_error_category_test.dart.
@visibleForTesting
String dioExceptionTypeCode(DioExceptionType type) {
  switch (type) {
    case DioExceptionType.connectionTimeout:
      return 'connectionTimeout';
    case DioExceptionType.sendTimeout:
      return 'sendTimeout';
    case DioExceptionType.receiveTimeout:
      return 'receiveTimeout';
    case DioExceptionType.connectionError:
      return 'connectionError';
    case DioExceptionType.badCertificate:
      return 'badCertificate';
    case DioExceptionType.badResponse:
      return 'badResponse';
    case DioExceptionType.cancel:
      return 'cancel';
    case DioExceptionType.unknown:
      return 'unknown';
  }
}

 
class ApiService {
  static final ApiService _instance = ApiService._internal();
  factory ApiService() => _instance;
  
  ApiService._internal() {
    // Charger les variables d'environnement
    final baseUrl = dotenv.env['API_BASE_URL'] ?? 'http://localhost:8082/v1/api';
    
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
      sendTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ));
    
    // Configurer l'URL WebSocket
    _wsBaseUrl = dotenv.env['WS_BASE_URL'] ?? 'ws://localhost:8082/v1/api';
    
    // Ajouter les interceptors
    _dio.interceptors.add(AuthInterceptor());
    _dio.interceptors.add(ErrorInterceptor());
    
    // Logger interceptor - minimal logging only
    _dio.interceptors.add(
      LogInterceptor(
        requestBody: false,
        responseBody: false,
        requestHeader: false,
        responseHeader: false,
        error: true,
        logPrint: (obj) => debugPrint('[API] $obj'),
      ),
    );
  }
  
  late final Dio _dio;
  static late final String _wsBaseUrl;
  
  Dio get dio => _dio;
  static String get wsBaseUrl => _wsBaseUrl;
  static String get apiBaseUrl => dotenv.env['API_BASE_URL'] ?? 'http://localhost:8082/v1/api';
  
  /// Méthode générique pour les requêtes GET
  ///
  /// [skipAuth] lets a caller mark a request as not needing auth --
  /// AuthInterceptor.onRequest reads this via `options.extra['skipAuth']`
  /// and, when true, skips its two FlutterSecureStorage reads
  /// (getToken/getUserId) entirely before calling handler.next(). Reuses
  /// Dio's own Options.extra request-metadata channel rather than inventing
  /// a parallel mechanism. Only used for backend routes verified public by
  /// contract -- see CategoryApi.getCountries (GET /countries,
  /// backend/internal/routes/routes.go:229 -- api.Get("/countries",
  /// publicListLimit, cache5min, h.GetCountries), no jwtMiddleware/OptionalJWT).
  Future<T?> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
    bool skipAuth = false,
  }) async {
    try {
      final effectiveOptions = buildRequestOptions(options: options, skipAuth: skipAuth);
      final response = await _dio.get<String>(
        path,
        queryParameters: queryParameters,
        options: effectiveOptions,
      );
      // Parse JSON manually to avoid web platform type issues
      if (response.data != null) {
        final decoded = jsonDecode(response.data!);
        return decoded as T?;
      }
      return null;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }
  
  /// Méthode générique pour les requêtes POST
  ///
  /// [skipAuth] mirrors the same mechanism added to get() -- see that
  /// method's doc comment. Used by AuthApi's verified-public auth endpoints
  /// (register/login/otp/reset-password) so they don't wait on
  /// AuthInterceptor's secure-storage reads, matching the backend route
  /// contract (backend/internal/routes/routes.go setupAuthRoutes/
  /// setupAuthRoutesV2 -- only /auth/logout and /auth/change-password carry
  /// jwtMiddleware).
  Future<T?> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
    bool skipAuth = false,
  }) async {
    try {
      final effectiveOptions = buildRequestOptions(options: options, skipAuth: skipAuth);
      final response = await _dio.post<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: effectiveOptions,
      );
      return response.data;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }
  
  /// Méthode générique pour les requêtes PUT
  Future<T?> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      final response = await _dio.put<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
      return response.data;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }
  
  /// Méthode générique pour les requêtes DELETE
  Future<T?> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      final response = await _dio.delete<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
      return response.data;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }
  
  /// Méthode pour upload de fichiers (multipart)
  Future<T?> upload<T>(
    String path, {
    required FormData data,
    Options? options,
  }) async {
    try {
      final response = await _dio.post<T>(
        path,
        data: data,
        options: options,
      );
      return response.data;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }
  
  /// Gestion centralisée des erreurs
  ///
  /// `code` is split per DioExceptionType (rather than collapsing every
  /// timeout variant and several other types into a single generic
  /// 'connection_error'/'unknown_error' string) so a caller can distinguish
  /// timeout vs. connection reset vs. TLS vs. other failure categories. No
  /// behavior change for callers that only branch on exception TYPE
  /// (ApiException vs UnauthorizedException, etc.) -- only the `code` string
  /// value is more specific.
  Exception _handleError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return ApiException('Délai de connexion dépassé', code: dioExceptionTypeCode(error.type));

      case DioExceptionType.connectionError:
        return ApiException('Erreur de connexion: ${error.message}', code: dioExceptionTypeCode(error.type));

      case DioExceptionType.badCertificate:
        return ApiException('Erreur de certificat', code: dioExceptionTypeCode(error.type));

      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final errorData = error.response?.data?['error'];
        final code = errorData?['code'] as String? ?? _codeForStatus(statusCode);
        final message = errorData?['message'] ?? 
                       error.response?.data?['message'] ?? 
                       'Erreur serveur';
        
        switch (statusCode) {
          case 400:
            return ApiException(message, code: code, httpStatus: statusCode);
          case 401:
            return UnauthorizedException(message, code: code, httpStatus: statusCode);
          case 403:
            return ForbiddenException(message, code: code, httpStatus: statusCode);
          case 404:
            return NotFoundException(message, code: code, httpStatus: statusCode);
          case 409:
            return ApiException(message, code: code, httpStatus: statusCode);
          case 422:
            return ApiException(message, code: code, httpStatus: statusCode);
          case 429:
            return RateLimitException(message, code: code, httpStatus: statusCode);
          case 500:
            return ServerException(message, code: code, httpStatus: statusCode);
          default:
            return ApiException(message, code: code, httpStatus: statusCode);
        }
      
      case DioExceptionType.cancel:
        return ApiException('Requête annulée', code: dioExceptionTypeCode(error.type));

      case DioExceptionType.unknown:
        return ApiException('Erreur de connexion: ${error.message}', code: dioExceptionTypeCode(error.type));
    }
  }
  
  String _codeForStatus(int? statusCode) {
    switch (statusCode) {
      case 400: return 'bad_request';
      case 401: return 'unauthorized';
      case 403: return 'forbidden';
      case 404: return 'not_found';
      case 409: return 'conflict';
      case 422: return 'validation_error';
      case 429: return 'too_many_requests';
      case 500: return 'server_error';
      default:  return 'error';
    }
  }
}

/// Exceptions personnalisées pour l'API
class ApiException implements Exception {
  final String message;
  final String? code;
  // Optional HTTP status, populated only for badResponse -- lets a caller
  // distinguish e.g. a 500 from a 404 without parsing `code`.
  final int? httpStatus;
  const ApiException(this.message, {this.code, this.httpStatus});

  @override
  String toString() => message;
}

class UnauthorizedException extends ApiException {
  const UnauthorizedException(super.message, {super.code, super.httpStatus});
}

class ForbiddenException extends ApiException {
  const ForbiddenException(super.message, {super.code, super.httpStatus});
}

class NotFoundException extends ApiException {
  const NotFoundException(super.message, {super.code, super.httpStatus});
}

class RateLimitException extends ApiException {
  const RateLimitException(super.message, {super.code, super.httpStatus});
}

class ServerException extends ApiException {
  const ServerException(super.message, {super.code, super.httpStatus});
}
