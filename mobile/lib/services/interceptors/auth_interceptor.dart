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
    // A verified public, pre-login endpoint (see ApiService.get/post's
    // skipAuth param, used by CategoryApi.getCountries and AuthApi's public
    // methods) sets options.extra['skipAuth'] so it never performs these two
    // FlutterSecureStorage reads. This matters because these reads are
    // awaited here, INSIDE onRequest, before handler.next(options) hands the
    // request to Dio -- so they are NOT covered by Dio's
    // connectTimeout/sendTimeout/receiveTimeout (those only bound the
    // socket-level phases that begin once Dio's adapter takes over). A
    // public endpoint has no reason to wait on auth state at all.
    if (options.extra['skipAuth'] == true) {
      return handler.next(options);
    }

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
