import 'package:flutter/material.dart';
import '../../models/conversation.dart';

class ConversationTile extends StatelessWidget {
  final UserConversation conversation;
  final VoidCallback onTap;

  const ConversationTile({
    super.key,
    required this.conversation,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isUnread = conversation.unreadCount > 0;
    
    return ListTile(
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      leading: _buildAvatar(),
      title: Row(
        children: [
          Expanded(
            child: Text(
              _getTitle(),
              style: TextStyle(
                fontWeight: isUnread ? FontWeight.bold : FontWeight.normal,
                fontSize: 16,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          if (conversation.isMuted)
            Icon(
              Icons.volume_off,
              size: 16,
              color: Colors.grey[400],
            ),
        ],
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const SizedBox(height: 4),
          Text(
            conversation.lastMessagePreview ?? 'Aucun message',
            style: TextStyle(
              color: isUnread 
                  ? theme.colorScheme.onSurface 
                  : Colors.grey[600],
              fontWeight: isUnread ? FontWeight.w500 : FontWeight.normal,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 4),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                _formatTime(conversation.lastMessageAt),
                style: TextStyle(
                  fontSize: 12,
                  color: Colors.grey[500],
                ),
              ),
              if (conversation.unreadCount > 0)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primary,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    conversation.unreadCount > 99 
                        ? '99+' 
                        : conversation.unreadCount.toString(),
                    style: TextStyle(
                      color: theme.colorScheme.onPrimary,
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAvatar() {
    // TODO: Afficher l'avatar du participant
    return CircleAvatar(
      radius: 28,
      backgroundColor: Colors.grey[300],
      child: Icon(
        conversation.type == 'direct' 
            ? Icons.person 
            : Icons.group,
        color: Colors.grey[600],
      ),
    );
  }

  String _getTitle() {
    if (conversation.title != null && conversation.title!.isNotEmpty) {
      return conversation.title!;
    }
    
    switch (conversation.type) {
      case 'direct':
        return 'Conversation';
      case 'group':
        return 'Groupe';
      case 'support':
        return 'Support';
      default:
        return 'Chat';
    }
  }

  String _formatTime(DateTime? time) {
    if (time == null) return '';
    
    final now = DateTime.now();
    final difference = now.difference(time);
    
    if (difference.inDays == 0) {
      // Aujourd'hui - afficher l'heure
      return '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
    } else if (difference.inDays == 1) {
      return 'Hier';
    } else if (difference.inDays < 7) {
      // Cette semaine
      final days = ['Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam', 'Dim'];
      return days[time.weekday - 1];
    } else {
      // Plus ancien
      return '${time.day}/${time.month}/${time.year}';
    }
  }
}
