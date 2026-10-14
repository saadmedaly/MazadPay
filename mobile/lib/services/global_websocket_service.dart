import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:mezadpay/services/auth_service.dart';

/// Customer Request #20 (real-time admin -> mobile sync): connection state
/// for [GlobalWebsocketService], exposed so providers/UI can react to it
/// without depending on internal fields (e.g. a Home banner "reconnecting"
/// indicator, if ever added -- not required by this ticket, but the states
/// themselves are).
enum GlobalWsConnectionState { disconnected, connecting, connected, reconnecting }

/// Pure backoff-delay decision for [GlobalWebsocketService]'s reconnect
/// timer: bounded exponential backoff (1s, 2s, 4s, 8s, 16s, capped at 30s),
/// keyed by the zero-based attempt number (0 = first retry after the
/// initial drop). Extracted as a standalone top-level function so it can be
/// tested directly without needing a live/fake WebSocket connection.
int reconnectBackoffSeconds(int attempt, {int baseSeconds = 1, int maxSeconds = 30}) {
  if (attempt < 0) attempt = 0;
  final backoff = baseSeconds * (1 << attempt.clamp(0, 30));
  return backoff > maxSeconds ? maxSeconds : backoff;
}

/// Authenticated, app-wide WebSocket connection for list/content
/// invalidation events (auction.created, auction.updated,
/// auction.status_changed, auction.deleted, faq.updated, banner.updated,
/// category.updated) -- the mobile side of Customer #20's global channel
/// (backend: GET /ws/global, see backend/internal/handlers/global_ws_handler.go).
///
/// Deliberately a SEPARATE class from [WebsocketService] (the existing
/// per-auction /ws/auction/:id client) rather than an extension of it:
/// WebsocketService's single-`_channel`/`_currentAuctionId` contract is
/// already relied upon by auction_provider_api.dart and works correctly for
/// bids -- risking that working path to retrofit a second, independent
/// connection onto the same class was avoided per the "smallest safe
/// change" scope. This class owns its own channel, its own reconnect/
/// backoff state machine, and is a singleton exactly like WebsocketService
/// (one physical global connection for the whole app, regardless of how
/// many providers/screens want its events).
class GlobalWebsocketService {
  static final GlobalWebsocketService _instance = GlobalWebsocketService._internal();
  factory GlobalWebsocketService() => _instance;
  GlobalWebsocketService._internal();

  static const int _maxBackoffSeconds = 30;
  static const int _baseBackoffSeconds = 1;

  WebSocketChannel? _channel;
  StreamSubscription? _subscription;
  Timer? _reconnectTimer;
  final _controller = StreamController<Map<String, dynamic>>.broadcast();
  final _stateController = StreamController<GlobalWsConnectionState>.broadcast();

  GlobalWsConnectionState _state = GlobalWsConnectionState.disconnected;
  int _reconnectAttempt = 0;
  bool _explicitlyDisconnected = false;
  // Guards against overlapping connect() calls (e.g. app resume firing
  // while a reconnect timer is already in flight) creating a second
  // physical connection -- "no duplicate global connections" per spec.
  bool _connecting = false;
  // Gap 2, item 8: the exact token the backend most recently rejected as
  // invalid/expired. While set, _handleDisconnect stops scheduling
  // reconnects -- retrying the SAME rejected token forever would be a
  // pointless infinite loop (bounded by backoff, but never able to
  // succeed). Cleared the moment connect() is asked to try a DIFFERENT
  // token (e.g. after a fresh login writes a new one), so a stale
  // rejection can never block a legitimately renewed session.
  String? _lastRejectedToken;

  Stream<Map<String, dynamic>> get stream => _controller.stream;
  Stream<GlobalWsConnectionState> get stateStream => _stateController.stream;
  GlobalWsConnectionState get state => _state;
  bool get isConnected => _state == GlobalWsConnectionState.connected;

  void _setState(GlobalWsConnectionState s) {
    _state = s;
    _stateController.add(s);
  }

  /// Opens the global connection if not already connected/connecting. Safe
  /// to call repeatedly (e.g. from app-resume and from a screen's initState)
  /// -- a second call while already connected or already connecting is a
  /// no-op, preventing duplicate physical connections.
  Future<void> connect() async {
    if (_connecting || _state == GlobalWsConnectionState.connected) return;
    _explicitlyDisconnected = false;
    _connecting = true;
    _setState(_reconnectAttempt > 0 ? GlobalWsConnectionState.reconnecting : GlobalWsConnectionState.connecting);

    try {
      final wsUrl = dotenv.env['WS_URL'] ?? 'ws://localhost:8082';
      final token = await AuthService().getToken();
      if (token == null || token.isEmpty) {
        // Not logged in yet -- nothing to authenticate the global channel
        // with. Never connect unauthenticated; caller (app resume/login
        // flow) retries connect() once a session exists.
        _connecting = false;
        _setState(GlobalWsConnectionState.disconnected);
        return;
      }
      if (token != _lastRejectedToken) {
        // A different (e.g. freshly renewed) token than the one that was
        // last rejected -- clear the rejection lock so this attempt, and
        // any future disconnect of THIS attempt, are free to reconnect
        // normally again.
        _lastRejectedToken = null;
      } else {
        // Still the exact same token the backend already told us is
        // invalid/expired -- refuse to spin retrying it. Caller must
        // obtain a new token (login/refresh) and call connect() again.
        _connecting = false;
        _setState(GlobalWsConnectionState.disconnected);
        return;
      }
      final url = '$wsUrl/ws/global?token=${Uri.encodeQueryComponent(token)}';

      debugPrint('🌐 Connecting to global WebSocket');

      final connectingWithToken = token;
      _channel = WebSocketChannel.connect(Uri.parse(url));
      await _subscription?.cancel();
      _subscription = _channel!.stream.listen(
        (message) {
          try {
            final data = jsonDecode(message);
            if (data is Map<String, dynamic>) {
              // The backend rejects an invalid/expired JWT with
              // {"error": "..."} then closes (see global_ws_handler.go) --
              // detected here, before onDone/onError fire, so
              // _handleDisconnect can tell "auth rejected this exact token"
              // apart from "a normal network drop" (Gap 2, item 8: a
              // rejected token must not retry in a tight bounded loop
              // forever against a token that can never become valid again
              // without a fresh login).
              if (data['error'] != null) {
                _lastRejectedToken = connectingWithToken;
                debugPrint('🌐 Global WebSocket rejected by server: ${data['error']}');
                return;
              }
              _controller.add(data);
            }
          } catch (e) {
            debugPrint('⚠️ Error decoding global WS message: $e');
          }
        },
        onDone: () {
          debugPrint('🌐 Global WebSocket closed');
          _connecting = false;
          _handleDisconnect();
        },
        onError: (error) {
          debugPrint('❌ Global WebSocket error: $error');
          _connecting = false;
          _handleDisconnect();
        },
      );

      _reconnectAttempt = 0;
      _connecting = false;
      _setState(GlobalWsConnectionState.connected);
    } catch (e) {
      debugPrint('❌ Could not connect to global WebSocket: $e');
      _connecting = false;
      _handleDisconnect();
    }
  }

  void _handleDisconnect() {
    if (_state == GlobalWsConnectionState.disconnected && _explicitlyDisconnected) {
      return;
    }
    _setState(GlobalWsConnectionState.disconnected);
    if (_explicitlyDisconnected) return;
    if (_lastRejectedToken != null) {
      // The server just told us this token is invalid/expired (set by the
      // message handler above, which runs before this onDone/onError
      // fires) -- do not schedule a reconnect against a token that cannot
      // become valid on its own; connect() will refuse to retry it anyway,
      // but skipping _scheduleReconnect entirely also avoids bumping
      // _reconnectAttempt/backoff state for a case backoff can't help.
      return;
    }
    _scheduleReconnect();
  }

  /// Bounded exponential backoff (1s, 2s, 4s, 8s, 16s, capped at 30s) with
  /// jitter -- prevents a reconnect storm (every client retrying in
  /// lockstep the instant the backend comes back) while still recovering
  /// promptly from a single dropped connection. No cap on attempt COUNT:
  /// the app keeps trying indefinitely at the 30s ceiling rather than ever
  /// giving up, since there is no user-facing "give up" state in this spec.
  void _scheduleReconnect() {
    _reconnectTimer?.cancel();
    final attempt = _reconnectAttempt;
    _reconnectAttempt++;
    final backoffSeconds = reconnectBackoffSeconds(attempt, baseSeconds: _baseBackoffSeconds, maxSeconds: _maxBackoffSeconds);
    final jitterMs = Random().nextInt(500);
    _setState(GlobalWsConnectionState.reconnecting);
    _reconnectTimer = Timer(Duration(seconds: backoffSeconds, milliseconds: jitterMs), () {
      if (_explicitlyDisconnected) return;
      connect();
    });
  }

  /// Explicit disconnect (e.g. logout) -- stops any pending reconnect and
  /// will not reconnect until connect() is called again.
  void disconnect() {
    _explicitlyDisconnected = true;
    _reconnectTimer?.cancel();
    _subscription?.cancel();
    _channel?.sink.close();
    _channel = null;
    _reconnectAttempt = 0;
    _lastRejectedToken = null;
    _setState(GlobalWsConnectionState.disconnected);
  }
}
