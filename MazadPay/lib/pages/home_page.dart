import 'package:mezadpay/l10n/app_localizations.dart';

import 'package:flutter/material.dart';

import 'package:mezadpay/widgets/side_menu_drawer.dart';

import 'package:mezadpay/pages/account_page.dart';

import 'package:mezadpay/pages/services_page.dart';

import 'package:mezadpay/pages/create_ad_start_page.dart';

import 'package:mezadpay/pages/notifications_page.dart';

import 'package:flutter/services.dart';

import 'package:google_fonts/google_fonts.dart';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:shared_preferences/shared_preferences.dart';



import '../widgets/side_menu_drawer.dart';

import '../widgets/live_indicator.dart';

import 'all_auctions_page.dart';

import 'auction_details_page.dart';

import 'notifications_page.dart';

import '../services/auction_api.dart';

import '../services/category_api.dart';

import '../services/cache_service.dart';

import '../providers/locale_provider.dart';

import '../providers/location_provider.dart';

import '../providers/favorites_provider.dart';



class HomePage extends ConsumerStatefulWidget {

  const HomePage({super.key});



  @override

  ConsumerState<HomePage> createState() => _HomePageState();

}



class _HomePageState extends ConsumerState<HomePage> {

  int _currentIndex = 0;

  bool _showLocationModal = true;

  int _selectedCityIndex = 0;

  List<String> _cities = [];

  List<Map<String, dynamic>> _dbCities = []; // Villes depuis la BD

  Set<String> _citiesWithAuctions = {}; // IDs des villes qui ont des enchères

  bool _isLoadingCities = true;

  final AuctionApi _auctionApi = AuctionApi();

  final CategoryApi _categoryApi = CategoryApi();

  List<Map<String, dynamic>> _auctions = [];

  bool _isLoading = true;



  @override

  void initState() {

    super.initState();

    // Load cities that have auctions (filtered automatically)

    _loadCitiesWithAuctions();

    // Check if location modal was shown before

    _checkFirstTimeLocation();

  }



  Future<void> _loadCitiesWithAuctions() async {

    try {

      // === OPTIMISTIC UI: Charger depuis le cache d'abord (instantané) ===

      final cachedCities = await CacheService.instance.getCachedCities();

      final isCacheValid = await CacheService.instance.isCitiesCacheValid();

      

      if (cachedCities != null && isCacheValid && cachedCities.isNotEmpty) {

        // Afficher les données du cache immédiatement (0ms)

        // Réorganiser les villes du cache: mettre celles avec enchères en premier

        Set<String> cachedCitiesWithAuctions = {};

        // Essayer de récupérer les villes avec enchères depuis le cache des enchères

        final cachedAuctions = await CacheService.instance.getCachedAuctions();

        if (cachedAuctions != null) {

          for (var auction in cachedAuctions) {

            final locationId = auction['location_id']?.toString() ?? 

                              auction['location']?.toString() ??

                              auction['city_id']?.toString();

            if (locationId != null && locationId.isNotEmpty) {

              cachedCitiesWithAuctions.add(locationId);

            }

          }

        }

        

        List<Map<String, dynamic>> citiesWithAuctionsList = [];

        List<Map<String, dynamic>> citiesWithoutAuctionsList = [];

        

        for (var city in cachedCities) {

          final cityId = city['id']?.toString();

          if (cityId != null && cachedCitiesWithAuctions.contains(cityId)) {

            citiesWithAuctionsList.add(city);

          } else {

            citiesWithoutAuctionsList.add(city);

          }

        }

        

        final reorderedCities = [...citiesWithAuctionsList, ...citiesWithoutAuctionsList];

        

        if (!mounted) return;

        setState(() {

          _dbCities = reorderedCities;

          _citiesWithAuctions = cachedCitiesWithAuctions;

          _cities = _dbCities.map((city) => _getCityName(context, city)).toList();

          _isLoadingCities = false;

        });

        

        // Charger les enchères de la première ville depuis le cache aussi

        if (citiesWithAuctionsList.isNotEmpty) {

          final cityId = reorderedCities[0]['id']?.toString();

          final cityAuctions = cachedAuctions?.where((auction) {

            final auctionCityId = auction['location_id']?.toString() ?? 

                                auction['location']?.toString() ??

                                auction['city_id']?.toString();

            return auctionCityId == cityId;

          }).toList();

          

          if (cityAuctions != null && cityAuctions.isNotEmpty) {

            setState(() {

              _selectedCityIndex = 0;

              _auctions = cityAuctions;

              _isLoading = false;

            });

          } else {

            _loadAuctions(locationId: cityId);

          }

        }

      }

      

      // === CHARGER DEPUIS L'API EN ARRIÈRE-PLAN ===

      

      // Étape 1: Charger toutes les enchères actives

      final auctionsResponse = await _auctionApi.getAuctions(

        page: 1,

        limit: 100,

        status: 'active',

      );

      

      // Étape 2: Extraire les IDs des villes qui ont des enchères

      Set<String> citiesWithAuctions = {};

      

      if (auctionsResponse.success && auctionsResponse.data != null) {

        final allAuctions = auctionsResponse.data!.map((item) => item as Map<String, dynamic>).toList();

        

        // Cache les enchères

        await CacheService.instance.cacheAuctions(allAuctions);

        

        for (var auction in allAuctions) {

          final locationId = auction['location_id']?.toString() ?? 

                            auction['location']?.toString() ??

                            auction['city_id']?.toString();

          if (locationId != null && locationId.isNotEmpty) {

            citiesWithAuctions.add(locationId);

          }

        }

      }

      

      // Étape 3: Charger TOUTES les villes depuis l'API

      final locationsResponse = await _categoryApi.getLocations();

      

      if (locationsResponse.success && locationsResponse.data != null) {

        final allLocations = locationsResponse.data!;

        

        if (allLocations is List && allLocations.isNotEmpty) {

          final allCities = allLocations.map((item) => item as Map<String, dynamic>).toList();

          

          // Cache les villes

          await CacheService.instance.cacheCities(allCities);

          

          if (!mounted) return;

          

          // Réorganiser les villes : mettre celles avec enchères en PREMIER

          List<Map<String, dynamic>> citiesWithAuctionsList = [];

          List<Map<String, dynamic>> citiesWithoutAuctionsList = [];

          

          for (var city in allCities) {

            final cityId = city['id']?.toString();

            if (cityId != null && citiesWithAuctions.contains(cityId)) {

              citiesWithAuctionsList.add(city);

            } else {

              citiesWithoutAuctionsList.add(city);

            }

          }

          

          // Combiner : villes avec enchères d'abord, puis celles sans

          final reorderedCities = [...citiesWithAuctionsList, ...citiesWithoutAuctionsList];

          

          // Étape 4: Mettre à jour l'état avec les villes réorganisées (mise à jour silencieuse)

          setState(() {

            _dbCities = reorderedCities;

            _citiesWithAuctions = citiesWithAuctions;

            _cities = _dbCities.map((city) => _getCityName(context, city)).toList();

            _isLoadingCities = false;

          });

          

          // Étape 5: Sélectionner la PREMIÈRE ville (qui a maintenant des enchères)

          if (citiesWithAuctionsList.isNotEmpty) {

            setState(() {

              _selectedCityIndex = 0;

            });

            final cityId = reorderedCities[0]['id']?.toString();

            _loadAuctions(locationId: cityId);

          } else {

            setState(() {

              _auctions = [];

              _isLoading = false;

            });

          }

          return;

        }

      }

      

      // Pas de données - liste vide

      if (!mounted) return;

      setState(() {

        _cities = [];

        _isLoadingCities = false;

      });

    } catch (e) {

      debugPrint('Error loading cities with auctions: $e');

      if (!mounted) return;

      setState(() {

        _cities = [];

        _isLoadingCities = false;

      });

    }

  }



  String _getCityName(BuildContext context, Map<String, dynamic> city) {

    // Récupérer la langue actuelle de l'app

    final locale = Localizations.localeOf(context).languageCode;

    

    // Retourner le nom selon la langue, avec fallback

    String cityName;

    switch (locale) {

      case 'ar':

        cityName = city['city_name_ar']?.toString() ?? 

                   city['name_ar']?.toString() ?? 

                   city['name']?.toString() ?? 

                   city['city_name']?.toString() ?? 

                   'Unknown';

        break;

      case 'fr':

        // Correction spéciale pour Nouadhibou (API retourne arabe dans city_name_fr)

        String frenchName = city['city_name_fr']?.toString() ?? 

                           city['name_fr']?.toString() ?? 

                           city['city_name_ar']?.toString() ?? 

                           city['name']?.toString() ?? 

                           'Unknown';

        

        // Si le nom français contient des caractères arabes, utiliser une traduction statique

        if (frenchName.contains('انواذيبو') || frenchName == 'انواذيبو') {

          cityName = 'Nouadhibou';

        } else if (frenchName.contains('انواكشوط') || frenchName == 'انواكشوط') {

          cityName = 'Nouakchott';

        } else {

          cityName = frenchName;

        }

        break;

      case 'en':

        cityName = city['city_name_en']?.toString() ?? 

                   city['name_en']?.toString() ?? 

                   city['city_name_ar']?.toString() ?? 

                   city['name']?.toString() ?? 

                   'Unknown';

        

        // Correction pour l'anglais aussi

        if (cityName.contains('انواذيبو') || cityName == 'انواذيبو') {

          cityName = 'Nouadhibou';

        } else if (cityName.contains('انواكشوط') || cityName == 'انواكشوط') {

          cityName = 'Nouakchott';

        }

        break;

      default:

        cityName = city['city_name_ar']?.toString() ?? 

                   city['name']?.toString() ?? 

                   'Unknown';

    }

    

    return cityName;

  }



  Future<void> _checkFirstTimeLocation() async {

    final prefs = await SharedPreferences.getInstance();

    final hasShownLocationModal = prefs.getBool('has_shown_location_modal') ?? false;

    

    if (!hasShownLocationModal && mounted) {

      final locationState = ref.read(locationProvider);

      if (!locationState.hasDetected) {

        // Delay to ensure context is ready

        Future.delayed(const Duration(milliseconds: 500), () {

          if (mounted) {

            _showLocationPermissionDialog();

          }

        });

      }

    }

  }



  Future<void> _markLocationModalShown() async {

    final prefs = await SharedPreferences.getInstance();

    await prefs.setBool('has_shown_location_modal', true);

  }



  Future<void> _loadAuctions({String? locationId}) async {

    try {

      final response = await _auctionApi.getAuctions(

        page: 1,

        limit: 10,

        status: 'active',

        locationId: locationId,

      );



      setState(() {

        _isLoading = false;

        if (response.success && response.data != null) {

          // response.data est maintenant directement une List<dynamic>

          _auctions = response.data!.map((item) => item as Map<String, dynamic>).toList();

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

                    onPressed: () async {

                      // Mark modal as shown so it doesn't appear again

                      await _markLocationModalShown();

                      if (mounted) {

                        Navigator.of(context).pop();

                      }

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

              // Location indicator

              Consumer(builder: (context, ref, child) {

                final locationState = ref.watch(locationProvider);

                final locationNotifier = ref.read(locationProvider.notifier);



                return Row(

                  children: [

                    // Location chip

                    GestureDetector(

                      onTap: () => locationNotifier.refreshLocation(),

                      child: Container(

                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),

                        decoration: BoxDecoration(

                          color: isDarkMode 

                              ? const Color(0xFF2D2D2D) 

                              : const Color(0xFFE8F5E9),

                          borderRadius: BorderRadius.circular(20),

                          border: Border.all(

                            color: const Color(0xFF2E7D32).withOpacity(0.3),

                          ),

                        ),

                        child: Row(

                          mainAxisSize: MainAxisSize.min,

                          children: [

                            Icon(

                              Icons.location_on,

                              size: 14,

                              color: const Color(0xFF2E7D32),

                            ),

                            const SizedBox(width: 4),

                            if (locationState.isLoading)

                              SizedBox(

                                width: 12,

                                height: 12,

                                child: CircularProgressIndicator(

                                  strokeWidth: 2,

                                  valueColor: AlwaysStoppedAnimation<Color>(

                                    const Color(0xFF2E7D32),

                                  ),

                                ),

                              )

                            else

                              Text(

                                locationState.location?.displayName ?? '...',

                                style: TextStyle(

                                  fontSize: 12,

                                  fontWeight: FontWeight.w600,

                                  color: const Color(0xFF2E7D32),

                                ),

                              ),

                          ],

                        ),

                      ),

                    ),

                    const SizedBox(width: 8),

                    // Notifications

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

                );

              }),

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



              // City tabs - only show if cities are loaded

              if (_cities.isNotEmpty)

              Padding(

                padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 10.0),

                child: Container(

                  height: 48,

                  decoration: BoxDecoration(

                    color: Colors.transparent,

                    borderRadius: BorderRadius.circular(24),

                    border: Border.all(color: const Color(0xFF0084FF).withOpacity(0.3)),

                  ),

                  child: SingleChildScrollView(

                    scrollDirection: Axis.horizontal,

                    child: Row(

                      mainAxisSize: MainAxisSize.min,

                      children: _cities.asMap().entries.map((entry) {

                        int idx = entry.key;

                        String city = entry.value;

                        bool isSelected = _selectedCityIndex == idx;

                        

                        return GestureDetector(

                          onTap: () {

                            setState(() {

                              _selectedCityIndex = idx;

                            });

                            // Recharger les enchères seulement si la ville en a

                            if (_dbCities.isNotEmpty && idx < _dbCities.length) {

                              final selectedCity = _dbCities[idx];

                              final cityId = selectedCity['id']?.toString();

                              

                              // Vérifier si cette ville a des enchères

                              if (cityId != null && _citiesWithAuctions.contains(cityId)) {

                                _loadAuctions(locationId: cityId);

                              } else {

                                // Ville sans enchères - vider la liste

                                setState(() {

                                  _auctions = [];

                                  _isLoading = false;

                                });

                              }

                            }

                          },

                          child: AnimatedContainer(

                            duration: const Duration(milliseconds: 200),

                            margin: const EdgeInsets.symmetric(horizontal: 4),

                            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),

                            decoration: BoxDecoration(

                              color: isSelected ? primaryBlue : Colors.transparent,

                              borderRadius: BorderRadius.circular(24),

                            ),

                            child: Text(

                              city,

                              style: TextStyle(

                                color: isSelected ? Colors.white : (isDarkMode ? Colors.white70 : Colors.black87),

                                fontWeight: FontWeight.bold,

                              ),

                            ),

                          ),

                        );

                      }).toList(),

                    ),

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

                        LiveIndicator(),

                      ],

                    ),

                  ],

                ),

              ),

              const SizedBox(height: 16),

              

              // Horizontal list of auctions

              SizedBox(

                height: 210,

                child: _isLoading

                    ? const Center(child: CircularProgressIndicator())

                    : _auctions.isEmpty

                        ? _buildEmptyAuctionsWidget(context, isDarkMode)

                        : ListView.builder(

                            scrollDirection: Axis.horizontal,

                            padding: const EdgeInsets.symmetric(horizontal: 16),

                            itemCount: _auctions.length,

                            itemBuilder: (context, index) {

                              final auction = _auctions[index];

                              return _buildAuctionCard(

                                auction,

                                isDarkMode,

                                index + 1,

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



  Widget _buildEmptyAuctionsWidget(BuildContext context, bool isDarkMode) {

    return Center(

      child: Column(

        mainAxisAlignment: MainAxisAlignment.center,

        children: [

          Icon(

            Icons.location_off_outlined,

            size: 48,

            color: isDarkMode ? Colors.grey[400] : Colors.grey[500],

          ),

          const SizedBox(height: 16),

          Text(

            AppLocalizations.of(context)?.page_not_available ?? 'Cette page n\'est pas accessible pour le moment',

            textAlign: TextAlign.center,

            style: TextStyle(

              fontSize: 16,

              fontWeight: FontWeight.w500,

              color: isDarkMode ? Colors.grey[300] : Colors.grey[600],

            ),

          ),

          const SizedBox(height: 8),

          Text(

            AppLocalizations.of(context)?.return_to_first_city ?? 'Retournez à la première ville',

            textAlign: TextAlign.center,

            style: TextStyle(

              fontSize: 14,

              color: isDarkMode ? Colors.grey[400] : Colors.grey[500],

            ),

          ),

          const SizedBox(height: 24),

          ElevatedButton.icon(

            onPressed: () {

              // Trouver la première ville qui a des enchères

              int firstCityWithAuctionsIndex = -1;

              for (int i = 0; i < _dbCities.length; i++) {

                final cityId = _dbCities[i]['id']?.toString();

                if (cityId != null && _citiesWithAuctions.contains(cityId)) {

                  firstCityWithAuctionsIndex = i;

                  break;

                }

              }

              

              if (firstCityWithAuctionsIndex != -1) {

                setState(() {

                  _selectedCityIndex = firstCityWithAuctionsIndex;

                });

                final cityId = _dbCities[firstCityWithAuctionsIndex]['id']?.toString();

                _loadAuctions(locationId: cityId);

              }

            },

            icon: const Icon(Icons.arrow_back, size: 18),

            label: Text(AppLocalizations.of(context)?.return_button ?? 'Retour'),

            style: ElevatedButton.styleFrom(

              backgroundColor: const Color(0xFF0084FF),

              foregroundColor: Colors.white,

              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),

              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),

            ),

          ),

        ],

      ),

    );

  }





  Widget _buildAuctionCard(Map<String, dynamic> auction, bool isDarkMode, int displayNumber) {

    final favoritesAsync = ref.watch(favoritesProvider);

    final id = auction['id']?.toString() ?? '';

    final isFavorite = favoritesAsync.value?.contains(id) ?? false;

    

    // Extraction des données API avec fallbacks intelligents pour le titre

    final locale = Localizations.localeOf(context).languageCode;

    String title = '';

    

    // 1. Essayer d'abord la langue actuelle de l'app

    switch (locale) {

      case 'ar':

        title = auction['title_ar']?.toString() ?? '';

        break;

      case 'fr':

        title = auction['title_fr']?.toString() ?? '';

        break;

      case 'en':

        title = auction['title_en']?.toString() ?? '';

        break;

    }

    

    // 2. Si vide, essayer l'arabe (langue par défaut)

    if (title.isEmpty) {

      title = auction['title_ar']?.toString() ?? '';

    }

    

    // 3. Si toujours vide, essayer les autres langues

    if (title.isEmpty) {

      title = auction['title_fr']?.toString() ??

              auction['title_en']?.toString() ??

              auction['title']?.toString() ??

              'Sans titre';

    }

    

    final price = '${auction['current_price'] ?? auction['current_bid'] ?? 0} MRU';

    final bidCount = auction['bidder_count'] ?? auction['bids'] ?? 0;

    

    // Formatage du temps restant

    String time = '';

    if (auction['end_time'] != null) {

      try {

        final endTime = DateTime.parse(auction['end_time']);

        final now = DateTime.now();

        final diff = endTime.difference(now);

        

        if (diff.isNegative) {

          time = 'Terminé';

        } else if (diff.inDays > 0) {

          time = '${diff.inDays}j ${diff.inHours % 24}h restants';

        } else if (diff.inHours > 0) {

          time = '${diff.inHours}h ${diff.inMinutes % 60}m restants';

        } else {

          time = '${diff.inMinutes}m restants';

        }

      } catch (e) {

        time = auction['end_time']?.toString() ?? '';

      }

    }

    

    // Gestion des images - utiliser le tableau images depuis l'API (affiche seulement la première)

    String imageUrl = 'assets/corolla.png';

    bool isNetworkImage = false;

    

    // Vérifier si l'API retourne images (tableau de toutes les images)

    if (auction['images'] != null && auction['images'] is List && (auction['images'] as List).isNotEmpty) {

      final images = auction['images'] as List;

      if (images.isNotEmpty && images[0] != null) {

        imageUrl = images[0].toString();

        isNetworkImage = imageUrl.startsWith('http');

      }

    }



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

                  child: isNetworkImage

                    ? Image.network(

                        imageUrl,

                        height: 110,

                        width: double.infinity,

                        fit: BoxFit.cover,

                        errorBuilder: (c, e, s) => Image.asset(

                          'assets/corolla.png',

                          height: 110,

                          width: double.infinity,

                          fit: BoxFit.cover,

                        ),

                        loadingBuilder: (context, child, loadingProgress) {

                          if (loadingProgress == null) return child;

                          return Container(

                            height: 110,

                            color: Colors.grey[200],

                            child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),

                          );

                        },

                      )

                    : Image.asset(

                        imageUrl,

                        height: 110,

                        width: double.infinity,

                        fit: BoxFit.cover,

                        errorBuilder: (c, e, s) => Container(height: 110, color: Colors.grey[200], child: const Icon(Icons.image, color: Colors.grey)),

                      ),

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

                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),

                    decoration: BoxDecoration(

                      color: const Color(0xFF0084FF),

                      borderRadius: BorderRadius.circular(20),

                      boxShadow: [

                        BoxShadow(

                          color: Colors.black.withOpacity(0.2),

                          blurRadius: 4,

                          offset: const Offset(0, 2),

                        ),

                      ],

                    ),

                    child: Text(

                      '#$displayNumber',

                      style: const TextStyle(

                        color: Colors.white,

                        fontSize: 11,

                        fontWeight: FontWeight.bold,

                      ),

                    ),

                  ),

                ),

              ],

            ),

            Padding(

              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),

              child: Column(

                crossAxisAlignment: CrossAxisAlignment.start,

                children: [

                  Text(title, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12)),

                  const SizedBox(height: 4),

                  Row(

                    children: [

                      const Icon(Icons.timer_outlined, color: Colors.red, size: 12),

                      const SizedBox(width: 4),

                      Flexible(

                        child: Text(

                          time,

                          style: const TextStyle(color: Colors.red, fontSize: 11, fontWeight: FontWeight.bold),

                          overflow: TextOverflow.ellipsis,

                          maxLines: 1,

                        ),

                      ),

                    ],

                  ),

                  const SizedBox(height: 4),

                  Row(

                    mainAxisAlignment: MainAxisAlignment.spaceBetween,

                    children: [

                      Flexible(

                        child: Text(

                          price,

                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12),

                          overflow: TextOverflow.ellipsis,

                          maxLines: 1,

                        ),

                      ),

                      Row(

                        mainAxisSize: MainAxisSize.min,

                        children: [

                          const Icon(Icons.gavel_outlined, size: 12, color: Colors.grey),

                          const SizedBox(width: 2),

                          Text(

                            '$bidCount',

                            style: const TextStyle(fontSize: 10, color: Colors.grey),

                          ),

                        ],

                      ),

                    ],

                  ),

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