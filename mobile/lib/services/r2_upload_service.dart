import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:dio/dio.dart';
import 'package:path/path.dart' as path;
import 'api_service.dart';

 
class R2UploadService {
  static final R2UploadService _instance = R2UploadService._internal();
  factory R2UploadService() => _instance;
  R2UploadService._internal();

  final ApiService _apiService = ApiService();

  // Max file sizes
  static const int maxImageSize = 10 * 1024 * 1024; // 10MB
  static const int maxVideoSize = 100 * 1024 * 1024; // 100MB
  static const int maxAudioSize = 50 * 1024 * 1024; // 50MB

  /// Upload une image pour le chat vers R2
  Future<String?> uploadChatImage({
    required String conversationId,
    required File imageFile,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille
      final fileSize = await imageFile.length();
      if (fileSize > maxImageSize) {
        throw Exception('Image trop grande (max 10MB)');
      }

      // Créer le FormData
      final fileName = path.basename(imageFile.path);
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          imageFile.path,
          filename: fileName,
        ),
        'type': 'image',
      });

      // Upload vers l'API backend
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/conversations/$conversationId/upload',
        data: formData,
      );

      if (response != null && response['success'] == true) {
        return response['data']['url'] as String?;
      }
      
      return null;
    } catch (e) {
      debugPrint('Error uploading chat image to R2: $e');
      return null;
    }
  }

  /// Upload une vidéo pour le chat vers R2
  Future<Map<String, dynamic>?> uploadChatVideo({
    required String conversationId,
    required File videoFile,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille
      final fileSize = await videoFile.length();
      if (fileSize > maxVideoSize) {
        throw Exception('Vidéo trop grande (max 100MB)');
      }

      // Créer le FormData
      final fileName = path.basename(videoFile.path);
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          videoFile.path,
          filename: fileName,
        ),
        'type': 'video',
      });

      // Upload avec suivi de progression
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/conversations/$conversationId/upload',
        data: formData,
      );

      if (response != null && response['success'] == true) {
        return {
          'url': response['data']['url'],
          'name': response['data']['name'],
          'size': response['data']['size'],
        };
      }
      
      return null;
    } catch (e) {
      debugPrint('Error uploading chat video to R2: $e');
      return null;
    }
  }

  /// Upload un fichier audio pour le chat vers R2
  Future<Map<String, dynamic>?> uploadChatAudio({
    required String conversationId,
    required File audioFile,
    int? duration,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille
      final fileSize = await audioFile.length();
      if (fileSize > maxAudioSize) {
        throw Exception('Audio trop grand (max 50MB)');
      }

      // Créer le FormData
      final fileName = path.basename(audioFile.path);
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          audioFile.path,
          filename: fileName,
        ),
        'type': 'audio',
      });

      // Upload vers l'API backend
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/conversations/$conversationId/upload',
        data: formData,
      );

      if (response != null && response['success'] == true) {
        return {
          'url': response['data']['url'],
          'name': response['data']['name'],
          'size': response['data']['size'],
          'duration': duration ?? 0,
        };
      }
      
      return null;
    } catch (e) {
      debugPrint('Error uploading chat audio to R2: $e');
      return null;
    }
  }

  /// Upload des images pour une auction vers R2
  /// Retourne la liste des URLs des images uploadées
  Future<List<String>> uploadAuctionImages({
    required String auctionId,
    required List<File> images,
    Function(double progress)? onProgress,
  }) async {
    final List<String> uploadedUrls = [];

    for (int i = 0; i < images.length; i++) {
      try {
        final image = images[i];
        
        // Vérifier la taille
        final fileSize = await image.length();
        if (fileSize > maxImageSize) {
          debugPrint('Image $i trop grande, ignorée');
          continue;
        }

        // Créer le FormData
        final fileName = path.basename(image.path);
        final formData = FormData.fromMap({
          'images': await MultipartFile.fromFile(
            image.path,
            filename: fileName,
          ),
        });

        // Upload vers l'API backend
        final response = await _apiService.upload<Map<String, dynamic>>(
          '/auctions/$auctionId/images',
          data: formData,
        );

        if (response != null && response['success'] == true) {
          final urls = response['data']['urls'] as List<dynamic>?;
          if (urls != null && urls.isNotEmpty) {
            uploadedUrls.add(urls.first as String);
          }
        }

        // Notifier la progression
        if (onProgress != null) {
          onProgress((i + 1) / images.length);
        }
      } catch (e) {
        debugPrint('Error uploading auction image $i: $e');
      }
    }

    return uploadedUrls;
  }

  /// Upload une image de profil utilisateur vers R2
  Future<String?> uploadAvatar({
    required File imageFile,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille
      final fileSize = await imageFile.length();
      if (fileSize > maxImageSize) {
        throw Exception('Image trop grande (max 10MB)');
      }

      // Créer le FormData
      final fileName = path.basename(imageFile.path);
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          imageFile.path,
          filename: fileName,
        ),
      });

      // Upload vers l'API backend
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/users/me/avatar/upload',
        data: formData,
      );

      if (response != null && response['success'] == true) {
        return response['data']['url'] as String?;
      }
      
      return null;
    } catch (e) {
      debugPrint('Error uploading avatar to R2: $e');
      return null;
    }
  }

  /// Upload une image pour KYC vers R2
  Future<String?> uploadKYCImage({
    required String type, // 'id_card_front', 'id_card_back', 'selfie'
    required File imageFile,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille
      final fileSize = await imageFile.length();
      if (fileSize > maxImageSize) {
        throw Exception('Image trop grande (max 10MB)');
      }

      // Créer le FormData
      final fileName = path.basename(imageFile.path);
      final formData = FormData.fromMap({
        type: await MultipartFile.fromFile(
          imageFile.path,
          filename: fileName,
        ),
      });

      // Upload vers l'API backend
      final response = await _apiService.upload<Map<String, dynamic>>(
        '/users/kyc',
        data: formData,
      );

      if (response != null && response['success'] == true) {
        return response['data']['url'] as String?;
      }
      
      return null;
    } catch (e) {
      debugPrint('Error uploading KYC image to R2: $e');
      return null;
    }
  }

  /// Formater la taille du fichier pour l'affichage
  static String formatFileSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

  /// Formater la durée pour l'affichage
  static String formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    return '$minutes:${remainingSeconds.toString().padLeft(2, '0')}';
  }
}
