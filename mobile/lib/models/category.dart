/// Modèle Category basé sur le backend Go
/// Correspond à `backend/internal/models/auction.go`
class Category {
  final int id;
  final String nameAr;
  final String nameFr;
  final String nameEn;
  final int? parentId;
  final String? iconName;
  final int displayOrder;
  final bool isActive;
  final String? imageUrl;
  final bool hasSubcategories;

  Category({
    required this.id,
    required this.nameAr,
    required this.nameFr,
    required this.nameEn,
    this.parentId,
    this.iconName,
    this.displayOrder = 0,
    this.isActive = true,
    this.imageUrl,
    this.hasSubcategories = false,
  });

  factory Category.fromJson(Map<String, dynamic> json) {
    return Category(
      id: json['id'] ?? 0,
      nameAr: json['name_ar'] ?? '',
      nameFr: json['name_fr'] ?? '',
      nameEn: json['name_en'] ?? '',
      parentId: json['parent_id'],
      iconName: json['icon_name'],
      displayOrder: json['display_order'] ?? 0,
      isActive: json['is_active'] ?? true,
      imageUrl: json['image_url'],
      hasSubcategories: json['has_subcategories'] ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name_ar': nameAr,
      'name_fr': nameFr,
      'name_en': nameEn,
      'parent_id': parentId,
      'icon_name': iconName,
      'display_order': displayOrder,
      'is_active': isActive,
      'image_url': imageUrl,
      'has_subcategories': hasSubcategories,
    };
  }

  /// Récupère le nom dans la langue spécifiée
  String getName(String languageCode) {
    switch (languageCode) {
      case 'ar':
        return nameAr;
      case 'fr':
        return nameFr;
      case 'en':
        return nameEn;
      default:
        return nameAr;
    }
  }

  /// Vérifie si c'est une catégorie parente
  bool get isParent => parentId == null;

  /// Vérifie si c'est une sous-catégorie
  bool get isSubcategory => parentId != null;
}

/// Modèle Location (Ville/Zone)
class Location {
  final int id;
  final int? countryId;
  final String cityNameAr;
  final String cityNameFr;
  final String areaNameAr;
  final String areaNameFr;
  final Country? country;

  Location({
    required this.id,
    this.countryId,
    required this.cityNameAr,
    required this.cityNameFr,
    required this.areaNameAr,
    required this.areaNameFr,
    this.country,
  });

  factory Location.fromJson(Map<String, dynamic> json) {
    return Location(
      id: json['id'] ?? 0,
      countryId: json['country_id'],
      cityNameAr: json['city_name_ar'] ?? '',
      cityNameFr: json['city_name_fr'] ?? '',
      areaNameAr: json['area_name_ar'] ?? '',
      areaNameFr: json['area_name_fr'] ?? '',
      country: json['country'] != null
          ? Country.fromJson(json['country'])
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'country_id': countryId,
      'city_name_ar': cityNameAr,
      'city_name_fr': cityNameFr,
      'area_name_ar': areaNameAr,
      'area_name_fr': areaNameFr,
      'country': country?.toJson(),
    };
  }

  /// Récupère le nom de la ville dans la langue spécifiée
  String getCityName(String languageCode) {
    switch (languageCode) {
      case 'ar':
        return cityNameAr;
      case 'fr':
        return cityNameFr;
      default:
        return cityNameAr;
    }
  }

  /// Récupère le nom de la zone dans la langue spécifiée
  String getAreaName(String languageCode) {
    switch (languageCode) {
      case 'ar':
        return areaNameAr;
      case 'fr':
        return areaNameFr;
      default:
        return areaNameAr;
    }
  }

  /// Nom complet (ville + zone)
  String getFullName(String languageCode) {
    final city = getCityName(languageCode);
    final area = getAreaName(languageCode);
    if (area.isNotEmpty && area != city) {
      return '$city - $area';
    }
    return city;
  }
}

/// Modèle Country (Pays)
class Country {
  final int id;
  final String code;
  final String nameAr;
  final String nameFr;
  final String nameEn;
  final String flagEmoji;
  final bool isActive;

  Country({
    required this.id,
    required this.code,
    required this.nameAr,
    required this.nameFr,
    required this.nameEn,
    required this.flagEmoji,
    this.isActive = true,
  });

  factory Country.fromJson(Map<String, dynamic> json) {
    return Country(
      id: json['id'] ?? 0,
      code: json['code'] ?? '',
      nameAr: json['name_ar'] ?? '',
      nameFr: json['name_fr'] ?? '',
      nameEn: json['name_en'] ?? '',
      flagEmoji: json['flag_emoji'] ?? '',
      isActive: json['is_active'] ?? true,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'code': code,
      'name_ar': nameAr,
      'name_fr': nameFr,
      'name_en': nameEn,
      'flag_emoji': flagEmoji,
      'is_active': isActive,
    };
  }

  /// Récupère le nom dans la langue spécifiée
  String getName(String languageCode) {
    switch (languageCode) {
      case 'ar':
        return nameAr;
      case 'fr':
        return nameFr;
      case 'en':
        return nameEn;
      default:
        return nameAr;
    }
  }
}
