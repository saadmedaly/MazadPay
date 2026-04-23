// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'favorites_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$favoritesCountHash() => r'641242a3928be9784ffcb3cc73268578057c8e76';

/// Provider pour le nombre de favoris
///
/// Copied from [favoritesCount].
@ProviderFor(favoritesCount)
final favoritesCountProvider = AutoDisposeFutureProvider<int>.internal(
  favoritesCount,
  name: r'favoritesCountProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$favoritesCountHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

@Deprecated('Will be removed in 3.0. Use Ref instead')
// ignore: unused_element
typedef FavoritesCountRef = AutoDisposeFutureProviderRef<int>;
String _$isAuctionFavoriteHash() => r'afa88fcb71627198ea61cdae1cd8de3a2c0937b8';

/// Copied from Dart SDK
class _SystemHash {
  _SystemHash._();

  static int combine(int hash, int value) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + value);
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x0007ffff & hash) << 10));
    return hash ^ (hash >> 6);
  }

  static int finish(int hash) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x03ffffff & hash) << 3));
    // ignore: parameter_assignments
    hash = hash ^ (hash >> 11);
    return 0x1fffffff & (hash + ((0x00003fff & hash) << 15));
  }
}

/// Provider pour vérifier si une enchère spécifique est en favori
///
/// Copied from [isAuctionFavorite].
@ProviderFor(isAuctionFavorite)
const isAuctionFavoriteProvider = IsAuctionFavoriteFamily();

/// Provider pour vérifier si une enchère spécifique est en favori
///
/// Copied from [isAuctionFavorite].
class IsAuctionFavoriteFamily extends Family<bool> {
  /// Provider pour vérifier si une enchère spécifique est en favori
  ///
  /// Copied from [isAuctionFavorite].
  const IsAuctionFavoriteFamily();

  /// Provider pour vérifier si une enchère spécifique est en favori
  ///
  /// Copied from [isAuctionFavorite].
  IsAuctionFavoriteProvider call(String auctionId) {
    return IsAuctionFavoriteProvider(auctionId);
  }

  @override
  IsAuctionFavoriteProvider getProviderOverride(
    covariant IsAuctionFavoriteProvider provider,
  ) {
    return call(provider.auctionId);
  }

  static const Iterable<ProviderOrFamily>? _dependencies = null;

  @override
  Iterable<ProviderOrFamily>? get dependencies => _dependencies;

  static const Iterable<ProviderOrFamily>? _allTransitiveDependencies = null;

  @override
  Iterable<ProviderOrFamily>? get allTransitiveDependencies =>
      _allTransitiveDependencies;

  @override
  String? get name => r'isAuctionFavoriteProvider';
}

/// Provider pour vérifier si une enchère spécifique est en favori
///
/// Copied from [isAuctionFavorite].
class IsAuctionFavoriteProvider extends AutoDisposeProvider<bool> {
  /// Provider pour vérifier si une enchère spécifique est en favori
  ///
  /// Copied from [isAuctionFavorite].
  IsAuctionFavoriteProvider(String auctionId)
    : this._internal(
        (ref) => isAuctionFavorite(ref as IsAuctionFavoriteRef, auctionId),
        from: isAuctionFavoriteProvider,
        name: r'isAuctionFavoriteProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$isAuctionFavoriteHash,
        dependencies: IsAuctionFavoriteFamily._dependencies,
        allTransitiveDependencies:
            IsAuctionFavoriteFamily._allTransitiveDependencies,
        auctionId: auctionId,
      );

  IsAuctionFavoriteProvider._internal(
    super._createNotifier, {
    required super.name,
    required super.dependencies,
    required super.allTransitiveDependencies,
    required super.debugGetCreateSourceHash,
    required super.from,
    required this.auctionId,
  }) : super.internal();

  final String auctionId;

  @override
  Override overrideWith(bool Function(IsAuctionFavoriteRef provider) create) {
    return ProviderOverride(
      origin: this,
      override: IsAuctionFavoriteProvider._internal(
        (ref) => create(ref as IsAuctionFavoriteRef),
        from: from,
        name: null,
        dependencies: null,
        allTransitiveDependencies: null,
        debugGetCreateSourceHash: null,
        auctionId: auctionId,
      ),
    );
  }

  @override
  AutoDisposeProviderElement<bool> createElement() {
    return _IsAuctionFavoriteProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is IsAuctionFavoriteProvider && other.auctionId == auctionId;
  }

  @override
  int get hashCode {
    var hash = _SystemHash.combine(0, runtimeType.hashCode);
    hash = _SystemHash.combine(hash, auctionId.hashCode);

    return _SystemHash.finish(hash);
  }
}

@Deprecated('Will be removed in 3.0. Use Ref instead')
// ignore: unused_element
mixin IsAuctionFavoriteRef on AutoDisposeProviderRef<bool> {
  /// The parameter `auctionId` of this provider.
  String get auctionId;
}

class _IsAuctionFavoriteProviderElement extends AutoDisposeProviderElement<bool>
    with IsAuctionFavoriteRef {
  _IsAuctionFavoriteProviderElement(super.provider);

  @override
  String get auctionId => (origin as IsAuctionFavoriteProvider).auctionId;
}

String _$favoritesHash() => r'86dacc364eec0e21dca7880ea2da4380951b3331';

/// Provider pour les favoris avec persistance locale
///
/// Copied from [Favorites].
@ProviderFor(Favorites)
final favoritesProvider =
    AutoDisposeAsyncNotifierProvider<Favorites, Set<String>>.internal(
      Favorites.new,
      name: r'favoritesProvider',
      debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
          ? null
          : _$favoritesHash,
      dependencies: null,
      allTransitiveDependencies: null,
    );

typedef _$Favorites = AutoDisposeAsyncNotifier<Set<String>>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
