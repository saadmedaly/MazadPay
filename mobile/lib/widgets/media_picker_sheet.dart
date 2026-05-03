import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';


class MediaPickerSheet extends StatefulWidget {
  final Function(String) onMediaSelected;
  
  const MediaPickerSheet({super.key, required this.onMediaSelected});

  @override
  State<MediaPickerSheet> createState() => _MediaPickerSheetState();
}

class _MediaPickerSheetState extends State<MediaPickerSheet> with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Container(
      height: MediaQuery.of(context).size.height * 0.7,
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
      ),
      child: Column(
        children: [
          const SizedBox(height: 12),
          Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[300], borderRadius: BorderRadius.circular(2))),
          const SizedBox(height: 24),
          
          // Custom Tab Bar
          TabBar(
            controller: _tabController,
            labelColor: const Color(0xFF0081FF),
            unselectedLabelColor: Colors.grey,
            indicatorColor: const Color(0xFF0081FF),
            indicatorSize: TabBarIndicatorSize.label,
            labelStyle: TextStyle(fontFamily: 'Plus Jakarta Sans', fontWeight: FontWeight.bold),
            tabs: [
              Tab(text: AppLocalizations.of(context)!.text_266),
              Tab(text: AppLocalizations.of(context)!.text_377),
            ],
          ),
          
          // Tab Content
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildMediaGrid(isDarkMode, isVideo: false),
                _buildMediaGrid(isDarkMode, isVideo: true),
              ],
            ),
          ),
          
          // Selection Button
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: () => Navigator.of(context).pop(),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF0081FF),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                ),
                child: Text(AppLocalizations.of(context)!.text_378, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontWeight: FontWeight.bold, color: Colors.white)),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMediaGrid(bool isDarkMode, {required bool isVideo}) {
    // Dummy media list
    final List<String> dummyAssets = [
      'assets/smaah.png',
      'assets/iphone.png',
      'assets/corolla.png',
      'assets/raf4.png',
      'assets/house.png',
      'assets/laptop.png',
    ];

    return GridView.builder(
      padding: const EdgeInsets.all(16),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        crossAxisSpacing: 8,
        mainAxisSpacing: 8,
      ),
      itemCount: 12,
      itemBuilder: (context, index) {
        final asset = dummyAssets[index % dummyAssets.length];
        return GestureDetector(
          onTap: () => widget.onMediaSelected(asset),
          child: Stack(
            children: [
              Container(
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  image: DecorationImage(image: AssetImage(asset), fit: BoxFit.cover),
                ),
              ),
              if (isVideo)
                const Center(child: Icon(Icons.play_circle_fill, color: Colors.white, size: 32)),
              Positioned(
                top: 8, right: 8,
                child: Container(
                  padding: const EdgeInsets.all(2),
                  decoration: const BoxDecoration(color: Colors.white, shape: BoxShape.circle),
                  child: const Icon(Icons.check_circle, color: Color(0xFF0081FF), size: 18),
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}