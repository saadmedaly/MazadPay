// lib/data/datasources/remote/api_client.dart
// Client API avec Dio - configuration complète pour MazadPay

import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/api_constants.dart';
import '../../../core/errors/exceptions.dart';
import '../../repositories/auth_repository.dart';

// Provider pour le client API
final apiClientProvider = Provider<Dio>((ref) {
  final authRepo = ref.watch(authRepositoryProvider);
  return ApiClient(authRepo: authRepo).dio;
});

class ApiClient {
  late final Dio dio;
  final AuthRepository authRepo;
  
  ApiClient({required this.authRepo}) {
    dio = Dio(
      BaseOptions(
        baseUrl: ApiConstants.baseUrl,
        connectTimeout: const Duration(milliseconds: ApiConstants.connectTimeout),
        receiveTimeout: const Duration(milliseconds: ApiConstants.receiveTimeout),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        validateStatus: (status) => status! < 500,
      ),
    );
    
    _setupInterceptors();
  }
  
  void _setupInterceptors() {
    // Interceptor pour logging (debug uniquement)
    if (kDebugMode) {
      dio.interceptors.add(
        LogInterceptor(
          requestBody: true,
          responseBody: true,
          logPrint: (object) => debugPrint('[DIO] $object'),
        ),
      );
    }
    
    // Interceptor pour l'authentification
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          // Ajouter le token JWT si disponible
          final token = await authRepo.getToken();
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          
          // Ajouter la langue préférée
          final lang = await authRepo.getLanguagePreference();
          options.headers['Accept-Language'] = lang;
          
          return handler.next(options);
        },
        onError: (error, handler) async {
          // Gérer le 401 - Token expiré
          if (error.response?.statusCode == 401) {
            try {
              final refreshed = await authRepo.refreshToken();
              if (refreshed) {
                // Réessayer la requête avec le nouveau token
                final opts = error.requestOptions;
                final token = await authRepo.getToken();
                opts.headers['Authorization'] = 'Bearer $token';
                
                final response = await dio.fetch(opts);
                return handler.resolve(response);
              } else {
                // Rediriger vers login si refresh échoue
                await authRepo.logout();
                throw UnauthorizedException();
              }
            } catch (e) {
              await authRepo.logout();
              throw UnauthorizedException();
            }
          }
          
          return handler.next(error);
        },
      ),
    );
    
    // Interceptor pour retry automatique
    dio.interceptors.add(
      RetryInterceptor(
        dio: dio,
        retries: 3,
        retryDelays: const [
          Duration(seconds: 1),
          Duration(seconds: 2),
          Duration(seconds: 3),
        ],
      ),
    );
  }
  
  // Gestion centralisée des erreurs
  static AppException handleError(dynamic error) {
    if (error is DioException) {
      switch (error.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
          return const TimeoutException();
          
        case DioExceptionType.connectionError:
          return const NetworkException();
          
        case DioExceptionType.badResponse:
          final statusCode = error.response?.statusCode;
          final data = error.response?.data;
          final message = data?['message'] ?? data?['error'] ?? 'Server error';
          
          switch (statusCode) {
            case 400:
              return ValidationException(message, errors: data?['errors']);
            case 401:
              return const UnauthorizedException();
            case 403:
              return const UnauthorizedException('Access forbidden');
            case 404:
              return const NotFoundException();
            case 409:
              return ConflictException(message);
            case 422:
              return ValidationException(message, errors: data?['errors']);
            default:
              return ServerException(message, statusCode);
          }
          
        default:
          return const UnknownException();
      }
    }
    
    if (error is SocketException) {
      return const NetworkException();
    }
    
    return const UnknownException();
  }
}

// Interceptor personnalisé pour retry
class RetryInterceptor extends Interceptor {
  final Dio dio;
  final int retries;
  final List<Duration> retryDelays;
  
  RetryInterceptor({
    required this.dio,
    this.retries = 3,
    required this.retryDelays,
  });
  
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (_shouldRetry(err) && err.requestOptions.extra['retry_count'] != retries) {
      final retryCount = (err.requestOptions.extra['retry_count'] ?? 0) + 1;
      err.requestOptions.extra['retry_count'] = retryCount;
      
      await Future.delayed(retryDelays[retryCount - 1]);
      
      try {
        final response = await dio.fetch(err.requestOptions);
        return handler.resolve(response);
      } catch (e) {
        return handler.next(err);
      }
    }
    
    return handler.next(err);
  }
  
  bool _shouldRetry(DioException error) {
    return error.type == DioExceptionType.connectionError ||
           error.type == DioExceptionType.connectionTimeout ||
           error.type == DioExceptionType.receiveTimeout ||
           (error.response?.statusCode == 503); // Service unavailable
  }
}
