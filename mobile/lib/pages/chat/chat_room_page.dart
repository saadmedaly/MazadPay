import 'dart:async';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import '../../models/message.dart';
import '../../services/chat_service.dart';
import '../../services/chat_file_service.dart';
import '../../widgets/chat/message_bubble.dart';
import '../../widgets/chat/chat_input.dart';
import '../../l10n/app_localizations.dart';

class ChatRoomPage extends StatefulWidget {
  final String conversationId;
  final String title;

  const ChatRoomPage({
    super.key,
    required this.conversationId,
    required this.title,
  });

  @override
  State<ChatRoomPage> createState() => _ChatRoomPageState();
}

class _ChatRoomPageState extends State<ChatRoomPage> {
  final ChatService _chatService = ChatService();
  final ChatFileService _fileService = ChatFileService();
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final ImagePicker _imagePicker = ImagePicker();
  
  List<Message> _messages = [];
  bool _isLoading = true;
  bool _isSending = false;
  bool _isTyping = false;
  String? _error;
  int _offset = 0;
  final int _limit = 50;
  bool _hasMore = true;
  
  Timer? _typingTimer;
  String? _currentUserId;

  @override
  void initState() {
    super.initState();
    debugPrint('Initializing ChatRoomPage for conversation: ${widget.conversationId}');
    _currentUserId = _chatService.currentUserId;
    debugPrint('Current user ID in ChatRoomPage: $_currentUserId');
    _setupStreams();
    _loadMessages();
    _joinConversation();
  }

  void _setupStreams() {
    // Nouveaux messages
    _chatService.messageStream.listen((message) {
      if (message.conversationId == widget.conversationId) {
        setState(() {
          _messages.insert(0, message);
        });
        _scrollToBottom();
      }
    });

    // Typing indicators
    _chatService.typingStream.listen((data) {
      final userId = data['user_id'] as String?;
      final typing = data['typing'] as bool? ?? false;
      
      if (userId != _currentUserId && typing) {
        setState(() {
          _isTyping = true;
        });
      } else {
        setState(() {
          _isTyping = false;
        });
      }
    });

    // Status des messages
    _chatService.statusStream.listen((data) {
      _updateMessageStatus(data);
    });
  }

  Future<void> _loadMessages({bool refresh = false}) async {
    if (refresh) {
      _offset = 0;
      _hasMore = true;
    }

    if (!_hasMore && !refresh) return;

    setState(() {
      if (refresh) _isLoading = true;
    });

    try {
      final messages = await _chatService.getMessages(
        widget.conversationId,
        limit: _limit,
        offset: _offset,
      );

      setState(() {
        if (refresh) {
          _messages = messages;
        } else {
          _messages.addAll(messages);
        }
        _hasMore = messages.length == _limit;
        _offset += messages.length;
        _isLoading = false;
        _error = null;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  void _joinConversation() {
    _chatService.joinConversation(widget.conversationId);
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        0,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  void _onTyping() {
    _typingTimer?.cancel();
    _chatService.sendTypingStart(widget.conversationId);
    
    _typingTimer = Timer(const Duration(seconds: 2), () {
      _chatService.sendTypingStop(widget.conversationId);
    });
  }

  Future<void> _sendTextMessage() async {
    final text = _messageController.text.trim();
    if (text.isEmpty) return;

    setState(() {
      _isSending = true;
    });

    try {
      final request = SendMessageRequest(
        type: 'text',
        content: text,
      );

      await _chatService.sendMessage(widget.conversationId, request);
      
      _messageController.clear();
      _chatService.sendTypingStop(widget.conversationId);
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'envoi: $e')),
      );
    } finally {
      setState(() {
        _isSending = false;
      });
    }
  }

  Future<void> _sendImage() async {
    final XFile? image = await _imagePicker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1080,
    );

    if (image == null) return;

    setState(() {
      _isSending = true;
    });

    try {
      final file = File(image.path);
      final url = await _fileService.uploadImage(
        userId: _currentUserId!,
        imageFile: file,
      );

      if (url != null) {
        final request = SendMessageRequest(
          type: 'image',
          fileUrl: url,
          fileName: image.name,
          fileSize: await file.length(),
        );

        await _chatService.sendMessage(widget.conversationId, request);
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'upload: $e')),
      );
    } finally {
      setState(() {
        _isSending = false;
      });
    }
  }

  Future<void> _sendVideo() async {
    final XFile? video = await _imagePicker.pickVideo(
      source: ImageSource.gallery,
    );

    if (video == null) return;

    setState(() {
      _isSending = true;
    });

    try {
      final file = File(video.path);
      final result = await _fileService.uploadVideo(
        userId: _currentUserId!,
        videoFile: file,
      );

      if (result != null) {
        final request = SendMessageRequest(
          type: 'video',
          fileUrl: result['url'],
          fileName: video.name,
          fileSize: await file.length(),
          fileDuration: result['duration'],
          thumbnailUrl: result['thumbnail'],
        );

        await _chatService.sendMessage(widget.conversationId, request);
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'upload: $e')),
      );
    } finally {
      setState(() {
        _isSending = false;
      });
    }
  }

  void _updateMessageStatus(Map<String, dynamic> data) {
    // Mettre à jour le statut des messages localement
    final messageId = data['message_id'] as String?;
    final status = data['status'] as String?;
    
    if (messageId != null && status != null) {
      setState(() {
        final index = _messages.indexWhere((m) => m.id == messageId);
        if (index != -1) {
          // Mettre à jour le statut
        }
      });
    }
  }

  @override
  void dispose() {
    _chatService.leaveConversation(widget.conversationId);
    _messageController.dispose();
    _scrollController.dispose();
    _typingTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    
    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(widget.title),
            if (_isTyping)
              Text(
                localizations?.text_403 ?? 'En train d\'écrire...',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.normal,
                ),
              ),
          ],
        ),
        actions: [
          PopupMenuButton<String>(
            onSelected: (value) {
              if (value == 'info') {
                // TODO: Afficher les informations de la conversation
              }
            },
            itemBuilder: (context) => [
              PopupMenuItem(
                value: 'info',
                child: Text(localizations?.info ?? 'Informations'),
              ),
            ],
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _buildMessageList(),
          ),
          ChatInput(
            controller: _messageController,
            onSend: _sendTextMessage,
            onImagePick: _sendImage,
            onVideoPick: _sendVideo,
            onTyping: _onTyping,
            isSending: _isSending,
          ),
        ],
      ),
    );
  }

  Widget _buildMessageList() {
    if (_isLoading && _messages.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null && _messages.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('Erreur: $_error'),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => _loadMessages(refresh: true),
              child: const Text('Réessayer'),
            ),
          ],
        ),
      );
    }

    if (_messages.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.chat_bubble_outline,
              size: 64,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 16),
            Text(
              localizations?.text_404 ?? 'Aucun message',
              style: TextStyle(
                fontSize: 18,
                color: Colors.grey[600],
              ),
            ),
            const SizedBox(height: 8),
            Text(
              localizations?.text_405 ?? 'Envoyez votre premier message',
              style: TextStyle(
                fontSize: 14,
                color: Colors.grey[500],
              ),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      reverse: true,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 16),
      itemCount: _messages.length + (_hasMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (index == _messages.length) {
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(16),
              child: CircularProgressIndicator(),
            ),
          );
        }

        final message = _messages[index];
        final isMe = message.senderId == _currentUserId;
        
        return MessageBubble(
          message: message,
          isMe: isMe,
          onTap: () {
            // TODO: Gérer le tap sur le message (répondre, supprimer, etc.)
          },
        );
      },
    );
  }
}
