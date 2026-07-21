import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Service de gestion du JWT token
/// Utilise flutter_secure_storage pour stocker le token de manière sécurisée
class AuthService {
  static final AuthService _instance = AuthService._internal();
  factory AuthService() => _instance;
  
  AuthService._internal();
  
  final FlutterSecureStorage _storage = const FlutterSecureStorage(
    aOptions: AndroidOptions(
      encryptedSharedPreferences: true,
    ),
    iOptions: IOSOptions(
      accessibility: KeychainAccessibility.first_unlock,
    ),
  );
  
  static const String _tokenKey = 'jwt_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _userIdKey = 'user_id';
  static const String _hasRegisteredKey = 'has_registered';
  
  /// Sauvegarder le JWT token
  Future<void> saveToken(String token) async {
    await _storage.write(key: _tokenKey, value: token);
    await saveHasRegistered(true);
  }
  
  /// Sauvegarder l'état d'inscription
  Future<void> saveHasRegistered(bool value) async {
    await _storage.write(key: _hasRegisteredKey, value: value.toString());
  }
  
  /// Récupérer l'état d'inscription (si l'utilisateur s'est déjà inscrit ou connecté auparavant)
  Future<bool> hasRegistered() async {
    final val = await _storage.read(key: _hasRegisteredKey);
    return val == 'true';
  }
  
  /// Sauvegarder le refresh token
  Future<void> saveRefreshToken(String refreshToken) async {
    await _storage.write(key: _refreshTokenKey, value: refreshToken);
  }
  
  /// Sauvegarder l'ID utilisateur
  Future<void> saveUserId(String userId) async {
    await _storage.write(key: _userIdKey, value: userId);
  }
  
  /// Récupérer le JWT token
  Future<String?> getToken() async {
    return await _storage.read(key: _tokenKey);
  }
  
  /// Récupérer le refresh token
  Future<String?> getRefreshToken() async {
    return await _storage.read(key: _refreshTokenKey);
  }
  
  /// Récupérer l'ID utilisateur
  Future<String?> getUserId() async {
    return await _storage.read(key: _userIdKey);
  }
  
  /// Vérifier si l'utilisateur est connecté (présence du token uniquement)
  Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null && token.isNotEmpty;
  }

  /// Vérifier si le token existe ET n'est pas expiré (lecture locale du payload JWT,
  /// aucun appel réseau). Retourne false si le token est absent, malformé ou expiré.
  Future<bool> hasValidSession() async {
    final token = await getToken();
    if (token == null || token.isEmpty) return false;

    final parts = token.split('.');
    if (parts.length != 3) return true; // format inattendu: laisser passer, l'API tranchera

    try {
      final normalized = base64Url.normalize(parts[1]);
      final payload = jsonDecode(utf8.decode(base64Url.decode(normalized)));
      final exp = payload['exp'];
      if (exp is int) {
        final expiry = DateTime.fromMillisecondsSinceEpoch(exp * 1000);
        return expiry.isAfter(DateTime.now());
      }
      return true;
    } catch (_) {
      return true; // décodage impossible: laisser passer, l'API tranchera
    }
  }
  
  /// Déconnexion - Supprimer tous les tokens
  Future<void> logout() async {
    await _storage.delete(key: _tokenKey);
    await _storage.delete(key: _refreshTokenKey);
    await _storage.delete(key: _userIdKey);
  }
  
  /// Nettoyer toutes les données stockées
  Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}
