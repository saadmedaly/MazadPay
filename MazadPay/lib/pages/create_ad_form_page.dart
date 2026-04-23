import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

import 'ad_success_page.dart';
import '../widgets/media_picker_sheet.dart';
import '../services/auction_api.dart';

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
  
  late String _selectedMainCategory;
  late String _selectedSubCategory;
  late String _selectedCity;
  final List<String> _selectedImages = []; // List of assets or paths
  final AuctionApi _auctionApi = AuctionApi();
  bool _isLoading = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final l10n = AppLocalizations.of(context)!;
    _selectedMainCategory = l10n.text_86;
    _selectedSubCategory = l10n.text_87;
    _selectedCity = l10n.text_88;
    
    // Initialize our sub-category mapping
    _subCategoriesByMain = {
      l10n.text_86: [l10n.text_114, l10n.text_115, l10n.text_116, l10n.text_117, l10n.text_118],
      l10n.text_119: [l10n.text_120, l10n.text_121, l10n.text_122, l10n.text_123, l10n.text_124, l10n.text_125, l10n.text_126, l10n.text_127, l10n.text_128, l10n.text_129],
      l10n.text_130: [l10n.text_120, l10n.text_121, l10n.text_122, l10n.text_123, l10n.text_124, l10n.text_125, l10n.text_126, l10n.text_127, l10n.text_128, l10n.text_129],
      l10n.text_131: [l10n.text_120, l10n.text_121, l10n.text_122, l10n.text_123, l10n.text_124, l10n.text_125, l10n.text_126, l10n.text_127, l10n.text_128, l10n.text_129],
      l10n.text_132: [l10n.text_120, l10n.text_121, l10n.text_122, l10n.text_123, l10n.text_124, l10n.text_125, l10n.text_126, l10n.text_127, l10n.text_128, l10n.text_129],
      l10n.text_133: [l10n.text_120, l10n.text_121, l10n.text_122, l10n.text_123, l10n.text_124, l10n.text_125, l10n.text_126, l10n.text_127, l10n.text_128, l10n.text_129],
    };
  }

  Map<String, List<String>> _subCategoriesByMain = {};

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

    if (_selectedImages.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_add_image)),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final response = await _auctionApi.createAuction(
        title: name,
        description: description,
        startingPrice: price,
        category: _selectedMainCategory,
        subCategory: _selectedSubCategory,
        location: _selectedCity,
        images: _selectedImages,
        phone: phone,
      );

      setState(() => _isLoading = false);

      if (response.success) {
        Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => const AdSuccessPage()),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(response.message ?? AppLocalizations.of(context)!.error_create_auction)),
        );
      }
    } catch (e) {
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

  Widget _buildCategorySelector(BuildContext context, {required String value, required IconData icon, required VoidCallback onTap}) {
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
                 value,
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
        itemCount: _selectedImages.length + 1,
        itemBuilder: (context, index) {
          if (index == _selectedImages.length) {
            return GestureDetector(
              onTap: () => _showMediaPicker(context),
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
          return Container(
            width: 100,
            margin: const EdgeInsetsDirectional.only(start: 12),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(16),
              image: DecorationImage(image: AssetImage(_selectedImages[index]), fit: BoxFit.cover),
            ),
          );
        },
      ),
    );
  }

  void _showCitySheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => _CategorySheet(
        title: AppLocalizations.of(context)!.text_103,
        items: [AppLocalizations.of(context)!.text_88, AppLocalizations.of(context)!.text_104, AppLocalizations.of(context)!.text_105, AppLocalizations.of(context)!.text_106, AppLocalizations.of(context)!.text_107, AppLocalizations.of(context)!.text_108, AppLocalizations.of(context)!.text_109, AppLocalizations.of(context)!.text_110, AppLocalizations.of(context)!.text_111, AppLocalizations.of(context)!.text_112, AppLocalizations.of(context)!.text_113],
        onSelected: (val) {
          setState(() => _selectedCity = val);
          Navigator.of(context).pop();
        },
      ),
    );
  }


  void _showCategorySheet(BuildContext context, bool isMain) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => _CategorySheet(
        title: isMain ? AppLocalizations.of(context)!.text_98 : AppLocalizations.of(context)!.text_99,
        items: isMain
          ? [AppLocalizations.of(context)!.text_86, AppLocalizations.of(context)!.text_119, AppLocalizations.of(context)!.text_130, AppLocalizations.of(context)!.text_134, AppLocalizations.of(context)!.text_135, AppLocalizations.of(context)!.text_136, AppLocalizations.of(context)!.text_124, AppLocalizations.of(context)!.text_137, AppLocalizations.of(context)!.text_132, AppLocalizations.of(context)!.text_133, AppLocalizations.of(context)!.text_125, AppLocalizations.of(context)!.text_138, AppLocalizations.of(context)!.text_139, AppLocalizations.of(context)!.text_140]
          : (_subCategoriesByMain[_selectedMainCategory] ?? [AppLocalizations.of(context)!.text_120, AppLocalizations.of(context)!.text_121, AppLocalizations.of(context)!.text_122, AppLocalizations.of(context)!.text_123, AppLocalizations.of(context)!.text_124, AppLocalizations.of(context)!.text_125, AppLocalizations.of(context)!.text_126, AppLocalizations.of(context)!.text_127, AppLocalizations.of(context)!.text_128, AppLocalizations.of(context)!.text_129]),
        onSelected: (val) {
          setState(() {
            if (isMain) {
              _selectedMainCategory = val;
            } else {
              _selectedSubCategory = val;
            }
          });
          Navigator.of(context).pop();
        },
      ),
    );
  }

  void _showMediaPicker(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => MediaPickerSheet(
        onMediaSelected: (asset) {
          setState(() {
            if (_selectedImages.length < 5 && !_selectedImages.contains(asset)) {
              _selectedImages.add(asset);
            }
          });
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