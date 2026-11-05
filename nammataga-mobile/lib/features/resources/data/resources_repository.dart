import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../domain/external_link.dart';
import '../domain/resource_category.dart';
import '../domain/resource_document.dart';

/// Repository interacting with backend resources, categories, documents, and external links.
class ResourcesRepository {
  final Dio _dio;

  ResourcesRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  /// Fetches all resource categories.
  /// If backend doesn't return subcategories in category list, populates
  /// `subcategories: ['Central', 'State']` for `scheme-gos`.
  Future<List<ResourceCategory>> getCategories() async {
    try {
      final response = await _dio.get(ApiEndpoints.resourceCategories);
      if (response.statusCode == 200 && response.data is List) {
        final categories = (response.data as List).map((item) {
          final category =
              ResourceCategory.fromJson(item as Map<String, dynamic>);
          if ((category.subcategories == null ||
                  category.subcategories!.isEmpty) &&
              category.id == 'scheme-gos') {
            return category.copyWith(
              subcategories: const ['Central', 'State'],
            );
          }
          return category;
        }).toList();
        return categories;
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load resource categories',
      );
    }
  }

  /// Fetches documents belonging to [categoryId], optionally filtered by [subcategory].
  /// Applies [sortDocuments] to sort descending by year and ascending by title.
  Future<List<ResourceDocument>> getDocuments(
    String categoryId, {
    String? subcategory,
  }) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.documentsByCategory(categoryId, subcategory),
      );
      if (response.statusCode == 200 && response.data is List) {
        final docs = (response.data as List)
            .map((item) =>
                ResourceDocument.fromJson(item as Map<String, dynamic>))
            .toList();
        return sortDocuments(docs);
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load documents for category: $categoryId',
      );
    }
  }

  /// Fetches list of external links and sorts alphabetically ascending by title.
  Future<List<ExternalLink>> getExternalLinks() async {
    try {
      final response = await _dio.get(ApiEndpoints.externalLinks);
      if (response.statusCode == 200 && response.data is List) {
        final links = (response.data as List)
            .map((item) =>
                ExternalLink.fromJson(item as Map<String, dynamic>))
            .toList();
        return sortExternalLinks(links);
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load external links',
      );
    }
  }
}
