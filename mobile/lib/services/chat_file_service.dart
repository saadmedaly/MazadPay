import 'dart:io';
import 'dart:typed_data';
import 'package:firebase_storage/firebase_storage.dart';
import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as path;
import 'package:uuid/uuid.dart';
import 'package:video_compress/video_compress.dart';
import 'package:flutter_image_compress/flutter_image_compress.dart';

class ChatFileService {
  static final ChatFileService _instance = ChatFileService._internal();
  factory ChatFileService() => _instance;
  ChatFileService._internal();

  FirebaseStorage? _storageInstance;
  
  FirebaseStorage get _storage {
    if (_storageInstance == null) {
      try {
        _storageInstance = FirebaseStorage.instance;
      } catch (e) {
        debugPrint('⚠️ Firebase Storage not initialized: $e');
        throw Exception('Firebase Storage is not available on this platform or not initialized.');
      }
    }
    return _storageInstance!;
  }
  final _uuid = const Uuid();
  
  // Max file size: 10MB
  static const int maxFileSize = 10 * 1024 * 1024;
  
  // Upload progress callback
  Function(double progress)? onProgress;

  /// Upload un fichier vers Firebase Storage
  Future<String?> uploadFile({
    required String userId,
    required File file,
    required String type, // 'audio', 'video', 'image', 'file'
    String? customFileName,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier si Firebase est dispo (surtout sur Web)
      if (kIsWeb) {
        debugPrint('⚠️ File upload is currently limited on web due to Firebase config.');
      }

      // Vérifier la taille du fichier
      final fileSize = await file.length();
      if (fileSize > maxFileSize) {
        throw Exception('File size exceeds 10MB limit');
      }

      // Générer le chemin de stockage
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final fileName = customFileName ?? path.basename(file.path);
      final extension = path.extension(fileName);
      final storagePath = 'chat-files/$type/$userId/${timestamp}_$_uuid$extension';

      // Référence Firebase Storage
      final ref = _storage.ref().child(storagePath);

      // Métadonnées
      final metadata = SettableMetadata(
        contentType: _getContentType(type, extension),
        customMetadata: {
          'uploadedBy': userId,
          'uploadedAt': DateTime.now().toIso8601String(),
          'originalName': fileName,
        },
      );

      // Upload avec suivi de progression
      final uploadTask = ref.putFile(file, metadata);

      if (onProgress != null) {
        uploadTask.snapshotEvents.listen((TaskSnapshot snapshot) {
          final progress = snapshot.bytesTransferred / snapshot.totalBytes;
          onProgress(progress);
        });
      }

      // Attendre la fin de l'upload
      final snapshot = await uploadTask;
      
      // Obtenir l'URL de téléchargement
      final downloadUrl = await snapshot.ref.getDownloadURL();
      
      return downloadUrl;
    } catch (e) {
      debugPrint('Error uploading file: $e');
      return null;
    }
  }

  /// Upload des données binaires (pour les images compressées, etc.)
  Future<String?> uploadBytes({
    required String userId,
    required Uint8List bytes,
    required String type,
    required String extension,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille
      if (bytes.length > maxFileSize) {
        throw Exception('File size exceeds 10MB limit');
      }

      // Générer le chemin de stockage
      final timestamp = DateTime.now().millisecondsSinceEpoch;
      final storagePath = 'chat-files/$type/$userId/${timestamp}_$_uuid.$extension';

      // Référence Firebase Storage
      final ref = _storage.ref().child(storagePath);

      // Métadonnées
      final metadata = SettableMetadata(
        contentType: _getContentType(type, extension),
        customMetadata: {
          'uploadedBy': userId,
          'uploadedAt': DateTime.now().toIso8601String(),
        },
      );

      // Upload
      final uploadTask = ref.putData(bytes, metadata);

      if (onProgress != null) {
        uploadTask.snapshotEvents.listen((TaskSnapshot snapshot) {
          final progress = snapshot.bytesTransferred / snapshot.totalBytes;
          onProgress(progress);
        });
      }

      final snapshot = await uploadTask;
      final downloadUrl = await snapshot.ref.getDownloadURL();
      
      return downloadUrl;
    } catch (e) {
      debugPrint('Error uploading bytes: $e');
      return null;
    }
  }

  /// Compresser et uploader une image
  Future<String?> uploadImage({
    required String userId,
    required File imageFile,
    int quality = 85,
    int maxWidth = 1920,
    int maxHeight = 1080,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Compresser l'image
      final compressedBytes = await FlutterImageCompress.compressWithFile(
        imageFile.absolute.path,
        quality: quality,
        minWidth: maxWidth,
        minHeight: maxHeight,
        rotate: 0,
      );

      if (compressedBytes == null) {
        throw Exception('Failed to compress image');
      }

      // Déterminer l'extension
      final extension = path.extension(imageFile.path).toLowerCase();
      final targetExtension = extension == '.png' ? 'png' : 'jpg';

      // Upload
      return await uploadBytes(
        userId: userId,
        bytes: compressedBytes,
        type: 'images',
        extension: targetExtension,
        onProgress: onProgress,
      );
    } catch (e) {
      debugPrint('Error uploading image: $e');
      // Fallback: upload original file
      return await uploadFile(
        userId: userId,
        file: imageFile,
        type: 'images',
        onProgress: onProgress,
      );
    }
  }

  /// Compresser et uploader une vidéo
  Future<Map<String, dynamic>?> uploadVideo({
    required String userId,
    required File videoFile,
    VideoQuality quality = VideoQuality.DefaultQuality,
    bool deleteOrigin = false,
    Function(double progress)? onProgress,
  }) async {
    try {
      // Vérifier la taille originale
      final originalSize = await videoFile.length();
      
       if (originalSize <= maxFileSize) {
        final url = await uploadFile(
          userId: userId,
          file: videoFile,
          type: 'video',
          onProgress: onProgress,
        );
        
        // Obtenir la durée
        final MediaInfo? info = await VideoCompress.getMediaInfo(videoFile.absolute.path);
        
        final thumbnailFile = await VideoCompress.getFileThumbnail(videoFile.path);
        String? thumbnailUrl;
        if (thumbnailFile != null) {
          thumbnailUrl = await uploadFile(
            userId: userId,
            file: thumbnailFile,
            type: 'thumbnails',
          );
        }

        return {
          'url': url,
          'duration': (info?.duration ?? 0).toInt() ~/ 1000,
          'thumbnail': thumbnailUrl,
        };
      }

      // Compresser la vidéo
      final subscription = VideoCompress.compressProgress$.subscribe((progress) {
        debugPrint('Video compression: $progress%');
      });

      final MediaInfo? info = await VideoCompress.compressVideo(
        videoFile.absolute.path,
        quality: quality,
        deleteOrigin: deleteOrigin,
      );

      subscription.unsubscribe();

      if (info?.file == null) {
        throw Exception('Failed to compress video');
      }

      final compressedFile = info!.file!;
      
      // Upload la vidéo compressée
      final url = await uploadFile(
        userId: userId,
        file: compressedFile,
        type: 'video',
        onProgress: onProgress,
      );

      // Générer une miniature
      String? thumbnailUrl;
      final thumbnailFile = await VideoCompress.getFileThumbnail(compressedFile.path);
      if (thumbnailFile != null) {
        final thumbnailBytes = await thumbnailFile.readAsBytes();
        thumbnailUrl = await uploadBytes(
          userId: userId,
          bytes: thumbnailBytes,
          type: 'thumbnails',
          extension: 'jpg',
        );
      }

      // Nettoyer le fichier compressé temporaire
      if (deleteOrigin && compressedFile.path != videoFile.path) {
        await compressedFile.delete();
      }

      return {
        'url': url,
        'duration': (info.duration ?? 0).toInt() ~/ 1000,
        'thumbnail': thumbnailUrl,
      };
    } catch (e) {
      debugPrint('Error uploading video: $e');
      return null;
    }
  }

  /// Uploader un fichier audio
  Future<Map<String, dynamic>?> uploadAudio({
    required String userId,
    required File audioFile,
    int? duration, // En secondes
    Function(double progress)? onProgress,
  }) async {
    try {
      final url = await uploadFile(
        userId: userId,
        file: audioFile,
        type: 'audio',
        onProgress: onProgress,
      );

      return {
        'url': url,
        'duration': duration ?? 0,
      };
    } catch (e) {
      debugPrint('Error uploading audio: $e');
      return null;
    }
  }

  /// Supprimer un fichier de Firebase Storage
  Future<bool> deleteFile(String fileUrl) async {
    try {
      final ref = _storage.refFromURL(fileUrl);
      await ref.delete();
      return true;
    } catch (e) {
      debugPrint('Error deleting file: $e');
      return false;
    }
  }

  /// Obtenir le type MIME
  String _getContentType(String type, String extension) {
    switch (type) {
      case 'image':
        switch (extension.toLowerCase()) {
          case '.png':
            return 'image/png';
          case '.gif':
            return 'image/gif';
          case '.webp':
            return 'image/webp';
          case '.jpg':
          case '.jpeg':
          default:
            return 'image/jpeg';
        }
      case 'video':
        switch (extension.toLowerCase()) {
          case '.mov':
            return 'video/quicktime';
          case '.avi':
            return 'video/x-msvideo';
          case '.wmv':
            return 'video/x-ms-wmv';
          case '.mp4':
          default:
            return 'video/mp4';
        }
      case 'audio':
        switch (extension.toLowerCase()) {
          case '.wav':
            return 'audio/wav';
          case '.ogg':
            return 'audio/ogg';
          case '.m4a':
            return 'audio/mp4';
          case '.mp3':
          default:
            return 'audio/mpeg';
        }
      case 'file':
        switch (extension.toLowerCase()) {
          case '.pdf':
            return 'application/pdf';
          case '.doc':
            return 'application/msword';
          case '.docx':
            return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
          default:
            return 'application/octet-stream';
        }
      default:
        return 'application/octet-stream';
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
