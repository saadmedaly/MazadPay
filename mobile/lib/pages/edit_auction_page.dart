import 'package:flutter/material.dart';
import '../services/auction_api.dart';
import '../utils/money_formatter.dart';

class EditAuctionPage extends StatefulWidget {
  final Map<String, dynamic> auction;
  const EditAuctionPage({super.key, required this.auction});

  @override
  State<EditAuctionPage> createState() => _EditAuctionPageState();
}

class _EditAuctionPageState extends State<EditAuctionPage> {
  final _titleController = TextEditingController();
  final _descController = TextEditingController();
  final _priceController = TextEditingController();
  final AuctionApi _api = AuctionApi();
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    final a = widget.auction;
    _titleController.text = a['title_ar']?.toString() ?? a['title']?.toString() ?? '';
    _descController.text = a['description_ar']?.toString() ?? a['description']?.toString() ?? '';
    _priceController.text = (a['current_price'] ?? a['start_price'] ?? '').toString();
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descController.dispose();
    _priceController.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final price = double.tryParse(_priceController.text.trim());
    if (price == null || price <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('أدخل سعراً صحيحاً')),
      );
      return;
    }
    setState(() => _isLoading = true);
    try {
      final id = widget.auction['id']?.toString() ?? '';
      final response = await _api.updateAuction(
        auctionId: id,
        title: _titleController.text.trim(),
        description: _descController.text.trim(),
        startingPrice: price,
      );
      if (!mounted) return;
      if (response.success) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('تم حفظ التعديلات بنجاح'), backgroundColor: Color(0xFF00C58D)),
        );
        Navigator.of(context).pop(true); // signal refresh
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(response.message ?? 'فشل حفظ التعديلات')),
        );
      }
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('خطأ: $e')),
      );
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Scaffold(
      backgroundColor: isDark ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        title: const Text(
          'تعديل المزاد',
          style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold),
        ),
        leading: IconButton(
          icon: Icon(Icons.arrow_back_ios, color: isDark ? Colors.white : Colors.black, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _label('العنوان'),
            _field(_titleController, hint: 'عنوان المزاد'),
            const SizedBox(height: 20),
            _label('الوصف'),
            _field(_descController, hint: 'وصف المزاد', maxLines: 4),
            const SizedBox(height: 20),
            _label('السعر (${widget.auction['currency_code']?.toString() ?? MoneyFormatter.fallbackCurrencyCode})'),
            _field(_priceController, hint: 'السعر', keyboardType: TextInputType.number),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.orange.withOpacity(0.1),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: Colors.orange.withOpacity(0.3)),
              ),
              child: const Row(
                children: [
                  Icon(Icons.info_outline, color: Colors.orange, size: 18),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'لا يمكن تعديل وقت بداية أو نهاية المزاد بعد الإنشاء.',
                      style: TextStyle(fontSize: 12, color: Colors.orange),
                      textDirection: TextDirection.rtl,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 40),
            SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: _isLoading ? null : _save,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF0081FF),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                ),
                child: _isLoading
                    ? const CircularProgressIndicator(color: Colors.white)
                    : const Text(
                        'حفظ التعديلات',
                        style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 16, fontWeight: FontWeight.bold, color: Colors.white),
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _label(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        text,
        style: const TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 14, fontWeight: FontWeight.bold),
      ),
    );
  }

  Widget _field(TextEditingController controller, {required String hint, int maxLines = 1, TextInputType? keyboardType}) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      decoration: BoxDecoration(
        color: isDark ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Colors.grey.withOpacity(0.2)),
      ),
      child: TextField(
        controller: controller,
        maxLines: maxLines,
        keyboardType: keyboardType,
        textAlign: TextAlign.right,
        decoration: InputDecoration(
          hintText: hint,
          hintStyle: const TextStyle(color: Colors.grey, fontSize: 13),
          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          border: InputBorder.none,
        ),
      ),
    );
  }
}
