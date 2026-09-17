import 'dart:async';

import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart' hide TextDirection;
import 'package:url_launcher/url_launcher.dart';

import 'auction_winner_page.dart';
import 'support_page.dart' show SupportContactUris;
import '../services/auction_api.dart';
import '../services/cache_service.dart';
import '../services/realtime_sync_service.dart';
import '../utils/money_formatter.dart';
import '../utils/auction_image.dart';

// Customer Request #30: MazadPay's own official WhatsApp number (matches
// support_page.dart's canonical whatsappNumber -- the same value already
// confirmed there to agree with login_page.dart and app_modals.dart), used
// for winners to manually coordinate payment. Deliberately NOT the
// per-auction seller contact number (auction_details_page.dart's
// _supportPhone) -- this is company-to-winner, not buyer-to-seller.
const String kMazadPayWhatsAppNumber = '47601175';

/// Pure builder for the pay-button's WhatsApp message text, extracted so it
/// is directly unit-testable and shared between both launch URIs below.
/// Never includes JWT/UUIDs/internal IDs/email/private phone data -- only
/// the auction title, LOT number (if present), and the winning amount,
/// which are all already shown on-screen to the winner themselves.
String buildWinnerPaymentMessage({
  required String auctionTitle,
  required String? lotNumber,
  required String formattedAmount,
}) {
  final buffer = StringBuffer('مرحباً، لقد فزت بمزاد ');
  buffer.write('"$auctionTitle"');
  if (lotNumber != null && lotNumber.trim().isNotEmpty) {
    buffer.write(' (LOT-$lotNumber)');
  }
  buffer.write(' بمبلغ $formattedAmount. أرغب في إكمال عملية الدفع.');
  return buffer.toString();
}

/// Primary launch target: the universal https://wa.me/... link (works via
/// the WhatsApp app when installed, or the browser otherwise).
Uri buildWinnerPaymentWhatsAppUri({
  required String auctionTitle,
  required String? lotNumber,
  required String formattedAmount,
}) {
  final text = buildWinnerPaymentMessage(
    auctionTitle: auctionTitle,
    lotNumber: lotNumber,
    formattedAmount: formattedAmount,
  );
  final base = SupportContactUris.whatsApp(kMazadPayWhatsAppNumber);
  return base.replace(queryParameters: {'text': text});
}

/// REAL DEVICE BUG fallback: WhatsApp's own native `whatsapp://send` deep
/// link. On Android 11+ (targetSdk 30+), canLaunchUrl()/launchUrl() for the
/// https://wa.me/... link can fail even with WhatsApp installed unless the
/// host app declares package-visibility `<queries>` for that exact intent
/// (fixed in AndroidManifest.xml) -- this native-scheme URI is a second,
/// independent path to the same conversation/prefilled message, used only
/// if the primary wa.me launch fails. iOS/other platforms never reach this
/// path since the primary launch already succeeds there.
Uri buildWinnerPaymentWhatsAppNativeUri({
  required String auctionTitle,
  required String? lotNumber,
  required String formattedAmount,
}) {
  final text = buildWinnerPaymentMessage(
    auctionTitle: auctionTitle,
    lotNumber: lotNumber,
    formattedAmount: formattedAmount,
  );
  return Uri.parse('whatsapp://send').replace(queryParameters: {
    'phone': '222$kMazadPayWhatsAppNumber',
    'text': text,
  });
}

class MyWinningsPage extends ConsumerStatefulWidget {
  const MyWinningsPage({super.key});

  @override
  ConsumerState<MyWinningsPage> createState() => _MyWinningsPageState();
}

class _MyWinningsPageState extends ConsumerState<MyWinningsPage> with WidgetsBindingObserver {
  final AuctionApi _auctionApi = AuctionApi();
  List<Map<String, dynamic>> _winnings = [];
  bool _isLoading = true;
  String? _error;

  // Customer #23: reuses Customer #20's existing global-channel event
  // stream -- no new realtime architecture. auction.status_changed already
  // fires for every real auction closure (see
  // AuctionService.FinalizeExpiredAuction/emitAuctionEvent); this page just
  // needed to actually listen for it and refetch, which it never did
  // before.
  StreamSubscription<RealtimeEvent>? _realtimeSubscription;
  StreamSubscription<void>? _catchUpSubscription;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _loadWinnings();

    _realtimeSubscription = RealtimeSyncService().events.listen((event) {
      if (event.type == RealtimeEventType.auctionStatusChanged && mounted) {
        _loadWinnings();
      }
    });
    // App-resume refresh (Customer #20's existing catch-up signal): fires on
    // reconnect/resume, refetches current state -- never replays historical
    // wins as a new popup (this page only refreshes its own list, it never
    // opens AuctionWinnerPage automatically -- see notification_handler.dart
    // for the foreground winner coordinator that owns that behavior).
    _catchUpSubscription = RealtimeSyncService().catchUpSignal.listen((_) {
      if (mounted) _loadWinnings();
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && mounted) {
      _loadWinnings();
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _realtimeSubscription?.cancel();
    _catchUpSubscription?.cancel();
    super.dispose();
  }

  // Customer Request #30 / REAL DEVICE BUG fix: opens WhatsApp to MazadPay's
  // own official number with a safe prefilled message so the winner can
  // manually coordinate payment with the company.
  //
  // canLaunchUrl() is deliberately NOT used as a gate before launching: on
  // Android 11+ (targetSdk 30+, package visibility), it can return false for
  // an https://wa.me/... URI even when WhatsApp IS installed, unless the
  // host app declares <queries> visibility for that exact intent (now fixed
  // in AndroidManifest.xml) -- a real-device bug this exact pattern caused
  // ("تعذّر فتح واتساب" even with WhatsApp present). Instead this attempts
  // launchUrl directly (letting a genuine failure throw), then falls back to
  // WhatsApp's native whatsapp://send deep link before finally showing the
  // error snackbar -- never trusting canLaunchUrl's possibly-false result as
  // the sole signal. iOS/other platforms are unaffected (no package-
  // visibility restriction there) and simply succeed on the first attempt.
  Future<void> _openPaymentWhatsApp({
    required BuildContext context,
    required String auctionTitle,
    required String? lotNumber,
    required String formattedAmount,
  }) async {
    final primaryUri = buildWinnerPaymentWhatsAppUri(
      auctionTitle: auctionTitle,
      lotNumber: lotNumber,
      formattedAmount: formattedAmount,
    );
    final nativeUri = buildWinnerPaymentWhatsAppNativeUri(
      auctionTitle: auctionTitle,
      lotNumber: lotNumber,
      formattedAmount: formattedAmount,
    );
    try {
      final launched = await launchUrl(primaryUri, mode: LaunchMode.externalApplication);
      if (launched) return;
      // launchUrl returning false (no exception) is itself a "could not
      // launch" signal on some platforms -- try the native fallback before
      // giving up.
      final fallbackLaunched = await launchUrl(nativeUri, mode: LaunchMode.externalApplication);
      if (fallbackLaunched) return;
      if (!context.mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('تعذّر فتح واتساب. تأكد من تثبيته على جهازك.')),
      );
    } catch (_) {
      // The primary launch threw -- try the native whatsapp:// fallback
      // before surfacing an error.
      try {
        final fallbackLaunched = await launchUrl(nativeUri, mode: LaunchMode.externalApplication);
        if (fallbackLaunched) return;
      } catch (_) {
        // fall through to the error snackbar below
      }
      // context.mounted, not the State's own `mounted`: `context` is a
      // parameter here, not necessarily this State's own BuildContext, so
      // the analyzer correctly treats a bare `mounted` check as unrelated
      // to whether THIS specific context is still safe to use after the
      // await above.
      if (!context.mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('تعذّر فتح واتساب. تأكد من تثبيته على جهازك.')),
      );
    }
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

        // Cache mes gains (dans la forme Map attendue par CacheService)
        await CacheService.instance.cacheMyWinnings({'data': auctionList});

        if (!mounted) return;
        setState(() {
          _isLoading = false;
          _error = null;
          _winnings = auctionList.map((item) => item as Map<String, dynamic>).toList();
        });
      } else {
        // Bug L fix: the fetch genuinely failed (network/timeout/backend
        // error) -- previously this fell through silently with no error set,
        // rendering the misleading "no winnings yet" empty state even when a
        // real win existed server-side but simply couldn't be fetched right
        // now. Only trust the cache's absence as "genuinely no winnings" when
        // the API call actually succeeded; a failed call with no warm cache
        // must show the retry/error state, never a false empty state.
        if (cachedMyWinnings == null && mounted) {
          setState(() {
            _isLoading = false;
            _error = response.error?.message ?? response.message ?? 'fetch_failed';
          });
        }
      }
    } catch (e) {
      if (!mounted) return;
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
              // Customer feedback #11: a user who already had this page open
              // (or returns to it) when an auction they won just finalized
              // must be able to pull down to see the new win without leaving
              // and reopening the page -- pull-to-refresh added to both the
              // list and the empty state (RefreshIndicator requires a
              // scrollable child, so the empty state is wrapped in one).
              // AlwaysScrollableScrollPhysics is required here specifically:
              // the empty state's content is shorter than the viewport, and
              // default ScrollPhysics can refuse to register the overscroll
              // drag needed to trigger RefreshIndicator when there is nothing
              // to actually scroll.
              : _winnings.isEmpty
                  ? RefreshIndicator(
                      onRefresh: _loadWinnings,
                      color: const Color(0xFF0081FF),
                      child: ListView(
                        physics: const AlwaysScrollableScrollPhysics(),
                        children: [
                          SizedBox(
                            height: MediaQuery.of(context).size.height * 0.6,
                            child: Center(
                              child: Text(
                                l10n.no_winnings,
                                style: const TextStyle(color: Colors.grey, fontSize: 16),
                              ),
                            ),
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _loadWinnings,
                      color: const Color(0xFF0081FF),
                      child: ListView.separated(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.all(20),
                        itemCount: _winnings.length,
                        separatorBuilder: (context, index) => const SizedBox(height: 16),
                        itemBuilder: (context, index) {
                          final winning = _winnings[index];
                          return _buildWinningItem(context, winning, isDarkMode, l10n);
                        },
                      ),
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
    
    // Customer #23: current_price is the ONLY field the backend actually
    // sends for the settled winning amount (models.Auction has no separate
    // winning_price/final_price column) -- final_price/current_bid were
    // never real backend fields, just dead fallback branches that could
    // never fire. Kept as a defensive secondary fallback only in case a
    // legacy cached response shape still has them, per the implementation
    // brief ("if legacy fallback compatibility is demonstrably needed, keep
    // it only as a secondary compatibility fallback") -- current_price is
    // tried first.
    final price = MoneyFormatter.format(
      num.tryParse((winning['current_price'] ?? winning['final_price'] ?? winning['current_bid'] ?? 0).toString()) ?? 0,
      winning['currency_code']?.toString(),
    );
    final isPaid = winning['is_paid'] == true || winning['payment_status'] == 'paid';

    // Customer #23: end_time IS returned by the backend (models.Auction's
    // json:"end_time") but was never rendered on this card at all.
    final endTimeRaw = winning['end_time']?.toString();
    final endDate = endTimeRaw != null ? DateTime.tryParse(endTimeRaw) : null;
    final endDateLabel = endDate != null ? DateFormat('yyyy-MM-dd').format(endDate.toLocal()) : null;

    // Gestion des images. Customer feedback #11: /users/me/winnings returns
    // auctions via AuctionRepository.ListPaginated, whose image_urls field is a
    // comma-separated STRING (see models.Auction.ImageURLs), not a List -- the
    // "is List" check below always failed for it, so every won item silently
    // fell through to the generic placeholder asset even when real images
    // existed. Mirrors the same comma-separated-string-or-List handling
    // already established in my_auctions_page.dart for this identical field.
    //
    // Client feedback: Bug H -- the placeholder used to default to
    // 'assets/corolla.png' (a real, bundled stock photo), which rendered
    // successfully as though it were the auction's own image whenever no
    // real image existed. Defaults to '' instead, so the no-image case
    // reaches the same neutral Icons.image_not_supported placeholder below
    // as a genuinely broken/failed image load.
    String imageUrl = '';
    final rawImageUrls = winning['image_urls'];
    if (rawImageUrls != null && rawImageUrls.toString().isNotEmpty) {
      if (rawImageUrls is List && rawImageUrls.isNotEmpty) {
        imageUrl = resolveAuctionImageUrl(rawImageUrls.map((e) => e.toString()).toList()) ?? '';
      } else {
        imageUrl = resolveAuctionImageUrl(rawImageUrls.toString().split(',')) ?? '';
      }
    } else if (winning['images'] != null && winning['images'] is List && (winning['images'] as List).isNotEmpty) {
      imageUrl = resolveAuctionImageUrl((winning['images'] as List).map((e) => e.toString()).toList()) ?? '';
    } else if (winning['image_url'] != null) {
      imageUrl = winning['image_url'].toString();
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
                child: imageUrl.isEmpty
                    ? Container(
                        width: 80,
                        height: 80,
                        color: Colors.grey[200],
                        child: const Icon(Icons.image_not_supported, color: Colors.grey),
                      )
                    : imageUrl.startsWith('http')
                    ? Image.network(
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
                      )
                    : Image.asset(
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
                    if (endDateLabel != null) ...[
                      const SizedBox(height: 2),
                      Text(
                        endDateLabel,
                        style: TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          fontSize: 11,
                          color: isDarkMode ? Colors.grey[400] : Colors.grey[600],
                        ),
                      ),
                    ],
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
                    onPressed: () => _openPaymentWhatsApp(
                      context: context,
                      auctionTitle: title,
                      lotNumber: winning['lot_number']?.toString(),
                      formattedAmount: price,
                    ),
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
