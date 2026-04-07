import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../providers/auction_provider.dart';

class BidActionSheet extends ConsumerStatefulWidget {
  final String auctionId;
  const BidActionSheet({super.key, required this.auctionId});

  @override
  ConsumerState<BidActionSheet> createState() => _BidActionSheetState();
}

class _BidActionSheetState extends ConsumerState<BidActionSheet> {
  int _step = 1; // 1: Increase amount, 2: Final confirm
  double _bidAmount = 0.0;

  @override
  void initState() {
    super.initState();
    final auction = ref.read(auctionNotifierProvider(widget.auctionId));
    _bidAmount = auction.currentPrice + auction.minIncrement;
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[300], borderRadius: BorderRadius.circular(2))),
          const SizedBox(height: 24),
          _step == 1 ? _buildStep1(isDarkMode) : _buildStep2(isDarkMode),
          const SizedBox(height: 32),
          _buildActionButton(isDarkMode),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Widget _buildStep1(bool isDarkMode) {
    final auction = ref.read(auctionNotifierProvider(widget.auctionId));
    return Column(
      children: [
        Text('تاكيد المزايدة', style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold)),
        const SizedBox(height: 24),
        
        // Header info box
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: isDarkMode ? Colors.black26 : const Color(0xFFF9FAFB),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: Colors.grey.withOpacity(0.1)),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
               _buildInfoColumn('11', 'مزايدات', Colors.red),
               _buildVerticalDivider(),
               _buildInfoColumn('${auction.currentPrice.toStringAsFixed(0)}', 'أعلى مزايدة', Colors.black),
               _buildVerticalDivider(),
               _buildInfoColumn('09 : 00', 'الوقت المتبقي', Colors.red),
            ],
          ),
        ),
        
        const SizedBox(height: 32),
        
        // Amount selector
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          decoration: BoxDecoration(
            color: isDarkMode ? Colors.black26 : const Color(0xFFF2F4F7),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildIconButton(Icons.add, () => setState(() => _bidAmount += auction.minIncrement)),
              Expanded(
                child: Center(
                  child: Text(
                    '${_bidAmount.toStringAsFixed(0)} أوقية',
                    style: GoogleFonts.plusJakartaSans(fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
              _buildIconButton(Icons.remove, () {
                if (_bidAmount > auction.currentPrice + auction.minIncrement) {
                  setState(() => _bidAmount -= auction.minIncrement);
                }
              }),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildStep2(bool isDarkMode) {
    return Column(
      children: [
        Text('تاكيد المزايدة', style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold)),
        const SizedBox(height: 32),
        Text('${_bidAmount.toStringAsFixed(0)} أوقية', 
            style: GoogleFonts.plusJakartaSans(fontSize: 28, fontWeight: FontWeight.bold, color: const Color(0xFF0081FF))),
        const SizedBox(height: 16),
        Text('الموافقة على مبلغ المزايدة ؟', style: GoogleFonts.plusJakartaSans(fontSize: 16, color: Colors.grey)),
        const SizedBox(height: 32),
      ],
    );
  }

  Widget _buildActionButton(bool isDarkMode) {
    return SizedBox(
      width: double.infinity,
      height: 60,
      child: ElevatedButton(
        onPressed: () {
          if (_step == 1) {
            setState(() => _step = 2);
          } else {
            // Confirm Bid
            ref.read(auctionNotifierProvider(widget.auctionId).notifier).placeBid(_bidAmount);
            Navigator.of(context).pop();
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('تمت المزايدة بنجاح !')),
            );
          }
        },
        style: ElevatedButton.styleFrom(
          backgroundColor: const Color(0xFF0081FF),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.gavel, color: Colors.white),
            const SizedBox(width: 8),
            Text(_step == 1 ? 'قم بالمزايدة الان' : 'تاكيد', 
                style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold)),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoColumn(String value, String label, Color valueColor) {
    return Column(
      children: [
        Text(value, style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold, color: valueColor)),
        Text(label, style: GoogleFonts.plusJakartaSans(fontSize: 12, color: Colors.grey)),
      ],
    );
  }

  Widget _buildVerticalDivider() {
    return Container(width: 1, height: 40, color: Colors.grey.withOpacity(0.2));
  }

  Widget _buildIconButton(IconData icon, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 48,
        height: 48,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 5),
          ],
        ),
        child: Icon(icon, color: const Color(0xFF0081FF)),
      ),
    );
  }
}
