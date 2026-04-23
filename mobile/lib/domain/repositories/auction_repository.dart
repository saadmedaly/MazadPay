// lib/domain/repositories/auction_repository.dart
// Repository pour les enchères

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/errors/exceptions.dart';
import '../../core/constants/api_constants.dart';
import '../../data/datasources/remote/api_client.dart';
import '../../data/models/auction_model.dart';
import 'auth_repository.dart';

final auctionRepositoryProvider = Provider<AuctionRepository>((ref) {
  final dio = ref.watch(apiClientProvider);
  return AuctionRepositoryImpl(dio: dio);
});

abstract class AuctionRepository {
  Future<AuctionListResponse> getAuctions({
    int page = 1,
    int limit = 20,
    String? status,
    String? categoryId,
    String? search,
  });
  
  Future<AuctionModel> getAuctionDetail(String id);
  
  Future<BidModel> placeBid(String auctionId, PlaceBidRequest request);
  
  Future<List<AuctionModel>> getMyAuctions();
  
  Future<List<AuctionModel>> getFeaturedAuctions();
  
  Future<List<AuctionModel>> getEndingSoonAuctions();
  
  Future<List<AuctionModel>> getNearbyAuctions(double lat, double lng);
  
  Future<void> addToFavorites(String auctionId);
  
  Future<void> removeFromFavorites(String auctionId);
  
  Future<List<AuctionModel>> getFavorites();
  
  Future<List<BidModel>> getAuctionBids(String auctionId);
}

class AuctionRepositoryImpl implements AuctionRepository {
  final Dio dio;

  AuctionRepositoryImpl({required this.dio});

  @override
  Future<AuctionListResponse> getAuctions({
    int page = 1,
    int limit = 20,
    String? status,
    String? categoryId,
    String? search,
  }) async {
    try {
      final queryParams = <String, dynamic>{
        'page': page,
        'limit': limit,
      };
      
      if (status != null) queryParams['status'] = status;
      if (categoryId != null) queryParams['category_id'] = categoryId;
      if (search != null && search.isNotEmpty) queryParams['search'] = search;

      final response = await dio.get(
        ApiConstants.auctions,
        queryParameters: queryParams,
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        return AuctionListResponse.fromJson({
          'data': response.data['data'],
          'meta': response.data['meta'],
        });
      }

      throw ServerException(response.data['message'] ?? 'Failed to load auctions');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<AuctionModel> getAuctionDetail(String id) async {
    try {
      final response = await dio.get('${ApiConstants.auctionDetail}$id');

      if (response.statusCode == 200 && response.data['success'] == true) {
        return AuctionModel.fromJson(response.data['data']);
      }

      throw NotFoundException('Auction not found');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<BidModel> placeBid(String auctionId, PlaceBidRequest request) async {
    try {
      final response = await dio.post(
        '${ApiConstants.placeBid}$auctionId/bid',
        data: request.toJson(),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        return BidModel.fromJson(response.data['data']);
      }

      throw ServerException(response.data['message'] ?? 'Failed to place bid');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<List<AuctionModel>> getMyAuctions() async {
    try {
      final response = await dio.get('${ApiConstants.auctions}/my-auctions');

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => AuctionModel.fromJson(e)).toList();
      }

      throw ServerException('Failed to load my auctions');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<List<AuctionModel>> getFeaturedAuctions() async {
    try {
      final response = await dio.get(
        ApiConstants.auctions,
        queryParameters: {'featured': true, 'limit': 10},
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => AuctionModel.fromJson(e)).toList();
      }

      throw ServerException('Failed to load featured auctions');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<List<AuctionModel>> getEndingSoonAuctions() async {
    try {
      final response = await dio.get(
        ApiConstants.auctions,
        queryParameters: {'ending_soon': true, 'limit': 10},
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => AuctionModel.fromJson(e)).toList();
      }

      throw ServerException('Failed to load ending soon auctions');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<List<AuctionModel>> getNearbyAuctions(double lat, double lng) async {
    try {
      final response = await dio.get(
        ApiConstants.auctions,
        queryParameters: {
          'lat': lat,
          'lng': lng,
          'nearby': true,
        },
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => AuctionModel.fromJson(e)).toList();
      }

      throw ServerException('Failed to load nearby auctions');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> addToFavorites(String auctionId) async {
    try {
      await dio.post('${ApiConstants.favorites}/$auctionId');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> removeFromFavorites(String auctionId) async {
    try {
      await dio.delete('${ApiConstants.favorites}/$auctionId');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<List<AuctionModel>> getFavorites() async {
    try {
      final response = await dio.get(ApiConstants.favorites);

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => AuctionModel.fromJson(e)).toList();
      }

      throw ServerException('Failed to load favorites');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<List<BidModel>> getAuctionBids(String auctionId) async {
    try {
      final response = await dio.get(
        '${ApiConstants.auctionDetail}$auctionId/bids',
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => BidModel.fromJson(e)).toList();
      }

      throw ServerException('Failed to load bids');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }
}
