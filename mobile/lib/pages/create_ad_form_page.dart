import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb, Uint8List, debugPrint;
import 'package:image_picker/image_picker.dart';

import 'auction_pending_approval_page.dart';
import '../services/auction_api.dart';
import '../services/r2_upload_service.dart';
import '../services/category_api.dart';
import '../services/cache_service.dart';
import '../models/category.dart';

class CreateAdFormPage extends StatefulWidget {
  const CreateAdFormPage({super.key});

  @override
  State<CreateAdFormPage> createState() => _CreateAdFormPageState();
}

class _CreateAdFormPageState extends State<CreateAdFormPage> {
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _phoneController = TextEditingController();
  final _priceController = TextEditingController();
  
  String? _selectedMainCategory;     // display name (nameAr)
  int?    _selectedMainCategoryId;   // actual ID sent to API
  String? _selectedSubCategory;      // display name (nameAr)
  int?    _selectedSubCategoryId;    // actual ID sent to API
  String? _selectedCity;
  DateTime? _endTime;
  final _brandController = TextEditingController();
  // Backend accepte: new, used, refurbished, damaged
  String? _selectedCondition;
  static const List<Map<String, String>> _conditionOptions = [
    {'value': 'new', 'label': 'جديد'},
    {'value': 'used', 'label': 'مستعمل'},
    {'value': 'refurbished', 'label': 'مجدد'},
    {'value': 'damaged', 'label': 'تالف'},
  ];
  final List<XFile> _selectedImageFiles = []; // XFile works on web + mobile
  final AuctionApi _auctionApi = AuctionApi();
  final R2UploadService _r2UploadService = R2UploadService();
  final CategoryApi _categoryApi = CategoryApi();
  final ImagePicker _imagePicker = ImagePicker();
  bool _isLoading = false;
  bool _isUploadingImages = false;
  double _uploadProgress = 0.0;
  
  // Données depuis le cache/API
  List<Category> _categories = [];
  List<Location> _locations = [];
  @override
  void initState() {
    super.initState();
    _loadFromCache(); // Charger immédiatement depuis le cache
  }
  
  // Charger depuis le cache d'abord (rapide), puis fetch en arrière-plan
  Future<void> _loadFromCache() async {
    try {
      // Charger catégories depuis cache (le fetch en arrière-plan ci-dessous
      // rafraîchira le cache si les données sont périmées ou absentes)
      final cachedCategories = await CacheService.instance.getCachedCategories();
      if (cachedCategories != null && cachedCategories.isNotEmpty) {
        setState(() {
          _categories = cachedCategories
              .map((c) => Category.fromJson(c))
              .toList();
          if (_selectedMainCategory == null && _categories.isNotEmpty) {
            _selectedMainCategory = _categories.first.nameAr;
          }
        });
      }
      
      // Charger villes depuis cache
      final cachedCities = await CacheService.instance.getCachedCities();
      if (cachedCities != null && cachedCities.isNotEmpty) {
        setState(() {
          _locations = cachedCities
              .map((c) => Location.fromJson(c))
              .toList();
          if (_selectedCity == null && _locations.isNotEmpty) {
            _selectedCity = _locations.first.cityNameAr;
          }
        });
      }
    } catch (e) {
      debugPrint('Erreur chargement cache: $e');
    }
    
    // Fetch en arrière-plan pour mettre à jour le cache
    _fetchFromApi();
  }
  
  // Fetch depuis l'API en arrière-plan
  Future<void> _fetchFromApi() async {
    try {
      // Fetch categories
      final catResponse = await _categoryApi.getCategories();
      if (catResponse.success && catResponse.data != null) {
        // API returns List directly
        final categoriesData = catResponse.data!;
        final categories = categoriesData
            .map((c) => Category.fromJson(c as Map<String, dynamic>))
            .toList();
        setState(() {
          _categories = categories;
          if (_selectedMainCategory == null && _categories.isNotEmpty) {
            _selectedMainCategory = _categories.first.nameAr;
          }
        });
        // Sauvegarder dans le cache
        await CacheService.instance.cacheCategories(
          categories.map((c) => c.toJson()).toList()
        );
      }
      
      // Fetch locations
      final locResponse = await _categoryApi.getLocations();
      if (locResponse.success && locResponse.data != null) {
        // API returns List directly
        final locationsData = locResponse.data!;
        final locations = locationsData
            .map((l) => Location.fromJson(l as Map<String, dynamic>))
            .toList();
        setState(() {
          _locations = locations;
          if (_selectedCity == null && _locations.isNotEmpty) {
            _selectedCity = _locations.first.cityNameAr;
          }
        });
        // Sauvegarder dans le cache
        await CacheService.instance.cacheCities(
          locations.map((l) => l.toJson()).toList()
        );
      }
    } catch (e) {
      debugPrint('Erreur fetch API: $e');
    }
  }
  
  // Helper pour obtenir le nom en arabe
  Future<void> _submitAd() async {
    final name = _nameController.text.trim();
    final description = _descriptionController.text.trim();
    final phone = _phoneController.text.trim();
    final price = double.tryParse(_priceController.text);

    if (name.isEmpty || description.isEmpty || phone.isEmpty || price == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_fill_required_fields)),
      );
      return;
    }

    if (_selectedImageFiles.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_add_image)),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      // Étape 1: Créer l'enchère sans images
      final createResponse = await _auctionApi.createAuction(
        title: name,
        description: description,
        startingPrice: price,
        category: _selectedMainCategory ?? '',
        subCategory: _selectedSubCategory ?? '',
        categoryId: _selectedMainCategoryId,
        subCategoryId: _selectedSubCategoryId,
        location: _selectedCity ?? '',
        images: [], // Créer d'abord sans images
        phone: phone,
        condition: _selectedCondition,
        brand: _brandController.text.trim().isEmpty ? null : _brandController.text.trim(),
        endTime: _endTime,
      );

      if (!createResponse.success || createResponse.data == null) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(createResponse.message ?? AppLocalizations.of(context)!.error_create_auction)),
        );
        return;
      }

      // Étape 2: Récupérer l'ID de l'enchère créée
      final auctionId = createResponse.data!['id'] as String?;
      if (auctionId == null) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_create_auction)),
        );
        return;
      }

      // Étape 3: Uploader les images vers R2
      setState(() {
        _isUploadingImages = true;
        _uploadProgress = 0.0;
      });

      debugPrint('[Upload] auction_id=$auctionId images=${_selectedImageFiles.length}');
      final uploadedUrls = await _r2UploadService.uploadAuctionXFiles(
        auctionId: auctionId,
        images: _selectedImageFiles,
        onProgress: (progress) {
          setState(() => _uploadProgress = progress);
        },
      );
      debugPrint('[Upload] uploaded=${uploadedUrls.length} urls=$uploadedUrls');

      if (!mounted) return;
      setState(() {
        _isUploadingImages = false;
        _isLoading = false;
      });

      if (uploadedUrls.isEmpty) {
        // L'enchère est créée mais sans images
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Enchère créée mais les images n\'ont pas pu être uploadées')),
        );
      }

      final response = createResponse;

      if (!mounted) return;
      setState(() => _isLoading = false);

      if (response.success) {
        Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => const AuctionPendingApprovalPage()),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(response.message ?? AppLocalizations.of(context)!.error_create_auction)),
        );
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_connection)),
      );
    }
  }

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
            AppLocalizations.of(context)!.text_89,
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
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                AppLocalizations.of(context)!.text_90,
                style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
              ),
              const SizedBox(height: 32),
              
              _buildLabel(context, AppLocalizations.of(context)!.text_91),
              _buildTextField(context, controller: _nameController, hint: AppLocalizations.of(context)!.text_92),
              
              const SizedBox(height: 24),
              _buildLabel(context, AppLocalizations.of(context)!.text_93),
              _buildTextField(context, controller: _descriptionController, hint: AppLocalizations.of(context)!.text_94, maxLines: 3, icon: Icons.description_outlined),
              
              const SizedBox(height: 24),
              _buildLabel(context, AppLocalizations.of(context)!.text_40),
              _buildTextField(context, controller: _phoneController, hint: AppLocalizations.of(context)!.text_95, icon: Icons.phone_outlined, keyboardType: TextInputType.phone),
              
              const SizedBox(height: 24),
              _buildLabel(context, AppLocalizations.of(context)!.text_96),
              _buildTextField(context, controller: _priceController, hint: AppLocalizations.of(context)!.text_97, icon: Icons.attach_money_outlined, keyboardType: TextInputType.number, suffixText: 'MRU'),
              
              const SizedBox(height: 24),
              _buildLabel(context, AppLocalizations.of(context)!.text_98),
              _buildCategorySelector(context, value: _selectedMainCategory, icon: Icons.category_outlined, onTap: () => _showCategorySheet(context, true)),
              
              const SizedBox(height: 24),
              _buildLabel(context, AppLocalizations.of(context)!.text_99),
              _buildCategorySelector(context, value: _selectedSubCategory, icon: Icons.grid_view_outlined, onTap: () => _showCategorySheet(context, false)),
              
              const SizedBox(height: 24),
              _buildLabel(context, AppLocalizations.of(context)!.text_100),
              _buildCategorySelector(context, value: _selectedCity, icon: Icons.location_on_outlined, onTap: () => _showCitySheet(context)),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'حالة المنتج'),
              _buildConditionSelector(context),

              const SizedBox(height: 24),
              _buildLabel(context, 'الماركة (اختياري)'),
              _buildTextField(context, controller: _brandController, hint: 'مثال: Samsung، Toyota...', icon: Icons.local_offer_outlined),

              const SizedBox(height: 24),
              _buildLabel(context, 'تاريخ انتهاء المزاد'),
              _buildDateSelector(context),
              
              const SizedBox(height: 32),
              _buildLabel(context, AppLocalizations.of(context)!.text_101),
              _buildMediaSection(context),
              
              const SizedBox(height: 48),
              
              // Submit Button
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _submitAd,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0081FF),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  child: _isLoading
                      ? const CircularProgressIndicator(color: Colors.white)
                      : Text(
                          AppLocalizations.of(context)!.text_102,
                          style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                ),
              ),
              const SizedBox(height: 24),
            ],
          ),
        ),
      );
  }

  Widget _buildLabel(BuildContext context, String text) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 8, end: 4),
      child: Row(
        children: [
          Text(
            text,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white70 : Colors.black87,
            ),
          ),
          const SizedBox(width: 4),
          const Text('*', style: TextStyle(color: Colors.red)),
        ],
      ),
    );
  }

  Widget _buildTextField(BuildContext context, {
    required TextEditingController controller,
    required String hint,
    int maxLines = 1,
    IconData? icon,
    TextInputType? keyboardType,
    String? suffixText,
  }) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.withOpacity(0.2)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: TextField(
        controller: controller,
        maxLines: maxLines,
        keyboardType: keyboardType,
        textAlign: TextAlign.right,
        decoration: InputDecoration(
          hintText: hint,
          hintStyle: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey, fontSize: 13),
          contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          border: InputBorder.none,
          prefixIcon: icon != null ? Icon(icon, color: const Color(0xFF0081FF).withOpacity(0.5), size: 20) : null,
          suffixText: suffixText,
          suffixStyle: const TextStyle(color: Colors.grey, fontWeight: FontWeight.bold),
        ),
      ),
    );
  }

  Widget _buildDateSelector(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: () async {
        final DateTime? picked = await showDatePicker(
          context: context,
          initialDate: _endTime ?? DateTime.now().add(const Duration(days: 7)),
          firstDate: DateTime.now(),
          lastDate: DateTime.now().add(const Duration(days: 365)),
          builder: (context, child) {
            return Theme(
              data: Theme.of(context).copyWith(
                colorScheme: isDarkMode
                    ? const ColorScheme.dark(primary: Color(0xFF0081FF))
                    : const ColorScheme.light(primary: Color(0xFF0081FF)),
              ),
              child: child!,
            );
          },
        );
        if (picked != null) {
          setState(() => _endTime = picked);
        }
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFF0081FF).withOpacity(0.1)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          children: [
            Icon(Icons.calendar_today_outlined, color: const Color(0xFF0081FF).withOpacity(0.5), size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                _endTime != null
                    ? '${_endTime!.day}/${_endTime!.month}/${_endTime!.year}'
                    : 'اختر تاريخ الانتهاء...',
                style: TextStyle(fontFamily: 'Plus Jakarta Sans',
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
              ),
            ),
            const Icon(Icons.arrow_drop_down, color: Colors.grey),
          ],
        ),
      ),
    );
  }

  Widget _buildConditionSelector(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: _conditionOptions.map((opt) {
        final selected = _selectedCondition == opt['value'];
        return ChoiceChip(
          label: Text(opt['label']!),
          selected: selected,
          onSelected: (_) => setState(() => _selectedCondition = opt['value']),
          selectedColor: const Color(0xFF0081FF),
          labelStyle: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontWeight: FontWeight.w600,
            color: selected ? Colors.white : (isDarkMode ? Colors.white70 : Colors.black87),
          ),
          backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          side: BorderSide(color: const Color(0xFF0081FF).withOpacity(selected ? 0 : 0.2)),
        );
      }).toList(),
    );
  }

  Widget _buildCategorySelector(BuildContext context, {required String? value, required IconData icon, required VoidCallback onTap}) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFF0081FF).withOpacity(0.1)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          children: [
             Icon(icon, color: const Color(0xFF0081FF).withOpacity(0.5), size: 20),
             const SizedBox(width: 12),
             Expanded(
               child: Text(
                 value ?? 'اختر...',
                 style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                   fontSize: 14,
                   fontWeight: FontWeight.w600,
                   color: isDarkMode ? Colors.white : Colors.black,
                 ),
               ),
             ),
             const Icon(Icons.arrow_drop_down, color: Colors.grey),
          ],
        ),
      ),
    );
  }

  Widget _buildMediaSection(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return SizedBox(
      height: 100,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        itemCount: _selectedImageFiles.length + 1,
        itemBuilder: (context, index) {
          if (index == _selectedImageFiles.length) {
            return GestureDetector(
              onTap: () => _pickImages(),
              child: Container(
                width: 100,
                margin: const EdgeInsetsDirectional.only(start: 12),
                decoration: BoxDecoration(
                  color: isDarkMode ? Colors.white10 : Colors.grey[100],
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFF0081FF).withOpacity(0.2), style: BorderStyle.solid),
                ),
                child: const Icon(Icons.add_a_photo_outlined, color: Color(0xFF0081FF), size: 32),
              ),
            );
          }
          return Stack(
            children: [
              Container(
                width: 100,
                margin: const EdgeInsetsDirectional.only(start: 12),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(16),
                ),
                child: kIsWeb
                    ? FutureBuilder<Uint8List>(
                        future: _selectedImageFiles[index].readAsBytes(),
                        builder: (_, snap) => snap.hasData
                            ? ClipRRect(
                                borderRadius: BorderRadius.circular(16),
                                child: Image.memory(snap.data!, width: 100, height: 100, fit: BoxFit.cover),
                              )
                            : Container(width: 100, height: 100, color: Colors.grey[200]),
                      )
                    : ClipRRect(
                        borderRadius: BorderRadius.circular(16),
                        child: Image.network(
                          _selectedImageFiles[index].path,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                          errorBuilder: (_, __, ___) => Container(width: 100, height: 100, color: Colors.grey[200]),
                        ),
                      ),
              ),
              Positioned(
                top: 4,
                right: 4,
                child: GestureDetector(
                  onTap: () {
                    setState(() {
                      _selectedImageFiles.removeAt(index);
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.all(4),
                    decoration: const BoxDecoration(
                      color: Colors.red,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.close, color: Colors.white, size: 16),
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  /// Pick images from gallery
  Future<void> _pickImages() async {
    try {
      final List<XFile> pickedFiles = await _imagePicker.pickMultiImage(
        maxWidth: 1920,
        maxHeight: 1080,
        imageQuality: 85,
      );

      if (pickedFiles == null || pickedFiles.isEmpty) return;

      // Limiter à 5 images maximum
      final remainingSlots = 5 - _selectedImageFiles.length;
      if (remainingSlots <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Maximum 5 images allowed')),
        );
        return;
      }

      final filesToAdd = pickedFiles.take(remainingSlots).toList();

      setState(() {
        _selectedImageFiles.addAll(filesToAdd);
      });
    } catch (e) {
      debugPrint('Error picking images: $e');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error selecting images: $e')),
      );
    }
  }

  void _showCitySheet(BuildContext context) {
    if (_locations.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('جاري تحميل المواقع...'))
      );
      return;
    }
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => _CategorySheet(
        title: AppLocalizations.of(context)!.text_103,
        items: _locations.map((l) => l.cityNameAr).cast<String>().toList(),
        onSelected: (val) {
          setState(() => _selectedCity = val);
          Navigator.of(context).pop();
        },
      ),
    );
  }


  void _showCategorySheet(BuildContext context, bool isMain) {
    if (_categories.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('جاري تحميل الفئات...'))
      );
      return;
    }

    final parentCategories = _categories.where((c) => c.parentId == null).toList();

    List<Category> displayItems;
    if (isMain) {
      displayItems = parentCategories;
    } else {
      // Find selected parent by ID (more reliable than nameAr)
      final parentId = _selectedMainCategoryId;
      if (parentId == null) {
        // Fallback: search by nameAr
        final parent = _categories.firstWhere(
          (c) => c.nameAr == _selectedMainCategory && c.parentId == null,
          orElse: () => Category(id: 0, nameAr: '', nameFr: '', nameEn: ''),
        );
        displayItems = parent.id == 0 ? [] : _categories.where((c) => c.parentId == parent.id).toList();
      } else {
        displayItems = _categories.where((c) => c.parentId == parentId).toList();
      }
      debugPrint('[CreateAd] selectedCategoryId=$parentId selectedCategoryName=$_selectedMainCategory subcategories=${displayItems.length}');
      debugPrint('[CreateAd] subcategory names: ${displayItems.map((c) => c.nameAr).toList()}');
    }

    if (!isMain && displayItems.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('لا توجد فئات فرعية لهذه الفئة'))
      );
      return;
    }

    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => _CategorySheet(
        title: isMain ? AppLocalizations.of(context)!.text_98 : AppLocalizations.of(context)!.text_99,
        items: displayItems.map((c) => c.nameAr).toList(),
        onSelected: (val) {
          final selected = displayItems.firstWhere(
            (c) => c.nameAr == val,
            orElse: () => Category(id: 0, nameAr: val, nameFr: val, nameEn: val),
          );
          setState(() {
            if (isMain) {
              _selectedMainCategory = val;
              _selectedMainCategoryId = selected.id == 0 ? null : selected.id;
              _selectedSubCategory = null;
              _selectedSubCategoryId = null;
            } else {
              _selectedSubCategory = val;
              _selectedSubCategoryId = selected.id == 0 ? null : selected.id;
            }
          });
          Navigator.of(context).pop();
        },
      ),
    );
  }
}

class _CategorySheet extends StatefulWidget {
  final String title;
  final List<String> items;
  final Function(String) onSelected;

  const _CategorySheet({required this.title, required this.items, required this.onSelected});

  @override
  State<_CategorySheet> createState() => _CategorySheetState();
}

class _CategorySheetState extends State<_CategorySheet> {
  late List<String> _filtered;

  @override
  void initState() {
    super.initState();
    _filtered = widget.items;
  }

  void _onSearch(String query) {
    setState(() {
      _filtered = query.isEmpty
          ? widget.items
          : widget.items.where((item) => item.contains(query)).toList();
    });
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Container(
      padding: EdgeInsetsDirectional.only(bottom: MediaQuery.of(context).viewInsets.bottom),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const SizedBox(height: 12),
          Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[300], borderRadius: BorderRadius.circular(2))),
          const SizedBox(height: 24),
          Text(
            widget.title,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: TextField(
              textAlign: TextAlign.right,
              onChanged: _onSearch,
              decoration: InputDecoration(
                hintText: AppLocalizations.of(context)!.text_141,
                prefixIcon: const Icon(Icons.search, color: Colors.grey),
                filled: true,
                fillColor: isDarkMode ? Colors.black26 : Colors.grey[100],
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
              ),
            ),
          ),
          const SizedBox(height: 16),
          Flexible(
            child: ListView.separated(
              shrinkWrap: true,
              itemCount: _filtered.length,
              separatorBuilder: (context, index) => const Divider(height: 1),
              itemBuilder: (context, index) => ListTile(
                title: Text(_filtered[index], textAlign: TextAlign.right, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontWeight: FontWeight.w500)),
                onTap: () => widget.onSelected(_filtered[index]),
              ),
            ),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }
}