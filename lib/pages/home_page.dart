import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_page.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _currentIndex = 0;
  bool _showLocationModal = true;
  int _selectedCityIndex = 0;
  final List<String> _cities = ['انواكشوط', 'انواذيبو'];

  @override
  void initState() {
    super.initState();
    // Show the modal after the first frame is rendered
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_showLocationModal) {
        _showLocationPermissionDialog();
      }
    });
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
                const Text(
                  'خدمات مزاد موريتانيا بحاجة إلى\nموقعك',
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, height: 1.4),
                ),
                const SizedBox(height: 12),
                const Text(
                  'لتجربة أفضل يرجى تفعيل الموقع الجغرافي في هاتفك',
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
                    child: const Text('تفعيل خدمة تحديد الموقع', style: TextStyle(fontWeight: FontWeight.bold)),
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

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
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
              // Center: Logo
              Image.asset('logo.png', height: 35, errorBuilder: (c, e, s) => const Text('MazadPay', style: TextStyle(color: Colors.black, fontWeight: FontWeight.bold))),
              // Right: Notifications
              IconButton(
                icon: Icon(Icons.notifications_outlined, color: isDarkMode ? Colors.white : Colors.black, size: 28),
                onPressed: () {},
              ),
            ],
          ),
        ),
        body: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Promo Banner
              Padding(
                padding: const EdgeInsets.all(20.0),
                child: Container(
                  width: double.infinity,
                  height: 100,
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
                      begin: Alignment.centerRight,
                      end: Alignment.centerLeft,
                    ),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: const [
                        Text(
                          'اربح وقتك',
                          style: TextStyle(color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                  ),
                ),
              ),

              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 10.0),
                child: Container(
                  height: 48,
                  decoration: BoxDecoration(
                    color: isDarkMode ? const Color(0xFF1D1D1D) : const Color(0xFFE8F0FE),
                    borderRadius: BorderRadius.circular(24),
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
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 8.0),
                child: SizedBox(
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
              ),

              // Dots indicator for banner
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  _buildDot(true), _buildDot(false), _buildDot(false),
                ],
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
                        const Text(
                          'مزاد لايف',
                          style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(width: 8),
                        Container(width: 10, height: 10, decoration: const BoxDecoration(color: Colors.red, shape: BoxShape.circle)),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              
              // Horizontal list of auctions
              SizedBox(
                height: 240,
                child: ListView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  children: [
                    _buildAuctionCard('سمعه للبيع كرفور', 'MRU 2,500,000', '23h 42m 15s', 'smaah.png', '7', isDarkMode),
                    _buildAuctionCard('12 Pro Max', 'MRU 12,000', '5h 42m 15s', 'iphone.png', '17', isDarkMode),
                    _buildAuctionCard('كورولا 2017', 'MRU 285,000', '23h 42m 15s', 'corolla.png', '12', isDarkMode),
                    _buildAuctionCard('راف 4 2017', 'MRU 320,000', '11h 20m 5s', 'raf4.png', '3', isDarkMode),
                    _buildAuctionCard('دار في ربينة', 'MRU 5,000,000', '2h 10m 30s', 'house.png', '5', isDarkMode),
                    _buildAuctionCard('لابتوب Lenovo', 'MRU 35,000', '18h 55m 00s', 'laptop.png', '9', isDarkMode),
                  ],
                ),
              ),
              
              // View More Button
              Padding(
                padding: const EdgeInsets.all(20.0),
                child: SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: () {},
                    style: ElevatedButton.styleFrom(
                      backgroundColor: primaryBlue,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                    child: const Text('عرض مزيد من المزادات', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  ),
                ),
              ),

              // Sponsors/Brands (الماركات)
              const Padding(
                padding: EdgeInsets.symmetric(horizontal: 20.0),
                child: Text(
                  'الماركات',
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
                    _buildSponsorLogo('Bankily.png'),
                    _buildSponsorLogo('Masrivi.png'),
                    _buildSponsorLogo('Click.png'),
                    _buildSponsorLogo('Sedad.png'),
                  ],
                ),
              ),
              const SizedBox(height: 40), // Space for bottom nav
            ],
          ),
        ),
        
        // Custom Bottom Navigation Bar
        floatingActionButton: Padding(
          padding: const EdgeInsets.only(top: 20.0),
          child: SizedBox(
            height: 70,
            width: 70,
            child: FloatingActionButton(
              onPressed: () {},
              backgroundColor: Colors.transparent,
              elevation: 0,
              highlightElevation: 0,
              child: Image.asset(
                'botum_bar.png',
                fit: BoxFit.contain,
              ),
            ),
          ),
        ),
        floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
        bottomNavigationBar: BottomAppBar(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          shape: const CircularNotchedRectangle(),
          notchMargin: 8,
          child: SizedBox(
            height: 60,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildNavItem(Icons.home_outlined, Icons.home, 'الرئيسية', 0),
                _buildNavItem(Icons.local_shipping_outlined, Icons.local_shipping, 'توصيل', 1),
                const SizedBox(width: 48), // Space for FAB
                _buildNavItem(Icons.storefront_outlined, Icons.storefront, 'التجارة الالكترونية', 2),
                _buildNavItem(Icons.person_outline, Icons.person, 'حسابي', 3),
              ],
            ),
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
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Image.asset(
          'announcement.png',
          width: double.infinity,
          fit: BoxFit.fill,
          errorBuilder: (c, e, s) => Container(
            decoration: BoxDecoration(
              color: const Color(0xFF81B8E8),
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Center(child: Text('Announcement', style: TextStyle(color: Colors.white))),
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

  Widget _buildAuctionCard(String title, String price, String time, String imageUrl, String id, bool isDarkMode) {
    return Container(
      width: 140,
      margin: const EdgeInsets.symmetric(horizontal: 8),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.withOpacity(0.2)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Stack(
            children: [
              ClipRRect(
                borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
                child: Image.asset(imageUrl, height: 120, width: double.infinity, fit: BoxFit.cover, errorBuilder: (c,e,s) => Container(height: 120, color: Colors.grey[200], child: const Icon(Icons.image, color: Colors.grey))),
              ),
              Positioned(
                top: 8, left: 8,
                child: Container(
                  padding: const EdgeInsets.all(4),
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
