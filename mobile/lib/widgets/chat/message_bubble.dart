import 'package:flutter/material.dart';
import '../../models/message.dart';

class MessageBubble extends StatelessWidget {
  final Message message;
  final bool isMe;
  final VoidCallback? onTap;

  const MessageBubble({
    super.key,
    required this.message,
    required this.isMe,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return Align(
      alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
          padding: _getPadding(),
          decoration: BoxDecoration(
            color: message.isDeleted 
                ? Colors.grey[300]
                : isMe 
                    ? theme.colorScheme.primary 
                    : Colors.grey[200],
            borderRadius: BorderRadius.only(
              topLeft: const Radius.circular(16),
              topRight: const Radius.circular(16),
              bottomLeft: Radius.circular(isMe ? 16 : 4),
              bottomRight: Radius.circular(isMe ? 4 : 16),
            ),
          ),
          constraints: BoxConstraints(
            maxWidth: MediaQuery.of(context).size.width * 0.75,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (!isMe && message.sender != null)
                Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Text(
                    message.sender!.fullName ?? 'Unknown',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      color: theme.colorScheme.primary,
                    ),
                  ),
                ),
              _buildContent(context),
              const SizedBox(height: 4),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    _formatTime(message.createdAt),
                    style: TextStyle(
                      fontSize: 11,
                      color: isMe 
                          ? theme.colorScheme.onPrimary.withOpacity(0.7)
                          : Colors.grey[600],
                    ),
                  ),
                  if (isMe) ...[
                    const SizedBox(width: 4),
                    _buildStatusIcon(context),
                  ],
                  if (message.isEdited)
                    Text(
                      ' • modifié',
                      style: TextStyle(
                        fontSize: 10,
                        color: isMe 
                            ? theme.colorScheme.onPrimary.withOpacity(0.7)
                            : Colors.grey[500],
                      ),
                    ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  EdgeInsets _getPadding() {
    if (message.isDeleted) {
      return const EdgeInsets.symmetric(horizontal: 12, vertical: 8);
    }
    
    switch (message.type) {
      case 'image':
      case 'video':
        return const EdgeInsets.all(4);
      case 'audio':
        return const EdgeInsets.symmetric(horizontal: 12, vertical: 8);
      default:
        return const EdgeInsets.symmetric(horizontal: 12, vertical: 8);
    }
  }

  Widget _buildContent(BuildContext context) {
    if (message.isDeleted) {
      return Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            Icons.delete_outline,
            size: 16,
            color: Colors.grey[600],
          ),
          const SizedBox(width: 4),
          Text(
            'Message supprimé',
            style: TextStyle(
              color: Colors.grey[600],
              fontStyle: FontStyle.italic,
            ),
          ),
        ],
      );
    }

    switch (message.type) {
      case 'text':
        return Text(
          message.content ?? '',
          style: TextStyle(
            color: isMe 
                ? Theme.of(context).colorScheme.onPrimary 
                : Colors.black87,
          ),
        );
        
      case 'image':
        return _buildImage(context);
        
      case 'video':
        return _buildVideo(context);
        
      case 'audio':
        return _buildAudio(context);
        
      case 'file':
        return _buildFile(context);
        
      default:
        return Text(
          message.content ?? '',
          style: TextStyle(
            color: isMe 
                ? Theme.of(context).colorScheme.onPrimary 
                : Colors.black87,
          ),
        );
    }
  }

  Widget _buildImage(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: Image.network(
        message.fileUrl!,
        width: 200,
        fit: BoxFit.cover,
        loadingBuilder: (context, child, loadingProgress) {
          if (loadingProgress == null) return child;
          return Container(
            width: 200,
            height: 150,
            color: Colors.grey[300],
            child: const Center(
              child: CircularProgressIndicator(),
            ),
          );
        },
        errorBuilder: (context, error, stackTrace) {
          return Container(
            width: 200,
            height: 150,
            color: Colors.grey[300],
            child: const Icon(Icons.error),
          );
        },
      ),
    );
  }

  Widget _buildVideo(BuildContext context) {
    return GestureDetector(
      onTap: () {
        // TODO: Ouvrir le lecteur vidéo
      },
      child: Container(
        width: 200,
        height: 150,
        decoration: BoxDecoration(
          color: Colors.black,
          borderRadius: BorderRadius.circular(12),
          image: message.thumbnailUrl != null
              ? DecorationImage(
                  image: NetworkImage(message.thumbnailUrl!),
                  fit: BoxFit.cover,
                )
              : null,
        ),
        child: Center(
          child: Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.black54,
              shape: BoxShape.circle,
            ),
            child: const Icon(
              Icons.play_arrow,
              color: Colors.white,
              size: 32,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildAudio(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(
          Icons.mic,
          color: isMe 
              ? Theme.of(context).colorScheme.onPrimary 
              : Colors.grey[700],
        ),
        const SizedBox(width: 8),
        Text(
          'Message audio',
          style: TextStyle(
            color: isMe 
                ? Theme.of(context).colorScheme.onPrimary 
                : Colors.grey[700],
          ),
        ),
        if (message.fileDuration != null) ...[
          const SizedBox(width: 8),
          Text(
            _formatDuration(message.fileDuration!),
            style: TextStyle(
              fontSize: 12,
              color: isMe 
                  ? Theme.of(context).colorScheme.onPrimary.withOpacity(0.7)
                  : Colors.grey[600],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildFile(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(
          Icons.insert_drive_file,
          color: isMe 
              ? Theme.of(context).colorScheme.onPrimary 
              : Colors.grey[700],
        ),
        const SizedBox(width: 8),
        Flexible(
          child: Text(
            message.fileName ?? 'Fichier',
            style: TextStyle(
              color: isMe 
                  ? Theme.of(context).colorScheme.onPrimary 
                  : Colors.grey[700],
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        if (message.fileSize != null) ...[
          const SizedBox(width: 8),
          Text(
            _formatFileSize(message.fileSize!),
            style: TextStyle(
              fontSize: 12,
              color: isMe 
                  ? Theme.of(context).colorScheme.onPrimary.withOpacity(0.7)
                  : Colors.grey[600],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildStatusIcon(BuildContext context) {
    IconData icon = Icons.check;
    Color color = Theme.of(context).colorScheme.onPrimary.withOpacity(0.7);
    
    return Icon(
      icon,
      size: 14,
      color: color,
    );
  }

  String _formatTime(DateTime time) {
    return '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    return '$minutes:${remainingSeconds.toString().padLeft(2, '0')}';
  }

  String _formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }
}
