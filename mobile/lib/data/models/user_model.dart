// lib/data/models/user_model.dart
// Modèle utilisateur avec JSON serialization

import 'package:freezed_annotation/freezed_annotation.dart';

part 'user_model.freezed.dart';
part 'user_model.g.dart';

@freezed
class UserModel with _$UserModel {
  const factory UserModel({
    required String id,
    required String phone,
    String? fullName,
    String? email,
    String? profilePicUrl,
    String? city,
    String? countryCode,
    String? address,
    String? postalCode,
    String? dateOfBirth,
    String? gender,
    @Default('ar') String languagePref,
    @Default(true) bool notificationsEnabled,
    @Default(true) bool isActive,
    @Default('user') String role,
    @Default(false) bool isVerified,
    @Default(false) bool profileCompleted,
    String? lastLoginAt,
    String? createdAt,
  }) = _UserModel;
  
  factory UserModel.fromJson(Map<String, dynamic> json) => 
      _$UserModelFromJson(json);
}

// Extension pour formater le téléphone masqué
extension UserModelExtension on UserModel {
  String get maskedPhone {
    if (phone.length < 4) return phone;
    return '####${phone.substring(phone.length - 4)}';
  }
  
  String get displayName => fullName ?? phone;
  
  String get initials {
    if (fullName != null && fullName!.isNotEmpty) {
      return fullName!.split(' ').map((e) => e[0]).take(2).join().toUpperCase();
    }
    return phone.substring(0, 2).toUpperCase();
  }
}
