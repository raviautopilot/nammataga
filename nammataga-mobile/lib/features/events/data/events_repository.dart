import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../domain/event_model.dart';
import '../domain/gallery_image_model.dart';

/// Data repository communicating with NammaTAGA Events and Gallery APIs.
class EventsRepository {
  final Dio _dio;

  EventsRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  /// Fetches list of upcoming association events.
  Future<List<EventModel>> getUpcomingEvents() async {
    try {
      final response = await _dio.get(ApiEndpoints.upcomingEvents);
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .map((item) => EventModel.fromJson(item as Map<String, dynamic>))
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load upcoming events',
      );
    }
  }

  /// Fetches all available gallery years.
  Future<List<int>> getGalleryYears() async {
    try {
      final response = await _dio.get(ApiEndpoints.galleryYears);
      if (response.statusCode == 200 && response.data is List) {
        final years = (response.data as List)
            .map((y) {
              if (y is int) return y;
              return int.tryParse(y.toString()) ?? 0;
            })
            .where((y) => y > 0)
            .toList();
        years.sort((a, b) => b.compareTo(a)); // Descending order
        return years;
      }
      return [2026, 2025];
    } catch (_) {
      return [2026, 2025];
    }
  }

  /// Helper to fetch images for a specific year.
  Future<List<GalleryImageModel>> _fetchImagesForYear(int year) async {
    try {
      final response = await _dio.get(ApiEndpoints.galleryByYear(year));
      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .map((item) =>
                GalleryImageModel.fromJson(item as Map<String, dynamic>))
            .toList();
      }
    } catch (_) {}
    return [];
  }

  /// Fetches gallery photos. When [year] is null ("All Memories"), fetches across ALL
  /// available years (both 2026 and 2025) and combines them into a complete chronological archive.
  Future<List<GalleryImageModel>> getGalleryImages({int? year}) async {
    try {
      if (year != null) {
        return await _fetchImagesForYear(year);
      }

      // "All Memories" requested: fetch all known years from API and fallback list
      final fetchedYears = await getGalleryYears();
      final Set<int> allYears = {2026, 2025, ...fetchedYears};
      final sortedYears = allYears.toList()..sort((a, b) => b.compareTo(a));

      final results = await Future.wait(
        sortedYears.map((y) => _fetchImagesForYear(y)),
      );

      final aggregated = <GalleryImageModel>[];
      for (final list in results) {
        aggregated.addAll(list);
      }

      // Deduplicate by ID to prevent duplicate tiles
      final seenIds = <String>{};
      final uniqueImages = <GalleryImageModel>[];
      for (final img in aggregated) {
        if (img.id.isNotEmpty) {
          if (seenIds.add(img.id)) {
            uniqueImages.add(img);
          }
        } else {
          uniqueImages.add(img);
        }
      }

      return uniqueImages;
    } on DioException catch (e) {
      throw Exception(
        e.response?.data?['error'] ??
            e.response?.data?['message'] ??
            'Failed to load gallery images',
      );
    }
  }
}
