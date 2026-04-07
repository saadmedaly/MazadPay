import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'ad_success_page.dart';
import '../widgets/media_picker_sheet.dart';

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
  
  String _selectedMainCategory = 'سيارات';
  String _selectedSubCategory = 'أجهزة منزلية';
  String _selectedCity = 'انواكشوط';
  final List<String> _selectedImages = []; // List of assets or paths

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            'مزايدة',
            style: GoogleFonts.plusJakartaSans(
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
                'إنشاء إعلان جديد',
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
              ),
              const SizedBox(height: 32),
              
              _buildLabel(context, 'اسم الإعلان'),
              _buildTextField(context, controller: _nameController, hint: 'أدخل اسم منتجك'),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'الوصف'),
              _buildTextField(context, controller: _descriptionController, hint: 'اكتب وصفا', maxLines: 3, icon: Icons.description_outlined),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'رقم الهاتف'),
              _buildTextField(context, controller: _phoneController, hint: 'لتواصل معك', icon: Icons.phone_outlined, keyboardType: TextInputType.phone),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'السعر'),
              _buildTextField(context, controller: _priceController, hint: 'أدخل السعر', icon: Icons.attach_money_outlined, keyboardType: TextInputType.number, suffixText: 'MRU'),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'اختر الفئة الرئيسية'),
              _buildCategorySelector(context, value: _selectedMainCategory, icon: Icons.category_outlined, onTap: () => _showCategorySheet(context, true)),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'اختر الفئة الفرعية'),
              _buildCategorySelector(context, value: _selectedSubCategory, icon: Icons.grid_view_outlined, onTap: () => _showCategorySheet(context, false)),
              
              const SizedBox(height: 24),
              _buildLabel(context, 'اختر الموقع'),
              _buildCategorySelector(context, value: _selectedCity, icon: Icons.location_on_outlined, onTap: () => _showCitySheet(context)),
              
              const SizedBox(height: 32),
              _buildLabel(context, 'صور وفيديوهات المنتج'),
              _buildMediaSection(context),
              
              const SizedBox(height: 48),
              
              // Submit Button
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton(
                  onPressed: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(builder: (context) => const AdSuccessPage()),
                    );
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0081FF),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  child: Text(
                    'نشر الإعلان',
                    style: GoogleFonts.plusJakartaSans(
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
      ),
    );
  }

  Widget _buildLabel(BuildContext context, String text) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsets.only(bottom: 8, right: 4),
      child: Row(
        children: [
          Text(
            text,
            style: GoogleFonts.plusJakartaSans(
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
          hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 13),
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
                 style: GoogleFonts.plusJakartaSans(
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
                margin: const EdgeInsets.only(left: 12),
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
            margin: const EdgeInsets.only(left: 12),
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
        title: 'اختر المدينة',
        items: ['انواكشوط', 'انواذيبو', 'كيهيدي', 'روصو', 'أطار', 'زويرات', 'العيون', 'النعمة', 'تجكجة', 'كبني'],
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
        title: isMain ? 'اختر الفئة الرئيسية' : 'اختر الفئة الفرعية',
        items: isMain 
          ? ['عقارات', 'سيارات', 'هواتف', 'الكترونيات', 'ساعات', 'دراجات']
          : ['أجهزة منزلية', 'قطع أرضية', 'مستلزمات رجالية', 'مستلزمات نسائية', 'حيوانات', 'شاحنات', 'مواد ثقيلة', 'أرقام هاتف', 'ربّاخة', 'بيع مشاريع'],
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

class _CategorySheet extends StatelessWidget {
  final String title;
  final List<String> items;
  final Function(String) onSelected;
  
  const _CategorySheet({required this.title, required this.items, required this.onSelected});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Container(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
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
            title,
            style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: TextField(
              textAlign: TextAlign.right,
              decoration: InputDecoration(
                hintText: 'بحث...',
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
              itemCount: items.length,
              separatorBuilder: (context, index) => const Divider(height: 1),
              itemBuilder: (context, index) => ListTile(
                title: Text(items[index], textAlign: TextAlign.right, style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w500)),
                onTap: () => onSelected(items[index]),
              ),
            ),
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }
}
