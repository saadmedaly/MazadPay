import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/category_api.dart';
import '../services/realtime_sync_service.dart';

final categoryApiProvider = Provider<CategoryApi>((ref) => CategoryApi());

/// Provider pour les catégories depuis l'API
///
/// Customer #20: categoriesProvider previously only ever refetched on
/// provider recreation (app restart) -- no invalidation call site existed
/// anywhere. This subscribes to RealtimeSyncService.events for
/// category.updated (and to catchUpSignal, for the resume/reconnect case)
/// and calls ref.invalidateSelf(), which causes any active `ref.watch` of
/// this exact provider to rerun getCategories() -- the smallest targeted
/// fix (categories only, per Phase 7 "do not refresh every provider for
/// every event") without restructuring this into a StateNotifier.
final categoriesProvider = FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final categoryApi = ref.watch(categoryApiProvider);

  final eventSub = RealtimeSyncService().events.listen((event) {
    if (event.type == RealtimeEventType.categoryUpdated) {
      ref.invalidateSelf();
    }
  });
  final catchUpSub = RealtimeSyncService().catchUpSignal.listen((_) {
    ref.invalidateSelf();
  });
  ref.onDispose(() {
    eventSub.cancel();
    catchUpSub.cancel();
  });

  final response = await categoryApi.getCategories();

  if (response.success && response.data != null) {
    final data = response.data!;
    // La réponse est maintenant directement une liste
    final list = data as List;
    return list.whereType<Map>().map((item) => Map<String, dynamic>.from(item)).toList();
    }
  return [];
});

/// Provider pour les sous-catégories d'une catégorie spécifique
final subCategoriesProvider = FutureProvider.family<List<Map<String, dynamic>>, String>((ref, categoryId) async {
  final categoryApi = ref.watch(categoryApiProvider);
  final response = await categoryApi.getSubCategories(categoryId);

  if (response.success && response.data != null) {
    final data = response.data!;
    if (data is List) {
      return (data as List).whereType<Map>().map((item) => Map<String, dynamic>.from(item)).toList();
    } else {
      final subCategories = data['sub_categories'] ?? data['subCategories'] ?? data['data'] ?? [];
      if (subCategories is List) {
        return subCategories.whereType<Map>().map((item) => Map<String, dynamic>.from(item)).toList();
      }
    }
  
  }
  return [];
});
