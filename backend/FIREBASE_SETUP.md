# Firebase Cloud Messaging (FCM) Configuration

## État Actuel
- **Service FCM** : 100% implémenté côté backend
- **Configuration** : Variable d'environnement prête
- **Manque** : Fichier service-account.json

## Configuration Requise

### 1. Variable d'Environnement
```bash
# Dans .env
FIREBASE_SERVICE_ACCOUNT_PATH=./configs/firebase-service-account.json
```

### 2. Structure du Fichier Service Account
Créer le fichier `backend/configs/firebase-service-account.json` :

```json
{
  "type": "service_account",
  "project_id": "votre-projet-firebase",
  "private_key_id": "votre-private-key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n",
  "client_email": "firebase-adminsdk-xxxxx@votre-projet.iam.gserviceaccount.com",
  "client_id": "votre-client-id",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-xxxxx%40votre-projet.iam.gserviceaccount.com",
  "universe_domain": "googleapis.com"
}
```

## Étapes d'Installation

### Étape 1: Créer le Projet Firebase
1. Aller sur [Firebase Console](https://console.firebase.google.com/)
2. Créer un nouveau projet ou utiliser un projet existant
3. Activer Cloud Messaging

### Étape 2: Générer le Service Account
1. Dans Firebase Console > Paramètres du projet > Comptes de service
2. Cliquer sur "Générer une nouvelle clé privée"
3. Choisir "Firebase Admin SDK"
4. Télécharger le fichier JSON
5. Renommer en `firebase-service-account.json`
6. Placer dans `backend/configs/`

### Étape 3: Configurer Web App (Frontend)
1. Dans Firebase Console > Paramètres du projet > Général
2. Ajouter une application Web
3. Copier la configuration dans `web/src/lib/firebase.ts`

### Étape 4: Activer FCM
1. Dans Firebase Console > Cloud Messaging
2. Configurer les clés Web Push (optionnel)
3. Noter la clé serveur pour les tests

## Vérification

### Backend
```bash
# Vérifier que le service FCM est bien initialisé
grep -r "Firebase.*initialized" logs/app.log

# Tester l'endpoint de sauvegarde de token
curl -X POST http://localhost:8082/v1/api/notifications/token \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"fcm_token": "test_token", "device_id": "web", "platform": "web"}'
```

### Frontend
```javascript
// Vérifier que Firebase est initialisé
import { messaging } from '@/lib/firebase';
console.log('Firebase Messaging:', messaging);
```

## Problèmes Courants

### 1. Fichier Non Trouvé
```
Error: no such file or directory, open './configs/firebase-service-account.json'
```
**Solution**: Créer le fichier avec le contenu JSON valide

### 2. Permissions Invalide
```
Error: 7 PERMISSION_DENIED: The caller does not have permission
```
**Solution**: Vérifier que le service account a les permissions FCM

### 3. Token Invalide
```
Error: registration-token-not-registered
```
**Solution**: Le token FCM a expiré, générer un nouveau token

## Mode Dégradé
Si Firebase n'est pas configuré, le backend fonctionne normalement mais sans notifications push.

## Test d'Intégration
```bash
# Démarrer le backend
cd backend && go run cmd/server/main.go

# Envoyer une notification test
curl -X POST http://localhost:8082/v1/api/admin/notifications/send \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test", "body": "Notification test", "user_id": "USER_UUID"}'
```
