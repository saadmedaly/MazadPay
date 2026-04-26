import 'package:hive_flutter/hive_flutter.dart';
import 'dart:convert';

class CacheService {
  static const String _citiesBoxName = 'cities_cache';
  static const String _auctionsBoxName = 'auctions_cache';
  static const String _auctionDetailsBoxName = 'auction_details_cache';
  static const String _categoriesBoxName = 'categories_cache';
  static const String _countriesBoxName = 'countries_cache';
  static const String _subCategoriesBoxName = 'sub_categories_cache';
  static const String _reportReasonsBoxName = 'report_reasons_cache';
  static const String _myAuctionsBoxName = 'my_auctions_cache';
  static const String _myWinningsBoxName = 'my_winnings_cache';
  
  static const String _citiesKey = 'cities_data';
  static const String _auctionsKey = 'auctions_data';
  static const String _categoriesKey = 'categories_data';
  static const String _countriesKey = 'countries_data';
  static const String _reportReasonsKey = 'report_reasons_data';
  static const String _myAuctionsKey = 'my_auctions_data';
  static const String _myWinningsKey = 'my_winnings_data';
  
  static const String _citiesTimestampKey = 'cities_timestamp';
  static const String _auctionsTimestampKey = 'auctions_timestamp';
  static const String _categoriesTimestampKey = 'categories_timestamp';
  static const String _countriesTimestampKey = 'countries_timestamp';
  static const String _reportReasonsTimestampKey = 'report_reasons_timestamp';
  static const String _myAuctionsTimestampKey = 'my_auctions_timestamp';
  static const String _myWinningsTimestampKey = 'my_winnings_timestamp';
  
  // TTL en millisecondes
  static const int _citiesCacheTTL = 3600000; // 1 heure pour les villes (changent rarement)
  static const int _auctionsCacheTTL = 300000; // 5 minutes pour les enchères (changent plus souvent)
  static const int _auctionDetailsCacheTTL = 60000; // 1 minute pour les détails d'enchère (changent très souvent)
  static const int _categoriesCacheTTL = 7200000; // 2 heures pour les catégories (changent très rarement)
  static const int _countriesCacheTTL = 86400000; // 24 heures pour les pays (changent rarement)
  static const int _subCategoriesCacheTTL = 7200000; // 2 heures pour les sous-catégories
  static const int _reportReasonsCacheTTL = 86400000; // 24 heures pour les raisons de signalement
  static const int _myAuctionsCacheTTL = 300000; // 5 minutes pour mes enchères
  static const int _myWinningsCacheTTL = 300000; // 5 minutes pour mes gains
  
  static CacheService? _instance;
  Box? _citiesBox;
  Box? _auctionsBox;
  Box? _auctionDetailsBox;
  Box? _categoriesBox;
  Box? _countriesBox;
  Box? _subCategoriesBox;
  Box? _reportReasonsBox;
  Box? _myAuctionsBox;
  Box? _myWinningsBox;
  
  bool _isInitialized = false;
  
  static CacheService get instance {
    _instance ??= CacheService._();
    return _instance!;
  }
  
  CacheService._();
  
  Future<void> init() async {
    if (_isInitialized) return;
    
    await Hive.initFlutter();
    _citiesBox = await Hive.openBox(_citiesBoxName);
    _auctionsBox = await Hive.openBox(_auctionsBoxName);
    _auctionDetailsBox = await Hive.openBox(_auctionDetailsBoxName);
    _categoriesBox = await Hive.openBox(_categoriesBoxName);
    _countriesBox = await Hive.openBox(_countriesBoxName);
    _subCategoriesBox = await Hive.openBox(_subCategoriesBoxName);
    _reportReasonsBox = await Hive.openBox(_reportReasonsBoxName);
    _myAuctionsBox = await Hive.openBox(_myAuctionsBoxName);
    _myWinningsBox = await Hive.openBox(_myWinningsBoxName);
    
    _isInitialized = true;
  }
  
  Future<void> _ensureInitialized() async {
    if (!_isInitialized) {
      await init();
    }
  }
  
  // Cache pour les villes
  Future<void> cacheCities(List<Map<String, dynamic>> cities) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _citiesBox!.put(_citiesKey, jsonEncode(cities));
    await _citiesBox!.put(_citiesTimestampKey, timestamp);
  }
  
  Future<List<Map<String, dynamic>>?> getCachedCities() async {
    await _ensureInitialized();
    final cachedData = _citiesBox!.get(_citiesKey);
    final timestamp = _citiesBox!.get(_citiesTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _citiesCacheTTL) {
      return null;
    }
    
    try {
      final decoded = jsonDecode(cachedData) as List;
      return decoded.map((e) => e as Map<String, dynamic>).toList();
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isCitiesCacheValid() async {
    await _ensureInitialized();
    final timestamp = _citiesBox!.get(_citiesTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _citiesCacheTTL;
  }
  
  // Cache pour les enchères
  Future<void> cacheAuctions(List<Map<String, dynamic>> auctions) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _auctionsBox!.put(_auctionsKey, jsonEncode(auctions));
    await _auctionsBox!.put(_auctionsTimestampKey, timestamp);
  }
  
  Future<List<Map<String, dynamic>>?> getCachedAuctions() async {
    await _ensureInitialized();
    final cachedData = _auctionsBox!.get(_auctionsKey);
    final timestamp = _auctionsBox!.get(_auctionsTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _auctionsCacheTTL) {
      return null;
    }
    
    try {
      final decoded = jsonDecode(cachedData) as List;
      return decoded.map((e) => e as Map<String, dynamic>).toList();
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isAuctionsCacheValid() async {
    await _ensureInitialized();
    final timestamp = _auctionsBox!.get(_auctionsTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _auctionsCacheTTL;
  }
  
  // Cache pour les détails d'une enchère
  Future<void> cacheAuctionDetail(String auctionId, Map<String, dynamic> auctionDetail) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _auctionDetailsBox!.put(auctionId, jsonEncode(auctionDetail));
    await _auctionDetailsBox!.put('${auctionId}_timestamp', timestamp);
  }
  
  Future<Map<String, dynamic>?> getCachedAuctionDetail(String auctionId) async {
    await _ensureInitialized();
    final cachedData = _auctionDetailsBox!.get(auctionId);
    final timestamp = _auctionDetailsBox!.get('${auctionId}_timestamp');
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _auctionDetailsCacheTTL) {
      return null;
    }
    
    try {
      return jsonDecode(cachedData) as Map<String, dynamic>;
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isAuctionDetailCacheValid(String auctionId) async {
    await _ensureInitialized();
    final timestamp = _auctionDetailsBox!.get('${auctionId}_timestamp');
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _auctionDetailsCacheTTL;
  }
  
  // Cache pour les catégories
  Future<void> cacheCategories(List<Map<String, dynamic>> categories) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _categoriesBox!.put(_categoriesKey, jsonEncode(categories));
    await _categoriesBox!.put(_categoriesTimestampKey, timestamp);
  }
  
  Future<List<Map<String, dynamic>>?> getCachedCategories() async {
    await _ensureInitialized();
    final cachedData = _categoriesBox!.get(_categoriesKey);
    final timestamp = _categoriesBox!.get(_categoriesTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _categoriesCacheTTL) {
      return null;
    }
    
    try {
      final decoded = jsonDecode(cachedData) as List;
      return decoded.map((e) => e as Map<String, dynamic>).toList();
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isCategoriesCacheValid() async {
    await _ensureInitialized();
    final timestamp = _categoriesBox!.get(_categoriesTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _categoriesCacheTTL;
  }
  
  // Cache pour les pays
  Future<void> cacheCountries(List<Map<String, dynamic>> countries) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _countriesBox!.put(_countriesKey, jsonEncode(countries));
    await _countriesBox!.put(_countriesTimestampKey, timestamp);
  }
  
  Future<List<Map<String, dynamic>>?> getCachedCountries() async {
    await _ensureInitialized();
    final cachedData = _countriesBox!.get(_countriesKey);
    final timestamp = _countriesBox!.get(_countriesTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _countriesCacheTTL) {
      return null;
    }
    
    try {
      final decoded = jsonDecode(cachedData) as List;
      return decoded.map((e) => e as Map<String, dynamic>).toList();
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isCountriesCacheValid() async {
    await _ensureInitialized();
    final timestamp = _countriesBox!.get(_countriesTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _countriesCacheTTL;
  }
  
  // Cache pour les sous-catégories
  Future<void> cacheSubCategories(String categoryId, List<Map<String, dynamic>> subCategories) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _subCategoriesBox!.put(categoryId, jsonEncode(subCategories));
    await _subCategoriesBox!.put('${categoryId}_timestamp', timestamp);
  }
  
  Future<List<Map<String, dynamic>>?> getCachedSubCategories(String categoryId) async {
    await _ensureInitialized();
    final cachedData = _subCategoriesBox!.get(categoryId);
    final timestamp = _subCategoriesBox!.get('${categoryId}_timestamp');
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _subCategoriesCacheTTL) {
      return null;
    }
    
    try {
      final decoded = jsonDecode(cachedData) as List;
      return decoded.map((e) => e as Map<String, dynamic>).toList();
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isSubCategoriesCacheValid(String categoryId) async {
    await _ensureInitialized();
    final timestamp = _subCategoriesBox!.get('${categoryId}_timestamp');
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _subCategoriesCacheTTL;
  }
  
  // Cache pour les raisons de signalement
  Future<void> cacheReportReasons(List<Map<String, dynamic>> reportReasons) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _reportReasonsBox!.put(_reportReasonsKey, jsonEncode(reportReasons));
    await _reportReasonsBox!.put(_reportReasonsTimestampKey, timestamp);
  }
  
  Future<List<Map<String, dynamic>>?> getCachedReportReasons() async {
    await _ensureInitialized();
    final cachedData = _reportReasonsBox!.get(_reportReasonsKey);
    final timestamp = _reportReasonsBox!.get(_reportReasonsTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _reportReasonsCacheTTL) {
      return null;
    }
    
    try {
      final decoded = jsonDecode(cachedData) as List;
      return decoded.map((e) => e as Map<String, dynamic>).toList();
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isReportReasonsCacheValid() async {
    await _ensureInitialized();
    final timestamp = _reportReasonsBox!.get(_reportReasonsTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _reportReasonsCacheTTL;
  }
  
  // Cache pour mes enchères
  Future<void> cacheMyAuctions(Map<String, dynamic> myAuctions) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _myAuctionsBox!.put(_myAuctionsKey, jsonEncode(myAuctions));
    await _myAuctionsBox!.put(_myAuctionsTimestampKey, timestamp);
  }
  
  Future<Map<String, dynamic>?> getCachedMyAuctions() async {
    await _ensureInitialized();
    final cachedData = _myAuctionsBox!.get(_myAuctionsKey);
    final timestamp = _myAuctionsBox!.get(_myAuctionsTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _myAuctionsCacheTTL) {
      return null;
    }
    
    try {
      return jsonDecode(cachedData) as Map<String, dynamic>;
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isMyAuctionsCacheValid() async {
    await _ensureInitialized();
    final timestamp = _myAuctionsBox!.get(_myAuctionsTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _myAuctionsCacheTTL;
  }
  
  // Cache pour mes gains
  Future<void> cacheMyWinnings(Map<String, dynamic> myWinnings) async {
    await _ensureInitialized();
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    await _myWinningsBox!.put(_myWinningsKey, jsonEncode(myWinnings));
    await _myWinningsBox!.put(_myWinningsTimestampKey, timestamp);
  }
  
  Future<Map<String, dynamic>?> getCachedMyWinnings() async {
    await _ensureInitialized();
    final cachedData = _myWinningsBox!.get(_myWinningsKey);
    final timestamp = _myWinningsBox!.get(_myWinningsTimestampKey);
    
    if (cachedData == null || timestamp == null) {
      return null;
    }
    
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - timestamp > _myWinningsCacheTTL) {
      return null;
    }
    
    try {
      return jsonDecode(cachedData) as Map<String, dynamic>;
    } catch (e) {
      return null;
    }
  }
  
  Future<bool> isMyWinningsCacheValid() async {
    await _ensureInitialized();
    final timestamp = _myWinningsBox!.get(_myWinningsTimestampKey);
    if (timestamp == null) return false;
    
    final now = DateTime.now().millisecondsSinceEpoch;
    return now - timestamp <= _myWinningsCacheTTL;
  }
  
  // Clear cache
  Future<void> clearCitiesCache() async {
    await _ensureInitialized();
    await _citiesBox!.delete(_citiesKey);
    await _citiesBox!.delete(_citiesTimestampKey);
  }
  
  Future<void> clearAuctionsCache() async {
    await _ensureInitialized();
    await _auctionsBox!.delete(_auctionsKey);
    await _auctionsBox!.delete(_auctionsTimestampKey);
  }
  
  Future<void> clearAuctionDetailCache(String auctionId) async {
    await _ensureInitialized();
    await _auctionDetailsBox!.delete(auctionId);
    await _auctionDetailsBox!.delete('${auctionId}_timestamp');
  }
  
  Future<void> clearCategoriesCache() async {
    await _ensureInitialized();
    await _categoriesBox!.delete(_categoriesKey);
    await _categoriesBox!.delete(_categoriesTimestampKey);
  }
  
  Future<void> clearCountriesCache() async {
    await _ensureInitialized();
    await _countriesBox!.delete(_countriesKey);
    await _countriesBox!.delete(_countriesTimestampKey);
  }
  
  Future<void> clearAllCache() async {
    await _ensureInitialized();
    await _citiesBox!.clear();
    await _auctionsBox!.clear();
    await _auctionDetailsBox!.clear();
    await _categoriesBox!.clear();
    await _countriesBox!.clear();
    await _subCategoriesBox!.clear();
    await _reportReasonsBox!.clear();
    await _myAuctionsBox!.clear();
    await _myWinningsBox!.clear();
  }
}
