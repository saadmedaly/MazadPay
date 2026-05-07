import 'dart:io';
import 'dart:typed_data';
import 'dart:developer' as developer;
import 'package:dio/dio.dart';
import 'package:flutter_image_compress/flutter_image_compress.dart';

class MediaUploadService {
  final Dio dio;

  MediaUploadService(this.dio);

  /// Orchestre le flux complet d'upload vers Cloudflare R2
  /// Retourne l'URL publique de l'image (ou null si échec)
  Future<String?> uploadImage({
    required File file,
    required String context, // 'auction' ou 'profile'
    Function(double)? onProgress,
  }) async {
    try {
      // --- ÉTAPE 1 : COMPRESSION CÔTÉ CLIENT ---
      final originalSize = await file.length();
      Uint8List uploadBytes;

      // Si le fichier est déjà léger (< 200KB), on ne gaspille pas le CPU du téléphone
      if (originalSize < 200 * 1024) {
        developer.log('Fichier < 200KB. Compression ignorée.', name: 'MediaUpload');
        uploadBytes = await file.readAsBytes();
      } else {
        developer.log('Taille originale: ${(originalSize / 1024).toStringAsFixed(1)} KB', name: 'MediaUpload');
        
        final compressed = await FlutterImageCompress.compressWithFile(
          file.absolute.path,
          minWidth: 1200,
          minHeight: 1200, // minWidth/minHeight garantissent que le côté long ne dépasse pas 1200px
          quality: 80,
          format: CompressFormat.webp,
        );

        if (compressed == null) throw Exception('Échec de la compression image');
        
        uploadBytes = compressed;
        developer.log('Taille compressée (WebP): ${(uploadBytes.length / 1024).toStringAsFixed(1)} KB', name: 'MediaUpload');
      }

      // --- ÉTAPE 2 : DEMANDER LA PRESIGNED URL (Fiber Backend) ---
      final presignResponse = await dio.post('/api/v1/media/presign', data: {
        'file_type': 'image/webp',
        'file_size': uploadBytes.length,
        'context': context,
      });

      if (presignResponse.statusCode != 200) {
        throw Exception('Impossible d\'obtenir la signature R2');
      }

      final uploadUrl = presignResponse.data['upload_url'];
      final mediaId = presignResponse.data['media_id'];
      final publicUrl = presignResponse.data['public_url'];

      // --- ÉTAPE 3 : UPLOAD DIRECT VERS CLOUDFLARE R2 ---
      // 0 bande passante pour le backend Fiber !
      final r2Response = await dio.put(
        uploadUrl,
        data: Stream.fromIterable([uploadBytes]), // Streaming évite la copie en RAM par Dio
        options: Options(
          headers: {
            'Content-Type': 'image/webp',
            'Content-Length': uploadBytes.length.toString(),
          },
          // Ne pas utiliser le baseUrl du backend pour cet appel
          receiveTimeout: const Duration(seconds: 30),
          sendTimeout: const Duration(seconds: 30),
        ),
        onSendProgress: (sent, total) {
          if (onProgress != null && total > 0) {
            onProgress(sent / total);
          }
        },
      );

      if (r2Response.statusCode != 200) {
        throw Exception('Échec de l\'upload vers Cloudflare R2');
      }

      // --- ÉTAPE 4 : CONFIRMATION AU BACKEND ---
      await dio.put('/api/v1/media/$mediaId/confirm');

      return publicUrl;
      
    } catch (e, stack) {
      developer.log('Erreur d\'upload R2', error: e, stackTrace: stack, name: 'MediaUpload');
      return null;
    }
  }
}
