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

  group('Upcoming Events Filtering & Date Handling', () {
    test('Past dates (e.g. May 2026, March 2025) are flagged as isUpcoming: false and isPast: true', () {
      final pastEvent1 = EventModel.fromJson({
        'id': 'p1',
        'title': 'Past Hello Event',
        'date': '2026-05-18',
        'location': 'Chennai',
        'status': 'upcoming', // Backend returned upcoming but date passed!
      });

      final pastEvent2 = EventModel.fromJson({
        'id': 'p2',
        'title': 'Annual General Meeting 2025',
        'date': '2025-03-15 16:50',
        'location': 'TAGA Towers',
        'status': 'upcoming',
      });

      expect(pastEvent1.isUpcoming, isFalse);
      expect(pastEvent1.isPast, isTrue);
      expect(pastEvent1.daysRemaining, isNull);
      expect(pastEvent1.formattedTime, 'All Day');

      expect(pastEvent2.isUpcoming, isFalse);
      expect(pastEvent2.isPast, isTrue);
      expect(pastEvent2.daysRemaining, isNull);
    });

    test('Future dates (e.g. Nov 2026, Dec 2026, Jan 2027) are flagged as isUpcoming: true', () {
      final futureEvent1 = EventModel.fromJson({
        'id': 'f1',
        'title': 'TAGA State Agriculture Convention 2026',
        'date': '2026-11-28 09:30',
        'location': 'TAGA Auditorium, Chennai',
        'status': 'upcoming',
      });

      final futureEvent2 = EventModel.fromJson({
        'id': 'f2',
        'title': 'TAGA Conclave 2027',
        'date': '2027-01-24 14:00',
        'location': 'Trichy',
        'status': 'upcoming',
      });

      expect(futureEvent1.isUpcoming, isTrue);
      expect(futureEvent1.isPast, isFalse);
      expect(futureEvent1.daysRemaining, isNotNull);
      expect(futureEvent1.daysRemaining! > 0, isTrue);
      expect(futureEvent1.formattedTime, '09:30 AM');

      expect(futureEvent2.isUpcoming, isTrue);
      expect(futureEvent2.isPast, isFalse);
    });

    test('Status "completed" or "cancelled" is excluded even with future date', () {
      final completedEvent = EventModel.fromJson({
        'id': 'c1',
        'title': 'Cancelled Event',
        'date': '2026-11-28 09:30',
        'location': 'Chennai',
        'status': 'completed',
      });

      expect(completedEvent.isUpcoming, isFalse);
      expect(completedEvent.isPast, isTrue);
    });
  });
}

class _MockUpcomingEventsNotifier extends UpcomingEventsNotifier {
  final List<EventModel> mockEvents;
  _MockUpcomingEventsNotifier(this.mockEvents);

  @override
  Future<List<EventModel>> build() async => mockEvents;
}
