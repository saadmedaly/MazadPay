import 'dart:async';
import 'dart:io';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:file_picker/file_picker.dart';
import '../../models/auction.dart';
import '../../models/message.dart';
import '../../services/chat_service.dart';
import '../../services/r2_upload_service.dart';
import '../../widgets/chat/message_bubble.dart';
import '../../widgets/chat/chat_input.dart';
import '../../l10n/app_localizations.dart';
import '../../models/user.dart';
import '../auction_details_page.dart';
import '../../services/auction_api.dart';

class ChatRoomPage extends StatefulWidget {
  final String conversationId;
  final String title;
  final Auction? linkedAuction;
  final String? initialMessage;

  const ChatRoomPage({
    super.key,
    required this.conversationId,
    required this.title,
    this.linkedAuction,
    this.initialMessage,
  });

  @override
  State<ChatRoomPage> createState() => _ChatRoomPageState();
}

class _ChatRoomPageState extends State<ChatRoomPage> {
  final ChatService _chatService = ChatService();
  final R2UploadService _fileService = R2UploadService();
  final TextEditingController _messageController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  final ImagePicker _imagePicker = ImagePicker();
  
  List<Message> _messages = [];
  bool _isLoading = false;
  bool _isSending = false;
  bool _showAuctionContext = false;
  Auction? _currentAuction;
  bool _isTyping = false;
  String? _error;
  int _offset = 0;
  final int _limit = 50;
  bool _hasMore = true;
  
  Timer? _typingTimer;
  String? _currentUserId;
  User? _partnerUser;

  StreamSubscription? _messageSubscription;
  StreamSubscription? _typingSubscription;
  StreamSubscription? _statusSubscription;

  @override
  void initState() {
    super.initState();
    debugPrint('Initializing ChatRoomPage for conversation: ${widget.conversationId}');
    _showAuctionContext = widget.linkedAuction != null;
    _currentAuction = widget.linkedAuction;
    _currentUserId = _chatService.currentUserId;
    debugPrint('Current user ID in ChatRoomPage: $_currentUserId');
    _setupStreams();
    _loadMessages();
    _joinConversation();
    _loadPartnerInfo();
  }

  Future<void> _loadPartnerInfo() async {
    try {
      final response = await _chatService.getConversation(widget.conversationId);
      if (response != null) {
        final otherParticipant = response.participants.firstWhere(
          (p) => p.userId != _currentUserId,
          orElse: () => response.participants.first,
        );
        if (mounted) {
          setState(() {
            _partnerUser = otherParticipant.user;
          });
          
          // Fetch auction if in metadata
          final auctionId = response.conversation.metadata?['auction_id'] as String?;
          if (auctionId != null && _currentAuction == null) {
            _fetchAuction(auctionId);
          }
        }
      }
    } catch (e) {
      debugPrint('Error loading partner info: $e');
    }
  }

  Future<void> _fetchAuction(String auctionId) async {
    try {
      final auctionApi = AuctionApi();
      final response = await auctionApi.getAuctionById(auctionId);
      if (mounted && response.success && response.data != null) {
        setState(() {
          _currentAuction = Auction.fromJson(response.data!);
          _showAuctionContext = true;
        });
      }
    } catch (e) {
      debugPrint('Error fetching auction for chat context: $e');
    }
  }

  void _setupStreams() {
    // Nouveaux messages
    _messageSubscription = _chatService.messageStream.listen((message) {
      if (message.conversationId == widget.conversationId) {
        if (!mounted) return;
        setState(() {
          _messages.insert(0, message);
        });
        _scrollToBottom();
        if (message.senderId != _currentUserId) {
          _markMessagesAsRead();
        }
      }
    });

    // Typing indicators
    _typingSubscription = _chatService.typingStream.listen((data) {
      if (!mounted) return;
      final userId = data['user_id'] as String?;
      final typing = data['typing'] as bool? ?? false;
      
      if (userId != _currentUserId && typing) {
        if (mounted) {
          setState(() {
            _isTyping = true;
          });
        }
      } else {
        if (mounted) {
          setState(() {
            _isTyping = false;
          });
        }
      }
    });

    // Status des messages
    _statusSubscription = _chatService.statusStream.listen((data) {
      if (!mounted) return;
      _updateMessageStatus(data);
    });
  }

  Future<void> _loadMessages({bool refresh = false}) async {
    if (refresh) {
      _offset = 0;
      _hasMore = true;
    }

    if (!_hasMore && !refresh) return;

    if (mounted) {
      setState(() {
        if (refresh) _isLoading = true;
      });
    }

    try {
      final messages = await _chatService.getMessages(
        widget.conversationId,
        limit: _limit,
        offset: _offset,
      );

      if (!mounted) return;
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
      _markMessagesAsRead();
    } catch (e) {
      if (!mounted) return;
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

    if (mounted) {
      setState(() {
        _isSending = true;
      });
    }

    try {
      Map<String, dynamic>? metadata;
      if (_showAuctionContext && _currentAuction != null) {
        metadata = {
          'auction_id': _currentAuction!.id,
          'auction_title': _currentAuction!.title,
          'auction_image': (_currentAuction!.imageUrls != null && _currentAuction!.imageUrls!.isNotEmpty)
              ? _currentAuction!.imageUrls![0]
              : null,
          'auction_description': _currentAuction!.description ?? '',
        };
      }

      final request = SendMessageRequest(
        type: 'text',
        content: text,
        metadata: metadata,
      );

      await _chatService.sendMessage(widget.conversationId, request);
      
      _messageController.clear();
      _chatService.sendTypingStop(widget.conversationId);
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'envoi: $e')),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isSending = false;
        });
      }
    }
  }

  Future<void> _sendImage() async {
    final XFile? image = await _imagePicker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1080,
    );

    if (image == null) return;

    if (mounted) {
      setState(() {
        _isSending = true;
      });
    }

    try {
      final file = File(image.path);
      // Upload vers Cloudflare R2 via backend API
      final url = await _fileService.uploadChatImage(
        conversationId: widget.conversationId,
        imageFile: file,
      );

      if (url != null) {
        Map<String, dynamic>? metadata;
        if (_showAuctionContext && _currentAuction != null) {
          metadata = {
            'auction_id': _currentAuction!.id,
            'auction_title': _currentAuction!.title,
            'auction_image': (_currentAuction!.imageUrls != null && _currentAuction!.imageUrls!.isNotEmpty)
                ? _currentAuction!.imageUrls![0]
                : null,
            'auction_description': _currentAuction!.description ?? '',
          };
        }

        final request = SendMessageRequest(
          type: 'image',
          fileUrl: url,
          fileName: image.name,
          fileSize: await file.length(),
          metadata: metadata,
        );

        await _chatService.sendMessage(widget.conversationId, request);
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Erreur lors de l\'upload de l\'image')),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'upload: $e')),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isSending = false;
        });
      }
    }
  }

  Future<void> _sendVideo() async {
    final XFile? video = await _imagePicker.pickVideo(
      source: ImageSource.gallery,
    );

    if (video == null) return;

    if (mounted) {
      setState(() {
        _isSending = true;
      });
    }

    try {
      final file = File(video.path);
      // Upload vers Cloudflare R2 via backend API
      final result = await _fileService.uploadChatVideo(
        conversationId: widget.conversationId,
        videoFile: file,
      );

      if (result != null) {
        final request = SendMessageRequest(
          type: 'video',
          fileUrl: result['url'],
          fileName: video.name,
          fileSize: await file.length(),
        );

        await _chatService.sendMessage(widget.conversationId, request);
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Erreur lors de l\'upload de la vidéo')),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'upload: $e')),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isSending = false;
        });
      }
    }
  }

  Future<void> _sendAudio() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.audio,
        allowMultiple: false,
      );

      if (result == null || result.files.isEmpty) return;

      if (mounted) {
        setState(() {
          _isSending = true;
        });
      }

      final filePath = result.files.single.path;
      if (filePath == null) return;
      
      final file = File(filePath);
      
      // Upload vers Cloudflare R2 via backend API
      final uploadResult = await _fileService.uploadChatAudio(
        conversationId: widget.conversationId,
        audioFile: file,
        duration: 0, // TODO: Récupérer la durée si possible
      );

      if (uploadResult != null) {
        final request = SendMessageRequest(
          type: 'audio',
          fileUrl: uploadResult['url'],
          fileName: result.files.single.name,
          fileSize: await file.length(),
          fileDuration: uploadResult['duration'],
        );

        await _chatService.sendMessage(widget.conversationId, request);
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Erreur lors de l\'upload de l\'audio')),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur d\'upload: $e')),
      );
    } finally {
      if (mounted) {
        setState(() {
          _isSending = false;
        });
      }
    }
  }


  void _markMessagesAsRead() {
    if (_messages.isEmpty) return;
    
    // Trouver le dernier message non lu du partenaire
    for (final message in _messages) {
      if (message.senderId != _currentUserId) {
        // Vérifier si j'ai déjà lu ce message
        final hasRead = message.status?.any((s) => s.userId == _currentUserId && s.status == 'read') ?? false;
        if (!hasRead) {
          _chatService.markAsRead(widget.conversationId, message.id);
        }
        // Une fois qu'on a trouvé le dernier message (ou qu'on s'arrête au premier non lu), on peut arrêter
        // Car markAsRead côté serveur devrait marquer tous les messages précédents comme lus
        break;
      }
    }
  }

  void _updateMessageStatus(Map<String, dynamic> data) {
    final messageId = data['message_id'] as String?;
    final status = data['status'] as String?;
    final userId = data['user_id'] as String?;
    
    if (messageId != null && status != null && userId != null) {
      if (!mounted) return;
      setState(() {
        final index = _messages.indexWhere((m) => m.id == messageId);
        if (index != -1) {
          final message = _messages[index];
          final currentStatuses = List<MessageStatus>.from(message.status ?? []);
          
          final statusIndex = currentStatuses.indexWhere((s) => s.userId == userId);
          if (statusIndex != -1) {
            currentStatuses[statusIndex] = MessageStatus(
              id: currentStatuses[statusIndex].id,
              messageId: messageId,
              userId: userId,
              status: status,
              updatedAt: DateTime.now(),
            );
          } else {
            currentStatuses.add(MessageStatus(
              id: '',
              messageId: messageId,
              userId: userId,
              status: status,
              updatedAt: DateTime.now(),
            ));
          }

          _messages[index] = Message(
            id: message.id,
            conversationId: message.conversationId,
            senderId: message.senderId,
            type: message.type,
            content: message.content,
            fileName: message.fileName,
            fileUrl: message.fileUrl,
            fileSize: message.fileSize,
            fileDuration: message.fileDuration,
            mimeType: message.mimeType,
            thumbnailUrl: message.thumbnailUrl,
            replyToId: message.replyToId,
            isEdited: message.isEdited,
            isDeleted: message.isDeleted,
            deletedAt: message.deletedAt,
            metadata: message.metadata,
            createdAt: message.createdAt,
            updatedAt: message.updatedAt,
            sender: message.sender,
            replyTo: message.replyTo,
            status: currentStatuses,
          );
        }
      });
    }
  }

  Future<void> _showConversationInfo() async {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final response = await _chatService.getConversation(widget.conversationId);
      if (!mounted) return;
      Navigator.pop(context); // Close loading dialog

      if (response != null) {
        final otherParticipant = response.participants.firstWhere(
          (p) => p.userId != _currentUserId,
          orElse: () => response.participants.first,
        );

        final user = otherParticipant.user;
        if (user != null) {
          final localizations = AppLocalizations.of(context);
          showDialog(
            context: context,
            builder: (context) => AlertDialog(
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
              title: Text(localizations?.info ?? 'Informations', textAlign: TextAlign.center),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CircleAvatar(
                    radius: 40,
                    backgroundColor: Colors.grey[200],
                    backgroundImage: user.profilePicUrl != null && user.profilePicUrl!.isNotEmpty
                        ? NetworkImage(user.profilePicUrl!) 
                        : null,
                    child: user.profilePicUrl == null || user.profilePicUrl!.isEmpty
                        ? const Icon(Icons.person, size: 40, color: Colors.grey) 
                        : null,
                  ),
                  const SizedBox(height: 16),
                  Text(user.fullName ?? 'Utilisateur', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
                  const SizedBox(height: 8),
                  if (user.phone.isNotEmpty) 
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.phone, size: 16, color: Colors.grey),
                        const SizedBox(width: 8),
                        Text(user.phone, style: const TextStyle(color: Colors.grey)),
                      ],
                    ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Impossible de charger les informations')),
          );
        }
      }
    } catch (e) {
      if (!mounted) return;
      Navigator.pop(context); // close loading
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Erreur: $e')),
      );
    }
  }


  @override
  void dispose() {
    debugPrint('Disposing ChatRoomPage for conversation: ${widget.conversationId}');
    _messageSubscription?.cancel();
    _typingSubscription?.cancel();
    _statusSubscription?.cancel();
    _chatService.leaveConversation(widget.conversationId);
    _messageController.dispose();
    _scrollController.dispose();
    _typingTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Scaffold(
      appBar: AppBar(
        title: InkWell(
          onTap: _showConversationInfo,
          child: Row(
            children: [
              CircleAvatar(
                radius: 18,
                backgroundColor: Colors.grey[200],
                backgroundImage: _getProfileImageProvider(_partnerUser?.profilePicUrl),
                child: (_partnerUser?.profilePicUrl == null || _partnerUser!.profilePicUrl!.isEmpty)
                    ? const Icon(Icons.person, size: 20, color: Colors.grey)
                    : null,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      _partnerUser?.fullName ?? widget.title,
                      style: const TextStyle(fontSize: 16),
                      overflow: TextOverflow.ellipsis,
                    ),
                    if (_isTyping)
                      Text(
                        localizations?.text_403 ?? 'En train d\'écrire...',
                        style: const TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.normal,
                          color: Colors.green,
                        ),
                      )
                    else if (_partnerUser?.phone != null)
                      Text(
                        _partnerUser!.phone!,
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.normal,
                          color: isDarkMode ? Colors.grey[400] : Colors.black54,
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.info_outline),
            onPressed: _showConversationInfo,
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _buildMessageList(),
          ),
          _buildAuctionBanner(),
          ChatInput(
            controller: _messageController,
            onSend: _sendTextMessage,
            onImagePick: _sendImage,
            onVideoPick: _sendVideo,
            onAudioPick: _sendAudio,
            onTyping: _onTyping,
            isSending: _isSending,
          ),
        ],
      ),
    );
  }

  Widget _buildMessageList() {
    final localizations = AppLocalizations.of(context);
    
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

  Widget _buildAuctionBanner() {
    if (!_showAuctionContext) return const SizedBox.shrink();
    
    // Use widget.linkedAuction if available, otherwise fetch it if we only have auctionId
    final auction = _currentAuction ?? widget.linkedAuction;
    if (auction == null) return const SizedBox.shrink();

    final isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final primaryColor = Theme.of(context).primaryColor;

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: isDarkMode ? Colors.black45 : Colors.grey[200],
        border: Border(
          top: BorderSide(color: isDarkMode ? Colors.white10 : Colors.black12, width: 0.5),
          bottom: BorderSide(color: isDarkMode ? Colors.white10 : Colors.black12, width: 0.5),
        ),
      ),
      child: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: GestureDetector(
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => AuctionDetailsPage(auctionId: auction.id),
                  ),
                );
              },
              child: Container(
                decoration: BoxDecoration(
                  color: isDarkMode ? Colors.white.withOpacity(0.05) : Colors.white.withOpacity(0.8),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: IntrinsicHeight(
                  child: Row(
                    children: [
                      // Left accent bar
                      Container(
                        width: 4,
                        decoration: BoxDecoration(
                          color: primaryColor,
                          borderRadius: const BorderRadius.only(
                            topLeft: Radius.circular(8),
                            bottomLeft: Radius.circular(8),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      // Text Content
                      Expanded(
                        child: Padding(
                          padding: const EdgeInsets.symmetric(vertical: 8),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                auction.title,
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                  color: primaryColor,
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                              const SizedBox(height: 4),
                              Text(
                                auction.description.isNotEmpty 
                                    ? auction.description 
                                    : (widget.initialMessage ?? 'Aucune description'),
                                style: TextStyle(
                                  fontSize: 12,
                                  color: isDarkMode ? Colors.grey[300] : Colors.grey[700],
                                ),
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ],
                          ),
                        ),
                      ),
                      // Auction Image
                      Padding(
                        padding: const EdgeInsets.all(6),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(4),
                          child: SizedBox(
                            width: 48,
                            height: 48,
                            child: (auction.imageUrls.isNotEmpty)
                                ? _buildAuctionImage(auction.imageUrls.first)
                                : Container(
                                    color: Colors.grey[300],
                                    child: const Icon(Icons.image, size: 24, color: Colors.grey),
                                  ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          // Close button
          Positioned(
            top: 2,
            right: 6,
            child: IconButton(
              icon: const Icon(Icons.close, size: 16),
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(),
              onPressed: () {
                if (mounted) {
                  setState(() {
                    _showAuctionContext = false;
                  });
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  ImageProvider? _getProfileImageProvider(String? url) {
    if (url == null || url.isEmpty) return null;
    if (url.startsWith('data:image')) {
      try {
        final base64String = url.split(',').last;
        return MemoryImage(base64.decode(base64String));
      } catch (e) {
        return null;
      }
    }
    return NetworkImage(url);
  }

  Widget _buildAuctionImage(String imageUrl) {
    if (imageUrl.startsWith('data:image')) {
      try {
        final base64String = imageUrl.split(',').last;
        return Image.memory(
          base64.decode(base64String),
          width: 44,
          height: 44,
          fit: BoxFit.cover,
          errorBuilder: (context, error, stackTrace) => const Icon(Icons.broken_image, size: 20),
        );
      } catch (e) {
        return const Icon(Icons.broken_image, size: 20);
      }
    }

    return Image.network(
      imageUrl,
      width: 44,
      height: 44,
      fit: BoxFit.cover,
      errorBuilder: (context, error, stackTrace) => const Icon(Icons.broken_image, size: 20),
      loadingBuilder: (context, child, loadingProgress) {
        if (loadingProgress == null) return child;
        return Center(
          child: SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(
              strokeWidth: 2,
              value: loadingProgress.expectedTotalBytes != null
                  ? loadingProgress.cumulativeBytesLoaded / loadingProgress.expectedTotalBytes!
                  : null,
            ),
          ),
        );
      },
    );
  }
}
