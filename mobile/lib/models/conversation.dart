import 'package:uuid/uuid.dart';
import 'user.dart';

class Conversation {
  final String id;
  final String type; // 'direct', 'group', 'support'
  final String? title;
  final String? createdBy;
  final DateTime createdAt;
  final DateTime updatedAt;
  final DateTime? lastMessageAt;
  final String? lastMessagePreview;
  final String? lastMessageSenderId;
  final bool isActive;
  final Map<String, dynamic>? metadata;

  Conversation({
    required this.id,
    required this.type,
    this.title,
    this.createdBy,
    required this.createdAt,
    required this.updatedAt,
    this.lastMessageAt,
    this.lastMessagePreview,
    this.lastMessageSenderId,
    this.isActive = true,
    this.metadata,
  });

  factory Conversation.fromJson(Map<String, dynamic> json) {
    return Conversation(
      id: json['id'] ?? '',
      type: json['type'] ?? 'direct',
      title: json['title'],
      createdBy: json['created_by'],
      createdAt: DateTime.parse(json['created_at'] ?? DateTime.now().toIso8601String()),
      updatedAt: DateTime.parse(json['updated_at'] ?? DateTime.now().toIso8601String()),
      lastMessageAt: json['last_message_at'] != null 
          ? DateTime.parse(json['last_message_at']) 
          : null,
      lastMessagePreview: json['last_message_preview'],
      lastMessageSenderId: json['last_message_sender_id'],
      isActive: json['is_active'] ?? true,
      metadata: json['metadata'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type,
      'title': title,
      'created_by': createdBy,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
      'last_message_at': lastMessageAt?.toIso8601String(),
      'last_message_preview': lastMessagePreview,
      'last_message_sender_id': lastMessageSenderId,
      'is_active': isActive,
      'metadata': metadata,
    };
  }
}

class ConversationParticipant {
  final String id;
  final String conversationId;
  final String userId;
  final String role; // 'owner', 'admin', 'member'
  final DateTime joinedAt;
  final DateTime? lastReadAt;
  final String? lastReadMessageId;
  final bool isMuted;
  final int unreadCount;
  final User? user;

  ConversationParticipant({
    required this.id,
    required this.conversationId,
    required this.userId,
    this.role = 'member',
    required this.joinedAt,
    this.lastReadAt,
    this.lastReadMessageId,
    this.isMuted = false,
    this.unreadCount = 0,
    this.user,
  });

  factory ConversationParticipant.fromJson(Map<String, dynamic> json) {
    return ConversationParticipant(
      id: json['id'] ?? '',
      conversationId: json['conversation_id'] ?? '',
      userId: json['user_id'] ?? '',
      role: json['role'] ?? 'member',
      joinedAt: DateTime.parse(json['joined_at'] ?? DateTime.now().toIso8601String()),
      lastReadAt: json['last_read_at'] != null 
          ? DateTime.parse(json['last_read_at']) 
          : null,
      lastReadMessageId: json['last_read_message_id'],
      isMuted: json['is_muted'] ?? false,
      unreadCount: json['unread_count'] ?? 0,
      user: json['user'] != null ? User.fromJson(json['user']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'conversation_id': conversationId,
      'user_id': userId,
      'role': role,
      'joined_at': joinedAt.toIso8601String(),
      'last_read_at': lastReadAt?.toIso8601String(),
      'last_read_message_id': lastReadMessageId,
      'is_muted': isMuted,
      'unread_count': unreadCount,
      'user': user?.toJson(),
    };
  }
}

class UserConversation {
  final String userId;
  final String conversationId;
  final String type;
  final String? title;
  final DateTime? lastMessageAt;
  final String? lastMessagePreview;
  final String? lastMessageSenderId;
  final bool isActive;
  final String role;
  final DateTime joinedAt;
  final DateTime? lastReadAt;
  final int unreadCount;
  final bool isMuted;

  UserConversation({
    required this.userId,
    required this.conversationId,
    required this.type,
    this.title,
    this.lastMessageAt,
    this.lastMessagePreview,
    this.lastMessageSenderId,
    required this.isActive,
    required this.role,
    required this.joinedAt,
    this.lastReadAt,
    required this.unreadCount,
    required this.isMuted,
  });

  factory UserConversation.fromJson(Map<String, dynamic> json) {
    return UserConversation(
      userId: json['user_id'] ?? '',
      conversationId: json['conversation_id'] ?? '',
      type: json['type'] ?? 'direct',
      title: json['title'],
      lastMessageAt: json['last_message_at'] != null 
          ? DateTime.parse(json['last_message_at']) 
          : null,
      lastMessagePreview: json['last_message_preview'],
      lastMessageSenderId: json['last_message_sender_id'],
      isActive: json['is_active'] ?? true,
      role: json['role'] ?? 'member',
      joinedAt: DateTime.parse(json['joined_at'] ?? DateTime.now().toIso8601String()),
      lastReadAt: json['last_read_at'] != null 
          ? DateTime.parse(json['last_read_at']) 
          : null,
      unreadCount: json['unread_count'] ?? 0,
      isMuted: json['is_muted'] ?? false,
    );
  }
}

class ConversationResponse {
  final Conversation conversation;
  final List<ConversationParticipant> participants;

  ConversationResponse({
    required this.conversation,
    required this.participants,
  });

  factory ConversationResponse.fromJson(Map<String, dynamic> json) {
    return ConversationResponse(
      conversation: Conversation.fromJson(json['conversation'] ?? {}),
      participants: (json['participants'] as List<dynamic>?)
          ?.map((e) => ConversationParticipant.fromJson(e))
          .toList() ?? [],
    );
  }
}
