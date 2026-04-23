// lib/services/websocket_service.dart
// Service WebSocket pour les mises en temps réel

import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart';

import '../core/constants/api_constants.dart';
import '../data/models/notification_model.dart';
import '../domain/repositories/auth_repository.dart';

// Provider pour le service WebSocket
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  final authRepo = ref.watch(authRepositoryProvider);
  return WebSocketService(authRepository: authRepo);
});

// Stream provider pour les événements WebSocket d'une enchère spécifique
final auctionWebSocketProvider = StreamProvider.family<WSEvent, String>((ref, auctionId) {
  final wsService = ref.watch(webSocketServiceProvider);
  return wsService.connectToAuction(auctionId);
});

class WebSocketService {
  final AuthRepository authRepository;
  WebSocketChannel? _channel;
  final Map<String, StreamController<WSEvent>> _auctionControllers = {};
  Timer? _reconnectTimer;
  String? _currentAuctionId;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 5;

  WebSocketService({required this.authRepository});

  /// Connecte au WebSocket d'une enchère spécifique
  Stream<WSEvent> connectToAuction(String auctionId) {
    // Retourne le stream existant si déjà connecté
    if (_auctionControllers.containsKey(auctionId)) {
      return _auctionControllers[auctionId]!.stream;
    }

    // Crée un nouveau controller
    final controller = StreamController<WSEvent>.broadcast(
      onCancel: () => disconnectFromAuction(auctionId),
    );
    _auctionControllers[auctionId] = controller;

    _connect(auctionId);
    return controller.stream;
  }

  /// Établit la connexion WebSocket
  Future<void> _connect(String auctionId) async {
    try {
      _currentAuctionId = auctionId;
      
      // Récupère le token JWT
      final token = await authRepository.getToken();
      if (token == null) {
        _auctionControllers[auctionId]?.addError('Not authenticated');
        return;
      }

      // Construit l'URL WebSocket avec le token
      final wsUrl = '${ApiConstants.wsBaseUrl}/auction/$auctionId?token=$token';

      // Établit la connexion
      _channel = IOWebSocketChannel.connect(
        wsUrl,
        pingInterval: const Duration(seconds: 30),
      );

      _reconnectAttempts = 0;

      // Écoute les messages entrants
      _channel!.stream.listen(
        (message) => _handleMessage(auctionId, message),
        onError: (error) => _handleError(auctionId, error),
        onDone: () => _handleDisconnect(auctionId),
      );
    } catch (e) {
      _handleError(auctionId, e);
    }
  }

  /// Gère les messages entrants
  void _handleMessage(String auctionId, dynamic message) {
    try {
      final data = jsonDecode(message);
      final event = WSEvent.fromJson(data);
      
      _auctionControllers[auctionId]?.add(event);
    } catch (e) {
      print('WebSocket message parse error: $e');
    }
  }

  /// Gère les erreurs
  void _handleError(String auctionId, dynamic error) {
    print('WebSocket error for auction $auctionId: $error');
    _auctionControllers[auctionId]?.addError(error);
    _scheduleReconnect(auctionId);
  }

  /// Gère la déconnexion
  void _handleDisconnect(String auctionId) {
    print('WebSocket disconnected for auction $auctionId');
    _scheduleReconnect(auctionId);
  }

  /// Programme une reconnexion avec backoff exponentiel
  void _scheduleReconnect(String auctionId) {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      _auctionControllers[auctionId]?.addError('Max reconnection attempts reached');
      return;
    }

    _reconnectTimer?.cancel();
    _reconnectAttempts++;

    // Backoff exponentiel: 1s, 2s, 4s, 8s, 16s
    final delay = Duration(seconds: 1 << (_reconnectAttempts - 1));
    
    _reconnectTimer = Timer(delay, () {
      if (_auctionControllers[auctionId]?.hasListener ?? false) {
        _connect(auctionId);
      }
    });
  }

  /// Déconnecte d'une enchère spécifique
  void disconnectFromAuction(String auctionId) {
    _auctionControllers[auctionId]?.close();
    _auctionControllers.remove(auctionId);
    
    if (_currentAuctionId == auctionId) {
      _channel?.sink.close();
      _channel = null;
      _reconnectTimer?.cancel();
    }
  }

  /// Déconnecte toutes les connexions
  void disconnectAll() {
    _reconnectTimer?.cancel();
    _channel?.sink.close();
    _channel = null;
    
    for (final controller in _auctionControllers.values) {
      controller.close();
    }
    _auctionControllers.clear();
  }

  /// Vérifie si connecté à une enchère
  bool isConnectedToAuction(String auctionId) {
    return _auctionControllers.containsKey(auctionId) &&
           _auctionControllers[auctionId]?.hasListener == true;
  }
}

/// Extension pour parser les payloads WebSocket
extension WSEventExtension on WSEvent {
  BidPlacedPayload? get bidPlacedPayload {
    if (type == 'bid_placed' && payload is Map<String, dynamic>) {
      return BidPlacedPayload.fromJson(payload);
    }
    return null;
  }

  TimerTickPayload? get timerTickPayload {
    if (type == 'timer_tick' && payload is Map<String, dynamic>) {
      return TimerTickPayload.fromJson(payload);
    }
    return null;
  }

  AuctionEndedPayload? get auctionEndedPayload {
    if (type == 'auction_ended' && payload is Map<String, dynamic>) {
      return AuctionEndedPayload.fromJson(payload);
    }
    return null;
  }
}
