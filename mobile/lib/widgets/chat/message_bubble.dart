import 'package:flutter/material.dart';
import 'dart:convert';
import '../../models/message.dart';
import '../../pages/auction_details_page.dart';

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
              _buildAuctionHeader(context),
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
                          ? theme.colorScheme.onPrimary.withValues(alpha: 0.7)
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
                            ? theme.colorScheme.onPrimary.withValues(alpha: 0.7)
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

  Widget _buildAuctionHeader(BuildContext context) {
    if (message.metadata == null || message.metadata!['auction_id'] == null) {
      return const SizedBox.shrink();
    }

    final auctionTitle = message.metadata!['auction_title'] ?? 'Enchère';
    final auctionImage = message.metadata!['auction_image'] as String?;
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return InkWell(
      onTap: () {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (context) => AuctionDetailsPage(
              auctionId: message.metadata!['auction_id'].toString(),
            ),
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 4),
        decoration: BoxDecoration(
          color: (isMe ? Colors.black26 : Colors.black12),
          borderRadius: BorderRadius.circular(8),
        ),
        child: IntrinsicHeight(
          child: Row(
            children: [
              Container(
                width: 4,
                decoration: BoxDecoration(
                  color: isMe ? Colors.white70 : theme.colorScheme.primary,
                  borderRadius: const BorderRadius.only(
                    topLeft: Radius.circular(8),
                    bottomLeft: Radius.circular(8),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 4),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            isMe ? 'Vous' : (message.sender?.fullName ?? 'Vendeur'),
                            style: TextStyle(
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: isMe ? Colors.white70 : theme.colorScheme.primary,
                            ),
                          ),
                          const Icon(
                            Icons.arrow_forward_ios,
                            size: 10,
                            color: Colors.grey,
                          ),
                        ],
                      ),
                      const SizedBox(height: 2),
                      Text(
                        auctionTitle,
                        style: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                          color: isMe ? Colors.white : Colors.black87,
                          overflow: TextOverflow.ellipsis,
                        ),
                        maxLines: 1,
                      ),
                    ],
                  ),
                ),
              ),
              if (auctionImage != null)
                Padding(
                  padding: const EdgeInsets.all(4),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(4),
                    child: _buildHeaderImage(auctionImage),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHeaderImage(String imageUrl) {
    if (imageUrl.isEmpty) {
      return const SizedBox(width: 40, height: 40);
    }
    if (imageUrl.startsWith('data:image')) {
      try {
        final base64String = imageUrl.split(',').last;
        return Image.memory(
          base64.decode(base64String),
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          errorBuilder: (context, error, stackTrace) =>
              const Icon(Icons.broken_image, size: 20),
        );
      } catch (e) {
        return const Icon(Icons.image, size: 20);
      }
    }
    return Image.network(
      imageUrl,
      width: 40,
      height: 40,
      fit: BoxFit.cover,
      errorBuilder: (context, error, stackTrace) => const Icon(Icons.image, size: 20),
    );
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
    final imageUrl = message.fileUrl!;
    
    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: _buildImageWidget(imageUrl),
    );
  }

  Widget _buildImageWidget(String imageUrl) {
    if (imageUrl.startsWith('data:image')) {
      try {
        final base64String = imageUrl.split(',').last;
        return Image.memory(
          base64Decode(base64String),
          width: 200,
          fit: BoxFit.cover,
          errorBuilder: (context, error, stackTrace) => _buildErrorWidget(),
        );
      } catch (e) {
        return _buildErrorWidget();
      }
    }

    return Image.network(
      imageUrl,
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
      errorBuilder: (context, error, stackTrace) => _buildErrorWidget(),
    );
  }

  Widget _buildErrorWidget() {
    return Container(
      width: 200,
      height: 150,
      color: Colors.grey[300],
      child: const Icon(Icons.error),
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
                  ? Theme.of(context).colorScheme.onPrimary.withValues(alpha: 0.7)
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
                  ? Theme.of(context).colorScheme.onPrimary.withValues(alpha: 0.7)
                  : Colors.grey[600],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildStatusIcon(BuildContext context) {
    if (message.status == null || message.status!.isEmpty) {
      return Icon(
        Icons.access_time,
        size: 14,
        color: Theme.of(context).colorScheme.onPrimary.withValues(alpha: 0.7),
      );
    }

    final statuses = message.status!;
    bool isRead = statuses.any((s) => s.status == 'read');
    bool isDelivered = statuses.any((s) => s.status == 'delivered' || s.status == 'read');

    IconData icon;
    Color color = Theme.of(context).colorScheme.onPrimary;

    if (isRead) {
      icon = Icons.done_all;
      color = Colors.lightBlueAccent;
    } else if (isDelivered) {
      icon = Icons.done_all;
      color = color.withValues(alpha: 0.7);
    } else {
      icon = Icons.done;
      color = color.withValues(alpha: 0.7);
    }

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
