import 'dart:async';
import 'package:flutter/material.dart';
import '../../models/conversation.dart';
import '../../services/chat_service.dart';
import 'chat_room_page.dart';
import '../../widgets/chat/conversation_tile.dart';
import '../../widgets/chat/user_search_dialog.dart';
import '../../l10n/app_localizations.dart';

class ChatListPage extends StatefulWidget {
  const ChatListPage({super.key});

  @override
  State<ChatListPage> createState() => _ChatListPageState();
}

class _ChatListPageState extends State<ChatListPage> {
  final ChatService _chatService = ChatService();
  List<UserConversation> _conversations = [];
  bool _isLoading = true;
  String? _error;
  int _offset = 0;
  final int _limit = 20;
  bool _hasMore = true;
  bool _isSearching = false;
  final TextEditingController _searchController = TextEditingController();
  List<UserConversation> _filteredConversations = [];
  StreamSubscription? _messageSubscription;

  @override
  void initState() {
    super.initState();
    _loadConversations();
    _setupStreams();
  }

  void _setupStreams() {
    // Écouter les nouveaux messages pour actualiser la liste
    _messageSubscription = _chatService.messageStream.listen((message) {
      _refreshConversations();
    });
  }

  @override
  void dispose() {
    _messageSubscription?.cancel();
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadConversations({bool refresh = false}) async {
    if (refresh) {
      _offset = 0;
      _hasMore = true;
    }

    if (!_hasMore && !refresh) return;

    if (!mounted) return;
    setState(() {
      if (refresh) _isLoading = true;
    });

    try {
      final conversations = await _chatService.getConversations(
        limit: _limit,
        offset: _offset,
      );

      if (!mounted) return;
      setState(() {
        if (refresh) {
          _conversations = conversations;
        } else {
          _conversations.addAll(conversations);
        }
        _filteredConversations = _conversations;
        _hasMore = conversations.length == _limit;
        _offset += conversations.length;
        _isLoading = false;
        _error = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  void _refreshConversations() {
    _loadConversations(refresh: true);
  }

  void _onSearchTextChanged(String query) {
    setState(() {
      if (query.isEmpty) {
        _filteredConversations = _conversations;
      } else {
        _filteredConversations = _conversations.where((conv) {
          final title = (conv.title ?? _getParticipantName(conv)).toLowerCase();
          return title.contains(query.toLowerCase());
        }).toList();
      }
    });
  }

  void _navigateToChat(UserConversation conversation) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ChatRoomPage(
          conversationId: conversation.conversationId,
          title: conversation.title ?? _getParticipantName(conversation),
        ),
      ),
    ).then((_) => _refreshConversations());
  }

  String _getParticipantName(UserConversation conversation) {
    if (conversation.title != null && conversation.title!.isNotEmpty) {
      return conversation.title!;
    }

    if (conversation.type == 'direct' && conversation.participantsList.isNotEmpty) {
      try {
        final other = conversation.participantsList.firstWhere(
          (p) => p.userId != conversation.userId,
        );
        if (other.user?.fullName != null) {
          return other.user!.fullName!;
        }
      } catch (_) {}
    }

    return 'Chat';
  }


  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    
    return Scaffold(
      appBar: AppBar(
        title: _isSearching 
            ? TextField(
                controller: _searchController,
                autofocus: true,
                decoration: InputDecoration(
                  hintText: localizations?.text_392 ?? 'Rechercher...',
                  border: InputBorder.none,
                  hintStyle: const TextStyle(color: Colors.white70),
                ),
                style: const TextStyle(color: Colors.white),
                onChanged: _onSearchTextChanged,
              )
            : Text(localizations?.text_392 ?? 'Messages'),
        actions: [
          IconButton(
            icon: Icon(_isSearching ? Icons.close : Icons.search),
            onPressed: () {
              setState(() {
                _isSearching = !_isSearching;
                if (!_isSearching) {
                  _searchController.clear();
                  _filteredConversations = _conversations;
                }
              });
            },
          ),
          PopupMenuButton<String>(
            onSelected: (value) {
              if (value == 'new_chat') {
                _showNewChatDialog();
              }
            },
            itemBuilder: (context) => [
              PopupMenuItem(
                value: 'new_chat',
                child: Text(localizations?.text_402 ?? 'Nouvelle conversation'),
              ),
            ],
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () => _loadConversations(refresh: true),
        child: _buildBody(),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading && _conversations.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null && _conversations.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('Erreur: $_error'),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _refreshConversations,
              child: const Text('Réessayer'),
            ),
          ],
        ),
      );
    }

    if (_conversations.isEmpty) {
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
              'Aucune conversation',
              style: TextStyle(
                fontSize: 18,
                color: Colors.grey[600],
              ),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: _showNewChatDialog,
              child: const Text('Démarrer une conversation'),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      itemCount: _filteredConversations.length + (_hasMore && !_isSearching ? 1 : 0),
      padding: const EdgeInsets.symmetric(vertical: 8),
      itemBuilder: (context, index) {
        if (index == _filteredConversations.length) {
          // Loading more indicator
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(16),
              child: CircularProgressIndicator(),
            ),
          );
        }

        final conversation = _filteredConversations[index];
        return ConversationTile(
          conversation: conversation,
          onTap: () => _navigateToChat(conversation),
        );
      },
    );
  }

  void _showNewChatDialog() {
    showDialog(
      context: context,
      builder: (context) => const UserSearchDialog(),
    ).then((_) => _refreshConversations());
  }
}

