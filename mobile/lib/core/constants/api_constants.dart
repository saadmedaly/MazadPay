// lib/core/constants/api_constants.dart
// Configuration API pour MazadPay Mobile

class ApiConstants {
  // Base URL - changer selon l'environnement
  static const String baseUrl = 'http://localhost:8082/v1/api';
  static const String wsBaseUrl = 'ws://localhost:8082/ws';
  
  // Timeout
  static const int connectTimeout = 10000; // 10 secondes
  static const int receiveTimeout = 10000;
  
  // Endpoints
  static const String login = '/auth/login';
  static const String register = '/auth/register';
  static const String logout = '/auth/logout';
  static const String refreshToken = '/auth/refresh';
  static const String changePassword = '/auth/change-password';
  static const String sendOtp = '/auth/otp/send';
  static const String verifyOtp = '/auth/otp/verify';
  static const String resetPassword = '/auth/reset-password';
  
  static const String me = '/users/me';
  static const String updateProfile = '/users/me';
  static const String auctions = '/auctions';
  static const String auctionDetail = '/auctions/';
  static const String placeBid = '/auctions/';
  static const String favorites = '/users/favorites';
  static const String notifications = '/notifications';
  static const String pushTokens = '/notifications/push-tokens';
  
  // WebSocket
  static String wsAuction(String auctionId) => '/ws/auction/$auctionId';
}

// Messages d'erreur multilingues
class ErrorMessages {
  static const Map<String, Map<String, String>> messages = {
    'ar': {
      'network_error': 'خطأ في الاتصال بالشبكة',
      'server_error': 'خطأ في الخادم',
      'timeout': 'انتهت مهلة الطلب',
      'unauthorized': 'جلسة منتهية، يرجى تسجيل الدخول مرة أخرى',
      'not_found': 'الصفحة غير موجودة',
      'validation_error': 'خطأ في البيانات المدخلة',
      'unknown_error': 'حدث خطأ غير متوقع',
    },
    'fr': {
      'network_error': 'Erreur de connexion réseau',
      'server_error': 'Erreur serveur',
      'timeout': 'Délai d\'attente dépassé',
      'unauthorized': 'Session expirée, veuillez vous reconnecter',
      'not_found': 'Page non trouvée',
      'validation_error': 'Erreur de validation des données',
      'unknown_error': 'Une erreur inattendue s\'est produite',
    },
    'en': {
      'network_error': 'Network connection error',
      'server_error': 'Server error',
      'timeout': 'Request timeout',
      'unauthorized': 'Session expired, please login again',
      'not_found': 'Page not found',
      'validation_error': 'Data validation error',
      'unknown_error': 'An unexpected error occurred',
    },
  };
  
  static String get(String key, String lang) {
    return messages[lang]?[key] ?? messages['ar']![key] ?? key;
  }
}
