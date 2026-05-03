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
import 'package:mezadpay/pages/auction_details_page.dart';
import '../services/auction_api.dart';
import '../providers/category_provider.dart';
import '../services/category_api.dart';

class AllAuctionsPage extends ConsumerStatefulWidget {
  const AllAuctionsPage({super.key});

  @override
  ConsumerState<AllAuctionsPage> createState() => _AllAuctionsPageState();
}

class _AllAuctionsPageState extends ConsumerState<AllAuctionsPage> {
  // Constantes de couleur
  static const Color primaryBlue = Color(0xFF0084FF);

  final TextEditingController _searchController = TextEditingController();
  final AuctionApi _auctionApi = AuctionApi();

  List<Map<String, dynamic>> _allAuctions = [];
  List<Map<String, dynamic>> _filteredAuctions = [];
  List<Map<String, dynamic>> _categories = [];
  List<Map<String, dynamic>> _subCategories = [];
  Map<String, List<Map<String, dynamic>>> _subCategoriesMap = {};
  int _activeTabIndex = 0;
  int _selectedCategoryIndex = 0;
  int _selectedSubCategoryIndex = -1;
  bool _isLoading = true;
  bool _isLoadingCategories = true;
  String? _selectedCategoryId;

  // Pagination variables
  int _currentPage = 1;
  bool _hasMoreData = true;
  bool _isLoadingMore = false;
  final ScrollController _scrollController = ScrollController();

  // Advanced filter variables
  bool _showAdvancedFilter = false;
  double _minPrice = 0;
  double _maxPrice = 1000000;
  RangeValues _priceRange = const RangeValues(0, 1000000);
  String _sortBy = 'newest'; // newest, price_asc, price_desc, ending_soon
  DateTime? _startDate;
  DateTime? _endDate;

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchChanged);
    _scrollController.addListener(_onScroll);
    _loadCategories();
    _loadAuctions();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 200) {
      if (!_isLoadingMore && _hasMoreData) {
        _loadMoreAuctions();
      }
    }
  }

  Future<void> _loadCategories() async {
    try {
      final categoryApi = CategoryApi();
      final response = await categoryApi.getCategories();

      print('=== DEBUG CATEGORIES ===');
      print('Response success: ${response.success}');
      print('Response data type: ${response.data.runtimeType}');

      List<Map<String, dynamic>> parentCategories = [];
      Map<String, List<Map<String, dynamic>>> subCategoriesMap = {};

      if (response.success && response.data != null) {
        final categoriesList = response.data!;
        print('CategoriesList length: ${categoriesList.length}');

        // Separate parent categories and subcategories
        for (var item in categoriesList) {
          final cat = item as Map<String, dynamic>;
          final parentId = cat['parent_id'];
          print('Category: ${cat['name_ar']}, parent_id: $parentId');
          
          if (parentId == null) {
            // This is a parent category
            parentCategories.add({
              'id': cat['id']?.toString() ?? '',
              'name_ar': cat['name_ar'] ?? cat['name'] ?? '',
              'name_fr': cat['name_fr'] ?? cat['name'] ?? '',
              'name_en': cat['name_en'] ?? cat['name'] ?? '',
              'image': cat['image_url'] ?? cat['image'] ?? 'assets/auctions/other.png',
              'key': cat['key'] ?? cat['slug'] ?? cat['id']?.toString() ?? '',
              'count': cat['auction_count'] ?? cat['count'] ?? 0,
              'has_subcategories': cat['has_subcategories'] ?? false,
            });
          } else {
            // This is a subcategory, store it under its parent
            final parentKey = parentId.toString();
            if (!subCategoriesMap.containsKey(parentKey)) {
              subCategoriesMap[parentKey] = [];
            }
            subCategoriesMap[parentKey]!.add({
              'id': cat['id']?.toString() ?? '',
              'name_ar': cat['name_ar'] ?? cat['name'] ?? '',
              'name_fr': cat['name_fr'] ?? cat['name'] ?? '',
              'name_en': cat['name_en'] ?? cat['name'] ?? '',
              'image': cat['image_url'] ?? cat['image'] ?? 'assets/auctions/other.png',
              'key': cat['key'] ?? cat['slug'] ?? cat['id']?.toString() ?? '',
              'parent_id': parentKey,
            });
          }
        }
      }

      print('Parent categories count: ${parentCategories.length}');
      print('SubCategoriesMap keys: ${subCategoriesMap.keys}');

      setState(() {
        _categories = [
          {'id': 'all', 'name_ar': 'الكل', 'name_fr': 'Tout', 'name_en': 'All', 'image': 'assets/auctions/other.png', 'key': 'all'},
          ...parentCategories,
        ];
        // Store subcategories map for quick access
        _subCategoriesMap = subCategoriesMap;
        _isLoadingCategories = false;
      });
      print('Final _categories length: ${_categories.length}');
    } catch (e) {
      print('Error loading categories: $e');
      setState(() {
        _isLoadingCategories = false;
        // En cas d'erreur, on garde la catégorie "All" par défaut
        _categories = [
          {'id': 'all', 'name_ar': 'الكل', 'name_fr': 'Tout', 'name_en': 'All', 'image': 'assets/auctions/other.png', 'key': 'all'},
        ];
      });
    }
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
  }

  void _onSearchChanged() {
    _filterAuctions();
  }

  Future<void> _loadAuctions({bool reset = true}) async {
    if (reset) {
      setState(() {
        _isLoading = true;
        _currentPage = 1;
        _hasMoreData = true;
      });
    }

    try {
      final response = await _auctionApi.getAuctions(
        page: _currentPage,
        limit: 20,
        status: _activeTabIndex == 0 ? 'active' : 'completed',
        categoryId: _selectedCategoryId,
        minPrice: _minPrice > 0 ? _minPrice.toInt() : null,
        maxPrice: _maxPrice < 1000000 ? _maxPrice.toInt() : null,
        sortBy: _sortBy,
      );

      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
        if (response.success && response.data != null) {
          final dynamic responseData = response.data!;
          List<dynamic> auctionList = [];
          int? totalCount;

          // La réponse peut être directement une liste ou un objet avec 'data' ou 'auctions'
          if (responseData is List) {
            auctionList = responseData;
          } else if (responseData is Map<String, dynamic>) {
            auctionList = (responseData['auctions'] ?? responseData['data'] ?? []) as List<dynamic>;
            totalCount = responseData['total'] ?? responseData['count'];
          }

          final newAuctions = auctionList.map((item) => item as Map<String, dynamic>).toList();

          if (reset) {
            _allAuctions = newAuctions;
          } else {
            _allAuctions.addAll(newAuctions);
          }

          _filteredAuctions = List.from(_allAuctions);

          // Vérifier s'il y a plus de données
          if (newAuctions.length < 20 || (totalCount != null && _allAuctions.length >= totalCount)) {
            _hasMoreData = false;
          }
        }
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_loading_auctions)),
        );
      }
    }
  }

  Future<void> _loadMoreAuctions() async {
    if (_isLoadingMore || !_hasMoreData) return;

    setState(() {
      _isLoadingMore = true;
      _currentPage++;
    });

    await _loadAuctions(reset: false);
  }

  void _filterAuctions() {
    setState(() {
      _filteredAuctions = _allAuctions.where((auction) {
        final title = auction['title_ar'] ?? auction['title'] ?? '';
        final matchesSearch = title.toLowerCase().contains(_searchController.text.toLowerCase());
        final matchesStatus = _activeTabIndex == 0 ? auction['status'] == 'active' : auction['status'] == 'finished';

        bool matchesCategory = true;
        if (_selectedCategoryIndex != 0 && _categories.isNotEmpty) {
          final categoryId = _categories[_selectedCategoryIndex]['id']?.toString();
          final auctionCategoryId = auction['category_id']?.toString() ?? auction['category']?.toString();
          if (categoryId != null && categoryId != 'all') {
            matchesCategory = auctionCategoryId == categoryId;
          }
        }

        bool matchesSubCategory = true;
        if (_selectedSubCategoryIndex != -1 && _subCategories.isNotEmpty) {
          final subCategoryId = _subCategories[_selectedSubCategoryIndex]['id']?.toString();
          final auctionSubCategoryId = auction['sub_category_id']?.toString() ?? auction['subCategory']?.toString();
          if (subCategoryId != null) {
            matchesSubCategory = auctionSubCategoryId == subCategoryId;
          }
        }

        // Filtre par prix
        bool matchesPrice = true;
        final price = (auction['current_price'] ?? auction['current_bid'] ?? auction['price'] ?? 0);
        final priceValue = price is num ? price.toDouble() : double.tryParse(price.toString()) ?? 0;
        if (priceValue < _minPrice || priceValue > _maxPrice) {
          matchesPrice = false;
        }

        // Filtre par date de début/fin
        bool matchesDate = true;
        if (_startDate != null || _endDate != null) {
          final auctionStart = auction['start_time'] != null ? DateTime.tryParse(auction['start_time']) : null;
          final auctionEnd = auction['end_time'] != null ? DateTime.tryParse(auction['end_time']) : null;
          
          if (_startDate != null && auctionStart != null && auctionStart.isBefore(_startDate!)) {
            matchesDate = false;
          }
          if (_endDate != null && auctionEnd != null && auctionEnd.isAfter(_endDate!)) {
            matchesDate = false;
          }
        }

        return matchesSearch && matchesStatus && matchesCategory && matchesSubCategory && matchesPrice && matchesDate;
      }).toList();
    });
  }

  void _showAdvancedFilterSheet() {
    final locale = Localizations.localeOf(context).languageCode;
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              height: MediaQuery.of(context).size.height * 0.75,
              decoration: BoxDecoration(
                color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
              ),
              child: Column(
                children: [
                  // Header
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      border: Border(
                        bottom: BorderSide(
                          color: isDarkMode ? Colors.grey.shade800 : Colors.grey.shade200,
                        ),
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          locale == 'ar' ? 'تصفية متقدمة' : (locale == 'fr' ? 'Filtre avancé' : 'Advanced Filter'),
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: isDarkMode ? Colors.white : Colors.black,
                          ),
                        ),
                        IconButton(
                          icon: const Icon(Icons.close),
                          onPressed: () => Navigator.pop(context),
                        ),
                      ],
                    ),
                  ),
                  // Filter content
                  Expanded(
                    child: SingleChildScrollView(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // Prix
                          Text(
                            locale == 'ar' ? 'السعر (MRU)' : (locale == 'fr' ? 'Prix (MRU)' : 'Price (MRU)'),
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: isDarkMode ? Colors.white : Colors.black,
                            ),
                          ),
                          const SizedBox(height: 8),
                          RangeSlider(
                            values: _priceRange,
                            min: 0,
                            max: 1000000,
                            divisions: 100,
                            labels: RangeLabels(
                              '${_priceRange.start.toInt()} MRU',
                              '${_priceRange.end.toInt()} MRU',
                            ),
                            onChanged: (values) {
                              setModalState(() {
                                _priceRange = values;
                                _minPrice = values.start;
                                _maxPrice = values.end;
                              });
                            },
                          ),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                '${_priceRange.start.toInt()} MRU',
                                style: TextStyle(color: isDarkMode ? Colors.grey.shade400 : Colors.grey),
                              ),
                              Text(
                                '${_priceRange.end.toInt()} MRU',
                                style: TextStyle(color: isDarkMode ? Colors.grey.shade400 : Colors.grey),
                              ),
                            ],
                          ),
                          const SizedBox(height: 24),

                          // Tri
                          Text(
                            locale == 'ar' ? 'ترتيب حسب' : (locale == 'fr' ? 'Trier par' : 'Sort by'),
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: isDarkMode ? Colors.white : Colors.black,
                            ),
                          ),
                          const SizedBox(height: 8),
                          _buildSortOption(
                            setModalState,
                            'newest',
                            locale == 'ar' ? 'الأحدث' : (locale == 'fr' ? 'Plus récent' : 'Newest'),
                            Icons.access_time,
                          ),
                          _buildSortOption(
                            setModalState,
                            'price_asc',
                            locale == 'ar' ? 'السعر: من الأقل للأعلى' : (locale == 'fr' ? 'Prix: croissant' : 'Price: low to high'),
                            Icons.arrow_upward,
                          ),
                          _buildSortOption(
                            setModalState,
                            'price_desc',
                            locale == 'ar' ? 'السعر: من الأعلى للأقل' : (locale == 'fr' ? 'Prix: décroissant' : 'Price: high to low'),
                            Icons.arrow_downward,
                          ),
                          _buildSortOption(
                            setModalState,
                            'ending_soon',
                            locale == 'ar' ? 'تنتهي قريباً' : (locale == 'fr' ? 'Termine bientôt' : 'Ending soon'),
                            Icons.timer,
                          ),
                        ],
                      ),
                    ),
                  ),
                  // Actions
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      border: Border(
                        top: BorderSide(
                          color: isDarkMode ? Colors.grey.shade800 : Colors.grey.shade200,
                        ),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: OutlinedButton(
                            onPressed: () {
                              setModalState(() {
                                _priceRange = const RangeValues(0, 1000000);
                                _minPrice = 0;
                                _maxPrice = 1000000;
                                _sortBy = 'newest';
                              });
                            },
                            style: OutlinedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              side: BorderSide(color: primaryBlue),
                            ),
                            child: Text(
                              locale == 'ar' ? 'إعادة' : (locale == 'fr' ? 'Réinitialiser' : 'Reset'),
                              style: TextStyle(color: primaryBlue),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton(
                            onPressed: () {
                              Navigator.pop(context);
                              _loadAuctions();
                            },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: primaryBlue,
                              padding: const EdgeInsets.symmetric(vertical: 12),
                            ),
                            child: Text(
                              locale == 'ar' ? 'تطبيق' : (locale == 'fr' ? 'Appliquer' : 'Apply'),
                              style: const TextStyle(color: Colors.white),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildSortOption(StateSetter setModalState, String value, String label, IconData icon) {
    final isSelected = _sortBy == value;
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return GestureDetector(
      onTap: () {
        setModalState(() {
          _sortBy = value;
        });
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF0084FF).withOpacity(0.1) : (isDarkMode ? Colors.grey.shade800 : Colors.grey.shade100),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(
            color: isSelected ? const Color(0xFF0084FF) : Colors.transparent,
            width: 1.5,
          ),
        ),
        child: Row(
          children: [
            Icon(
              icon,
              size: 20,
              color: isSelected ? const Color(0xFF0084FF) : (isDarkMode ? Colors.grey.shade400 : Colors.grey),
            ),
            const SizedBox(width: 12),
            Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 14,
                fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                color: isSelected ? const Color(0xFF0084FF) : (isDarkMode ? Colors.white : Colors.black87),
              ),
            ),
            const Spacer(),
            if (isSelected)
              const Icon(
                Icons.check_circle,
                color: Color(0xFF0084FF),
                size: 20,
              ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      endDrawer: const SideMenuDrawer(),
      appBar: AppBar(
        backgroundColor: primaryBlue,
        elevation: 0,
        toolbarHeight: 60,
        centerTitle: true,
        title: Text(
          _getAuctionTypesTitle(context),
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
        controller: _scrollController,
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
                      String title = _getCategoryName(cat);

                      return GestureDetector(
                        onTap: () {
                          setState(() {
                            _selectedCategoryIndex = index;
                            _selectedSubCategoryIndex = -1;
                            _selectedCategoryId = cat['id']?.toString();
                          });
                          // Load subcategories from the map if this is a parent category
                          if (cat['id']?.toString() != 'all') {
                            final categoryId = cat['id']?.toString() ?? '';
                            setState(() {
                              _subCategories = _subCategoriesMap[categoryId] ?? [];
                            });
                          } else {
                            setState(() => _subCategories = []);
                          }
                          // Recharger les enchères depuis le serveur avec le nouveau filtre
                          _loadAuctions();
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
                                child: SizedBox(
                                  width: 100,
                                  height: 110,
                                  child: _buildAuctionImage(cat['image']?.toString()),
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

                // Search Bar with Filter Button
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 12.0),
                  child: Row(
                    children: [
                      // Search Field
                      Expanded(
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
                              hintText: _getSearchHint(context),
                              hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 15),
                              prefixIcon: const Icon(Icons.search, color: Colors.grey, size: 22),
                              border: InputBorder.none,
                              contentPadding: const EdgeInsets.symmetric(vertical: 12),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      // Filter Button
                      GestureDetector(
                        onTap: _showAdvancedFilterSheet,
                        child: Container(
                          height: 48,
                          width: 48,
                          decoration: BoxDecoration(
                            color: (_minPrice > 0 || _maxPrice < 1000000 || _sortBy != 'newest')
                                ? primaryBlue
                                : (isDarkMode ? const Color(0xFF1D1D1D) : Colors.white),
                            borderRadius: BorderRadius.circular(24),
                            border: Border.all(color: Colors.grey.withValues(alpha: 0.1)),
                            boxShadow: [
                              BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 10, offset: const Offset(0, 4))
                            ],
                          ),
                          child: Icon(
                            Icons.tune,
                            color: (_minPrice > 0 || _maxPrice < 1000000 || _sortBy != 'newest')
                                ? Colors.white
                                : Colors.grey,
                            size: 22,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),

                // شريط العناصر العام — يظهر لكل فئة لها محتوى
                Builder(builder: (context) {
                  // Utiliser les sous-catégories de l'API
                  final items = _subCategories;
                  if (items.isEmpty) return const SizedBox.shrink();

                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: Text(
                            _getCategoryName(_categories[_selectedCategoryIndex]),
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
                            final bool isSelected = _selectedSubCategoryIndex == index;
                            final String label = _getCategoryName(item);

                            return GestureDetector(
                              onTap: () => setState(() {
                                _selectedSubCategoryIndex = index;
                                _filterAuctions();
                              }),
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
                                      child: SizedBox(
                                        height: 52,
                                        width: 90,
                                        child: _buildAuctionImage(item['image']?.toString()),
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
                    // Labels dynamiques selon la langue
                    final locale = Localizations.localeOf(context).languageCode;
                    String activeLabel, finishedLabel;
                    switch (locale) {
                      case 'ar':
                        activeLabel = 'مزايدات نشطة';
                        finishedLabel = 'مزايدات منتهية';
                        break;
                      case 'fr':
                        activeLabel = 'Enchères actives';
                        finishedLabel = 'Enchères terminées';
                        break;
                      case 'en':
                      default:
                        activeLabel = 'Active Auctions';
                        finishedLabel = 'Finished Auctions';
                        break;
                    }
                    
                    return Row(
                      children: [
                        _buildStitchTab(1, finishedLabel, finishedCount.toString().padLeft(2, '0'), const Color(0xFFFFEDED), const Color(0xFFFF3B30)),
                        const SizedBox(width: 12),
                        _buildStitchTab(0, activeLabel, activeCount.toString().padLeft(2, '0'), const Color(0xFFE8F5E9), const Color(0xFF2E7D32)),
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
          // Loading indicator for pagination
          if (_isLoadingMore)
            SliverToBoxAdapter(
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: const Center(
                  child: CircularProgressIndicator(),
                ),
              ),
            ),
          // End of list message
          if (!_hasMoreData && _filteredAuctions.isNotEmpty && !_isLoading)
            SliverToBoxAdapter(
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: Text(
                    Localizations.localeOf(context).languageCode == 'ar'
                        ? 'لا توجد مزادات أخرى'
                        : (Localizations.localeOf(context).languageCode == 'fr'
                            ? 'Plus d\'enchères'
                            : 'No more auctions'),
                    style: GoogleFonts.plusJakartaSans(
                      color: isDarkMode ? Colors.grey.shade400 : Colors.grey,
                      fontSize: 14,
                    ),
                  ),
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
          });
          // Recharger les enchères depuis le serveur avec le nouveau statut
          _loadAuctions();
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
                Flexible(
                  child: Text(
                    label,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                      color: isSelected ? textColor : Colors.grey,
                    ),
                    overflow: TextOverflow.ellipsis,
                    maxLines: 1,
                  ),
                ),
                const SizedBox(width: 6),
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
    const Color softRed = Color(0xFFFF3B30);
    const Color lightGreyBg = Color(0xFFF2F2F7);
    const Color darkGreyBg = Color(0xFF2C2C2E);

    final id = auction['id']?.toString() ?? '';
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(id) ?? false;

    final bool isFinished = auction['status'] == 'finished';

    // Récupérer le titre selon la langue actuelle avec fallback intelligent
    final locale = Localizations.localeOf(context).languageCode;
    String title = '';

    // 1. Essayer d'abord la langue actuelle de l'app
    switch (locale) {
      case 'ar':
        title = auction['title_ar']?.toString() ?? '';
        break;
      case 'fr':
        title = auction['title_fr']?.toString() ?? '';
        break;
      case 'en':
        title = auction['title_en']?.toString() ?? '';
        break;
    }

    // 2. Si vide, essayer l'arabe (langue par défaut de l'app)
    if (title.isEmpty) {
      title = auction['title_ar']?.toString() ?? '';
    }

    // 3. Si toujours vide, essayer les autres langues dans l'ordre
    if (title.isEmpty) {
      title = auction['title_fr']?.toString() ??
              auction['title_en']?.toString() ??
              auction['title']?.toString() ??
              '';
    }

    // 4. Si toujours vide, afficher "Sans titre"
    if (title.isEmpty) title = _getNoTitleText(context);

    // Obtenir l'URL de l'image
    String? imageUrl;
    if (auction['images'] is List && (auction['images'] as List).isNotEmpty) {
      final firstImage = (auction['images'] as List)[0];
      if (firstImage is Map<String, dynamic>) {
        imageUrl = firstImage['url']?.toString() ?? firstImage['image_url']?.toString();
      } else {
        imageUrl = firstImage.toString();
      }
    } else if (auction['image'] != null) {
      imageUrl = auction['image'].toString();
    } else if (auction['image_url'] != null) {
      imageUrl = auction['image_url'].toString();
    }

    return GestureDetector(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => AuctionDetailsPage(auctionId: id),
          ),
        );
      },
      child: Container(
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
                  child: _buildAuctionImage(imageUrl),
                ),
              ),
              // Left: Content
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      // Title
                      Text(
                        title,
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
                      "${auction['current_price'] ?? auction['current_bid'] ?? auction['price'] ?? 0} MRU",
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
                              (auction['bidder_count'] ?? auction['bids'] ?? 0).toString(),
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
                                child: Builder(builder: (context) {
                                  // Formater la date si disponible
                                  String timeText = '0';
                                  final rawTime = auction['end_time'] ?? auction['ends_at'] ?? auction['time'];
                                  if (rawTime != null) {
                                    try {
                                      final dateTime = DateTime.parse(rawTime.toString());
                                      // Format: DD/MM/YYYY HH:MM
                                      final day = dateTime.day.toString().padLeft(2, '0');
                                      final month = dateTime.month.toString().padLeft(2, '0');
                                      final year = dateTime.year.toString();
                                      final hour = dateTime.hour.toString().padLeft(2, '0');
                                      final minute = dateTime.minute.toString().padLeft(2, '0');
                                      timeText = '$day/$month/$year\t$hour:$minute';
                                    } catch (e) {
                                      // Si parsing échoue, afficher la valeur brute
                                      timeText = rawTime.toString();
                                    }
                                  }
                                  
                                  return Text(
                                    isFinished ? 'انتهى المزاد' : timeText,
                                    style: GoogleFonts.plusJakartaSans(
                                      color: softRed,
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                    ),
                                    overflow: TextOverflow.ellipsis,
                                  );
                                }),
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
                      child: Builder(builder: (context) {
                        // Récupérer la ville avec fallback intelligent
                        final locale = Localizations.localeOf(context).languageCode;
                        String cityName = '';
                        
                        // 1. Essayer d'abord la langue actuelle de l'app
                        switch (locale) {
                          case 'ar':
                            cityName = auction['city_name_ar']?.toString() ?? auction['city_ar']?.toString() ?? '';
                            break;
                          case 'fr':
                            cityName = auction['city_name_fr']?.toString() ?? auction['city_fr']?.toString() ?? '';
                            break;
                          case 'en':
                            cityName = auction['city_name_en']?.toString() ?? auction['city_en']?.toString() ?? '';
                            break;
                        }
                        
                        // 2. Si vide, essayer l'arabe (langue par défaut)
                        if (cityName.isEmpty) {
                          cityName = auction['city_name_ar']?.toString() ?? auction['city_ar']?.toString() ?? '';
                        }
                        
                        // 3. Si toujours vide, essayer les autres langues
                        if (cityName.isEmpty) {
                          cityName = auction['city_name_fr']?.toString() ??
                                     auction['city_name_en']?.toString() ??
                                     auction['city']?.toString() ??
                                     auction['location']?.toString() ??
                                     '';
                        }
                        
                        return Text(
                          "$cityName . ${(auction['category'] ?? '').toString().split(' ').first}",
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 10, // Reduced from 11
                            color: const Color(0xFF9AA5B4),
                            fontWeight: FontWeight.w500,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        );
                      }),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
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

  Widget _buildAuctionImage(String? imageUrl) {
    if (imageUrl == null || imageUrl.isEmpty) {
      return Container(
        color: Colors.grey[200],
        child: const Icon(Icons.image_not_supported, color: Colors.grey),
      );
    }

    // Si c'est une URL réseau
    if (imageUrl.startsWith('http')) {
      return Image.network(
        imageUrl,
        fit: BoxFit.cover,
        errorBuilder: (c, e, s) => Container(
          color: Colors.grey[200],
          child: const Icon(Icons.image_not_supported, color: Colors.grey),
        ),
        loadingBuilder: (context, child, loadingProgress) {
          if (loadingProgress == null) return child;
          return Container(
            color: Colors.grey[200],
            child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
          );
        },
      );
    }

    // Sinon c'est un asset local
    return Image.asset(
      imageUrl,
      fit: BoxFit.cover,
      errorBuilder: (c, e, s) => Container(
        color: Colors.grey[200],
        child: const Icon(Icons.image_not_supported, color: Colors.grey),
      ),
    );
  }

  String _getNoTitleText(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'بدون عنوان';
      case 'fr':
        return 'Sans titre';
      case 'en':
      default:
        return 'No title';
    }
  }

  String _getFinishedText(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'انتهى المزاد';
      case 'fr':
        return 'Enchère terminée';
      case 'en':
      default:
        return 'Auction ended';
    }
  }

  String _getAuctionTypesTitle(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'أنواع المزادات';
      case 'fr':
        return 'Types d\'enchères';
      case 'en':
      default:
        return 'Auction Types';
    }
  }

  String _getSearchHint(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'البحث';
      case 'fr':
        return 'Rechercher';
      case 'en':
      default:
        return 'Search';
    }
  }

  String _getCategoryName(Map<String, dynamic> category) {
    final locale = Localizations.localeOf(context).languageCode;
    String name = '';

    // 1. Essayer d'abord la langue actuelle de l'app
    switch (locale) {
      case 'ar':
        name = category['name_ar']?.toString() ?? '';
        break;
      case 'fr':
        name = category['name_fr']?.toString() ?? '';
        break;
      case 'en':
        name = category['name_en']?.toString() ?? '';
        break;
    }

    // 2. Si vide, essayer l'arabe (langue par défaut)
    if (name.isEmpty) {
      name = category['name_ar']?.toString() ?? '';
    }

    // 3. Si toujours vide, essayer les autres langues
    if (name.isEmpty) {
      name = category['name_fr']?.toString() ??
             category['name_en']?.toString() ??
             category['name']?.toString() ??
             '';
    }

    return name;
  }

  String _getLocalizedTitle(String key, AppLocalizations l10n) {
    // Si c'est une catégorie de l'API (commence par name_)
    if (key.startsWith('name_')) {
      return key;
    }
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
