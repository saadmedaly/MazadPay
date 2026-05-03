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
  
  /// Sauvegarder le JWT token
  Future<void> saveToken(String token) async {
    await _storage.write(key: _tokenKey, value: token);
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
  
  /// Vérifier si l'utilisateur est connecté
  Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null && token.isNotEmpty;
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
