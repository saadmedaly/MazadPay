import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/pages/create_ad_start_page.dart';
import 'package:mezadpay/pages/notifications_page.dart';
import 'package:mezadpay/widgets/live_indicator.dart';
import 'package:mezadpay/pages/auction_details_page.dart';


import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';
import '../services/auction_api.dart';

class HomePage extends ConsumerStatefulWidget {
  const HomePage({super.key});

  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  int _currentIndex = 0;
  bool _showLocationModal = true;
  int _selectedCityIndex = 0;
  late List<String> _cities;
  final AuctionApi _auctionApi = AuctionApi();
  List<Map<String, dynamic>> _auctions = [];
  bool _isLoading = true;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _cities = [AppLocalizations.of(context)!.text_193, AppLocalizations.of(context)!.text_194];
  }

  @override
  void initState() {
    super.initState();
    // Show the modal after the first frame is rendered
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_showLocationModal) {
        _showLocationPermissionDialog();
      }
    });
    // Load auctions from API
    _loadAuctions();
  }

  Future<void> _loadAuctions() async {
    try {
      final response = await _auctionApi.getAuctions(
        page: 1,
        limit: 10,
        status: 'active',
      );

      setState(() {
        _isLoading = false;
        if (response.success && response.data != null) {
          _auctions = List.from(response.data!['auctions'] ?? []);
        }
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_loading_auctions)),
        );
      }
    }
  }

  void _showLocationPermissionDialog() {
    showDialog(
      context: context,
      barrierColor: Colors.black.withOpacity(0.4),
      builder: (BuildContext context) {
        return Dialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
          elevation: 0,
          backgroundColor: Colors.transparent,
          child: Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Illustration placeholder
                SizedBox(
                  height: 120,
                  width: 120,
                  child: Stack(
                    alignment: Alignment.center,
                    children: [
                      Icon(Icons.language, size: 100, color: Colors.blue[100]),
                      const Positioned(
                        top: 10,
                        child: Icon(Icons.location_on, size: 40, color: Colors.red),
                      ),
                      // Add other dots as per design...
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                Text(
                  AppLocalizations.of(context)!.text_195,
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, height: 1.4),
                ),
                const SizedBox(height: 12),
                Text(
                  AppLocalizations.of(context)!.text_196,
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 14, color: Colors.grey),
                ),
                const SizedBox(height: 24),
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: () {
                      setState(() {
                        _showLocationModal = false;
                      });
                      Navigator.of(context).pop();
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0084FF),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                    child: Text(AppLocalizations.of(context)!.text_197, style: TextStyle(fontWeight: FontWeight.bold)),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        // Pass the Drawer widget as endDrawer so it opens from the left side in RTL
        endDrawer: const SideMenuDrawer(),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          toolbarHeight: 70,
          automaticallyImplyLeading: false, // Build custom layout
          title: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              // Logo/Title on the Right (in RTL)
              Row(
                textDirection: TextDirection.ltr,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    'M',
                    style: TextStyle(
                      fontSize: 26,
                      fontWeight: FontWeight.w900,
                      fontStyle: FontStyle.italic,
                      color: isDarkMode ? Colors.white : const Color(0xFF135BEC),
                    ),
                  ),
                  Text(
                    'azad',
                    style: TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.w900,
                      color: isDarkMode ? Colors.white : Colors.black,
                    ),
                  ),
                  Text(
                    'Pay',
                    style: TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.w900,
                      color: isDarkMode ? Colors.white : const Color(0xFF135BEC),
                    ),
                  ),
                ],
              ),
              // Left side icons (Notifications)
              IconButton(
                icon: Icon(Icons.notifications_outlined, color: isDarkMode ? Colors.white : Colors.black, size: 28),
                onPressed: () {
                   Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const NotificationsPage()),
                  );
                },
              ),
            ],
          ),
        ),
        body: _selectedCityIndex == 1
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    AppLocalizations.of(context)!.text_34,
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey),
                  ),
                  const SizedBox(height: 24),
                  ElevatedButton(
                    onPressed: () {
                      setState(() {
                        _selectedCityIndex = 0;
                      });
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0084FF),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 14),
                    ),
                    child: Text(
                      AppLocalizations.of(context)!.text_198,
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: Colors.white),
                    ),
                  ),
                ],
              ),
            )
          : SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 8),

              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 10.0),
                child: Container(
                  height: 48,
                  decoration: BoxDecoration(
                    color: Colors.transparent,
                    borderRadius: BorderRadius.circular(24),
                    border: Border.all(color: const Color(0xFF0084FF).withOpacity(0.3)),
                  ),
                  child: Row(
                    children: _cities.asMap().entries.map((entry) {
                      int idx = entry.key;
                      String city = entry.value;
                      bool isSelected = _selectedCityIndex == idx;
                      
                      return Expanded(
                        child: GestureDetector(
                          onTap: () {
                            setState(() {
                              _selectedCityIndex = idx;
                            });
                          },
                          child: AnimatedContainer(
                            duration: const Duration(milliseconds: 200),
                            decoration: BoxDecoration(
                              color: isSelected ? primaryBlue : Colors.transparent,
                              borderRadius: BorderRadius.circular(24),
                            ),
                            child: Center(
                              child: Text(
                                city,
                                style: TextStyle(
                                  color: isSelected ? Colors.white : (isDarkMode ? Colors.white70 : Colors.black87),
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ),
              ),

              // Hero Banner (Announcement Carousel)
              SizedBox(
                height: 160,
                width: double.infinity,
                child: PageView(
                  children: [
                    _buildBannerCard(isDarkMode),
                    _buildBannerCard(isDarkMode),
                    _buildBannerCard(isDarkMode),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Live Auctions (مزاد لايف)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20.0),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Text(AppLocalizations.of(context)!.text_199, style: const TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold)),

                        const SizedBox(width: 8),
                        const LiveIndicator(),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              
              // Horizontal list of auctions
              SizedBox(
                height: 240,
                child: _isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : _auctions.isEmpty
                        ? const Center(child: Text('Aucune enchère disponible'))
                        : ListView.builder(
                            scrollDirection: Axis.horizontal,
                            padding: const EdgeInsets.symmetric(horizontal: 16),
                            itemCount: _auctions.length,
                            itemBuilder: (context, index) {
                              final auction = _auctions[index];
                              return _buildAuctionCard(
                                auction['title'] ?? 'Sans titre',
                                '${auction['current_bid'] ?? 0} MRU',
                                auction['ends_at'] ?? '',
                                auction['images']?.cast<String>() ?? ['assets/corolla.png'],
                                auction['id']?.toString() ?? '',
                                isDarkMode,
                              );
                            },
                          ),
              ),
              const SizedBox(height: 16),
              
              // View More Button
              Padding(
                padding: const EdgeInsets.all(20.0),
                child: SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(builder: (context) => const AllAuctionsPage()),
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: primaryBlue,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                    child: Text(AppLocalizations.of(context)!.text_200, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  ),
                ),
              ),


              // Sponsors/Brands (الروعات)
              Padding(
                padding: EdgeInsets.symmetric(horizontal: 20.0),
                child: Text(
                  AppLocalizations.of(context)!.text_201,
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
              ),
              const SizedBox(height: 16),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20.0),
                child: GridView.count(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  crossAxisCount: 2,
                  crossAxisSpacing: 12,
                  mainAxisSpacing: 12,
                  childAspectRatio: 2.2,
                  children: [
                    _buildSponsorLogo('assets/Bankily.png'),
                  ],
                ),
              ),
              const SizedBox(height: 40), // Space for bottom nav
            ],
          ),
        ),
        
        // Custom Bottom Navigation Bar
        floatingActionButton: GestureDetector(
          onTap: () {
            Navigator.of(context).push(
              MaterialPageRoute(builder: (context) => const CreateAdStartPage()),
            );
          },
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                height: 70,
                width: 70,
                child: FloatingActionButton(
                  onPressed: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(builder: (context) => const CreateAdStartPage()),
                    );
                  },
                  backgroundColor: Colors.transparent,
                  elevation: 0,
                  highlightElevation: 0,
                  child: Image.asset(
                    'assets/botum_bar.png',
                    fit: BoxFit.contain,
                  ),
                ),
              ),
              Text(
                "إعلان جديد",
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                  color: Colors.grey[600],
                ),
              ),
            ],
          ),
        ),
        floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
        bottomNavigationBar: BottomAppBar(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          shape: const CircularNotchedRectangle(),
          notchMargin: 8,
          child: SizedBox(
            height: 70, // Increased height
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildNavItem(Icons.home_outlined, Icons.home, AppLocalizations.of(context)!.text_1, 0),
                _buildNavItem(Icons.local_shipping_outlined, Icons.local_shipping, AppLocalizations.of(context)!.text_32, 1),
                const SizedBox(width: 48), // Space for FAB
                _buildNavItem(Icons.storefront_outlined, Icons.storefront, AppLocalizations.of(context)!.text_33, 2),
                _buildNavItem(Icons.person_outline, Icons.person, AppLocalizations.of(context)!.text_19, 3),
              ],
            ),
          ),
        ),
      );
  }

  Widget _buildNavItem(IconData icon, IconData activeIcon, String label, int index) {
    bool isSelected = _currentIndex == index;
    const Color primaryBlue = Color(0xFF0084FF);
    return InkWell(
      onTap: () {
        if (index == 1) {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (context) => ServicesPage()),
          );
        } else if (index == 2) {
          setState(() => _currentIndex = index);
        } else if (index == 3) {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (context) => AccountPage()),
          );
        } else {
          setState(() => _currentIndex = index);
        }
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(isSelected ? activeIcon : icon, color: isSelected ? primaryBlue : Colors.grey[600]),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
              color: isSelected ? primaryBlue : Colors.grey[600],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBannerCard(bool isDarkMode) {
    return SizedBox(
      width: double.infinity,
      height: 160,
      child: Image.asset(
        'assets/announcement.png',
        width: double.infinity,
        height: 160,
        fit: BoxFit.cover,
        errorBuilder: (c, e, s) => Container(
          height: 160,
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
              begin: AlignmentDirectional.centerEnd,
              end: AlignmentDirectional.centerStart,
            ),
          ),
          child: const Center(
            child: Text('Announcement', style: TextStyle(color: Colors.white, fontSize: 16)),
          ),
        ),
      ),
    );
  }

  Widget _buildDot(bool isActive) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 4),
      width: 8, height: 8,
      decoration: BoxDecoration(
        color: isActive ? const Color(0xFF0084FF) : Colors.grey[300],
        shape: BoxShape.circle,
      ),
    );
  }

  Widget _buildAuctionCard(String title, String price, String time, List<String> imageUrls, String id, bool isDarkMode) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(id) ?? false;

    return GestureDetector(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => AuctionDetailsPage(auctionId: id)),
        );
      },
      child: Container(
        width: 140,
        margin: const EdgeInsets.symmetric(horizontal: 8),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFD0D5DD),
            width: 1.5,
          ),
          boxShadow: [
            BoxShadow(
              color: isDarkMode
                  ? Colors.black.withOpacity(0.3)
                  : const Color(0xFF101828).withOpacity(0.08),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Stack(
              children: [
                ClipRRect(
                  borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
                  child: Image.asset(imageUrls[0], height: 120, width: double.infinity, fit: BoxFit.cover, errorBuilder: (c,e,s) => Container(height: 120, color: Colors.grey[200], child: const Icon(Icons.image, color: Colors.grey))),
                ),
 
                Positioned(
                  top: 4, right: 4,
                  child: IconButton(
                    iconSize: 20,
                    onPressed: () {
                      ref.read(favoritesProvider.notifier).toggleFavorite(id);
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(isFavorite ? AppLocalizations.of(context)!.text_55 : AppLocalizations.of(context)!.text_56),
                          duration: const Duration(seconds: 1),
                          behavior: SnackBarBehavior.floating,
                        ),
                      );
                    },
                    icon: Container(
                      padding: const EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.5),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(
                        isFavorite ? Icons.favorite : Icons.favorite_border,
                        color: isFavorite ? Colors.red : Colors.black,
                        size: 16,
                      ),
                    ),
                  ),
                ),
                Positioned(
                  top: 8, left: 8,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(color: const Color(0xFF0084FF), borderRadius: BorderRadius.circular(4)),
                    child: Text(id, style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
                  ),
                ),
              ],
            ),
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(Icons.timer_outlined, color: Colors.red, size: 12),
                      const SizedBox(width: 4),
                      Text(time, style: const TextStyle(color: Colors.red, fontSize: 11, fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(price, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSponsorLogo(String assetPath) {
    return Container(
      width: 90,
      margin: const EdgeInsets.symmetric(horizontal: 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withOpacity(0.2)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 6,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(12),
        child: Image.asset(
          assetPath,
          fit: BoxFit.cover,
          errorBuilder: (c, e, s) => const Center(child: Icon(Icons.storefront, color: Colors.grey)),
        ),
      ),
    );
  }
}