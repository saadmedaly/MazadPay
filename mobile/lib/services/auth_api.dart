import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/services/auth_service.dart';

/// Service API pour l'authentification
class AuthApi {
  final ApiService _apiService = ApiService();
  final AuthService _authService = AuthService();
  
  /// Connexion utilisateur
  Future<ApiResponse<Map<String, dynamic>>> login({
    required String phone,
    required String pin,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/login',
        data: {
          'phone': phone,
          'pin': pin,
        },
      );
      
      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(response);
      
      // Si succès, sauvegarder le token
      if (apiResponse.success && apiResponse.data != null) {
        final token = apiResponse.data!['token'];
        final user = apiResponse.data!['user'];
        await _authService.saveToken(token);
        if (user != null && user['id'] != null) {
          await _authService.saveUserId(user['id'].toString());
        }
      }
      
      return apiResponse;
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Inscription utilisateur
  Future<ApiResponse<Map<String, dynamic>>> register({
    required String phone,
    required String pin,
    required String fullName,
    String? email,
    String? city,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/register',
        data: {
          'phone': phone,
          'pin': pin,
          'full_name': fullName,
          if (email != null) 'email': email,
          if (city != null) 'city': city,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Envoyer OTP
  Future<ApiResponse<Map<String, dynamic>>> sendOTP({
    required String phone,
    required String purpose, // 'register' ou 'reset_password'
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/otp/send',
        data: {
          'phone': phone,
          'purpose': purpose,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Vérifier OTP
  Future<ApiResponse<Map<String, dynamic>>> verifyOTP({
    required String phone,
    required String code,
    required String purpose,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/otp/verify',
        data: {
          'phone': phone,
          'code': code,
          'purpose': purpose,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Réinitialiser le mot de passe
  Future<ApiResponse<Map<String, dynamic>>> resetPassword({
    required String phone,
    required String newPin,
    required String otpCode,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/reset-password',
        data: {
          'phone': phone,
          'new_pin': newPin,
          'otp_code': otpCode,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Changer le mot de passe
  Future<ApiResponse<Map<String, dynamic>>> changePassword({
    required String oldPin,
    required String newPin,
  }) async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/auth/change-password',
        data: {
          'old_pin': oldPin,
          'new_pin': newPin,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Déconnexion
  Future<void> logout() async {
    try {
      await _apiService.post('/auth/logout');
    } catch (e) {
      // Ignorer l'erreur lors de la déconnexion
    } finally {
      await _authService.logout();
    }
  }
}
