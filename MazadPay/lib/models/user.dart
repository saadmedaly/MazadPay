/// Modèle utilisateur basé sur le backend Go
/// Correspond à `backend/internal/models/user.go`
class User {
  final String id;
  final String phone;
  final String? fullName;
  final String? email;
  final String? profilePicUrl;
  final String? city;
  final String languagePref;
  final bool notificationsEnabled;
  final DateTime? termsAcceptedAt;
  final bool isActive;
  final String role;
  final bool isVerified;
  final DateTime? lastLoginAt;
  final DateTime createdAt;
  final DateTime updatedAt;
  final String? countryCode;
  final DateTime? dateOfBirth;
  final String? address;
  final String? postalCode;
  final String? gender;
  final bool profileCompleted;
  final String? kycStatus;

  User({
    required this.id,
    required this.phone,
    this.fullName,
    this.email,
    this.profilePicUrl,
    this.city,
    this.languagePref = 'ar',
    this.notificationsEnabled = true,
    this.termsAcceptedAt,
    this.isActive = true,
    this.role = 'user',
    this.isVerified = false,
    this.lastLoginAt,
    required this.createdAt,
    required this.updatedAt,
    this.countryCode,
    this.dateOfBirth,
    this.address,
    this.postalCode,
    this.gender,
    this.profileCompleted = false,
    this.kycStatus,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? '',
      phone: json['phone'] ?? '',
      fullName: json['full_name'],
      email: json['email'],
      profilePicUrl: json['profile_pic_url'],
      city: json['city'],
      languagePref: json['language_pref'] ?? 'ar',
      notificationsEnabled: json['notifications_enabled'] ?? true,
      termsAcceptedAt: json['terms_accepted_at'] != null
          ? DateTime.parse(json['terms_accepted_at'])
          : null,
      isActive: json['is_active'] ?? true,
      role: json['role'] ?? 'user',
      isVerified: json['is_verified'] ?? false,
      lastLoginAt: json['last_login_at'] != null
          ? DateTime.parse(json['last_login_at'])
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
      countryCode: json['country_code'],
      dateOfBirth: json['date_of_birth'] != null
          ? DateTime.parse(json['date_of_birth'])
          : null,
      address: json['address'],
      postalCode: json['postal_code'],
      gender: json['gender'],
      profileCompleted: json['profile_completed'] ?? false,
      kycStatus: json['kyc_status'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'phone': phone,
      'full_name': fullName,
      'email': email,
      'profile_pic_url': profilePicUrl,
      'city': city,
      'language_pref': languagePref,
      'notifications_enabled': notificationsEnabled,
      'terms_accepted_at': termsAcceptedAt?.toIso8601String(),
      'is_active': isActive,
      'role': role,
      'is_verified': isVerified,
      'last_login_at': lastLoginAt?.toIso8601String(),
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
      'country_code': countryCode,
      'date_of_birth': dateOfBirth?.toIso8601String(),
      'address': address,
      'postal_code': postalCode,
      'gender': gender,
      'profile_completed': profileCompleted,
      'kyc_status': kycStatus,
    };
  }

  /// Masque le numéro de téléphone (####xxxx)
  String get maskedPhone {
    if (phone.length < 4) return '####';
    return '####${phone.substring(phone.length - 4)}';
  }

  /// Vérifie si l'utilisateur est admin
  bool get isAdmin => role == 'admin' || role == 'super_admin';

  /// Vérifie si l'utilisateur est super admin
  bool get isSuperAdmin => role == 'super_admin';

  /// Copie l'utilisateur avec des modifications
  User copyWith({
    String? id,
    String? phone,
    String? fullName,
    String? email,
    String? profilePicUrl,
    String? city,
    String? languagePref,
    bool? notificationsEnabled,
    DateTime? termsAcceptedAt,
    bool? isActive,
    String? role,
    bool? isVerified,
    DateTime? lastLoginAt,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? countryCode,
    DateTime? dateOfBirth,
    String? address,
    String? postalCode,
    String? gender,
    bool? profileCompleted,
    String? kycStatus,
  }) {
    return User(
      id: id ?? this.id,
      phone: phone ?? this.phone,
      fullName: fullName ?? this.fullName,
      email: email ?? this.email,
      profilePicUrl: profilePicUrl ?? this.profilePicUrl,
      city: city ?? this.city,
      languagePref: languagePref ?? this.languagePref,
      notificationsEnabled: notificationsEnabled ?? this.notificationsEnabled,
      termsAcceptedAt: termsAcceptedAt ?? this.termsAcceptedAt,
      isActive: isActive ?? this.isActive,
      role: role ?? this.role,
      isVerified: isVerified ?? this.isVerified,
      lastLoginAt: lastLoginAt ?? this.lastLoginAt,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      countryCode: countryCode ?? this.countryCode,
      dateOfBirth: dateOfBirth ?? this.dateOfBirth,
      address: address ?? this.address,
      postalCode: postalCode ?? this.postalCode,
      gender: gender ?? this.gender,
      profileCompleted: profileCompleted ?? this.profileCompleted,
      kycStatus: kycStatus ?? this.kycStatus,
    );
  }
}

/// Statuts KYC possibles
class KycStatus {
  static const String pending = 'pending';
  static const String verified = 'verified';
  static const String rejected = 'rejected';
  static const String notSubmitted = 'not_submitted';
}

/// Rôles utilisateur
class UserRole {
  static const String user = 'user';
  static const String admin = 'admin';
  static const String superAdmin = 'super_admin';
}
