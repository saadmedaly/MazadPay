import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/favorites_provider.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/pages/create_ad_start_page.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/pages/home_page.dart';

class AllAuctionsPage extends ConsumerStatefulWidget {
  const AllAuctionsPage({super.key});

  @override
  ConsumerState<AllAuctionsPage> createState() => _AllAuctionsPageState();
}

class _AllAuctionsPageState extends ConsumerState<AllAuctionsPage> {
  final TextEditingController _searchController = TextEditingController();

  List<Map<String, String>> _allAuctions = [];
  List<Map<String, String>> _filteredAuctions = [];
  int _activeTabIndex = 0;
  int _selectedCategoryIndex = 0;
  int _selectedSubCategoryIndex = 0;


  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final l10n = AppLocalizations.of(context)!;
    _allAuctions = [
      {
        'id': '1',
        'title': 'Toyota prado tx 2017',
        'price': '15,000 MRU',
        'bids': '15',
        'time': '23h 30m 18s',
        'location': l10n.text_113,
        'category': 'cars',
        'subCategory': '4x4',
        'image': 'assets/car0.png',
        'status': 'active',
        'postedTime': 'منذ 5 ساعات'
      },
      {
        'id': '2',
        'title': 'هلكس خليجية نظيفة',
        'price': '760,000 MRU',
        'bids': '7',
        'time': '23h 30m 18s',
        'location': l10n.text_104,
        'category': 'cars',
        'subCategory': 'standard',
        'image': 'assets/car1.jpg',
        'status': 'active',
        'postedTime': 'منذ 2 ساعات'
      },
      {
        'id': '5',
        'title': 'سيارة تاكسي بحالة ممتازة',
        'price': '450,000 MRU',
        'bids': '3',
        'time': '10h 15m 30s',
        'location': l10n.text_111,
        'category': 'cars',
        'subCategory': 'taxi',
        'image': 'assets/car2.jpg',
        'status': 'active',
        'postedTime': 'منذ ساعة'
      },
      {
        'id': '3',
        'title': 'TOYOTA RAV4 2008 D4D',
        'price': '420,000 MRU',
        'bids': '20',
        'time': l10n.text_364,
        'location': l10n.text_113,
        'category': 'cars',
        'subCategory': '4x4',
        'image': 'assets/car3.jpg',
        'status': 'finished',
        'postedTime': 'منذ يوم'
      },
      {
        'id': '4',
        'title': 'iPhone 15 Pro Max',
        'price': '55,000 MRU',
        'bids': '12',
        'time': '12h 10m 05s',
        'location': l10n.text_113,
        'category': 'phones',
        'image': 'assets/phone1.jpg',
        'status': 'active',
        'postedTime': 'منذ 3 ساعات'
      },
      {
        'id': '6',
        'title': 'Samsung Galaxy S24 Ultra',
        'price': '38,000 MRU',
        'bids': '8',
        'time': '05h 45m 20s',
        'location': l10n.text_113,
        'category': 'phones',
        'image': 'assets/phone2.jpg',
        'status': 'active',
        'postedTime': 'منذ 6 ساعات'
      },
      {
        'id': '7',
        'title': 'فيلا راقية وسط العاصمة',
        'price': '2,500,000 MRU',
        'bids': '5',
        'time': '48h 00m 00s',
        'location': l10n.text_113,
        'category': 'real_estate',
        'image': 'assets/maison1.jpg',
        'status': 'active',
        'postedTime': 'منذ يومين'
      },
      {
        'id': '8',
        'title': 'شقة 3 غرف للبيع',
        'price': '850,000 MRU',
        'bids': '11',
        'time': '24h 00m 00s',
        'location': l10n.text_113,
        'category': 'real_estate',
        'image': 'assets/maison2.jpg',
        'status': 'active',
        'postedTime': 'منذ يوم'
      },
      {
        'id': '9',
        'title': 'غسالة أوتوماتيك 7 كيلو',
        'price': '25,000 MRU',
        'bids': '4',
        'time': '08h 30m 00s',
        'location': l10n.text_113,
        'category': 'home_appliances',
        'image': 'assets/Appareils de maison1.jpg',
        'status': 'active',
        'postedTime': 'منذ 4 ساعات'
      },
      {
        'id': '10',
        'title': 'ثلاجة سامسونج 500 لتر',
        'price': '35,000 MRU',
        'bids': '6',
        'time': '15h 00m 00s',
        'location': l10n.text_113,
        'category': 'home_appliances',
        'image': 'assets/Appareils de maison2.jpg',
        'status': 'active',
        'postedTime': 'منذ 7 ساعات'
      },
      // ── أجهزة منزلية (باقي الصور) ───────────────────────
      {'id': 'ha3', 'title': 'مكيف سبليت 2 طن',           'price': '42,000 MRU',  'bids': '5',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'home_appliances', 'image': 'assets/Appareils de maison3.jpg', 'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'ha4', 'title': 'بوتوغاز مع أسطوانة',        'price': '12,000 MRU',  'bids': '2',  'time': '06h 00m 00s', 'location': l10n.text_113, 'category': 'home_appliances', 'image': 'assets/Appareils de maison4.jpg', 'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'ha5', 'title': 'تلفزيون سامسونج 65 بوصة',   'price': '55,000 MRU',  'bids': '7',  'time': '14h 00m 00s', 'location': l10n.text_113, 'category': 'home_appliances', 'image': 'assets/Appareils de maison5.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'ha6', 'title': 'مجفف ملابس LG',              'price': '28,000 MRU',  'bids': '3',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'home_appliances', 'image': 'assets/Appareils de maison6.jpg', 'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 'ha7', 'title': 'طباخ كهربائي 6 شعلات',      'price': '18,000 MRU',  'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'home_appliances', 'image': 'assets/Appareils de maison7.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'ha8', 'title': 'مكنسة كهربائية دايسون',      'price': '22,000 MRU',  'bids': '4',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'home_appliances', 'image': 'assets/Appareils de maison8.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      // ── الإلكترونيات ─────────────────────────────────────
      {'id': 'el1', 'title': 'لابتوب ماك بوك برو M3',      'price': '120,000 MRU', 'bids': '9',  'time': '18h 00m 00s', 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique1.jpg', 'status': 'active',   'postedTime': 'منذ ساعة'},
      {'id': 'el2', 'title': 'تابلت آيباد برو 12.9',       'price': '75,000 MRU',  'bids': '6',  'time': '14h 00m 00s', 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique2.jpg', 'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'el3', 'title': 'كاميرا سوني A7 IV',          'price': '90,000 MRU',  'bids': '5',  'time': '12h 00m 00s', 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique3.jpg', 'status': 'active',   'postedTime': 'منذ 4 ساعات'},
      {'id': 'el4', 'title': 'بلاي ستيشن 5 جديدة',         'price': '45,000 MRU',  'bids': '11', 'time': '20h 00m 00s', 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique4.jpg', 'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'el5', 'title': 'سماعات سوني WH-1000XM5',     'price': '18,000 MRU',  'bids': '4',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique5.jpg', 'status': 'active',   'postedTime': 'منذ 6 ساعات'},
      {'id': 'el6', 'title': 'شاشة LG UltraWide 34',       'price': '52,000 MRU',  'bids': '3',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique6.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'el7', 'title': 'طابعة ليزر HP LaserJet',     'price': '28,000 MRU',  'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique7.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'el8', 'title': 'راوتر واي فاي 6 Asus',       'price': '12,000 MRU',  'bids': '2',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'electronics', 'image': 'assets/Électronique8.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      // ── الهواتف (باقي الصور) ─────────────────────────────
      {'id': 'ph3', 'title': 'Huawei P60 Pro',             'price': '42,000 MRU',  'bids': '5',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'phones', 'image': 'assets/phone3.jpg', 'status': 'active',   'postedTime': 'منذ 4 ساعات'},
      {'id': 'ph4', 'title': 'Xiaomi 14 Ultra',            'price': '35,000 MRU',  'bids': '3',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'phones', 'image': 'assets/phone4.jpg', 'status': 'active',   'postedTime': 'منذ 6 ساعات'},
      {'id': 'ph5', 'title': 'Google Pixel 8 Pro',         'price': '48,000 MRU',  'bids': '7',  'time': '16h 00m 00s', 'location': l10n.text_113, 'category': 'phones', 'image': 'assets/phone5.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'ph6', 'title': 'OnePlus 12 256GB',           'price': '28,000 MRU',  'bids': '2',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'phones', 'image': 'assets/phone6.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      // ── العقارات (باقي الصور) ────────────────────────────
      {'id': 're3', 'title': 'منزل 5 غرف بحديقة',          'price': '1,800,000 MRU','bids': '8', 'time': '72h 00m 00s', 'location': l10n.text_113, 'category': 'real_estate', 'image': 'assets/maison3.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 're4', 'title': 'شقة مفروشة للإيجار',         'price': '120,000 MRU', 'bids': '4',  'time': '24h 00m 00s', 'location': l10n.text_113, 'category': 'real_estate', 'image': 'assets/maison4.jpg', 'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 're5', 'title': 'مبنى تجاري وسط المدينة',     'price': '4,500,000 MRU','bids': '12','time': '96h 00m 00s', 'location': l10n.text_113, 'category': 'real_estate', 'image': 'assets/maison5.jpg', 'status': 'active',   'postedTime': 'منذ 3 أيام'},
      // ── الأثاث ──────────────────────────────────────────
      {'id': 'f1',  'title': 'طقم صالون فاخر',         'price': '45,000 MRU',  'bids': '3',  'time': '12h 00m 00s', 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles1.jpg',  'status': 'active',   'postedTime': 'منذ ساعتين'},
      {'id': 'f2',  'title': 'غرفة نوم كاملة',          'price': '60,000 MRU',  'bids': '5',  'time': '18h 00m 00s', 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles2.jpg',  'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'f3',  'title': 'طاولة سفرة خشب طبيعي',   'price': '22,000 MRU',  'bids': '2',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles3.jpg',  'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'f4',  'title': 'مكتبة خشبية كبيرة',       'price': '18,000 MRU',  'bids': '1',  'time': '06h 00m 00s', 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles4.jpg',  'status': 'active',   'postedTime': 'منذ 6 ساعات'},
      {'id': 'f5',  'title': 'كنبة جلد إيطالي',         'price': '35,000 MRU',  'bids': '4',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles5.jpg',  'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'f6',  'title': 'طاولة قهوة زجاج',         'price': '8,500 MRU',   'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles6.jpg',  'status': 'finished', 'postedTime': 'منذ يومين'},
      {'id': 'f7',  'title': 'خزانة ملابس 6 أبواب',     'price': '27,000 MRU',  'bids': '2',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles7.jpg',  'status': 'finished', 'postedTime': 'منذ 3 أيام'},
      {'id': 'f8',  'title': 'سرير مع دولاب',           'price': '32,000 MRU',  'bids': '3',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'furniture', 'image': 'assets/Meubles8.jpg',  'status': 'finished', 'postedTime': 'منذ أسبوع'},
      // ── قطع أرضية ───────────────────────────────────────
      {'id': 't1',  'title': 'قطعة أرضية حي تفاريغ',    'price': '500,000 MRU', 'bids': '7',  'time': '48h 00m 00s', 'location': l10n.text_113, 'category': 'land_plots', 'image': 'assets/Terrains1.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 't2',  'title': 'أرض سكنية 500م²',          'price': '350,000 MRU', 'bids': '4',  'time': '24h 00m 00s', 'location': l10n.text_113, 'category': 'land_plots', 'image': 'assets/Terrains2.jpg', 'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 't3',  'title': 'قطعة تجارية وسط المدينة', 'price': '900,000 MRU', 'bids': '10', 'time': '72h 00m 00s', 'location': l10n.text_113, 'category': 'land_plots', 'image': 'assets/Terrains3.jpg', 'status': 'active',   'postedTime': 'منذ 4 ساعات'},
      {'id': 't4',  'title': 'أرض زراعية مسورة',         'price': '200,000 MRU', 'bids': '2',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'land_plots', 'image': 'assets/Terrains4.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      // ── معدات ثقيلة ─────────────────────────────────────
      {'id': 'ml1', 'title': 'حفارة كاتربيلار 2019',    'price': '1,200,000 MRU','bids': '8', 'time': '36h 00m 00s', 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd1.jpg',  'status': 'active',   'postedTime': 'منذ ساعة'},
      {'id': 'ml2', 'title': 'جرافة كوماتسو',            'price': '950,000 MRU', 'bids': '5',  'time': '20h 00m 00s', 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd2.jpg',  'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'ml3', 'title': 'رافعة شوكية 5 طن',         'price': '400,000 MRU', 'bids': '3',  'time': '14h 00m 00s', 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd3.jpg',  'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'ml4', 'title': 'خلاطة خرسانة كبيرة',      'price': '180,000 MRU', 'bids': '2',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd4.jpg',  'status': 'active',   'postedTime': 'منذ 6 ساعات'},
      {'id': 'ml5', 'title': 'ضاغط هواء صناعي',          'price': '95,000 MRU',  'bids': '1',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd5.jpg',  'status': 'active',   'postedTime': 'منذ 8 ساعات'},
      {'id': 'ml6', 'title': 'شاحنة قلاب هيونداي',       'price': '620,000 MRU', 'bids': '6',  'time': '16h 00m 00s', 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd6.jpg',  'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'ml7', 'title': 'مضخة مياه صناعية',         'price': '75,000 MRU',  'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd7.jpg',  'status': 'finished', 'postedTime': 'منذ يومين'},
      {'id': 'ml8', 'title': 'مولد كهرباء 100 كيلوات',  'price': '300,000 MRU', 'bids': '4',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd8.jpg',  'status': 'finished', 'postedTime': 'منذ 3 أيام'},
      {'id': 'ml9', 'title': 'آلة تسوية أرض',            'price': '550,000 MRU', 'bids': '3',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd9.jpg',  'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'ml10','title': 'كرين متحرك 20 طن',         'price': '2,000,000 MRU','bids': '9', 'time': l10n.text_364, 'location': l10n.text_113, 'category': 'heavy_materials', 'image': 'assets/Matériel lourd10.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      // ── الدراجات ────────────────────────────────────────
      {'id': 'mo1', 'title': 'دراجة ياماها R15 2022',    'price': '85,000 MRU',  'bids': '6',  'time': '14h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos1.jpg',  'status': 'active',   'postedTime': 'منذ ساعتين'},
      {'id': 'mo2', 'title': 'دراجة هوندا CB500',         'price': '72,000 MRU',  'bids': '4',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos2.jpg',  'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'mo3', 'title': 'دراجة سوزوكي GSX',          'price': '65,000 MRU',  'bids': '3',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos3.jpg',  'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'mo4', 'title': 'دراجة كاوازاكي نينجا',      'price': '110,000 MRU', 'bids': '8',  'time': '22h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos4.jpg',  'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'mo5', 'title': 'دراجة BMW G310R',           'price': '130,000 MRU', 'bids': '5',  'time': '18h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos5.jpg',  'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'mo6', 'title': 'دراجة كلاسيكية هارلي',      'price': '200,000 MRU', 'bids': '10', 'time': '30h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos6.jpg',  'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 'mo7', 'title': 'دراجة دوكاتي مونستر',       'price': '160,000 MRU', 'bids': '7',  'time': '16h 00m 00s', 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos7.jpg',  'status': 'active',   'postedTime': 'منذ 3 أيام'},
      {'id': 'mo8', 'title': 'دراجة ترياومف 2020',        'price': '145,000 MRU', 'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos8.jpg',  'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'mo9', 'title': 'دراجة أبريليا RS 660',      'price': '175,000 MRU', 'bids': '2',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos9.jpg',  'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'mo10','title': 'دراجة KTM Duke 390',        'price': '95,000 MRU',  'bids': '4',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos10.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      {'id': 'mo11','title': 'دراجة هيوسانج GT650',       'price': '88,000 MRU',  'bids': '1',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos11.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      {'id': 'mo12','title': 'دراجة موتو غوتسي V7',       'price': '120,000 MRU', 'bids': '3',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'bikes', 'image': 'assets/Motos12.jpg', 'status': 'finished', 'postedTime': 'منذ شهر'},
      // ── الشاحنات ────────────────────────────────────────
      {'id': 'c1',  'title': 'شاحنة مرسيدس أكتروس',      'price': '1,500,000 MRU','bids': '9', 'time': '40h 00m 00s', 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions1.jpg', 'status': 'active',   'postedTime': 'منذ ساعة'},
      {'id': 'c2',  'title': 'شاحنة مان TGX 480',         'price': '1,200,000 MRU','bids': '6', 'time': '28h 00m 00s', 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions2.jpg', 'status': 'active',   'postedTime': 'منذ 4 ساعات'},
      {'id': 'c3',  'title': 'شاحنة فولفو FH16',          'price': '1,800,000 MRU','bids': '11','time': '48h 00m 00s', 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions3.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'c4',  'title': 'شاحنة سكانيا R500',         'price': '1,600,000 MRU','bids': '7', 'time': '32h 00m 00s', 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions4.jpg', 'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 'c5',  'title': 'شاحنة رينو T520',           'price': '900,000 MRU', 'bids': '4',  'time': '20h 00m 00s', 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions5.jpg', 'status': 'active',   'postedTime': 'منذ 3 أيام'},
      {'id': 'c6',  'title': 'شاحنة إيفيكو ستيليس',       'price': '750,000 MRU', 'bids': '3',  'time': '16h 00m 00s', 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions6.jpg', 'status': 'active',   'postedTime': 'منذ 3 أيام'},
      {'id': 'c7',  'title': 'شاحنة دايملر أكتروس 1844', 'price': '2,000,000 MRU','bids': '0', 'time': l10n.text_364, 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions7.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'c8',  'title': 'شاحنة نيسان كوندور',        'price': '680,000 MRU', 'bids': '5',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'trucks', 'image': 'assets/Camions8.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      // ── مستلزمات رجالية ─────────────────────────────────
      {'id': 'mh1', 'title': 'ساعة رولكس أصلية',          'price': '250,000 MRU', 'bids': '15', 'time': '24h 00m 00s', 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes1.jpg',  'status': 'active',   'postedTime': 'منذ ساعة'},
      {'id': 'mh2', 'title': 'حقيبة جلد رجالية فاخرة',   'price': '35,000 MRU',  'bids': '4',  'time': '12h 00m 00s', 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes2.jpg',  'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'mh3', 'title': 'بدلة رسمية ماركة',          'price': '28,000 MRU',  'bids': '2',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes3.jpg',  'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'mh4', 'title': 'حذاء جلد إيطالي',           'price': '18,000 MRU',  'bids': '3',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes4.jpg',  'status': 'active',   'postedTime': 'منذ 6 ساعات'},
      {'id': 'mh5', 'title': 'نظارة شمسية راي بان',        'price': '12,000 MRU',  'bids': '1',  'time': '06h 00m 00s', 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes5.jpg',  'status': 'active',   'postedTime': 'منذ 8 ساعات'},
      {'id': 'mh6', 'title': 'عطر رجالي كريد',            'price': '22,000 MRU',  'bids': '5',  'time': '14h 00m 00s', 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes6.jpg',  'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'mh7', 'title': 'حزام جلد ماركة هيرمس',      'price': '15,000 MRU',  'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes7.jpg',  'status': 'finished', 'postedTime': 'منذ يومين'},
      {'id': 'mh8', 'title': 'قميص رسمي بيير كاردان',     'price': '9,500 MRU',   'bids': '2',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes8.jpg',  'status': 'finished', 'postedTime': 'منذ 3 أيام'},
      {'id': 'mh9', 'title': 'خاتم ذهب رجالي',            'price': '45,000 MRU',  'bids': '6',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes9.jpg',  'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'mh10','title': 'محفظة جلد فاخرة',           'price': '8,000 MRU',   'bids': '1',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'mens', 'image': 'assets/Fournitures pour hommes10.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      // ── مستلزمات نسائية ─────────────────────────────────
      {'id': 'wm1', 'title': 'حقيبة شانيل أصلية',         'price': '180,000 MRU', 'bids': '12', 'time': '20h 00m 00s', 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes1.jpg', 'status': 'active',   'postedTime': 'منذ ساعة'},
      {'id': 'wm2', 'title': 'عطر شنل نمبر 5',            'price': '30,000 MRU',  'bids': '5',  'time': '14h 00m 00s', 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes2.jpg', 'status': 'active',   'postedTime': 'منذ 3 ساعات'},
      {'id': 'wm3', 'title': 'ساعة فندي نسائية',           'price': '95,000 MRU',  'bids': '8',  'time': '18h 00m 00s', 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes3.jpg', 'status': 'active',   'postedTime': 'منذ 5 ساعات'},
      {'id': 'wm4', 'title': 'مجوهرات ذهب 18 قيراط',      'price': '120,000 MRU', 'bids': '10', 'time': '24h 00m 00s', 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes4.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'wm5', 'title': 'فستان سواريه فاخر',          'price': '25,000 MRU',  'bids': '3',  'time': '10h 00m 00s', 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes5.jpg', 'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 'wm6', 'title': 'حذاء كعب عالٍ لويس فيتون', 'price': '42,000 MRU',  'bids': '4',  'time': '12h 00m 00s', 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes6.jpg', 'status': 'active',   'postedTime': 'منذ 3 أيام'},
      {'id': 'wm7', 'title': 'نظارة شمسية غوتشي نسائية',  'price': '16,000 MRU',  'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes7.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'wm8', 'title': 'خاتم ألماس نسائي',           'price': '200,000 MRU', 'bids': '14', 'time': l10n.text_364, 'location': l10n.text_113, 'category': 'womens', 'image': 'assets/Fournitures pour femmes8.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
      // ── حيوانات ──────────────────────────────────────────
      {'id': 'an1', 'title': 'حصان عربي أصيل',            'price': '300,000 MRU', 'bids': '9',  'time': '36h 00m 00s', 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal1.jpg', 'status': 'active',   'postedTime': 'منذ ساعة'},
      {'id': 'an2', 'title': 'جمل مهجن للبيع',             'price': '150,000 MRU', 'bids': '5',  'time': '24h 00m 00s', 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal2.jpg', 'status': 'active',   'postedTime': 'منذ 4 ساعات'},
      {'id': 'an3', 'title': 'ماعز محلية للبيع',           'price': '25,000 MRU',  'bids': '3',  'time': '12h 00m 00s', 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal3.jpg', 'status': 'active',   'postedTime': 'منذ 6 ساعات'},
      {'id': 'an4', 'title': 'خروف حاشي للعيد',            'price': '40,000 MRU',  'bids': '7',  'time': '08h 00m 00s', 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal4.jpg', 'status': 'active',   'postedTime': 'منذ يوم'},
      {'id': 'an5', 'title': 'بقرة حلوب فريزيان',          'price': '85,000 MRU',  'bids': '4',  'time': '18h 00m 00s', 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal5.jpg', 'status': 'active',   'postedTime': 'منذ يومين'},
      {'id': 'an6', 'title': 'دجاج بلدي 50 رأس',           'price': '15,000 MRU',  'bids': '2',  'time': '06h 00m 00s', 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal6.jpg', 'status': 'active',   'postedTime': 'منذ 3 أيام'},
      {'id': 'an7', 'title': 'كبش إيل دو فرانس',           'price': '55,000 MRU',  'bids': '0',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal7.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوع'},
      {'id': 'an8', 'title': 'إبل للمزرعة',                'price': '220,000 MRU', 'bids': '6',  'time': l10n.text_364, 'location': l10n.text_113, 'category': 'animals', 'image': 'assets/animal8.jpg', 'status': 'finished', 'postedTime': 'منذ أسبوعين'},
    ];
    _filterAuctions();
  }

  @override
  void initState() {
    super.initState();
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged() {
    _filterAuctions();
  }

  void _filterAuctions() {
    setState(() {
      _filteredAuctions = _allAuctions.where((auction) {
        final matchesSearch = auction['title']!.toLowerCase().contains(_searchController.text.toLowerCase());
        final matchesStatus = _activeTabIndex == 0 ? auction['status'] == 'active' : auction['status'] == 'finished';

        bool matchesCategory = true;
        if (_selectedCategoryIndex != 0) {
          final categoryKey = _categories[_selectedCategoryIndex]['key'];
          matchesCategory = auction['category'] == categoryKey;
        }

        bool matchesSubCategory = true;
        if (_selectedCategoryIndex == 1 && _selectedSubCategoryIndex != -1) {
          final subKey = _carSubCategories[_selectedSubCategoryIndex]['key'];
          if (auction.containsKey('subCategory')) {
            matchesSubCategory = auction['subCategory'] == subKey;
          }
        }

        return matchesSearch && matchesStatus && matchesCategory && matchesSubCategory;
      }).toList();
    });
  }

  final List<Map<String, dynamic>> _categories = [
    {'title_key': 'text_74',  'image': 'assets/auctions/other.png',            'key': 'all',             'count': 76},
    {'title_key': 'text_86',  'image': 'assets/auctions/cars.png',             'key': 'cars',            'count': 4},
    {'title_key': 'text_130', 'image': 'assets/auctions/phones.png',           'key': 'phones',          'count': 6},
    {'title_key': 'text_119', 'image': 'assets/auctions/houses.png',           'key': 'real_estate',     'count': 5},
    {'title_key': 'text_127', 'image': 'assets/auctions/phone_numbers.png',    'key': 'phone_numbers',   'count': 0},
    {'title_key': 'text_120', 'image': 'assets/auctions/home_appliances.png',  'key': 'home_appliances', 'count': 8},
    {'title_key': 'text_124', 'image': 'assets/auctions/animals.png',          'key': 'animals',         'count': 8},
    {'title_key': 'text_123', 'image': 'assets/auctions/womens_accessories.png','key': 'womens',         'count': 8},
    {'title_key': 'text_122', 'image': 'assets/auctions/mens_accessories.png', 'key': 'mens',            'count': 10},
    {'title_key': 'text_125', 'image': 'assets/auctions/heavy_equipment.png',  'key': 'trucks',          'count': 8},
    {'title_key': 'text_131', 'image': 'assets/auctions/phones.png',                    'key': 'electronics',    'count': 8},
    {'title_key': 'text_129', 'image': 'assets/auctions/selling_projects.png',          'key': 'projects',       'count': 3},
    {'title_key': 'text_133', 'image': 'assets/auctions/Vélos.png',                     'key': 'bikes',          'count': 12},
    {'title_key': 'text_139', 'image': 'assets/auctions/Matériaux lourds.jpg',          'key': 'heavy_materials', 'count': 10},
    {'title_key': 'text_134', 'image': 'assets/auctions/Pièces de sol.jpg',             'key': 'land_plots',     'count': 4},
    {'title_key': 'text_137', 'image': 'assets/auctions/Meubles.jpg',                   'key': 'furniture',      'count': 8},
  ];

  final List<Map<String, dynamic>> _carSubCategories = [
    {'title_key': 'text_115', 'image': 'assets/car1.jpg', 'key': 'standard'},
    {'title_key': 'text_114', 'image': 'assets/car0.png', 'key': '4x4'},
    {'title_key': 'text_116', 'image': 'assets/car2.jpg', 'key': 'taxi'},
    {'title_key': 'text_117', 'image': 'assets/car4.jpg', 'key': 'damaged'},
    {'title_key': 'text_118', 'image': 'assets/car5.png', 'key': 'electric'},
  ];

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      endDrawer: const SideMenuDrawer(),
      appBar: AppBar(
        backgroundColor: primaryBlue,
        elevation: 0,
        toolbarHeight: 60,
        centerTitle: true,
        title: Text(
          "انواع المزادات",
          style: GoogleFonts.plusJakartaSans(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        leading: Builder(
          builder: (context) => IconButton(
            icon: const Icon(Icons.menu, color: Colors.white),
            onPressed: () => Scaffold.of(context).openEndDrawer(),
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.arrow_forward_ios, color: Colors.white, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          SliverToBoxAdapter(
            child: Column(
              children: [
                // Top Categories Horizontal List (Photo-card style)
                const SizedBox(height: 14),
                SizedBox(
                  height: 110,
                  child: ListView.builder(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    itemCount: _categories.length,
                    itemBuilder: (context, index) {
                      final cat = _categories[index];
                      bool isSelected = _selectedCategoryIndex == index;
                      String title = _getLocalizedTitle(cat['title_key'], l10n);

                      return GestureDetector(
                        onTap: () {
                          setState(() {
                            _selectedCategoryIndex = index;
                            _selectedSubCategoryIndex = -1;
                            _filterAuctions();
                          });
                        },
                        child: Container(
                          width: 100,
                          margin: const EdgeInsets.only(right: 10, top: 4, bottom: 4),
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: isSelected ? primaryBlue : Colors.transparent,
                              width: 2.5,
                            ),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withValues(alpha: isSelected ? 0.15 : 0.07),
                                blurRadius: 8,
                                offset: const Offset(0, 4),
                              ),
                            ],
                          ),
                          child: Stack(
                            clipBehavior: Clip.none,
                            children: [
                              // Full image background
                              ClipRRect(
                                borderRadius: BorderRadius.circular(14),
                                child: Image.asset(
                                  cat['image'],
                                  width: 100,
                                  height: 110,
                                  fit: BoxFit.cover,
                                  errorBuilder: (c, e, s) => Container(
                                    color: isDarkMode ? const Color(0xFF2C2C2C) : const Color(0xFFEEEEEE),
                                    child: Center(child: Icon(Icons.category, color: isSelected ? primaryBlue : Colors.grey, size: 36)),
                                  ),
                                ),
                              ),
                              // Dark gradient overlay at bottom
                              ClipRRect(
                                borderRadius: BorderRadius.circular(14),
                                child: Container(
                                  decoration: BoxDecoration(
                                    gradient: LinearGradient(
                                      begin: Alignment.topCenter,
                                      end: Alignment.bottomCenter,
                                      colors: [
                                        Colors.transparent,
                                        Colors.black.withValues(alpha: 0.55),
                                      ],
                                      stops: const [0.4, 1.0],
                                    ),
                                  ),
                                ),
                              ),
                              // Title at bottom
                              Positioned(
                                bottom: 8,
                                left: 4,
                                right: 4,
                                child: Text(
                                  title,
                                  textAlign: TextAlign.center,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                              // Count badge top-right
                              Positioned(
                                top: -4,
                                right: -4,
                                child: Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                                  decoration: const BoxDecoration(
                                    color: Color(0xFFFF3B30),
                                    shape: BoxShape.circle,
                                  ),
                                  child: Text(
                                    cat['count'].toString(),
                                    style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),

                // Search Bar
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 12.0),
                  child: Container(
                    height: 48,
                    decoration: BoxDecoration(
                      color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                      borderRadius: BorderRadius.circular(24),
                      border: Border.all(color: Colors.grey.withValues(alpha: 0.1)),
                      boxShadow: [
                        BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 10, offset: const Offset(0, 4))
                      ],
                    ),
                    child: TextField(
                      controller: _searchController,
                      textAlign: TextAlign.center,
                      decoration: InputDecoration(
                        hintText: "البحث",
                        hintStyle: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 15),
                        prefixIcon: const Icon(Icons.search, color: Colors.grey, size: 22),
                        border: InputBorder.none,
                        contentPadding: const EdgeInsets.symmetric(vertical: 12),
                      ),
                    ),
                  ),
                ),

                // Sub-Categories selection (cars only)
                if (_selectedCategoryIndex == 1) ...[
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                    child: Align(
                      alignment: Alignment.centerRight,
                      child: Text(
                        "اختر نوعية السيارة التي تبحث عنها",
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: isDarkMode ? Colors.white : Colors.black,
                        ),
                      ),
                    ),
                  ),
                  SizedBox(
                    height: 100,
                    child: ListView.builder(
                      scrollDirection: Axis.horizontal,
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      itemCount: _carSubCategories.length,
                      itemBuilder: (context, index) {
                        final sub = _carSubCategories[index];
                        bool isSelected = _selectedSubCategoryIndex == index;
                        return GestureDetector(
                          onTap: () {
                            setState(() {
                              _selectedSubCategoryIndex = index;
                              _filterAuctions();
                            });
                          },
                          child: Container(
                            width: 110,
                            margin: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                            decoration: BoxDecoration(
                              color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: isSelected ? primaryBlue : Colors.grey.withValues(alpha: 0.15),
                                width: isSelected ? 2 : 1,
                              ),
                              boxShadow: [
                                BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 6, offset: const Offset(0, 3))
                              ],
                            ),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Image.asset(
                                  sub['image'],
                                  height: 52,
                                  fit: BoxFit.contain,
                                  errorBuilder: (c, e, s) => Icon(Icons.directions_car, color: isSelected ? primaryBlue : Colors.grey, size: 30),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  _getLocalizedTitle(sub['title_key'], l10n),
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 12,
                                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                                    color: isSelected ? primaryBlue : (isDarkMode ? Colors.white70 : Colors.black87),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ],

                // Custom Tabs — dynamic count
                const SizedBox(height: 16),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: Builder(builder: (context) {
                    // Count active vs finished from all auctions (matching category filter)
                    int activeCount = _allAuctions.where((a) {
                      if (a['status'] != 'active') return false;
                      if (_selectedCategoryIndex != 0) {
                        return a['category'] == _categories[_selectedCategoryIndex]['key'];
                      }
                      return true;
                    }).length;
                    int finishedCount = _allAuctions.where((a) {
                      if (a['status'] != 'finished') return false;
                      if (_selectedCategoryIndex != 0) {
                        return a['category'] == _categories[_selectedCategoryIndex]['key'];
                      }
                      return true;
                    }).length;
                    return Row(
                      children: [
                        _buildStitchTab(1, "مزايدات منتهية", finishedCount.toString().padLeft(2, '0'), const Color(0xFFFFEDED), const Color(0xFFFF3B30)),
                        const SizedBox(width: 12),
                        _buildStitchTab(0, "مزايدات نشطة", activeCount.toString().padLeft(2, '0'), const Color(0xFFE8F5E9), const Color(0xFF2E7D32)),
                      ],
                    );
                  }),
                ),
                const SizedBox(height: 16),
              ],
            ),
          ),

          // Auction List
          _filteredAuctions.isEmpty
              ? SliverFillRemaining(
                  child: Center(
                    child: Text(
                      l10n.text_54,
                      style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 16),
                    ),
                  ),
                )
              : SliverPadding(
                  padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, index) => _buildHorizontalAuctionCard(
                        context,
                        isDarkMode,
                        _filteredAuctions[index],
                      ),
                      childCount: _filteredAuctions.length,
                    ),
                  ),
                ),
        ],
      ),

      floatingActionButton: FloatingActionButton(
        onPressed: () => Navigator.of(context).push(MaterialPageRoute(builder: (context) => const CreateAdStartPage())),
        backgroundColor: Colors.transparent,
        elevation: 0,
        highlightElevation: 0,
        child: Image.asset('assets/botum_bar.png', fit: BoxFit.contain),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: BottomAppBar(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        elevation: 0,
        shape: const CircularNotchedRectangle(),
        notchMargin: 6,
        child: SizedBox(
          height: 60,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildSimpleNavItem(Icons.home_outlined, l10n.text_1, 0),
              _buildSimpleNavItem(Icons.local_shipping_outlined, l10n.text_32, 1),
              const SizedBox(width: 48),
              _buildSimpleNavItem(Icons.storefront_outlined, l10n.text_33, 2),
              _buildSimpleNavItem(Icons.person_outline, l10n.text_19, 3),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStitchTab(int index, String label, String count, Color bgColor, Color textColor) {
    bool isSelected = _activeTabIndex == index;
    return Expanded(
      child: GestureDetector(
        onTap: () {
          setState(() {
            _activeTabIndex = index;
            _filterAuctions();
          });
        },
        child: Container(
          height: 45,
          decoration: BoxDecoration(
            color: isSelected ? bgColor : Colors.grey.withValues(alpha: 0.05),
            borderRadius: BorderRadius.circular(12),
            border: isSelected ? Border.all(color: textColor.withValues(alpha: 0.3)) : null,
          ),
          child: Center(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  label,
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 14,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                    color: isSelected ? textColor : Colors.grey,
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: isSelected ? textColor : Colors.grey.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    count,
                    style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHorizontalAuctionCard(BuildContext context, bool isDarkMode, Map<String, String> auction) {
    const Color primaryBlue = Color(0xFF0084FF);
    const Color softRed = Color(0xFFFF3B30);
    const Color lightGreyBg = Color(0xFFF2F2F7);
    const Color darkGreyBg = Color(0xFF2C2C2E);
    
    final id = auction['id']!;
    final isFavorite = ref.watch(favoritesProvider).contains(id);

    final bool isFinished = auction['status'] == 'finished';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
            spreadRadius: -5,
          ),
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.03),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: IntrinsicHeight(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Right: Image
            ClipRRect(
              borderRadius: const BorderRadius.only(
                topRight: Radius.circular(16),
                bottomRight: Radius.circular(16),
              ),
              child: SizedBox(
                width: 125,
                height: 110,
                child: Image.asset(
                  auction['image']!,
                  fit: BoxFit.cover,
                  errorBuilder: (c, e, s) => Container(
                    color: Colors.grey[200],
                    child: const Icon(Icons.image_not_supported, color: Colors.grey),
                  ),
                ),
              ),
            ),
            // Left: Content
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6), // Reduced from 10
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    // Title
                    Text(
                      auction['title']!,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 14, // Reduced from 15
                        fontWeight: FontWeight.w800,
                        color: isDarkMode ? Colors.white : Colors.black,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      textAlign: TextAlign.right,
                    ),
                    const SizedBox(height: 1), // Reduced from 2
                    // Price
                    Text(
                      "${auction['price']?.split(' ')[0]} أوقية جديدة",
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13, // Reduced from 14
                        fontWeight: FontWeight.w600,
                        color: isDarkMode ? Colors.white70 : Colors.black87,
                      ),
                      textAlign: TextAlign.right,
                    ),
                    const SizedBox(height: 6), // Reduced from 10
                    // Interaction Row: [Timer] [Heart] [Bids+Gavel]
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        // Bid count + gavel (leftmost = endmost in RTL code)
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              auction['bids']!,
                              style: GoogleFonts.plusJakartaSans(
                                color: softRed,
                                fontWeight: FontWeight.w900,
                                fontSize: 14, // Reduced from 16
                              ),
                            ),
                            const SizedBox(width: 3),
                            Icon(
                              Icons.gavel_rounded,
                              color: isDarkMode ? Colors.white54 : Colors.black54,
                              size: 16, // Reduced from 18
                            ),
                          ],
                        ),
                        // Heart / favorite
                        GestureDetector(
                          onTap: () => ref.read(favoritesProvider.notifier).toggleFavorite(id),
                          child: Icon(
                            isFavorite ? Icons.favorite : Icons.favorite_border,
                            color: isFavorite ? softRed : (isDarkMode ? Colors.white54 : Colors.grey),
                            size: 20, // Reduced from 22
                          ),
                        ),
                        // Timer (rightmost = startmost in RTL code)
                        Flexible(
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Flexible(
                                child: Text(
                                  isFinished ? 'انتهى المزاد' : auction['time']!,
                                  style: GoogleFonts.plusJakartaSans(
                                    color: softRed,
                                    fontSize: 11,
                                    fontWeight: FontWeight.w800,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              const SizedBox(width: 3),
                              const Icon(Icons.access_time_rounded, color: softRed, size: 14),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const Spacer(),
                    // Posted time + location
                    Directionality(
                      textDirection: TextDirection.rtl,
                      child: Text(
                        "${auction['postedTime']} . ${auction['location']}",
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 10, // Reduced from 11
                          color: const Color(0xFF9AA5B4),
                          fontWeight: FontWeight.w500,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSimpleNavItem(IconData icon, String label, int index) {
    return InkWell(
      onTap: () {
        if (index == 0) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const HomePage()));
        } else if (index == 1) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const ServicesPage()));
        } else if (index == 3) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const AccountPage()));
        }
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: Colors.grey[600], size: 24),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 9, color: Colors.grey)),
        ],
      ),
    );
  }

  String _getLocalizedTitle(String key, AppLocalizations l10n) {
    switch (key) {
      case 'text_74': return l10n.text_74;
      case 'text_86': return l10n.text_86;
      case 'text_130': return l10n.text_130;
      case 'text_119': return l10n.text_119;
      case 'text_127': return l10n.text_127;
      case 'text_120': return l10n.text_120;
      case 'text_114': return l10n.text_114;
      case 'text_115': return l10n.text_115;
      case 'text_116': return l10n.text_116;
      case 'text_117': return l10n.text_117;
      case 'text_118': return l10n.text_118;
      case 'text_122': return l10n.text_122;
      case 'text_123': return l10n.text_123;
      case 'text_124': return l10n.text_124;
      case 'text_125': return l10n.text_125;
      case 'text_129': return l10n.text_129;
      case 'text_131': return l10n.text_131;
      case 'text_133': return l10n.text_133;
      case 'text_134': return l10n.text_134;
      case 'text_137': return l10n.text_137;
      case 'text_139': return l10n.text_139;
      default: return "Category";
    }
  }
}
