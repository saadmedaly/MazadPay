// lib/data/models/auction_model.dart
// Modèle Enchère (Auction) avec JSON serialization

import 'package:freezed_annotation/freezed_annotation.dart';

part 'auction_model.freezed.dart';
part 'auction_model.g.dart';

@freezed
class AuctionModel with _$AuctionModel {
  const factory AuctionModel({
    required String id,
    required String title,
    String? description,
    String? shortDescription,
    required double startPrice,
    required double currentPrice,
    double? buyNowPrice,
    double? reservePrice,
    String? currency,
    required String status, // draft, active, paused, ended, cancelled
    String? categoryId,
    String? categoryName,
    String? subCategoryId,
    String? subCategoryName,
    String? locationId,
    String? locationName,
    String? sellerId,
    String? sellerName,
    String? mainImageUrl,
    List<String>? imageUrls,
    String? videoUrl,
    int? quantity,
    String? condition, // new, used, refurbished
    int? viewCount,
    int? bidCount,
    String? winnerId,
    String? endedAt,
    String? startedAt,
    String? endTime,
    String? createdAt,
    String? updatedAt,
    bool? isFavorite,
    bool? isUserBidding,
    double? userLastBid,
    // Pour les enchères en temps réel
    @Default(0) int secondsLeft,
    @Default(false) bool isLive,
  }) = _AuctionModel;
  
  factory AuctionModel.fromJson(Map<String, dynamic> json) => 
      _$AuctionModelFromJson(json);
}

// Modèle pour placer une mise
@freezed
class PlaceBidRequest with _$PlaceBidRequest {
  const factory PlaceBidRequest({
    required double amount,
    String? pin, // Pour vérification supplémentaire sur mobile
  }) = _PlaceBidRequest;
  
  factory PlaceBidRequest.fromJson(Map<String, dynamic> json) => 
      _$PlaceBidRequestFromJson(json);
}

@freezed
class BidModel with _$BidModel {
  const factory BidModel({
    required String id,
    required String auctionId,
    required String userId,
    required double amount,
    String? bidderName,
    String? bidderPhone,
    required String createdAt,
    bool? isAutoBid,
    double? maxAutoBid,
  }) = _BidModel;
  
  factory BidModel.fromJson(Map<String, dynamic> json) => 
      _$BidModelFromJson(json);
}

// Réponse paginée pour les enchères
@freezed
class AuctionListResponse with _$AuctionListResponse {
  const factory AuctionListResponse({
    required List<AuctionModel> data,
    required PaginationMeta meta,
  }) = _AuctionListResponse;
  
  factory AuctionListResponse.fromJson(Map<String, dynamic> json) => 
      _$AuctionListResponseFromJson(json);
}

@freezed
class PaginationMeta with _$PaginationMeta {
  const factory PaginationMeta({
    required int total,
    required int page,
    required int limit,
    required int totalPages,
  }) = _PaginationMeta;
  
  factory PaginationMeta.fromJson(Map<String, dynamic> json) => 
      _$PaginationMetaFromJson(json);
}
