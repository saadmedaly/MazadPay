import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'package:mezadpay/services/favorites_service.dart';
import 'package:mezadpay/services/api_service.dart';
import 'auction_details_page.dart';
import '../utils/money_formatter.dart';

class FavoritesPage extends ConsumerStatefulWidget {
  const FavoritesPage({super.key});

  @override
  ConsumerState<FavoritesPage> createState() => _FavoritesPageState();
}

class _FavoritesPageState extends ConsumerState<FavoritesPage> {
  Future<List<Map<String, dynamic>>>? _auctionsFuture;

  void _loadAuctions() {
    setState(() {
      _auctionsFuture = FavoritesService().getFavoriteAuctions();
    });
  }

  @override
  void initState() {
    super.initState();
    _auctionsFuture = FavoritesService().getFavoriteAuctions();
  }

  @override
  Widget build(BuildContext context) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        title: Text(
          AppLocalizations.of(context)!.text_28,
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: isDarkMode ? Colors.white : Colors.black,
          ),
        ),
        leading: IconButton(
          icon: Icon(Icons.arrow_back_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.sync, size: 20),
            onPressed: () async {
              final service = FavoritesService();
              await service.syncPendingFavorites();
              await service.migrateLocalFavorites();
              ref.read(favoritesProvider.notifier).refresh();
              _loadAuctions();
              if (context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text(AppLocalizations.of(context)!.favorites_synced)),
                );
              }
            },
          ),
        ],
      ),
      body: favoritesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.error_outline, size: 60, color: Colors.red.withOpacity(0.5)),
              const SizedBox(height: 16),
              Text(
                '${AppLocalizations.of(context)!.error_loading_favorites}: $error',
                style: const TextStyle(color: Colors.grey),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => ref.read(favoritesProvider.notifier).refresh(),
                child: Text(AppLocalizations.of(context)!.retry),
              ),
            ],
          ),
        ),
        data: (favoriteIds) {
          if (favoriteIds.isEmpty) {
            return _buildEmptyState(isDarkMode);
          }
          return FutureBuilder<List<Map<String, dynamic>>>(
            future: _auctionsFuture,
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const Center(child: CircularProgressIndicator());
              }
              if (snapshot.hasError) {
                return Center(child: Text('خطأ في تحميل المزادات', style: TextStyle(color: Colors.grey)));
              }

              final auctions = snapshot.data ?? [];

              if (auctions.isEmpty) {
                final idList = favoriteIds.toList();
                return ListView.builder(
                  padding: const EdgeInsets.all(20),
                  itemCount: idList.length,
                  itemBuilder: (context, index) => _buildPlaceholder(idList[index]),
                );
              }

              return GridView.builder(
                padding: const EdgeInsets.all(20),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  crossAxisSpacing: 16,
                  mainAxisSpacing: 16,
                  childAspectRatio: 0.58,
                ),
                itemCount: auctions.length,
                itemBuilder: (context, index) => _buildFavoriteItem(auctions[index], isDarkMode),
              );
            },
          );
        },
      ),
    );
  }

  Widget _buildEmptyState(bool isDarkMode) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.favorite_border, size: 80, color: Colors.grey.withOpacity(0.3)),
          const SizedBox(height: 16),
          Text(
            AppLocalizations.of(context)!.text_191,
            style: const TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 16, color: Colors.grey),
          ),
        ],
      ),
    );
  }

  Widget _buildPlaceholder(String auctionId) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: const Icon(Icons.favorite, color: Colors.red),
        title: Text('${AppLocalizations.of(context)!.auction} #$auctionId'),
        subtitle: Text(AppLocalizations.of(context)!.data_not_available_offline),
        trailing: IconButton(
          icon: const Icon(Icons.delete_outline, color: Colors.red),
          onPressed: () => ref.read(favoritesProvider.notifier).removeFavorite(auctionId),
        ),
      ),
    );
  }

  Widget _buildFavoriteItem(Map<String, dynamic> auction, bool isDarkMode) {
    final auctionId = auction['id']?.toString() ?? '';

    // Title: prefer Arabic, fallback to French/English
    final title = (auction['title_ar']?.toString().isNotEmpty == true
            ? auction['title_ar']
            : null) ??
        auction['title_fr']?.toString() ??
        auction['title_en']?.toString() ??
        auction['title']?.toString() ??
        '';

    // Price
    final price = auction['current_price'] ?? auction['start_price'] ?? 0;

    // Image URL: images list first, then image_url field
    String? imageUrl;
    final images = auction['images'];
    if (images is List && images.isNotEmpty) {
      imageUrl = images.first?.toString();
    }
    imageUrl ??= auction['image_url']?.toString() ?? auction['cover_image_url']?.toString();
    // Fix relative URLs — derive host from API_BASE_URL env
    if (imageUrl != null && imageUrl.startsWith('/')) {
      final uri = Uri.tryParse(ApiService.apiBaseUrl);
      final host = uri != null
          ? '${uri.scheme}://${uri.host}${uri.hasPort && uri.port != 80 && uri.port != 443 ? ":${uri.port}" : ""}'
          : 'http://localhost:8082';
      imageUrl = '$host$imageUrl';
    }

    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Stack(
            children: [
              ClipRRect(
                borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
                child: imageUrl != null && imageUrl.startsWith('http')
                    ? Image.network(
                        imageUrl,
                        height: 120,
                        width: double.infinity,
                        fit: BoxFit.cover,
                        loadingBuilder: (c, child, progress) => progress == null
                            ? child
                            : Container(
                                height: 120,
                                color: Colors.grey[200],
                                child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                              ),
                        errorBuilder: (c, e, s) => _imagePlaceholder(),
                      )
                    : _imagePlaceholder(),
              ),
              Positioned(
                top: 8,
                left: 8,
                child: GestureDetector(
                  onTap: () => ref.read(favoritesProvider.notifier).removeFavorite(auctionId),
                  child: Container(
                    padding: const EdgeInsets.all(4),
                    decoration: const BoxDecoration(color: Colors.white, shape: BoxShape.circle),
                    child: const Icon(Icons.favorite, color: Colors.red, size: 20),
                  ),
                ),
              ),
            ],
          ),
          Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title.isNotEmpty ? title : AppLocalizations.of(context)!.no_title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  MoneyFormatter.format(
                    num.tryParse(price.toString()) ?? 0,
                    auction['currency_code']?.toString(),
                  ),
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF0081FF),
                  ),
                ),
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: auctionId.isNotEmpty
                        ? () => Navigator.of(context).push(
                              MaterialPageRoute(builder: (_) => AuctionDetailsPage(auctionId: auctionId)),
                            )
                        : null,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0081FF),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 4),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                    ),
                    child: Text(
                      AppLocalizations.of(context)!.text_192,
                      style: const TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 12),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _imagePlaceholder() {
    return Container(
      height: 120,
      width: double.infinity,
      color: Colors.grey[200],
      child: const Icon(Icons.image_not_supported, color: Colors.grey),
    );
  }
}
