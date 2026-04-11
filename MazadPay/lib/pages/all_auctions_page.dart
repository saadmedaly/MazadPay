import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'auction_details_page.dart';

class AllAuctionsPage extends ConsumerStatefulWidget {
  const AllAuctionsPage({super.key});

  @override
  ConsumerState<AllAuctionsPage> createState() => _AllAuctionsPageState();
}

class _AllAuctionsPageState extends ConsumerState<AllAuctionsPage> {
  final TextEditingController _searchController = TextEditingController();
  
  List<Map<String, String>> _allAuctions = [];
  List<Map<String, String>> _filteredAuctions = [];
  int _activeTabIndex = 0; // 0 for Active, 1 for Finished

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final l10n = AppLocalizations.of(context)!;
    _allAuctions = [
      {
        'id': '1',
        'title': 'Toyota prado tx 2017',
        'price': '15,000 MRU',
        'bids': '15',
        'time': '23h 30m 18s',
        'location': l10n.text_171, // "تفرغ زينة"
        'postedTime': 'منذ 23 دقائق',
        'image': 'assets/raf4.png',
        'status': 'active'
      },
      {
        'id': '2',
        'title': 'هلكس خليجية نظيفة',
        'price': '760,000 MRU',
        'bids': '7',
        'time': '23h 30m 18s',
        'location': l10n.text_172, // "عرفات"
        'postedTime': 'منذ 23 دقائق',
        'image': 'assets/corolla.png',
        'status': 'active'
      },
      {
        'id': '3',
        'title': 'TOYOTA RAV4 2008 D4D',
        'price': '420,000 MRU',
        'bids': '20',
        'time': l10n.text_364, // "انتهى المزاد"
        'location': l10n.text_171,
        'postedTime': 'منذ 24 ساعة',
        'image': 'assets/raf4.png',
        'status': 'finished'
      },
    ];
    _filterAuctions();
  }

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged() {
    _filterAuctions();
  }

  void _filterAuctions() {
    setState(() {
      _filteredAuctions = _allAuctions.where((auction) {
        final matchesSearch = auction['title']!.toLowerCase().contains(_searchController.text.toLowerCase());
        final matchesStatus = _activeTabIndex == 0 ? auction['status'] == 'active' : auction['status'] == 'finished';
        return matchesSearch && matchesStatus;
      }).toList();
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);
    
    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: primaryBlue,
          elevation: 0,
          centerTitle: true,
          title: Text(
            l10n.text_2, // "جميع المزادات"
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios, color: Colors.white, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.menu, color: Colors.white),
              onPressed: () {}, // Drawer trigger
            ),
          ],
        ),
        body: Column(
          children: [
            // Search Bar
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 10.0),
              child: Container(
                decoration: BoxDecoration(
                  color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.04),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    hintText: l10n.text_53, // "ابحث عن مزاد..."
                    hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 14),
                    prefixIcon: const Icon(Icons.search, color: Colors.grey),
                    border: InputBorder.none,
                    contentPadding: const EdgeInsets.symmetric(vertical: 15),
                  ),
                ),
              ),
            ),
            
            // Grid of Auctions
            Expanded(
              child: _filteredAuctions.isEmpty
                  ? Center(
                      child: Text(
                        l10n.text_54, // "لا توجد نتائج بحث"
                        style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 16),
                      ),
                    )
                  : GridView.builder(
                      padding: const EdgeInsets.all(20),
                      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 2,
                        crossAxisSpacing: 16,
                        mainAxisSpacing: 20,
                        childAspectRatio: 0.40, // Providing more vertical space for info row
                      ),
                      itemCount: _filteredAuctions.length,
                      itemBuilder: (context, index) {
                        final auction = _filteredAuctions[index];
                        return _buildAuctionItem(
                          context,
                          isDarkMode,
                          auction['title']!,
                          auction['price']!,
                          auction['time']!,
                          auction['id']!,
                          auction['image']!,
                          auction['bids']!,
                        );
                      },
                    ),
            ),
          ],
        ),
      );
  }

  Widget _buildAuctionItem(
    BuildContext context,
    bool isDarkMode,
    String title,
    String price,
    String time,
    String id,
    String imageUrl,
    String bids,
  ) {
    final l10n = AppLocalizations.of(context)!;
    const Color primaryBlue = Color(0xFF0084FF);
    final isFavorite = ref.watch(favoritesProvider).contains(id);

    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
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
                  imageUrl,
                  height: 120,
                  width: double.infinity,
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(
                    height: 120,
                    color: Colors.grey[200],
                    child: const Icon(Icons.image, color: Colors.grey),
                  ),
                ),
              ),
              Positioned(
                top: 8,
                right: 8,
                child: GestureDetector(
                  onTap: () {
                    ref.read(favoritesProvider.notifier).toggleFavorite(id);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text(isFavorite ? l10n.text_55 : l10n.text_56),
                        duration: const Duration(seconds: 1),
                        behavior: SnackBarBehavior.floating,
                      ),
                    );
                  },
                  child: Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: isDarkMode ? Colors.black45 : Colors.white.withOpacity(0.8),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      isFavorite ? Icons.favorite : Icons.favorite_border,
                      color: isFavorite ? Colors.red : (isDarkMode ? Colors.white : Colors.black),
                      size: 18,
                    ),
                  ),
                ),
              ),
              Positioned(
                top: 8,
                left: 8,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: primaryBlue,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    'ID: $id',
                    style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
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
                Text(
                  title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: isDarkMode ? Colors.white : Colors.black,
                  ),
                ),
                const SizedBox(height: 6),
                Row(
                  children: [
                    const Icon(Icons.timer_outlined, color: Colors.red, size: 10),
                    const SizedBox(width: 4),
                    Flexible(
                      child: Text(
                        time,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(color: Colors.red, fontSize: 10, fontWeight: FontWeight.bold),
                      ),
                    ),
                    const SizedBox(width: 4),
                    const Icon(Icons.gavel, color: Colors.grey, size: 10),
                    const SizedBox(width: 2),
                    Flexible(
                      child: Text(
                        '$bids ${l10n.text_371}', // "مزايدات" or "bids"
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(color: Colors.grey, fontSize: 10),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Text(
                  price,
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: primaryBlue,
                  ),
                ),
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(builder: (context) => AuctionDetailsPage(auctionId: id)),
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: primaryBlue,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                      elevation: 0,
                    ),
                    child: Text(
                      l10n.text_3, // "زايد الآن"
                      style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.bold),
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

