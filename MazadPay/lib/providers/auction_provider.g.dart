// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'auction_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$auctionNotifierHash() => r'3fba4c019b9c6b4533366677859f448174a81571';

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

abstract class _$AuctionNotifier extends BuildlessAutoDisposeNotifier<Auction> {
  late final String id;

  Auction build(String id);
}

/// See also [AuctionNotifier].
@ProviderFor(AuctionNotifier)
const auctionNotifierProvider = AuctionNotifierFamily();

/// See also [AuctionNotifier].
class AuctionNotifierFamily extends Family<Auction> {
  /// See also [AuctionNotifier].
  const AuctionNotifierFamily();

  /// See also [AuctionNotifier].
  AuctionNotifierProvider call(String id) {
    return AuctionNotifierProvider(id);
  }

  @override
  AuctionNotifierProvider getProviderOverride(
    covariant AuctionNotifierProvider provider,
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
  String? get name => r'auctionNotifierProvider';
}

/// See also [AuctionNotifier].
class AuctionNotifierProvider
    extends AutoDisposeNotifierProviderImpl<AuctionNotifier, Auction> {
  /// See also [AuctionNotifier].
  AuctionNotifierProvider(String id)
    : this._internal(
        () => AuctionNotifier()..id = id,
        from: auctionNotifierProvider,
        name: r'auctionNotifierProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$auctionNotifierHash,
        dependencies: AuctionNotifierFamily._dependencies,
        allTransitiveDependencies:
            AuctionNotifierFamily._allTransitiveDependencies,
        id: id,
      );

  AuctionNotifierProvider._internal(
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
  Auction runNotifierBuild(covariant AuctionNotifier notifier) {
    return notifier.build(id);
  }

  @override
  Override overrideWith(AuctionNotifier Function() create) {
    return ProviderOverride(
      origin: this,
      override: AuctionNotifierProvider._internal(
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
  AutoDisposeNotifierProviderElement<AuctionNotifier, Auction> createElement() {
    return _AuctionNotifierProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is AuctionNotifierProvider && other.id == id;
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
mixin AuctionNotifierRef on AutoDisposeNotifierProviderRef<Auction> {
  /// The parameter `id` of this provider.
  String get id;
}

class _AuctionNotifierProviderElement
    extends AutoDisposeNotifierProviderElement<AuctionNotifier, Auction>
    with AuctionNotifierRef {
  _AuctionNotifierProviderElement(super.provider);

  @override
  String get id => (origin as AuctionNotifierProvider).id;
}

String _$auctionHistoryHash() => r'526fdbf8c5ebf4bcada187dcc7ea72c60502a4db';

abstract class _$AuctionHistory
    extends BuildlessAutoDisposeNotifier<List<BidEntry>> {
  late final String auctionId;

  List<BidEntry> build(String auctionId);
}

/// See also [AuctionHistory].
@ProviderFor(AuctionHistory)
const auctionHistoryProvider = AuctionHistoryFamily();

/// See also [AuctionHistory].
class AuctionHistoryFamily extends Family<List<BidEntry>> {
  /// See also [AuctionHistory].
  const AuctionHistoryFamily();

  /// See also [AuctionHistory].
  AuctionHistoryProvider call(String auctionId) {
    return AuctionHistoryProvider(auctionId);
  }

  @override
  AuctionHistoryProvider getProviderOverride(
    covariant AuctionHistoryProvider provider,
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
  String? get name => r'auctionHistoryProvider';
}

/// See also [AuctionHistory].
class AuctionHistoryProvider
    extends AutoDisposeNotifierProviderImpl<AuctionHistory, List<BidEntry>> {
  /// See also [AuctionHistory].
  AuctionHistoryProvider(String auctionId)
    : this._internal(
        () => AuctionHistory()..auctionId = auctionId,
        from: auctionHistoryProvider,
        name: r'auctionHistoryProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$auctionHistoryHash,
        dependencies: AuctionHistoryFamily._dependencies,
        allTransitiveDependencies:
            AuctionHistoryFamily._allTransitiveDependencies,
        auctionId: auctionId,
      );

  AuctionHistoryProvider._internal(
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
  List<BidEntry> runNotifierBuild(covariant AuctionHistory notifier) {
    return notifier.build(auctionId);
  }

  @override
  Override overrideWith(AuctionHistory Function() create) {
    return ProviderOverride(
      origin: this,
      override: AuctionHistoryProvider._internal(
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
  AutoDisposeNotifierProviderElement<AuctionHistory, List<BidEntry>>
  createElement() {
    return _AuctionHistoryProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is AuctionHistoryProvider && other.auctionId == auctionId;
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
mixin AuctionHistoryRef on AutoDisposeNotifierProviderRef<List<BidEntry>> {
  /// The parameter `auctionId` of this provider.
  String get auctionId;
}

class _AuctionHistoryProviderElement
    extends AutoDisposeNotifierProviderElement<AuctionHistory, List<BidEntry>>
    with AuctionHistoryRef {
  _AuctionHistoryProviderElement(super.provider);

  @override
  String get auctionId => (origin as AuctionHistoryProvider).auctionId;
}

// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
