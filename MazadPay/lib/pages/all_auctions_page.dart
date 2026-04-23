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
import '../services/auction_api.dart';

class AllAuctionsPage extends ConsumerStatefulWidget {
  const AllAuctionsPage({super.key});

  @override
  ConsumerState<AllAuctionsPage> createState() => _AllAuctionsPageState();
}

class _AllAuctionsPageState extends ConsumerState<AllAuctionsPage> {
  final TextEditingController _searchController = TextEditingController();
  final AuctionApi _auctionApi = AuctionApi();

  List<Map<String, dynamic>> _allAuctions = [];
  List<Map<String, dynamic>> _filteredAuctions = [];
  int _activeTabIndex = 0;
  int _selectedCategoryIndex = 0;
  int _selectedSubCategoryIndex = 0;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchChanged);
    _loadAuctions();
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    super.dispose();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
  }

  void _onSearchChanged() {
    _filterAuctions();
  }

  Future<void> _loadAuctions() async {
    try {
      final response = await _auctionApi.getAuctions(
        page: 1,
        limit: 50,
        status: _activeTabIndex == 0 ? 'active' : 'completed',
      );

      setState(() {
        _isLoading = false;
        if (response.success && response.data != null) {
          _allAuctions = List.from(response.data!['auctions'] ?? []);
          _filteredAuctions = List.from(_allAuctions);
        }
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_loading_auctions)),
        );
      }
    }
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
    {'title_key': 'text_86',  'image': 'assets/auctions/cars.png',             'key': 'cars',            'count': 4},
    {'title_key': 'text_130', 'image': 'assets/auctions/phones.png',           'key': 'phones',          'count': 6},
    {'title_key': 'text_119', 'image': 'assets/auctions/houses.png',           'key': 'real_estate',     'count': 5},
    {'title_key': 'text_127', 'image': 'assets/auctions/phone_numbers.png',    'key': 'phone_numbers',   'count': 0},
    {'title_key': 'text_120', 'image': 'assets/auctions/home_appliances.png',  'key': 'home_appliances', 'count': 8},
    {'title_key': 'text_124', 'image': 'assets/auctions/animals.png',          'key': 'animals',         'count': 8},
    {'title_key': 'text_123', 'image': 'assets/auctions/womens_accessories.png','key': 'womens',         'count': 8},
    {'title_key': 'text_122', 'image': 'assets/auctions/mens_accessories.png', 'key': 'mens',            'count': 10},
    {'title_key': 'text_125', 'image': 'assets/auctions/heavy_equipment.png',  'key': 'trucks',          'count': 8},
    {'title_key': 'text_131', 'image': 'assets/auctions/phones.png',                    'key': 'electronics',    'count': 8},
    {'title_key': 'text_129', 'image': 'assets/auctions/selling_projects.png',          'key': 'projects',       'count': 3},
    {'title_key': 'text_133', 'image': 'assets/auctions/Vélos.png',                     'key': 'bikes',          'count': 12},
    {'title_key': 'text_139', 'image': 'assets/auctions/Matériaux lourds.jpg',          'key': 'heavy_materials', 'count': 10},
    {'title_key': 'text_134', 'image': 'assets/auctions/Pièces de sol.jpg',             'key': 'land_plots',     'count': 4},
    {'title_key': 'text_137', 'image': 'assets/auctions/Meubles.jpg',                   'key': 'furniture',      'count': 8},
  ];

  final List<Map<String, dynamic>> _carSubCategories = [
    {'title_key': 'text_115', 'image': 'assets/car1.jpg', 'key': 'standard'},
    {'title_key': 'text_114', 'image': 'assets/car0.png', 'key': '4x4'},
    {'title_key': 'text_116', 'image': 'assets/car2.jpg', 'key': 'taxi'},
    {'title_key': 'text_117', 'image': 'assets/car4.jpg', 'key': 'damaged'},
    {'title_key': 'text_118', 'image': 'assets/car5.png', 'key': 'electric'},
  ];

  // خريطة تربط كل فئة بقائمة عناصرها المرئية في شريط التمرير
  late final Map<String, List<Map<String, dynamic>>> _categoryItems = {
    'cars': _carSubCategories,
    'phones': [
      {'label': 'iPhone 15 Pro Max', 'image': 'assets/phone1.jpg'},
      {'label': 'Samsung S24 Ultra', 'image': 'assets/phone2.jpg'},
      {'label': 'Huawei P60 Pro',    'image': 'assets/phone3.jpg'},
      {'label': 'Xiaomi 14 Ultra',   'image': 'assets/phone4.jpg'},
      {'label': 'Google Pixel 8',    'image': 'assets/phone5.jpg'},
      {'label': 'OnePlus 12',        'image': 'assets/phone6.jpg'},
    ],
    'electronics': [
      {'label': 'لابتوب',     'image': 'assets/Électronique1.jpg'},
      {'label': 'آيباد',      'image': 'assets/Électronique2.jpg'},
      {'label': 'كاميرا',     'image': 'assets/Électronique3.jpg'},
      {'label': 'بلاي ستيشن','image': 'assets/Électronique4.jpg'},
      {'label': 'سماعات',     'image': 'assets/Électronique5.jpg'},
      {'label': 'شاشة',       'image': 'assets/Électronique6.jpg'},
      {'label': 'طابعة',      'image': 'assets/Électronique7.jpg'},
      {'label': 'راوتر',      'image': 'assets/Électronique8.jpg'},
    ],
    'real_estate': [
      {'label': 'فيلا',         'image': 'assets/maison1.jpg'},
      {'label': 'شقة',          'image': 'assets/maison2.jpg'},
      {'label': 'منزل',         'image': 'assets/maison3.jpg'},
      {'label': 'شقة مفروشة',   'image': 'assets/maison4.jpg'},
      {'label': 'مبنى تجاري',   'image': 'assets/maison5.jpg'},
    ],
    'home_appliances': [
      {'label': 'غسالة',   'image': 'assets/Appareils de maison1.jpg'},
      {'label': 'ثلاجة',   'image': 'assets/Appareils de maison2.jpg'},
      {'label': 'مكيف',    'image': 'assets/Appareils de maison3.jpg'},
      {'label': 'بوتوغاز', 'image': 'assets/Appareils de maison4.jpg'},
      {'label': 'تلفزيون', 'image': 'assets/Appareils de maison5.jpg'},
      {'label': 'مجفف',    'image': 'assets/Appareils de maison6.jpg'},
      {'label': 'طباخ',    'image': 'assets/Appareils de maison7.jpg'},
      {'label': 'مكنسة',   'image': 'assets/Appareils de maison8.jpg'},
    ],
    'animals': [
      {'label': 'حصان',  'image': 'assets/animal1.jpg'},
      {'label': 'جمل',   'image': 'assets/animal2.jpg'},
      {'label': 'ماعز',  'image': 'assets/animal3.jpg'},
      {'label': 'خروف',  'image': 'assets/animal4.jpg'},
      {'label': 'بقرة',  'image': 'assets/animal5.jpg'},
      {'label': 'دجاج',  'image': 'assets/animal6.jpg'},
      {'label': 'كبش',   'image': 'assets/animal7.jpg'},
      {'label': 'إبل',   'image': 'assets/animal8.jpg'},
    ],
    'womens': [
      {'label': 'حقيبة شانيل',   'image': 'assets/Fournitures pour femmes1.jpg'},
      {'label': 'عطر',            'image': 'assets/Fournitures pour femmes2.jpg'},
      {'label': 'ساعة نسائية',    'image': 'assets/Fournitures pour femmes3.jpg'},
      {'label': 'مجوهرات',        'image': 'assets/Fournitures pour femmes4.jpg'},
      {'label': 'فستان',          'image': 'assets/Fournitures pour femmes5.jpg'},
      {'label': 'حذاء',           'image': 'assets/Fournitures pour femmes6.jpg'},
      {'label': 'نظارة',          'image': 'assets/Fournitures pour femmes7.jpg'},
      {'label': 'خاتم ألماس',     'image': 'assets/Fournitures pour femmes8.jpg'},
    ],
    'mens': [
      {'label': 'ساعة رولكس',  'image': 'assets/Fournitures pour hommes1.jpg'},
      {'label': 'حقيبة جلد',   'image': 'assets/Fournitures pour hommes2.jpg'},
      {'label': 'بدلة',         'image': 'assets/Fournitures pour hommes3.jpg'},
      {'label': 'حذاء جلد',    'image': 'assets/Fournitures pour hommes4.jpg'},
      {'label': 'نظارة',        'image': 'assets/Fournitures pour hommes5.jpg'},
      {'label': 'عطر',          'image': 'assets/Fournitures pour hommes6.jpg'},
      {'label': 'حزام هيرمس',  'image': 'assets/Fournitures pour hommes7.jpg'},
      {'label': 'قميص',         'image': 'assets/Fournitures pour hommes8.jpg'},
      {'label': 'خاتم ذهب',    'image': 'assets/Fournitures pour hommes9.jpg'},
      {'label': 'محفظة',        'image': 'assets/Fournitures pour hommes10.jpg'},
    ],
    'trucks': [
      {'label': 'مرسيدس أكتروس', 'image': 'assets/Camions1.jpg'},
      {'label': 'مان TGX',        'image': 'assets/Camions2.jpg'},
      {'label': 'فولفو FH16',     'image': 'assets/Camions3.jpg'},
      {'label': 'سكانيا R500',    'image': 'assets/Camions4.jpg'},
      {'label': 'رينو T520',      'image': 'assets/Camions5.jpg'},
      {'label': 'إيفيكو',         'image': 'assets/Camions6.jpg'},
      {'label': 'دايملر',         'image': 'assets/Camions7.jpg'},
      {'label': 'نيسان كوندور',   'image': 'assets/Camions8.jpg'},
    ],
    'bikes': [
      {'label': 'ياماها R15',  'image': 'assets/Motos1.jpg'},
      {'label': 'هوندا CB500', 'image': 'assets/Motos2.jpg'},
      {'label': 'سوزوكي GSX',  'image': 'assets/Motos3.jpg'},
      {'label': 'كاوازاكي',    'image': 'assets/Motos4.jpg'},
      {'label': 'BMW G310R',   'image': 'assets/Motos5.jpg'},
      {'label': 'هارلي',       'image': 'assets/Motos6.jpg'},
      {'label': 'دوكاتي',      'image': 'assets/Motos7.jpg'},
      {'label': 'ترياومف',     'image': 'assets/Motos8.jpg'},
      {'label': 'أبريليا',     'image': 'assets/Motos9.jpg'},
      {'label': 'KTM Duke',    'image': 'assets/Motos10.jpg'},
      {'label': 'هيوسانج',     'image': 'assets/Motos11.jpg'},
      {'label': 'موتو غوتسي',  'image': 'assets/Motos12.jpg'},
    ],
    'heavy_materials': [
      {'label': 'حفارة',     'image': 'assets/Matériel lourd1.jpg'},
      {'label': 'جرافة',     'image': 'assets/Matériel lourd2.jpg'},
      {'label': 'رافعة شوكية','image': 'assets/Matériel lourd3.jpg'},
      {'label': 'خلاطة',     'image': 'assets/Matériel lourd4.jpg'},
      {'label': 'ضاغط هواء', 'image': 'assets/Matériel lourd5.jpg'},
      {'label': 'شاحنة قلاب','image': 'assets/Matériel lourd6.jpg'},
      {'label': 'مضخة مياه', 'image': 'assets/Matériel lourd7.jpg'},
      {'label': 'مولد كهرباء','image': 'assets/Matériel lourd8.jpg'},
      {'label': 'آلة تسوية', 'image': 'assets/Matériel lourd9.jpg'},
      {'label': 'كرين 20 طن','image': 'assets/Matériel lourd10.jpg'},
    ],
    'land_plots': [
      {'label': 'حي تفاريغ',  'image': 'assets/Terrains1.jpg'},
      {'label': 'أرض سكنية',  'image': 'assets/Terrains2.jpg'},
      {'label': 'قطعة تجارية','image': 'assets/Terrains3.jpg'},
      {'label': 'أرض زراعية', 'image': 'assets/Terrains4.jpg'},
    ],
    'furniture': [
      {'label': 'طقم صالون',  'image': 'assets/Meubles1.jpg'},
      {'label': 'غرفة نوم',   'image': 'assets/Meubles2.jpg'},
      {'label': 'طاولة سفرة', 'image': 'assets/Meubles3.jpg'},
      {'label': 'مكتبة',      'image': 'assets/Meubles4.jpg'},
      {'label': 'كنبة جلد',   'image': 'assets/Meubles5.jpg'},
      {'label': 'طاولة قهوة', 'image': 'assets/Meubles6.jpg'},
      {'label': 'خزانة ملابس','image': 'assets/Meubles7.jpg'},
      {'label': 'سرير',       'image': 'assets/Meubles8.jpg'},
    ],
    'projects': [
      {'label': 'مشروع 1', 'image': 'assets/auctions/selling_projects.png'},
      {'label': 'مشروع 2', 'image': 'assets/auctions/selling_projects.png'},
      {'label': 'مشروع 3', 'image': 'assets/auctions/selling_projects.png'},
    ],
  };

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
                // Top Categories Horizontal List (Photo-card style)
                const SizedBox(height: 14),
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
                            _selectedSubCategoryIndex = -1;
                            _filterAuctions();
                          });
                        },
                        child: Container(
                          width: 100,
                          margin: const EdgeInsets.only(right: 10, top: 4, bottom: 4),
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: isSelected ? primaryBlue : Colors.transparent,
                              width: 2.5,
                            ),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withValues(alpha: isSelected ? 0.15 : 0.07),
                                blurRadius: 8,
                                offset: const Offset(0, 4),
                              ),
                            ],
                          ),
                          child: Stack(
                            clipBehavior: Clip.none,
                            children: [
                              // Full image background
                              ClipRRect(
                                borderRadius: BorderRadius.circular(14),
                                child: Image.asset(
                                  cat['image'],
                                  width: 100,
                                  height: 110,
                                  fit: BoxFit.cover,
                                  errorBuilder: (c, e, s) => Container(
                                    color: isDarkMode ? const Color(0xFF2C2C2C) : const Color(0xFFEEEEEE),
                                    child: Center(child: Icon(Icons.category, color: isSelected ? primaryBlue : Colors.grey, size: 36)),
                                  ),
                                ),
                              ),
                              // Dark gradient overlay at bottom
                              ClipRRect(
                                borderRadius: BorderRadius.circular(14),
                                child: Container(
                                  decoration: BoxDecoration(
                                    gradient: LinearGradient(
                                      begin: Alignment.topCenter,
                                      end: Alignment.bottomCenter,
                                      colors: [
                                        Colors.transparent,
                                        Colors.black.withValues(alpha: 0.55),
                                      ],
                                      stops: const [0.4, 1.0],
                                    ),
                                  ),
                                ),
                              ),
                              // Title at bottom
                              Positioned(
                                bottom: 8,
                                left: 4,
                                right: 4,
                                child: Text(
                                  title,
                                  textAlign: TextAlign.center,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                              // Count badge top-right
                              Positioned(
                                top: -4,
                                right: -4,
                                child: Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
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

                // شريط العناصر العام — يظهر لكل فئة لها محتوى
                Builder(builder: (context) {
                  final catKey = _selectedCategoryIndex == 0
                      ? null
                      : _categories[_selectedCategoryIndex]['key'] as String;
                  final items = catKey != null ? (_categoryItems[catKey] ?? []) : [];
                  if (items.isEmpty) return const SizedBox.shrink();

                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: Text(
                            _getLocalizedTitle(_categories[_selectedCategoryIndex]['title_key'], l10n),
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
                          itemCount: items.length,
                          itemBuilder: (context, index) {
                            final item = items[index];
                            final bool isCars = catKey == 'cars';
                            final bool isSelected = isCars
                                ? _selectedSubCategoryIndex == index
                                : false;
                            final String imagePath = item['image'] as String;
                            final String label = isCars
                                ? _getLocalizedTitle(item['title_key'] as String, l10n)
                                : item['label'] as String;

                            return GestureDetector(
                              onTap: isCars
                                  ? () => setState(() {
                                        _selectedSubCategoryIndex = index;
                                        _filterAuctions();
                                      })
                                  : null,
                              child: Container(
                                width: 110,
                                margin: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                                decoration: BoxDecoration(
                                  color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(
                                    color: isSelected ? primaryBlue : Colors.grey.withValues(alpha: 0.15),
                                    width: isSelected ? 2 : 1,
                                  ),
                                  boxShadow: [
                                    BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 6, offset: const Offset(0, 3))
                                  ],
                                ),
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    ClipRRect(
                                      borderRadius: BorderRadius.circular(10),
                                      child: Image.asset(
                                        imagePath,
                                        height: 52,
                                        width: 90,
                                        fit: BoxFit.cover,
                                        errorBuilder: (c, e, s) => Icon(Icons.image_not_supported, color: Colors.grey, size: 30),
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      label,
                                      textAlign: TextAlign.center,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                      style: GoogleFonts.plusJakartaSans(
                                        fontSize: 11,
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
                  );
                }),

                // Custom Tabs — dynamic count
                const SizedBox(height: 16),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Builder(builder: (context) {
                    // Count active vs finished from all auctions (matching category filter)
                    int activeCount = _allAuctions.where((a) {
                      if (a['status'] != 'active') return false;
                      if (_selectedCategoryIndex != 0) {
                        return a['category'] == _categories[_selectedCategoryIndex]['key'];
                      }
                      return true;
                    }).length;
                    int finishedCount = _allAuctions.where((a) {
                      if (a['status'] != 'finished') return false;
                      if (_selectedCategoryIndex != 0) {
                        return a['category'] == _categories[_selectedCategoryIndex]['key'];
                      }
                      return true;
                    }).length;
                    return Row(
                      children: [
                        _buildStitchTab(1, "مزايدات منتهية", finishedCount.toString().padLeft(2, '0'), const Color(0xFFFFEDED), const Color(0xFFFF3B30)),
                        const SizedBox(width: 12),
                        _buildStitchTab(0, "مزايدات نشطة", activeCount.toString().padLeft(2, '0'), const Color(0xFFE8F5E9), const Color(0xFF2E7D32)),
                      ],
                    );
                  }),
                ),
                const SizedBox(height: 16),
              ],
            ),
          ),

          // Auction List
          _isLoading
              ? SliverFillRemaining(
                  child: Center(
                    child: CircularProgressIndicator(),
                  ),
                )
              : _filteredAuctions.isEmpty
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

  Widget _buildHorizontalAuctionCard(BuildContext context, bool isDarkMode, Map<String, dynamic> auction) {
    const Color primaryBlue = Color(0xFF0084FF);
    const Color softRed = Color(0xFFFF3B30);
    const Color lightGreyBg = Color(0xFFF2F2F7);
    const Color darkGreyBg = Color(0xFF2C2C2E);
    
    final id = auction['id']?.toString() ?? '';
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(id) ?? false;

    final bool isFinished = auction['status'] == 'finished';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
            spreadRadius: -5,
          ),
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.03),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: IntrinsicHeight(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Right: Image
            ClipRRect(
              borderRadius: const BorderRadius.only(
                topRight: Radius.circular(16),
                bottomRight: Radius.circular(16),
              ),
              child: SizedBox(
                width: 125,
                height: 110,
                child: Image.asset(
                  auction['images'] is List && (auction['images'] as List).isNotEmpty 
                      ? auction['images'][0] 
                      : auction['image']?.toString() ?? 'assets/car0.png',
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(
                    color: Colors.grey[200],
                    child: const Icon(Icons.image_not_supported, color: Colors.grey),
                  ),
                ),
              ),
            ),
            // Left: Content
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6), // Reduced from 10
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    // Title
                    Text(
                      auction['title']?.toString() ?? 'Sans titre',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 14,
                        fontWeight: FontWeight.w800,
                        color: isDarkMode ? Colors.white : Colors.black,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      textAlign: TextAlign.right,
                    ),
                    const SizedBox(height: 1),
                    // Price
                    Text(
                      "${auction['current_bid'] ?? auction['price'] ?? 0} MRU",
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: isDarkMode ? Colors.white70 : Colors.black87,
                      ),
                      textAlign: TextAlign.right,
                    ),
                    const SizedBox(height: 6), // Reduced from 10
                    // Interaction Row: [Timer] [Heart] [Bids+Gavel]
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        // Bid count + gavel (leftmost = endmost in RTL code)
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              auction['bids']?.toString() ?? '0',
                              style: GoogleFonts.plusJakartaSans(
                                color: softRed,
                                fontWeight: FontWeight.w900,
                                fontSize: 14,
                              ),
                            ),
                            const SizedBox(width: 3),
                            Icon(
                              Icons.gavel_rounded,
                              color: isDarkMode ? Colors.white54 : Colors.black54,
                              size: 16,
                            ),
                          ],
                        ),
                        // Heart / favorite
                        GestureDetector(
                          onTap: () => ref.read(favoritesProvider.notifier).toggleFavorite(id),
                          child: Icon(
                            isFavorite ? Icons.favorite : Icons.favorite_border,
                            color: isFavorite ? softRed : (isDarkMode ? Colors.white54 : Colors.grey),
                            size: 20, // Reduced from 22
                          ),
                        ),
                        // Timer (rightmost = startmost in RTL code)
                        Flexible(
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Flexible(
                                child: Text(
                                  isFinished ? 'انتهى المزاد' : (auction['ends_at']?.toString() ?? auction['time']?.toString() ?? ''),
                                  style: GoogleFonts.plusJakartaSans(
                                    color: softRed,
                                    fontSize: 11,
                                    fontWeight: FontWeight.w800,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              const SizedBox(width: 3),
                              const Icon(Icons.access_time_rounded, color: softRed, size: 14),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const Spacer(),
                    // Posted time + location
                    Directionality(
                      textDirection: TextDirection.rtl,
                      child: Text(
                        "${auction['postedTime']} . ${auction['location']}",
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 10, // Reduced from 11
                          color: const Color(0xFF9AA5B4),
                          fontWeight: FontWeight.w500,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
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
      case 'text_134': return l10n.text_134;
      case 'text_137': return l10n.text_137;
      case 'text_139': return l10n.text_139;
      default: return "Category";
    }
  }
}
