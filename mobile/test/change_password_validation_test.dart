import 'package:flutter_test/flutter_test.dart';

// Customer request #6A: Change Password in the Account page. Audit found
// this already fully implemented and wired (mobile/lib/pages/
// account_profile_page.dart's _ChangePasswordSheet -> authNotifierProvider.
// changePassword -> AuthApi.changePassword -> PUT /auth/change-password,
// protected by jwtMiddleware, backend/internal/handlers/auth_handler.go
// ChangePasswordRequest: old_pin required (no length rule, accepts a legacy
// 4-digit PIN), new_pin required+min=8+max=72). No code was changed for
// #6A -- these tests cover the existing validation predicates
// (_ChangePasswordSheet's TextFormField validators), reproduced here rather
// than pumping the real widget, since _ChangePasswordSheet is a private
// class inside a page that constructs UserApi()/Dio/dotenv in its own
// initState -- the same coupling concern already documented in
// password_mismatch_test.dart for the registration/reset password forms.
//
// current password validator: required only (matches backend: old_pin has
// no min-length rule, so a legacy 4-digit PIN must be accepted as "current").
String? currentPasswordValidator(String? v) {
  if (v == null || v.isEmpty) return 'Required';
  return null;
}

// new password validator: required + min 8 chars (matches backend: new_pin
// validate:"required,min=8,max=72").
String? newPasswordValidator(String? v) {
  if (v == null || v.isEmpty) return 'Required';
  if (v.length < 8) return 'At least 8 characters';
  return null;
}

// confirm password validator: required + must match the new password.
String? confirmPasswordValidator(String? v, String newPassword) {
  if (v == null || v.isEmpty) return 'Required';
  if (v != newPassword) return 'Passwords do not match';
  return null;
}

void main() {
  group('current password validator (accepts legacy 4-digit PIN, backend has no min-length rule)', () {
    test('empty current password is rejected', () {
      expect(currentPasswordValidator(''), isNotNull);
    });

    test('a 4-digit legacy PIN as current password is accepted client-side (backend decides validity)', () {
      expect(currentPasswordValidator('1234'), isNull);
    });

    test('an 8+ char current password is accepted', () {
      expect(currentPasswordValidator('oldpassword123'), isNull);
    });
  });

  group('new password validator (matches backend new_pin: required, min 8)', () {
    test('empty new password is rejected', () {
      expect(newPasswordValidator(''), isNotNull);
    });

    test('a new password under 8 characters is rejected', () {
      expect(newPasswordValidator('short1'), isNotNull);
    });

    test('a new password of exactly 8 characters is accepted', () {
      expect(newPasswordValidator('exactly8'), isNull);
    });

    test('a long new password is accepted', () {
      expect(newPasswordValidator('a very long new password indeed'), isNull);
    });
  });

  group('confirm password validator', () {
    test('empty confirmation is rejected', () {
      expect(confirmPasswordValidator('', 'newpassword123'), isNotNull);
    });

    test('mismatched confirmation is rejected', () {
      expect(confirmPasswordValidator('different123', 'newpassword123'), isNotNull);
    });

    test('matching confirmation is accepted', () {
      expect(confirmPasswordValidator('newpassword123', 'newpassword123'), isNull);
    });
  });
}
