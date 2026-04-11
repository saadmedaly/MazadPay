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
  
  // Categories with images and counts
  final List<Map<String, dynamic>> _categories = [
    {'title': 'سيارات', 'count': 45, 'image': 'assets/cars.png'},
    {'title': 'عقارات', 'count': 31, 'image': 'assets/building.png'},
    {'title': 'قطع ارضية', 'count': 22, 'image': 'assets/morceau.png'},
    {'title': 'أجهزة', 'count': 15, 'image': 'assets/iphone.png'},
  ];

  // Car subtypes
  final List<Map<String, String>> _carSubtypes = [
    {'title': 'رباعية 4x4', 'image': 'assets/raf4.png'},
    {'title': 'عادية', 'image': 'assets/corolla.png'},
    {'title': 'تاكسي', 'image': 'assets/corolla.png'}, // Placeholder for taxi
  ];

  // Expanded mock data for auctions
  final List<Map<String, String>> _allAuctions = [
    {
      'id': '1',
      'title': 'Toyota prado tx 2017',
      'price': '15,000 أوقية جديدة',
      'bids': '15',
      'time': '23h 30m 18s',
      'location': 'تفرغ زينة',
      'postedTime': 'منذ 23 دقائق',
      'image': 'assets/raf4.png',
      'status': 'active'
    },
    {
      'id': '2',
      'title': 'هلكس خليجية نظيفة',
      'price': '760,000 أوقية جديدة',
      'bids': '7',
      'time': '23h 30m 18s',
      'location': 'عرفات',
      'postedTime': 'منذ 23 دقائق',
      'image': 'assets/corolla.png',
      'status': 'active'
    },
    {
      'id': '3',
      'title': 'TOYOTA RAV4 2008 D4D',
      'price': '420,000 أوقية جديدة',
      'bids': '20',
      'time': 'انتهى المزاد',
      'location': 'تفرغ زينة',
      'postedTime': 'منذ 24 ساعة',
      'image': 'assets/raf4.png',
      'status': 'finished'
    },
  ];

  List<Map<String, String>> _filteredAuctions = [];
  int _selectedCategoryIndex = 0;
  int _selectedSubtypeIndex = 0;
  int _activeTabIndex = 0; // 0 for Active, 1 for Finished

  @override
  void initState() {
    super.initState();
    _filteredAuctions = _allAuctions.where((a) => a['status'] == 'active').toList();
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
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);
    
    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: primaryBlue,
          elevation: 0,
          centerTitle: true,
          title: Text(
            'انواع المزايدات',
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
        body: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Top Categories Section
              _buildTopCategories(isDarkMode),
              
              const SizedBox(height: 16),
              
              // Search Bar Section
              _buildSearchBar(isDarkMode),
              
              const SizedBox(height: 16),
              
              // Divider
              const Divider(thickness: 1, height: 1),
              
              const SizedBox(height: 16),
              
              // Sub-category Selection
              _buildSubCategorySelection(isDarkMode),
              
              const SizedBox(height: 16),
              
              // Auction Tabs
              _buildAuctionTabs(isDarkMode),
              
              const SizedBox(height: 16),
              
              // Auctions List
              _buildAuctionsList(isDarkMode),
              
              const SizedBox(height: 80), // To avoid bottom bar overlap
            ],
          ),
        ),
        floatingActionButton: _buildFAB(context),
        floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
        bottomNavigationBar: _buildBottomAppBar(isDarkMode),
      ),
    );
  }

  Widget _buildTopCategories(bool isDarkMode) {
    return Container(
      height: 130,
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: _categories.length,
        itemBuilder: (context, index) {
          final cat = _categories[index];
          bool isSelected = _selectedCategoryIndex == index;
          return GestureDetector(
            onTap: () => setState(() => _selectedCategoryIndex = index),
            child: Container(
              width: 100,
              margin: const EdgeInsets.symmetric(horizontal: 6),
              child: Column(
                children: [
                  Stack(
                    clipBehavior: Clip.none,
                    children: [
                      Container(
                        height: 80,
                        width: 90,
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(12),
                          image: DecorationImage(
                            image: AssetImage(cat['image']),
                            fit: BoxFit.cover,
                          ),
                          border: isSelected ? Border.all(color: const Color(0xFF0084FF), width: 2) : null,
                        ),
                      ),
                      Positioned(
                        top: -5,
                        right: -5,
                        child: Container(
                          padding: const EdgeInsets.all(4),
                          decoration: const BoxDecoration(
                            color: Color(0xFFFF5C5C),
                            shape: BoxShape.circle,
                          ),
                          child: Text(
                            cat['count'].toString(),
                            style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Text(
                    cat['title'],
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: isDarkMode ? Colors.white : Colors.black,
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildSearchBar(bool isDarkMode) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20.0),
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
          textAlign: TextAlign.right,
          decoration: InputDecoration(
            hintText: 'البحث',
            hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 14),
            prefixIcon: const Icon(Icons.search, color: Color(0xFFFF8B8B)), // Pinkish icon as per image
            border: InputBorder.none,
            contentPadding: const EdgeInsets.symmetric(vertical: 15, horizontal: 16),
          ),
        ),
      ),
    );
  }

  Widget _buildSubCategorySelection(bool isDarkMode) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20.0),
          child: Text(
            'اختر نوعية السيارة التي تبحث عنها',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 90,
          child: ListView.builder(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16),
            itemCount: _carSubtypes.length,
            itemBuilder: (context, index) {
              final sub = _carSubtypes[index];
              bool isSelected = _selectedSubtypeIndex == index;
              return GestureDetector(
                onTap: () => setState(() => _selectedSubtypeIndex = index),
                child: Container(
                  width: 120,
                  margin: const EdgeInsets.symmetric(horizontal: 6),
                  child: Column(
                    children: [
                      Container(
                        height: 50,
                        width: 110,
                        decoration: BoxDecoration(
                          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: isSelected ? const Color(0xFF0084FF) : Colors.grey.withOpacity(0.2),
                            width: isSelected ? 2 : 1,
                          ),
                        ),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(12),
                          child: Image.asset(sub['image']!, fit: BoxFit.contain),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        sub['title']!,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                          color: isDarkMode ? Colors.white : Colors.black,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildAuctionTabs(bool isDarkMode) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20.0),
      child: Row(
        children: [
          Expanded(
            child: GestureDetector(
              onTap: () {
                setState(() => _activeTabIndex = 0);
                _filterAuctions();
              },
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 12),
                decoration: BoxDecoration(
                  color: _activeTabIndex == 0 ? const Color(0xFFEFFFEE) : Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: _activeTabIndex == 0 ? const Color(0xFF4CAF50) : Colors.grey.withOpacity(0.2)),
                ),
                child: Center(
                  child: Text(
                    'مزايدات نشطة 06',
                    style: GoogleFonts.plusJakartaSans(
                      color: _activeTabIndex == 0 ? const Color(0xFF4CAF50) : Colors.grey,
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: GestureDetector(
              onTap: () {
                setState(() => _activeTabIndex = 1);
                _filterAuctions();
              },
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 12),
                decoration: BoxDecoration(
                  color: _activeTabIndex == 1 ? const Color(0xFFFFEEEE) : Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: _activeTabIndex == 1 ? const Color(0xFFFF5252) : Colors.grey.withOpacity(0.2)),
                ),
                child: Center(
                  child: Text(
                    'مزايدات منتهية 01',
                    style: GoogleFonts.plusJakartaSans(
                      color: _activeTabIndex == 1 ? const Color(0xFFFF5252) : Colors.grey,
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAuctionsList(bool isDarkMode) {
    if (_filteredAuctions.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(40.0),
          child: Text(
            'لا توجد نتائج بحث',
            style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 16),
          ),
        ),
      );
    }
    return ListView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.symmetric(horizontal: 20),
      itemCount: _filteredAuctions.length,
      itemBuilder: (context, index) {
        final auction = _filteredAuctions[index];
        return _buildAuctionListItem(context, isDarkMode, auction);
      },
    );
  }

  Widget _buildAuctionListItem(BuildContext context, bool isDarkMode, Map<String, String> auction) {
    final favorites = ref.watch(favoritesProvider);
    final isFavorite = favorites.contains(auction['id']);
    const Color primaryBlue = Color(0xFF0084FF);

    return GestureDetector(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => AuctionDetailsPage(auctionId: auction['id']!),
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          children: [
            // Left Info Side
            Expanded(
              flex: 3,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          auction['title']!,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                            color: isDarkMode ? Colors.white : Colors.black,
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      GestureDetector(
                        onTap: () => ref.read(favoritesProvider.notifier).toggleFavorite(auction['id']!),
                        child: Icon(
                          isFavorite ? Icons.favorite : Icons.favorite_border,
                          color: isFavorite ? Colors.red : Colors.grey,
                          size: 20,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    auction['price']!,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                      color: isDarkMode ? Colors.white70 : Colors.black87,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: Colors.grey.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.gavel, size: 14, color: Colors.black54),
                            const SizedBox(width: 4),
                            Text(auction['bids']!, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Flexible(
                        child: Row(
                          children: [
                            const Icon(Icons.timer_outlined, size: 14, color: Colors.red),
                            const SizedBox(width: 4),
                            Flexible(
                              child: Text(
                                auction['time']!,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(color: Colors.red, fontSize: 11, fontWeight: FontWeight.bold),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '${auction['postedTime']} . ${auction['location']}',
                    style: GoogleFonts.plusJakartaSans(fontSize: 10, color: Colors.grey),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            // Right Image Side
            Expanded(
              flex: 2,
              child: ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: Image.asset(
                  auction['image']!,
                  height: 90,
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(height: 90, color: Colors.grey[200]),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }


  Widget _buildFAB(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          height: 70,
          width: 70,
          child: FloatingActionButton(
            onPressed: () {},
            backgroundColor: Colors.transparent,
            elevation: 0,
            child: Image.asset(
              'assets/botum_bar.png',
              fit: BoxFit.contain,
            ),
          ),
        ),
        const Text(
          'إعلان جديد',
          style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.black54),
        ),
      ],
    );
  }

  Widget _buildBottomAppBar(bool isDarkMode) {
    return BottomAppBar(
      color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      shape: const CircularNotchedRectangle(),
      notchMargin: 8,
      child: SizedBox(
        height: 60,
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _buildNavItem(Icons.home_outlined, 'الرئيسية', true),
            _buildNavItem(Icons.local_shipping_outlined, 'توصيل', false),
            const SizedBox(width: 60), // Space for FAB
            _buildNavItem(Icons.storefront_outlined, 'التجارة الالكترونية', false),
            _buildNavItem(Icons.person_outline, 'حسابي', false),
          ],
        ),
      ),
    );
  }

  Widget _buildNavItem(IconData icon, String label, bool isSelected) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Icon(icon, color: isSelected ? const Color(0xFF135BEC) : Colors.grey),
        const SizedBox(height: 4),
        Text(
          label,
          style: TextStyle(
            fontSize: 10,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
            color: isSelected ? const Color(0xFF135BEC) : Colors.grey,
          ),
        ),
      ],
    );
  }

}
