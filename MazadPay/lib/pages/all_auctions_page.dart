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
  int _activeTabIndex = 0;
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
        'location': l10n.text_113,
        'category': 'cars',
        'subCategory': '4x4',
        'image': 'assets/car0.png',
        'status': 'active',
        'postedTime': 'منذ 5 ساعات'
      },
      {
        'id': '2',
        'title': 'هلكس خليجية نظيفة',
        'price': '760,000 MRU',
        'bids': '7',
        'time': '23h 30m 18s',
        'location': l10n.text_104,
        'category': 'cars',
        'subCategory': 'standard',
        'image': 'assets/car1.jpg',
        'status': 'active',
        'postedTime': 'منذ 2 ساعات'
      },
      {
        'id': '5',
        'title': 'سيارة تاكسي بحالة ممتازة',
        'price': '450,000 MRU',
        'bids': '3',
        'time': '10h 15m 30s',
        'location': l10n.text_111,
        'category': 'cars',
        'subCategory': 'taxi',
        'image': 'assets/car2.jpg',
        'status': 'active',
        'postedTime': 'منذ ساعة'
      },
      {
        'id': '3',
        'title': 'TOYOTA RAV4 2008 D4D',
        'price': '420,000 MRU',
        'bids': '20',
        'time': l10n.text_364,
        'location': l10n.text_113,
        'category': 'cars',
        'subCategory': '4x4',
        'image': 'assets/car3.jpg',
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
        'image': 'assets/phone1.jpg',
        'status': 'active',
        'postedTime': 'منذ 3 ساعات'
      },
      {
        'id': '6',
        'title': 'Samsung Galaxy S24 Ultra',
        'price': '38,000 MRU',
        'bids': '8',
        'time': '05h 45m 20s',
        'location': l10n.text_113,
        'category': 'phones',
        'image': 'assets/phone2.jpg',
        'status': 'active',
        'postedTime': 'منذ 6 ساعات'
      },
      {
        'id': '7',
        'title': 'فيلا راقية وسط العاصمة',
        'price': '2,500,000 MRU',
        'bids': '5',
        'time': '48h 00m 00s',
        'location': l10n.text_113,
        'category': 'real_estate',
        'image': 'assets/maison1.jpg',
        'status': 'active',
        'postedTime': 'منذ يومين'
      },
      {
        'id': '8',
        'title': 'شقة 3 غرف للبيع',
        'price': '850,000 MRU',
        'bids': '11',
        'time': '24h 00m 00s',
        'location': l10n.text_113,
        'category': 'real_estate',
        'image': 'assets/maison2.jpg',
        'status': 'active',
        'postedTime': 'منذ يوم'
      },
      {
        'id': '9',
        'title': 'غسالة أوتوماتيك 7 كيلو',
        'price': '25,000 MRU',
        'bids': '4',
        'time': '08h 30m 00s',
        'location': l10n.text_113,
        'category': 'home_appliances',
        'image': 'assets/Appareils de maison1.jpg',
        'status': 'active',
        'postedTime': 'منذ 4 ساعات'
      },
      {
        'id': '10',
        'title': 'ثلاجة سامسونج 500 لتر',
        'price': '35,000 MRU',
        'bids': '6',
        'time': '15h 00m 00s',
        'location': l10n.text_113,
        'category': 'home_appliances',
        'image': 'assets/Appareils de maison2.jpg',
        'status': 'active',
        'postedTime': 'منذ 7 ساعات'
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

        bool matchesSubCategory = true;
        if (_selectedCategoryIndex == 1 && _selectedSubCategoryIndex != -1) {
          final subKey = _carSubCategories[_selectedSubCategoryIndex]['key'];
          if (auction.containsKey('subCategory')) {
            matchesSubCategory = auction['subCategory'] == subKey;
          }
        }

        return matchesSearch && matchesStatus && matchesCategory && matchesSubCategory;
      }).toList();
    });
  }

  final List<Map<String, dynamic>> _categories = [
    {'title_key': 'text_74',  'image': 'assets/auctions/other.png',            'key': 'all',             'count': 76},
    {'title_key': 'text_86',  'image': 'assets/auctions/cars.png',             'key': 'cars',            'count': 45},
    {'title_key': 'text_130', 'image': 'assets/auctions/phones.png',           'key': 'phones',          'count': 12},
    {'title_key': 'text_119', 'image': 'assets/auctions/houses.png',           'key': 'real_estate',     'count': 31},
    {'title_key': 'text_127', 'image': 'assets/auctions/phone_numbers.png',    'key': 'phone_numbers',   'count': 15},
    {'title_key': 'text_120', 'image': 'assets/auctions/home_appliances.png',  'key': 'home_appliances', 'count': 8},
    {'title_key': 'text_124', 'image': 'assets/auctions/animals.png',          'key': 'animals',         'count': 5},
    {'title_key': 'text_123', 'image': 'assets/auctions/womens_accessories.png','key': 'womens',         'count': 9},
    {'title_key': 'text_122', 'image': 'assets/auctions/mens_accessories.png', 'key': 'mens',            'count': 7},
    {'title_key': 'text_125', 'image': 'assets/auctions/heavy_equipment.png',  'key': 'trucks',          'count': 6},
    {'title_key': 'text_131', 'image': 'assets/auctions/phones.png',           'key': 'electronics',     'count': 11},
    {'title_key': 'text_129', 'image': 'assets/auctions/selling_projects.png', 'key': 'projects',        'count': 3},
    {'title_key': 'text_133', 'image': 'assets/auctions/other.png',            'key': 'bikes',           'count': 4},
  ];

  final List<Map<String, dynamic>> _carSubCategories = [
    {'title_key': 'text_115', 'image': 'assets/car1.jpg', 'key': 'standard'},
    {'title_key': 'text_114', 'image': 'assets/car0.png', 'key': '4x4'},
    {'title_key': 'text_116', 'image': 'assets/car2.jpg', 'key': 'taxi'},
    {'title_key': 'text_117', 'image': 'assets/car4.jpg', 'key': 'damaged'},
    {'title_key': 'text_118', 'image': 'assets/car5.png', 'key': 'electric'},
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
          "انواع المزادات",
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
      body: CustomScrollView(
        slivers: [
          SliverToBoxAdapter(
            child: Column(
              children: [
                // Top Categories Horizontal List
                const SizedBox(height: 12),
                SizedBox(
                  height: 115,
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
                            _selectedSubCategoryIndex = -1;
                            _filterAuctions();
                          });
                        },
                        child: Container(
                          width: 105,
                          margin: const EdgeInsets.only(right: 12, top: 4, bottom: 4),
                          decoration: BoxDecoration(
                            color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                            borderRadius: BorderRadius.circular(20),
                            boxShadow: isSelected
                                ? [BoxShadow(color: primaryBlue.withValues(alpha: 0.2), blurRadius: 8, offset: const Offset(0, 4))]
                                : [BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 4, offset: const Offset(0, 2))],
                            border: Border.all(
                              color: isSelected ? primaryBlue : Colors.transparent,
                              width: 2,
                            ),
                          ),
                          child: Stack(
                            clipBehavior: Clip.none,
                            children: [
                              Center(
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Image.asset(
                                      cat['image'],
                                      height: 44,
                                      width: 44,
                                      fit: BoxFit.contain,
                                      errorBuilder: (c, e, s) => Icon(Icons.category, color: isSelected ? primaryBlue : Colors.grey, size: 32),
                                    ),
                                    const SizedBox(height: 6),
                                    Text(
                                      title,
                                      textAlign: TextAlign.center,
                                      style: GoogleFonts.plusJakartaSans(
                                        fontSize: 12,
                                        fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                                        color: isSelected ? primaryBlue : (isDarkMode ? Colors.white70 : Colors.black87),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Positioned(
                                top: -4,
                                right: -4,
                                child: Container(
                                  padding: const EdgeInsets.all(6),
                                  decoration: const BoxDecoration(
                                    color: Color(0xFFFF3B30),
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
                        ),
                      );
                    },
                  ),
                ),

                // Search Bar
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 12.0),
                  child: Container(
                    height: 48,
                    decoration: BoxDecoration(
                      color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                      borderRadius: BorderRadius.circular(24),
                      border: Border.all(color: Colors.grey.withValues(alpha: 0.1)),
                      boxShadow: [
                        BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 10, offset: const Offset(0, 4))
                      ],
                    ),
                    child: TextField(
                      controller: _searchController,
                      textAlign: TextAlign.center,
                      decoration: InputDecoration(
                        hintText: "البحث",
                        hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 15),
                        prefixIcon: const Icon(Icons.search, color: Colors.grey, size: 22),
                        border: InputBorder.none,
                        contentPadding: const EdgeInsets.symmetric(vertical: 12),
                      ),
                    ),
                  ),
                ),

                // Sub-Categories selection (cars only)
                if (_selectedCategoryIndex == 1) ...[
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                    child: Align(
                      alignment: Alignment.centerRight,
                      child: Text(
                        "اختر نوعية السيارة التي تبحث عنها",
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: isDarkMode ? Colors.white : Colors.black,
                        ),
                      ),
                    ),
                  ),
                  SizedBox(
                    height: 100,
                    child: ListView.builder(
                      scrollDirection: Axis.horizontal,
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: _carSubCategories.length,
                      itemBuilder: (context, index) {
                        final sub = _carSubCategories[index];
                        bool isSelected = _selectedSubCategoryIndex == index;
                        return GestureDetector(
                          onTap: () {
                            setState(() {
                              _selectedSubCategoryIndex = index;
                              _filterAuctions();
                            });
                          },
                          child: Container(
                            width: 110,
                            margin: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                            decoration: BoxDecoration(
                              color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: isSelected ? primaryBlue : Colors.transparent,
                                width: 2,
                              ),
                              boxShadow: [
                                BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4, offset: const Offset(0, 2))
                              ],
                            ),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Image.asset(
                                  sub['image'],
                                  height: 50,
                                  fit: BoxFit.contain,
                                  errorBuilder: (c, e, s) => Icon(Icons.directions_car, color: isSelected ? primaryBlue : Colors.grey, size: 30),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  _getLocalizedTitle(sub['title_key'], l10n),
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12,
                                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
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
                ],

                // Custom Tabs
                const SizedBox(height: 16),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Row(
                    children: [
                      _buildStitchTab(1, "مزايدات منتهية", "00", const Color(0xFFFFDEDE), const Color(0xFFFF3B30)),
                      const SizedBox(width: 12),
                      _buildStitchTab(0, "مزايدات نشطة", "02", const Color(0xFFE8F5E9), const Color(0xFF2E7D32)),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
              ],
            ),
          ),

          // Auction List
          _filteredAuctions.isEmpty
              ? SliverFillRemaining(
                  child: Center(
                    child: Text(
                      l10n.text_54,
                      style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 16),
                    ),
                  ),
                )
              : SliverPadding(
                  padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, index) => _buildHorizontalAuctionCard(
                        context,
                        isDarkMode,
                        _filteredAuctions[index],
                      ),
                      childCount: _filteredAuctions.length,
                    ),
                  ),
                ),
        ],
      ),

      floatingActionButton: FloatingActionButton(
        onPressed: () => Navigator.of(context).push(MaterialPageRoute(builder: (context) => const CreateAdStartPage())),
        backgroundColor: Colors.transparent,
        elevation: 0,
        highlightElevation: 0,
        child: Image.asset('assets/botum_bar.png', fit: BoxFit.contain),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: BottomAppBar(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        elevation: 0,
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

  Widget _buildStitchTab(int index, String label, String count, Color bgColor, Color textColor) {
    bool isSelected = _activeTabIndex == index;
    return Expanded(
      child: GestureDetector(
        onTap: () {
          setState(() {
            _activeTabIndex = index;
            _filterAuctions();
          });
        },
        child: Container(
          height: 45,
          decoration: BoxDecoration(
            color: isSelected ? bgColor : Colors.grey.withValues(alpha: 0.05),
            borderRadius: BorderRadius.circular(12),
            border: isSelected ? Border.all(color: textColor.withValues(alpha: 0.3)) : null,
          ),
          child: Center(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  label,
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 14,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                    color: isSelected ? textColor : Colors.grey,
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: isSelected ? textColor : Colors.grey.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    count,
                    style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold),
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
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Left: Details
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    auction['title']!,
                    style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.bold),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(Icons.timer_outlined, color: Colors.red, size: 13),
                      const SizedBox(width: 4),
                      Flexible(
                        child: Text(
                          auction['time']!,
                          style: const TextStyle(color: Colors.red, fontSize: 11, fontWeight: FontWeight.bold),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(Icons.gavel, color: Colors.grey, size: 13),
                      const SizedBox(width: 4),
                      Text(auction['bids']!, style: const TextStyle(color: Colors.grey, fontSize: 11)),
                      const Spacer(),
                      GestureDetector(
                        onTap: () => ref.read(favoritesProvider.notifier).toggleFavorite(id),
                        child: Icon(
                          isFavorite ? Icons.favorite : Icons.favorite_border,
                          color: isFavorite ? Colors.red : Colors.grey,
                          size: 18,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Text(
                    auction['price']!,
                    style: const TextStyle(color: primaryBlue, fontSize: 14, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Icon(Icons.location_on_outlined, size: 11, color: Colors.grey[400]),
                      const SizedBox(width: 2),
                      Flexible(
                        child: Text(
                          auction['location']!,
                          style: TextStyle(fontSize: 10, color: Colors.grey[400]),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
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
                  height: 110,
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(width: 120, height: 110, color: Colors.grey[200]),
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
    );
  }

  Widget _buildSimpleNavItem(IconData icon, String label, int index) {
    return InkWell(
      onTap: () {
        if (index == 0) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const HomePage()));
        } else if (index == 1) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const ServicesPage()));
        } else if (index == 3) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const AccountPage()));
        }
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
      case 'text_114': return l10n.text_114;
      case 'text_115': return l10n.text_115;
      case 'text_116': return l10n.text_116;
      case 'text_117': return l10n.text_117;
      case 'text_118': return l10n.text_118;
      case 'text_122': return l10n.text_122;
      case 'text_123': return l10n.text_123;
      case 'text_124': return l10n.text_124;
      case 'text_125': return l10n.text_125;
      case 'text_129': return l10n.text_129;
      case 'text_131': return l10n.text_131;
      case 'text_133': return l10n.text_133;
      default: return "Category";
    }
  }
}
