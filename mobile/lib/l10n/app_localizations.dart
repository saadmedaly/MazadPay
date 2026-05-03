import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_ar.dart';
import 'app_localizations_en.dart';
import 'app_localizations_fr.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('ar'),
    Locale('en'),
    Locale('fr'),
  ];

  /// No description provided for @text_1.
  ///
  /// In ar, this message translates to:
  /// **'الرئيسية'**
  String get text_1;

  /// No description provided for @text_2.
  ///
  /// In ar, this message translates to:
  /// **'جميع المزادات'**
  String get text_2;

  /// No description provided for @text_3.
  ///
  /// In ar, this message translates to:
  /// **'زايد الآن'**
  String get text_3;

  /// No description provided for @text_4.
  ///
  /// In ar, this message translates to:
  /// **'حول مزاد باي (Mazad Pay)'**
  String get text_4;

  /// No description provided for @text_5.
  ///
  /// In ar, this message translates to:
  /// **'مزاد باي هو أول تطبيق للمزادات الرقمية في موريتانيا، والمنصة الرائدة التي تنقل مفهوم المزايدة التقليدية إلى تجربة تقنية حديثة، آمنة، وشفافة. نحن فخورون لكوننا أول من أطلق هذا النظام المتكامل في السوق الموريتاني، لنجمع لك بين إثارة المزايدة وتلبية احتياجاتك اليومية في تطبيق واحد وبسواعد وطنية.'**
  String get text_5;

  /// No description provided for @text_6.
  ///
  /// In ar, this message translates to:
  /// **'خدماتنا الرائدة:'**
  String get text_6;

  /// No description provided for @text_7.
  ///
  /// In ar, this message translates to:
  /// **'من خلال تطبيقنا، نضع بين يديك باقة متنوعة من الخدمات المصممة خصيصاً لمجتمعنا:'**
  String get text_7;

  /// No description provided for @text_8.
  ///
  /// In ar, this message translates to:
  /// **'المزادات الرقمية (الأولى في موريتانيا): نظام مزايدة علني وشفاف على السيارات، العقارات، والإلكترونيات، يضمن حقوق الجميع بكل موثوقية.'**
  String get text_8;

  /// No description provided for @text_9.
  ///
  /// In ar, this message translates to:
  /// **'التجارة الإلكترونية العالمية: منصة تسوق ذكية تربطك بأفضل المواقع العالمية مثل (Amazon, AliExpress, Shein) لتصلك منتجاتك المفضلة أينما كنت.'**
  String get text_9;

  /// No description provided for @text_10.
  ///
  /// In ar, this message translates to:
  /// **'خدمات التوصيل والنقل الشاملة: حلول متكاملة تشمل نقل الأشخاص، شحن البضائع، وخدمات سحب ونقل السيارات (الحاملات).'**
  String get text_10;

  /// No description provided for @text_11.
  ///
  /// In ar, this message translates to:
  /// **'حجز الأنشطة الرياضية: إمكانية حجز مختلف أنواع الرياضات؛ بما في ذلك صالات اللياقة البدنية، الملاعب، وحصص السباحة.'**
  String get text_11;

  /// No description provided for @text_12.
  ///
  /// In ar, this message translates to:
  /// **'طلب الطعام: استمتع بطلب وجباتك من أشهر المطاعم ومقاهي بضغطة زر.'**
  String get text_12;

  /// No description provided for @text_13.
  ///
  /// In ar, this message translates to:
  /// **'حجز الفنادق والإقامات: خطط لرحلاتك داخل وخارج موريتانيا بأسعار تنافسية.'**
  String get text_13;

  /// No description provided for @text_14.
  ///
  /// In ar, this message translates to:
  /// **'المواعيد الطبية: تنظيم وتسهيل حجز المواعيد في كبرى المصحات والعيادات لضمان راحتك وصحتك.'**
  String get text_14;

  /// No description provided for @text_15.
  ///
  /// In ar, this message translates to:
  /// **'رؤيتنا'**
  String get text_15;

  /// No description provided for @text_16.
  ///
  /// In ar, this message translates to:
  /// **'الريادة المطلقة كأول وجهة رقمية للمزادات في موريتانيا، وتقديم حلول شاملة تدمج بين التجارة الإلكترونية، مختلف أنواع توصيل، وخدمات، لتعزيز نمط حياة المواطن الموريتاني وتسهيل معاملاته.'**
  String get text_16;

  /// No description provided for @text_17.
  ///
  /// In ar, this message translates to:
  /// **'التزامنا'**
  String get text_17;

  /// No description provided for @text_18.
  ///
  /// In ar, this message translates to:
  /// **'بصفتنا أول مزاد رقمي في البلاد، يلتزم فريق مزاد باي بتوفير بيئة رقمية آمنة كلياً، تلتزم بأعلى معايير المصداقية والشفافية، وتدعم التحول الرقمي في موريتانيا.'**
  String get text_18;

  /// No description provided for @text_19.
  ///
  /// In ar, this message translates to:
  /// **'حسابي'**
  String get text_19;

  /// No description provided for @text_20.
  ///
  /// In ar, this message translates to:
  /// **'مبلغ التأمين'**
  String get text_20;

  /// No description provided for @text_21.
  ///
  /// In ar, this message translates to:
  /// **'الرصيد المتوفر'**
  String get text_21;

  /// No description provided for @text_22.
  ///
  /// In ar, this message translates to:
  /// **'اختر طريقة'**
  String get text_22;

  /// No description provided for @text_23.
  ///
  /// In ar, this message translates to:
  /// **'قم بالإيداع الان'**
  String get text_23;

  /// No description provided for @text_24.
  ///
  /// In ar, this message translates to:
  /// **'ابدا رحلة المزايدة الخاصة بك!'**
  String get text_24;

  /// No description provided for @text_25.
  ///
  /// In ar, this message translates to:
  /// **'قم باسترجاع مبلغ التأمين'**
  String get text_25;

  /// No description provided for @text_26.
  ///
  /// In ar, this message translates to:
  /// **'أنشطتك'**
  String get text_26;

  /// No description provided for @text_27.
  ///
  /// In ar, this message translates to:
  /// **'مزاداتي'**
  String get text_27;

  /// No description provided for @text_28.
  ///
  /// In ar, this message translates to:
  /// **'المفضلة'**
  String get text_28;

  /// No description provided for @text_29.
  ///
  /// In ar, this message translates to:
  /// **'العناصر التي فزت بها'**
  String get text_29;

  /// No description provided for @text_30.
  ///
  /// In ar, this message translates to:
  /// **'مركز التواصل'**
  String get text_30;

  /// No description provided for @text_31.
  ///
  /// In ar, this message translates to:
  /// **'معلومات الحساب الشخصي'**
  String get text_31;

  /// No description provided for @text_32.
  ///
  /// In ar, this message translates to:
  /// **'توصيل'**
  String get text_32;

  /// No description provided for @text_33.
  ///
  /// In ar, this message translates to:
  /// **'التجارة الالكترونية'**
  String get text_33;

  /// No description provided for @text_34.
  ///
  /// In ar, this message translates to:
  /// **'الصفحة غير متاحة حاليا'**
  String get text_34;

  /// No description provided for @text_35.
  ///
  /// In ar, this message translates to:
  /// **'معلومات الحساب'**
  String get text_35;

  /// No description provided for @text_36.
  ///
  /// In ar, this message translates to:
  /// **'ب'**
  String get text_36;

  /// No description provided for @text_37.
  ///
  /// In ar, this message translates to:
  /// **'بدال سيديا'**
  String get text_37;

  /// No description provided for @text_38.
  ///
  /// In ar, this message translates to:
  /// **'معلوماتي'**
  String get text_38;

  /// No description provided for @text_39.
  ///
  /// In ar, this message translates to:
  /// **'الاسم الكامل'**
  String get text_39;

  /// No description provided for @text_40.
  ///
  /// In ar, this message translates to:
  /// **'رقم الهاتف'**
  String get text_40;

  /// No description provided for @text_41.
  ///
  /// In ar, this message translates to:
  /// **'البريد الإلكتروني'**
  String get text_41;

  /// No description provided for @text_42.
  ///
  /// In ar, this message translates to:
  /// **'المدينة'**
  String get text_42;

  /// No description provided for @text_43.
  ///
  /// In ar, this message translates to:
  /// **'نواكشوط'**
  String get text_43;

  /// No description provided for @text_44.
  ///
  /// In ar, this message translates to:
  /// **'الإعدادات'**
  String get text_44;

  /// No description provided for @text_45.
  ///
  /// In ar, this message translates to:
  /// **'تغيير كلمة المرور'**
  String get text_45;

  /// No description provided for @text_46.
  ///
  /// In ar, this message translates to:
  /// **'اللغة'**
  String get text_46;

  /// No description provided for @text_47.
  ///
  /// In ar, this message translates to:
  /// **'العربية'**
  String get text_47;

  /// No description provided for @text_48.
  ///
  /// In ar, this message translates to:
  /// **'الإشعارات'**
  String get text_48;

  /// No description provided for @text_49.
  ///
  /// In ar, this message translates to:
  /// **'تسجيل الخروج'**
  String get text_49;

  /// No description provided for @text_50.
  ///
  /// In ar, this message translates to:
  /// **'تم نشر الإعلان بنجاح!'**
  String get text_50;

  /// No description provided for @text_51.
  ///
  /// In ar, this message translates to:
  /// **'يتم مراجعته و سيظهر للجميع قريباً'**
  String get text_51;

  /// No description provided for @text_52.
  ///
  /// In ar, this message translates to:
  /// **'حَسناً'**
  String get text_52;

  /// No description provided for @text_53.
  ///
  /// In ar, this message translates to:
  /// **'ابحث عن مزاد...'**
  String get text_53;

  /// No description provided for @text_54.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد نتائج بحث'**
  String get text_54;

  /// No description provided for @text_55.
  ///
  /// In ar, this message translates to:
  /// **'تمت الإزالة من المفضلة'**
  String get text_55;

  /// No description provided for @text_56.
  ///
  /// In ar, this message translates to:
  /// **'تم الإضافة إلى المفضلة'**
  String get text_56;

  /// No description provided for @text_57.
  ///
  /// In ar, this message translates to:
  /// **'السعر الحالي'**
  String get text_57;

  /// No description provided for @text_58.
  ///
  /// In ar, this message translates to:
  /// **'الحد الأدنى للزيادة'**
  String get text_58;

  /// No description provided for @text_59.
  ///
  /// In ar, this message translates to:
  /// **'ثانية'**
  String get text_59;

  /// No description provided for @text_60.
  ///
  /// In ar, this message translates to:
  /// **'دقيقة'**
  String get text_60;

  /// No description provided for @text_61.
  ///
  /// In ar, this message translates to:
  /// **'ساعة'**
  String get text_61;

  /// No description provided for @text_62.
  ///
  /// In ar, this message translates to:
  /// **'أحكام وشروط المزايدة'**
  String get text_62;

  /// No description provided for @text_63.
  ///
  /// In ar, this message translates to:
  /// **'تواصل مع فريق الدعم'**
  String get text_63;

  /// No description provided for @text_64.
  ///
  /// In ar, this message translates to:
  /// **'مشاهدة كل المزايدين'**
  String get text_64;

  /// No description provided for @text_65.
  ///
  /// In ar, this message translates to:
  /// **'الشركة المصنعة'**
  String get text_65;

  /// No description provided for @text_66.
  ///
  /// In ar, this message translates to:
  /// **'نقل'**
  String get text_66;

  /// No description provided for @text_67.
  ///
  /// In ar, this message translates to:
  /// **'نوع الوقود'**
  String get text_67;

  /// No description provided for @text_68.
  ///
  /// In ar, this message translates to:
  /// **'السنة'**
  String get text_68;

  /// No description provided for @text_69.
  ///
  /// In ar, this message translates to:
  /// **'عدد الأميال'**
  String get text_69;

  /// No description provided for @text_70.
  ///
  /// In ar, this message translates to:
  /// **'الطراز'**
  String get text_70;

  /// No description provided for @text_71.
  ///
  /// In ar, this message translates to:
  /// **'انت الان المزايد الأعلى'**
  String get text_71;

  /// No description provided for @text_72.
  ///
  /// In ar, this message translates to:
  /// **'قم بالمزايدة الان'**
  String get text_72;

  /// No description provided for @text_73.
  ///
  /// In ar, this message translates to:
  /// **'مزادات متشابهة'**
  String get text_73;

  /// No description provided for @text_74.
  ///
  /// In ar, this message translates to:
  /// **'عرض الكل'**
  String get text_74;

  /// No description provided for @text_75.
  ///
  /// In ar, this message translates to:
  /// **'سجل المزايدات'**
  String get text_75;

  /// No description provided for @text_76.
  ///
  /// In ar, this message translates to:
  /// **'الفائز : محمد احمد'**
  String get text_76;

  /// No description provided for @text_77.
  ///
  /// In ar, this message translates to:
  /// **'الملخص'**
  String get text_77;

  /// No description provided for @text_78.
  ///
  /// In ar, this message translates to:
  /// **'عدد مزايدات'**
  String get text_78;

  /// No description provided for @text_79.
  ///
  /// In ar, this message translates to:
  /// **'عدد المزايدون'**
  String get text_79;

  /// No description provided for @text_80.
  ///
  /// In ar, this message translates to:
  /// **'المزايدات الأخيرة'**
  String get text_80;

  /// No description provided for @text_81.
  ///
  /// In ar, this message translates to:
  /// **'مبروك!'**
  String get text_81;

  /// No description provided for @text_82.
  ///
  /// In ar, this message translates to:
  /// **'ربحت المزاد'**
  String get text_82;

  /// No description provided for @text_83.
  ///
  /// In ar, this message translates to:
  /// **'محمد احمد سيديا'**
  String get text_83;

  /// No description provided for @text_84.
  ///
  /// In ar, this message translates to:
  /// **'الفائز الاول بالمزاد'**
  String get text_84;

  /// No description provided for @text_85.
  ///
  /// In ar, this message translates to:
  /// **'أكمل عملية الدفع'**
  String get text_85;

  /// No description provided for @text_86.
  ///
  /// In ar, this message translates to:
  /// **'سيارات'**
  String get text_86;

  /// No description provided for @text_87.
  ///
  /// In ar, this message translates to:
  /// **'سيارات رباعية الدفع'**
  String get text_87;

  /// No description provided for @text_88.
  ///
  /// In ar, this message translates to:
  /// **'مدريد'**
  String get text_88;

  /// No description provided for @text_89.
  ///
  /// In ar, this message translates to:
  /// **'مزايدة'**
  String get text_89;

  /// No description provided for @text_90.
  ///
  /// In ar, this message translates to:
  /// **'إنشاء إعلان جديد'**
  String get text_90;

  /// No description provided for @text_91.
  ///
  /// In ar, this message translates to:
  /// **'اسم الإعلان'**
  String get text_91;

  /// No description provided for @text_92.
  ///
  /// In ar, this message translates to:
  /// **'أدخل اسم منتجك'**
  String get text_92;

  /// No description provided for @text_93.
  ///
  /// In ar, this message translates to:
  /// **'الوصف'**
  String get text_93;

  /// No description provided for @text_94.
  ///
  /// In ar, this message translates to:
  /// **'اكتب وصفا'**
  String get text_94;

  /// No description provided for @text_95.
  ///
  /// In ar, this message translates to:
  /// **'لتواصل معك'**
  String get text_95;

  /// No description provided for @text_96.
  ///
  /// In ar, this message translates to:
  /// **'السعر'**
  String get text_96;

  /// No description provided for @text_97.
  ///
  /// In ar, this message translates to:
  /// **'أدخل السعر'**
  String get text_97;

  /// No description provided for @text_98.
  ///
  /// In ar, this message translates to:
  /// **'اختر الفئة الرئيسية'**
  String get text_98;

  /// No description provided for @text_99.
  ///
  /// In ar, this message translates to:
  /// **'اختر الفئة الفرعية'**
  String get text_99;

  /// No description provided for @text_100.
  ///
  /// In ar, this message translates to:
  /// **'اختر الموقع'**
  String get text_100;

  /// No description provided for @text_101.
  ///
  /// In ar, this message translates to:
  /// **'صور وفيديوهات المنتج'**
  String get text_101;

  /// No description provided for @text_102.
  ///
  /// In ar, this message translates to:
  /// **'نشر الإعلان'**
  String get text_102;

  /// No description provided for @text_103.
  ///
  /// In ar, this message translates to:
  /// **'اختر المدينة'**
  String get text_103;

  /// No description provided for @text_104.
  ///
  /// In ar, this message translates to:
  /// **'عرفات'**
  String get text_104;

  /// No description provided for @text_105.
  ///
  /// In ar, this message translates to:
  /// **'توجنين'**
  String get text_105;

  /// No description provided for @text_106.
  ///
  /// In ar, this message translates to:
  /// **'لكصر'**
  String get text_106;

  /// No description provided for @text_107.
  ///
  /// In ar, this message translates to:
  /// **'دار نعيم'**
  String get text_107;

  /// No description provided for @text_108.
  ///
  /// In ar, this message translates to:
  /// **'عين الطلح'**
  String get text_108;

  /// No description provided for @text_109.
  ///
  /// In ar, this message translates to:
  /// **'الميناء'**
  String get text_109;

  /// No description provided for @text_110.
  ///
  /// In ar, this message translates to:
  /// **'الرياض'**
  String get text_110;

  /// No description provided for @text_111.
  ///
  /// In ar, this message translates to:
  /// **'السبخة'**
  String get text_111;

  /// No description provided for @text_112.
  ///
  /// In ar, this message translates to:
  /// **'تيارت'**
  String get text_112;

  /// No description provided for @text_113.
  ///
  /// In ar, this message translates to:
  /// **'تفرغ زينة'**
  String get text_113;

  /// No description provided for @text_114.
  ///
  /// In ar, this message translates to:
  /// **'سيارات رباعية دفع'**
  String get text_114;

  /// No description provided for @text_115.
  ///
  /// In ar, this message translates to:
  /// **'سيارات عادية'**
  String get text_115;

  /// No description provided for @text_116.
  ///
  /// In ar, this message translates to:
  /// **'سيارات تاكسي'**
  String get text_116;

  /// No description provided for @text_117.
  ///
  /// In ar, this message translates to:
  /// **'سيارات مضروبة'**
  String get text_117;

  /// No description provided for @text_118.
  ///
  /// In ar, this message translates to:
  /// **'سيارات كهربائية'**
  String get text_118;

  /// No description provided for @text_119.
  ///
  /// In ar, this message translates to:
  /// **'عقارات'**
  String get text_119;

  /// No description provided for @text_120.
  ///
  /// In ar, this message translates to:
  /// **'أجهزة منزلية'**
  String get text_120;

  /// No description provided for @text_121.
  ///
  /// In ar, this message translates to:
  /// **'قطع أرضية'**
  String get text_121;

  /// No description provided for @text_122.
  ///
  /// In ar, this message translates to:
  /// **'مستلزمات رجالية'**
  String get text_122;

  /// No description provided for @text_123.
  ///
  /// In ar, this message translates to:
  /// **'مستلزمات نسائية'**
  String get text_123;

  /// No description provided for @text_124.
  ///
  /// In ar, this message translates to:
  /// **'حيوانات'**
  String get text_124;

  /// No description provided for @text_125.
  ///
  /// In ar, this message translates to:
  /// **'شاحنات'**
  String get text_125;

  /// No description provided for @text_126.
  ///
  /// In ar, this message translates to:
  /// **'مواد ثقيلة'**
  String get text_126;

  /// No description provided for @text_127.
  ///
  /// In ar, this message translates to:
  /// **'أرقام هاتف'**
  String get text_127;

  /// No description provided for @text_128.
  ///
  /// In ar, this message translates to:
  /// **'ربّاخة'**
  String get text_128;

  /// No description provided for @text_129.
  ///
  /// In ar, this message translates to:
  /// **'بيع مشاريع'**
  String get text_129;

  /// No description provided for @text_130.
  ///
  /// In ar, this message translates to:
  /// **'هواتف'**
  String get text_130;

  /// No description provided for @text_131.
  ///
  /// In ar, this message translates to:
  /// **'الكترونيات'**
  String get text_131;

  /// No description provided for @text_132.
  ///
  /// In ar, this message translates to:
  /// **'ساعات'**
  String get text_132;

  /// No description provided for @text_133.
  ///
  /// In ar, this message translates to:
  /// **'دراجات'**
  String get text_133;

  /// No description provided for @text_134.
  ///
  /// In ar, this message translates to:
  /// **'قطع ارضية'**
  String get text_134;

  /// No description provided for @text_135.
  ///
  /// In ar, this message translates to:
  /// **'الالكترونيات'**
  String get text_135;

  /// No description provided for @text_136.
  ///
  /// In ar, this message translates to:
  /// **'أدوات منزلية'**
  String get text_136;

  /// No description provided for @text_137.
  ///
  /// In ar, this message translates to:
  /// **'الاثاث'**
  String get text_137;

  /// No description provided for @text_138.
  ///
  /// In ar, this message translates to:
  /// **'رباخة'**
  String get text_138;

  /// No description provided for @text_139.
  ///
  /// In ar, this message translates to:
  /// **'معدات ثقيلة'**
  String get text_139;

  /// No description provided for @text_140.
  ///
  /// In ar, this message translates to:
  /// **'ارقام ذهبية'**
  String get text_140;

  /// No description provided for @text_141.
  ///
  /// In ar, this message translates to:
  /// **'بحث...'**
  String get text_141;

  /// No description provided for @text_142.
  ///
  /// In ar, this message translates to:
  /// **'أثاث'**
  String get text_142;

  /// No description provided for @text_143.
  ///
  /// In ar, this message translates to:
  /// **'هل يوجد لديك منتجات تعرضها للمزايدة ؟ اختر اعلانك من هنا'**
  String get text_143;

  /// No description provided for @text_144.
  ///
  /// In ar, this message translates to:
  /// **'متابعة'**
  String get text_144;

  /// No description provided for @text_145.
  ///
  /// In ar, this message translates to:
  /// **'مزاد باي'**
  String get text_145;

  /// No description provided for @text_146.
  ///
  /// In ar, this message translates to:
  /// **'إنشاء ملفك الشخصي'**
  String get text_146;

  /// No description provided for @text_147.
  ///
  /// In ar, this message translates to:
  /// **'أدخل معلوماتك لإنهاء إنشاء حسابك'**
  String get text_147;

  /// No description provided for @text_148.
  ///
  /// In ar, this message translates to:
  /// **'أدخل اسمك الكامل'**
  String get text_148;

  /// No description provided for @text_149.
  ///
  /// In ar, this message translates to:
  /// **'تفاصيل التوصيل'**
  String get text_149;

  /// No description provided for @text_150.
  ///
  /// In ar, this message translates to:
  /// **'حالة التوصيل'**
  String get text_150;

  /// No description provided for @text_151.
  ///
  /// In ar, this message translates to:
  /// **'رقم التتبع'**
  String get text_151;

  /// No description provided for @text_152.
  ///
  /// In ar, this message translates to:
  /// **'الوصول المتوقع'**
  String get text_152;

  /// No description provided for @text_153.
  ///
  /// In ar, this message translates to:
  /// **'اليوم، 04:30 م'**
  String get text_153;

  /// No description provided for @text_154.
  ///
  /// In ar, this message translates to:
  /// **'تم استلام الطلب'**
  String get text_154;

  /// No description provided for @text_155.
  ///
  /// In ar, this message translates to:
  /// **'تم تأكيد طلبك بنجاح.'**
  String get text_155;

  /// No description provided for @text_156.
  ///
  /// In ar, this message translates to:
  /// **'10:00 ص'**
  String get text_156;

  /// No description provided for @text_157.
  ///
  /// In ar, this message translates to:
  /// **'في مرحلة التجهيز'**
  String get text_157;

  /// No description provided for @text_158.
  ///
  /// In ar, this message translates to:
  /// **'يتم فحص المنتج وتجهيزه للشحن.'**
  String get text_158;

  /// No description provided for @text_159.
  ///
  /// In ar, this message translates to:
  /// **'11:20 ص'**
  String get text_159;

  /// No description provided for @text_160.
  ///
  /// In ar, this message translates to:
  /// **'تم الشحن'**
  String get text_160;

  /// No description provided for @text_161.
  ///
  /// In ar, this message translates to:
  /// **'الطلب في طريقه إلى مدينتك.'**
  String get text_161;

  /// No description provided for @text_162.
  ///
  /// In ar, this message translates to:
  /// **'01:45 م'**
  String get text_162;

  /// No description provided for @text_163.
  ///
  /// In ar, this message translates to:
  /// **'قيد التوصيل'**
  String get text_163;

  /// No description provided for @text_164.
  ///
  /// In ar, this message translates to:
  /// **'المندوب في طريقه إلى عنوانك.'**
  String get text_164;

  /// No description provided for @text_165.
  ///
  /// In ar, this message translates to:
  /// **'قريباً'**
  String get text_165;

  /// No description provided for @text_166.
  ///
  /// In ar, this message translates to:
  /// **'عنوان التوصيل'**
  String get text_166;

  /// No description provided for @text_167.
  ///
  /// In ar, this message translates to:
  /// **'نواكشوط، تفرغ زينة، شارع المختار ولد داداه'**
  String get text_167;

  /// No description provided for @text_168.
  ///
  /// In ar, this message translates to:
  /// **'محمد الأمين'**
  String get text_168;

  /// No description provided for @text_169.
  ///
  /// In ar, this message translates to:
  /// **'مندوب التوصيل'**
  String get text_169;

  /// No description provided for @text_170.
  ///
  /// In ar, this message translates to:
  /// **'مصرفي'**
  String get text_170;

  /// No description provided for @text_171.
  ///
  /// In ar, this message translates to:
  /// **'استخدموا رمز التاجر الخاص بنا لإتمام عملية الدفع عبر مصرفي'**
  String get text_171;

  /// No description provided for @text_172.
  ///
  /// In ar, this message translates to:
  /// **'بنكلي'**
  String get text_172;

  /// No description provided for @text_173.
  ///
  /// In ar, this message translates to:
  /// **'استخدموا رمز التاجر الخاص بنا لإتمام عملية الدفع عبر بنكلي'**
  String get text_173;

  /// No description provided for @text_174.
  ///
  /// In ar, this message translates to:
  /// **'سداد'**
  String get text_174;

  /// No description provided for @text_175.
  ///
  /// In ar, this message translates to:
  /// **'استخدموا رمز التاجر الخاص بنا لإتمام عملية الدفع عبر سداد'**
  String get text_175;

  /// No description provided for @text_176.
  ///
  /// In ar, this message translates to:
  /// **'كليك'**
  String get text_176;

  /// No description provided for @text_177.
  ///
  /// In ar, this message translates to:
  /// **'استخدموا رمز التاجر الخاص بنا لإتمام عملية الدفع عبر كليك'**
  String get text_177;

  /// No description provided for @text_178.
  ///
  /// In ar, this message translates to:
  /// **'دفع من خلال تطبيقات البنكية'**
  String get text_178;

  /// No description provided for @text_179.
  ///
  /// In ar, this message translates to:
  /// **'يرجى قراءة الشروط والأحكام التالية لاسترداد المعاملات بعناية'**
  String get text_179;

  /// No description provided for @text_180.
  ///
  /// In ar, this message translates to:
  /// **'شروط الدفع والتأمين'**
  String get text_180;

  /// No description provided for @text_181.
  ///
  /// In ar, this message translates to:
  /// **'١. تقبل التطبيق وسائل الدفع: بنكلي، مصرفي، السداد.'**
  String get text_181;

  /// No description provided for @text_182.
  ///
  /// In ar, this message translates to:
  /// **'٢. يتطلب بعض المزادات دفع مبلغ تأمين لضمان جدية المزايدة.'**
  String get text_182;

  /// No description provided for @text_183.
  ///
  /// In ar, this message translates to:
  /// **'٣. يُسترجع مبلغ التأمين تلقائياً في حال عدم الفوز بالمزاد خلال ساعة حسب مزود الدفع.'**
  String get text_183;

  /// No description provided for @text_184.
  ///
  /// In ar, this message translates to:
  /// **'٤. لا يُسترجع التأمين في حال الفوز وعدم إتمام الدفع أو مخالفة شروط التطبيق.'**
  String get text_184;

  /// No description provided for @text_185.
  ///
  /// In ar, this message translates to:
  /// **'٥. يحق لإدارة التطبيق إلغاء أي مزاد عند وجود خطأ تقني أو شبهة احتيال، مع إعادة التأمين عند الإلغاء.'**
  String get text_185;

  /// No description provided for @text_186.
  ///
  /// In ar, this message translates to:
  /// **'باستخدامك للتطبيق، فإنك توافق على هذه الشروط'**
  String get text_186;

  /// No description provided for @text_187.
  ///
  /// In ar, this message translates to:
  /// **'إلغاء'**
  String get text_187;

  /// No description provided for @text_188.
  ///
  /// In ar, this message translates to:
  /// **'اوافق على شروط'**
  String get text_188;

  /// No description provided for @text_189.
  ///
  /// In ar, this message translates to:
  /// **'إضافة إيداع'**
  String get text_189;

  /// No description provided for @text_190.
  ///
  /// In ar, this message translates to:
  /// **'طريقة الدفع'**
  String get text_190;

  /// No description provided for @text_191.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد عناصر في المفضلة'**
  String get text_191;

  /// No description provided for @text_192.
  ///
  /// In ar, this message translates to:
  /// **'زايد الان'**
  String get text_192;

  /// No description provided for @text_193.
  ///
  /// In ar, this message translates to:
  /// **'انواكشوط'**
  String get text_193;

  /// No description provided for @text_194.
  ///
  /// In ar, this message translates to:
  /// **'انواذيبو'**
  String get text_194;

  /// No description provided for @text_195.
  ///
  /// In ar, this message translates to:
  /// **'خدمات مزاد موريتانيا بحاجة إلى موقعك'**
  String get text_195;

  /// No description provided for @text_196.
  ///
  /// In ar, this message translates to:
  /// **'لتجربة أفضل يرجى تفعيل الموقع الجغرافي في هاتفك'**
  String get text_196;

  /// No description provided for @text_197.
  ///
  /// In ar, this message translates to:
  /// **'تفعيل  تحديد الموقع'**
  String get text_197;

  /// No description provided for @text_198.
  ///
  /// In ar, this message translates to:
  /// **'العودة إلى انواكشوط'**
  String get text_198;

  /// No description provided for @text_199.
  ///
  /// In ar, this message translates to:
  /// **'مزاد لايف'**
  String get text_199;

  /// No description provided for @text_200.
  ///
  /// In ar, this message translates to:
  /// **'عرض مزيد من المزادات'**
  String get text_200;

  /// No description provided for @text_201.
  ///
  /// In ar, this message translates to:
  /// **'الرعاة'**
  String get text_201;

  /// No description provided for @text_202.
  ///
  /// In ar, this message translates to:
  /// **'إعلان جديد'**
  String get text_202;

  /// No description provided for @text_203.
  ///
  /// In ar, this message translates to:
  /// **'كيف يمكنني دفع من من أي تطبيق بنكي'**
  String get text_203;

  /// No description provided for @text_204.
  ///
  /// In ar, this message translates to:
  /// **'كيفية المزايدة على السيارات'**
  String get text_204;

  /// No description provided for @text_205.
  ///
  /// In ar, this message translates to:
  /// **'طريقة استلام سيارتك بعد الفوز'**
  String get text_205;

  /// No description provided for @text_206.
  ///
  /// In ar, this message translates to:
  /// **'شرح نظام العمولات والشحن'**
  String get text_206;

  /// No description provided for @text_207.
  ///
  /// In ar, this message translates to:
  /// **'الأسئلة الشائعة'**
  String get text_207;

  /// No description provided for @text_208.
  ///
  /// In ar, this message translates to:
  /// **'كيفية مزايدة والشحن'**
  String get text_208;

  /// No description provided for @text_209.
  ///
  /// In ar, this message translates to:
  /// **'فيديوهات تعلمية'**
  String get text_209;

  /// No description provided for @text_210.
  ///
  /// In ar, this message translates to:
  /// **'اختر اللغة'**
  String get text_210;

  /// No description provided for @text_211.
  ///
  /// In ar, this message translates to:
  /// **'التالي'**
  String get text_211;

  /// No description provided for @text_212.
  ///
  /// In ar, this message translates to:
  /// **'يرجى التحقق من البيانات'**
  String get text_212;

  /// No description provided for @text_213.
  ///
  /// In ar, this message translates to:
  /// **'أدخل رقم هاتفك'**
  String get text_213;

  /// No description provided for @text_214.
  ///
  /// In ar, this message translates to:
  /// **'كلمة السر'**
  String get text_214;

  /// No description provided for @text_215.
  ///
  /// In ar, this message translates to:
  /// **'هل نسيت كلمة المرور؟'**
  String get text_215;

  /// No description provided for @text_216.
  ///
  /// In ar, this message translates to:
  /// **'مستخدم جديد؟ سجل الآن!'**
  String get text_216;

  /// No description provided for @text_217.
  ///
  /// In ar, this message translates to:
  /// **'تسجيل الدخول'**
  String get text_217;

  /// No description provided for @text_218.
  ///
  /// In ar, this message translates to:
  /// **'اتصل بنا'**
  String get text_218;

  /// No description provided for @text_219.
  ///
  /// In ar, this message translates to:
  /// **'اختر الدولة'**
  String get text_219;

  /// No description provided for @text_220.
  ///
  /// In ar, this message translates to:
  /// **'موريتانيا'**
  String get text_220;

  /// No description provided for @text_221.
  ///
  /// In ar, this message translates to:
  /// **'السنغال'**
  String get text_221;

  /// No description provided for @text_222.
  ///
  /// In ar, this message translates to:
  /// **'المغرب'**
  String get text_222;

  /// No description provided for @text_223.
  ///
  /// In ar, this message translates to:
  /// **'تونس'**
  String get text_223;

  /// No description provided for @text_224.
  ///
  /// In ar, this message translates to:
  /// **'تواصل معنا'**
  String get text_224;

  /// No description provided for @text_225.
  ///
  /// In ar, this message translates to:
  /// **'واتس اب'**
  String get text_225;

  /// No description provided for @text_226.
  ///
  /// In ar, this message translates to:
  /// **'أول تطبيق مزاد في موريتانيا'**
  String get text_226;

  /// No description provided for @text_227.
  ///
  /// In ar, this message translates to:
  /// **'اكتشف تجربة جديدة وفريدة في عالم المزادات الموريتاني.'**
  String get text_227;

  /// No description provided for @text_228.
  ///
  /// In ar, this message translates to:
  /// **'تواصل مباشر'**
  String get text_228;

  /// No description provided for @text_229.
  ///
  /// In ar, this message translates to:
  /// **'اتصل أو راسل المعلن مباشرة عبر التطبيق وبسهولة تامة.'**
  String get text_229;

  /// No description provided for @text_230.
  ///
  /// In ar, this message translates to:
  /// **'صاحب الصفقة الرابحة'**
  String get text_230;

  /// No description provided for @text_231.
  ///
  /// In ar, this message translates to:
  /// **'ادخل عالم المزايدات، وارفع عرضك بثقة لتكون صاحب الصفقة.'**
  String get text_231;

  /// No description provided for @text_232.
  ///
  /// In ar, this message translates to:
  /// **'ابدأ الآن'**
  String get text_232;

  /// No description provided for @text_233.
  ///
  /// In ar, this message translates to:
  /// **'تخطي'**
  String get text_233;

  /// No description provided for @text_234.
  ///
  /// In ar, this message translates to:
  /// **'مزايدة فائزة'**
  String get text_234;

  /// No description provided for @text_235.
  ///
  /// In ar, this message translates to:
  /// **'سعر أعلى منك'**
  String get text_235;

  /// No description provided for @text_236.
  ///
  /// In ar, this message translates to:
  /// **'العناصر التي فزت بها'**
  String get text_236;

  /// No description provided for @text_237.
  ///
  /// In ar, this message translates to:
  /// **'مدفوع'**
  String get text_237;

  /// No description provided for @text_238.
  ///
  /// In ar, this message translates to:
  /// **'في انتظار الدفع'**
  String get text_238;

  /// No description provided for @text_239.
  ///
  /// In ar, this message translates to:
  /// **'عرض التفاصيل'**
  String get text_239;

  /// No description provided for @text_240.
  ///
  /// In ar, this message translates to:
  /// **'ادفع الان'**
  String get text_240;

  /// No description provided for @text_241.
  ///
  /// In ar, this message translates to:
  /// **'هل تحتاج مساعدة؟'**
  String get text_241;

  /// No description provided for @text_242.
  ///
  /// In ar, this message translates to:
  /// **'تواصل مع الدعم'**
  String get text_242;

  /// No description provided for @text_243.
  ///
  /// In ar, this message translates to:
  /// **'كلمة المرور الجديدة'**
  String get text_243;

  /// No description provided for @text_244.
  ///
  /// In ar, this message translates to:
  /// **'قم بتأكيد كلمة المرور الجديدة'**
  String get text_244;

  /// No description provided for @text_245.
  ///
  /// In ar, this message translates to:
  /// **'تحديد ككل كمقروء'**
  String get text_245;

  /// No description provided for @text_246.
  ///
  /// In ar, this message translates to:
  /// **'منذ 5 دقائق'**
  String get text_246;

  /// No description provided for @text_247.
  ///
  /// In ar, this message translates to:
  /// **'تنبيه مزايدة'**
  String get text_247;

  /// No description provided for @text_248.
  ///
  /// In ar, this message translates to:
  /// **'لقد تم تجاوز عرضك في مزاد Toyota Corolla.'**
  String get text_248;

  /// No description provided for @text_249.
  ///
  /// In ar, this message translates to:
  /// **'مبروك الفوز!'**
  String get text_249;

  /// No description provided for @text_250.
  ///
  /// In ar, this message translates to:
  /// **'لقد فزت بمزاد iPhone 15 Pro Max.'**
  String get text_250;

  /// No description provided for @text_251.
  ///
  /// In ar, this message translates to:
  /// **'تأكيد الدفع'**
  String get text_251;

  /// No description provided for @text_252.
  ///
  /// In ar, this message translates to:
  /// **'تم استلام دفعة مبلغ التأمين بنجاح.'**
  String get text_252;

  /// No description provided for @text_253.
  ///
  /// In ar, this message translates to:
  /// **'تحديث النظام'**
  String get text_253;

  /// No description provided for @text_254.
  ///
  /// In ar, this message translates to:
  /// **'هناك ميزات جديدة متاحة في التطبيق الآن.'**
  String get text_254;

  /// No description provided for @text_255.
  ///
  /// In ar, this message translates to:
  /// **'تحديث الإعلان'**
  String get text_255;

  /// No description provided for @text_256.
  ///
  /// In ar, this message translates to:
  /// **'تمت الموافقة على إعلانك ونشره في التطبيق.'**
  String get text_256;

  /// No description provided for @text_257.
  ///
  /// In ar, this message translates to:
  /// **'  أول تطبيق مزاد  في موريتانيا'**
  String get text_257;

  /// No description provided for @text_258.
  ///
  /// In ar, this message translates to:
  /// **'اتصل أو راسل المعلن مباشرة عبر التطبيق'**
  String get text_258;

  /// No description provided for @text_259.
  ///
  /// In ar, this message translates to:
  /// **' اطلب أي خدمة من التطبيق توصلك عند باب دارك'**
  String get text_259;

  /// No description provided for @text_260.
  ///
  /// In ar, this message translates to:
  /// **'أدخل رمز التحقق'**
  String get text_260;

  /// No description provided for @text_261.
  ///
  /// In ar, this message translates to:
  /// **'ادخل الرمز المكون من 6 أرقام المرسل عبر الواتساب إلى'**
  String get text_261;

  /// No description provided for @text_262.
  ///
  /// In ar, this message translates to:
  /// **'إعادة إرسال الرمز خلال '**
  String get text_262;

  /// No description provided for @text_263.
  ///
  /// In ar, this message translates to:
  /// **'تحقق'**
  String get text_263;

  /// No description provided for @text_264.
  ///
  /// In ar, this message translates to:
  /// **'تغيير طريقة التحقق'**
  String get text_264;

  /// No description provided for @text_265.
  ///
  /// In ar, this message translates to:
  /// **'يمكن لهذا التطبيق الوصول إلى الصور التي تختارها فقط.'**
  String get text_265;

  /// No description provided for @text_266.
  ///
  /// In ar, this message translates to:
  /// **'الصور'**
  String get text_266;

  /// No description provided for @text_267.
  ///
  /// In ar, this message translates to:
  /// **'الألبومات'**
  String get text_267;

  /// No description provided for @text_268.
  ///
  /// In ar, this message translates to:
  /// **'استخدم رمز التاجر لإتمام عملية الدفع'**
  String get text_268;

  /// No description provided for @text_269.
  ///
  /// In ar, this message translates to:
  /// **'التاريخ والوقت'**
  String get text_269;

  /// No description provided for @text_270.
  ///
  /// In ar, this message translates to:
  /// **'2026/04/12 م'**
  String get text_270;

  /// No description provided for @text_271.
  ///
  /// In ar, this message translates to:
  /// **'الحالة'**
  String get text_271;

  /// No description provided for @text_272.
  ///
  /// In ar, this message translates to:
  /// **'رسوم إنشاء الطلب'**
  String get text_272;

  /// No description provided for @text_273.
  ///
  /// In ar, this message translates to:
  /// **'رقم هاتف المدفوع'**
  String get text_273;

  /// No description provided for @text_274.
  ///
  /// In ar, this message translates to:
  /// **'يرجى تحويل مبلغ التأمين وارفاق صورة الحوالة في الاسفل'**
  String get text_274;

  /// No description provided for @text_275.
  ///
  /// In ar, this message translates to:
  /// **'تم إرفاق صورة الوصل'**
  String get text_275;

  /// No description provided for @text_276.
  ///
  /// In ar, this message translates to:
  /// **'يرجى الضغط لإرفاق صورة الحوالة'**
  String get text_276;

  /// No description provided for @text_277.
  ///
  /// In ar, this message translates to:
  /// **'يرجى إرفاق صورة الوصل أولاً'**
  String get text_277;

  /// No description provided for @text_278.
  ///
  /// In ar, this message translates to:
  /// **'إتمام الدفع'**
  String get text_278;

  /// No description provided for @text_279.
  ///
  /// In ar, this message translates to:
  /// **'تم استلام طلبك بنجاح'**
  String get text_279;

  /// No description provided for @text_280.
  ///
  /// In ar, this message translates to:
  /// **'قيد المراجعة'**
  String get text_280;

  /// No description provided for @text_281.
  ///
  /// In ar, this message translates to:
  /// **'سيتم مراجعة طلبك من قبل الإدارة في أقرب وقت ممكن وسيتم إضافة الرصيد إلى حسابك.'**
  String get text_281;

  /// No description provided for @text_282.
  ///
  /// In ar, this message translates to:
  /// **'العودة للرئيسية'**
  String get text_282;

  /// No description provided for @text_283.
  ///
  /// In ar, this message translates to:
  /// **'يرجى التسجيل برقم جوالك لاستخدام هذه الميزة.'**
  String get text_283;

  /// No description provided for @text_284.
  ///
  /// In ar, this message translates to:
  /// **'سياسة الخصوصية'**
  String get text_284;

  /// No description provided for @text_285.
  ///
  /// In ar, this message translates to:
  /// **'1. المقدمة'**
  String get text_285;

  /// No description provided for @text_286.
  ///
  /// In ar, this message translates to:
  /// **'نحن في مزاد باي (MazadPay) نلتزم بحماية خصوصيتك ومعلوماتك الشخصية. توضح هذه السياسة كيفية جمع واستخدام وحماية بياناتك عند استخدام تطبيقنا خدماتنا.'**
  String get text_286;

  /// No description provided for @text_287.
  ///
  /// In ar, this message translates to:
  /// **'2. المعلومات التي نجمعها'**
  String get text_287;

  /// No description provided for @text_288.
  ///
  /// In ar, this message translates to:
  /// **'نقوم بجمع المعلومات التي تقدمها لنا مباشرة عند إنشاء حساب، مثل الاسم، رقم الهاتف، والبريد الإلكتروني. كما نسجل بيانات المعاملات المالية والمزايدات التي تشارك بها.'**
  String get text_288;

  /// No description provided for @text_289.
  ///
  /// In ar, this message translates to:
  /// **'3. كيفية استخدام البيانات'**
  String get text_289;

  /// No description provided for @text_290.
  ///
  /// In ar, this message translates to:
  /// **'نستخدم معلوماتك لمعالجة المزايدات، إدارة حسابك، إرسال تنبيهات السعر، وضمان أمان المعاملات المالية داخل التطبيق.'**
  String get text_290;

  /// No description provided for @text_291.
  ///
  /// In ar, this message translates to:
  /// **'4. حماية المعلومات'**
  String get text_291;

  /// No description provided for @text_292.
  ///
  /// In ar, this message translates to:
  /// **'نحن نستخدم تقنيات تشفير متطورة لحماية بياناتك من الوصول غير المصرح به. معلوماتك المالية يتم معالجتها من خلال بوابات دفع آمنة ومعتمدة.'**
  String get text_292;

  /// No description provided for @text_293.
  ///
  /// In ar, this message translates to:
  /// **'5. التغييرات على هذه السياسة'**
  String get text_293;

  /// No description provided for @text_294.
  ///
  /// In ar, this message translates to:
  /// **'نحتفظ بالحق في تحديث سياسة الخصوصية هذه من وقت لآخر. سيتم إخطارك بأي تغييرات جوهرية عبر البريد الإلكتروني أو تنبيه داخل التطبيق.'**
  String get text_294;

  /// No description provided for @text_295.
  ///
  /// In ar, this message translates to:
  /// **'آخر تحديث: أبريل 2026'**
  String get text_295;

  /// No description provided for @text_296.
  ///
  /// In ar, this message translates to:
  /// **'اربح وقتك'**
  String get text_296;

  /// No description provided for @text_297.
  ///
  /// In ar, this message translates to:
  /// **'الخدمات'**
  String get text_297;

  /// No description provided for @text_298.
  ///
  /// In ar, this message translates to:
  /// **'كورس'**
  String get text_298;

  /// No description provided for @text_299.
  ///
  /// In ar, this message translates to:
  /// **'كورس عبر المدن'**
  String get text_299;

  /// No description provided for @text_300.
  ///
  /// In ar, this message translates to:
  /// **'نقل البضائع'**
  String get text_300;

  /// No description provided for @text_301.
  ///
  /// In ar, this message translates to:
  /// **'اخرى'**
  String get text_301;

  /// No description provided for @text_302.
  ///
  /// In ar, this message translates to:
  /// **'تم فتح حساب بنجاح'**
  String get text_302;

  /// No description provided for @text_303.
  ///
  /// In ar, this message translates to:
  /// **'رجوع'**
  String get text_303;

  /// No description provided for @text_304.
  ///
  /// In ar, this message translates to:
  /// **'زايد الآن لتكون الفائز الأقرب بالصفقة!'**
  String get text_304;

  /// No description provided for @text_305.
  ///
  /// In ar, this message translates to:
  /// **'انضم إلى آلاف المزايدين واقتنص الفرص الحصرية بلمسة واحدة.'**
  String get text_305;

  /// No description provided for @text_306.
  ///
  /// In ar, this message translates to:
  /// **'ابدأ المزايدة الآن'**
  String get text_306;

  /// No description provided for @text_307.
  ///
  /// In ar, this message translates to:
  /// **'تخطي '**
  String get text_307;

  /// No description provided for @text_308.
  ///
  /// In ar, this message translates to:
  /// **'تخطي للجولة'**
  String get text_308;

  /// No description provided for @text_309.
  ///
  /// In ar, this message translates to:
  /// **'مرکز الدعم'**
  String get text_309;

  /// No description provided for @text_310.
  ///
  /// In ar, this message translates to:
  /// **'كيف يمكننا مساعدتك؟'**
  String get text_310;

  /// No description provided for @text_311.
  ///
  /// In ar, this message translates to:
  /// **'فريقنا متاح لمساعدتك في أي وقت.'**
  String get text_311;

  /// No description provided for @text_312.
  ///
  /// In ar, this message translates to:
  /// **'تحدث معنا على واتساب'**
  String get text_312;

  /// No description provided for @text_313.
  ///
  /// In ar, this message translates to:
  /// **'استجابة سريعة ومباشرة'**
  String get text_313;

  /// No description provided for @text_314.
  ///
  /// In ar, this message translates to:
  /// **'اتصال هاتفي'**
  String get text_314;

  /// No description provided for @text_315.
  ///
  /// In ar, this message translates to:
  /// **'كيف أبدأ المزايدة؟'**
  String get text_315;

  /// No description provided for @text_316.
  ///
  /// In ar, this message translates to:
  /// **'كيف يمكنني استرجاع مبلغ التأمين؟'**
  String get text_316;

  /// No description provided for @text_317.
  ///
  /// In ar, this message translates to:
  /// **'ما هي طرق الدفع المتاحة؟'**
  String get text_317;

  /// No description provided for @text_318.
  ///
  /// In ar, this message translates to:
  /// **'شروط الاستخدام'**
  String get text_318;

  /// No description provided for @text_319.
  ///
  /// In ar, this message translates to:
  /// **'السابق'**
  String get text_319;

  /// No description provided for @text_320.
  ///
  /// In ar, this message translates to:
  /// **'أوافق على الشروط'**
  String get text_320;

  /// No description provided for @text_321.
  ///
  /// In ar, this message translates to:
  /// **'آخر تحديث: أبريل 2026'**
  String get text_321;

  /// No description provided for @text_322.
  ///
  /// In ar, this message translates to:
  /// **'شروط استخدام تطبيق \"مزاد موريتانيا\"'**
  String get text_322;

  /// No description provided for @text_323.
  ///
  /// In ar, this message translates to:
  /// **'الموافقة على الشروط'**
  String get text_323;

  /// No description provided for @text_324.
  ///
  /// In ar, this message translates to:
  /// **'باستخدامك التطبيق، فإنك توافق على الالتزام بهذه الشروط والقوانين المعمول بها في موريتانيا.'**
  String get text_324;

  /// No description provided for @text_325.
  ///
  /// In ar, this message translates to:
  /// **'الأهلية'**
  String get text_325;

  /// No description provided for @text_326.
  ///
  /// In ar, this message translates to:
  /// **'يحق فقط للأشخاص الذين بلغوا 18 سنة أو أكثر المشاركة في المزادات.'**
  String get text_326;

  /// No description provided for @text_327.
  ///
  /// In ar, this message translates to:
  /// **'الحساب والأمان'**
  String get text_327;

  /// No description provided for @text_328.
  ///
  /// In ar, this message translates to:
  /// **'يجب تسجيل الحساب باستخدام رقم هاتف صحيح وفعال.'**
  String get text_328;

  /// No description provided for @text_329.
  ///
  /// In ar, this message translates to:
  /// **'أنت مسؤول عن سرية كلمة المرور وجميع الأنشطة التي تتم بحسابك.'**
  String get text_329;

  /// No description provided for @text_330.
  ///
  /// In ar, this message translates to:
  /// **'التعديلات والإشعارات:'**
  String get text_330;

  /// No description provided for @text_331.
  ///
  /// In ar, this message translates to:
  /// **'يحق لنا تعديل الشروط أو إضافة مزايا جديدة في أي وقت.'**
  String get text_331;

  /// No description provided for @text_332.
  ///
  /// In ar, this message translates to:
  /// **'أي تغييرات مهمة سنرسل إشعاراً للمستخدمين.'**
  String get text_332;

  /// No description provided for @text_333.
  ///
  /// In ar, this message translates to:
  /// **'1. إنهاء الحساب:'**
  String get text_333;

  /// No description provided for @text_334.
  ///
  /// In ar, this message translates to:
  /// **'نحتفظ بحق تعليق أو حذف أي حساب ينتهك الشروط أو يضر بمستخدمي التطبيق.'**
  String get text_334;

  /// No description provided for @text_335.
  ///
  /// In ar, this message translates to:
  /// **'2. القانون الواجب التطبيق:'**
  String get text_335;

  /// No description provided for @text_336.
  ///
  /// In ar, this message translates to:
  /// **'تخضع هذه الشروط لقوانين الجمهورية الإسلامية الموريتانية، وأي نزاع يتم حله وفقاً لها.'**
  String get text_336;

  /// No description provided for @text_337.
  ///
  /// In ar, this message translates to:
  /// **'شروط المشاركة في المزادات:'**
  String get text_337;

  /// No description provided for @text_338.
  ///
  /// In ar, this message translates to:
  /// **'يتطلب بعض المزادات دفع مبلغ تأمين لضمان جدية المزايدة.'**
  String get text_338;

  /// No description provided for @text_339.
  ///
  /// In ar, this message translates to:
  /// **'يُسترجع مبلغ التأمين تلقائيًا في حال عدم الفوز بالمزاد خلال ساعة من انتهاء المزاد.'**
  String get text_339;

  /// No description provided for @text_340.
  ///
  /// In ar, this message translates to:
  /// **'لا يُسترجع مبلغ التأمين في حال الفوز بالمزاد وعدم إتمام عملية الدفع أو عند مخالفة شروط التطبيق.'**
  String get text_340;

  /// No description provided for @text_341.
  ///
  /// In ar, this message translates to:
  /// **'يحق لإدارة التطبيق إلغاء أي مزاد في حال وجود خطأ تقني أو شبهة احتيال، ويتم في هذه الحالة إعادة مبلغ التأمين للمشاركين.'**
  String get text_341;

  /// No description provided for @text_342.
  ///
  /// In ar, this message translates to:
  /// **'يشترط للاشتراك في التطبيق دفع رسوم اشتراك سنوية قدرها 100 أوقية جديدة للاستفادة من خدمات المزاد والمشاركة فيه'**
  String get text_342;

  /// No description provided for @text_343.
  ///
  /// In ar, this message translates to:
  /// **'استرجاع مبلغ التأمين'**
  String get text_343;

  /// No description provided for @text_344.
  ///
  /// In ar, this message translates to:
  /// **'مبلغ الاسترجاع'**
  String get text_344;

  /// No description provided for @text_345.
  ///
  /// In ar, this message translates to:
  /// **'طريقة الاستلام'**
  String get text_345;

  /// No description provided for @text_346.
  ///
  /// In ar, this message translates to:
  /// **'حوالة بنكية'**
  String get text_346;

  /// No description provided for @text_347.
  ///
  /// In ar, this message translates to:
  /// **'خدمة موبايل (Bankily/Masrivi)'**
  String get text_347;

  /// No description provided for @text_348.
  ///
  /// In ar, this message translates to:
  /// **'تأكيد طلب الاسترجاع'**
  String get text_348;

  /// No description provided for @text_349.
  ///
  /// In ar, this message translates to:
  /// **'الرصيد القابل للاسترجاع'**
  String get text_349;

  /// No description provided for @text_350.
  ///
  /// In ar, this message translates to:
  /// **'تم استلام طلبك!'**
  String get text_350;

  /// No description provided for @text_351.
  ///
  /// In ar, this message translates to:
  /// **'سيتم معالجة طلب استرجاع مبلغ التأمين خلال 24 ساعة.'**
  String get text_351;

  /// No description provided for @text_352.
  ///
  /// In ar, this message translates to:
  /// **'سيارة نظيفة جدا، صيانة دورية، محرك ممتاز.'**
  String get text_352;

  /// No description provided for @text_353.
  ///
  /// In ar, this message translates to:
  /// **'تويوتا'**
  String get text_353;

  /// No description provided for @text_354.
  ///
  /// In ar, this message translates to:
  /// **'بنزين'**
  String get text_354;

  /// No description provided for @text_355.
  ///
  /// In ar, this message translates to:
  /// **'أوتوماتيكي'**
  String get text_355;

  /// No description provided for @text_356.
  ///
  /// In ar, this message translates to:
  /// **'تويوتا برادو'**
  String get text_356;

  /// No description provided for @text_357.
  ///
  /// In ar, this message translates to:
  /// **'محمد احمد'**
  String get text_357;

  /// No description provided for @text_358.
  ///
  /// In ar, this message translates to:
  /// **'مريم سيدي'**
  String get text_358;

  /// No description provided for @text_359.
  ///
  /// In ar, this message translates to:
  /// **'علي صمبا'**
  String get text_359;

  /// No description provided for @text_360.
  ///
  /// In ar, this message translates to:
  /// **'قيم تجربتك'**
  String get text_360;

  /// No description provided for @text_361.
  ///
  /// In ar, this message translates to:
  /// **'رأيك يهمنا لتحسين جولة خدماتنا وإسعادكم!'**
  String get text_361;

  /// No description provided for @text_362.
  ///
  /// In ar, this message translates to:
  /// **'اكتب ملاحظاتك hier (اختياري)...'**
  String get text_362;

  /// No description provided for @text_363.
  ///
  /// In ar, this message translates to:
  /// **'إرسال التقييم'**
  String get text_363;

  /// No description provided for @text_364.
  ///
  /// In ar, this message translates to:
  /// **'انتهى المزاد'**
  String get text_364;

  /// No description provided for @text_365.
  ///
  /// In ar, this message translates to:
  /// **'ألف مبروك!'**
  String get text_365;

  /// No description provided for @text_366.
  ///
  /// In ar, this message translates to:
  /// **'انتهى المزاد وانت المزايد الأعلى'**
  String get text_366;

  /// No description provided for @text_367.
  ///
  /// In ar, this message translates to:
  /// **'إستكمال إجراءات الدفع والشراء'**
  String get text_367;

  /// No description provided for @text_368.
  ///
  /// In ar, this message translates to:
  /// **'تاكيد المزايدة'**
  String get text_368;

  /// No description provided for @text_369.
  ///
  /// In ar, this message translates to:
  /// **'مزايدات'**
  String get text_369;

  /// No description provided for @text_370.
  ///
  /// In ar, this message translates to:
  /// **'أعلى مزايدة'**
  String get text_370;

  /// No description provided for @text_371.
  ///
  /// In ar, this message translates to:
  /// **'الوقت المتبقي'**
  String get text_371;

  /// No description provided for @text_372.
  ///
  /// In ar, this message translates to:
  /// **'الموافقة على مبلغ المزايدة ؟'**
  String get text_372;

  /// No description provided for @text_373.
  ///
  /// In ar, this message translates to:
  /// **'تمت المزايدة بنجاح !'**
  String get text_373;

  /// No description provided for @text_374.
  ///
  /// In ar, this message translates to:
  /// **'تاكيد'**
  String get text_374;

  /// No description provided for @text_375.
  ///
  /// In ar, this message translates to:
  /// **'مزاد '**
  String get text_375;

  /// No description provided for @text_376.
  ///
  /// In ar, this message translates to:
  /// **'باي'**
  String get text_376;

  /// No description provided for @text_377.
  ///
  /// In ar, this message translates to:
  /// **'الفيديوهات'**
  String get text_377;

  /// No description provided for @text_378.
  ///
  /// In ar, this message translates to:
  /// **'إضافة'**
  String get text_378;

  /// No description provided for @text_379.
  ///
  /// In ar, this message translates to:
  /// **'قم بالإيداع الآن'**
  String get text_379;

  /// No description provided for @text_380.
  ///
  /// In ar, this message translates to:
  /// **'للتواصل معنا'**
  String get text_380;

  /// No description provided for @text_381.
  ///
  /// In ar, this message translates to:
  /// **'المعلومات الشخصية'**
  String get text_381;

  /// No description provided for @text_382.
  ///
  /// In ar, this message translates to:
  /// **'تغيير اللغة'**
  String get text_382;

  /// No description provided for @text_383.
  ///
  /// In ar, this message translates to:
  /// **'قيم التطبيق'**
  String get text_383;

  /// No description provided for @text_384.
  ///
  /// In ar, this message translates to:
  /// **'شاركتنا رأيك'**
  String get text_384;

  /// No description provided for @text_385.
  ///
  /// In ar, this message translates to:
  /// **'الشروط والأحكام'**
  String get text_385;

  /// No description provided for @text_386.
  ///
  /// In ar, this message translates to:
  /// **'مساعدة/الأسئلة الشائعة'**
  String get text_386;

  /// No description provided for @text_387.
  ///
  /// In ar, this message translates to:
  /// **'عن مزاد أبي'**
  String get text_387;

  /// No description provided for @text_388.
  ///
  /// In ar, this message translates to:
  /// **'مشاركة'**
  String get text_388;

  /// No description provided for @text_389.
  ///
  /// In ar, this message translates to:
  /// **'هل تعرف شخصًا مهتمًا؟'**
  String get text_389;

  /// No description provided for @text_390.
  ///
  /// In ar, this message translates to:
  /// **'حول مزاد أبي'**
  String get text_390;

  /// No description provided for @error_connection.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في الاتصال'**
  String get error_connection;

  /// No description provided for @error_login_failed.
  ///
  /// In ar, this message translates to:
  /// **'فشل تسجيل الدخول'**
  String get error_login_failed;

  /// No description provided for @error_invalid_credentials.
  ///
  /// In ar, this message translates to:
  /// **'بيانات الاعتماد غير صالحة'**
  String get error_invalid_credentials;

  /// No description provided for @error_phone_required.
  ///
  /// In ar, this message translates to:
  /// **'رقم الهاتف مطلوب'**
  String get error_phone_required;

  /// No description provided for @error_otp_required.
  ///
  /// In ar, this message translates to:
  /// **'رمز OTP مطلوب'**
  String get error_otp_required;

  /// No description provided for @error_otp_invalid.
  ///
  /// In ar, this message translates to:
  /// **'رمز OTP غير صالح'**
  String get error_otp_invalid;

  /// No description provided for @error_loading_auctions.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل المزادات'**
  String get error_loading_auctions;

  /// No description provided for @error_loading_auction_details.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل تفاصيل المزاد'**
  String get error_loading_auction_details;

  /// No description provided for @error_loading_favorites.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل المفضلة'**
  String get error_loading_favorites;

  /// No description provided for @error_loading_balance.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل الرصيد'**
  String get error_loading_balance;

  /// No description provided for @error_deposit_failed.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في الإيداع'**
  String get error_deposit_failed;

  /// No description provided for @error_withdraw_failed.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في السحب'**
  String get error_withdraw_failed;

  /// No description provided for @error_insufficient_balance.
  ///
  /// In ar, this message translates to:
  /// **'رصيد غير كافٍ'**
  String get error_insufficient_balance;

  /// No description provided for @error_loading_notifications.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل الإشعارات'**
  String get error_loading_notifications;

  /// No description provided for @error_loading_profile.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل الملف الشخصي'**
  String get error_loading_profile;

  /// No description provided for @error_create_auction.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في إنشاء الإعلان'**
  String get error_create_auction;

  /// No description provided for @error_fill_required_fields.
  ///
  /// In ar, this message translates to:
  /// **'يرجى ملء جميع الحقول المطلوبة'**
  String get error_fill_required_fields;

  /// No description provided for @error_add_image.
  ///
  /// In ar, this message translates to:
  /// **'يرجى إضافة صورة واحدة على الأقل'**
  String get error_add_image;

  /// No description provided for @error_invalid_amount.
  ///
  /// In ar, this message translates to:
  /// **'يرجى إدخال مبلغ صحيح'**
  String get error_invalid_amount;

  /// No description provided for @error_no_data.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد بيانات متاحة'**
  String get error_no_data;

  /// No description provided for @text_391.
  ///
  /// In ar, this message translates to:
  /// **'لديك حساب بالفعل؟'**
  String get text_391;

  /// No description provided for @favorites_synced.
  ///
  /// In ar, this message translates to:
  /// **'تم مزامنة المفضلة'**
  String get favorites_synced;

  /// No description provided for @favorites_local_storage.
  ///
  /// In ar, this message translates to:
  /// **'المفضلة محفوظة على الجهاز'**
  String get favorites_local_storage;

  /// No description provided for @favorites_sync_error.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في مزامنة المفضلة'**
  String get favorites_sync_error;

  /// No description provided for @retry.
  ///
  /// In ar, this message translates to:
  /// **'إعادة المحاولة'**
  String get retry;

  /// No description provided for @sync.
  ///
  /// In ar, this message translates to:
  /// **'مزامنة'**
  String get sync;

  /// No description provided for @profile_created.
  ///
  /// In ar, this message translates to:
  /// **'تم إنشاء الملف الشخصي بنجاح'**
  String get profile_created;

  /// No description provided for @error_password_mismatch.
  ///
  /// In ar, this message translates to:
  /// **'كلمات المرور غير متطابقة'**
  String get error_password_mismatch;

  /// No description provided for @error_password_too_short.
  ///
  /// In ar, this message translates to:
  /// **'يجب أن تكون كلمة المرور 4 أحرف على الأقل'**
  String get error_password_too_short;

  /// No description provided for @error_register.
  ///
  /// In ar, this message translates to:
  /// **'خطأ أثناء التسجيل'**
  String get error_register;

  /// No description provided for @auction_created.
  ///
  /// In ar, this message translates to:
  /// **'تم إنشاء المزاد بنجاح'**
  String get auction_created;

  /// No description provided for @bid_placed.
  ///
  /// In ar, this message translates to:
  /// **'تم تقديم المزايدة بنجاح'**
  String get bid_placed;

  /// No description provided for @bid_error.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تقديم المزايدة'**
  String get bid_error;

  /// No description provided for @deposit_success.
  ///
  /// In ar, this message translates to:
  /// **'تم الإيداع بنجاح'**
  String get deposit_success;

  /// No description provided for @withdraw_success.
  ///
  /// In ar, this message translates to:
  /// **'تم السحب بنجاح'**
  String get withdraw_success;

  /// No description provided for @no_auctions_available.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد مزادات متاحة حالياً'**
  String get no_auctions_available;

  /// No description provided for @no_notifications.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد إشعارات'**
  String get no_notifications;

  /// No description provided for @no_favorites.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد عناصر في المفضلة'**
  String get no_favorites;

  /// No description provided for @no_winnings.
  ///
  /// In ar, this message translates to:
  /// **'لم تربح أي مزادات بعد'**
  String get no_winnings;

  /// No description provided for @no_my_auctions.
  ///
  /// In ar, this message translates to:
  /// **'لم تقم بإنشاء أي مزادات'**
  String get no_my_auctions;

  /// No description provided for @no_banners.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد إعلانات متاحة'**
  String get no_banners;

  /// No description provided for @no_data.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد بيانات متاحة'**
  String get no_data;

  /// No description provided for @no_title.
  ///
  /// In ar, this message translates to:
  /// **'بدون عنوان'**
  String get no_title;

  /// No description provided for @auction.
  ///
  /// In ar, this message translates to:
  /// **'مزاد'**
  String get auction;

  /// No description provided for @data_not_available_offline.
  ///
  /// In ar, this message translates to:
  /// **'البيانات غير متوفرة في وضع عدم الاتصال'**
  String get data_not_available_offline;

  /// No description provided for @error_generic.
  ///
  /// In ar, this message translates to:
  /// **'حدث خطأ: {error}'**
  String error_generic(Object error);

  /// No description provided for @error_loading_winnings.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل المزادات المربوحة'**
  String get error_loading_winnings;

  /// No description provided for @error_loading_my_auctions.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل مزاداتي'**
  String get error_loading_my_auctions;

  /// No description provided for @error_sync_favorites.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في مزامنة المفضلة'**
  String get error_sync_favorites;

  /// No description provided for @error_upload_image.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في رفع الصورة'**
  String get error_upload_image;

  /// No description provided for @error_network.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في الاتصال بالشبكة'**
  String get error_network;

  /// No description provided for @error_server.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في الخادم'**
  String get error_server;

  /// No description provided for @error_timeout.
  ///
  /// In ar, this message translates to:
  /// **'انتهت مهلة الاتصال'**
  String get error_timeout;

  /// No description provided for @error_loading_auction.
  ///
  /// In ar, this message translates to:
  /// **'خطأ في تحميل المزاد'**
  String get error_loading_auction;

  /// No description provided for @no_related_auctions.
  ///
  /// In ar, this message translates to:
  /// **'لا توجد مزادات مشابهة متاحة'**
  String get no_related_auctions;

  /// No description provided for @loading.
  ///
  /// In ar, this message translates to:
  /// **'جاري التحميل...'**
  String get loading;

  /// No description provided for @pull_to_refresh.
  ///
  /// In ar, this message translates to:
  /// **'اسحب للتحديث'**
  String get pull_to_refresh;

  /// No description provided for @view_more.
  ///
  /// In ar, this message translates to:
  /// **'عرض المزيد'**
  String get view_more;

  /// No description provided for @show_less.
  ///
  /// In ar, this message translates to:
  /// **'عرض أقل'**
  String get show_less;

  /// No description provided for @try_again.
  ///
  /// In ar, this message translates to:
  /// **'حاول مرة أخرى'**
  String get try_again;

  /// No description provided for @cancel.
  ///
  /// In ar, this message translates to:
  /// **'إلغاء'**
  String get cancel;

  /// No description provided for @confirm.
  ///
  /// In ar, this message translates to:
  /// **'تأكيد'**
  String get confirm;

  /// No description provided for @save.
  ///
  /// In ar, this message translates to:
  /// **'حفظ'**
  String get save;

  /// No description provided for @delete.
  ///
  /// In ar, this message translates to:
  /// **'حذف'**
  String get delete;

  /// No description provided for @edit.
  ///
  /// In ar, this message translates to:
  /// **'تعديل'**
  String get edit;

  /// No description provided for @close.
  ///
  /// In ar, this message translates to:
  /// **'إغلاق'**
  String get close;

  /// No description provided for @success.
  ///
  /// In ar, this message translates to:
  /// **'نجاح'**
  String get success;

  /// No description provided for @error.
  ///
  /// In ar, this message translates to:
  /// **'خطأ'**
  String get error;

  /// No description provided for @warning.
  ///
  /// In ar, this message translates to:
  /// **'تحذير'**
  String get warning;

  /// No description provided for @info.
  ///
  /// In ar, this message translates to:
  /// **'معلومة'**
  String get info;

  /// No description provided for @page_not_available.
  ///
  /// In ar, this message translates to:
  /// **'هذه الصفحة غير متوفرة حالياً'**
  String get page_not_available;

  /// No description provided for @return_to_first_city.
  ///
  /// In ar, this message translates to:
  /// **'العودة إلى المدينة الأولى'**
  String get return_to_first_city;

  /// No description provided for @return_button.
  ///
  /// In ar, this message translates to:
  /// **'عودة'**
  String get return_button;

  /// No description provided for @views.
  ///
  /// In ar, this message translates to:
  /// **'مشاهدة'**
  String get views;

  /// No description provided for @lot.
  ///
  /// In ar, this message translates to:
  /// **'رقم اللوط'**
  String get lot;

  /// No description provided for @day.
  ///
  /// In ar, this message translates to:
  /// **'يوم'**
  String get day;

  /// No description provided for @text_392.
  ///
  /// In ar, this message translates to:
  /// **'رسائلي'**
  String get text_392;

  /// No description provided for @text_393.
  ///
  /// In ar, this message translates to:
  /// **'مزايدة تلقائية'**
  String get text_393;

  /// No description provided for @text_394.
  ///
  /// In ar, this message translates to:
  /// **'تمييز المزاد'**
  String get text_394;

  /// No description provided for @text_395.
  ///
  /// In ar, this message translates to:
  /// **'بلاغ عن المزاد'**
  String get text_395;

  /// No description provided for @text_396.
  ///
  /// In ar, this message translates to:
  /// **'إعادة نشر المزاد'**
  String get text_396;

  /// No description provided for @text_397.
  ///
  /// In ar, this message translates to:
  /// **'إلغاء المزاد'**
  String get text_397;

  /// No description provided for @text_398.
  ///
  /// In ar, this message translates to:
  /// **'تمديد المزاد'**
  String get text_398;

  /// No description provided for @text_399.
  ///
  /// In ar, this message translates to:
  /// **'إضافة صور'**
  String get text_399;

  /// No description provided for @text_400.
  ///
  /// In ar, this message translates to:
  /// **'تواصل مع البائع'**
  String get text_400;

  /// No description provided for @text_401.
  ///
  /// In ar, this message translates to:
  /// **'طلباتي'**
  String get text_401;

  /// No description provided for @text_402.
  ///
  /// In ar, this message translates to:
  /// **'محادثة جديدة'**
  String get text_402;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['ar', 'en', 'fr'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'ar':
      return AppLocalizationsAr();
    case 'en':
      return AppLocalizationsEn();
    case 'fr':
      return AppLocalizationsFr();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
