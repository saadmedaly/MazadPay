import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

import 'dart:math' as math;
import 'create_ad_form_page.dart';

class CreateAdStartPage extends StatelessWidget {
  const CreateAdStartPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    final categories = [
      {'icon': Icons.home_outlined, 'label': AppLocalizations.of(context)!.text_119},
      {'icon': Icons.directions_car_outlined, 'label': AppLocalizations.of(context)!.text_86},
      {'icon': Icons.smartphone_outlined, 'label': AppLocalizations.of(context)!.text_130},
      {'icon': Icons.laptop_outlined, 'label': AppLocalizations.of(context)!.text_131},
      {'icon': Icons.watch_outlined, 'label': AppLocalizations.of(context)!.text_132},
      {'icon': Icons.pedal_bike_outlined, 'label': AppLocalizations.of(context)!.text_133},
      {'icon': Icons.pets_outlined, 'label': AppLocalizations.of(context)!.text_124},
      {'icon': Icons.chair_outlined, 'label': AppLocalizations.of(context)!.text_142},
    ];

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : Colors.white,
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
          icon: Icon(Icons.close, color: isDarkMode ? Colors.white : Colors.black),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: Column(
        children: [
          const Spacer(),
          // Circular Category Selector
          Center(
            child: SizedBox(
              width: 300,
              height: 300,
              child: Stack(
                alignment: Alignment.center,
                children: [
                  // Central Hammer Icon
                  _buildCentralHammer(context),
                  
                  // Orbiting Icons
                  ...List.generate(categories.length, (index) {
                    final angle = (index * 2 * math.pi / categories.length) - (math.pi / 2);
                    const radius = 120.0;
                    final x = radius * math.cos(angle);
                    final y = radius * math.sin(angle);
                    
                    return Transform.translate(
                      offset: Offset(x, y),
                      child: _buildCategoryIcon(
                        context, 
                        categories[index]['icon'] as IconData,
                        categories[index]['label'] as String,
                      ),
                    );
                  }),

                  // Decorative Lines (Simplified)
                  CustomPaint(
                    size: const Size(300, 300),
                    painter: _CircularLinesPainter(isDarkMode, categories.length),
                  ),
                ],
              ),
            ),
          ),
          const Spacer(),
          
          // Bottom Text
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 40),
            child: Text(
              AppLocalizations.of(context)!.text_143,
              textAlign: TextAlign.center,
              style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: isDarkMode ? Colors.white70 : Colors.black87,
                height: 1.5,
              ),
            ),
          ),
          const SizedBox(height: 32),
          
          // Continue Button
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 40),
            child: SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute(builder: (context) => const CreateAdFormPage()),
                  );
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF0081FF),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  elevation: 0,
                ),
                child: Text(
                  AppLocalizations.of(context)!.text_144,
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCentralHammer(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Container(
      width: 80,
      height: 80,
      decoration: BoxDecoration(
        color: Colors.white,
        shape: BoxShape.circle,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(color: const Color(0xFF0081FF).withOpacity(0.1), width: 2),
      ),
      child: const Center(
        child: Icon(
          Icons.gavel,
          color: Color(0xFF0081FF),
          size: 40,
        ),
      ),
    );
  }

  Widget _buildCategoryIcon(BuildContext context, IconData icon, String label) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Container(
      width: 48,
      height: 48,
      decoration: BoxDecoration(
        color: Colors.white,
        shape: BoxShape.circle,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 4),
          ),
        ],
        border: Border.all(color: Colors.grey.withOpacity(0.1), width: 1.5),
      ),
      child: Center(
        child: Icon(
          icon,
          color: const Color(0xFF0081FF),
          size: 24,
        ),
      ),
    );
  }
}

class _CircularLinesPainter extends CustomPainter {
  final bool isDarkMode;
  final int count;
  _CircularLinesPainter(this.isDarkMode, this.count);

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = isDarkMode ? Colors.white.withOpacity(0.08) : Colors.black.withOpacity(0.04)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.0;

    final center = Offset(size.width / 2, size.height / 2);
    
    // Draw outer circle
    canvas.drawCircle(center, 120, paint);
    
    // Draw connecting lines to each category
    for (int i = 0; i < count; i++) {
      final angle = (i * 2 * math.pi / count) - (math.pi / 2);
      final radius = 120.0;
      final x = radius * math.cos(angle);
      final y = radius * math.sin(angle);
      
      // Stop the line a bit before the category icon
      final lineEnd = Offset(center.dx + x * 0.8, center.dy + y * 0.8);
      canvas.drawLine(center, lineEnd, paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
