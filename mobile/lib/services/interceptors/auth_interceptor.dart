import 'package:dio/dio.dart';
import 'package:mezadpay/services/auth_service.dart';

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
      // Token expiré ou invalide
      // Optionnel: Essayer de rafraîchir le token
      // Pour l'instant, on déconnecte l'utilisateur
      await _authService.logout();
      
      // Note: Dans une implémentation complète, vous pourriez:
      // 1. Essayer de rafraîchir le token avec refresh_token
      // 2. Si réussi, réessayer la requête originale
      // 3. Si échec, déconnecter et rediriger vers login
    }
    
    return handler.next(err);
  }
}
