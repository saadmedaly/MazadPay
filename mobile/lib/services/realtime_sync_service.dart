import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:mezadpay/services/global_websocket_service.dart';

/// Customer Request #20: the global-channel event types this app currently
/// reacts to, mirrored 1:1 from backend/internal/models/websocket_global.go's
/// constants -- kept in exactly one place on each side so the literal
/// strings can't drift silently.
class RealtimeEventType {
  static const auctionCreated = 'auction.created';
  static const auctionUpdated = 'auction.updated';
  static const auctionStatusChanged = 'auction.status_changed';
  static const auctionDeleted = 'auction.deleted';
  static const faqUpdated = 'faq.updated';
  static const bannerUpdated = 'banner.updated';
  static const categoryUpdated = 'category.updated';
  /// Customer #20 hardening round: a PRIVATE, owner-only event (see backend
  /// GlobalHub.BroadcastToUser) -- if it reaches this client at all, it is
  /// necessarily about one of this user's own requests.
  static const requestUpdated = 'request.updated';
}

/// A single parsed global-channel event -- the mobile-side equivalent of
/// backend's models.GlobalWSEvent. entityId is the raw string the backend
/// sent (a UUID for auctions, an int-as-string for FAQ/banner/category).
class RealtimeEvent {
  final String type;
  final String entityType;
  final String entityId;

  RealtimeEvent({required this.type, required this.entityType, required this.entityId});

  factory RealtimeEvent.fromJson(Map<String, dynamic> json) {
    return RealtimeEvent(
      type: json['type']?.toString() ?? '',
      entityType: json['entity_type']?.toString() ?? '',
      entityId: json['entity_id']?.toString() ?? '',
    );
  }

  bool get isAuctionEvent =>
      type == RealtimeEventType.auctionCreated ||
      type == RealtimeEventType.auctionUpdated ||
      type == RealtimeEventType.auctionStatusChanged ||
      type == RealtimeEventType.auctionDeleted;
}

/// RealtimeSyncService owns the app's single [GlobalWebsocketService]
/// connection and re-publishes its raw messages as typed [RealtimeEvent]s,
/// plus a separate "catch up now" signal for app-resume/reconnect (Phase 6:
/// "do NOT attempt to replay missed websocket messages -- use REST refetch
/// as recovery"). Screens/providers that need targeted invalidation listen
/// to [events] (a specific entity changed) or [catchUpSignal] (something may
/// have changed while we were away -- refetch your own current data,
/// regardless of whether a specific event for it arrived).
///
/// A plain singleton (not a Riverpod provider) deliberately, matching the
/// existing WebsocketService/AuthService/CacheService pattern already used
/// throughout this codebase for app-wide singletons that outlive any single
/// widget's provider scope.
class RealtimeSyncService {
  static final RealtimeSyncService _instance = RealtimeSyncService._internal();
  factory RealtimeSyncService() => _instance;
  RealtimeSyncService._internal();

  final _eventsController = StreamController<RealtimeEvent>.broadcast();
  final _catchUpController = StreamController<void>.broadcast();
  StreamSubscription? _wsSubscription;
  StreamSubscription? _wsStateSubscription;
  GlobalWsConnectionState? _previousConnectionState;

  /// A specific, typed invalidation event -- e.g. "auction X was updated".
  Stream<RealtimeEvent> get events => _eventsController.stream;

  /// Fired on app resume and on a successful (re)connection of the global
  /// channel -- listeners should refetch whatever data they currently show,
  /// since any number of events could have been missed while disconnected.
  Stream<void> get catchUpSignal => _catchUpController.stream;

  bool _started = false;

  /// Starts the global connection and begins forwarding its events. Safe to
  /// call multiple times (e.g. once after login, again defensively on
  /// resume) -- GlobalWebsocketService.connect() itself no-ops if already
  /// connected/connecting, so this never creates duplicate listeners: the
  /// subscription below is only ever created once, guarded by [_started].
  Future<void> start() async {
    if (!_started) {
      _started = true;
      _wsSubscription = GlobalWebsocketService().stream.listen((data) {
        try {
          final event = RealtimeEvent.fromJson(data);
          if (event.type.isEmpty) return; // e.g. the "connected" initial_state message
          _eventsController.add(event);
        } catch (e) {
          debugPrint('⚠️ RealtimeSyncService: failed to parse event: $e');
        }
      });

      _wsStateSubscription = GlobalWebsocketService().stateStream.listen((state) {
        // Catch up exactly on the transition INTO connected -- covers both
        // "just reconnected after a drop" and "connected for the first
        // time this session" (the latter matters too: a user could have
        // opened the app while offline, landed on cached/stale data, then
        // come online for the first time).
        if (state == GlobalWsConnectionState.connected &&
            _previousConnectionState != GlobalWsConnectionState.connected) {
          _catchUpController.add(null);
        }
        _previousConnectionState = state;
      });
    }
    await GlobalWebsocketService().connect();
  }

  /// Called by the app-root lifecycle observer on AppLifecycleState.resumed
  /// (Phase 6: "the app must trigger targeted refresh/invalidation ...
  /// without requiring user interaction"). Reconnects the global channel if
  /// it dropped while backgrounded, AND fires catch-up immediately -- do not
  /// wait for the reconnect round-trip before refreshing visible screens,
  /// since a brief background/foreground cycle may not have dropped the
  /// socket at all, yet admin edits could still have happened meanwhile.
  Future<void> onAppResumed() async {
    _catchUpController.add(null);
    await GlobalWebsocketService().connect();
  }

  void stop() {
    _wsSubscription?.cancel();
    _wsStateSubscription?.cancel();
    _started = false;
    _previousConnectionState = null;
    GlobalWebsocketService().disconnect();
  }

  /// Customer #20 hardening (Gap 2 -- user switch safety): stop() THEN
  /// start(), guaranteeing any existing connection/reconnect timer (e.g.
  /// belonging to a previous user, or left over from a login flow retried)
  /// is torn down before the new identity's connection opens. Called right
  /// after a successful login/register -- see AuthApi.login(). Never reuses
  /// a stale token: start() -> GlobalWebsocketService().connect() always
  /// reads AuthService().getToken() fresh at call time, well after this
  /// restart runs, so it is always the just-saved token for the current
  /// user, never a previous user's.
  Future<void> restart() async {
    stop();
    await start();
  }
}
