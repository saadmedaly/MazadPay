import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'auction_details_page.dart';
import '../services/auction_api.dart';
import '../services/cache_service.dart';

class MyAuctionsPage extends ConsumerStatefulWidget {
  const MyAuctionsPage({super.key});

  @override
  ConsumerState<MyAuctionsPage> createState() => _MyAuctionsPageState();
}

class _MyAuctionsPageState extends ConsumerState<MyAuctionsPage> {
  final AuctionApi _auctionApi = AuctionApi();
  List<Map<String, dynamic>> _myAuctions = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadMyAuctions();
  }

  Future<void> _loadMyAuctions() async {
    try {
      // === OPTIMISTIC UI: Charger depuis le cache d'abord (instantané) ===
      final cachedMyAuctions = await CacheService.instance.getCachedMyAuctions();
      final isCacheValid = await CacheService.instance.isMyAuctionsCacheValid();

      if (cachedMyAuctions != null && isCacheValid) {
        // Afficher les données du cache immédiatement (0ms)
        List<dynamic> auctionList = [];
        // cachedMyAuctions est toujours un Map
        auctionList = (cachedMyAuctions['auctions'] ?? cachedMyAuctions['data'] ?? []) as List<dynamic>;
        
        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _myAuctions = auctionList.map((item) => item as Map<String, dynamic>).toList();
        });
      }

      // === CHARGER DEPUIS L'API EN ARRIÈRE-PLAN ===
      final response = await _auctionApi.getMyAuctions();

      if (response.success && response.data != null) {
        final dynamic responseData = response.data!;
        List<dynamic> auctionList = [];
        // La réponse peut être directement une liste ou un objet avec 'data' ou 'auctions'
        if (responseData is List) {
          auctionList = responseData;
        } else if (responseData is Map<String, dynamic>) {
          auctionList = (responseData['auctions'] ?? responseData['data'] ?? []) as List<dynamic>;
        }
        
        // Cache mes enchères (seulement si c'est un Map)
        if (responseData is Map<String, dynamic>) {
          await CacheService.instance.cacheMyAuctions(responseData);
        }

        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _myAuctions = auctionList.map((item) => item as Map<String, dynamic>).toList();
        });
      } else {
        // Si le cache existe mais l'API échoue, garder le cache
        if (cachedMyAuctions == null && mounted) {
          setState(() => _isLoading = false);
        }
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
        _error = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        title: Text(
          l10n.text_27,
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
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        l10n.error_loading_my_auctions,
                        style: const TextStyle(color: Colors.grey),
                      ),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadMyAuctions,
                        child: Text(l10n.retry),
                      ),
                    ],
                  ),
                )
              : _myAuctions.isEmpty
                  ? Center(
                      child: Text(
                        l10n.no_my_auctions,
                        style: const TextStyle(color: Colors.grey, fontSize: 16),
                      ),
                    )
                  : ListView.separated(
                      padding: const EdgeInsets.all(20),
                      itemCount: _myAuctions.length,
                      separatorBuilder: (context, index) => const SizedBox(height: 16),
                      itemBuilder: (context, index) {
                        final auction = _myAuctions[index];
                        return _buildAuctionItem(context, auction, isDarkMode, l10n);
                      },
                    ),
    );
  }

  Widget _buildAuctionItem(BuildContext context, Map<String, dynamic> auction, bool isDarkMode, AppLocalizations l10n) {
    // Extraction des données API avec fallbacks
    final id = auction['id']?.toString() ?? '';
    
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
              l10n.no_title;
    }
    
    final price = '${auction['current_price'] ?? auction['current_bid'] ?? 0} MRU';
    final time = auction['end_time'] ?? auction['ends_at'] ?? '';
    final status = auction['status'] ?? 'active';
    final isWinning = status == 'finished' || status == 'won';

    // Gestion des images
    String imageUrl = 'assets/corolla.png';
    if (auction['images'] != null && auction['images'] is List && (auction['images'] as List).isNotEmpty) {
      imageUrl = (auction['images'] as List)[0].toString();
    } else if (auction['image'] != null) {
      imageUrl = auction['image'].toString();
    }

    return GestureDetector(
      onTap: () {
        if (id.isNotEmpty) {
          Navigator.of(context).push(
            MaterialPageRoute(builder: (context) => AuctionDetailsPage(auctionId: id)),
          );
        }
      },
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: Colors.grey.withOpacity(0.1)),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
          ],
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: Image.asset(
                imageUrl,
                width: 100,
                height: 100,
                fit: BoxFit.cover,
                errorBuilder: (c, e, s) => Container(
                  width: 100,
                  height: 100,
                  color: Colors.grey[200],
                  child: const Icon(Icons.image_not_supported, color: Colors.grey),
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    price,
                    style: const TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Color(0xFF0081FF),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Flexible(
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: isWinning
                                ? const Color(0xFF00C58D).withOpacity(0.1)
                                : const Color(0xFFE31B23).withOpacity(0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            isWinning ? l10n.text_234 : l10n.text_235,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              fontFamily: 'Plus Jakarta Sans',
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: isWinning ? const Color(0xFF00C58D) : const Color(0xFFE31B23),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        time,
                        style: const TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          fontSize: 12,
                          color: Colors.grey,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            IconButton(
              icon: const Icon(Icons.arrow_forward_ios, size: 16, color: Colors.grey),
              onPressed: () {
                if (id.isNotEmpty) {
                  Navigator.of(context).push(
                    MaterialPageRoute(builder: (context) => AuctionDetailsPage(auctionId: id)),
                  );
                }
              },
            ),
          ],
        ),
      ),
    );
  }
}
