import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'package:mezadpay/services/favorites_service.dart';
import 'auction_details_page.dart';

class FavoritesPage extends ConsumerWidget {
  const FavoritesPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
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
          // Bouton de synchronisation manuelle
          IconButton(
            icon: const Icon(Icons.sync, size: 20),
            onPressed: () async {
              final service = FavoritesService();
              await service.syncPendingFavorites();
              await service.migrateLocalFavorites();
              ref.read(favoritesProvider.notifier).refresh();
              if (context.mounted) {
                final l10n = AppLocalizations.of(context)!;
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text(l10n.favorites_synced)),
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
            return _buildEmptyState(context, isDarkMode);
          }
          return _buildFavoritesList(context, ref, isDarkMode, favoriteIds.toList());
        },
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context, bool isDarkMode) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.favorite_border, size: 80, color: Colors.grey.withOpacity(0.3)),
          const SizedBox(height: 16),
          Text(
            AppLocalizations.of(context)!.text_191,
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 16,
              color: Colors.grey,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            AppLocalizations.of(context)!.favorites_local_storage,
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 12,
              color: Colors.grey.withOpacity(0.7),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFavoritesList(BuildContext context, WidgetRef ref, bool isDarkMode, List<String> favoriteIds) {
    // TODO: Récupérer les données complètes des enchères depuis le cache ou l'API
    // Pour l'instant, affichons juste les IDs avec des placeholders
    return FutureBuilder<List<Map<String, dynamic>>>(
      future: FavoritesService().getFavoriteAuctions(),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        final auctions = snapshot.data ?? [];
        
        if (auctions.isEmpty) {
          // Si pas de données en cache, afficher les IDs avec placeholder
          return ListView.builder(
            padding: const EdgeInsets.all(20),
            itemCount: favoriteIds.length,
            itemBuilder: (context, index) {
              return _buildFavoriteItemPlaceholder(
                context,
                ref,
                isDarkMode,
                favoriteIds[index],
              );
            },
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
          itemBuilder: (context, index) {
            final auction = auctions[index];
            return _buildFavoriteItem(
              context,
              ref,
              isDarkMode,
              auction,
            );
          },
        );
      },
    );
  }

  Widget _buildFavoriteItemPlaceholder(
    BuildContext context,
    WidgetRef ref,
    bool isDarkMode,
    String auctionId,
  ) {
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

  Widget _buildFavoriteItem(
    BuildContext context,
    WidgetRef ref,
    bool isDarkMode,
    Map<String, dynamic> auction,
  ) {
    final auctionId = auction['id']?.toString() ?? auction['auction_id']?.toString() ?? '';
    
    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Stack(
            children: [
              ClipRRect(
                borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
                child: Image.asset(
                  auction['images'] is List && (auction['images'] as List).isNotEmpty
                      ? auction['images'][0]
                      : 'assets/corolla.png',
                  height: 120,
                  width: double.infinity,
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(
                    height: 120,
                    color: Colors.grey[300],
                    child: const Icon(Icons.image_not_supported, color: Colors.grey),
                  ),
                ),
              ),
              Positioned(
                top: 8,
                left: 8,
                child: GestureDetector(
                  onTap: () => ref.read(favoritesProvider.notifier).removeFavorite(auctionId),
                  child: Container(
                    padding: const EdgeInsets.all(4),
                    decoration: const BoxDecoration(
                      color: Colors.white,
                      shape: BoxShape.circle,
                    ),
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
                Builder(builder: (context) {
                  // Récupérer le titre avec fallback intelligent
                  final locale = Localizations.localeOf(context).languageCode;
                  String title = '';
                  
                  // 1. Essayer d'abord la langue actuelle de l'app
                  switch (locale) {
                    case 'ar':
                      title = auction['title_ar']?.toString() ?? '';
                      break;
                    case 'fr':
                      title = auction['title_fr']?.toString() ?? '';
                      break;
                    case 'en':
                      title = auction['title_en']?.toString() ?? '';
                      break;
                  }
                  
                  // 2. Si vide, essayer l'arabe (langue par défaut)
                  if (title.isEmpty) {
                    title = auction['title_ar']?.toString() ?? '';
                  }
                  
                  // 3. Si toujours vide, essayer les autres langues
                  if (title.isEmpty) {
                    title = auction['title_fr']?.toString() ??
                            auction['title_en']?.toString() ??
                            auction['title']?.toString() ??
                            '';
                  }
                  
                  // 4. Si toujours vide, afficher "Sans titre"
                  if (title.isEmpty) title = AppLocalizations.of(context)!.no_title;
                  
                  return Text(
                    title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  );
                }),
                const SizedBox(height: 8),
                Text(
                  '${auction['current_price'] ?? auction['current_bid'] ?? auction['price'] ?? 0} MRU',
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
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => AuctionDetailsPage(
                            auctionId: auctionId,
                          ),
                        ),
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0081FF),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 4),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    child: Text(
                      AppLocalizations.of(context)!.text_192,
                      style: const TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 12,
                      ),
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
}
