import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/events_repository.dart';
import '../domain/event_model.dart';
import '../domain/gallery_image_model.dart';

/// Provider for the EventsRepository singleton.
final eventsRepositoryProvider = Provider<EventsRepository>((ref) {
  return EventsRepository();
});

/// Async notifier managing upcoming events state and refresh capabilities.
class UpcomingEventsNotifier extends AsyncNotifier<List<EventModel>> {
  @override
  Future<List<EventModel>> build() async {
    final repo = ref.watch(eventsRepositoryProvider);
    final events = await repo.getUpcomingEvents();
    return events.where((e) => e.isUpcoming).toList();
  }

  /// Triggers a pull-to-refresh reload.
  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final repo = ref.read(eventsRepositoryProvider);
      final events = await repo.getUpcomingEvents();
      return events.where((e) => e.isUpcoming).toList();
    });
  }
}

final upcomingEventsProvider =
    AsyncNotifierProvider<UpcomingEventsNotifier, List<EventModel>>(
  UpcomingEventsNotifier.new,
);

/// Provider fetching available gallery years.
final galleryYearsProvider = FutureProvider<List<int>>((ref) async {
  final repo = ref.watch(eventsRepositoryProvider);
  return repo.getGalleryYears();
});

/// Notifier for selected gallery year (null = all memories).
class SelectedGalleryYearNotifier extends Notifier<int?> {
  @override
  int? build() => null;

  void setYear(int? year) => state = year;
}

final selectedGalleryYearProvider =
    NotifierProvider<SelectedGalleryYearNotifier, int?>(
  SelectedGalleryYearNotifier.new,
);

/// Provider fetching gallery images for a given year.
final galleryImagesProvider =
    FutureProvider.family<List<GalleryImageModel>, int?>((ref, year) async {
  final repo = ref.watch(eventsRepositoryProvider);
  return repo.getGalleryImages(year: year);
});

/// Notifier for real-time upcoming events search queries.
class EventsSearchQueryNotifier extends Notifier<String> {
  @override
  String build() => '';

  void setQuery(String query) => state = query;
}

final eventsSearchQueryProvider =
    NotifierProvider<EventsSearchQueryNotifier, String>(
  EventsSearchQueryNotifier.new,
);
