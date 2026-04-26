import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'auction_winner_page.dart';
import '../services/auction_api.dart';
import '../services/cache_service.dart';

class MyWinningsPage extends ConsumerStatefulWidget {
  const MyWinningsPage({super.key});

  @override
  ConsumerState<MyWinningsPage> createState() => _MyWinningsPageState();
}

class _MyWinningsPageState extends ConsumerState<MyWinningsPage> {
  final AuctionApi _auctionApi = AuctionApi();
  List<Map<String, dynamic>> _winnings = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadWinnings();
  }

  Future<void> _loadWinnings() async {
    try {
      // === OPTIMISTIC UI: Charger depuis le cache d'abord (instantané) ===
      final cachedMyWinnings = await CacheService.instance.getCachedMyWinnings();
      final isCacheValid = await CacheService.instance.isMyWinningsCacheValid();

      if (cachedMyWinnings != null && isCacheValid) {
        // Afficher les données du cache immédiatement (0ms)
        List<dynamic> auctionList = [];
        // cachedMyWinnings est toujours un Map
        auctionList = (cachedMyWinnings['auctions'] ?? cachedMyWinnings['data'] ?? []) as List<dynamic>;
        
        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _winnings = auctionList.map((item) => item as Map<String, dynamic>).toList();
        });
      }

      // === CHARGER DEPUIS L'API EN ARRIÈRE-PLAN ===
      final response = await _auctionApi.getMyWinnings();

      if (response.success && response.data != null) {
        final dynamic responseData = response.data!;
        List<dynamic> auctionList = [];
        // La réponse peut être directement une liste ou un objet avec 'data' ou 'auctions'
        if (responseData is List) {
          auctionList = responseData;
        } else if (responseData is Map<String, dynamic>) {
          auctionList = (responseData['auctions'] ?? responseData['data'] ?? []) as List<dynamic>;
        }
        
        // Cache mes gains (seulement si c'est un Map)
        if (responseData is Map<String, dynamic>) {
          await CacheService.instance.cacheMyWinnings(responseData);
        }

        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _winnings = auctionList.map((item) => item as Map<String, dynamic>).toList();
        });
      } else {
        // Si le cache existe mais l'API échoue, garder le cache
        if (cachedMyWinnings == null && mounted) {
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
          l10n.text_236,
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
                        l10n.error_loading_winnings,
                        style: const TextStyle(color: Colors.grey),
                      ),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadWinnings,
                        child: Text(l10n.retry),
                      ),
                    ],
                  ),
                )
              : _winnings.isEmpty
                  ? Center(
                      child: Text(
                        l10n.no_winnings,
                        style: const TextStyle(color: Colors.grey, fontSize: 16),
                      ),
                    )
                  : ListView.separated(
                      padding: const EdgeInsets.all(20),
                      itemCount: _winnings.length,
                      separatorBuilder: (context, index) => const SizedBox(height: 16),
                      itemBuilder: (context, index) {
                        final winning = _winnings[index];
                        return _buildWinningItem(context, winning, isDarkMode, l10n);
                      },
                    ),
    );
  }

  Widget _buildWinningItem(BuildContext context, Map<String, dynamic> winning, bool isDarkMode, AppLocalizations l10n) {
    // Extraction des données API avec fallbacks
    final id = winning['id']?.toString() ?? '';
    
    // Récupérer le titre avec fallback intelligent
    final locale = Localizations.localeOf(context).languageCode;
    String title = '';
    
    // 1. Essayer d'abord la langue actuelle de l'app
    switch (locale) {
      case 'ar':
        title = winning['title_ar']?.toString() ?? '';
        break;
      case 'fr':
        title = winning['title_fr']?.toString() ?? '';
        break;
      case 'en':
        title = winning['title_en']?.toString() ?? '';
        break;
    }
    
    // 2. Si vide, essayer l'arabe (langue par défaut)
    if (title.isEmpty) {
      title = winning['title_ar']?.toString() ?? '';
    }
    
    // 3. Si toujours vide, essayer les autres langues
    if (title.isEmpty) {
      title = winning['title_fr']?.toString() ??
              winning['title_en']?.toString() ??
              winning['title']?.toString() ??
              l10n.no_title;
    }
    
    final price = '${winning['current_price'] ?? winning['final_price'] ?? winning['current_bid'] ?? 0} MRU';
    final isPaid = winning['is_paid'] == true || winning['payment_status'] == 'paid';

    // Gestion des images
    String imageUrl = 'assets/corolla.png';
    if (winning['images'] != null && winning['images'] is List && (winning['images'] as List).isNotEmpty) {
      imageUrl = (winning['images'] as List)[0].toString();
    } else if (winning['image'] != null) {
      imageUrl = winning['image'].toString();
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: Column(
        children: [
          Row(
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: Image.asset(
                  imageUrl,
                  width: 80,
                  height: 80,
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(
                    width: 80,
                    height: 80,
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
                    const SizedBox(height: 4),
                    Text(
                      price,
                      style: const TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0081FF),
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: isPaid ? const Color(0xFF00C58D).withOpacity(0.1) : Colors.orange.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  isPaid ? l10n.text_237 : l10n.text_238,
                  style: TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: isPaid ? const Color(0xFF00C58D) : Colors.orange,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: OutlinedButton(
                  onPressed: () {
                    if (id.isNotEmpty) {
                      Navigator.of(context).push(
                        MaterialPageRoute(builder: (context) => AuctionWinnerPage(auctionId: id)),
                      );
                    }
                  },
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: Text(
                    l10n.text_239,
                    style: const TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              if (!isPaid) ...[
                const SizedBox(width: 12),
                Expanded(
                  child: ElevatedButton(
                    onPressed: () {
                      // TODO: Implémenter le paiement
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0081FF),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                    child: Text(
                      l10n.text_240,
                      style: const TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }
}
