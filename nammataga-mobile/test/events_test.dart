import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/events/domain/event_model.dart';
import 'package:nammataga_mobile/features/events/domain/gallery_image_model.dart';
import 'package:nammataga_mobile/features/events/presentation/events_screen.dart';
import 'package:nammataga_mobile/features/events/presentation/widgets/event_card.dart';
import 'package:nammataga_mobile/features/events/providers/events_provider.dart';

void main() {
  setUp(() {
    Animate.restartOnHotReload = false;
  });

  final sampleEvents = [
    const EventModel(
      id: 'e1',
      title: 'State Agriculture Convention 2026',
      dateRaw: '2026-11-28 09:30',
      location: 'TAGA Auditorium, Chennai',
      description: 'Annual flagship convention with technical keynotes.',
      attendees: 300,
      status: 'upcoming',
    ),
    const EventModel(
      id: 'e2',
      title: 'AgTech & AI Workshop',
      dateRaw: '2026-12-15 10:00',
      location: 'TNAU Campus, Coimbatore',
      description: 'Smart irrigation & drone diagnostic workshop.',
      attendees: 150,
      status: 'upcoming',
    ),
  ];

  final sampleImages = [
    const GalleryImageModel(
      id: 'g1',
      title: 'Honorable Agriculture Minister Keynote',
      event: 'State Conference',
      imageUrl: '/images/eventImages/2025/test1.jpeg',
      date: '2026-04-20',
      year: 2025,
    ),
  ];

  Widget createTestWidget() {
    return ProviderScope(
      overrides: [
        upcomingEventsProvider.overrideWith(
          () => _MockUpcomingEventsNotifier(sampleEvents),
        ),
        galleryYearsProvider.overrideWith((ref) => Future.value([2026, 2025])),
        galleryImagesProvider.overrideWith((ref, year) => Future.value(sampleImages)),
      ],
      child: const MaterialApp(
        home: EventsScreen(),
      ),
    );
  }

  testWidgets('EventsScreen renders curved header, title, and both tab buttons', (
    tester,
  ) async {
    await tester.pumpWidget(createTestWidget());
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(seconds: 1));

    expect(find.text('Events & Media'), findsOneWidget);
    expect(find.text('Upcoming Events'), findsOneWidget);
    expect(find.text('Photo Gallery'), findsOneWidget);
  });

  testWidgets('UpcomingEvents renders sample event cards with details', (
    tester,
  ) async {
    await tester.pumpWidget(createTestWidget());
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(seconds: 1));

    expect(find.text('State Agriculture Convention 2026'), findsOneWidget);
    expect(find.text('AgTech & AI Workshop'), findsOneWidget);
    expect(find.text('TAGA Auditorium, Chennai'), findsOneWidget);
    expect(find.byType(EventCard), findsNWidgets(2));
  });

  testWidgets('Tapping Photo Gallery tab switches to Gallery view with year filter', (
    tester,
  ) async {
    await tester.pumpWidget(createTestWidget());
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(seconds: 1));

    // Tap Photo Gallery tab
    await tester.tap(find.text('Photo Gallery'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pump(const Duration(seconds: 1));
    await tester.pump(const Duration(seconds: 1));

    expect(find.text('All Memories'), findsOneWidget);
    expect(find.text('2026'), findsOneWidget);
    expect(find.text('2025'), findsWidgets);
  });

  test('GalleryImageModel fullImageUrl properly percent-encodes spaces and ampersands', () {
    const model = GalleryImageModel(
      id: 'g2',
      title: 'Test',
      event: 'Test Event',
      imageUrl: '/images/eventImages/2025/Respected Agricultural Production Commissioner & Principal Secretary.jpeg',
      date: '2026-04-20',
      year: 2025,
    );

    final url = model.fullImageUrl;
    expect(url, contains('Respected%20Agricultural%20Production%20Commissioner%20%26%20Principal%20Secretary.jpeg'));
    expect(url.contains(' '), isFalse);
    expect(url.contains(' & '), isFalse);
  });

  test('GalleryImageModel fullImageUrl properly percent-encodes single quotes for web safety', () {
    const model = GalleryImageModel(
      id: 'g7',
      title: 'Newly promoted JDA\'s blessings',
      event: 'Test Event',
      imageUrl: '/images/eventImages/2026/Newly promoted JDA\'s getting blessings from Honorable Minister of Agriculture - Farmers welfare.jpeg',
      date: '2026-04-20',
      year: 2026,
    );

    final url = model.fullImageUrl;
    expect(url, contains('JDA%27s'));
    expect(url.contains("'"), isFalse);
  });
}

class _MockUpcomingEventsNotifier extends UpcomingEventsNotifier {
  final List<EventModel> mockEvents;
  _MockUpcomingEventsNotifier(this.mockEvents);

  @override
  Future<List<EventModel>> build() async => mockEvents;
}
