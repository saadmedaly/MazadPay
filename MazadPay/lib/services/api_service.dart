import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:dio/dio.dart';
import 'package:mezadpay/services/auth_service.dart';
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
    
    // Ajouter les interceptors
    _dio.interceptors.add(AuthInterceptor());
    _dio.interceptors.add(ErrorInterceptor());
    
    // Logger interceptor pour le debug (à désactiver en production)
    _dio.interceptors.add(
      LogInterceptor(
        requestBody: true,
        responseBody: true,
        requestHeader: true,
        responseHeader: false,
        error: true,
      ),
    );
  }
  
  late final Dio _dio;
  
  Dio get dio => _dio;
  
  /// Méthode générique pour les requêtes GET
  Future<T?> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      final response = await _dio.get<T>(
        path,
        queryParameters: queryParameters,
        options: options,
      );
      return response.data;
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
        return const ApiException('Délai de connexion dépassé');
      
      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final message = error.response?.data?['error']?['message'] ?? 
                       error.response?.data?['message'] ?? 
                       'Erreur serveur';
        
        switch (statusCode) {
          case 400:
            return ApiException(message);
          case 401:
            return UnauthorizedException('Non autorisé');
          case 403:
            return ForbiddenException('Accès refusé');
          case 404:
            return NotFoundException('Ressource non trouvée');
          case 429:
            return RateLimitException('Trop de requêtes. Veuillez réessayer plus tard.');
          case 500:
            return ServerException('Erreur serveur interne');
          default:
            return ApiException(message);
        }
      
      case DioExceptionType.cancel:
        return const ApiException('Requête annulée');
      
      case DioExceptionType.unknown:
        return ApiException('Erreur de connexion: ${error.message}');
      
      default:
        return ApiException('Erreur inconnue: ${error.message}');
    }
  }
}

/// Exceptions personnalisées pour l'API
class ApiException implements Exception {
  final String message;
  const ApiException(this.message);
  
  @override
  String toString() => message;
}

class UnauthorizedException extends ApiException {
  const UnauthorizedException(String message) : super(message);
}

class ForbiddenException extends ApiException {
  const ForbiddenException(String message) : super(message);
}

class NotFoundException extends ApiException {
  const NotFoundException(String message) : super(message);
}

class RateLimitException extends ApiException {
  const RateLimitException(String message) : super(message);
}

class ServerException extends ApiException {
  const ServerException(String message) : super(message);
}
