import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les enchères
class AuctionApi {
  final ApiService _apiService = ApiService();
  
  /// Lister toutes les enchères
  Future<ApiResponse<List<dynamic>>> getAuctions({
    int page = 1,
    int limit = 20,
    String? categoryId,
    String? locationId,
    String? status,
    int? minPrice,
    int? maxPrice,
    String? sortBy,
  }) async {
    try {
      final queryParams = <String, dynamic>{
        'page': page,
        'limit': limit,
      };

      if (categoryId != null) queryParams['category_id'] = categoryId;
      if (locationId != null) queryParams['location_id'] = locationId;
      if (status != null) queryParams['status'] = status;
      if (minPrice != null) queryParams['min_price'] = minPrice;
      if (maxPrice != null) queryParams['max_price'] = maxPrice;
      if (sortBy != null) queryParams['sort_by'] = sortBy;

      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions',
        queryParameters: queryParams,
      );
 
      final List<dynamic> auctionList = response?['data'] ?? [];
      return ApiResponse.success(auctionList);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer les détails d'une enchère
  Future<ApiResponse<Map<String, dynamic>>> getAuctionById(String id) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$id',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Alias pour getAuctionById (utilisé par les pages)
  Future<ApiResponse<Map<String, dynamic>>> getAuctionDetails(String id) async {
    return getAuctionById(id);
  }

  /// Récupérer les enchères créées par l'utilisateur connecté
  Future<ApiResponse<List<dynamic>>> getMyAuctions() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/me/auctions',
      );
      
      // La réponse est {"data": [...], "success": true}
      // On extrait la liste directement
      final success = response?['success'] as bool? ?? false;
      final data = response?['data'] as List<dynamic>? ?? [];
      final message = response?['message'] as String?;
      
      return ApiResponse<List<dynamic>>(
        success: success,
        data: data,
        message: message,
      );
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer les enchères gagnées par l'utilisateur
  Future<ApiResponse<Map<String, dynamic>>> getMyWinnings() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/me/winnings',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Créer une enchère (sera validée par l'admin)
  Future<ApiResponse<Map<String, dynamic>>> createAuction({
    required String title,
    required String description,
    required double startingPrice,
    required String category,
    required String subCategory,
    required String location,
    required List<String> images,
    required String phone,
    DateTime? endTime,
  }) async {
    try {
      // Calculer la date de fin par défaut (7 jours à partir de maintenant)
      final endDateTime = endTime ?? DateTime.now().add(const Duration(days: 7));
      
      // Créer l'enchère avec champs snake_case attendus par le backend
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions',
        data: {
          'title_ar': title,
          'title_fr': title,
          'title_en': title,
          'description_ar': description,
          'description_fr': description,
          'description_en': description,
          'start_price': startingPrice,
          'category': category,
          'sub_category': subCategory,
          'location': location,
          'images': images,
          'phone': phone,
          'end_time': '${endDateTime.toIso8601String()}Z',
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Incrémenter le compteur de vues
  Future<ApiResponse<Map<String, dynamic>>> incrementViews(String id) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$id/view',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Placer une enchère
  Future<ApiResponse<Map<String, dynamic>>> placeBid({
    required String auctionId,
    required double amount,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/bids',
        data: {
          'amount': amount,
        },
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Récupérer l'historique des enchères
  Future<ApiResponse<List<dynamic>>> getBidHistory(String auctionId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$auctionId/bids',
      );
      
      // API returns {"data": [...bids...], "success": true}
      final success = response?['success'] as bool? ?? false;
      final data = response?['data'] as List<dynamic>? ?? [];
      
      return ApiResponse<List<dynamic>>(
        success: success,
        data: data,
        message: response?['message'] as String?,
      );
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Obtenir le contact du vendeur
  Future<ApiResponse<Map<String, dynamic>>> getSellerContact(String auctionId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$auctionId/seller-contact',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les catégories
  Future<ApiResponse<Map<String, dynamic>>> getCategories() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/categories');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les locations
  Future<ApiResponse<Map<String, dynamic>>> getLocations({String? countryId}) async {
    try {
      final queryParams = <String, dynamic>{};
      if (countryId != null) queryParams['country_id'] = countryId;
      
      final response = await _apiService.get<Map<String, dynamic>>(
        '/locations',
        queryParameters: queryParams,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
  
  /// Lister les pays
  Future<ApiResponse<Map<String, dynamic>>> getCountries() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>('/countries');
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Mettre à jour une enchère
  Future<ApiResponse<Map<String, dynamic>>> updateAuction({
    required String auctionId,
    String? title,
    String? description,
    double? startingPrice,
    String? category,
    String? subCategory,
    String? location,
    List<String>? images,
    String? phone,
    DateTime? endTime,
  }) async {
    try {
      final data = <String, dynamic>{};
      
      if (title != null) {
        data['title_ar'] = title;
        data['title_fr'] = title;
        data['title_en'] = title;
      }
      if (description != null) {
        data['description_ar'] = description;
        data['description_fr'] = description;
        data['description_en'] = description;
      }
      if (startingPrice != null) data['start_price'] = startingPrice;
      if (category != null) data['category'] = category;
      if (subCategory != null) data['sub_category'] = subCategory;
      if (location != null) data['location'] = location;
      if (images != null) data['images'] = images;
      if (phone != null) data['phone'] = phone;
      if (endTime != null) data['end_time'] = '${endTime.toIso8601String()}Z';

      final response = await _apiService.put<Map<String, dynamic>>(
        '/auctions/$auctionId',
        data: data,
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Supprimer une enchère
  Future<ApiResponse<Map<String, dynamic>>> deleteAuction(String auctionId) async {
    try {
      final response = await _apiService.delete<Map<String, dynamic>>(
        '/auctions/$auctionId',
      );
      
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  // --- NOUVEAUX ENDPOINTS : CYCLE DE VIE ---

  /// Annuler une enchère
  Future<ApiResponse<Map<String, dynamic>>> cancelAuction(String auctionId, {String? reason}) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/cancel',
        data: reason != null ? {'reason': reason} : null,
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Remettre en ligne une enchère
  Future<ApiResponse<Map<String, dynamic>>> relistAuction(String auctionId) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/relist',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Prolonger une enchère
  Future<ApiResponse<Map<String, dynamic>>> extendAuction(String auctionId, {required int hours}) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/extend',
        data: {'hours': hours},
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Ajouter des images supplémentaires après la création
  Future<ApiResponse<Map<String, dynamic>>> addImages(String auctionId, List<String> images) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/images',
        data: {'images': images},
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  // --- NOUVEAUX ENDPOINTS : AUTO-BID ---

  /// Configurer une enchère automatique
  Future<ApiResponse<Map<String, dynamic>>> createAutoBid({
    required String auctionId,
    required double maxAmount,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/auto-bid',
        data: {'max_amount': maxAmount},
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Mettre à jour une enchère automatique
  Future<ApiResponse<Map<String, dynamic>>> updateAutoBid({
    required String auctionId,
    required double maxAmount,
  }) async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/auctions/$auctionId/auto-bid',
        data: {'max_amount': maxAmount},
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Annuler une enchère automatique
  Future<ApiResponse<Map<String, dynamic>>> cancelAutoBid(String auctionId) async {
    try {
      final response = await _apiService.delete<Map<String, dynamic>>(
        '/auctions/$auctionId/auto-bid',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer mes enchères automatiques
  Future<ApiResponse<List<dynamic>>> getMyAutoBids() async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/users/auto-bids',
      );
      final success = response?['success'] as bool? ?? false;
      final data = response?['data'] as List<dynamic>? ?? [];
      return ApiResponse<List<dynamic>>(
        success: success,
        data: data,
        message: response?['message'] as String?,
      );
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  // --- NOUVEAUX ENDPOINTS : BOOST ---

  /// Booster une enchère
  Future<ApiResponse<Map<String, dynamic>>> boostAuction({
    required String auctionId,
    required String boostType,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/boost',
        data: {'boost_type': boostType},
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  // --- NOUVEAUX ENDPOINTS : MODÉRATION ET RÉSULTATS ---

  /// Signaler une enchère
  Future<ApiResponse<Map<String, dynamic>>> reportAuction({
    required String auctionId,
    required String reasonId,
    String? details,
  }) async {
    try {
      final data = <String, dynamic>{'reason_id': reasonId};
      if (details != null && details.isNotEmpty) {
        data['details'] = details;
      }
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/report',
        data: data,
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Obtenir le gagnant d'une enchère
  Future<ApiResponse<Map<String, dynamic>>> getAuctionWinner(String auctionId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$auctionId/winner',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Contacter le vendeur (Initier la conversation)
  Future<ApiResponse<Map<String, dynamic>>> contactSeller(String auctionId) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auctions/$auctionId/contact',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Obtenir le statut d'une offre personnelle
  Future<ApiResponse<Map<String, dynamic>>> getBidStatus(String auctionId) async {
    try {
      final response = await _apiService.get<Map<String, dynamic>>(
        '/auctions/$auctionId/bid-status',
      );
      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
