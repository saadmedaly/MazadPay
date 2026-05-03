import 'package:uuid/uuid.dart';
import 'user.dart';

class Message {
  final String id;
  final String conversationId;
  final String? senderId;
  final String type; // 'text', 'audio', 'video', 'image', 'file', 'system'
  final String? content;
  final String? fileName;
  final String? fileUrl;
  final int? fileSize;
  final int? fileDuration; // En secondes
  final String? mimeType;
  final String? thumbnailUrl;
  final String? replyToId;
  final bool isEdited;
  final bool isDeleted;
  final DateTime? deletedAt;
  final Map<String, dynamic>? metadata;
  final DateTime createdAt;
  final DateTime updatedAt;
  
  // Relations
  final User? sender;
  final Message? replyTo;
  final List<MessageStatus>? status;

  Message({
    required this.id,
    required this.conversationId,
    this.senderId,
    required this.type,
    this.content,
    this.fileName,
    this.fileUrl,
    this.fileSize,
    this.fileDuration,
    this.mimeType,
    this.thumbnailUrl,
    this.replyToId,
    this.isEdited = false,
    this.isDeleted = false,
    this.deletedAt,
    this.metadata,
    required this.createdAt,
    required this.updatedAt,
    this.sender,
    this.replyTo,
    this.status,
  });

  factory Message.fromJson(Map<String, dynamic> json) {
    return Message(
      id: json['id'] ?? '',
      conversationId: json['conversation_id'] ?? '',
      senderId: json['sender_id'],
      type: json['type'] ?? 'text',
      content: json['content'],
      fileName: json['file_name'],
      fileUrl: json['file_url'],
      fileSize: json['file_size'],
      fileDuration: json['file_duration'],
      mimeType: json['mime_type'],
      thumbnailUrl: json['thumbnail_url'],
      replyToId: json['reply_to_id'],
      isEdited: json['is_edited'] ?? false,
      isDeleted: json['is_deleted'] ?? false,
      deletedAt: json['deleted_at'] != null 
          ? DateTime.parse(json['deleted_at']) 
          : null,
      metadata: json['metadata'],
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
      updatedAt: DateTime.parse(json['updated_at'] ?? DateTime.now().toIso8601String()),
      sender: json['sender'] != null ? User.fromJson(json['sender']) : null,
      replyTo: json['reply_to'] != null ? Message.fromJson(json['reply_to']) : null,
      status: (json['status'] as List<dynamic>?)
          ?.map((e) => MessageStatus.fromJson(e))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'conversation_id': conversationId,
      'sender_id': senderId,
      'type': type,
      'content': content,
      'file_name': fileName,
      'file_url': fileUrl,
      'file_size': fileSize,
      'file_duration': fileDuration,
      'mime_type': mimeType,
      'thumbnail_url': thumbnailUrl,
      'reply_to_id': replyToId,
      'is_edited': isEdited,
      'is_deleted': isDeleted,
      'deleted_at': deletedAt?.toIso8601String(),
      'metadata': metadata,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  bool get isMedia => type == 'audio' || type == 'video' || type == 'image' || type == 'file';
  bool get isText => type == 'text';
  bool get isSystem => type == 'system';
  
  String get displayContent {
    if (isDeleted) return 'Message supprimé';
    if (content != null) return content!;
    if (type == 'audio') return '🎤 Message audio';
    if (type == 'video') return '🎥 Vidéo';
    if (type == 'image') return '📷 Image';
    if (type == 'file') return '📎 Fichier';
    return '';
  }
}

class MessageStatus {
  final String id;
  final String messageId;
  final String userId;
  final String status; // 'sent', 'delivered', 'read'
  final DateTime updatedAt;

  MessageStatus({
    required this.id,
    required this.messageId,
    required this.userId,
    required this.status,
    required this.updatedAt,
  });

  factory MessageStatus.fromJson(Map<String, dynamic> json) {
    return MessageStatus(
      id: json['id'] ?? '',
      messageId: json['message_id'] ?? '',
      userId: json['user_id'] ?? '',
      status: json['status'] ?? 'sent',
      updatedAt: DateTime.parse(json['updated_at'] ?? DateTime.now().toIso8601String()),
    );
  }
}

class WebSocketMessage {
  final String event;
  final String? conversationId;
  final dynamic data;
  final DateTime timestamp;

  WebSocketMessage({
    required this.event,
    this.conversationId,
    required this.data,
    required this.timestamp,
  });

  factory WebSocketMessage.fromJson(Map<String, dynamic> json) {
    return WebSocketMessage(
      event: json['event'] ?? '',
      conversationId: json['conversation_id'],
      data: json['data'],
      timestamp: DateTime.parse(json['timestamp'] ?? DateTime.now().toIso8601String()),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'event': event,
      'conversation_id': conversationId,
      'data': data,
      'timestamp': timestamp.toIso8601String(),
    };
  }
}

// WebSocket Event Types
class ChatEventTypes {
  static const String messageNew = 'message:new';
  static const String messageRead = 'message:read';
  static const String messageDelivered = 'message:delivered';
  static const String typingStart = 'typing:start';
  static const String typingStop = 'typing:stop';
  static const String userOnline = 'user:online';
  static const String userOffline = 'user:offline';
  static const String conversationJoin = 'conversation:join';
  static const String conversationLeave = 'conversation:leave';
  static const String error = 'error';
}

// Request Models
class CreateConversationRequest {
  final String type;
  final String? title;
  final List<String> userIds;
  final String? initialMessage;

  CreateConversationRequest({
    required this.type,
    this.title,
    required this.userIds,
    this.initialMessage,
  });

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'title': title,
      'user_ids': userIds,
      'initial_message': initialMessage,
    };
  }
}

class SendMessageRequest {
  final String type;
  final String? content;
  final String? fileName;
  final String? fileUrl;
  final int? fileSize;
  final int? fileDuration;
  final String? mimeType;
  final String? thumbnailUrl;
  final String? replyToId;

  SendMessageRequest({
    required this.type,
    this.content,
    this.fileName,
    this.fileUrl,
    this.fileSize,
    this.fileDuration,
    this.mimeType,
    this.thumbnailUrl,
    this.replyToId,
  });

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'content': content,
      'file_name': fileName,
      'file_url': fileUrl,
      'file_size': fileSize,
      'file_duration': fileDuration,
      'mime_type': mimeType,
      'thumbnail_url': thumbnailUrl,
      'reply_to_id': replyToId,
    };
  }
}

class MarkReadRequest {
  final String? messageId;

  MarkReadRequest({this.messageId});

  Map<String, dynamic> toJson() {
    return {
      'message_id': messageId,
    };
  }
}
