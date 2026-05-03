import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../models/conversation.dart';
import '../../services/chat_service.dart';
import '../../services/auth_service.dart';
import 'chat_room_page.dart';
import '../../widgets/chat/conversation_tile.dart';
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

  @override
  void initState() {
    super.initState();
    _loadConversations();
    _setupStreams();
  }

  void _setupStreams() {
    // Écouter les nouveaux messages pour actualiser la liste
    _chatService.messageStream.listen((message) {
      _refreshConversations();
    });
  }

  Future<void> _loadConversations({bool refresh = false}) async {
    if (refresh) {
      _offset = 0;
      _hasMore = true;
    }

    if (!_hasMore && !refresh) return;

    setState(() {
      if (refresh) _isLoading = true;
    });

    try {
      final conversations = await _chatService.getConversations(
        limit: _limit,
        offset: _offset,
      );

      setState(() {
        if (refresh) {
          _conversations = conversations;
        } else {
          _conversations.addAll(conversations);
        }
        _hasMore = conversations.length == _limit;
        _offset += conversations.length;
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

  void _refreshConversations() {
    _loadConversations(refresh: true);
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
    // Pour une conversation directe, retourner le nom de l'autre participant
    // Simplifié - à adapter selon vos besoins
    return conversation.title ?? 'Chat';
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context);
    
    return Scaffold(
      appBar: AppBar(
        title: Text(localizations?.text_392 ?? 'Messages'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {
              // TODO: Implémenter la recherche
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
      itemCount: _conversations.length + (_hasMore ? 1 : 0),
      padding: const EdgeInsets.symmetric(vertical: 8),
      itemBuilder: (context, index) {
        if (index == _conversations.length) {
          // Loading more indicator
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(16),
              child: CircularProgressIndicator(),
            ),
          );
        }

        final conversation = _conversations[index];
        return ConversationTile(
          conversation: conversation,
          onTap: () => _navigateToChat(conversation),
        );
      },
    );
  }

  void _showNewChatDialog() {
    // TODO: Implémenter le dialogue de nouvelle conversation
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Nouvelle conversation'),
        content: const Text('Recherchez un utilisateur pour démarrer une conversation'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Annuler'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              // TODO: Naviguer vers la page de recherche d'utilisateurs
            },
            child: const Text('Rechercher'),
          ),
        ],
      ),
    );
  }
}
