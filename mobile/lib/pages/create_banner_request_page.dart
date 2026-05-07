import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/services/request_api.dart';
import 'package:mezadpay/services/r2_upload_service.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:path/path.dart' as path;
import 'package:dio/dio.dart';

class CreateBannerRequestPage extends StatefulWidget {
  const CreateBannerRequestPage({super.key});

  @override
  State<CreateBannerRequestPage> createState() => _CreateBannerRequestPageState();
}

class _CreateBannerRequestPageState extends State<CreateBannerRequestPage> {
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _linkController = TextEditingController();
  
  DateTime? _startDate;
  DateTime? _endDate;
  File? _selectedImage;
  XFile? _selectedXFile;
  Uint8List? _webImage;
  
  final RequestApi _requestApi = RequestApi();
  final ApiService _apiService = ApiService();
  final ImagePicker _imagePicker = ImagePicker();
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _startDate = DateTime.now().add(const Duration(days: 1));
    _endDate = DateTime.now().add(const Duration(days: 8));
  }

  Future<void> _pickImage() async {
    final XFile? pickedFile = await _imagePicker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1920,
      maxHeight: 1080,
      imageQuality: 85,
    );

    if (pickedFile != null) {
      if (kIsWeb) {
        final bytes = await pickedFile.readAsBytes();
        setState(() {
          _selectedXFile = pickedFile;
          _webImage = bytes;
          _selectedImage = null;
        });
      } else {
        setState(() {
          _selectedXFile = pickedFile;
          _selectedImage = File(pickedFile.path);
          _webImage = null;
        });
      }
    }
  }

  Future<String?> _uploadImage() async {
    try {
      final fileName = _selectedXFile?.name ?? 'banner.jpg';
      
      final formData = FormData.fromMap({
        'image': kIsWeb 
          ? MultipartFile.fromBytes(_webImage!, filename: fileName)
          : await MultipartFile.fromFile(_selectedImage!.path, filename: fileName),
      });

      final response = await _apiService.upload<Map<String, dynamic>>(
        '/requests/banners/upload',
        data: formData,
      );

      if (response != null && response['success'] == true) {
        return response['data']['url'] as String?;
      }
      return null;
    } catch (e) {
      debugPrint('Error uploading banner image: $e');
      return null;
    }
  }

  Future<void> _submitRequest() async {
    final lang = Localizations.localeOf(context).languageCode;
    final title = _titleController.text.trim();
    final description = _descriptionController.text.trim();
    final link = _linkController.text.trim();

    final hasImage = kIsWeb ? _webImage != null : _selectedImage != null;

    String msgRequired = 'Veuillez remplir tous les champs obligatoires';
    String msgUploadError = 'Erreur lors de l\'envoi de l\'image';
    String msgSuccess = 'Demande envoyée avec succès';
    String msgError = 'Une erreur est survenue';

    if (lang == 'ar') {
      msgRequired = 'يرجى ملء جميع الحقول المطلوبة';
      msgUploadError = 'خطأ أثناء إرسال الصورة';
      msgSuccess = 'تم إرسال الطلب بنجاح';
      msgError = 'حدث خطأ ما';
    } else if (lang == 'en') {
      msgRequired = 'Please fill all required fields';
      msgUploadError = 'Error uploading image';
      msgSuccess = 'Request submitted successfully';
      msgError = 'An error occurred';
    }

    if (title.isEmpty || description.isEmpty || !hasImage || _startDate == null || _endDate == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msgRequired)),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      // 1. Upload Image
      final imageUrl = await _uploadImage();
      if (imageUrl == null) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msgUploadError)),
        );
        return;
      }

      // 2. Submit Request
      final response = await _requestApi.createBannerRequest(
        titleAr: title,
        titleFr: title,
        titleEn: title,
        descriptionAr: description,
        descriptionFr: description,
        descriptionEn: description,
        imageUrl: imageUrl,
        linkUrl: link,
        startsAt: _startDate!,
        endsAt: _endDate!,
      );

      if (mounted) {
        setState(() => _isLoading = false);
        if (response.success) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(msgSuccess)),
          );
          Navigator.of(context).pop();
        } else {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(response.message ?? msgError)),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erreur: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final lang = Localizations.localeOf(context).languageCode;
    
    String pageTitle = 'Demander une bannière';
    String labelTitle = 'Titre de la publicité';
    String hintTitle = 'Ex: Promo de la semaine';
    String labelDesc = 'Description';
    String hintDesc = 'Détails de votre offre...';
    String labelLink = 'Lien de redirection (Optionnel)';
    String hintLink = 'https://votre-site.com';
    String labelStart = 'Date de début';
    String labelEnd = 'Date de fin';
    String labelImage = 'Image de la bannière';
    String btnSubmit = 'Envoyer la demande';

    if (lang == 'ar') {
      pageTitle = 'طلب إعلان';
      labelTitle = 'عنوان الإعلان';
      hintTitle = 'مثال: عرض الأسبوع';
      labelDesc = 'الوصف';
      hintDesc = 'تفاصيل عرضك...';
      labelLink = 'رابط التوجيه (اختياري)';
      hintLink = 'https://site.com';
      labelStart = 'تاريخ البدء';
      labelEnd = 'تاريخ الانتهاء';
      labelImage = 'صورة الإعلان';
      btnSubmit = 'إرسال الطلب';
    } else if (lang == 'en') {
      pageTitle = 'Request a Banner';
      labelTitle = 'Ad Title';
      hintTitle = 'Ex: Weekly Promo';
      labelDesc = 'Description';
      hintDesc = 'Details of your offer...';
      labelLink = 'Target Link (Optional)';
      hintLink = 'https://your-site.com';
      labelStart = 'Start Date';
      labelEnd = 'End Date';
      labelImage = 'Banner Image';
      btnSubmit = 'Submit Request';
    }

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      appBar: AppBar(
        title: Text(pageTitle, style: const TextStyle(fontWeight: FontWeight.bold)),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildLabel(labelTitle),
            _buildTextField(controller: _titleController, hint: hintTitle),
            
            const SizedBox(height: 20),
            _buildLabel(labelDesc),
            _buildTextField(controller: _descriptionController, hint: hintDesc, maxLines: 3),
            
            const SizedBox(height: 20),
            _buildLabel(labelLink),
            _buildTextField(controller: _linkController, hint: hintLink, icon: Icons.link),
            
            const SizedBox(height: 20),
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildLabel(labelStart),
                      _buildDateSelector(
                        value: _startDate,
                        onTap: () => _selectDate(true),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildLabel(labelEnd),
                      _buildDateSelector(
                        value: _endDate,
                        onTap: () => _selectDate(false),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            
            const SizedBox(height: 32),
            _buildLabel(labelImage),
            _buildImagePicker(),
            
            const SizedBox(height: 48),
            SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: _isLoading ? null : _submitRequest,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF0081FF),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                ),
                child: _isLoading
                    ? const CircularProgressIndicator(color: Colors.white)
                    : Text(
                        btnSubmit,
                        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white),
                      ),
              ),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildLabel(String text) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
          color: isDarkMode ? Colors.white70 : Colors.black87,
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String hint,
    int maxLines = 1,
    IconData? icon,
  }) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.withOpacity(0.2)),
      ),
      child: TextField(
        controller: controller,
        maxLines: maxLines,
        decoration: InputDecoration(
          hintText: hint,
          hintStyle: const TextStyle(color: Colors.grey, fontSize: 13),
          contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          border: InputBorder.none,
          prefixIcon: icon != null ? Icon(icon, color: const Color(0xFF0081FF).withOpacity(0.5), size: 20) : null,
        ),
      ),
    );
  }

  Widget _buildDateSelector({DateTime? value, required VoidCallback onTap}) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final lang = Localizations.localeOf(context).languageCode;
    String labelChoose = lang == 'ar' ? 'اختر...' : (lang == 'en' ? 'Choose...' : 'Choisir...');

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.grey.withOpacity(0.2)),
        ),
        child: Row(
          children: [
            const Icon(Icons.calendar_today_outlined, size: 18, color: Color(0xFF0081FF)),
            const SizedBox(width: 8),
            Text(
              value != null ? '${value.day}/${value.month}/${value.year}' : labelChoose,
              style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _selectDate(bool isStart) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: (isStart ? _startDate : _endDate) ?? DateTime.now().add(const Duration(days: 1)),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );
    if (picked != null) {
      setState(() {
        if (isStart) {
          _startDate = picked;
          if (_endDate != null && _endDate!.isBefore(_startDate!)) {
            _endDate = _startDate!.add(const Duration(days: 7));
          }
        } else {
          _endDate = picked;
        }
      });
    }
  }

  Widget _buildImagePicker() {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final lang = Localizations.localeOf(context).languageCode;
    
    return GestureDetector(
      onTap: _pickImage,
      child: Container(
        width: double.infinity,
        height: 150,
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.grey[100],
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFF0081FF).withOpacity(0.2), style: BorderStyle.solid),
        ),
        child: (_selectedImage != null || _webImage != null)
            ? Stack(
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: kIsWeb 
                      ? Image.memory(_webImage!, width: double.infinity, height: 150, fit: BoxFit.cover)
                      : Image.file(_selectedImage!, width: double.infinity, height: 150, fit: BoxFit.cover),
                  ),
                  Positioned(
                    top: 8,
                    right: 8,
                    child: GestureDetector(
                      onTap: () => setState(() {
                        _selectedImage = null;
                        _webImage = null;
                        _selectedXFile = null;
                      }),
                      child: Container(
                        padding: const EdgeInsets.all(4),
                        decoration: const BoxDecoration(color: Colors.red, shape: BoxShape.circle),
                        child: const Icon(Icons.close, color: Colors.white, size: 16),
                      ),
                    ),
                  ),
                ],
              )
            : Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.add_a_photo_outlined, color: Color(0xFF0081FF), size: 40),
                  const SizedBox(height: 8),
                  Text(
                    lang == 'ar' ? 'إضافة صورة' : (lang == 'en' ? 'Add an image' : 'Ajouter une image'),
                    style: const TextStyle(color: Colors.grey, fontWeight: FontWeight.w600),
                  ),
                ],
              ),
      ),
    );
  }
}
