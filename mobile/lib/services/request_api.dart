import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';

/// Service API pour les demandes (auction requests, banner requests)
class RequestApi {
  final ApiService _apiService = ApiService();

  /// Créer une demande d'enchère (auction_requests workflow).
  ///
  /// [status] optionnel : "draft" pour enregistrer sans soumettre, "pending"
  /// (ou omis) pour soumettre immédiatement. description_ar est obligatoire
  /// côté backend (10-5000 caractères).
  Future<ApiResponse<Map<String, dynamic>>> createAuctionRequest({
    required int categoryId,
    required int locationId,
    required String titleAr,
    String? titleFr,
    String? titleEn,
    required String descriptionAr,
    String? descriptionFr,
    String? descriptionEn,
    required double startPrice,
    double? minIncrement,
    double? insuranceAmount,
    double? reservePrice,
    double? buyNowPrice,
    DateTime? startDate,
    DateTime? endDate,
    Map<String, dynamic>? images,
    int quantity = 1,
    String? status,
  }) async {
    try {
      final data = <String, dynamic>{
        'category_id': categoryId,
        'location_id': locationId,
        'title_ar': titleAr,
        'title_fr': titleFr ?? titleAr,
        'title_en': titleEn ?? titleAr,
        'description_ar': descriptionAr,
        'description_fr': descriptionFr ?? descriptionAr,
        'description_en': descriptionEn ?? descriptionAr,
        'start_price': startPrice,
        'min_increment': minIncrement ?? 0,
        'insurance_amount': insuranceAmount ?? 0,
        'reserve_price': reservePrice,
        'buy_now_price': buyNowPrice,
        'start_date': (startDate ?? DateTime.now()).toUtc().toIso8601String(),
        'end_date': (endDate ?? DateTime.now().add(const Duration(days: 7)))
            .toUtc()
            .toIso8601String(),
        'images': images ?? {},
        'quantity': quantity,
      };
      if (status != null) data['status'] = status;

      final response = await _apiService.post<Map<String, dynamic>>(
        '/requests/auctions',
        data: data,
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Modifier et (re)soumettre une demande d'enchère existante — utilisé
  /// pour le flux "brouillon" / "rejeté -> corriger -> renvoyer". Ne
  /// fonctionne que si l'appelant est propriétaire ET que la demande est
  /// actuellement "draft" ou "rejected" (contrôlé côté backend).
  Future<ApiResponse<Map<String, dynamic>>> updateAuctionRequest({
    required String id,
    int? categoryId,
    int? locationId,
    String? titleAr,
    String? titleFr,
    String? titleEn,
    String? descriptionAr,
    String? descriptionFr,
    String? descriptionEn,
    double? startPrice,
    double? minIncrement,
    double? insuranceAmount,
    double? reservePrice,
    double? buyNowPrice,
    DateTime? startDate,
    DateTime? endDate,
    Map<String, dynamic>? images,
    int? quantity,
    String? status,
  }) async {
    try {
      final data = <String, dynamic>{};
      if (categoryId != null) data['category_id'] = categoryId;
      if (locationId != null) data['location_id'] = locationId;
      if (titleAr != null) data['title_ar'] = titleAr;
      if (titleFr != null) data['title_fr'] = titleFr;
      if (titleEn != null) data['title_en'] = titleEn;
      if (descriptionAr != null) data['description_ar'] = descriptionAr;
      if (descriptionFr != null) data['description_fr'] = descriptionFr;
      if (descriptionEn != null) data['description_en'] = descriptionEn;
      if (startPrice != null) data['start_price'] = startPrice;
      if (minIncrement != null) data['min_increment'] = minIncrement;
      if (insuranceAmount != null) data['insurance_amount'] = insuranceAmount;
      if (reservePrice != null) data['reserve_price'] = reservePrice;
      if (buyNowPrice != null) data['buy_now_price'] = buyNowPrice;
      if (startDate != null)
        data['start_date'] = startDate.toUtc().toIso8601String();
      if (endDate != null) data['end_date'] = endDate.toUtc().toIso8601String();
      if (images != null) data['images'] = images;
      if (quantity != null) data['quantity'] = quantity;
      if (status != null) data['status'] = status;

      final response = await _apiService.put<Map<String, dynamic>>(
        '/requests/auctions/$id',
        data: data,
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Créer une demande de bannière
  Future<ApiResponse<Map<String, dynamic>>> createBannerRequest({
    required String titleAr,
    required String titleFr,
    required String titleEn,
    required String descriptionAr,
    required String descriptionFr,
    required String descriptionEn,
    required String imageUrl,
    required String linkUrl,
    required DateTime startsAt,
    required DateTime endsAt,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/requests/banners',
        data: {
          'title_ar': titleAr,
          'title_fr': titleFr,
          'title_en': titleEn,
          'description_ar': descriptionAr,
          'description_fr': descriptionFr,
          'description_en': descriptionEn,
          'image_url': imageUrl,
          'link_url': linkUrl,
          'starts_at': startsAt.toIso8601String(),
          'ends_at': endsAt.toIso8601String(),
        },
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer mes demandes d'enchères
  Future<ApiResponse<Map<String, dynamic>>> getMyAuctionRequests({
    String? status,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (status != null) queryParams['status'] = status;

      final response = await _apiService.get<Map<String, dynamic>>(
        '/requests/auctions/my',
        queryParameters: queryParams,
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }

  /// Récupérer mes demandes de bannières
  Future<ApiResponse<Map<String, dynamic>>> getMyBannerRequests({
    String? status,
  }) async {
    try {
      final queryParams = <String, dynamic>{};
      if (status != null) queryParams['status'] = status;

      final response = await _apiService.get<Map<String, dynamic>>(
        '/requests/banners/my',
        queryParameters: queryParams,
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      return ApiResponse.error(e.toString());
    }
  }
}
