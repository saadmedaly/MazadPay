import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:mezadpay/services/auth_service.dart';

class WebsocketService {
  static final WebsocketService _instance = WebsocketService._internal();
  factory WebsocketService() => _instance;
  WebsocketService._internal();

  WebSocketChannel? _channel;
  final _controller = StreamController<Map<String, dynamic>>.broadcast();
  
  Stream<Map<String, dynamic>> get stream => _controller.stream;
  
  bool _isConnected = false;
  String? _currentAuctionId;

  // Live Staging root cause (client feedback, round 3, item 3): the backend
  // WS handler (ws_handler.go HandleAuction) requires ?token=<JWT> as a query
  // param and immediately closes with "Missing JWT token" otherwise -- this
  // method never appended one, so every WebSocket connection failed
  // silently (caught by onError, only logged to console). Fetches the
  // current JWT the same way AuthInterceptor does (AuthService().getToken())
  // and URL-encodes it into the connection URL; never logged.
  Future<void> connect(String auctionId) async {
    if (_isConnected && _currentAuctionId == auctionId) return;

    _currentAuctionId = auctionId;
    _disconnect();

    final wsUrl = dotenv.env['WS_URL'] ?? 'ws://localhost:8082';
    final token = await AuthService().getToken();
    final url = token != null && token.isNotEmpty
        ? '$wsUrl/ws/auction/$auctionId?token=${Uri.encodeQueryComponent(token)}'
        : '$wsUrl/ws/auction/$auctionId';

    debugPrint('🔌 Connecting to WebSocket for auction $auctionId');

    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));
      _isConnected = true;

      _channel!.stream.listen(
        (message) {
          try {
            final data = jsonDecode(message);
            _controller.add(data);
          } catch (e) {
            debugPrint('⚠️ Error decoding WS message: $e');
          }
        },
        onDone: () {
          debugPrint('🔌 WebSocket connection closed');
          _isConnected = false;
        },
        onError: (error) {
          debugPrint('❌ WebSocket error: $error');
          _isConnected = false;
        },
      );
    } catch (e) {
      debugPrint('❌ Could not connect to WebSocket: $e');
      _isConnected = false;
    }
  }

  void _disconnect() {
    _channel?.sink.close();
    _isConnected = false;
  }

  void close() {
    _disconnect();
    _controller.close();
  }
}
