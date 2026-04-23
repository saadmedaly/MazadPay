// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'auction_provider_api.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$auctionNotifierApiHash() =>
    r'9796454b558295dbfa7223dfd76db38c5296ce4c';

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

abstract class _$AuctionNotifierApi
    extends BuildlessAutoDisposeAsyncNotifier<Auction> {
  late final String id;

  FutureOr<Auction> build(String id);
}

/// Provider pour les enchères utilisant l'API backend
/// Remplace le provider mocké avec de vraies données
///
/// Copied from [AuctionNotifierApi].
@ProviderFor(AuctionNotifierApi)
const auctionNotifierApiProvider = AuctionNotifierApiFamily();

/// Provider pour les enchères utilisant l'API backend
/// Remplace le provider mocké avec de vraies données
///
/// Copied from [AuctionNotifierApi].
class AuctionNotifierApiFamily extends Family<AsyncValue<Auction>> {
  /// Provider pour les enchères utilisant l'API backend
  /// Remplace le provider mocké avec de vraies données
  ///
  /// Copied from [AuctionNotifierApi].
  const AuctionNotifierApiFamily();

  /// Provider pour les enchères utilisant l'API backend
  /// Remplace le provider mocké avec de vraies données
  ///
  /// Copied from [AuctionNotifierApi].
  AuctionNotifierApiProvider call(String id) {
    return AuctionNotifierApiProvider(id);
  }

  @override
  AuctionNotifierApiProvider getProviderOverride(
    covariant AuctionNotifierApiProvider provider,
  ) {
    return call(provider.id);
  }

  static const Iterable<ProviderOrFamily>? _dependencies = null;

  @override
  Iterable<ProviderOrFamily>? get dependencies => _dependencies;

  static const Iterable<ProviderOrFamily>? _allTransitiveDependencies = null;

  @override
  Iterable<ProviderOrFamily>? get allTransitiveDependencies =>
      _allTransitiveDependencies;

  @override
  String? get name => r'auctionNotifierApiProvider';
}

/// Provider pour les enchères utilisant l'API backend
/// Remplace le provider mocké avec de vraies données
///
/// Copied from [AuctionNotifierApi].
class AuctionNotifierApiProvider
    extends AutoDisposeAsyncNotifierProviderImpl<AuctionNotifierApi, Auction> {
  /// Provider pour les enchères utilisant l'API backend
  /// Remplace le provider mocké avec de vraies données
  ///
  /// Copied from [AuctionNotifierApi].
  AuctionNotifierApiProvider(String id)
    : this._internal(
        () => AuctionNotifierApi()..id = id,
        from: auctionNotifierApiProvider,
        name: r'auctionNotifierApiProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$auctionNotifierApiHash,
        dependencies: AuctionNotifierApiFamily._dependencies,
        allTransitiveDependencies:
            AuctionNotifierApiFamily._allTransitiveDependencies,
        id: id,
      );

  AuctionNotifierApiProvider._internal(
    super._createNotifier, {
    required super.name,
    required super.dependencies,
    required super.allTransitiveDependencies,
    required super.debugGetCreateSourceHash,
    required super.from,
    required this.id,
  }) : super.internal();

  final String id;

  @override
  FutureOr<Auction> runNotifierBuild(covariant AuctionNotifierApi notifier) {
    return notifier.build(id);
  }

  @override
  Override overrideWith(AuctionNotifierApi Function() create) {
    return ProviderOverride(
      origin: this,
      override: AuctionNotifierApiProvider._internal(
        () => create()..id = id,
        from: from,
        name: null,
        dependencies: null,
        allTransitiveDependencies: null,
        debugGetCreateSourceHash: null,
        id: id,
      ),
    );
  }

  @override
  AutoDisposeAsyncNotifierProviderElement<AuctionNotifierApi, Auction>
  createElement() {
    return _AuctionNotifierApiProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is AuctionNotifierApiProvider && other.id == id;
  }

  @override
  int get hashCode {
    var hash = _SystemHash.combine(0, runtimeType.hashCode);
    hash = _SystemHash.combine(hash, id.hashCode);

    return _SystemHash.finish(hash);
  }
}

@Deprecated('Will be removed in 3.0. Use Ref instead')
// ignore: unused_element
mixin AuctionNotifierApiRef on AutoDisposeAsyncNotifierProviderRef<Auction> {
  /// The parameter `id` of this provider.
  String get id;
}

class _AuctionNotifierApiProviderElement
    extends AutoDisposeAsyncNotifierProviderElement<AuctionNotifierApi, Auction>
    with AuctionNotifierApiRef {
  _AuctionNotifierApiProviderElement(super.provider);

  @override
  String get id => (origin as AuctionNotifierApiProvider).id;
}

String _$auctionHistoryApiHash() => r'fa09d64e5090762a2203683b6943881fcf3d3871';

abstract class _$AuctionHistoryApi
    extends BuildlessAutoDisposeAsyncNotifier<List<BidEntry>> {
  late final String auctionId;

  FutureOr<List<BidEntry>> build(String auctionId);
}

/// See also [AuctionHistoryApi].
@ProviderFor(AuctionHistoryApi)
const auctionHistoryApiProvider = AuctionHistoryApiFamily();

/// See also [AuctionHistoryApi].
class AuctionHistoryApiFamily extends Family<AsyncValue<List<BidEntry>>> {
  /// See also [AuctionHistoryApi].
  const AuctionHistoryApiFamily();

  /// See also [AuctionHistoryApi].
  AuctionHistoryApiProvider call(String auctionId) {
    return AuctionHistoryApiProvider(auctionId);
  }

  @override
  AuctionHistoryApiProvider getProviderOverride(
    covariant AuctionHistoryApiProvider provider,
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
  String? get name => r'auctionHistoryApiProvider';
}

/// See also [AuctionHistoryApi].
class AuctionHistoryApiProvider
    extends
        AutoDisposeAsyncNotifierProviderImpl<
          AuctionHistoryApi,
          List<BidEntry>
        > {
  /// See also [AuctionHistoryApi].
  AuctionHistoryApiProvider(String auctionId)
    : this._internal(
        () => AuctionHistoryApi()..auctionId = auctionId,
        from: auctionHistoryApiProvider,
        name: r'auctionHistoryApiProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$auctionHistoryApiHash,
        dependencies: AuctionHistoryApiFamily._dependencies,
        allTransitiveDependencies:
            AuctionHistoryApiFamily._allTransitiveDependencies,
        auctionId: auctionId,
      );

  AuctionHistoryApiProvider._internal(
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
  FutureOr<List<BidEntry>> runNotifierBuild(
    covariant AuctionHistoryApi notifier,
  ) {
    return notifier.build(auctionId);
  }

  @override
  Override overrideWith(AuctionHistoryApi Function() create) {
    return ProviderOverride(
      origin: this,
      override: AuctionHistoryApiProvider._internal(
        () => create()..auctionId = auctionId,
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
  AutoDisposeAsyncNotifierProviderElement<AuctionHistoryApi, List<BidEntry>>
  createElement() {
    return _AuctionHistoryApiProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is AuctionHistoryApiProvider && other.auctionId == auctionId;
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
mixin AuctionHistoryApiRef
    on AutoDisposeAsyncNotifierProviderRef<List<BidEntry>> {
  /// The parameter `auctionId` of this provider.
  String get auctionId;
}

class _AuctionHistoryApiProviderElement
    extends
        AutoDisposeAsyncNotifierProviderElement<
          AuctionHistoryApi,
          List<BidEntry>
        >
    with AuctionHistoryApiRef {
  _AuctionHistoryApiProviderElement(super.provider);

  @override
  String get auctionId => (origin as AuctionHistoryApiProvider).auctionId;
}

// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
