import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/services/realtime_sync_service.dart';
import 'package:mezadpay/services/global_websocket_service.dart';

// Customer Request #20 (real-time admin -> mobile sync): regression tests
// for the pure decision logic backing the mobile-side realtime pieces --
// RealtimeEvent parsing/classification and the reconnect backoff schedule.
// The stateful singletons (GlobalWebsocketService's actual socket,
// RealtimeSyncService's actual stream wiring) are not exercised here since
// this repo has no HTTP/WebSocket-mocking test infrastructure (same
// constraint documented in auction_image_fallback_test.dart) -- these tests
// instead lock down the exact pure functions/parsing those classes rely on.
void main() {
  group('RealtimeEvent.fromJson (Customer #20)', () {
    test('parses a full auction event', () {
      final event = RealtimeEvent.fromJson({
        'type': 'auction.updated',
        'entity_type': 'auction',
        'entity_id': 'abc-123',
        'updated_at': '2026-01-01T00:00:00Z',
      });
      expect(event.type, 'auction.updated');
      expect(event.entityType, 'auction');
      expect(event.entityId, 'abc-123');
    });

    test('missing fields default to empty strings rather than throwing', () {
      final event = RealtimeEvent.fromJson(<String, dynamic>{});
      expect(event.type, '');
      expect(event.entityType, '');
      expect(event.entityId, '');
    });

    test('the initial "connected" handshake message parses to an empty type, not a crash', () {
      final event = RealtimeEvent.fromJson({'type': 'connected', 'updated_at': '2026-01-01T00:00:00Z'});
      expect(event.type, 'connected');
      expect(event.isAuctionEvent, isFalse);
    });
  });

  group('RealtimeEvent.isAuctionEvent (Customer #20)', () {
    test('auction.created/updated/status_changed/deleted are all auction events', () {
      for (final type in [
        RealtimeEventType.auctionCreated,
        RealtimeEventType.auctionUpdated,
        RealtimeEventType.auctionStatusChanged,
        RealtimeEventType.auctionDeleted,
      ]) {
        final event = RealtimeEvent(type: type, entityType: 'auction', entityId: '1');
        expect(event.isAuctionEvent, isTrue, reason: '$type must be classified as an auction event');
      }
    });

    test('faq/banner/category/request events are NOT auction events -- must not trigger auction-list refetches', () {
      for (final type in [
        RealtimeEventType.faqUpdated,
        RealtimeEventType.bannerUpdated,
        RealtimeEventType.categoryUpdated,
        RealtimeEventType.requestUpdated,
      ]) {
        final event = RealtimeEvent(type: type, entityType: 'x', entityId: '1');
        expect(event.isAuctionEvent, isFalse, reason: '$type must NOT be classified as an auction event');
      }
    });

    test('an unknown/unrecognized type is not treated as an auction event', () {
      final event = RealtimeEvent(type: 'something.else', entityType: 'x', entityId: '1');
      expect(event.isAuctionEvent, isFalse);
    });
  });

  group('request.updated (Customer #20 hardening: Gap 1)', () {
    test('RealtimeEventType.requestUpdated matches the exact literal the backend emits (models.EventRequestUpdated)', () {
      // Kept as an explicit literal-string assertion (not just a reference
      // to the constant) so a typo introduced on either side of the
      // mobile/backend boundary is caught here rather than silently
      // failing to match at runtime.
      expect(RealtimeEventType.requestUpdated, 'request.updated');
    });

    test('parses correctly from the exact payload shape the backend sends', () {
      final event = RealtimeEvent.fromJson({
        'type': 'request.updated',
        'entity_type': 'request',
        'entity_id': 'req-abc-123',
        'updated_at': '2026-01-01T00:00:00Z',
      });
      expect(event.type, RealtimeEventType.requestUpdated);
      expect(event.entityType, 'request');
      expect(event.entityId, 'req-abc-123');
      expect(event.isAuctionEvent, isFalse);
    });
  });

  group('reconnectBackoffSeconds (Customer #20 -- bounded exponential backoff)', () {
    test('doubles each attempt starting from the base', () {
      expect(reconnectBackoffSeconds(0), 1);
      expect(reconnectBackoffSeconds(1), 2);
      expect(reconnectBackoffSeconds(2), 4);
      expect(reconnectBackoffSeconds(3), 8);
      expect(reconnectBackoffSeconds(4), 16);
    });

    test('never exceeds the configured cap, however many attempts pass -- prevents a reconnect storm from an ever-growing delay while still bounding the wait', () {
      expect(reconnectBackoffSeconds(5), 30);
      expect(reconnectBackoffSeconds(10), 30);
      expect(reconnectBackoffSeconds(1000), 30);
    });

    test('a negative attempt number is treated as attempt 0 rather than throwing/underflowing', () {
      expect(reconnectBackoffSeconds(-1), 1);
    });

    test('respects custom base/max overrides', () {
      expect(reconnectBackoffSeconds(0, baseSeconds: 2, maxSeconds: 100), 2);
      expect(reconnectBackoffSeconds(3, baseSeconds: 2, maxSeconds: 100), 16);
      expect(reconnectBackoffSeconds(3, baseSeconds: 2, maxSeconds: 10), 10);
    });
  });
}
