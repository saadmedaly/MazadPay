// lib/data/models/notification_model.dart
// Modèle Notification Push et In-App

import 'package:freezed_annotation/freezed_annotation.dart';

part 'notification_model.freezed.dart';
part 'notification_model.g.dart';

@freezed
class NotificationModel with _$NotificationModel {
  const factory NotificationModel({
    required String id,
    required String userId,
    required String title,
    required String body,
    String? type, // bid_placed, bid_outbid, auction_won, auction_ended, etc.
    String? imageUrl,
    @Default(false) bool isRead,
    String? actionUrl, // Deep link
    Map<String, dynamic>? data,
    String? createdAt,
  }) = _NotificationModel;
  
  factory NotificationModel.fromJson(Map<String, dynamic> json) => 
      _$NotificationModelFromJson(json);
}

@freezed
class PushTokenRequest with _$PushTokenRequest {
  const factory PushTokenRequest({
    required String fcmToken,
    String? deviceId,
    @Default('android') String platform, // android, ios, web
  }) = _PushTokenRequest;
  
  factory PushTokenRequest.fromJson(Map<String, dynamic> json) => 
      _$PushTokenRequestFromJson(json);
}

// Événements WebSocket
@freezed
class WSEvent with _$WSEvent {
  const factory WSEvent({
    required String type, // bid_placed, timer_tick, auction_ended, auction_won, initial_state
    required dynamic payload,
  }) = _WSEvent;
  
  factory WSEvent.fromJson(Map<String, dynamic> json) => 
      _$WSEventFromJson(json);
}

@freezed
class BidPlacedPayload with _$BidPlacedPayload {
  const factory BidPlacedPayload({
    required String auctionId,
    required double newPrice,
    required String bidderPhone,
    required int bidCount,
    required int secondsLeft,
  }) = _BidPlacedPayload;
  
  factory BidPlacedPayload.fromJson(Map<String, dynamic> json) => 
      _$BidPlacedPayloadFromJson(json);
}

@freezed
class TimerTickPayload with _$TimerTickPayload {
  const factory TimerTickPayload({
    required String auctionId,
    required int secondsLeft,
  }) = _TimerTickPayload;
  
  factory TimerTickPayload.fromJson(Map<String, dynamic> json) => 
      _$TimerTickPayloadFromJson(json);
}

@freezed
class AuctionEndedPayload with _$AuctionEndedPayload {
  const factory AuctionEndedPayload({
    required String auctionId,
    required double finalPrice,
    String? winnerPhone,
  }) = _AuctionEndedPayload;
  
  factory AuctionEndedPayload.fromJson(Map<String, dynamic> json) => 
      _$AuctionEndedPayloadFromJson(json);
}
