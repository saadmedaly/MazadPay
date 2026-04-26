import 'package:mezadpay/l10n/app_localizations.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/auction_provider.dart';
import '../models/auction.dart';
import '../models/bid.dart';
import '../widgets/bid_action_sheet.dart';
import '../widgets/auction_winner_dialog.dart';
import '../pages/auction_history_page.dart';
import '../providers/favorites_provider.dart';
import '../pages/all_auctions_page.dart';
import '../services/auction_api.dart';
import '../services/bid_api.dart';
import '../services/cache_service.dart';

class AuctionDetailsPage extends ConsumerStatefulWidget {
  final String auctionId;
  const AuctionDetailsPage({super.key, required this.auctionId});

  @override
  ConsumerState<AuctionDetailsPage> createState() => _AuctionDetailsPageState();
}

class _AuctionDetailsPageState extends ConsumerState<AuctionDetailsPage> {
  Timer? _timer;
  Duration _timeLeft = Duration.zero;
  final PageController _pageController = PageController();
  int _currentPage = 0;
  final AuctionApi _auctionApi = AuctionApi();
  final BidApi _bidApi = BidApi();
  bool _isLoading = true;
  Map<String, dynamic>? _auctionData;
  List<Bid>? _bidHistory;
  late Future<dynamic> _relatedAuctionsFuture; // Cache pour éviter appels multiples
  bool _isDisposed = false;

  @override
  void initState() {
    super.initState();
    _loadAuctionDetails();
    _loadBidHistory();
    _relatedAuctionsFuture = _auctionApi.getAuctions(); // Cache le Future
  }

  Future<void> _loadAuctionDetails() async {
    try {
      debugPrint('=== LOADING AUCTION DETAILS ===');
      debugPrint('Auction ID: ${widget.auctionId}');
      
      // === OPTIMISTIC UI: Charger depuis le cache d'abord (instantané) ===
      final cachedDetail = await CacheService.instance.getCachedAuctionDetail(widget.auctionId);
      final isCacheValid = await CacheService.instance.isAuctionDetailCacheValid(widget.auctionId);

      debugPrint('Cache valid: $isCacheValid');
      debugPrint('Cached detail: ${cachedDetail != null}');

      if (cachedDetail != null && isCacheValid) {
        // Afficher les données du cache immédiatement (0ms)
        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _auctionData = cachedDetail;
          debugPrint('Loaded from cache: ${_auctionData != null}');
          if (_auctionData != null) {
            _startTimer();
          }
        });
      }

      // === CHARGER DEPUIS L'API EN ARRIÈRE-PLAN ===
      debugPrint('Fetching from API...');
      final response = await _auctionApi.getAuctionDetails(widget.auctionId);

      debugPrint('API Response success: ${response.success}');
      debugPrint('API Response data: ${response.data != null}');

      if (response.success && response.data != null) {
        final data = response.data!;
        debugPrint('Data type: ${data.runtimeType}');
        debugPrint('Data keys: ${data is Map ? (data as Map).keys.toList() : "Not a map"}');
        
        Map<String, dynamic>? auctionData;
        
        // Extraction robuste avec multiples fallbacks pour différentes structures API
        if (data is Map<String, dynamic>) {
          auctionData = data['auction'] ?? data['data'] ?? data;
          debugPrint('Extracted auctionData: ${auctionData != null}');
          debugPrint('AuctionData keys: ${auctionData?.keys.toList()}');
          // API retourne les images dans une clé séparée - les fusionner
          if (data['images'] != null && auctionData != null) {
            auctionData['images'] = data['images'];
          }
          // Cache les détails de l'enchère
          if (auctionData != null) {
            await CacheService.instance.cacheAuctionDetail(widget.auctionId, auctionData);
          }
        } else if (data is List && data.isNotEmpty) {
          auctionData = data[0];
          debugPrint('Extracted from list: ${auctionData != null}');
          if (auctionData != null) {
            await CacheService.instance.cacheAuctionDetail(widget.auctionId, auctionData);
          }
        }

        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _auctionData = auctionData;
          debugPrint('Final auctionData: ${_auctionData != null}');
          if (_auctionData != null) {
            _startTimer();
          }
        });
      } else {
        debugPrint('API response failed or no data');
        // Si le cache existe mais l'API échoue, garder le cache
        if (cachedDetail == null && mounted) {
          setState(() => _isLoading = false);
        }
      }
    } catch (e) {
      debugPrint('Error loading auction details: $e');
      setState(() => _isLoading = false);
      if (mounted) {
        final errorStr = e.toString();
        // Check if auction not found
        if (errorStr.contains('not found') || errorStr.contains('404')) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Cette enchère n\'existe plus ou a été supprimée'),
              backgroundColor: Colors.red,
              duration: const Duration(seconds: 3),
            ),
          );
          // Navigate back after showing error
          Future.delayed(const Duration(seconds: 2), () {
            if (mounted) Navigator.of(context).pop();
          });
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(AppLocalizations.of(context)!.error_loading_auction_details)),
          );
        }
      }
    }
  }

  Future<void> _loadBidHistory() async {
    try {
      debugPrint('=== LOADING BID HISTORY ===');
      debugPrint('Auction ID: ${widget.auctionId}');
      final response = await _bidApi.getBidHistory(widget.auctionId);
      
      debugPrint('Bid history response success: ${response.success}');
      
      if (response.success && response.data != null) {
        final data = response.data!;
        List<Bid> bids = [];
        
        if (data is Map<String, dynamic>) {
          final bidsList = data['bids'] as List<dynamic>? ?? [];
          bids = bidsList.map((e) => Bid.fromJson(e as Map<String, dynamic>)).toList();
        } else if (data is List) {
          bids = (data as List<dynamic>).map((e) => Bid.fromJson(e as Map<String, dynamic>)).toList();
        }
        
        debugPrint('Loaded ${bids.length} bids');
        
        if (mounted) {
          setState(() {
            _bidHistory = bids;
          });
        }
      } else {
        if (mounted) {
          setState(() {
            _bidHistory = [];
          });
        }
      }
    } catch (e) {
      debugPrint('Error loading bid history: $e');
      if (mounted) {
        setState(() {
          _bidHistory = [];
        });
      }
    }
  }

  void _startTimer() {
    if (_auctionData == null) return;
    
    final endTimeStr = _auctionData!['end_time'] ?? _auctionData!['ends_at'] ?? '';
    final endTime = DateTime.tryParse(endTimeStr);
    if (endTime == null) return;
    
    _timeLeft = endTime.difference(DateTime.now());
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isDisposed || !mounted) {
        timer.cancel();
        return;
      }
      if (_timeLeft.inSeconds > 0) {
        try {
          if (mounted && !_isDisposed) {
            setState(() {
              _timeLeft = Duration(seconds: _timeLeft.inSeconds - 1);
            });
          }
        } catch (e) {
          timer.cancel();
        }
      } else {
        _timer?.cancel();
        _checkWinner();
      }
    });
  }

  void _checkWinner() {
    // TODO: Implement winner check using API
  }

  @override
  void dispose() {
    _timer?.cancel();
    _timer = null;
    _isDisposed = true;
    super.dispose();
  }

  String _formatDuration(Duration d) {
    String twoDigits(int n) => n.toString().padLeft(2, "0");
    final days = d.inDays;
    final hours = d.inHours.remainder(24);
    final minutes = d.inMinutes.remainder(60);
    final seconds = d.inSeconds.remainder(60);
    
    if (days > 0) {
      return "${twoDigits(days)}j ${twoDigits(hours)}:${twoDigits(minutes)}:${twoDigits(seconds)}";
    }
    return "${twoDigits(hours)}:${twoDigits(minutes)}:${twoDigits(seconds)}";
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    if (_isLoading) {
      return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    if (_auctionData == null) {
      return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: Center(child: Text(AppLocalizations.of(context)!.error_loading_auction)),
      );
    }

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: Column(
          children: [
            Expanded(
              child: CustomScrollView(
                slivers: [
                  _buildSliverAppBar(context, _auctionData!, isDarkMode),
                  SliverToBoxAdapter(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildPriceSection(context, _auctionData!, isDarkMode),
                        _buildMainInfoSection(context, _auctionData!, isDarkMode),
                        _buildExternalLinks(context, _auctionData!, isDarkMode),
                        _buildCarDetailsSection(context, _auctionData!, isDarkMode),
                        _buildBidHistorySection(context, isDarkMode),

                        // Active Auctions Section (NEW)
                        _buildActiveAuctionsSection(context, isDarkMode),
                        
                        const SizedBox(height: 100), // Space for bottom button
                      ],
                    ),
                  ),
                ],
              ),
            ),
            _buildBottomAction(context, _auctionData!, isDarkMode),
          ],
        ),
      );
  }

  Widget _buildSliverAppBar(BuildContext context, Map<String, dynamic> auction, bool isDarkMode) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(widget.auctionId) ?? false;

    return SliverAppBar(
      expandedHeight: 350,
      pinned: true,
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : const Color(0xFF0081FF),
      leadingWidth: 140,
      leading: Padding(
        padding: const EdgeInsetsDirectional.only(end: 16),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
             _buildCircularButton(Icons.share_outlined, () {}),
             const SizedBox(width: 8),
             _buildCircularButton(
                isFavorite ? Icons.favorite : Icons.favorite_border,
                () {
                  ref.read(favoritesProvider.notifier).toggleFavorite(widget.auctionId);
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text(isFavorite ? AppLocalizations.of(context)!.text_55 : AppLocalizations.of(context)!.text_56),
                      duration: const Duration(seconds: 1),
                      behavior: SnackBarBehavior.floating,
                    ),
                  );
                },
                isFavorite: isFavorite,
             ),
          ],
        ),
      ),
      actions: [
        _buildCircularButton(Icons.arrow_forward_ios, () => Navigator.of(context).pop(), isBack: true),
        const SizedBox(width: 16),
      ],
      flexibleSpace: FlexibleSpaceBar(
        background: Stack(
          fit: StackFit.expand,
          children: [
            PageView.builder(
              controller: _pageController,
              onPageChanged: (index) => setState(() => _currentPage = index),
              itemCount: (auction['images'] as List?)?.length ?? 1,
              itemBuilder: (context, index) {
                final images = auction['images'] as List?;
                if (images != null && images.isNotEmpty) {
                  // Gérer format objet (detail API) ou string (list API)
                  final imageItem = images[index];
                  String imageUrl;
                  if (imageItem is Map<String, dynamic>) {
                    imageUrl = imageItem['url']?.toString() ?? '';
                  } else {
                    imageUrl = imageItem.toString();
                  }
                  final isNetworkImage = imageUrl.startsWith('http');
                  if (isNetworkImage) {
                    return Image.network(
                      imageUrl,
                      fit: BoxFit.cover,
                      errorBuilder: (c, e, s) => Container(
                        color: Colors.grey[300],
                        child: const Icon(Icons.image_not_supported, color: Colors.grey),
                      ),
                      loadingBuilder: (context, child, loadingProgress) {
                        if (loadingProgress == null) return child;
                        return Container(
                          color: Colors.grey[200],
                          child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                        );
                      },
                    );
                  } else if (imageUrl.isNotEmpty && !imageUrl.startsWith('{')) {
                    // Asset local valide
                    return Image.asset(
                      imageUrl,
                      fit: BoxFit.cover,
                      errorBuilder: (c, e, s) => Container(
                        color: Colors.grey[300],
                        child: const Icon(Icons.image_not_supported, color: Colors.grey),
                      ),
                    );
                  } else {
                    // URL invalide ou vide - afficher placeholder
                    return Container(
                      color: Colors.grey[300],
                      child: const Icon(Icons.image, color: Colors.grey),
                    );
                  }
                }
                return Container(
                  color: Colors.grey[300],
                  child: const Icon(Icons.image, color: Colors.grey),
                );
              },
            ),
            Positioned(
              bottom: 24,
              right: 24,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.6),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '${_currentPage + 1}/${(auction['images'] as List?)?.length ?? 1} 🗂️',
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontWeight: FontWeight.bold),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCircularButton(IconData icon, VoidCallback onTap, {bool isBack = false, bool isFavorite = false}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 40,
        height: 40,
        decoration: const BoxDecoration(
          color: Colors.white24,
          shape: BoxShape.circle,
        ),
        child: Icon(
          icon, 
          color: isFavorite ? Colors.red : Colors.white, 
          size: isBack ? 18 : 20
        ),
      ),
    );
  }

  Widget _buildPriceSection(BuildContext context, Map<String, dynamic> auction, bool isDarkMode) {
    return Row(
      children: [
        Expanded(
          child: Container(
            height: 100,
            decoration: const BoxDecoration(
              color: Color(0xFF0081FF),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(AppLocalizations.of(context)!.text_57, style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white70, fontSize: 14)),
                const SizedBox(height: 4),
                Text('${auction['current_bid'] ?? auction['current_price'] ?? 0} MRU', 
                    style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontSize: 20, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
        ),
        Expanded(
          child: Container(
            height: 100,
            decoration: BoxDecoration(
              color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
              border: Border(left: BorderSide(color: Colors.grey.withOpacity(0.1))),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(AppLocalizations.of(context)!.text_58, style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey, fontSize: 14)),
                const SizedBox(height: 4),
                Text('${auction['min_increment'] ?? auction['minIncrement'] ?? 0} MRU', 
                    style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: const Color(0xFF0081FF), fontSize: 20, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildMainInfoSection(BuildContext context, Map<String, dynamic> auction, bool isDarkMode) {
    return Container(
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.03), blurRadius: 20, offset: const Offset(0, 10)),
        ],
      ),
      child: Builder(builder: (context) {
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

        return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Text(
                  title,
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 22, fontWeight: FontWeight.bold),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: 8),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text('${auction['views'] ?? 0}', style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey)),
                  const SizedBox(width: 4),
                  const Icon(Icons.visibility_outlined, size: 18, color: Colors.grey),
                ],
              ),
            ],
          ),
          const SizedBox(height: 24),
          
          // Timer Box
          Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: const Color(0xFFF9FAFB),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: Colors.grey.withOpacity(0.1)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildTimerUnit(AppLocalizations.of(context)!.text_59, _formatDuration(_timeLeft).split(':')[2].split(' ')[0]),
                _buildTimerUnit(AppLocalizations.of(context)!.text_60, _formatDuration(_timeLeft).split(':')[1]),
                _buildTimerUnit(AppLocalizations.of(context)!.text_61, _formatDuration(_timeLeft).split(':')[0].split(' ').last),
              ],
            ),
          ),
          
          const SizedBox(height: 24),
          
          // Bid stats and Call
          Row(
            children: [
              Expanded(
                child: GestureDetector(
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (context) => AuctionHistoryPage(auctionId: auction['id']?.toString() ?? widget.auctionId, auctionData: _mapAuctionData(auction))),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.gavel, color: Colors.black, size: 18),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          '${AppLocalizations.of(context)!.text_78} ${auction['bidder_count'] ?? auction['bidderCount'] ?? 0}', 
                          style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontWeight: FontWeight.bold),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: const Color(0xFF00C58D),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  children: [
                    Text(auction['phone_number']?.toString() ?? auction['phoneNumber']?.toString() ?? '', style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 8),
                    const Icon(Icons.phone, color: Colors.white, size: 18),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text('Lot #${auction['lot_number'] ?? auction['lotNumber'] ?? ''}', style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: const Color(0xFF0081FF), fontWeight: FontWeight.bold)),
        ],
      );
      }),
    );
  }

  Widget _buildTimerUnit(String label, String value) {
    return Column(
      children: [
        Text(value, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 24, fontWeight: FontWeight.bold, color: Colors.red)),
        Text(label, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 12, color: Colors.red)),
      ],
    );
  }

  Widget _buildExternalLinks(BuildContext context, Map<String, dynamic> auction, bool isDarkMode) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        children: [
          _buildListTile(Icons.info_outline, AppLocalizations.of(context)!.text_62, () {}, isDarkMode),
          const SizedBox(height: 12),
          _buildListTile(Icons.contact_support_outlined, AppLocalizations.of(context)!.text_63, () {}, isDarkMode),
          const SizedBox(height: 12),
          GestureDetector(
            onTap: () => Navigator.of(context).push(
              MaterialPageRoute(builder: (context) => AuctionHistoryPage(auctionId: auction['id']?.toString() ?? widget.auctionId, auctionData: _mapAuctionData(auction))),
            ),
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(vertical: 14),
              decoration: BoxDecoration(
                color: const Color(0xFF00C58D),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                AppLocalizations.of(context)!.text_64,
                textAlign: TextAlign.center,
                style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCarDetailsSection(BuildContext context, Map<String, dynamic> auction, bool isDarkMode) {
    if (auction['manufacturer'] == null && auction['fuel_type'] == null &&
        auction['transmission'] == null && auction['year'] == null &&
        auction['mileage'] == null && auction['model'] == null) {
      return const SizedBox.shrink();
    }
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFE4E7EC),
          width: 1.5,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: 24,
            runSpacing: 16,
            children: [
              if (auction['manufacturer'] != null)
                _buildCarDetailItem(Icons.home_work_outlined, AppLocalizations.of(context)!.text_65, auction['manufacturer']?.toString() ?? '', isDarkMode),
              if (auction['transmission'] != null)
                _buildCarDetailItem(Icons.vertical_split_outlined, AppLocalizations.of(context)!.text_66, auction['transmission']?.toString() ?? '', isDarkMode),
              if (auction['fuel_type'] != null)
                _buildCarDetailItem(Icons.local_gas_station_outlined, AppLocalizations.of(context)!.text_67, auction['fuel_type']?.toString() ?? '', isDarkMode),
              if (auction['year'] != null)
                _buildCarDetailItem(Icons.calendar_month_outlined, AppLocalizations.of(context)!.text_68, auction['year']?.toString() ?? '', isDarkMode),
              if (auction['mileage'] != null)
                _buildCarDetailItem(Icons.speed_rounded, AppLocalizations.of(context)!.text_69, auction['mileage']?.toString() ?? '', isDarkMode),
              if (auction['model'] != null)
                _buildCarDetailItem(Icons.directions_car_filled, AppLocalizations.of(context)!.text_70, auction['model']?.toString() ?? '', isDarkMode),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCarDetailItem(IconData icon, String label, String value, bool isDarkMode) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 20, color: isDarkMode ? Colors.white70 : Colors.black87),
            const SizedBox(width: 6),
            Text(label, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 13, color: Colors.grey, fontWeight: FontWeight.w500)),
          ],
        ),
        const SizedBox(height: 6),
        Text(value, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 15, fontWeight: FontWeight.w700)),
      ],
    );
  }

  Widget _buildListTile(IconData icon, String title, VoidCallback onTap, bool isDarkMode) {
    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFE4E7EC),
          width: 1.5,
        ),
      ),
      child: ListTile(
        dense: true,
        visualDensity: const VisualDensity(vertical: -2),
        leading: Icon(icon, color: Colors.grey, size: 20),
        title: Text(title, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 14)),
        trailing: const Icon(Icons.arrow_back_ios, size: 16, color: Colors.grey),
        onTap: onTap,
      ),
    );
  }

  Widget _buildBidHistorySection(BuildContext context, bool isDarkMode) {
    final bidderCount = _auctionData?['bidder_count'] ?? 0;
    final bidList = _bidHistory ?? [];
    
    if (bidList.isEmpty && bidderCount == 0) {
      return const SizedBox.shrink();
    }

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFE4E7EC),
          width: 1.5,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                AppLocalizations.of(context)!.text_64,
                style: TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                '${bidList.isNotEmpty ? bidList.length : bidderCount} ${AppLocalizations.of(context)!.text_78}',
                style: TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
                  fontSize: 14,
                  color: Colors.grey,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          if (bidList.isNotEmpty)
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: bidList.length > 5 ? 5 : bidList.length,
              separatorBuilder: (context, index) => Divider(
                color: Colors.grey.withOpacity(0.1),
                height: 1,
              ),
              itemBuilder: (context, index) {
                final bid = bidList[index];
                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 20,
                        backgroundColor: isDarkMode ? const Color(0xFF333333) : const Color(0xFFF5F5F5),
                        child: Text(
                          bid.bidderName?.substring(0, 1).toUpperCase() ?? '?',
                          style: TextStyle(
                            fontFamily: 'Plus Jakarta Sans',
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: isDarkMode ? Colors.white : Colors.black,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              bid.displayBidderName,
                              style: TextStyle(
                                fontFamily: 'Plus Jakarta Sans',
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: isDarkMode ? Colors.white : Colors.black,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              bid.getTimeAgo(),
                              style: TextStyle(
                                fontFamily: 'Plus Jakarta Sans',
                                fontSize: 12,
                                color: Colors.grey,
                              ),
                            ),
                          ],
                        ),
                      ),
                      Text(
                        '${bid.amount} UM',
                        style: TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: isDarkMode ? Colors.white : Colors.black,
                        ),
                      ),
                    ],
                  ),
                );
              },
            )
          else if (bidderCount > 0)
            Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Text(
                  '$bidderCount ${bidderCount == 1 ? 'enchérisseur a participé' : 'enchérisseurs ont participé'}',
                  style: TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    fontSize: 14,
                    color: Colors.grey,
                  ),
                ),
              ),
            ),
          if (bidList.length > 5)
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: GestureDetector(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => AuctionHistoryPage(auctionId: widget.auctionId),
                    ),
                  );
                },
                child: Center(
                  child: Text(
                    AppLocalizations.of(context)!.text_65,
                    style: TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: Theme.of(context).primaryColor,
                    ),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildBottomAction(BuildContext context, Map<String, dynamic> auction, bool isDarkMode) {
    // Extract auction ID with fallbacks - use widget.auctionId as primary source
    final auctionId = auction['id']?.toString() ?? widget.auctionId;
    
    // Extract auction data with fallbacks
    final currentPrice = double.tryParse(
      auction['current_bid']?.toString() ?? 
      auction['current_price']?.toString() ?? 
      auction['starting_price']?.toString() ?? 
      auction['start_price']?.toString() ?? '0'
    ) ?? 0;
    
    final minIncrement = double.tryParse(
      auction['min_increment']?.toString() ?? 
      auction['minIncrement']?.toString() ?? 
      '500'
    ) ?? 500;
    
    final bidCount = auction['bidder_count'] ?? auction['bid_count'] ?? auction['bids_count'] ?? 0;
    
    final timeLeftStr = _formatDuration(_timeLeft);

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 10, offset: const Offset(0, -5)),
        ],
      ),
      child: SafeArea(
        top: false,
        child: SizedBox(
          width: double.infinity,
          height: 60,
          child: ElevatedButton(
            onPressed: auction['is_user_highest_bidder'] == true ? null : () {
              showModalBottomSheet(
                context: context,
                backgroundColor: Colors.transparent,
                isScrollControlled: true,
                builder: (context) => BidActionSheet(
                  auctionId: auctionId,
                  currentPrice: currentPrice,
                  minIncrement: minIncrement,
                  bidCount: bidCount,
                  timeLeft: timeLeftStr,
                ),
              );
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: auction['is_user_highest_bidder'] == true ? const Color(0xFF00C58D) : const Color(0xFF0081FF),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
              disabledBackgroundColor: const Color(0xFF00C58D),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (auction['is_user_highest_bidder'] == true) ...[
                   const Icon(Icons.back_hand_outlined, color: Colors.white),
                   const SizedBox(width: 8),
                   Flexible(
                     child: Text(
                       AppLocalizations.of(context)!.text_71,
                       maxLines: 1,
                       overflow: TextOverflow.ellipsis,
                       style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontWeight: FontWeight.bold),
                     ),
                   ),
                ] else ...[
                   const Icon(Icons.gavel, color: Colors.white),
                   const SizedBox(width: 8),
                   Flexible(
                     child: Text(
                       AppLocalizations.of(context)!.text_72,
                       maxLines: 1,
                       overflow: TextOverflow.ellipsis,
                       style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontWeight: FontWeight.bold),
                     ),
                   ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
  Widget _buildActiveAuctionsSection(BuildContext context, bool isDarkMode) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 32),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20.0),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                AppLocalizations.of(context)!.text_73,
                style: const TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              GestureDetector(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const AllAuctionsPage()),
                  );
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: const Color(0xFF0081FF),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    AppLocalizations.of(context)!.text_74,
                    style: const TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        SizedBox(
          height: 240,
          child: FutureBuilder(
            future: _relatedAuctionsFuture,
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const Center(child: CircularProgressIndicator());
              }

              List<Map<String, dynamic>> relatedAuctions = [];
              if (snapshot.hasData && snapshot.data!.success && snapshot.data!.data != null) {
                final dynamic responseData = snapshot.data!.data!;
                List<dynamic> auctionList = [];
                // La réponse peut être directement une liste ou un objet avec 'data' ou 'auctions'
                if (responseData is List) {
                  auctionList = responseData;
                } else if (responseData is Map<String, dynamic>) {
                  auctionList = (responseData['auctions'] ?? responseData['data'] ?? []) as List<dynamic>;
                }
                relatedAuctions = auctionList.map((item) => item as Map<String, dynamic>).toList();
              }

              if (relatedAuctions.isEmpty) {
                return Center(
                  child: Text(
                    AppLocalizations.of(context)!.no_related_auctions,
                    style: const TextStyle(color: Colors.grey),
                  ),
                );
              }

              return ListView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                children: relatedAuctions.asMap().entries.map((entry) {
                  final index = entry.key;
                  final auction = entry.value;
                  final id = auction['id']?.toString() ?? '';
                  final displayNumber = index + 1; // Auto-incrément
                  final title = (auction['title_ar'] ?? auction['title'])?.toString() ?? AppLocalizations.of(context)!.no_title;
                  final price = '${auction['current_price'] ?? auction['current_bid'] ?? 0} MRU';
                  final timeRaw = auction['end_time'] ?? auction['ends_at'] ?? '';
                  final time = _formatDateTime(timeRaw);

                  // Gestion des images - format objet (detail API) ou string (list API)
                  String imagePath = 'assets/corolla.png';
                  if (auction['images'] != null && auction['images'] is List && (auction['images'] as List).isNotEmpty) {
                    final firstImage = (auction['images'] as List)[0];
                    if (firstImage is Map<String, dynamic>) {
                      imagePath = firstImage['url']?.toString() ?? 'assets/corolla.png';
                    } else {
                      imagePath = firstImage.toString();
                    }
                  } else if (auction['image'] != null) {
                    imagePath = auction['image'].toString();
                  }

                  return _buildRelatedAuctionCard(title, price, time, imagePath, id, isDarkMode, auction, displayNumber);
                }).toList(),
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildRelatedAuctionCard(String title, String price, String time, String imagePath, String id, bool isDarkMode, Map<String, dynamic> auction, int displayNumber) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(id) ?? false;

    return GestureDetector(
      onTap: () {
        if (id == widget.auctionId) return;
        Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => AuctionDetailsPage(auctionId: id)),
        );
      },
      child: Container(
        width: 140,
        margin: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFD0D5DD),
            width: 1.5,
          ),
          boxShadow: [
            BoxShadow(
              color: isDarkMode
                  ? Colors.black.withOpacity(0.3)
                  : const Color(0xFF101828).withOpacity(0.08),
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
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
                  child: imagePath.startsWith('http')
                    ? Image.network(
                        imagePath,
                        height: 120,
                        width: double.infinity,
                        fit: BoxFit.cover,
                        errorBuilder: (c, e, s) => Image.asset('assets/corolla.png', height: 120, width: double.infinity, fit: BoxFit.cover),
                        loadingBuilder: (context, child, loadingProgress) {
                          if (loadingProgress == null) return child;
                          return Container(
                            height: 120,
                            color: Colors.grey[200],
                            child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                          );
                        },
                      )
                    : imagePath.isNotEmpty && !imagePath.startsWith('{')
                      ? Image.asset(imagePath, height: 120, width: double.infinity, fit: BoxFit.cover)
                      : Container(
                          height: 120,
                          color: Colors.grey[300],
                          child: const Center(child: Icon(Icons.image, color: Colors.grey)),
                        ),
                ),
                Positioned(
                  top: 4, right: 4,
                  child: IconButton(
                    iconSize: 18,
                    onPressed: () {
                      ref.read(favoritesProvider.notifier).toggleFavorite(id);
                    },
                    icon: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.5),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Icon(
                        isFavorite ? Icons.favorite : Icons.favorite_border,
                        color: isFavorite ? Colors.red : Colors.black,
                        size: 14,
                      ),
                    ),
                  ),
                ),
                Positioned(
                  top: 8, left: 8,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: const Color(0xFF0084FF),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text('#$displayNumber', style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
                  ),
                ),
              ],
            ),
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 11)),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(Icons.timer_outlined, size: 12, color: Colors.red),
                      const SizedBox(width: 3),
                      Text(time, style: const TextStyle(fontSize: 10, color: Colors.red, fontWeight: FontWeight.w600)),
                    ],
                  ),
                  const SizedBox(height: 2),
                  Text(price, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 10, color: Color(0xFF0081FF))),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatDateTime(String dateTimeStr) {
    if (dateTimeStr.isEmpty) return '';
    try {
      final dateTime = DateTime.parse(dateTimeStr);
      final day = dateTime.day.toString().padLeft(2, '0');
      final month = dateTime.month.toString().padLeft(2, '0');
      final year = dateTime.year;
      final hour = dateTime.hour.toString().padLeft(2, '0');
      final minute = dateTime.minute.toString().padLeft(2, '0');
      return '$day/$month/$year\t $hour:$minute';
    } catch (e) {
      return dateTimeStr;
    }
  }

  Auction _mapAuctionData(Map<String, dynamic> auction) {
    List<String> imageUrls = ['assets/corolla.png'];
    if (auction['images'] != null && auction['images'] is List) {
      imageUrls = (auction['images'] as List).map((img) {
        if (img is Map<String, dynamic>) {
          return img['url']?.toString() ?? 'assets/corolla.png';
        } else {
          return img.toString();
        }
      }).toList();
    } else if (auction['image_urls'] != null && auction['image_urls'] is List) {
      imageUrls = (auction['image_urls'] as List).map((img) {
        if (img is Map<String, dynamic>) {
          return img['url']?.toString() ?? 'assets/corolla.png';
        } else {
          return img.toString();
        }
      }).toList();
    }

    return Auction(
      id: auction['id']?.toString() ?? '',
      title: auction['title_ar'] ?? auction['title'] ?? 'Unknown',
      description: auction['description_ar'] ?? auction['description'] ?? '',
      imageUrls: imageUrls.isNotEmpty ? imageUrls : ['assets/corolla.png'],
      startPrice: double.tryParse(auction['start_price']?.toString() ?? '') ?? double.tryParse(auction['starting_price']?.toString() ?? '') ?? 0,
      currentPrice: double.tryParse(auction['current_price']?.toString() ?? '') ?? double.tryParse(auction['current_bid']?.toString() ?? '') ?? 0,
      minIncrement: double.tryParse(auction['min_increment']?.toString() ?? '') ?? 500,
      endTime: auction['end_time'] != null
          ? DateTime.parse(auction['end_time'])
          : DateTime.now().add(const Duration(hours: 13)),
      bidderCount: auction['bidder_count'] ?? auction['bid_count'] ?? 0,
      views: auction['views'] ?? 0,
      lotNumber: auction['lot_number'] ?? 'N/A',
      phoneNumber: auction['seller_phone'] ?? 'N/A',
      manufacturer: auction['manufacturer'] ?? '',
      fuelType: auction['fuel_type'] ?? '',
      transmission: auction['transmission'] ?? '',
      year: auction['year']?.toString() ?? '',
      mileage: auction['mileage']?.toString() ?? '',
      model: auction['model'] ?? '',
    );
  }
}
