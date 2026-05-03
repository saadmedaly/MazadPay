// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$userNotifierHash() => r'737fa0a9f764b6183914fcf8fc4a86581fce3519';

/// Provider pour la gestion du profil utilisateur
///
/// Copied from [UserNotifier].
@ProviderFor(UserNotifier)
final userNotifierProvider =
    AutoDisposeNotifierProvider<UserNotifier, UserState>.internal(
      UserNotifier.new,
      name: r'userNotifierProvider',
      debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
          ? null
          : _$userNotifierHash,
      dependencies: null,
      allTransitiveDependencies: null,
    );

typedef _$UserNotifier = AutoDisposeNotifier<UserState>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
