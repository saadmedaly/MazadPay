import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class WebsocketService {
  static final WebsocketService _instance = WebsocketService._internal();
  factory WebsocketService() => _instance;
  WebsocketService._internal();

  WebSocketChannel? _channel;
  final _controller = StreamController<Map<String, dynamic>>.broadcast();
  
  Stream<Map<String, dynamic>> get stream => _controller.stream;
  
  bool _isConnected = false;
  String? _currentAuctionId;

  void connect(String auctionId) {
    if (_isConnected && _currentAuctionId == auctionId) return;
    
    _currentAuctionId = auctionId;
    _disconnect();

    final wsUrl = dotenv.env['WS_URL'] ?? 'ws://localhost:8082';
    final url = '$wsUrl/ws/auction/$auctionId';
    
    debugPrint('🔌 Connecting to WebSocket: $url');
    
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
