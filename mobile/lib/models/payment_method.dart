// Model for payment method fetched from backend
class PaymentMethod {
  final int id;
  final String code;
  final String nameAr;
  final String nameFr;
  final String? nameEn;
  final String? logoUrl;
  final bool isActive;
  final int? countryId;

  PaymentMethod({
    required this.id,
    required this.code,
    required this.nameAr,
    required this.nameFr,
    this.nameEn,
    this.logoUrl,
    required this.isActive,
    this.countryId,
  });

  factory PaymentMethod.fromJson(Map<String, dynamic> json) {
    return PaymentMethod(
      id: json['id'] as int,
      code: json['code'] as String,
      nameAr: json['name_ar'] as String,
      nameFr: json['name_fr'] as String,
      nameEn: json['name_en'] as String?,
      logoUrl: json['logo_url'] as String?,
      isActive: json['is_active'] as bool? ?? true,
      countryId: json['country_id'] as int?,
    );
  }
}
