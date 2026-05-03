# Système de Messagerie - Guide de Configuration

## Vue d'ensemble

Le système de messagerie MazadPay permet la communication en temps réel entre:
- **Utilisateurs** (acheteurs)
- **Vendeurs**
- **Administrateurs**

### Fonctionnalités

- ✅ Conversations directes (1-à-1)
- ✅ Conversations de groupe
- ✅ Support client (chat avec admin)
- ✅ Messages texte
- ✅ Messages média (audio, vidéo, images)
- ✅ Fichiers (max 10MB)
- ✅ Indicateurs de frappe en temps réel
- ✅ Statuts de message (envoyé, livré, lu)
- ✅ Notifications push hors ligne
- ✅ Chiffrement via HTTPS/WSS

---

## Architecture

### Backend (Go + PostgreSQL)

**Tables créées:**
- `conversations` - Stockage des conversations
- `conversation_participants` - Membres des conversations
- `messages` - Contenu des messages
- `message_status` - Statut de lecture par utilisateur

**Services:**
- `ChatService` - Logique métier du chat
- `ChatHub` - Gestion WebSocket

### Frontend (Flutter)

**Modèles:**
- `Conversation` / `ConversationParticipant` / `UserConversation`
- `Message` / `MessageStatus`

**Services:**
- `ChatService` - WebSocket + API REST
- `ChatFileService` - Upload Firebase Storage

**UI:**
- `ChatListPage` - Liste des conversations
- `ChatRoomPage` - Conversation active
- `ConversationTile`, `MessageBubble`, `ChatInput`

---

## Configuration

### 1. Migrations SQL

Exécutez la migration 000033:
```bash
cd backend
migrate -path migrations -database "postgres://user:password@localhost/mazadpay?sslmode=disable" up
```

### 2. Firebase Storage (Médias)

**Dépendances Flutter:**
```yaml
dependencies:
  firebase_storage: ^11.6.0
  firebase_core: ^2.24.0
  image_picker: ^1.0.7
  video_compress: ^3.1.2
  flutter_image_compress: ^2.1.0
```

**Configuration:**
1. Copiez `firebase/storage.rules` dans votre projet Firebase
2. Déployez les règles:
   ```bash
   firebase deploy --only storage
   ```

### 3. WebSocket (Go)

Le WebSocket est automatiquement configuré dans `routes/routes.go`:

```go
// Le hub démarre automatiquement
go chatHub.Run()
```

**Endpoint WebSocket:**
```
wss://api.mazadpay.com/v1/api/chat/ws
```

### 4. Routes API REST

| Méthode | Endpoint | Description |
|-----------|----------|-------------|
| GET | `/conversations` | Liste des conversations |
| POST | `/conversations` | Créer une conversation |
| GET | `/conversations/:id` | Détails d'une conversation |
| POST | `/conversations/:id/join` | Rejoindre une conversation |
| POST | `/conversations/:id/leave` | Quitter une conversation |
| GET | `/conversations/direct/:user_id` | Conversation directe |
| GET | `/conversations/:id/messages` | Messages d'une conversation |
| POST | `/conversations/:id/messages` | Envoyer un message |
| POST | `/conversations/:id/read` | Marquer comme lu |
| PUT | `/messages/:id` | Modifier un message |
| DELETE | `/messages/:id` | Supprimer un message |

---

## Utilisation (Flutter)

### 1. Connexion WebSocket

```dart
final chatService = ChatService();
await chatService.connect(userId, authToken);
```

### 2. Rejoindre une conversation

```dart
chatService.joinConversation(conversationId);
```

### 3. Envoyer un message texte

```dart
await chatService.sendMessage(conversationId, SendMessageRequest(
  type: 'text',
  content: 'Bonjour!',
));
```

### 4. Envoyer une image

```dart
final file = File(imagePath);
final url = await ChatFileService().uploadImage(
  userId: userId,
  imageFile: file,
);

await chatService.sendMessage(conversationId, SendMessageRequest(
  type: 'image',
  fileUrl: url,
  fileName: 'image.jpg',
));
```

### 5. Écouter les messages

```dart
chatService.messageStream.listen((message) {
  // Traiter le nouveau message
});
```

---

## Types de Messages Supportés

| Type | Description | Max Size |
|------|-------------|----------|
| `text` | Texte simple | - |
| `image` | JPG, PNG, GIF, WebP | 10MB |
| `video` | MP4, MOV, AVI | 10MB |
| `audio` | MP3, WAV, M4A, OGG | 10MB |
| `file` | PDF, DOC, etc. | 10MB |

---

## Sécurité

1. **Authentification JWT** requise pour toutes les routes
2. **WebSocket authentifié** via header `Authorization: Bearer {token}`
3. **Autorisation** - les utilisateurs ne peuvent accéder qu'à leurs conversations
4. **Limite de taille** - 10MB max via validation backend et Firebase Storage rules
5. **Types de fichiers** - validation côté backend et Firebase

---

## Events WebSocket

| Event | Description |
|-------|-------------|
| `message:new` | Nouveau message reçu |
| `message:read` | Message marqué comme lu |
| `message:delivered` | Message livré |
| `typing:start` | Utilisateur commence à écrire |
| `typing:stop` | Utilisateur arrête d'écrire |
| `user:online` | Utilisateur connecté |
| `user:offline` | Utilisateur déconnecté |
| `conversation:join` | Rejoindre une room |
| `conversation:leave` | Quitter une room |

---

## Prochaines Étapes

- [ ] Tests unitaires backend
- [ ] Tests widget Flutter
- [ ] Compression vidéo plus avancée
- [ ] Messages vocaux (enregistrement)
- [ ] Réactions aux messages
- [ ] Messages épinglés
- [ ] Recherche dans les conversations
