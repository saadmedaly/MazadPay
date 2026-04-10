import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../providers/auction_provider.dart';
import '../models/auction.dart';

class AuctionWinnerPage extends ConsumerWidget {
  final String auctionId;
  const AuctionWinnerPage({super.key, required this.auctionId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auction = ref.watch(auctionNotifierProvider(auctionId));
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: Stack(
          children: [
            // Fireworks Background (Simulated with icons)
            _buildFireworksDecoration(),
            
            SafeArea(
              child: SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                padding: const EdgeInsets.only(bottom: 24),
                child: Column(
                  children: [
                    _buildHeader(context),
                    const SizedBox(height: 20),
                    Text(
                      'مبروك!',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 40,
                        fontWeight: FontWeight.w900,
                        color: const Color(0xFF135BEC),
                      ),
                    ),
                    Text(
                      'ربحت المزاد',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 24,
                        fontWeight: FontWeight.bold,
                        color: isDarkMode ? Colors.white : Colors.black87,
                      ),
                    ),
                    const SizedBox(height: 40),
                    
                    _buildWinningAmountBox(auction, isDarkMode),
                    const SizedBox(height: 32),
                    
                    _buildProductCard(auction, isDarkMode),
                    const SizedBox(height: 32),
                    
                    _buildWinnerSummary(isDarkMode),
                    
                    const SizedBox(height: 40),
                    _buildFooterAction(context, isDarkMode),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFireworksDecoration() {
     return Stack(
       children: [
         Positioned(top: 100, left: 50, child: Icon(Icons.star, color: Colors.orange.withOpacity(0.3), size: 40)),
         Positioned(top: 150, right: 80, child: Icon(Icons.auto_awesome, color: Colors.blue.withOpacity(0.3), size: 50)),
         Positioned(top: 300, left: 30, child: Icon(Icons.favorite, color: Colors.red.withOpacity(0.2), size: 30)),
         Positioned(bottom: 200, right: 40, child: Icon(Icons.wb_sunny, color: Colors.yellow.withOpacity(0.3), size: 60)),
       ],
     );
  }

  Widget _buildHeader(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          IconButton(
            icon: const Icon(Icons.share_outlined, color: Color(0xFF135BEC)),
            onPressed: () {},
          ),
          IconButton(
            icon: const Icon(Icons.close, color: Colors.grey),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
  }

  Widget _buildWinningAmountBox(Auction auction, bool isDarkMode) {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
          decoration: const BoxDecoration(
            color: Color(0xFFE31B23),
            borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
          ),
          child: Text(
            '2026/02/15',
            style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold),
          ),
        ),
        Container(
          width: 300,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: const Color(0xFFFFCC00),
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 20, offset: const Offset(0, 10)),
            ],
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
               const Icon(Icons.check_circle, color: Color(0xFF00C58D), size: 32),
               const SizedBox(width: 12),
               Text(
                 '${auction.currentPrice.toStringAsFixed(0)} أوقية',
                 style: GoogleFonts.plusJakartaSans(fontSize: 28, fontWeight: FontWeight.w900, color: Colors.black),
               ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildProductCard(Auction auction, bool isDarkMode) {
     return Container(
       padding: const EdgeInsets.all(8),
       decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
          ],
       ),
       child: ClipRRect(
         borderRadius: BorderRadius.circular(12),
         child: Image.asset(auction.imageUrls[0], width: 300, height: 180, fit: BoxFit.cover),
       ),
     );
  }

  Widget _buildWinnerSummary(bool isDarkMode) {
    return Container(
      width: 350,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: const Color(0xFF135BEC).withOpacity(0.1)),
      ),
      child: Row(
        children: [
           Container(
             width: 60, height: 60,
             decoration: BoxDecoration(
               color: const Color(0xFFFFF7E6),
               borderRadius: BorderRadius.circular(12),
             ),
             child: const Center(child: Icon(Icons.emoji_events, color: Color(0xFFFFCC00), size: 40)),
           ),
           const SizedBox(width: 16),
           Column(
             crossAxisAlignment: CrossAxisAlignment.start,
             children: [
                Text('محمد احمد سيديا', style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold)),
                Text('الفائز الاول بالمزاد', style: GoogleFonts.plusJakartaSans(color: Colors.grey)),
             ],
           ),
        ],
      ),
    );
  }

  Widget _buildFooterAction(BuildContext context, bool isDarkMode) {
    return Padding(
      padding: const EdgeInsets.all(24.0),
      child: SizedBox(
        width: double.infinity,
        height: 60,
        child: ElevatedButton(
          onPressed: () {
            // Future payment gateway integration
          },
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF4A7DFF),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          ),
          child: Text(
            'أكمل عملية الدفع',
            style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
          ),
        ),
      ),
    );
  }
}
