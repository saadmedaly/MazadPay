import 'package:flutter/material.dart';

class ChatInput extends StatelessWidget {
  final TextEditingController controller;
  final VoidCallback onSend;
  final VoidCallback onImagePick;
  final VoidCallback onVideoPick;
  final VoidCallback onTyping;
  final bool isSending;

  const ChatInput({
    super.key,
    required this.controller,
    required this.onSend,
    required this.onImagePick,
    required this.onVideoPick,
    required this.onTyping,
    required this.isSending,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 4,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: Row(
          children: [
            // Bouton attachements
            PopupMenuButton<String>(
              icon: Icon(
                Icons.add_circle_outline,
                color: theme.colorScheme.primary,
              ),
              onSelected: (value) {
                if (value == 'image') {
                  onImagePick();
                } else if (value == 'video') {
                  onVideoPick();
                }
              },
              itemBuilder: (context) => [
                const PopupMenuItem(
                  value: 'image',
                  child: Row(
                    children: [
                      Icon(Icons.image),
                      SizedBox(width: 8),
                      Text('Image'),
                    ],
                  ),
                ),
                const PopupMenuItem(
                  value: 'video',
                  child: Row(
                    children: [
                      Icon(Icons.videocam),
                      SizedBox(width: 8),
                      Text('Vidéo'),
                    ],
                  ),
                ),
              ],
            ),
            
            // Champ de texte
            Expanded(
              child: TextField(
                controller: controller,
                onChanged: (_) => onTyping(),
                onSubmitted: (_) => onSend(),
                maxLines: null,
                textCapitalization: TextCapitalization.sentences,
                decoration: InputDecoration(
                  hintText: 'Message...',
                  filled: true,
                  fillColor: theme.colorScheme.surface,
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 10,
                  ),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide(
                      color: theme.colorScheme.primary.withOpacity(0.5),
                    ),
                  ),
                ),
              ),
            ),
            
            // Bouton envoyer
            IconButton(
              onPressed: isSending ? null : onSend,
              icon: isSending
                  ? const SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Icon(
                      Icons.send,
                      color: controller.text.trim().isEmpty
                          ? Colors.grey
                          : theme.colorScheme.primary,
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
