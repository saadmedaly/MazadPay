import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

/// Interceptor pour la gestion centralisée des erreurs API
class ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    String errorMessage = 'Une erreur est survenue';
    
    // Log l'erreur pour le debug
    if (kDebugMode) {
      debugPrint('API Error: ${err.toString()}');
      debugPrint('Response: ${err.response?.data}');
    }
    
    // Personnaliser le message d'erreur selon le type
    switch (err.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        errorMessage = 'Délai de connexion dépassé. Veuillez vérifier votre connexion.';
        break;
      
      case DioExceptionType.badResponse:
        errorMessage = _handleBadResponse(err);
        break;
      
      case DioExceptionType.cancel:
        errorMessage = 'Requête annulée';
        break;
      
      case DioExceptionType.unknown:
        if (err.error.toString().contains('SocketException')) {
          errorMessage = 'Impossible de se connecter au serveur. Veuillez vérifier votre connexion internet.';
        } else {
          errorMessage = 'Erreur de connexion: ${err.message}';
        }
        break;
      
      default:
        errorMessage = 'Erreur inconnue: ${err.message}';
    }
    
    // Créer une nouvelle DioException avec le message personnalisé
    final customError = DioException(
      requestOptions: err.requestOptions,
      response: err.response,
      type: err.type,
      error: errorMessage,
      message: errorMessage,
    );
    
    return handler.next(customError);
  }
  
  String _handleBadResponse(DioException err) {
    final statusCode = err.response?.statusCode;
    final data = err.response?.data;
    
    // Extraire le message d'erreur de la réponse
    String message = 'Erreur serveur';
    
    if (data is Map<String, dynamic>) {
      message = data['error']?['message'] ?? 
                 data['message'] ?? 
                 data['error'] ?? 
                 message;
    }
    
    // Personnaliser selon le code de statut
    switch (statusCode) {
      case 400:
        return message;
      case 401:
        return 'Session expirée. Veuillez vous reconnecter.';
      case 403:
        return 'Accès refusé. Vous n\'avez pas les permissions nécessaires.';
      case 404:
        return 'Ressource non trouvée.';
      case 409:
        return message;
      case 422:
        return message;
      case 429:
        return 'Trop de requêtes. Veuillez réessayer dans quelques instants.';
      case 500:
        return 'Erreur serveur interne. Veuillez réessayer plus tard.';
      case 503:
        return 'Service temporairement indisponible.';
      default:
        return message;
    }
  }
  
  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    // Log la réponse pour le debug
    if (kDebugMode) {
      debugPrint('API Response: ${response.statusCode} - ${response.requestOptions.path}');
    }
    
    return handler.next(response);
  }
  
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    // Log la requête pour le debug
    if (kDebugMode) {
      debugPrint('API Request: ${options.method} ${options.path}');
    }
    
    return handler.next(options);
  }
}
