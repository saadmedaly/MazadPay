import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:flutter/foundation.dart';
import '../models/message.dart';
import '../models/conversation.dart';
import 'api_service.dart';

class ChatService extends ChangeNotifier {
  static final ChatService _instance = ChatService._internal();
  factory ChatService() => _instance;
  ChatService._internal();

  WebSocketChannel? _channel;
  Timer? _pingTimer;
  Timer? _reconnectTimer;
  bool _isConnected = false;
  String? _currentUserId;
  String? _authToken;
  int _reconnectAttempts = 0;
  static const int maxReconnectAttempts = 5;
  static const Duration reconnectDelay = Duration(seconds: 3);

  // Streams
  final _messageController = StreamController<Message>.broadcast();
  final _typingController = StreamController<Map<String, dynamic>>.broadcast();
  final _statusController = StreamController<Map<String, dynamic>>.broadcast();
  final _connectionController = StreamController<bool>.broadcast();
  final _onlineUsersController = StreamController<List<String>>.broadcast();

  Stream<Message> get messageStream => _messageController.stream;
  Stream<Map<String, dynamic>> get typingStream => _typingController.stream;
  Stream<Map<String, dynamic>> get statusStream => _statusController.stream;
  Stream<bool> get connectionStream => _connectionController.stream;
  Stream<List<String>> get onlineUsersStream => _onlineUsersController.stream;

  bool get isConnected => _isConnected;
  String? get currentUserId => _currentUserId;

  // Active conversations
  final Set<String> _joinedConversations = {};
  Set<String> get joinedConversations => _joinedConversations;

  Future<void> connect(String userId, String authToken) async {
    _currentUserId = userId;
    _authToken = authToken;
    
    await _connect();
  }

  Future<void> _connect() async {
    if (_isConnected) return;

    try {
      final wsUrl = '${ApiService.wsBaseUrl}/chat/ws?token=$_authToken';
      
      _channel = WebSocketChannel.connect(
        Uri.parse(wsUrl),
      );

      _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
      );

      _isConnected = true;
      _reconnectAttempts = 0;
      _connectionController.add(true);
      _startPingTimer();
      
      notifyListeners();
    } catch (e) {
      debugPrint('WebSocket connection error: $e');
      _scheduleReconnect();
    }
  }

  void _onMessage(dynamic message) {
    try {
      if (message == null || message.toString().trim().isEmpty) return;
      
      final dynamic parsed = jsonDecode(message);
      if (parsed == null) {
        debugPrint('Received null JSON from WebSocket');
        return;
      }
      
      if (parsed is! Map<String, dynamic>) {
        debugPrint('Received non-Map JSON from WebSocket: $parsed');
        return;
      }

      final wsMessage = WebSocketMessage.fromJson(parsed);

      switch (wsMessage.event) {
        case ChatEventTypes.messageNew:
          if (wsMessage.data == null) {
            debugPrint('Received message:new event with null data');
            return;
          }
          final msg = Message.fromJson(wsMessage.data as Map<String, dynamic>);
          _messageController.add(msg);
          break;
          
        case ChatEventTypes.messageRead:
        case ChatEventTypes.messageDelivered:
          if (wsMessage.data != null) {
            _statusController.add(wsMessage.data);
          }
          break;
          
        case ChatEventTypes.typingStart:
        case ChatEventTypes.typingStop:
          if (wsMessage.data != null) {
            _typingController.add(wsMessage.data);
          }
          break;
          
        case ChatEventTypes.userOnline:
        case ChatEventTypes.userOffline:
          if (wsMessage.data != null && wsMessage.data is Map) {
            final userId = wsMessage.data['user_id'];
            if (userId != null) {
              _onlineUsersController.add([userId.toString()]);
            }
          }
          break;
          
        case ChatEventTypes.error:
          debugPrint('Chat error: ${wsMessage.data}');
          break;
      }
    } catch (e) {
      debugPrint('Error parsing WebSocket message: $e');
    }
  }

  void _onError(error) {
    debugPrint('WebSocket error: $error');
    _isConnected = false;
    _connectionController.add(false);
    notifyListeners();
    _scheduleReconnect();
  }

  void _onDone() {
    debugPrint('WebSocket connection closed');
    _isConnected = false;
    _connectionController.add(false);
    notifyListeners();
    _scheduleReconnect();
  }

  void _scheduleReconnect() {
    _pingTimer?.cancel();
    
    if (_reconnectAttempts < maxReconnectAttempts) {
      _reconnectAttempts++;
      debugPrint('Reconnecting attempt $_reconnectAttempts/$maxReconnectAttempts');
      
      _reconnectTimer?.cancel();
      _reconnectTimer = Timer(reconnectDelay * _reconnectAttempts, () {
        _connect();
      });
    }
  }

  void _startPingTimer() {
    _pingTimer?.cancel();
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (_isConnected) {
        _send({'event': 'ping'});
      }
    });
  }

  void _send(Map<String, dynamic> data) {
    if (_isConnected && _channel != null) {
      _channel!.sink.add(jsonEncode(data));
    }
  }

  // Public Methods
  void joinConversation(String conversationId) {
    if (_joinedConversations.contains(conversationId)) return;
    
    _joinedConversations.add(conversationId);
    _send({
      'event': ChatEventTypes.conversationJoin,
      'conversation_id': conversationId,
    });
  }

  void leaveConversation(String conversationId) {
    _joinedConversations.remove(conversationId);
    _send({
      'event': ChatEventTypes.conversationLeave,
      'conversation_id': conversationId,
    });
  }

  void sendTypingStart(String conversationId) {
    _send({
      'event': ChatEventTypes.typingStart,
      'conversation_id': conversationId,
    });
  }

  void sendTypingStop(String conversationId) {
    _send({
      'event': ChatEventTypes.typingStop,
      'conversation_id': conversationId,
    });
  }

  void markAsRead(String conversationId, String? messageId) {
    _send({
      'event': ChatEventTypes.messageRead,
      'conversation_id': conversationId,
      'data': {
        'message_id': messageId,
      },
    });
  }

  void disconnect() {
    _pingTimer?.cancel();
    _reconnectTimer?.cancel();
    _channel?.sink.close();
    _isConnected = false;
    _joinedConversations.clear();
    _connectionController.add(false);
    notifyListeners();
  }

  @override
  void dispose() {
    disconnect();
    _messageController.close();
    _typingController.close();
    _statusController.close();
    _connectionController.close();
    _onlineUsersController.close();
    super.dispose();
  }

  // API Methods (REST)
  Future<List<UserConversation>> getConversations({int limit = 20, int offset = 0}) async {
    final response = await ApiService().get<Map<String, dynamic>>('/conversations?limit=$limit&offset=$offset');
    
    if (response != null && response['success'] == true && response['data'] != null) {
      final List<dynamic> list = response['data'] ?? [];
      return list.map((e) => UserConversation.fromJson(e)).toList();
    }
    return [];
  }

  Future<Conversation?> createConversation(CreateConversationRequest request) async {
    final response = await ApiService().post<Map<String, dynamic>>('/conversations', data: request.toJson());
    
    if (response != null && response['success'] == true && response['data'] != null) {
      return Conversation.fromJson(response['data']);
    }
    return null;
  }

  Future<ConversationResponse?> getConversation(String conversationId) async {
    final response = await ApiService().get<Map<String, dynamic>>('/conversations/$conversationId');
    
    if (response != null && response['success'] == true && response['data'] != null) {
      return ConversationResponse.fromJson(response['data']);
    }
    return null;
  }

  Future<Conversation?> getDirectConversation(String userId) async {
    final response = await ApiService().get<Map<String, dynamic>>('/conversations/direct/$userId');
    
    if (response != null && response['success'] == true && response['data'] != null) {
      return Conversation.fromJson(response['data']);
    }
    return null;
  }

  Future<List<Message>> getMessages(String conversationId, {int limit = 50, int offset = 0}) async {
    final response = await ApiService().get<Map<String, dynamic>>(
      '/conversations/$conversationId/messages?limit=$limit&offset=$offset',
    );
    
    if (response != null && response['success'] == true && response['data'] != null) {
      final List<dynamic> list = response['data'] ?? [];
      return list.map((e) => Message.fromJson(e)).toList();
    }
    return [];
  }

  Future<Message?> sendMessage(String conversationId, SendMessageRequest request) async {
    final response = await ApiService().post<Map<String, dynamic>>(
      '/conversations/$conversationId/messages',
      data: request.toJson(),
    );
    
    if (response != null && response['success'] == true && response['data'] != null) {
      return Message.fromJson(response['data']);
    }
    return null;
  }

  Future<bool> markConversationAsRead(String conversationId, {String? messageId}) async {
    final response = await ApiService().post<Map<String, dynamic>>(
      '/conversations/$conversationId/read',
      data: MarkReadRequest(messageId: messageId).toJson(),
    );
    return response != null && response['success'] == true;
  }

  Future<bool> editMessage(String messageId, String newContent) async {
    final response = await ApiService().put<Map<String, dynamic>>(
      '/messages/$messageId',
      data: {'content': newContent},
    );
    return response != null && response['success'] == true;
  }

  Future<bool> deleteMessage(String messageId) async {
    final response = await ApiService().delete<Map<String, dynamic>>('/messages/$messageId');
    return response != null && response['success'] == true;
  }

  Future<bool> joinConversationAPI(String conversationId) async {
    final response = await ApiService().post<Map<String, dynamic>>('/conversations/$conversationId/join');
    return response != null && response['success'] == true;
  }

  Future<bool> leaveConversationAPI(String conversationId) async {
    final response = await ApiService().post<Map<String, dynamic>>('/conversations/$conversationId/leave');
    return response != null && response['success'] == true;
  }
}
