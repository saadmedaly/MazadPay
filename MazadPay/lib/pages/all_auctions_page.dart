import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/pages/create_ad_start_page.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/pages/home_page.dart';

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
  int _selectedCategoryIndex = 0; 
  int _selectedSubCategoryIndex = 0;

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
        'location': l10n.text_113, // "تفرغ زينة"
        'category': 'cars',
        'image': 'assets/raf4.png',
        'status': 'active',
        'postedTime': 'منذ 5 ساعات'
      },
      {
        'id': '2',
        'title': 'هلكس خليجية نظيفة',
        'price': '760,000 MRU',
        'bids': '7',
        'time': '23h 30m 18s',
        'location': l10n.text_104, // "عرفات"
        'category': 'cars',
        'image': 'assets/corolla.png',
        'status': 'active',
        'postedTime': 'منذ 2 ساعات'
      },
      {
        'id': '3',
        'title': 'TOYOTA RAV4 2008 D4D',
        'price': '420,000 MRU',
        'bids': '20',
        'time': l10n.text_364, // "انتهى المزاد"
        'location': l10n.text_113,
        'category': 'cars',
        'image': 'assets/raf4.png',
        'status': 'finished',
        'postedTime': 'منذ يوم'
      },
      {
        'id': '4',
        'title': 'iPhone 15 Pro Max',
        'price': '55,000 MRU',
        'bids': '12',
        'time': '12h 10m 05s',
        'location': l10n.text_113,
        'category': 'phones',
        'image': 'assets/corolla.png',
        'status': 'active',
        'postedTime': 'منذ 3 ساعات'
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
        
        bool matchesCategory = true;
        if (_selectedCategoryIndex != 0) {
           final categoryKey = _categories[_selectedCategoryIndex]['key'];
           matchesCategory = auction['category'] == categoryKey;
        }
        
        return matchesSearch && matchesStatus && matchesCategory;
      }).toList();
    });
  }

  final List<Map<String, dynamic>> _categories = [
    {'title_key': 'text_74', 'image': 'assets/auctions/cars.png', 'key': 'all', 'count': 25},
    {'title_key': 'text_86', 'image': 'assets/auctions/cars.png', 'key': 'cars', 'count': 10},
    {'title_key': 'text_130', 'image': 'assets/auctions/phones.png', 'key': 'phones', 'count': 5},
    {'title_key': 'text_119', 'image': 'assets/auctions/houses.png', 'key': 'real_estate', 'count': 8},
    {'title_key': 'text_127', 'image': 'assets/auctions/phone_numbers.png', 'key': 'phone_numbers', 'count': 15},
    {'title_key': 'text_120', 'image': 'assets/auctions/home_appliances.png', 'key': 'home_appliances', 'count': 4},
  ];

  final List<Map<String, dynamic>> _subCategories = [
    {'title': '4x4', 'icon': Icons.directions_car_filled_outlined},
    {'title': 'يارة', 'icon': Icons.directions_car},
    {'title': 'تكسي', 'icon': Icons.local_taxi},
    {'title': 'باص', 'icon': Icons.airport_shuttle},
    {'title': 'شاحنة', 'icon': Icons.local_shipping},
  ];

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);
    
    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        endDrawer: const SideMenuDrawer(),
        appBar: AppBar(
          backgroundColor: primaryBlue,
          elevation: 0,
          toolbarHeight: 60,
          centerTitle: true,
          title: Text(
            "انواع المزادات", // "Types of Auctions" as per design
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
          leading: Builder(
            builder: (context) => IconButton(
              icon: const Icon(Icons.menu, color: Colors.white),
              onPressed: () => Scaffold.of(context).openEndDrawer(),
            ),
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.arrow_forward_ios, color: Colors.white, size: 20),
              onPressed: () => Navigator.of(context).pop(),
            ),
          ],
        ),
        body: Column(
          children: [
            // Top Categories Horizontal List
            const SizedBox(height: 12),
            SizedBox(
              height: 110,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: _categories.length,
                itemBuilder: (context, index) {
                  final cat = _categories[index];
                  bool isSelected = _selectedCategoryIndex == index;
                  String title = _getLocalizedTitle(cat['title_key'], l10n);

                  return GestureDetector(
                    onTap: () {
                      setState(() {
                        _selectedCategoryIndex = index;
                        _filterAuctions();
                      });
                    },
                    child: Container(
                      width: 100,
                      margin: const EdgeInsets.only(right: 12),
                      decoration: BoxDecoration(
                        color: isSelected ? primaryBlue.withValues(alpha: 0.1) : (isDarkMode ? const Color(0xFF1D1D1D) : Colors.white),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: isSelected ? primaryBlue : (isDarkMode ? const Color(0xFF333333) : Colors.grey.withValues(alpha: 0.1)),
                          width: 1.5,
                        ),
                      ),
                      child: Stack(
                        children: [
                          Positioned(
                            top: 6,
                            right: 6,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                              decoration: BoxDecoration(
                                color: Colors.orange,
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Text(
                                cat['count'].toString(),
                                style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                              ),
                            ),
                          ),
                          Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Image.asset(
                                  cat['image'],
                                  height: 48,
                                  width: 48,
                                  fit: BoxFit.contain,
                                  errorBuilder: (c,e,s) => Icon(Icons.category, color: isSelected ? primaryBlue : Colors.grey, size: 32),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  title,
                                  textAlign: TextAlign.center,
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 11,
                                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                                    color: isSelected ? primaryBlue : (isDarkMode ? Colors.white70 : Colors.black87),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),

            // Search Bar
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 12.0),
              child: Container(
                height: 40,
                decoration: BoxDecoration(
                  color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
                ),
                child: TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    hintText: "ابحث", // "Search"
                    hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 14),
                    prefixIcon: const Icon(Icons.search, color: Colors.grey, size: 20),
                    border: InputBorder.none,
                    contentPadding: const EdgeInsets.symmetric(vertical: 8),
                  ),
                ),
              ),
            ),
            
            // Sub-Categories selection
            const SizedBox(height: 4),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Align(
                alignment: Alignment.centerRight,
                child: Text(
                   "اختر نوعية السيارة التي تبحث عنها", // Example for cars
                  style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.grey[700]),
                ),
              ),
            ),
            const SizedBox(height: 8),
            SizedBox(
              height: 50,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: _subCategories.length,
                itemBuilder: (context, index) {
                  final sub = _subCategories[index];
                  bool isSelected = _selectedSubCategoryIndex == index;
                  return GestureDetector(
                    onTap: () => setState(() => _selectedSubCategoryIndex = index),
                    child: Container(
                      margin: const EdgeInsets.symmetric(horizontal: 6),
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      decoration: BoxDecoration(
                        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: isSelected ? primaryBlue : Colors.grey.withOpacity(0.2),
                          width: 1.5,
                        ),
                      ),
                      child: Row(
                        children: [
                          Icon(sub['icon'], size: 18, color: isSelected ? primaryBlue : Colors.grey),
                          const SizedBox(width: 6),
                          Text(
                            sub['title'],
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 12, 
                              fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                              color: isSelected ? primaryBlue : (isDarkMode ? Colors.white70 : Colors.black87),
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),

            // Custom Tabs
            const SizedBox(height: 16),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Container(
                height: 40,
                padding: const EdgeInsets.all(4),
                decoration: BoxDecoration(
                  color: isDarkMode ? const Color(0xFF1D1D1D) : const Color(0xFFF3F3F3),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  children: [
                    _buildDesignTab(1, "مزادات منتهية", 12),
                    _buildDesignTab(0, "مزادات نشطة", 150),
                  ],
                ),
              ),
            ),
            
            // Auction List (Vertical list of horizontal cards)
            Expanded(
              child: _filteredAuctions.isEmpty
                  ? Center(
                      child: Text(
                        l10n.text_54,
                        style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 16),
                      ),
                    )
                  : ListView.builder(
                      padding: const EdgeInsets.all(20),
                      itemCount: _filteredAuctions.length,
                      itemBuilder: (context, index) {
                        return _buildHorizontalAuctionCard(
                          context,
                          isDarkMode,
                          _filteredAuctions[index],
                        );
                      },
                    ),
            ),
          ],
        ),
        
        // Bottom Nav Bar (Matches designs)
        floatingActionButton: GestureDetector(
          onTap: () {
            Navigator.of(context).push(MaterialPageRoute(builder: (context) => const CreateAdStartPage()));
          },
          child: SizedBox(
            height: 60,
            width: 60,
            child: FloatingActionButton(
              onPressed: () => Navigator.of(context).push(MaterialPageRoute(builder: (context) => const CreateAdStartPage())),
              backgroundColor: Colors.transparent,
              elevation: 0,
              highlightElevation: 0,
              child: Image.asset('assets/botum_bar.png', fit: BoxFit.contain),
            ),
          ),
        ),
        floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
        bottomNavigationBar: BottomAppBar(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          shape: const CircularNotchedRectangle(),
          notchMargin: 6,
          child: SizedBox(
            height: 60,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildSimpleNavItem(Icons.home_outlined, l10n.text_1, 0),
                _buildSimpleNavItem(Icons.local_shipping_outlined, l10n.text_32, 1),
                const SizedBox(width: 48),
                _buildSimpleNavItem(Icons.storefront_outlined, l10n.text_33, 2),
                _buildSimpleNavItem(Icons.person_outline, l10n.text_19, 3),
              ],
            ),
          ),
        ),
      );
  }

  Widget _buildDesignTab(int index, String label, int count) {
    bool isSelected = _activeTabIndex == index;
    return Expanded(
      child: GestureDetector(
        onTap: () {
          setState(() {
            _activeTabIndex = index;
            _filterAuctions();
          });
        },
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          decoration: BoxDecoration(
            color: isSelected ? Colors.white : Colors.transparent,
            borderRadius: BorderRadius.circular(8),
            boxShadow: isSelected ? [BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 4, offset: const Offset(0, 2))] : [],
          ),
          child: Center(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  "($count)",
                  style: TextStyle(
                    fontSize: 12,
                    color: isSelected ? Colors.green : Colors.grey,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  ),
                ),
                const SizedBox(width: 4),
                Text(
                  label,
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 13,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                    color: isSelected ? Colors.black : Colors.grey,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHorizontalAuctionCard(BuildContext context, bool isDarkMode, Map<String, String> auction) {
    const Color primaryBlue = Color(0xFF0084FF);
    final id = auction['id']!;
    final isFavorite = ref.watch(favoritesProvider).contains(id);

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: IntrinsicHeight(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Left: Details
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      auction['title']!,
                      style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.bold),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        const Icon(Icons.timer_outlined, color: Colors.red, size: 14),
                        const SizedBox(width: 4),
                        Text(
                          auction['time']!,
                          style: const TextStyle(color: Colors.red, fontSize: 12, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        const Icon(Icons.gavel, color: Colors.grey, size: 14),
                        const SizedBox(width: 4),
                        Text(auction['bids']!, style: const TextStyle(color: Colors.grey, fontSize: 12)),
                        const Spacer(),
                        GestureDetector(
                          onTap: () => ref.read(favoritesProvider.notifier).toggleFavorite(id),
                          child: Icon(
                            isFavorite ? Icons.favorite : Icons.favorite_border,
                            color: isFavorite ? Colors.red : Colors.grey,
                            size: 20,
                          ),
                        ),
                      ],
                    ),
                    const Spacer(),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          auction['price']!,
                          style: const TextStyle(color: primaryBlue, fontSize: 15, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                         Icon(Icons.location_on_outlined, size: 12, color: Colors.grey[400]),
                        const SizedBox(width: 2),
                        Text(auction['location']!, style: TextStyle(fontSize: 10, color: Colors.grey[400])),
                        const Spacer(),
                        Text(auction['postedTime']!, style: TextStyle(fontSize: 10, color: Colors.grey[400])),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            // Right: Image
            ClipRRect(
              borderRadius: const BorderRadius.horizontal(right: Radius.circular(20)),
              child: Stack(
                children: [
                  Image.asset(
                    auction['image']!,
                    width: 120,
                    fit: BoxFit.cover,
                    errorBuilder: (c,e,s) => Container(width: 120, color: Colors.grey[200]),
                  ),
                  Positioned(
                    top: 8,
                    left: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(color: primaryBlue, borderRadius: BorderRadius.circular(4)),
                      child: Text("ID: $id", style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
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

  Widget _buildSimpleNavItem(IconData icon, String label, int index) {
     return InkWell(
      onTap: () {
        if (index == 0) Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const HomePage()));
        else if (index == 1) Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const ServicesPage()));
        else if (index == 3) Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const AccountPage()));
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: Colors.grey[600], size: 24),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 9, color: Colors.grey)),
        ],
      ),
    );
  }

  String _getLocalizedTitle(String key, AppLocalizations l10n) {
    switch (key) {
      case 'text_74': return l10n.text_74;
      case 'text_86': return l10n.text_86;
      case 'text_130': return l10n.text_130;
      case 'text_119': return l10n.text_119;
      case 'text_127': return l10n.text_127;
      case 'text_120': return l10n.text_120;
      default: return "Category";
    }
  }
}
