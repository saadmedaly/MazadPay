import 'package:mezadpay/l10n/app_localizations.dart';
import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/auction.dart';
import '../models/bid.dart';
import '../widgets/bid_action_sheet.dart';
import '../widgets/auction_description_section.dart';
import '../pages/auction_history_page.dart';
import 'all_auctions_page.dart';
import '../providers/favorites_provider.dart';
import '../services/auction_api.dart';
import '../services/bid_api.dart';
import '../services/cache_service.dart';
import '../services/auth_service.dart';
import '../pages/how_to_bid_page.dart';
import '../widgets/app_modals.dart';
import 'package:url_launcher/url_launcher.dart';
import '../providers/auction_provider_api.dart';
import '../models/bid.dart' as model_bid;
import 'package:cached_network_image/cached_network_image.dart';
import 'package:shimmer/shimmer.dart';
import '../utils/time_utils.dart';
import '../utils/money_formatter.dart';

class AuctionDetailsPage extends ConsumerStatefulWidget {
  final String auctionId;
  const AuctionDetailsPage({super.key, required this.auctionId});

  @override
  ConsumerState<AuctionDetailsPage> createState() => _AuctionDetailsPageState();
}

class _AuctionDetailsPageState extends ConsumerState<AuctionDetailsPage> {
  final AuctionApi _auctionApi = AuctionApi();
  final BidApi _bidApi = BidApi();

  // State from Provider
  String? _userId;

  // Timer state
  Duration _timeLeft = Duration.zero;
  Timer? _timer;

  // UI State
  int _currentPage = 0;
  final PageController _pageController = PageController();

  // Related auctions
  Future<dynamic>? _relatedAuctionsFuture;

  // Temporary bid history (could be refactored to provider later)
  List<model_bid.Bid>? _bidHistory;
  bool _isDisposed = false;

  @override
  void initState() {
    super.initState();
    debugPrint(
      'Initializing AuctionDetailsPage for auction: ${widget.auctionId}',
    );
    _loadUserId();
    _relatedAuctionsFuture = _auctionApi.getAuctions(); // Cache le Future

    // Incrémenter les vues une seule fois à l'ouverture
    _auctionApi.incrementViews(widget.auctionId);
  }

  Future<void> _loadUserId() async {
    final userId = await AuthService().getUserId();
    if (mounted) {
      setState(() {
        _userId = userId;
      });
    }
  }

  // ─── WhatsApp support contact ───────────────────────────────────────────
  static const String _supportPhone = '22236601175';

  Future<void> _openWhatsApp() async {
    final uri = Uri.parse('https://wa.me/$_supportPhone');
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('تعذّر فتح واتساب. تأكد من تثبيته على جهازك.'),
          ),
        );
      }
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('تعذّر فتح واتساب. تأكد من تثبيته على جهازك.'),
        ),
      );
    }
  }

  void _showTermsBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => DraggableScrollableSheet(
        expand: false,
        initialChildSize: 0.6,
        minChildSize: 0.4,
        maxChildSize: 0.9,
        builder: (_, controller) => Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'أحكام وشروط المزايدة',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              Expanded(
                child: ListView(
                  controller: controller,
                  children: const [
                    Text(
                      'تتم المزايدة وفق الشروط المحددة من إدارة MazadPay.\n\n'
                      '١. يجب على المزايد الالتزام بالسعر الذي يقدمه، وفي حال الفوز يجب إكمال عملية الدفع خلال المدة المحددة.\n\n'
                      '٢. لا يحق للمزايد سحب عرضه بعد تقديمه.\n\n'
                      '٣. تحتفظ الإدارة بحق إلغاء أي مزايدة مخالفة للأنظمة والقوانين المعمول بها.\n\n'
                      '٤. في حال عدم إتمام الدفع خلال المهلة المحددة، يحق للإدارة إلغاء الفوز ومنحه للمزايد التالي.\n\n'
                      '٥. جميع المعلومات المقدمة يجب أن تكون صحيحة ودقيقة.\n\n'
                      '٦. تخضع جميع النزاعات لقوانين جمهورية موريتانيا الإسلامية.',
                      style: TextStyle(
                        fontSize: 15,
                        height: 1.6,
                        color: Colors.black87,
                      ),
                      textDirection: TextDirection.rtl,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
  // ────────────────────────────────────────────────────────────────────────

  Future<void> _loadBidHistory() async {
    try {
      debugPrint('=== LOADING BID HISTORY ===');
      debugPrint('Auction ID: ${widget.auctionId}');
      final response = await _bidApi.getBidHistory(widget.auctionId);

      debugPrint('Bid history response success: ${response.success}');

      if (response.success && response.data != null) {
        final data = response.data!;
        List<Bid> bids = [];

        final bidsList = data['bids'] as List<dynamic>? ?? [];
        bids = bidsList
            .map((e) => Bid.fromJson(e as Map<String, dynamic>))
            .toList();

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

  void _startTimer(Auction auction) {
    final endTime = auction.endTime;
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
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final auctionAsync = ref.watch(
      auctionNotifierApiProvider(widget.auctionId),
    );

    return auctionAsync.when(
      loading: () => Scaffold(
        backgroundColor: isDarkMode
            ? const Color(0xFF121212)
            : const Color(0xFFFBFBFB),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (err, stack) => Scaffold(
        backgroundColor: isDarkMode
            ? const Color(0xFF121212)
            : const Color(0xFFFBFBFB),
        body: Center(
          child: Text(AppLocalizations.of(context)!.error_loading_auction),
        ),
      ),
      data: (auction) {
        // Démarrer le timer si pas encore fait
        if (_timer == null) {
          _startTimer(auction);
        }

        // Mettre à jour le temps restant à chaque build basé sur le timer
        _updateTimeLeft(auction.endTime);

        return Scaffold(
          backgroundColor: isDarkMode
              ? const Color(0xFF121212)
              : const Color(0xFFFBFBFB),
          body: Column(
            children: [
              Expanded(
                child: RefreshIndicator(
                  onRefresh: () async {
                    ref.invalidate(
                      auctionNotifierApiProvider(widget.auctionId),
                    );
                    await ref.read(
                      auctionNotifierApiProvider(widget.auctionId).future,
                    );
                  },
                  child: CustomScrollView(
                    slivers: [
                      _buildSliverAppBar(context, auction, isDarkMode),
                      SliverToBoxAdapter(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            _buildPriceSection(context, auction, isDarkMode),
                            _buildMainInfoSection(context, auction, isDarkMode),
                            _buildDescriptionSection(
                              context,
                              auction,
                              isDarkMode,
                            ),
                            _buildExternalLinks(context, auction, isDarkMode),
                            _buildCarDetailsSection(
                              context,
                              auction,
                              isDarkMode,
                            ),
                            _buildItemPropertiesSection(
                              context,
                              auction,
                              isDarkMode,
                            ),
                            _buildBidHistorySection(
                              context,
                              isDarkMode,
                              auction,
                            ),

                            // Action Buttons (Auto-Bid, Boost, Report, etc.)
                            _buildActionButtons(context, isDarkMode, auction),

                            // Active Auctions Section (NEW)
                            _buildActiveAuctionsSection(context, isDarkMode),

                            const SizedBox(
                              height: 100,
                            ), // Space for bottom button
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              _buildBottomAction(context, auction, isDarkMode),
            ],
          ),
        );
      },
    );
  }

  void _updateTimeLeft(DateTime endTime) {
    _timeLeft = endTime.difference(DateTime.now());
    if (_timeLeft.isNegative) _timeLeft = Duration.zero;
  }

  Widget _buildSliverAppBar(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite =
        favoritesAsync.value?.contains(widget.auctionId) ?? false;

    return SliverAppBar(
      expandedHeight: 350,
      pinned: true,
      backgroundColor: isDarkMode
          ? const Color(0xFF1D1D1D)
          : const Color(0xFF0081FF),
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
                ref
                    .read(favoritesProvider.notifier)
                    .toggleFavorite(widget.auctionId);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(
                      isFavorite
                          ? AppLocalizations.of(context)!.text_55
                          : AppLocalizations.of(context)!.text_56,
                    ),
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
        _buildCircularButton(
          Icons.arrow_forward_ios,
          () => Navigator.of(context).pop(),
          isBack: true,
        ),
        const SizedBox(width: 16),
      ],
      flexibleSpace: FlexibleSpaceBar(
        background: Stack(
          fit: StackFit.expand,
          children: [
            PageView.builder(
              controller: _pageController,
              onPageChanged: (index) => setState(() => _currentPage = index),
              itemCount: auction.imageUrls.length,
              itemBuilder: (context, index) {
                final imageUrl = auction.imageUrls[index];
                if (imageUrl.startsWith('http')) {
                  return CachedNetworkImage(
                    imageUrl: imageUrl,
                    fit: BoxFit.cover,
                    placeholder: (context, url) => Shimmer.fromColors(
                      baseColor: Colors.grey[300]!,
                      highlightColor: Colors.grey[100]!,
                      child: Container(color: Colors.white),
                    ),
                    errorWidget: (context, url, error) => Container(
                      color: Colors.grey[300],
                      child: const Icon(
                        Icons.image_not_supported,
                        color: Colors.grey,
                      ),
                    ),
                  );
                } else if (imageUrl.startsWith('data:image')) {
                  try {
                    final base64String = imageUrl.split(',').last;
                    return Image.memory(
                      base64.decode(base64String),
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stackTrace) => Container(
                        color: Colors.grey[300],
                        child: const Icon(
                          Icons.image_not_supported,
                          color: Colors.grey,
                        ),
                      ),
                    );
                  } catch (e) {
                    return Container(
                      color: Colors.grey[300],
                      child: const Icon(
                        Icons.image_not_supported,
                        color: Colors.grey,
                      ),
                    );
                  }
                } else if (imageUrl.isNotEmpty) {
                  return Image.asset(
                    imageUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) => Container(
                      color: Colors.grey[300],
                      child: const Icon(
                        Icons.image_not_supported,
                        color: Colors.grey,
                      ),
                    ),
                  );
                } else {
                  return Container(
                    color: Colors.grey[300],
                    child: const Icon(Icons.image, color: Colors.grey),
                  );
                }
              },
            ),
            Positioned(
              bottom: 24,
              right: 24,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
                decoration: BoxDecoration(
                  color: Colors.black.withValues(alpha: 0.6),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '${_currentPage + 1}/${auction.imageUrls.length} 🗂️',
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCircularButton(
    IconData icon,
    VoidCallback onTap, {
    bool isBack = false,
    bool isFavorite = false,
  }) {
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
          size: isBack ? 18 : 20,
        ),
      ),
    );
  }

  Widget _buildPriceSection(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    return Row(
      children: [
        Expanded(
          child: Container(
            height: 100,
            decoration: const BoxDecoration(color: Color(0xFF0081FF)),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  AppLocalizations.of(context)!.text_57,
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    color: Colors.white70,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  MoneyFormatter.format(auction.currentPrice, auction.currencyCode),
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    color: Colors.white,
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ),
        Expanded(
          child: Container(
            height: 100,
            decoration: BoxDecoration(
              color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
              border: Border(
                left: BorderSide(color: Colors.grey.withValues(alpha: 0.1)),
              ),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  AppLocalizations.of(context)!.text_58,
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    color: Colors.grey,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  MoneyFormatter.format(auction.minIncrement, auction.currencyCode),
                  style: const TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    color: Color(0xFF0081FF),
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDescriptionSection(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    return AuctionDescriptionSection(
      description: auction.description,
      isDarkMode: isDarkMode,
    );
  }

  Widget _buildMainInfoSection(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    return Container(
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.03),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Builder(
        builder: (context) {
          final title = auction.title.isEmpty
              ? AppLocalizations.of(context)!.no_title
              : auction.title;

          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(
                      title,
                      style: TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 22,
                        fontWeight: FontWeight.bold,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        '${auction.views}',
                        style: const TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          color: Colors.grey,
                        ),
                      ),
                      const SizedBox(width: 4),
                      const Icon(
                        Icons.visibility_outlined,
                        size: 18,
                        color: Colors.grey,
                      ),
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
                  border: Border.all(color: Colors.grey.withValues(alpha: 0.1)),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    // Client feedback (countdown correction): the label/value
                    // pairing here was previously shifted by one unit
                    // (text_59 "Sec" shown under the days value, text_60
                    // "Min" under hours, text_61 "Hour" under minutes) with
                    // no seconds displayed at all despite _timeLeft already
                    // being second-accurate. Fixed to the correct labels and
                    // a real seconds column reusing the same Timer.periodic.
                    _buildTimerUnit(
                      AppLocalizations.of(context)!.day,
                      _timeLeft.inDays.toString().padLeft(2, '0'),
                    ),
                    _buildTimerUnit(
                      AppLocalizations.of(context)!.text_61,
                      (_timeLeft.inHours % 24).toString().padLeft(2, '0'),
                    ),
                    _buildTimerUnit(
                      AppLocalizations.of(context)!.text_60,
                      (_timeLeft.inMinutes % 60).toString().padLeft(2, '0'),
                    ),
                    _buildTimerUnit(
                      AppLocalizations.of(context)!.text_59,
                      (_timeLeft.inSeconds % 60).toString().padLeft(2, '0'),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Stats Section - Description et Visites
              // The "عدد المزايدين"/"عدد المزايدات" (bid count) box used to
              // live here, duplicating the count already shown next to
              // "مشاهدة كل المزايدين" in _buildBidHistorySection below. It is
              // replaced with the auction's actual product description
              // (client feedback A9); null/empty description hides this
              // half of the row gracefully instead of showing a blank box.
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.grey[50],
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey.withOpacity(0.2)),
                ),
                child: Row(
                  children: [
                    // Product description (replaces duplicate bid-count box)
                    if (auction.description.trim().isNotEmpty)
                      Expanded(
                        child: Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(8),
                              decoration: BoxDecoration(
                                color: const Color(
                                  0xFF0081FF,
                                ).withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: const Icon(
                                Icons.description_outlined,
                                color: Color(0xFF0081FF),
                                size: 20,
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Text(
                                auction.description.trim(),
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                                style: TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 13,
                                  fontWeight: FontWeight.w600,
                                  color: isDarkMode
                                      ? Colors.white
                                      : Colors.black87,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),

                    // Separator (only when both sides have content)
                    if (auction.description.trim().isNotEmpty)
                      Container(
                        width: 1,
                        height: 40,
                        margin: const EdgeInsets.symmetric(horizontal: 8),
                        color: Colors.grey.withValues(alpha: 0.3),
                      ),

                    // Nombre de visiteurs
                    Expanded(
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: const Color(
                                0xFF00C58D,
                              ).withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(
                              Icons.visibility_outlined,
                              color: Color(0xFF00C58D),
                              size: 20,
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  '${auction.views}',
                                  style: const TextStyle(
                                    fontFamily: 'Plus Jakarta Sans',
                                    fontSize: 18,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xFF00C58D),
                                  ),
                                ),
                                Text(
                                  AppLocalizations.of(context)!.views,
                                  style: TextStyle(
                                    fontFamily: 'Plus Jakarta Sans',
                                    fontSize: 12,
                                    color: Colors.grey[600],
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // Phone and Lot
              Row(
                children: [
                  Expanded(
                    child: GestureDetector(
                      onTap: () => _openWhatsApp(),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 12,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFF25D366),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(
                              Icons.phone,
                              color: Colors.white,
                              size: 18,
                            ),
                            const SizedBox(width: 8),
                            const Text(
                              '36601175',
                              style: TextStyle(
                                fontFamily: 'Plus Jakarta Sans',
                                color: Colors.white,
                                fontWeight: FontWeight.bold,
                                fontSize: 14,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 12,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFF0081FF).withOpacity(0.1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(
                            Icons.inventory_2_outlined,
                            color: Color(0xFF0081FF),
                            size: 18,
                          ),
                          const SizedBox(width: 8),
                          Flexible(
                            child: Text(
                              (auction.lotNumber.isNotEmpty &&
                                      auction.lotNumber != 'N/A')
                                  ? auction.lotNumber
                                  : 'LOT-${auction.id.substring(0, 6).toUpperCase()}',
                              style: const TextStyle(
                                fontFamily: 'Plus Jakarta Sans',
                                color: Color(0xFF0081FF),
                                fontWeight: FontWeight.bold,
                                fontSize: 13,
                              ),
                              overflow: TextOverflow.ellipsis,
                              maxLines: 1,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildTimerUnit(String label, String value) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 24,
            fontWeight: FontWeight.bold,
            color: Colors.red,
          ),
        ),
        Text(
          label,
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 12,
            color: Colors.red,
          ),
        ),
      ],
    );
  }

  Widget _buildExternalLinks(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        children: [
          _buildListTile(
            Icons.info_outline,
            AppLocalizations.of(context)!.text_62,
            () => _showTermsBottomSheet(),
            isDarkMode,
          ),
          const SizedBox(height: 12),
          _buildListTile(
            Icons.play_circle_outline,
            'كيفية المزايدة والشحن',
            () {
              Navigator.of(
                context,
              ).push(MaterialPageRoute(builder: (_) => const HowToBidPage()));
            },
            isDarkMode,
          ),
          const SizedBox(height: 12),
          GestureDetector(
            onTap: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => AuctionHistoryPage(auctionId: auction.id),
              ),
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
                style: TextStyle(
                  fontFamily: 'Plus Jakarta Sans',
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

  Widget _buildCarDetailsSection(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    if (auction.manufacturer == null &&
        auction.fuelType == null &&
        auction.transmission == null &&
        auction.year == null &&
        auction.mileage == null &&
        auction.model == null) {
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
              if (auction.manufacturer != null)
                _buildCarDetailItem(
                  Icons.home_work_outlined,
                  AppLocalizations.of(context)!.text_65,
                  auction.manufacturer!,
                  isDarkMode,
                ),
              if (auction.transmission != null)
                _buildCarDetailItem(
                  Icons.vertical_split_outlined,
                  AppLocalizations.of(context)!.text_66,
                  auction.transmission!,
                  isDarkMode,
                ),
              if (auction.fuelType != null)
                _buildCarDetailItem(
                  Icons.local_gas_station_outlined,
                  AppLocalizations.of(context)!.text_67,
                  auction.fuelType!,
                  isDarkMode,
                ),
              if (auction.year != null)
                _buildCarDetailItem(
                  Icons.calendar_month_outlined,
                  AppLocalizations.of(context)!.text_68,
                  auction.year!,
                  isDarkMode,
                ),
              if (auction.mileage != null)
                _buildCarDetailItem(
                  Icons.speed_rounded,
                  AppLocalizations.of(context)!.text_69,
                  auction.mileage!,
                  isDarkMode,
                ),
              if (auction.model != null)
                _buildCarDetailItem(
                  Icons.directions_car_filled,
                  AppLocalizations.of(context)!.text_70,
                  auction.model!,
                  isDarkMode,
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildItemPropertiesSection(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    final Map<String, String> props = {};

    if (auction.condition != null && auction.condition!.isNotEmpty) {
      props['الحالة'] = _localizeCondition(auction.condition!);
    }
    if (auction.brand != null && auction.brand!.isNotEmpty) {
      props['الماركة'] = auction.brand!;
    }
    if (auction.category != null && auction.category!.isNotEmpty) {
      props['الفئة'] = auction.category!;
    }
    if (auction.city != null && auction.city!.isNotEmpty) {
      props['المدينة'] = auction.city!;
    }
    if (auction.lotNumber.isNotEmpty && auction.lotNumber != 'N/A') {
      props['رقم LOT'] = auction.lotNumber;
    }

    // Flatten item_details map
    if (auction.itemDetails != null) {
      for (final entry in auction.itemDetails!.entries) {
        final v = entry.value?.toString() ?? '';
        if (v.isNotEmpty && v != 'null') {
          props[entry.key] = v;
        }
      }
    }

    if (props.isEmpty) return const SizedBox.shrink();

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
          Text(
            'خصائص العنصر',
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 15,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black87,
            ),
          ),
          const SizedBox(height: 16),
          ...props.entries.map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SizedBox(
                    width: 90,
                    child: Text(
                      e.key,
                      style: TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 13,
                        color: isDarkMode ? Colors.grey[400] : Colors.grey[600],
                      ),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      e.value,
                      style: TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: isDarkMode ? Colors.white : Colors.black87,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _localizeCondition(String condition) {
    switch (condition.toLowerCase()) {
      case 'new':
        return 'جديد';
      case 'used':
        return 'مستعمل';
      case 'refurbished':
        return 'مجدد';
      case 'damaged':
        return 'تالف';
      default:
        return condition;
    }
  }

  Widget _buildCarDetailItem(
    IconData icon,
    String label,
    String value,
    bool isDarkMode,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(
              icon,
              size: 20,
              color: isDarkMode ? Colors.white70 : Colors.black87,
            ),
            const SizedBox(width: 6),
            Text(
              label,
              style: TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                fontSize: 13,
                color: Colors.grey,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        Text(
          value,
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 15,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }

  Widget _buildListTile(
    IconData icon,
    String title,
    VoidCallback onTap,
    bool isDarkMode,
  ) {
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
        title: Text(
          title,
          style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 14),
        ),
        trailing: const Icon(
          Icons.arrow_back_ios,
          size: 16,
          color: Colors.grey,
        ),
        onTap: onTap,
      ),
    );
  }

  Widget _buildBidHistorySection(
    BuildContext context,
    bool isDarkMode,
    Auction auction,
  ) {
    final bidderCount = auction.bidderCount;
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
              separatorBuilder: (context, index) =>
                  Divider(color: Colors.grey.withValues(alpha: 0.1), height: 1),
              itemBuilder: (context, index) {
                final bid = bidList[index];
                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 20,
                        backgroundColor: isDarkMode
                            ? const Color(0xFF333333)
                            : const Color(0xFFF5F5F5),
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
                  '$bidderCount ${AppLocalizations.of(context)!.text_79}',
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
                      builder: (context) =>
                          AuctionHistoryPage(auctionId: widget.auctionId),
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

  Widget _buildActionButtons(
    BuildContext context,
    bool isDarkMode,
    Auction auction,
  ) {
    final sellerId = auction.sellerId;
    final isOwner = _userId != null && _userId == sellerId;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Column(
        children: [
          _buildActionRow([
            _buildActionButton(
              context,
              icon: Icons.report_gmailerrorred_outlined,
              label: AppLocalizations.of(context)!.text_395,
              color: Colors.red,
              onTap: () => _handleReport(),
            ),
          ]),
          if (!isOwner) ...[const SizedBox(height: 12)],
        ],
      ),
    );
  }

  Widget _buildActionRow(List<Widget> children) {
    return Row(
      children:
          children
              .expand((w) => [Expanded(child: w), const SizedBox(width: 12)])
              .toList()
            ..removeLast(),
    );
  }

  Widget _buildActionButton(
    BuildContext context, {
    required IconData icon,
    required String label,
    required Color color,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: color.withValues(alpha: 0.3)),
        ),
        child: Column(
          children: [
            Icon(icon, color: color, size: 24),
            const SizedBox(height: 4),
            Text(
              label,
              style: TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: color,
              ),
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _handleAutoBid() async {
    final auctionAsync = ref.read(auctionNotifierApiProvider(widget.auctionId));
    final auction = auctionAsync.value;
    if (auction == null) return;

    final currentPrice = auction.currentPrice;

    AppModals.showAutoBidModal(
      context,
      auctionId: widget.auctionId,
      currentPrice: currentPrice,
      currencyCode: auction.currencyCode,
    );
  }

  Future<void> _handleReport() async {
    AppModals.showReportModal(context, auctionId: widget.auctionId);
  }

  Widget _buildBottomAction(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    final auctionId = auction.id;
    final currentPrice = auction.currentPrice;
    final minIncrement = auction.minIncrement;
    final bidCount = auction.bidderCount;
    final timeLeftStr = TimeUtils.formatDuration(context, _timeLeft);

    // صاحب المزاد لا يمكنه المزايدة
    final isOwner = _userId != null && _userId == auction.sellerId;

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, -5),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: SizedBox(
          width: double.infinity,
          height: 60,
          child: isOwner
              // ── صاحب المزاد: رسالة بدل زر المزايدة ──
              ? Container(
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.block, color: Colors.grey, size: 20),
                      SizedBox(width: 8),
                      Text(
                        'لا يمكنك المزايدة على مزادك الخاص',
                        style: TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          color: Colors.grey,
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                )
              // ── مستخدم عادي: زر المزايدة ──
              : ElevatedButton(
                  onPressed: auction.isUserHighestBidder
                      ? null
                      : () {
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
                              currencyCode: auction.currencyCode,
                            ),
                          );
                        },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: auction.isUserHighestBidder
                        ? const Color(0xFF00C58D)
                        : const Color(0xFF0081FF),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                    disabledBackgroundColor: const Color(0xFF00C58D),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      if (auction.isUserHighestBidder) ...[
                        const Icon(
                          Icons.back_hand_outlined,
                          color: Colors.white,
                        ),
                        const SizedBox(width: 8),
                        Flexible(
                          child: Text(
                            AppLocalizations.of(context)!.text_71,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              fontFamily: 'Plus Jakarta Sans',
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                            ),
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
                            style: const TextStyle(
                              fontFamily: 'Plus Jakarta Sans',
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                            ),
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
                  // "View All" (عرض الكل) next to "Similar Auctions" now opens
                  // the real Active Auctions listing page (with category
                  // filters + search), instead of bouncing back to Home
                  // (client feedback A8).
                  Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (context) => const AllAuctionsPage(),
                    ),
                  );
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 8,
                  ),
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

              List<Auction> relatedAuctions = [];
              final data = snapshot.data as dynamic;
              if (snapshot.hasData &&
                  data?.success == true &&
                  data?.data != null) {
                final dynamic responseData = data.data!;
                List<dynamic> auctionList = [];
                if (responseData is List) {
                  auctionList = responseData;
                } else if (responseData is Map<String, dynamic>) {
                  auctionList =
                      (responseData['auctions'] ?? responseData['data'] ?? [])
                          as List<dynamic>;
                }
                relatedAuctions = auctionList
                    .map(
                      (item) => Auction.fromJson(item as Map<String, dynamic>),
                    )
                    .toList();
              }

              if (relatedAuctions.isEmpty) {
                return Center(
                  child: Text(
                    AppLocalizations.of(context)!.no_related_auctions,
                    style: const TextStyle(color: Colors.grey),
                  ),
                );
              }

              return ListView.builder(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: relatedAuctions.length,
                itemBuilder: (context, index) {
                  final auction = relatedAuctions[index];
                  if (auction.id == widget.auctionId)
                    return const SizedBox.shrink();
                  return _buildRelatedAuctionCard(
                    context,
                    auction,
                    isDarkMode,
                    index + 1,
                  );
                },
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildRelatedAuctionCard(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
    int displayNumber,
  ) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(auction.id) ?? false;
    final imagePath = auction.imageUrls.isNotEmpty
        ? auction.imageUrls[0]
        : 'assets/corolla.png';

    return GestureDetector(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => AuctionDetailsPage(auctionId: auction.id),
          ),
        );
      },
      child: Container(
        width: 140,
        margin: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isDarkMode
                ? const Color(0xFF333333)
                : const Color(0xFFD0D5DD),
            width: 1.5,
          ),
          boxShadow: [
            BoxShadow(
              color: isDarkMode
                  ? Colors.black.withOpacity(0.3)
                  : const Color(0xFF101828).withValues(alpha: 0.08),
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
                  borderRadius: const BorderRadius.vertical(
                    top: Radius.circular(16),
                  ),
                  child: imagePath.startsWith('http')
                      ? CachedNetworkImage(
                          imageUrl: imagePath,
                          height: 120,
                          width: double.infinity,
                          fit: BoxFit.cover,
                          placeholder: (context, url) => Shimmer.fromColors(
                            baseColor: Colors.grey[300]!,
                            highlightColor: Colors.grey[100]!,
                            child: Container(color: Colors.white),
                          ),
                          errorWidget: (c, e, s) => Image.asset(
                            'assets/corolla.png',
                            height: 120,
                            width: double.infinity,
                            fit: BoxFit.cover,
                          ),
                        )
                      : Image.asset(
                          imagePath,
                          height: 120,
                          width: double.infinity,
                          fit: BoxFit.cover,
                          errorBuilder: (c, e, s) =>
                              Container(height: 120, color: Colors.grey[300]),
                        ),
                ),
                Positioned(
                  top: 4,
                  right: 4,
                  child: IconButton(
                    iconSize: 18,
                    onPressed: () {
                      ref
                          .read(favoritesProvider.notifier)
                          .toggleFavorite(auction.id);
                    },
                    icon: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.5),
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
                  top: 8,
                  left: 8,
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFF0084FF),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      '#$displayNumber',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
              ],
            ),
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    auction.title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 11,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(
                        Icons.timer_outlined,
                        size: 12,
                        color: Colors.red,
                      ),
                      const SizedBox(width: 3),
                      Text(
                        TimeUtils.formatDuration(
                          context,
                          auction.endTime.difference(DateTime.now()),
                        ),
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.red,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 2),
                  Text(
                    MoneyFormatter.format(auction.currentPrice, auction.currencyCode),
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 10,
                      color: Color(0xFF0081FF),
                    ),
                  ),
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
}
