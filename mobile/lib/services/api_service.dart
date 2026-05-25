import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:dio/dio.dart';
import 'package:mezadpay/services/interceptors/auth_interceptor.dart';
import 'package:mezadpay/services/interceptors/error_interceptor.dart';

 
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
  Future<T?> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      final response = await _dio.get<String>(
        path,
        queryParameters: queryParameters,
        options: options,
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
  Future<T?> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      final response = await _dio.post<T>(
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
  Exception _handleError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return const ApiException('Délai de connexion dépassé', code: 'connection_error');
      
      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final errorData = error.response?.data?['error'];
        final code = errorData?['code'] as String? ?? _codeForStatus(statusCode);
        final message = errorData?['message'] ?? 
                       error.response?.data?['message'] ?? 
                       'Erreur serveur';
        
        switch (statusCode) {
          case 400:
            return ApiException(message, code: code);
          case 401:
            return UnauthorizedException(message, code: code);
          case 403:
            return ForbiddenException(message, code: code);
          case 404:
            return NotFoundException(message, code: code);
          case 409:
            return ApiException(message, code: code);
          case 422:
            return ApiException(message, code: code);
          case 429:
            return RateLimitException(message, code: code);
          case 500:
            return ServerException(message, code: code);
          default:
            return ApiException(message, code: code);
        }
      
      case DioExceptionType.cancel:
        return const ApiException('Requête annulée', code: 'cancelled');
      
      case DioExceptionType.unknown:
        return ApiException('Erreur de connexion: ${error.message}', code: 'connection_error');
      
      default:
        return ApiException('Erreur inconnue: ${error.message}', code: 'unknown_error');
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
  const ApiException(this.message, {this.code});
  
  @override
  String toString() => message;
}

class UnauthorizedException extends ApiException {
  const UnauthorizedException(super.message, {super.code});
}

class ForbiddenException extends ApiException {
  const ForbiddenException(super.message, {super.code});
}

class NotFoundException extends ApiException {
  const NotFoundException(super.message, {super.code});
}

class RateLimitException extends ApiException {
  const RateLimitException(super.message, {super.code});
}

class ServerException extends ApiException {
  const ServerException(super.message, {super.code});
}
