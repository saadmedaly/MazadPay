import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/widgets/custom_app_bar.dart';
import 'package:mezadpay/pages/payment_success_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

class PaymentDetailsPage extends StatefulWidget {
  final String methodName;

  const PaymentDetailsPage({
    super.key,
    required this.methodName,
  });

  @override
  State<PaymentDetailsPage> createState() => _PaymentDetailsPageState();
}

class _PaymentDetailsPageState extends State<PaymentDetailsPage> {
  bool _receiptUploaded = false;
  String? _uploadedImagePath;

  void _showImagePicker(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
        return Container(
          height: MediaQuery.of(context).size.height * 0.75,
          decoration: BoxDecoration(
            color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
          ),
          child: Column(
            children: [
              const SizedBox(height: 12),
              Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(color: Colors.grey[300], borderRadius: BorderRadius.circular(2)),
              ),
              const SizedBox(height: 20),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Text(
                  l10n.appAccessPhotos,
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                ),
              ),
              const SizedBox(height: 20),
              // Toggle Buttons: Photos / Albums
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 40),
                child: Row(
                  children: [
                    Expanded(
                      child: Container(
                        height: 40,
                        decoration: BoxDecoration(
                          color: const Color(0xFFE8F0FE),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Center(
                          child: Text(l10n.photos,
                              style: const TextStyle(color: Color(0xFF0084FF), fontWeight: FontWeight.bold)),
                        ),
                      ),
                    ),
                    const SizedBox(width: 15),
                    Expanded(
                      child: Container(
                        height: 40,
                        decoration: BoxDecoration(
                          color: isDarkMode ? Colors.white10 : Colors.grey[100],
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Center(
                          child:
                              Text(l10n.albums, style: const TextStyle(color: Colors.grey, fontWeight: FontWeight.bold)),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              // Image Grid
              Expanded(
                child: GridView.builder(
                  padding: const EdgeInsets.all(2),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 3,
                    crossAxisSpacing: 2,
                    mainAxisSpacing: 2,
                  ),
                  itemCount: 15, // Mock images
                  itemBuilder: (context, index) {
                    // Using available assets as placeholders
                    final assets = [
                      'corolla.png',
                      'house.png',
                      'iphone.png',
                      'laptop.png',
                      'raf4.png',
                      'smaah.png',
                      'on1.png',
                      'on2.png',
                      'on3.png',
                      'announcement.png',
                      'logo.png',
                      'Bankily.png',
                      'Click.png',
                      'Masrivi.png',
                      'Sedad.png'
                    ];
                    final asset = assets[index % assets.length];

                    return GestureDetector(
                      onTap: () {
                        Navigator.pop(context);
                        setState(() {
                          _receiptUploaded = true;
                          _uploadedImagePath = asset;
                        });
                      },
                      child: Stack(
                        fit: StackFit.expand,
                        children: [
                          Image.asset(asset, fit: BoxFit.cover),
                          if (index == 4) // Example selected state as in screenshot
                            Container(
                              decoration: BoxDecoration(
                                border: Border.all(color: const Color(0xFF0084FF), width: 3),
                              ),
                              child: const Align(
                                alignment: Alignment.topRight,
                                child: Padding(
                                  padding: EdgeInsets.all(4.0),
                                  child: Icon(Icons.check_circle, color: Color(0xFF0084FF), size: 18),
                                ),
                              ),
                            ),
                        ],
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      drawer: const SideMenuDrawer(),
      appBar: CustomAppBar(
        title: l10n.addDeposit,
        showBack: true,
        onBackPress: () => Navigator.pop(context),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            const SizedBox(height: 20),

            // Merchant Info Section
            Text(
              l10n.payVia(widget.methodName),
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 28),
            ),
            const SizedBox(height: 12),
            Text(
              l10n.useMerchantCode,
              style: TextStyle(color: Colors.grey[600], fontSize: 16),
            ),
            const SizedBox(height: 8),
            const Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.copy_rounded, size: 24, color: Colors.black),
                SizedBox(width: 8),
                Text(
                  '07755',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 28),
                ),
              ],
            ),

            const SizedBox(height: 20),
            const Divider(),
            const SizedBox(height: 20),

            // Transfer Visual
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // Bank Logo Placeholder
                Container(
                  width: 70,
                  height: 40,
                  decoration: BoxDecoration(
                    color: isDarkMode ? Colors.white10 : Colors.grey[100],
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Center(
                    child: Text(widget.methodName == 'سداد' ? 'Sedad' : widget.methodName,
                        style: const TextStyle(fontWeight: FontWeight.bold)),
                  ),
                ),
                const SizedBox(width: 15),
                Expanded(
                  child: Column(
                    children: [
                      const Icon(Icons.sync_alt, size: 24),
                      Text(
                        l10n.amountToMazad,
                        textAlign: TextAlign.center,
                        style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Colors.grey[800]),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 15),
                // MazadPay Logo
                Image.asset('logo.png', height: 40, width: 90, fit: BoxFit.contain, errorBuilder: (c, e, s) => const Text('MazadPay')),
              ],
            ),

            const SizedBox(height: 30),

            // Order Details Card
            Container(
              width: double.infinity,
              decoration: BoxDecoration(
                color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
              ),
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  _buildDetailRow(l10n.orderCreationTime, '2025/12/27 م', isValueBold: true),
                  const SizedBox(height: 12),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(l10n.status, style: const TextStyle(color: Colors.grey)),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                        decoration: BoxDecoration(
                          color: const Color(0xFFFFEBD5),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          l10n.pendingPayment,
                          style: const TextStyle(color: Color(0xFFD47A2C), fontWeight: FontWeight.bold, fontSize: 13),
                        ),
                      ),
                    ],
                  ),
                  const Divider(height: 32),
                  _buildDetailRow(l10n.orderCreationFee, '100MRU', isValueBold: true),
                  const SizedBox(height: 16),
                  _buildDetailRow(l10n.payerPhoneNumber, '36601175', isValueBold: true),
                  const SizedBox(height: 16),
                  _buildDetailRow(l10n.totalAmountToPay(widget.methodName), '100MRU',
                      isValueBold: true, isLarge: true),
                ],
              ),
            ),

            const SizedBox(height: 30),

            // Receipt Upload Prompt
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, color: Colors.red, size: 24),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    l10n.uploadReceiptPrompt,
                    style: TextStyle(color: Colors.grey[800], fontSize: 14, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),

            const SizedBox(height: 20),

            // Upload Box
            GestureDetector(
              onTap: () => _showImagePicker(context),
              child: Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
                ),
                child: Center(
                  child: _receiptUploaded
                      ? Column(
                          children: [
                            if (_uploadedImagePath != null) ...[
                              ClipRRect(
                                borderRadius: BorderRadius.circular(8),
                                child: Image.asset(_uploadedImagePath!, height: 100, width: 100, fit: BoxFit.cover),
                              ),
                              const SizedBox(height: 10),
                            ],
                            const Icon(Icons.check_circle, color: Colors.green, size: 30),
                            const SizedBox(height: 8),
                            Text(l10n.receiptAttached, style: const TextStyle(fontWeight: FontWeight.bold)),
                          ],
                        )
                      : Text(
                          l10n.clickToUploadReceipt,
                          style: TextStyle(color: Colors.grey[600], fontWeight: FontWeight.bold),
                        ),
                ),
              ),
            ),

            const SizedBox(height: 40),

            // Final Action Button
            SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: () {
                  if (_receiptUploaded) {
                    Navigator.push(
                      context,
                      MaterialPageRoute(builder: (context) => const PaymentSuccessPage()),
                    );
                  } else {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(l10n.attachReceiptFirst)),
                    );
                  }
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: primaryBlue,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  elevation: 0,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.arrow_back, color: Colors.white),
                    const SizedBox(width: 12),
                    Flexible(
                      child: Text(
                        _receiptUploaded ? l10n.completePayment : l10n.payVia(widget.methodName),
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: Colors.white),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 40),
          ],
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value, {bool isValueBold = false, bool isLarge = false}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Expanded(child: Text(label, style: const TextStyle(color: Colors.grey))),
        Text(
          value,
          style: TextStyle(
            fontWeight: isValueBold ? FontWeight.bold : FontWeight.normal,
            fontSize: isLarge ? 18 : 15,
            color: Colors.black,
          ),
        ),
      ],
    );
  }
}
