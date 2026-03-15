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

  /// No description provided for @appName.
  ///
  /// In en, this message translates to:
  /// **'MazadPay'**
  String get appName;

  /// No description provided for @chooseLanguage.
  ///
  /// In en, this message translates to:
  /// **'Choose Language'**
  String get chooseLanguage;

  /// No description provided for @next.
  ///
  /// In en, this message translates to:
  /// **'Next'**
  String get next;

  /// No description provided for @login.
  ///
  /// In en, this message translates to:
  /// **'Login'**
  String get login;

  /// No description provided for @phoneNumber.
  ///
  /// In en, this message translates to:
  /// **'Phone Number'**
  String get phoneNumber;

  /// No description provided for @enterPhoneNumber.
  ///
  /// In en, this message translates to:
  /// **'Enter your phone number'**
  String get enterPhoneNumber;

  /// No description provided for @password.
  ///
  /// In en, this message translates to:
  /// **'Password'**
  String get password;

  /// No description provided for @enterPassword.
  ///
  /// In en, this message translates to:
  /// **'Enter your password'**
  String get enterPassword;

  /// No description provided for @forgotPassword.
  ///
  /// In en, this message translates to:
  /// **'Forgot Password?'**
  String get forgotPassword;

  /// No description provided for @newUser.
  ///
  /// In en, this message translates to:
  /// **'New User? Register Now!'**
  String get newUser;

  /// No description provided for @passwordRecoveryDevelopment.
  ///
  /// In en, this message translates to:
  /// **'Password recovery is under development'**
  String get passwordRecoveryDevelopment;

  /// No description provided for @contactUs.
  ///
  /// In en, this message translates to:
  /// **'Contact Us'**
  String get contactUs;

  /// No description provided for @selectCountry.
  ///
  /// In en, this message translates to:
  /// **'Select Country'**
  String get selectCountry;

  /// No description provided for @mauritania.
  ///
  /// In en, this message translates to:
  /// **'Mauritania'**
  String get mauritania;

  /// No description provided for @senegal.
  ///
  /// In en, this message translates to:
  /// **'Senegal'**
  String get senegal;

  /// No description provided for @morocco.
  ///
  /// In en, this message translates to:
  /// **'Morocco'**
  String get morocco;

  /// No description provided for @tunisia.
  ///
  /// In en, this message translates to:
  /// **'Tunisia'**
  String get tunisia;

  /// No description provided for @comingSoon.
  ///
  /// In en, this message translates to:
  /// **'Coming Soon'**
  String get comingSoon;

  /// No description provided for @whatsapp.
  ///
  /// In en, this message translates to:
  /// **'WhatsApp'**
  String get whatsapp;

  /// No description provided for @email.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get email;

  /// No description provided for @onboarding1.
  ///
  /// In en, this message translates to:
  /// **'First auction app in Mauritania'**
  String get onboarding1;

  /// No description provided for @onboarding2.
  ///
  /// In en, this message translates to:
  /// **'Call or message the advertiser directly through the app'**
  String get onboarding2;

  /// No description provided for @onboarding3.
  ///
  /// In en, this message translates to:
  /// **'Enter the world of bidding and place your bid with confidence'**
  String get onboarding3;

  /// No description provided for @startNow.
  ///
  /// In en, this message translates to:
  /// **'Start Now'**
  String get startNow;

  /// No description provided for @continueText.
  ///
  /// In en, this message translates to:
  /// **'Continue'**
  String get continueText;

  /// No description provided for @skip.
  ///
  /// In en, this message translates to:
  /// **'Skip'**
  String get skip;

  /// No description provided for @locationPermissionTitle.
  ///
  /// In en, this message translates to:
  /// **'Mazad Mauritania needs your location'**
  String get locationPermissionTitle;

  /// No description provided for @locationPermissionDesc.
  ///
  /// In en, this message translates to:
  /// **'For a better experience, please enable geolocation'**
  String get locationPermissionDesc;

  /// No description provided for @enableLocation.
  ///
  /// In en, this message translates to:
  /// **'Enable location service'**
  String get enableLocation;

  /// No description provided for @earnTime.
  ///
  /// In en, this message translates to:
  /// **'Save your time'**
  String get earnTime;

  /// No description provided for @nouakchott.
  ///
  /// In en, this message translates to:
  /// **'Nouakchott'**
  String get nouakchott;

  /// No description provided for @nouadhibou.
  ///
  /// In en, this message translates to:
  /// **'Nouadhibou'**
  String get nouadhibou;

  /// No description provided for @liveAuctions.
  ///
  /// In en, this message translates to:
  /// **'Live Auctions'**
  String get liveAuctions;

  /// No description provided for @viewMoreAuctions.
  ///
  /// In en, this message translates to:
  /// **'View more auctions'**
  String get viewMoreAuctions;

  /// No description provided for @brands.
  ///
  /// In en, this message translates to:
  /// **'Brands'**
  String get brands;

  /// No description provided for @home.
  ///
  /// In en, this message translates to:
  /// **'Home'**
  String get home;

  /// No description provided for @delivery.
  ///
  /// In en, this message translates to:
  /// **'Delivery'**
  String get delivery;

  /// No description provided for @ecommerce.
  ///
  /// In en, this message translates to:
  /// **'E-commerce'**
  String get ecommerce;

  /// No description provided for @myAccount.
  ///
  /// In en, this message translates to:
  /// **'My Account'**
  String get myAccount;

  /// No description provided for @depositNow.
  ///
  /// In en, this message translates to:
  /// **'Deposit Now'**
  String get depositNow;

  /// No description provided for @startBiddingJourney.
  ///
  /// In en, this message translates to:
  /// **'Start your bidding journey!'**
  String get startBiddingJourney;

  /// No description provided for @personalInfo.
  ///
  /// In en, this message translates to:
  /// **'Personal Information'**
  String get personalInfo;

  /// No description provided for @changeLanguage.
  ///
  /// In en, this message translates to:
  /// **'Change Language'**
  String get changeLanguage;

  /// No description provided for @contactUsSide.
  ///
  /// In en, this message translates to:
  /// **'Contact Us'**
  String get contactUsSide;

  /// No description provided for @rateApp.
  ///
  /// In en, this message translates to:
  /// **'Rate App'**
  String get rateApp;

  /// No description provided for @shareOpinion.
  ///
  /// In en, this message translates to:
  /// **'Share your opinion'**
  String get shareOpinion;

  /// No description provided for @termsConditions.
  ///
  /// In en, this message translates to:
  /// **'Terms & Conditions'**
  String get termsConditions;

  /// No description provided for @helpFaq.
  ///
  /// In en, this message translates to:
  /// **'Help/FAQ'**
  String get helpFaq;

  /// No description provided for @aboutMazad.
  ///
  /// In en, this message translates to:
  /// **'About Mazad Mauritania'**
  String get aboutMazad;

  /// No description provided for @shareApp.
  ///
  /// In en, this message translates to:
  /// **'Share App'**
  String get shareApp;

  /// No description provided for @shareAppDesc.
  ///
  /// In en, this message translates to:
  /// **'Do you know someone interested in bidding services?'**
  String get shareAppDesc;

  /// No description provided for @insuranceAmount.
  ///
  /// In en, this message translates to:
  /// **'Insurance Amount'**
  String get insuranceAmount;

  /// No description provided for @availableBalance.
  ///
  /// In en, this message translates to:
  /// **'Available Balance'**
  String get availableBalance;

  /// No description provided for @chooseMethod.
  ///
  /// In en, this message translates to:
  /// **'Choose a method'**
  String get chooseMethod;

  /// No description provided for @retrieveInsurance.
  ///
  /// In en, this message translates to:
  /// **'Retrieve insurance amount'**
  String get retrieveInsurance;

  /// No description provided for @yourActivities.
  ///
  /// In en, this message translates to:
  /// **'Your Activities'**
  String get yourActivities;

  /// No description provided for @myAuctions.
  ///
  /// In en, this message translates to:
  /// **'My Auctions'**
  String get myAuctions;

  /// No description provided for @favorites.
  ///
  /// In en, this message translates to:
  /// **'Favorites'**
  String get favorites;

  /// No description provided for @wonItems.
  ///
  /// In en, this message translates to:
  /// **'Won Items'**
  String get wonItems;

  /// No description provided for @communicationCenter.
  ///
  /// In en, this message translates to:
  /// **'Communication Center'**
  String get communicationCenter;

  /// No description provided for @personalAccountInfo.
  ///
  /// In en, this message translates to:
  /// **'Personal Account Info'**
  String get personalAccountInfo;

  /// No description provided for @services.
  ///
  /// In en, this message translates to:
  /// **'Services'**
  String get services;

  /// No description provided for @deliveryService.
  ///
  /// In en, this message translates to:
  /// **'Delivery'**
  String get deliveryService;

  /// No description provided for @course.
  ///
  /// In en, this message translates to:
  /// **'Course'**
  String get course;

  /// No description provided for @intercityCourse.
  ///
  /// In en, this message translates to:
  /// **'Intercity Course'**
  String get intercityCourse;

  /// No description provided for @goodsTransport.
  ///
  /// In en, this message translates to:
  /// **'Goods Transport'**
  String get goodsTransport;

  /// No description provided for @other.
  ///
  /// In en, this message translates to:
  /// **'Other'**
  String get other;

  /// No description provided for @termsOfUse.
  ///
  /// In en, this message translates to:
  /// **'Terms of Use'**
  String get termsOfUse;

  /// No description provided for @previous.
  ///
  /// In en, this message translates to:
  /// **'Previous'**
  String get previous;

  /// No description provided for @iAgreeToTerms.
  ///
  /// In en, this message translates to:
  /// **'I Agree to Terms'**
  String get iAgreeToTerms;

  /// No description provided for @lastUpdate.
  ///
  /// In en, this message translates to:
  /// **'Last Update: June 2024'**
  String get lastUpdate;

  /// No description provided for @termsOfUseTitle.
  ///
  /// In en, this message translates to:
  /// **'Terms of Use for \"Mazad Mauritania\" App'**
  String get termsOfUseTitle;

  /// No description provided for @agreeToTermsHeader.
  ///
  /// In en, this message translates to:
  /// **'Agreement to Terms'**
  String get agreeToTermsHeader;

  /// No description provided for @agreeToTermsBody.
  ///
  /// In en, this message translates to:
  /// **'By using the app, you agree to comply with these terms and the laws applicable in Mauritania.'**
  String get agreeToTermsBody;

  /// No description provided for @eligibilityHeader.
  ///
  /// In en, this message translates to:
  /// **'Eligibility'**
  String get eligibilityHeader;

  /// No description provided for @eligibilityBody.
  ///
  /// In en, this message translates to:
  /// **'Only individuals aged 18 or older are entitled to participate in auctions.'**
  String get eligibilityBody;

  /// No description provided for @accountSecurityHeader.
  ///
  /// In en, this message translates to:
  /// **'Account and Security'**
  String get accountSecurityHeader;

  /// No description provided for @accountSecurityBody1.
  ///
  /// In en, this message translates to:
  /// **'The account must be registered using a valid and active phone number.'**
  String get accountSecurityBody1;

  /// No description provided for @accountSecurityBody2.
  ///
  /// In en, this message translates to:
  /// **'You are responsible for password confidentiality and all activities occurring in your account.'**
  String get accountSecurityBody2;

  /// No description provided for @modificationsHeader.
  ///
  /// In en, this message translates to:
  /// **'Modifications and Notices:'**
  String get modificationsHeader;

  /// No description provided for @modificationsBullet1.
  ///
  /// In en, this message translates to:
  /// **'We reserve the right to modify the terms or add new features at any time.'**
  String get modificationsBullet1;

  /// No description provided for @modificationsBullet2.
  ///
  /// In en, this message translates to:
  /// **'Any important changes will be notified to users.'**
  String get modificationsBullet2;

  /// No description provided for @terminationHeader.
  ///
  /// In en, this message translates to:
  /// **'1. Account Termination:'**
  String get terminationHeader;

  /// No description provided for @terminationBullet1.
  ///
  /// In en, this message translates to:
  /// **'We reserve the right to suspend or delete any account that violates the terms or harms app users.'**
  String get terminationBullet1;

  /// No description provided for @lawHeader.
  ///
  /// In en, this message translates to:
  /// **'2. Applicable Law:'**
  String get lawHeader;

  /// No description provided for @lawBullet1.
  ///
  /// In en, this message translates to:
  /// **'These terms are subject to the laws of the Islamic Republic of Mauritania, and any dispute shall be resolved accordingly.'**
  String get lawBullet1;

  /// No description provided for @participationTermsHeader.
  ///
  /// In en, this message translates to:
  /// **'Auction Participation Terms:'**
  String get participationTermsHeader;

  /// No description provided for @participationBullet1.
  ///
  /// In en, this message translates to:
  /// **'Some auctions require payment of an insurance amount to ensure bidding seriousness.'**
  String get participationBullet1;

  /// No description provided for @participationBullet2.
  ///
  /// In en, this message translates to:
  /// **'The insurance amount is automatically returned if the auction is not won within an hour of its end.'**
  String get participationBullet2;

  /// No description provided for @participationBullet3.
  ///
  /// In en, this message translates to:
  /// **'The insurance amount is not returned in case of winning and failing to complete the payment or violating app terms.'**
  String get participationBullet3;

  /// No description provided for @participationBullet4.
  ///
  /// In en, this message translates to:
  /// **'App management has the right to cancel any auction in case of a technical error or suspicion of fraud, in which case the insurance amount is returned to participants.'**
  String get participationBullet4;

  /// No description provided for @participationBullet5.
  ///
  /// In en, this message translates to:
  /// **'Subscription to the app requires payment of an annual subscription fee of 100 MRU to benefit from and participate in auction services.'**
  String get participationBullet5;

  /// No description provided for @payThroughApps.
  ///
  /// In en, this message translates to:
  /// **'Pay through banking apps'**
  String get payThroughApps;

  /// No description provided for @readTermsCarefully.
  ///
  /// In en, this message translates to:
  /// **'Please read the following terms and conditions for transactions carefully'**
  String get readTermsCarefully;

  /// No description provided for @paymentAndInsuranceTerms.
  ///
  /// In en, this message translates to:
  /// **'Payment and Insurance Terms'**
  String get paymentAndInsuranceTerms;

  /// No description provided for @depositTerm1.
  ///
  /// In en, this message translates to:
  /// **'1. The app accepts payment methods: Bankily, Masrvi, Sedad.'**
  String get depositTerm1;

  /// No description provided for @depositTerm2.
  ///
  /// In en, this message translates to:
  /// **'2. Some auctions require payment of an insurance amount to ensure bidding seriousness.'**
  String get depositTerm2;

  /// No description provided for @depositTerm3.
  ///
  /// In en, this message translates to:
  /// **'3. The insurance amount is automatically returned if the auction is not won within an hour depending on the payment provider.'**
  String get depositTerm3;

  /// No description provided for @depositTerm4.
  ///
  /// In en, this message translates to:
  /// **'4. Insurance is not returned in case of winning and failing to complete the payment or violating app terms.'**
  String get depositTerm4;

  /// No description provided for @depositTerm5.
  ///
  /// In en, this message translates to:
  /// **'5. App management has the right to cancel any auction in case of a technical error or suspicion of fraud, with insurance returned upon cancellation.'**
  String get depositTerm5;

  /// No description provided for @usageAgreement.
  ///
  /// In en, this message translates to:
  /// **'By using the app, you agree to these terms'**
  String get usageAgreement;

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @agreeToTermsBrief.
  ///
  /// In en, this message translates to:
  /// **'I agree to terms'**
  String get agreeToTermsBrief;

  /// No description provided for @paymentMethod.
  ///
  /// In en, this message translates to:
  /// **'Payment Method'**
  String get paymentMethod;

  /// No description provided for @addDeposit.
  ///
  /// In en, this message translates to:
  /// **'Add Deposit'**
  String get addDeposit;

  /// No description provided for @payVia.
  ///
  /// In en, this message translates to:
  /// **'Pay via {method}'**
  String payVia(String method);

  /// No description provided for @onboardingTitle.
  ///
  /// In en, this message translates to:
  /// **'Bid Now to Be the Global Winner!'**
  String get onboardingTitle;

  /// No description provided for @onboardingDesc.
  ///
  /// In en, this message translates to:
  /// **'Join thousands of bidders and seize exclusive opportunities with one touch.'**
  String get onboardingDesc;

  /// No description provided for @skipToAuction.
  ///
  /// In en, this message translates to:
  /// **'Skip to Auction'**
  String get skipToAuction;

  /// No description provided for @appAccessPhotos.
  ///
  /// In en, this message translates to:
  /// **'Selection of photos is only accessible to this application.'**
  String get appAccessPhotos;

  /// No description provided for @photos.
  ///
  /// In en, this message translates to:
  /// **'Photos'**
  String get photos;

  /// No description provided for @albums.
  ///
  /// In en, this message translates to:
  /// **'Albums'**
  String get albums;

  /// No description provided for @useMerchantCode.
  ///
  /// In en, this message translates to:
  /// **'Use merchant code to complete payment'**
  String get useMerchantCode;

  /// No description provided for @amountToMazad.
  ///
  /// In en, this message translates to:
  /// **'100 MRU'**
  String get amountToMazad;

  /// No description provided for @orderCreationTime.
  ///
  /// In en, this message translates to:
  /// **'Date and Time'**
  String get orderCreationTime;

  /// No description provided for @status.
  ///
  /// In en, this message translates to:
  /// **'Status'**
  String get status;

  /// No description provided for @pendingPayment.
  ///
  /// In en, this message translates to:
  /// **'Pending Payment'**
  String get pendingPayment;

  /// No description provided for @orderCreationFee.
  ///
  /// In en, this message translates to:
  /// **'Order Creation Fee'**
  String get orderCreationFee;

  /// No description provided for @payerPhoneNumber.
  ///
  /// In en, this message translates to:
  /// **'Payer Phone Number'**
  String get payerPhoneNumber;

  /// No description provided for @totalAmountToPay.
  ///
  /// In en, this message translates to:
  /// **'Total amount to pay via {method}'**
  String totalAmountToPay(String method);

  /// No description provided for @uploadReceiptPrompt.
  ///
  /// In en, this message translates to:
  /// **'Please transfer the insurance amount and attach the receipt below'**
  String get uploadReceiptPrompt;

  /// No description provided for @clickToUploadReceipt.
  ///
  /// In en, this message translates to:
  /// **'Please click to attach the transfer receipt'**
  String get clickToUploadReceipt;

  /// No description provided for @receiptAttached.
  ///
  /// In en, this message translates to:
  /// **'Receipt attached'**
  String get receiptAttached;

  /// No description provided for @completePayment.
  ///
  /// In en, this message translates to:
  /// **'Complete Payment'**
  String get completePayment;

  /// No description provided for @attachReceiptFirst.
  ///
  /// In en, this message translates to:
  /// **'Please attach receipt first'**
  String get attachReceiptFirst;

  /// No description provided for @masrvi.
  ///
  /// In en, this message translates to:
  /// **'Masrvi'**
  String get masrvi;

  /// No description provided for @bankily.
  ///
  /// In en, this message translates to:
  /// **'Bankily'**
  String get bankily;

  /// No description provided for @sedad.
  ///
  /// In en, this message translates to:
  /// **'Sedad'**
  String get sedad;

  /// No description provided for @click.
  ///
  /// In en, this message translates to:
  /// **'Click'**
  String get click;

  /// No description provided for @masrviDesc.
  ///
  /// In en, this message translates to:
  /// **'Use our merchant code to complete the payment via Masrvi'**
  String get masrviDesc;

  /// No description provided for @bankilyDesc.
  ///
  /// In en, this message translates to:
  /// **'Use our merchant code to complete the payment via Bankily'**
  String get bankilyDesc;

  /// No description provided for @sedadDesc.
  ///
  /// In en, this message translates to:
  /// **'Use our merchant code to complete the payment via Sedad'**
  String get sedadDesc;

  /// No description provided for @clickDesc.
  ///
  /// In en, this message translates to:
  /// **'Use our merchant code to complete the payment via Click'**
  String get clickDesc;

  /// No description provided for @auctionSmaah.
  ///
  /// In en, this message translates to:
  /// **'Surface for sale Carrefour'**
  String get auctionSmaah;

  /// No description provided for @auctionIphone.
  ///
  /// In en, this message translates to:
  /// **'iPhone 12 Pro Max'**
  String get auctionIphone;

  /// No description provided for @auctionCorolla.
  ///
  /// In en, this message translates to:
  /// **'Corolla 2017'**
  String get auctionCorolla;

  /// No description provided for @auctionRaf4.
  ///
  /// In en, this message translates to:
  /// **'RAV4 2017'**
  String get auctionRaf4;

  /// No description provided for @auctionHouse.
  ///
  /// In en, this message translates to:
  /// **'House in Rabina'**
  String get auctionHouse;

  /// No description provided for @auctionLaptop.
  ///
  /// In en, this message translates to:
  /// **'Laptop Lenovo'**
  String get auctionLaptop;

  /// No description provided for @orderSuccessTitle.
  ///
  /// In en, this message translates to:
  /// **'Your order has been received successfully'**
  String get orderSuccessTitle;

  /// No description provided for @underReview.
  ///
  /// In en, this message translates to:
  /// **'Under Review'**
  String get underReview;

  /// No description provided for @orderReviewDesc.
  ///
  /// In en, this message translates to:
  /// **'Your request will be reviewed by the management as soon as possible and the balance will be added to your account.'**
  String get orderReviewDesc;

  /// No description provided for @backToHome.
  ///
  /// In en, this message translates to:
  /// **'Back to Home'**
  String get backToHome;

  /// No description provided for @needHelp.
  ///
  /// In en, this message translates to:
  /// **'Need help?'**
  String get needHelp;

  /// No description provided for @contactSupport.
  ///
  /// In en, this message translates to:
  /// **'Contact support'**
  String get contactSupport;

  /// No description provided for @registrationPrompt.
  ///
  /// In en, this message translates to:
  /// **'Please register with your mobile number to use this feature.'**
  String get registrationPrompt;

  /// No description provided for @enterVerificationCode.
  ///
  /// In en, this message translates to:
  /// **'Enter verification code'**
  String get enterVerificationCode;

  /// No description provided for @otpUnderPhone.
  ///
  /// In en, this message translates to:
  /// **'Enter the 6-digit code sent via WhatsApp to'**
  String get otpUnderPhone;

  /// No description provided for @resendCodeIn.
  ///
  /// In en, this message translates to:
  /// **'Resend code in'**
  String get resendCodeIn;

  /// No description provided for @verify.
  ///
  /// In en, this message translates to:
  /// **'Verify'**
  String get verify;

  /// No description provided for @changeVerificationMethod.
  ///
  /// In en, this message translates to:
  /// **'Change verification method'**
  String get changeVerificationMethod;

  /// No description provided for @accountCreatedSuccess.
  ///
  /// In en, this message translates to:
  /// **'Account created successfully'**
  String get accountCreatedSuccess;

  /// No description provided for @back.
  ///
  /// In en, this message translates to:
  /// **'Back'**
  String get back;

  /// No description provided for @newPassword.
  ///
  /// In en, this message translates to:
  /// **'New Password'**
  String get newPassword;

  /// No description provided for @confirmNewPassword.
  ///
  /// In en, this message translates to:
  /// **'Confirm New Password'**
  String get confirmNewPassword;

  /// No description provided for @accountInfo.
  ///
  /// In en, this message translates to:
  /// **'Account Information'**
  String get accountInfo;

  /// No description provided for @myInfo.
  ///
  /// In en, this message translates to:
  /// **'My Information'**
  String get myInfo;

  /// No description provided for @fullName.
  ///
  /// In en, this message translates to:
  /// **'Full Name'**
  String get fullName;

  /// No description provided for @city.
  ///
  /// In en, this message translates to:
  /// **'City'**
  String get city;

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @changePassword.
  ///
  /// In en, this message translates to:
  /// **'Change Password'**
  String get changePassword;

  /// No description provided for @notifications.
  ///
  /// In en, this message translates to:
  /// **'Notifications'**
  String get notifications;

  /// No description provided for @logout.
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get logout;

  /// No description provided for @dummyName.
  ///
  /// In en, this message translates to:
  /// **'Badal Sydia'**
  String get dummyName;

  /// No description provided for @dummyCity.
  ///
  /// In en, this message translates to:
  /// **'Nouakchott'**
  String get dummyCity;

  /// No description provided for @dummyInitial.
  ///
  /// In en, this message translates to:
  /// **'B'**
  String get dummyInitial;

  /// No description provided for @languageName.
  ///
  /// In en, this message translates to:
  /// **'English'**
  String get languageName;

  /// No description provided for @language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;
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
