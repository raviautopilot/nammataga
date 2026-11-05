import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/resources_repository.dart';
import '../domain/external_link.dart';
import '../domain/resource_category.dart';
import '../domain/resource_document.dart';

/// Provider for ResourcesRepository singleton.
final resourcesRepositoryProvider = Provider<ResourcesRepository>((ref) {
  return ResourcesRepository();
});

/// FutureProvider that fetches the list of available resource categories.
final resourceCategoriesProvider =
    FutureProvider<List<ResourceCategory>>((ref) async {
  final repo = ref.watch(resourcesRepositoryProvider);
  return repo.getCategories();
});

/// Notifier managing the currently selected resource category ID.
/// Automatically selects the first available category upon categories loading.
class SelectedCategoryNotifier extends Notifier<String?> {
  @override
  String? build() {
    ref.listen(resourceCategoriesProvider, (previous, next) {
      next.whenData((categories) {
        if (state == null && categories.isNotEmpty) {
          state = categories.first.id;
        }
      });
    });

    final cached = ref.read(resourceCategoriesProvider).value;
    if (cached != null && cached.isNotEmpty) {
      return cached.first.id;
    }
    return null;
  }

  /// Sets the active category ID, resetting subcategory selection.
  void setCategory(String? categoryId) {
    if (state != categoryId) {
      state = categoryId;
      ref.read(selectedSubcategoryProvider.notifier).clear();
    }
  }

  /// Convenience alias for setting category.
  void selectCategory(String? categoryId) => setCategory(categoryId);

  /// Clears selected category and subcategory.
  void clear() {
    state = null;
    ref.read(selectedSubcategoryProvider.notifier).clear();
  }
}

/// Provider managing the selected resource category ID.
final selectedCategoryProvider =
    NotifierProvider<SelectedCategoryNotifier, String?>(
  SelectedCategoryNotifier.new,
);

/// Notifier managing the selected subcategory ('Central', 'State', or null for all).
class SelectedSubcategoryNotifier extends Notifier<String?> {
  @override
  String? build() => null;

  void setSubcategory(String? subcategory) => state = subcategory;
  void selectSubcategory(String? subcategory) => state = subcategory;
  void clear() => state = null;
}

/// Provider managing the selected subcategory filter.
final selectedSubcategoryProvider =
    NotifierProvider<SelectedSubcategoryNotifier, String?>(
  SelectedSubcategoryNotifier.new,
);

/// FutureProvider family fetching documents for a specific category and optional subcategory.
final categoryDocumentsProvider = FutureProvider.family<
    List<ResourceDocument>,
    ({String categoryId, String? subcategory})>((ref, arg) async {
  final repo = ref.watch(resourcesRepositoryProvider);
  return repo.getDocuments(arg.categoryId, subcategory: arg.subcategory);
});

/// FutureProvider fetching all external links.
final externalLinksProvider = FutureProvider<List<ExternalLink>>((ref) async {
  final repo = ref.watch(resourcesRepositoryProvider);
  return repo.getExternalLinks();
});

/// Convenience provider fetching documents for the currently selected category & subcategory.
final currentCategoryDocumentsProvider =
    FutureProvider<List<ResourceDocument>>((ref) async {
  final categoryId = ref.watch(selectedCategoryProvider);
  if (categoryId == null || categoryId.isEmpty) {
    return [];
  }
  final subcategory = ref.watch(selectedSubcategoryProvider);
  return ref.watch(
    categoryDocumentsProvider((
      categoryId: categoryId,
      subcategory: subcategory,
    )).future,
  );
});
