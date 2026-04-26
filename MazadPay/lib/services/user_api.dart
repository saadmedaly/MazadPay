import 'package:dio/dio.dart';
import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les utilisateurs
class UserApi {
  final ApiService _apiService = ApiService();
  
  /// Récupérer le profil utilisateur
  Future<ApiResponse<Map<String, dynamic>>> getProfile() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/me');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Mettre à jour le profil
  Future<ApiResponse<Map<String, dynamic>>> updateProfile({
    String? fullName,
    String? email,
    String? phone,
    String? city,
  }) async {
    try {
      final data = <String, dynamic>{};
      if (fullName != null) data['full_name'] = fullName;
      if (email != null) data['email'] = email;
      if (phone != null) data['phone'] = phone;
      if (city != null) data['city'] = city;
      
      final response = await _apiService.put<Map<String, dynamic>>(
        '/users/me',
        data: data,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Mettre à jour l'avatar
  Future<ApiResponse<Map<String, dynamic>>> updateAvatar({
    required String avatarUrl,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/users/me/avatar',
        data: {
          'avatar_url': avatarUrl,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Changer la langue
  Future<ApiResponse<Map<String, dynamic>>> updateLanguage({
    required String language, // 'ar', 'fr', 'en'
  }) async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/users/me/language',
        data: {
          'language': language,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Mettre à jour les préférences de notifications
  Future<ApiResponse<Map<String, dynamic>>> updateNotificationPrefs({
    required Map<String, dynamic> preferences,
  }) async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/users/me/notification-prefs',
        data: preferences,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les favoris
  Future<ApiResponse<Map<String, dynamic>>> getFavorites() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/me/favorites');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Ajouter aux favoris
  Future<ApiResponse<Map<String, dynamic>>> addFavorite(String auctionId) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/users/me/favorites/$auctionId',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Supprimer des favoris
  Future<ApiResponse<Map<String, dynamic>>> removeFavorite(String auctionId) async {
    try {
      final response = await _apiService.delete<Map<String, dynamic>>(
        '/users/me/favorites/$auctionId',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer mes enchères
  Future<ApiResponse<Map<String, dynamic>>> getMyAuctions({String? status}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (status != null) queryParams['status'] = status;
      
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/me/auctions',
        queryParameters: queryParams,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer mes enchères placées
  Future<ApiResponse<Map<String, dynamic>>> getMyBids() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/me/bids');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer mes gains
  Future<ApiResponse<Map<String, dynamic>>> getMyWinnings() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/me/winnings');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer le statut KYC
  Future<ApiResponse<Map<String, dynamic>>> getKYCStatus() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/users/kyc');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Soumettre KYC
  Future<ApiResponse<Map<String, dynamic>>> submitKYC({
    required String idCardFront,
    required String idCardBack,
    required String selfie,
  }) async {
    try {
      // Note: Ceci nécessite l'implémentation du multipart upload
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/users/kyc',
        data: FormData.fromMap({
          'id_card_front': await MultipartFile.fromFile(idCardFront),
          'id_card_back': await MultipartFile.fromFile(idCardBack),
          'selfie': await MultipartFile.fromFile(selfie),
        }),
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
