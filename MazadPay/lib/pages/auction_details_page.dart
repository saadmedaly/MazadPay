import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../providers/auction_provider.dart';
import '../models/auction.dart';
import '../widgets/bid_action_sheet.dart';
import '../widgets/auction_winner_dialog.dart';
import '../pages/auction_history_page.dart';
import '../providers/favorites_provider.dart';
import '../pages/all_auctions_page.dart';

class AuctionDetailsPage extends ConsumerStatefulWidget {
  final String auctionId;
  const AuctionDetailsPage({super.key, required this.auctionId});

  @override
  ConsumerState<AuctionDetailsPage> createState() => _AuctionDetailsPageState();
}

class _AuctionDetailsPageState extends ConsumerState<AuctionDetailsPage> {
  late Timer _timer;
  Duration _timeLeft = Duration.zero;
  final PageController _pageController = PageController();
  int _currentPage = 0;

  @override
  void initState() {
    super.initState();
    _startTimer();
  }

  void _startTimer() {
    final auction = ref.read(auctionNotifierProvider(widget.auctionId));
    _timeLeft = auction.endTime.difference(DateTime.now());
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_timeLeft.inSeconds > 0) {
        setState(() {
          _timeLeft = Duration(seconds: _timeLeft.inSeconds - 1);
        });
      } else {
        _timer.cancel();
        _checkWinner();
      }
    });
  }

  void _checkWinner() {
    final auction = ref.read(auctionNotifierProvider(widget.auctionId));
    if (auction.isUserHighestBidder) {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (context) => AuctionWinnerDialog(auctionId: widget.auctionId),
      );
    }
  }

  @override
  void dispose() {
    _timer.cancel();
    super.dispose();
  }

  String _formatDuration(Duration d) {
    String twoDigits(int n) => n.toString().padLeft(2, "0");
    String hours = twoDigits(d.inHours);
    String minutes = twoDigits(d.inMinutes.remainder(60));
    String seconds = twoDigits(d.inSeconds.remainder(60));
    return "$hours : $minutes : $seconds";
  }

  @override
  Widget build(BuildContext context) {
    final auction = ref.watch(auctionNotifierProvider(widget.auctionId));
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: Column(
          children: [
            Expanded(
              child: CustomScrollView(
                slivers: [
                  _buildSliverAppBar(context, auction, isDarkMode),
                  SliverToBoxAdapter(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildPriceSection(context, auction, isDarkMode),
                        _buildMainInfoSection(context, auction, isDarkMode),
                        _buildExternalLinks(context, auction, isDarkMode),
                        _buildCarDetailsSection(context, auction, isDarkMode),

                        // Active Auctions Section (NEW)
                        _buildActiveAuctionsSection(context, isDarkMode),
                        
                        const SizedBox(height: 100), // Space for bottom button
                      ],
                    ),
                  ),
                ],
              ),
            ),
            _buildBottomAction(context, auction, isDarkMode),
          ],
        ),
      );
  }

  Widget _buildSliverAppBar(BuildContext context, Auction auction, bool isDarkMode) {
    final favorites = ref.watch(favoritesProvider);
    final isFavorite = favorites.contains(widget.auctionId);

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
              itemCount: auction.imageUrls.length,
              itemBuilder: (context, index) => Image.asset(auction.imageUrls[index], fit: BoxFit.cover),
            ),
            Positioned(
              bottom: 24,
              end: 24,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.6),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '${_currentPage + 1}/${auction.imageUrls.length} 🗂️',
                  style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold),
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

  Widget _buildPriceSection(BuildContext context, Auction auction, bool isDarkMode) {
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
                Text(AppLocalizations.of(context)!.text_57, style: GoogleFonts.plusJakartaSans(color: Colors.white70, fontSize: 14)),
                const SizedBox(height: 4),
                Text('${auction.currentPrice.toStringAsFixed(0)} أوقية جديدة', 
                    style: GoogleFonts.plusJakartaSans(color: Colors.white, fontSize: 20, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
        ),
        Expanded(
          child: Container(
            height: 100,
            decoration: BoxDecoration(
              color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
              border: Border(start: BorderSide(color: Colors.grey.withOpacity(0.1))),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(AppLocalizations.of(context)!.text_58, style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 14)),
                const SizedBox(height: 4),
                Text('${auction.minIncrement.toStringAsFixed(0)} أوقية جديدة', 
                    style: GoogleFonts.plusJakartaSans(color: const Color(0xFF0081FF), fontSize: 20, fontWeight: FontWeight.bold)),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildMainInfoSection(BuildContext context, Auction auction, bool isDarkMode) {
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
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Text(
                  auction.title,
                  style: GoogleFonts.plusJakartaSans(fontSize: 22, fontWeight: FontWeight.bold),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: 8),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text('${auction.views}', style: GoogleFonts.plusJakartaSans(color: Colors.grey)),
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
                _buildTimerUnit(AppLocalizations.of(context)!.text_59, _formatDuration(_timeLeft).split(' : ')[2]),
                _buildTimerUnit(AppLocalizations.of(context)!.text_60, _formatDuration(_timeLeft).split(' : ')[1]),
                _buildTimerUnit(AppLocalizations.of(context)!.text_61, _formatDuration(_timeLeft).split(' : ')[0]),
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
                    MaterialPageRoute(builder: (context) => AuctionHistoryPage(auctionId: auction.id)),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.gavel, color: Colors.black, size: 18),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'عدد المزايدات ${auction.bidderCount}', 
                          style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold),
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
                    Text(auction.phoneNumber, style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 8),
                    const Icon(Icons.phone, color: Colors.white, size: 18),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text('lot #${auction.lotNumber}', style: GoogleFonts.plusJakartaSans(color: const Color(0xFF0081FF), fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }

  Widget _buildTimerUnit(String label, String value) {
    return Column(
      children: [
        Text(value, style: GoogleFonts.plusJakartaSans(fontSize: 24, fontWeight: FontWeight.bold, color: Colors.red)),
        Text(label, style: GoogleFonts.plusJakartaSans(fontSize: 12, color: Colors.red)),
      ],
    );
  }

  Widget _buildExternalLinks(BuildContext context, Auction auction, bool isDarkMode) {
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
              MaterialPageRoute(builder: (context) => AuctionHistoryPage(auctionId: auction.id)),
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
                style: GoogleFonts.plusJakartaSans(
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

  Widget _buildCarDetailsSection(BuildContext context, Auction auction, bool isDarkMode) {
    if (auction.manufacturer == null && auction.fuelType == null &&
        auction.transmission == null && auction.year == null &&
        auction.mileage == null && auction.model == null) {
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
                _buildCarDetailItem(Icons.home_work_outlined, AppLocalizations.of(context)!.text_65, auction.manufacturer!, isDarkMode),
              if (auction.transmission != null)
                _buildCarDetailItem(Icons.vertical_split_outlined, AppLocalizations.of(context)!.text_66, auction.transmission!, isDarkMode),
              if (auction.fuelType != null)
                _buildCarDetailItem(Icons.local_gas_station_outlined, AppLocalizations.of(context)!.text_67, auction.fuelType!, isDarkMode),
              if (auction.year != null)
                _buildCarDetailItem(Icons.calendar_month_outlined, AppLocalizations.of(context)!.text_68, auction.year!, isDarkMode),
              if (auction.mileage != null)
                _buildCarDetailItem(Icons.speed_rounded, AppLocalizations.of(context)!.text_69, auction.mileage!, isDarkMode),
              if (auction.model != null)
                _buildCarDetailItem(Icons.directions_car_filled, AppLocalizations.of(context)!.text_70, auction.model!, isDarkMode),
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
            Text(label, style: GoogleFonts.plusJakartaSans(fontSize: 13, color: Colors.grey, fontWeight: FontWeight.w500)),
          ],
        ),
        const SizedBox(height: 6),
        Text(value, style: GoogleFonts.plusJakartaSans(fontSize: 15, fontWeight: FontWeight.w700)),
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
        title: Text(title, style: GoogleFonts.plusJakartaSans(fontSize: 14)),
        trailing: const Icon(Icons.arrow_back_ios, size: 16, color: Colors.grey),
        onTap: onTap,
      ),
    );
  }

  Widget _buildBottomAction(BuildContext context, Auction auction, bool isDarkMode) {
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
            onPressed: auction.isUserHighestBidder ? null : () {
              showModalBottomSheet(
                context: context,
                backgroundColor: Colors.transparent,
                isScrollControlled: true,
                builder: (context) => BidActionSheet(auctionId: auction.id),
              );
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: auction.isUserHighestBidder ? const Color(0xFF00C58D) : const Color(0xFF0081FF),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
              disabledBackgroundColor: const Color(0xFF00C58D),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (auction.isUserHighestBidder) ...[
                   const Icon(Icons.back_hand_outlined, color: Colors.white),
                   const SizedBox(width: 8),
                   Text(AppLocalizations.of(context)!.text_71, style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold)),
                ] else ...[
                   const Icon(Icons.gavel, color: Colors.white),
                   const SizedBox(width: 8),
                   Text(AppLocalizations.of(context)!.text_72, style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold)),
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
                style: GoogleFonts.plusJakartaSans(
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
                    style: GoogleFonts.plusJakartaSans(
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
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            children: [
              _buildRelatedAuctionCard('Toyota Corolla...', '300,000 MRU', '13:50:23', 'assets/corolla.png', '1', isDarkMode),
              _buildRelatedAuctionCard('Range Rover 2022', '1,200,000 MRU', '02:15:10', 'assets/corolla.png', '2', isDarkMode),
              _buildRelatedAuctionCard('iPhone 15 Pro', '45,000 MRU', '05:45:00', 'assets/corolla.png', '3', isDarkMode),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildRelatedAuctionCard(String title, String price, String time, String imagePath, String id, bool isDarkMode) {
    final favorites = ref.watch(favoritesProvider);
    final isFavorite = favorites.contains(id);

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
                  child: Image.asset(imagePath, height: 120, width: double.infinity, fit: BoxFit.cover),
                ),
                Positioned(
                  top: 4, end: 4,
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
                  top: 8, start: 8,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: const Color(0xFF0084FF),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(id, style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
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
}
