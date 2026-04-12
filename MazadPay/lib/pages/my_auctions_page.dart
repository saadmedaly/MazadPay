import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

import 'auction_details_page.dart';

class MyAuctionsPage extends StatelessWidget {
  const MyAuctionsPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            AppLocalizations.of(context)!.text_27,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
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
        body: ListView.separated(
          padding: const EdgeInsets.all(20),
          itemCount: 3,
          separatorBuilder: (context, index) => const SizedBox(height: 16),
          itemBuilder: (context, index) {
            final isWinning = index == 0;
            return _buildAuctionItem(context, isWinning, isDarkMode);
          },
        ),
      );
  }

  Widget _buildAuctionItem(BuildContext context, bool isWinning, bool isDarkMode) {
    return Container(
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
            child: Image.asset('assets/corolla.png', width: 100, height: 100, fit: BoxFit.cover),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Toyota Corolla 2018',
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 14, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 8),
                Text(
                  '307,000 MRU',
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: const Color(0xFF0081FF),
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Flexible(
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                        decoration: BoxDecoration(
                          color: isWinning ? const Color(0xFF00C58D).withOpacity(0.1) : const Color(0xFFE31B23).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          isWinning ? AppLocalizations.of(context)!.text_234 : AppLocalizations.of(context)!.text_235,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                            fontSize: 10,
                            fontWeight: FontWeight.bold,
                            color: isWinning ? const Color(0xFF00C58D) : const Color(0xFFE31B23),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '13:45:10',
                      style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 12, color: Colors.grey, fontWeight: FontWeight.bold),
                    ),
                  ],
                ),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(Icons.arrow_forward_ios, size: 16, color: Colors.grey),
            onPressed: () {
               Navigator.of(context).push(
                MaterialPageRoute(builder: (context) => const AuctionDetailsPage(auctionId: '1')),
              );
            },
          ),
        ],
      ),
    );
  }
}
