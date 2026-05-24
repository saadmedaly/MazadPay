import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/services/auth_service.dart';
import 'package:mezadpay/widgets/notification_handler.dart' show navigatorKey;
import 'package:mezadpay/pages/login_page.dart';

/// Interceptor pour ajouter automatiquement le JWT token à toutes les requêtes
class AuthInterceptor extends Interceptor {
  final AuthService _authService = AuthService();
  
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    // Ajouter le token JWT si disponible
    final token = await _authService.getToken();
    
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    
    // Ajouter l'ID utilisateur si disponible
    final userId = await _authService.getUserId();
    if (userId != null && userId.isNotEmpty) {
      options.headers['X-User-ID'] = userId;
    }
    
    return handler.next(options);
  }
  
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    // Gérer les erreurs 401 (Unauthorized) - Token expiré
    if (err.response?.statusCode == 401) {
      // Ignorer si l'URL est /auth/login pour éviter une boucle de redirection infinie
      if (err.requestOptions.path.contains('/auth/login')) {
        return handler.next(err);
      }

      // Token expiré ou invalide
      // Pour l'instant, on déconnecte l'utilisateur
      await _authService.logout();
      
      // Rediriger vers la page de login pour arrêter la boucle
      if (navigatorKey.currentState != null) {
        navigatorKey.currentState!.pushAndRemoveUntil(
          MaterialPageRoute(builder: (context) => const LoginPage()),
          (route) => false,
        );
      }
    }
    
    return handler.next(err);
  }
}
