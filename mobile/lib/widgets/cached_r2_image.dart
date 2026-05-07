import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:shimmer/shimmer.dart';

/// Widget hyper-optimisé pour le CDN Cloudflare R2
/// Charge la miniature en priorité, puis la version HD si nécessaire.
/// Met en cache pendant 7 jours via cached_network_image.
class CachedR2Image extends StatelessWidget {
  final String imageUrl;
  final double? width;
  final double? height;
  final BoxFit fit;
  final bool forceThumbnailOnly;

  const CachedR2Image({
    Key? key,
    required this.imageUrl,
    this.width,
    this.height,
    this.fit = BoxFit.cover,
    this.forceThumbnailOnly = false,
  }) : super(key: key);

  /// Génère l'URL de la miniature en remplaçant l'extension par "-thumb.jpg" ou "-thumb.webp"
  /// Fonctionne si le backend respecte la convention de nommage {key}-thumb.ext
  String get thumbnailUrl {
    // Si l'URL n'a pas d'extension reconnaissable, retourner l'original
    final lastDotIndex = imageUrl.lastIndexOf('.');
    if (lastDotIndex == -1 || lastDotIndex < imageUrl.length - 5) {
      return imageUrl;
    }
    
    final withoutExt = imageUrl.substring(0, lastDotIndex);
    final ext = imageUrl.substring(lastDotIndex);
    return '$withoutExt-thumb$ext';
  }

  @override
  Widget build(BuildContext context) {
    // La clé de cache naturelle de CachedNetworkImage est l'URL.
    // L'option maxAge est généralement configurée au niveau global du CacheManager,
    // mais cached_network_image respecte les en-têtes Cache-Control du CDN (7 jours !).

    final String targetUrl = forceThumbnailOnly ? thumbnailUrl : imageUrl;

    return CachedNetworkImage(
      imageUrl: targetUrl,
      width: width,
      height: height,
      fit: fit,
      // La magie du chargement progressif :
      // On affiche le thumbnail pendant que l'image HD télécharge (si on n'est pas en forceThumbnailOnly)
      placeholderFadeInDuration: const Duration(milliseconds: 300),
      placeholder: (context, url) {
        // Dans une GridView, on préfère souvent un shimmer, mais on peut combiner :
        // Si l'URL cible est la HD, on peut essayer de mettre le thumb en placeholder !
        if (!forceThumbnailOnly && imageUrl != thumbnailUrl) {
           return CachedNetworkImage(
             imageUrl: thumbnailUrl,
             width: width,
             height: height,
             fit: fit,
             placeholder: (context, url) => _buildShimmer(),
             errorWidget: (context, url, error) => _buildShimmer(),
           );
        }
        return _buildShimmer();
      },
      errorWidget: (context, url, error) => Container(
        width: width,
        height: height,
        color: Colors.grey[200],
        child: const Icon(Icons.broken_image, color: Colors.grey),
      ),
      // Optimisation mémoire en RAM (ne pas stocker des 4K décodées en RAM)
      memCacheWidth: width != null ? (width! * MediaQuery.of(context).devicePixelRatio).round() : null,
      // memCacheHeight: height != null ? (height! * MediaQuery.of(context).devicePixelRatio).round() : null,
    );
  }

  Widget _buildShimmer() {
    return Shimmer.fromColors(
      baseColor: Colors.grey[300]!,
      highlightColor: Colors.grey[100]!,
      child: Container(
        width: width ?? double.infinity,
        height: height ?? double.infinity,
        color: Colors.white,
      ),
    );
  }
}
