import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/features/home/presentation/home_screen.dart';
import 'package:nammataga_mobile/features/profile/domain/member_notification_model.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/presentation/announcements_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/profile_overview_screen.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';

const mockTestProfile = MemberProfile(
  id: 'usr_001',
  tagaId: 'TAGA-2024-0842',
  username: 'sivakumar',
  name: 'R. Sivakumar',
  initial: '',
  gender: 'Male',
  dateOfBirth: '1985-05-15',
  fatherName: 'P. Ramanathan',
  motherName: 'R. Meenakshi',
  mobileNumber: '9876543210',
  emailId: 'sivakumar.taga@gmail.com',
  designation: 'Agricultural Officer',
  recruitmentBatch: '2012',
  seniorityNumber: 'SN-4021',
  workingDistrict: 'Thanjavur',
  nativeDistrict: 'Madurai',
  residentialAddress: 'Thanjavur',
  permanentAddress: 'Madurai',
  educationalQualification: 'B.Sc.',
  tbfNumber: 'TBF-8921',
  cpsGpfNumber: 'CPS-548920',
  isPaid: true,
  subscriptionActive: true,
  paymentStatus: 'Paid',
);

void main() {
  group('MemberNotification Model', () {
    test('Parses published_at and formats date correctly', () {
      final json = {
        'id': 'ann_101',
        'title': 'General Body Meeting',
        'message': 'Meeting at TAGA auditorium.',
        'priority': 'urgent',
        'published_at': '2026-11-19T08:00:00Z',
      };
      final notification = MemberNotification.fromJson(json);

      expect(notification.id, 'ann_101');
      expect(notification.title, 'General Body Meeting');
      expect(notification.message, 'Meeting at TAGA auditorium.');
      expect(notification.priority, 'urgent');
      expect(notification.publishedAt, '2026-11-19T08:00:00Z');
      expect(notification.createdAt, '2026-11-19T08:00:00Z');
      expect(notification.formattedDate, 'Nov 19, 2026');
    });

    test('Falls back to created_at or sent_at when published_at is absent', () {
      final jsonCreatedAt = {
        'id': 'ann_102',
        'title': 'Notice 2',
        'message': 'Sample message',
        'priority': 'normal',
        'created_at': '2026-10-15T10:00:00Z',
      };
      final n1 = MemberNotification.fromJson(jsonCreatedAt);
      expect(n1.publishedAt, '2026-10-15T10:00:00Z');
      expect(n1.formattedDate, 'Oct 15, 2026');

      final jsonSentAt = {
        'id': 'ann_103',
        'title': 'Notice 3',
        'message': 'Sample message 3',
        'priority': 'high',
        'sent_at': '2026-12-01T10:00:00Z',
      };
      final n2 = MemberNotification.fromJson(jsonSentAt);
      expect(n2.publishedAt, '2026-12-01T10:00:00Z');
      expect(n2.formattedDate, 'Dec 1, 2026');
    });

    test('Handles malformed or missing dates gracefully', () {
      const n = MemberNotification(
        id: 'ann_empty',
        title: 'Empty Date',
        message: 'No date provided',
        priority: 'normal',
        publishedAt: '',
      );
      expect(n.formattedDate, 'Recent');
    });
  });

  group('HomeScreen Notification Bell', () {
    testWidgets('Renders bell icon with small green dot when active announcements exist', (tester) async {
      final activeAnnouncements = [
        const MemberNotification(
          id: 'ann_1',
          title: 'Special Event',
          message: 'Important update for all members',
          priority: 'urgent',
          publishedAt: '2026-11-19T00:00:00Z',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberAnnouncementsProvider.overrideWith((ref) => Future.value(activeAnnouncements)),
          ],
          child: const MaterialApp(home: HomeScreen()),
        ),
      );

      await tester.pump(const Duration(milliseconds: 200));

      // Notification bell icon is present
      expect(find.byIcon(LucideIcons.bell), findsOneWidget);
      // Small green dot indicator is visible
      expect(find.byKey(const Key('active_announcements_dot')), findsOneWidget);
      // Profile icon button is present after the bell
      expect(find.byIcon(LucideIcons.user), findsOneWidget);
      // No unread numeric count badge
      expect(find.text('1'), findsNothing);
    });

    testWidgets('Removes green dot indicator when there are no active announcements', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberAnnouncementsProvider.overrideWith((ref) => Future.value([])),
          ],
          child: const MaterialApp(home: HomeScreen()),
        ),
      );

      await tester.pump(const Duration(milliseconds: 200));

      // Notification bell icon is present
      expect(find.byIcon(LucideIcons.bell), findsOneWidget);
      // Green dot indicator is NOT present
      expect(find.byKey(const Key('active_announcements_dot')), findsNothing);
    });

    testWidgets('Tapping notification bell navigates to /profile/announcements', (tester) async {
      bool navigatedToAnnouncements = false;

      final router = GoRouter(
        initialLocation: '/home',
        routes: [
          GoRoute(
            path: '/home',
            builder: (context, state) => const HomeScreen(),
          ),
          GoRoute(
            path: '/profile/announcements',
            builder: (context, state) {
              navigatedToAnnouncements = true;
              return const Scaffold(body: Text('Announcements Destination Screen'));
            },
          ),
          GoRoute(
            path: '/profile',
            builder: (context, state) => const Scaffold(body: Text('Profile Screen')),
          ),
        ],
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberAnnouncementsProvider.overrideWith((ref) => Future.value([
              const MemberNotification(
                id: 'ann_1',
                title: 'Announcement Notice',
                message: 'Body message',
                priority: 'normal',
                publishedAt: '2026-11-19T00:00:00Z',
              ),
            ])),
          ],
          child: MaterialApp.router(
            routerConfig: router,
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 200));

      // Tap on the announcements bell button
      await tester.tap(find.byIcon(LucideIcons.bell));
      await tester.pumpAndSettle();

      expect(navigatedToAnnouncements, isTrue);
      expect(find.text('Announcements Destination Screen'), findsOneWidget);
    });
  });

  group('AnnouncementsScreen Simplified UI', () {
    testWidgets('Renders clean announcement card with Title, Message, Priority, and Date without read/unread button', (tester) async {
      final announcements = [
        const MemberNotification(
          id: 'ann_1',
          title: 'Special General Body Meeting 2026',
          message: 'All members are cordially invited to the meeting.',
          priority: 'urgent',
          publishedAt: '2026-11-19T00:00:00Z',
        ),
        const MemberNotification(
          id: 'ann_2',
          title: 'Subscription Reminder',
          message: 'Please complete your annual dues on time.',
          priority: 'normal',
          publishedAt: '2026-10-15T00:00:00Z',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberAnnouncementsProvider.overrideWith((ref) => Future.value(announcements)),
          ],
          child: const MaterialApp(home: AnnouncementsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      // Title & Message
      expect(find.text('Special General Body Meeting 2026'), findsOneWidget);
      expect(find.text('All members are cordially invited to the meeting.'), findsOneWidget);
      expect(find.text('URGENT'), findsOneWidget);
      expect(find.text('Nov 19, 2026'), findsOneWidget);

      expect(find.text('Subscription Reminder'), findsOneWidget);
      expect(find.text('Please complete your annual dues on time.'), findsOneWidget);
      expect(find.text('NORMAL'), findsOneWidget);
      expect(find.text('Oct 15, 2026'), findsOneWidget);

      // Verify elimination of read/unread elements
      expect(find.text('Tap to mark read'), findsNothing);
      expect(find.byIcon(LucideIcons.check), findsNothing);
    });

    testWidgets('Renders empty state when announcements list is empty', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberAnnouncementsProvider.overrideWith((ref) => Future.value([])),
          ],
          child: const MaterialApp(home: AnnouncementsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('No Announcements'), findsOneWidget);
      expect(
        find.text('You are all caught up on official association announcements.'),
        findsOneWidget,
      );
    });
  });

  group('ProfileOverviewScreen Simplified Navigation List', () {
    testWidgets('Does not display unread count badge on Official Announcements menu item', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockTestProfile)),
          ],
          child: const MaterialApp(home: ProfileOverviewScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Official Announcements'), findsOneWidget);
      // No unread badge
      expect(find.textContaining('new'), findsNothing);
    });
  });
}
