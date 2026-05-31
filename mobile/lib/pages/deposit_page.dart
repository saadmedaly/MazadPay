import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:image_picker/image_picker.dart';
import 'package:dio/dio.dart';
import 'package:path/path.dart' as path;
import 'package:mezadpay/pages/payment_success_page.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/services/payment_methods_service.dart';
import 'package:mezadpay/models/payment_method.dart';

const String _officialAccountNumber = '36601175';
const double _depositAmount = 500.0;

class DepositPage extends StatefulWidget {
  const DepositPage({super.key});

  @override
  State<DepositPage> createState() => _DepositPageState();
}

class _DepositPageState extends State<DepositPage> {
  final _phoneController = TextEditingController();
  final _notesController = TextEditingController();
  String? _selectedMethodCode;
  List<PaymentMethod> _methods = [];
  bool _loadingMethods = true;
  File? _receiptFile;
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    _loadMethods();
  }

  @override
  void dispose() {
    _phoneController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _loadMethods() async {
    try {
      final fetched = await PaymentMethodsService().getPaymentMethods();
      if (mounted) setState(() { _methods = fetched; _loadingMethods = false; });
    } catch (_) {
      if (mounted) setState(() => _loadingMethods = false);
    }
  }

  Future<void> _pickReceipt() async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(source: ImageSource.gallery, imageQuality: 85);
    if (picked != null && mounted) {
      setState(() => _receiptFile = File(picked.path));
    }
  }

  Future<void> _submit() async {
    if (_phoneController.text.trim().isEmpty) {
      _showSnack('يرجى إدخال رقم الهاتف');
      return;
    }
    if (_selectedMethodCode == null) {
      _showSnack('يرجى اختيار طريقة الدفع');
      return;
    }
    if (_receiptFile == null) {
      _showSnack('يرجى إرفاق إيصال التحويل');
      return;
    }

    setState(() => _submitting = true);

    try {
      final api = ApiService();

      // Step 1: create the deposit transaction
      final depositResp = await api.post<Map<String, dynamic>>(
        '/users/wallet/deposit',
        data: {
          'amount': _depositAmount,
          'gateway': _selectedMethodCode,
          'payment_method': _selectedMethodCode,
        },
      );

      if (depositResp == null || depositResp['success'] != true) {
        _showSnack('فشل إنشاء طلب الإيداع');
        setState(() => _submitting = false);
        return;
      }

      final txId = depositResp['data']?['id']?.toString();
      if (txId == null) {
        _showSnack('معرّف المعاملة غير متاح');
        setState(() => _submitting = false);
        return;
      }

      // Step 2: upload the receipt
      final fileName = path.basename(_receiptFile!.path);
      final formData = FormData.fromMap({
        'receipt': await MultipartFile.fromFile(_receiptFile!.path, filename: fileName),
        if (_notesController.text.trim().isNotEmpty)
          'note': _notesController.text.trim(),
        if (_phoneController.text.trim().isNotEmpty)
          'phone': _phoneController.text.trim(),
      });

      final receiptResp = await api.upload<Map<String, dynamic>>(
        '/users/wallet/transactions/$txId/receipt',
        data: formData,
      );

      if (receiptResp == null || receiptResp['success'] != true) {
        _showSnack('فشل رفع إيصال التحويل');
        setState(() => _submitting = false);
        return;
      }

      if (mounted) {
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (_) => const PaymentSuccessPage()),
        );
      }
    } catch (e) {
      _showSnack('حدث خطأ. يرجى المحاولة مجدداً');
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _showSnack(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
  }

  void _copyAccountNumber() {
    Clipboard.setData(const ClipboardData(text: _officialAccountNumber));
    _showSnack('تم نسخ الرقم');
  }

  String _methodName(PaymentMethod m) {
    final lang = Localizations.localeOf(context).languageCode;
    if (lang == 'ar') return m.nameAr;
    if (lang == 'fr') return m.nameFr;
    return m.nameEn ?? m.nameFr;
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    const blue = Color(0xFF0084FF);

    return Scaffold(
      backgroundColor: isDark ? const Color(0xFF121212) : const Color(0xFFF5F5F5),
      appBar: AppBar(
        backgroundColor: blue,
        elevation: 0,
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_forward_ios, color: Colors.white, size: 18),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          'طلب اشتراك جديد',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Phone number
            const Text('رقم الهاتف *', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 8),
            TextField(
              controller: _phoneController,
              keyboardType: TextInputType.phone,
              textAlign: TextAlign.right,
              decoration: InputDecoration(
                hintText: 'أدخل رقم هاتفك',
                hintStyle: const TextStyle(color: Colors.grey),
                filled: true,
                fillColor: isDark ? const Color(0xFF1D1D1D) : Colors.white,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: const BorderSide(color: blue),
                ),
              ),
            ),

            const SizedBox(height: 24),

            // Payment methods
            const Text('حسابات الدفع', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 8),

            if (_loadingMethods)
              const Center(child: CircularProgressIndicator())
            else
              ..._methods.map((m) => _buildMethodTile(m, isDark, blue)),

            const SizedBox(height: 16),

            // Payment instructions box
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFE8F4FF),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: const Color(0xFFB3D9FF)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  const Row(
                    textDirection: TextDirection.rtl,
                    children: [
                      Icon(Icons.info_outline, color: blue, size: 18),
                      SizedBox(width: 6),
                      Text('تعليمات الدفع', style: TextStyle(color: blue, fontWeight: FontWeight.bold, fontSize: 14)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'يرجى تحويل مبلغ الاشتراك إلى الرقم $_officialAccountNumber عبر تطبيق الدفع الذي اخترته، ثم إرفاق لقطة شاشة من عملية التحويل.',
                    textAlign: TextAlign.right,
                    style: const TextStyle(fontSize: 13, height: 1.5),
                  ),
                  const SizedBox(height: 12),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: const Color(0xFFB3D9FF)),
                    ),
                    child: Row(
                      textDirection: TextDirection.rtl,
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          _officialAccountNumber,
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18, letterSpacing: 1),
                        ),
                        GestureDetector(
                          onTap: _copyAccountNumber,
                          child: const Row(
                            textDirection: TextDirection.rtl,
                            children: [
                              Icon(Icons.copy_outlined, color: blue, size: 18),
                              SizedBox(width: 4),
                              Text('نسخ الرقم', style: TextStyle(color: blue, fontWeight: FontWeight.bold, fontSize: 13)),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 24),

            // Receipt upload
            const Text('إيصال التحويل *', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 8),
            GestureDetector(
              onTap: _pickReceipt,
              child: Container(
                width: double.infinity,
                height: _receiptFile != null ? 160 : 100,
                decoration: BoxDecoration(
                  color: isDark ? const Color(0xFF1D1D1D) : Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: _receiptFile != null ? blue : Colors.grey.shade300,
                    width: _receiptFile != null ? 2 : 1,
                  ),
                ),
                child: _receiptFile != null
                    ? Stack(
                        fit: StackFit.expand,
                        children: [
                          ClipRRect(
                            borderRadius: BorderRadius.circular(11),
                            child: Image.file(_receiptFile!, fit: BoxFit.cover),
                          ),
                          Positioned(
                            top: 8, right: 8,
                            child: GestureDetector(
                              onTap: () => setState(() => _receiptFile = null),
                              child: Container(
                                decoration: const BoxDecoration(color: Colors.red, shape: BoxShape.circle),
                                child: const Icon(Icons.close, color: Colors.white, size: 18),
                              ),
                            ),
                          ),
                        ],
                      )
                    : Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.upload_file_outlined, color: Colors.grey.shade400, size: 36),
                          const SizedBox(height: 8),
                          Text('يرجى الضغط لإرفاق صورة الحوالة',
                              style: TextStyle(color: Colors.grey.shade500, fontSize: 13)),
                        ],
                      ),
              ),
            ),

            const SizedBox(height: 24),

            // Notes
            const Text('ملاحظة (اختياري)', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 8),
            TextField(
              controller: _notesController,
              maxLines: 3,
              textAlign: TextAlign.right,
              decoration: InputDecoration(
                hintText: 'أي معلومات إضافية...',
                hintStyle: const TextStyle(color: Colors.grey),
                filled: true,
                fillColor: isDark ? const Color(0xFF1D1D1D) : Colors.white,
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey.shade300)),
                enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey.shade300)),
                focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: blue)),
              ),
            ),

            const SizedBox(height: 32),

            // Submit button
            SizedBox(
              width: double.infinity,
              height: 54,
              child: ElevatedButton(
                onPressed: _submitting ? null : _submit,
                style: ElevatedButton.styleFrom(
                  backgroundColor: blue,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  elevation: 0,
                ),
                child: _submitting
                    ? const CircularProgressIndicator(color: Colors.white)
                    : const Text('إتمام الدفع', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16)),
              ),
            ),

            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildMethodTile(PaymentMethod method, bool isDark, Color blue) {
    final selected = _selectedMethodCode == method.code;
    return GestureDetector(
      onTap: () => setState(() => _selectedMethodCode = method.code),
      child: Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: isDark ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: selected ? blue : Colors.grey.shade300,
            width: selected ? 2 : 1,
          ),
        ),
        child: Row(
          textDirection: TextDirection.rtl,
          children: [
            Icon(
              selected ? Icons.radio_button_checked : Icons.radio_button_unchecked,
              color: selected ? blue : Colors.grey,
              size: 22,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                _methodName(method),
                style: TextStyle(fontWeight: FontWeight.bold, color: isDark ? Colors.white : Colors.black),
              ),
            ),
            Text(
              _officialAccountNumber,
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
            ),
            const SizedBox(width: 12),
            GestureDetector(
              onTap: _copyAccountNumber,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F4FF),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: const Text('نسخ الرقم', style: TextStyle(color: Color(0xFF0084FF), fontSize: 12, fontWeight: FontWeight.bold)),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
