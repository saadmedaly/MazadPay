import 'package:mezadpay/l10n/app_localizations.dart';

import 'package:flutter/material.dart';

import 'package:mezadpay/pages/create_ad_start_page.dart';

import 'package:flutter/services.dart';

import 'package:google_fonts/google_fonts.dart';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:shared_preferences/shared_preferences.dart';



import '../widgets/side_menu_drawer.dart';
import '../widgets/live_indicator.dart';

import 'all_auctions_page.dart';

import 'auction_details_page.dart';

import '../services/auction_api.dart';

import '../services/category_api.dart';

import '../services/cache_service.dart';

import '../services/banner_api.dart';
import '../services/sponsor_api.dart';
import 'notifications_page.dart';
import 'my_auctions_page.dart';
import 'services_page.dart';
import 'account_page.dart';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:shimmer/shimmer.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'package:mezadpay/providers/location_provider.dart';
import 'package:mezadpay/providers/home_provider.dart';
import 'package:mezadpay/models/auction.dart';
import 'package:mezadpay/utils/time_utils.dart';



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

  final BannerApi _bannerApi = BannerApi();

  List<Map<String, dynamic>> _banners = []; // Bannières depuis la BD

  bool _isLoadingBanners = true;

  final SponsorApi _sponsorApi = SponsorApi();

  List<Map<String, dynamic>> _sponsors = [];

  bool _isLoadingSponsors = true;

  List<Map<String, dynamic>> _auctions = [];

  bool _isLoading = true;



  @override

  void initState() {

    super.initState();

    // Load cities that have auctions (filtered automatically)

    _loadCitiesWithAuctions();

    // Load banners from database

    _loadBanners();

    // Load sponsors from database

    _loadSponsors();

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

        _sortCitiesByUserLocation(citiesWithAuctionsList);
        _sortCitiesByUserLocation(citiesWithoutAuctionsList);

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

            if (!mounted) return;
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

          _sortCitiesByUserLocation(citiesWithAuctionsList);
          _sortCitiesByUserLocation(citiesWithoutAuctionsList);

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

            if (!mounted) return;
            setState(() {

              _selectedCityIndex = 0;

            });

            final cityId = reorderedCities[0]['id']?.toString();

            _loadAuctions(locationId: cityId);

          } else {

            if (!mounted) return;
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



  void _sortCitiesByUserLocation(List<Map<String, dynamic>> citiesList) {
    if (citiesList.isEmpty) return;

    final locationState = ref.read(locationProvider);
    final userLat = locationState.location?.latitude;
    final userLng = locationState.location?.longitude;
    final userCityName = locationState.location?.city ?? locationState.location?.displayName;

    // 1. Essayer de trouver une correspondance exacte par nom
    if (userCityName != null && userCityName.isNotEmpty) {
      final exactMatchIndex = citiesList.indexWhere((c) {
        final nameAr = c['name_ar']?.toString() ?? c['city_name_ar']?.toString() ?? '';
        final nameFr = c['name_fr']?.toString() ?? c['city_name_fr']?.toString() ?? '';
        final nameEn = c['name_en']?.toString() ?? c['city_name_fr']?.toString() ?? '';
        final name = c['name']?.toString() ?? '';
        
        final cleanUserCity = userCityName.toLowerCase().trim();
        
        return nameAr.toLowerCase().trim() == cleanUserCity ||
               nameFr.toLowerCase().trim() == cleanUserCity ||
               nameEn.toLowerCase().trim() == cleanUserCity ||
               name.toLowerCase().trim() == cleanUserCity ||
               (nameFr.isNotEmpty && nameFr.toLowerCase().contains(cleanUserCity)) ||
               (cleanUserCity.length > 3 && nameFr.toLowerCase().contains(cleanUserCity.substring(0, cleanUserCity.length - 1)));
      });

      if (exactMatchIndex != -1) {
        final city = citiesList.removeAt(exactMatchIndex);
        citiesList.insert(0, city);
        return;
      }
    }

    // 2. Si pas de correspondance par nom, utiliser la distance
    if (userLat != null && userLng != null && userLat != 0.0 && userLng != 0.0) {
      citiesList.sort((a, b) {
        final latA = double.tryParse(a['latitude']?.toString() ?? a['lat']?.toString() ?? '');
        final lngA = double.tryParse(a['longitude']?.toString() ?? a['lng']?.toString() ?? '');
        
        final latB = double.tryParse(b['latitude']?.toString() ?? b['lat']?.toString() ?? '');
        final lngB = double.tryParse(b['longitude']?.toString() ?? b['lng']?.toString() ?? '');

        if (latA == null || lngA == null) return 1;
        if (latB == null || lngB == null) return -1;

        final distA = (latA - userLat) * (latA - userLat) + (lngA - userLng) * (lngA - userLng);
        final distB = (latB - userLat) * (latB - userLat) + (lngB - userLng) * (lngB - userLng);

        return distA.compareTo(distB);
      });
      return;
    }

    // 3. FALLBACK: Si pas de localisation, mettre Nouakchott en premier
    final nouakchottIndex = citiesList.indexWhere((c) {
      final nameFr = c['name_fr']?.toString() ?? c['city_name_fr']?.toString() ?? '';
      return nameFr.toLowerCase().contains('nouakchott');
    });
    if (nouakchottIndex != -1) {
      final city = citiesList.removeAt(nouakchottIndex);
      citiesList.insert(0, city);
    }
  }

  String _getCityName(BuildContext context, Map<String, dynamic> city) {
    final locale = Localizations.localeOf(context).languageCode;
    
    if (locale == 'ar') {
      return city['city_name_ar']?.toString() ?? 
             city['name_ar']?.toString() ?? 
             city['name']?.toString() ?? 
             'غير معروف';
    } else {
      // Pour FR et EN, utiliser city_name_fr comme demandé
      String name = city['city_name_fr']?.toString() ?? 
                    city['name_fr']?.toString() ?? 
                    city['name']?.toString() ?? 
                    'Unknown';
      
      // Overrides pour les noms qui reviennent en Arabe de l'API
      if (name.contains('انواذيبو') || name == 'انواذيبو') return 'Nouadhibou';
      if (name.contains('انواكشوط') || name == 'انواكشوط') return 'Nouakchott';
      if (name.contains('أطار') || name == 'أطار') return 'Atar';
      if (name.contains('روصو') || name == 'روصو') return 'Rosso';
      
      return name;
    }
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

  Future<void> _loadBanners() async {
    try {
      debugPrint('=== LOADING BANNERS ===');
      
      final cachedBanners = await CacheService.instance.getCachedBanners();
      final isCacheValid = await CacheService.instance.isBannersCacheValid();
      
      if (cachedBanners != null && cachedBanners.isNotEmpty && isCacheValid) {
        if (!mounted) return;
        setState(() {
          _banners = cachedBanners;
          _isLoadingBanners = false;
        });
      }
      
      final response = await _bannerApi.getBanners();
      
      if (response.success && response.data != null) {
        final List<dynamic> bannerList = response.data!;
        debugPrint('Loaded ${bannerList.length} banners');
        
        final activeBanners = bannerList
            .map((item) => item as Map<String, dynamic>)
            .where((banner) => banner['is_active'] == true)
            .toList();
            
        await CacheService.instance.cacheBanners(activeBanners);
        
        debugPrint('Active banners: ${activeBanners.length}');
        
        if (!mounted) return;
        setState(() {
          _banners = activeBanners;
          _isLoadingBanners = false;
        });
      } else {
        debugPrint('Failed to load banners: ${response.message}');
        if (_banners.isEmpty && mounted) {
          setState(() {
            _isLoadingBanners = false;
          });
        }
      }
    } catch (e) {
      debugPrint('Error loading banners: $e');
      if (_banners.isEmpty && mounted) {
        setState(() {
          _isLoadingBanners = false;
        });
      }
    }
  }

  Future<void> _loadSponsors() async {
    try {
      final response = await _sponsorApi.getSponsors();
      
      if (response.success && response.data != null) {
        final List<dynamic> sponsorList = response.data!;
        
        if (!mounted) return;
        setState(() {
          _sponsors = sponsorList.map((item) => item as Map<String, dynamic>).toList();
          _isLoadingSponsors = false;
        });
      } else {
        if (mounted) {
          setState(() {
            _isLoadingSponsors = false;
          });
        }
      }
    } catch (e) {
      debugPrint('Error loading sponsors: $e');
      if (mounted) {
        setState(() {
          _isLoadingSponsors = false;
        });
      }
    }
  }


  /// Helper pour obtenir des textes localisés sans dépendre des clés AppLocalizations
  String _getLocalizedText(BuildContext context, String key, String defaultValue) {
    final locale = Localizations.localeOf(context).languageCode;
    
    // Map des traductions pour les clés temporelles
    final Map<String, Map<String, String>> translations = {
      'no_title': {
        'ar': 'بدون عنوان',
        'fr': 'Sans titre',
        'en': 'No title',
      },
      'ended': {
        'ar': 'انتهى',
        'fr': 'Terminé',
        'en': 'Ended',
      },
      'days_abbr': {
        'ar': 'ي',
        'fr': 'j',
        'en': 'd',
      },
      'hours_abbr': {
        'ar': 'س',
        'fr': 'h',
        'en': 'h',
      },
      'minutes_abbr': {
        'ar': 'د',
        'fr': 'm',
        'en': 'm',
      },
      'remaining': {
        'ar': 'متبقي',
        'fr': 'restants',
        'en': 'remaining',
      },
      'new_ad': {
        'ar': 'إعلان جديد',
        'fr': 'Nouvelle annonce',
        'en': 'New Ad',
      },
    };
    
    return translations[key]?[locale] ?? defaultValue;
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



      if (!mounted) return;
      setState(() {

        _isLoading = false;

        if (response.success && response.data != null) {

          // response.data est maintenant directement une List<dynamic>

          _auctions = response.data!.map((item) => item as Map<String, dynamic>).toList();

        }

      });

    } catch (e) {

      if (!mounted) return;
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

          ),

        ),

        body: IndexedStack(
          index: _currentIndex,
          children: [
            // Tab 0: الرئيسية
            _selectedCityIndex == 1
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.location_off_outlined, size: 64, color: Colors.grey[400]),
                      const SizedBox(height: 16),
                      const Text(
                        'لا توجد مزادات في انواذيبو حالياً',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 24),
                      ElevatedButton.icon(
                        onPressed: () => setState(() => _selectedCityIndex = 0),
                        icon: const Icon(Icons.arrow_forward, color: Colors.white),
                        label: const Text(
                          'العودة إلى الصفحة الرئيسية',
                          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: Colors.white),
                        ),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF0084FF),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                          padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 14),
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



              // City tabs — fixed list: نواكشوط / انواذيبو
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 10.0),
                child: Container(
                  height: 48,
                  decoration: BoxDecoration(
                    color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                    borderRadius: BorderRadius.circular(24),
                    border: Border.all(color: const Color(0xFF0084FF).withOpacity(0.3)),
                  ),
                  child: Row(
                    children: [
                      // نواكشوط — index 0 — RIGHT (RTL first)
                      Expanded(
                        child: GestureDetector(
                          onTap: () {
                            if (_selectedCityIndex == 0) return;
                            setState(() {
                              _selectedCityIndex = 0;
                              _isLoading = true;
                            });
                            _loadAuctions();
                          },
                          child: AnimatedContainer(
                            duration: const Duration(milliseconds: 200),
                            margin: const EdgeInsets.all(4),
                            decoration: BoxDecoration(
                              color: _selectedCityIndex == 0 ? primaryBlue : Colors.transparent,
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Center(
                              child: Text(
                                'نواكشوط',
                                style: TextStyle(
                                  color: _selectedCityIndex == 0 ? Colors.white : (isDarkMode ? Colors.white70 : Colors.black87),
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                      // انواذيبو — index 1 — LEFT
                      Expanded(
                        child: GestureDetector(
                          onTap: () {
                            if (_selectedCityIndex == 1) return;
                            setState(() {
                              _selectedCityIndex = 1;
                            });
                          },
                          child: AnimatedContainer(
                            duration: const Duration(milliseconds: 200),
                            margin: const EdgeInsets.all(4),
                            decoration: BoxDecoration(
                              color: _selectedCityIndex == 1 ? primaryBlue : Colors.transparent,
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Center(
                              child: Text(
                                'انواذيبو',
                                style: TextStyle(
                                  color: _selectedCityIndex == 1 ? Colors.white : (isDarkMode ? Colors.white70 : Colors.black87),
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),



              // Hero Banner (Announcement Carousel)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0),
                child: AspectRatio(
                  aspectRatio: 16 / 7,
                  child: PageView(
                    children: _isLoadingBanners
                        ? [_buildBannerCard(isDarkMode)]
                        : _banners.isEmpty
                            ? [_buildBannerCard(isDarkMode)]
                            : _banners.map((banner) => _buildDynamicBannerCard(banner, isDarkMode)).toList(),
                  ),
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
                child: ref.watch(homeAuctionsProvider).when(
                  data: (auctions) {
                    if (auctions.isEmpty) return _buildEmptyAuctionsWidget(context, isDarkMode);
                    return ListView.builder(
                      scrollDirection: Axis.horizontal,
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: auctions.length,
                      itemBuilder: (context, index) {
                        return _buildAuctionCard(context, auctions[index], isDarkMode, index + 1);
                      },
                    );
                  },
                  loading: () => const Center(child: CircularProgressIndicator()),
                  error: (err, stack) => Center(child: Text('Error: $err')),
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

              if (_isLoadingSponsors || _sponsors.isNotEmpty)
                Padding(
                  padding: EdgeInsets.symmetric(horizontal: 20.0),
                  child: Text(
                    AppLocalizations.of(context)!.text_201,
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                ),

              if (_isLoadingSponsors || _sponsors.isNotEmpty)
                const SizedBox(height: 16),

              if (_isLoadingSponsors)
                const Center(child: CircularProgressIndicator())
              else if (_sponsors.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20.0),
                  child: GridView.count(
                    shrinkWrap: true,
                    physics: const NeverScrollableScrollPhysics(),
                    crossAxisCount: 2,
                    crossAxisSpacing: 12,
                    mainAxisSpacing: 12,
                    childAspectRatio: 2.2,
                    children: _sponsors.map((sponsor) {
                      final logoUrl = sponsor['image_url'] ?? '';
                      return Container(
                        decoration: BoxDecoration(
                          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: Colors.grey.withOpacity(0.1)),
                        ),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(12),
                          child: logoUrl.isNotEmpty
                              ? CachedNetworkImage(
                                  imageUrl: logoUrl,
                                  fit: BoxFit.contain,
                                  placeholder: (context, url) => Shimmer.fromColors(
                                    baseColor: Colors.grey[300]!,
                                    highlightColor: Colors.grey[100]!,
                                    child: Container(color: Colors.white),
                                  ),
                                  errorWidget: (context, url, error) =>
                                      const Icon(Icons.image_not_supported, color: Colors.grey),
                                )
                              : const Icon(Icons.image, color: Colors.grey),
                        ),
                      );
                    }).toList(),
                  ),
                ),

              const SizedBox(height: 40), // Space for bottom nav

            ],

          ),

        ),

            // Tab 1: توصيل
            const ServicesPage(),
            // Tab 2: مزاداتي — opened via Navigator.push, never shown via IndexedStack
            // (index 2 pushes MyAuctionsPage directly, so we put a placeholder here)
            const SizedBox.shrink(),
            // Tab 3: حسابي
            const AccountPage(),
          ],
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
                _getLocalizedText(context, 'new_ad', "إعلان جديد"),
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

                _buildNavItem(Icons.gavel_outlined, Icons.gavel, AppLocalizations.of(context)!.text_27, 2),

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
        if (index == 2) {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const MyAuctionsPage()),
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
    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: Image.asset(
        'assets/announcement.png',
        width: double.infinity,
        fit: BoxFit.contain,
        errorBuilder: (c, e, s) => Container(
          color: Colors.grey[300],
          child: const Center(
            child: Icon(Icons.broken_image, size: 48, color: Colors.grey),
          ),
        ),
      ),
    );
  }
  
  Widget _buildDynamicBannerCard(Map<String, dynamic> banner, bool isDarkMode) {
    final imageUrl = banner['image_url']?.toString() ?? '';
    
    final locale = Localizations.localeOf(context).languageCode;
    String title = '';
    switch (locale) {
      case 'ar':
        title = banner['title_ar']?.toString() ?? banner['title']?.toString() ?? '';
        break;
      case 'fr':
        title = banner['title_fr']?.toString() ?? banner['title_ar']?.toString() ?? banner['title']?.toString() ?? '';
        break;
      case 'en':
        title = banner['title_en']?.toString() ?? banner['title_ar']?.toString() ?? banner['title']?.toString() ?? '';
        break;
      default:
        title = banner['title_ar']?.toString() ?? banner['title']?.toString() ?? '';
    }
    
    final targetUrl = banner['target_url']?.toString() ?? '';
    
    // Vérifier si c'est une vidéo
    final isVideo = imageUrl.toLowerCase().endsWith('.mp4') || 
                    imageUrl.toLowerCase().endsWith('.mov') ||
                    imageUrl.toLowerCase().endsWith('.avi') ||
                    imageUrl.toLowerCase().endsWith('.webm');
    
    return GestureDetector(
      onTap: () {
        if (targetUrl.isNotEmpty) {
          // Ouvrir le lien cible si disponible
          debugPrint('Navigate to: $targetUrl');
        }
      },
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Stack(
          children: [
            // Image ou vidéo
            if (isVideo && imageUrl.isNotEmpty)
              Container(
                color: Colors.grey[300],
                child: const Center(
                  child: Icon(Icons.play_circle_outline, size: 64, color: Colors.grey),
                ),
              )
            else if (imageUrl.startsWith('http'))
              CachedNetworkImage(
                imageUrl: imageUrl,
                width: double.infinity,
                fit: BoxFit.contain,
                placeholder: (context, url) => Shimmer.fromColors(
                  baseColor: Colors.grey[300]!,
                  highlightColor: Colors.grey[100]!,
                  child: Container(color: Colors.white),
                ),
                errorWidget: (context, url, error) {
                  return Container(
                    color: Colors.grey[300],
                    child: const Center(
                      child: Icon(Icons.broken_image, size: 48, color: Colors.grey),
                    ),
                  );
                },
              )
            else if (imageUrl.startsWith('assets/'))
              Image.asset(
                imageUrl,
                width: double.infinity,
                fit: BoxFit.contain,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    color: Colors.grey[300],
                    child: const Center(
                      child: Icon(Icons.broken_image, size: 48, color: Colors.grey),
                    ),
                  );
                },
              )
            else
              Container(
                color: Colors.grey[300],
                child: const Center(
                  child: Icon(Icons.broken_image, size: 48, color: Colors.grey),
                ),
              ),
            
            // Titre de la bannière (optionnel)
            if (title.isNotEmpty)
              Positioned(
                bottom: 0,
                left: 0,
                right: 0,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      colors: [
                        Colors.transparent,
                        Colors.black.withOpacity(0.7),
                      ],
                    ),
                  ),
                  child: Text(
                    title,
                    style: const TextStyle(
                      fontFamily: 'Plus Jakarta Sans',
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ),
          ],
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
            Icons.search_off_outlined,
            size: 64,
            color: isDarkMode ? Colors.grey[600] : Colors.grey[400],
          ),
          const SizedBox(height: 16),
          Text(
            AppLocalizations.of(context)?.no_auctions_found ?? 'No auctions found in this city',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white70 : Colors.black54,
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

  Widget _buildAuctionCard(BuildContext context, Auction auction, bool isDarkMode, int displayNumber) {
    final favoritesAsync = ref.watch(favoritesProvider);
    final isFavorite = favoritesAsync.value?.contains(auction.id) ?? false;
    final id = auction.id;
    final title = auction.title;
    final price = '${auction.currentPrice.toStringAsFixed(0)} MRU';
    final bidCount = auction.bidderCount;
    final time = TimeUtils.formatDuration(context, auction.endTime.difference(DateTime.now()));
    final imageUrl = auction.imageUrls.isNotEmpty ? auction.imageUrls[0] : 'assets/corolla.png';
    final isNetworkImage = imageUrl.startsWith('http');



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

                    ? CachedNetworkImage(

                        imageUrl: imageUrl,

                        height: 110,

                        width: double.infinity,

                        fit: BoxFit.cover,

                        placeholder: (context, url) => Shimmer.fromColors(

                          baseColor: Colors.grey[300]!,

                          highlightColor: Colors.grey[100]!,

                          child: Container(color: Colors.white),

                        ),

                        errorWidget: (c, e, s) => Image.asset(

                          'assets/corolla.png',

                          height: 110,

                          width: double.infinity,

                          fit: BoxFit.cover,

                        ),

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