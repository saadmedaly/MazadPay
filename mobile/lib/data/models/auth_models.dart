// lib/data/models/auth_models.dart
// Modèles pour l'authentification

import 'package:freezed_annotation/freezed_annotation.dart';

part 'auth_models.freezed.dart';
part 'auth_models.g.dart';

@freezed
class LoginRequest with _$LoginRequest {
  const factory LoginRequest({
    required String phone,
    required String pin,
  }) = _LoginRequest;
  
  factory LoginRequest.fromJson(Map<String, dynamic> json) => 
      _$LoginRequestFromJson(json);
}

@freezed
class LoginResponse with _$LoginResponse {
  const factory LoginResponse({
    required String token,
    String? refreshToken,
    required UserData user,
  }) = _LoginResponse;
  
  factory LoginResponse.fromJson(Map<String, dynamic> json) => 
      _$LoginResponseFromJson(json);
}

@freezed
class UserData with _$UserData {
  const factory UserData({
    required String id,
    required String phone,
    String? fullName,
    String? email,
    String? role,
    bool? isVerified,
  }) = _UserData;
  
  factory UserData.fromJson(Map<String, dynamic> json) => 
      _$UserDataFromJson(json);
}

@freezed
class RegisterRequest with _$RegisterRequest {
  const factory RegisterRequest({
    required String phone,
    required String pin,
    required String fullName,
    String? email,
    String? city,
  }) = _RegisterRequest;
  
  factory RegisterRequest.fromJson(Map<String, dynamic> json) => 
      _$RegisterRequestFromJson(json);
}

@freezed
class OtpRequest with _$OtpRequest {
  const factory OtpRequest({
    required String phone,
    required String purpose, // register, reset_password
  }) = _OtpRequest;
  
  factory OtpRequest.fromJson(Map<String, dynamic> json) => 
      _$OtpRequestFromJson(json);
}

@freezed
class OtpVerifyRequest with _$OtpVerifyRequest {
  const factory OtpVerifyRequest({
    required String phone,
    required String code,
    required String purpose,
  }) = _OtpVerifyRequest;
  
  factory OtpVerifyRequest.fromJson(Map<String, dynamic> json) => 
      _$OtpVerifyRequestFromJson(json);
}

@freezed
class ResetPasswordRequest with _$ResetPasswordRequest {
  const factory ResetPasswordRequest({
    required String phone,
    required String newPin,
    required String otpCode,
  }) = _ResetPasswordRequest;
  
  factory ResetPasswordRequest.fromJson(Map<String, dynamic> json) => 
      _$ResetPasswordRequestFromJson(json);
}

@freezed
class ChangePinRequest with _$ChangePinRequest {
  const factory ChangePinRequest({
    required String oldPin,
    required String newPin,
  }) = _ChangePinRequest;
  
  factory ChangePinRequest.fromJson(Map<String, dynamic> json) => 
      _$ChangePinRequestFromJson(json);
}

@freezed
class RefreshTokenRequest with _$RefreshTokenRequest {
  const factory RefreshTokenRequest({
    required String refreshToken,
  }) = _RefreshTokenRequest;
  
  factory RefreshTokenRequest.fromJson(Map<String, dynamic> json) => 
      _$RefreshTokenRequestFromJson(json);
}

// Session utilisateur stockée localement
@freezed
class UserSession with _$UserSession {
  const factory UserSession({
    required String token,
    String? refreshToken,
    required String userId,
    required String phone,
    String? fullName,
    required String loginTime,
    String? expiresAt,
    @Default('ar') String language,
  }) = _UserSession;
  
  factory UserSession.fromJson(Map<String, dynamic> json) => 
      _$UserSessionFromJson(json);
}
