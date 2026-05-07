import 'package:flutter/material.dart';
import 'dart:async';
import '../../services/user_api.dart';
import '../../services/chat_service.dart';
import '../../models/user.dart';
import '../../pages/chat/chat_room_page.dart';

class UserSearchDialog extends StatefulWidget {
  const UserSearchDialog({super.key});

  @override
  State<UserSearchDialog> createState() => _UserSearchDialogState();
}

class _UserSearchDialogState extends State<UserSearchDialog> {
  final UserApi _userApi = UserApi();
  final ChatService _chatService = ChatService();
  final TextEditingController _searchController = TextEditingController();
  
  List<User> _results = [];
  bool _isLoading = false;
  String? _error;
  Timer? _debounce;

  void _onSearch(String query) {
    if (_debounce?.isActive ?? false) _debounce!.cancel();
    _debounce = Timer(const Duration(milliseconds: 500), () async {
      if (query.length < 2) {
        setState(() {
          _results = [];
          _error = null;
        });
        return;
      }

      setState(() {
        _isLoading = true;
        _error = null;
      });

      try {
        final response = await _userApi.searchUsers(query);
        if (response.success && response.data != null) {
          setState(() {
            _results = response.data!.map((e) => User.fromJson(e as Map<String, dynamic>)).toList();
            _isLoading = false;
          });
        } else {
          setState(() {
            _error = response.message;
            _isLoading = false;
          });
        }
      } catch (e) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    });
  }

  void _startChat(User user) async {
    setState(() {
      _isLoading = true;
    });

    try {
      final conversation = await _chatService.getDirectConversation(user.id);
      if (!mounted) return;
      
      Navigator.pop(context); // Close dialog
      
      if (conversation != null) {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (context) => ChatRoomPage(
              conversationId: conversation.id,
              title: user.fullName ?? 'Chat',
            ),
          ),
        );
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  void dispose() {
    _debounce?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                const Text(
                  'Nouvelle conversation',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.pop(context),
                ),
              ],
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: 'Rechercher par nom, email ou téléphone',
                prefixIcon: const Icon(Icons.search),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                suffixIcon: _isLoading 
                    ? const Padding(
                        padding: EdgeInsets.all(12),
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : null,
              ),
              onChanged: _onSearch,
            ),
            const SizedBox(height: 16),
            if (_error != null)
              Text(_error!, style: const TextStyle(color: Colors.red)),
            ConstrainedBox(
              constraints: const BoxConstraints(maxHeight: 300),
              child: _results.isEmpty && !_isLoading
                  ? const Padding(
                      padding: EdgeInsets.all(32),
                      child: Text('Aucun utilisateur trouvé'),
                    )
                  : ListView.builder(
                      shrinkWrap: true,
                      itemCount: _results.length,
                      itemBuilder: (context, index) {
                        final user = _results[index];
                        return ListTile(
                          leading: CircleAvatar(
                            backgroundColor: Theme.of(context).colorScheme.primary.withValues(alpha: 0.1),
                            child: Text(
                              user.fullName != null && user.fullName!.isNotEmpty 
                                  ? user.fullName![0].toUpperCase() 
                                  : '?'
                            ),
                          ),
                          title: Text(user.fullName ?? 'Inconnu'),
                          subtitle: Text(user.email ?? user.phone),
                          onTap: () => _startChat(user),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
