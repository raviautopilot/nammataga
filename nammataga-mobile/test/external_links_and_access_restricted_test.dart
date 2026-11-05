import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/features/resources/domain/external_link.dart';
import 'package:nammataga_mobile/features/resources/presentation/resources_access_restricted_screen.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/external_links_view.dart';
import 'package:nammataga_mobile/features/resources/providers/resources_provider.dart';

void main() {
  group('ExternalLinksView Tests', () {
    testWidgets('renders list of provided external links with icons and titles', (tester) async {
      final testLinks = [
        const ExternalLink(
          title: 'Agriculture Department',
          url: 'https://www.tnagrisnet.tn.gov.in/',
        ),
        const ExternalLink(
          title: 'TNAU Agritech Portal',
          url: 'http://www.agritech.tnau.ac.in/',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ExternalLinksView(
                links: testLinks,
              ),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Agriculture Department'), findsOneWidget);
      expect(find.text('TNAU Agritech Portal'), findsOneWidget);
      expect(find.text('tnagrisnet.tn.gov.in'), findsOneWidget);
      expect(find.text('agritech.tnau.ac.in'), findsOneWidget);
      expect(find.byIcon(LucideIcons.externalLink), findsNWidgets(2));
    });

    testWidgets('triggers onLinkTap when a tile is tapped', (tester) async {
      ExternalLink? tappedLink;
      final testLinks = [
        const ExternalLink(
          title: 'Uzhavar App Portal',
          url: 'https://www.tnagrisnet.tn.gov.in/people_app',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ExternalLinksView(
                links: testLinks,
                onLinkTap: (link) {
                  tappedLink = link;
                },
              ),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      await tester.tap(find.text('Uzhavar App Portal'));
      await tester.pumpAndSettle();

      expect(tappedLink, isNotNull);
      expect(tappedLink!.title, 'Uzhavar App Portal');
      expect(tappedLink!.url, 'https://www.tnagrisnet.tn.gov.in/people_app');
    });

    testWidgets('renders empty state when links list is empty', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ExternalLinksView(
                links: [],
              ),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('No external links found'), findsOneWidget);
      expect(
        find.text(
          'Official government and departmental portals will appear here when published.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('renders error state and triggers onRetry when tapped', (tester) async {
      bool retried = false;

      await tester.pumpWidget(
        ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ExternalLinksView(
                links: const [],
                errorMessage: 'Failed to connect to government servers.',
                onRetry: () {
                  retried = true;
                },
              ),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Unable to Load External Links'), findsOneWidget);
      expect(find.text('Failed to connect to government servers.'), findsOneWidget);
      expect(find.text('Retry'), findsOneWidget);

      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();

      expect(retried, isTrue);
    });

    testWidgets('renders loading state when isLoading is true', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: Scaffold(
              body: ExternalLinksView(
                links: [],
                isLoading: true,
              ),
            ),
          ),
        ),
      );

      await tester.pump();

      // Loading state has 5 skeleton containers
      expect(find.byType(Container), findsWidgets);
      expect(find.text('No external links found'), findsNothing);
    });

    testWidgets('watches externalLinksProvider and falls back to defaultGovernmentLinks', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            externalLinksProvider.overrideWith((ref) async => <ExternalLink>[]),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: ExternalLinksView(),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Should fall back to defaultGovernmentLinks
      expect(find.text('Agriculture Department'), findsOneWidget);
      expect(find.text('Horticulture Department'), findsOneWidget);
      expect(find.text('Agri Engineering Department'), findsOneWidget);
    });

    test('defaultGovernmentLinks contains expected real government portals', () {
      expect(ExternalLinksView.defaultGovernmentLinks.length, greaterThanOrEqualTo(20));
      expect(
        ExternalLinksView.defaultGovernmentLinks.any((l) => l.title == 'Agriculture Department'),
        isTrue,
      );
      expect(
        ExternalLinksView.defaultGovernmentLinks.any((l) => l.title == 'Agri Engineering Department'),
        isTrue,
      );
      expect(
        ExternalLinksView.defaultGovernmentLinks.any((l) => l.title == 'IFHRMS Portal'),
        isTrue,
      );
    });
  });

  group('ResourcesAccessRestrictedScreen Tests', () {
    testWidgets('displays badge, heading, subtitle, and explanatory card matching website', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: ResourcesAccessRestrictedScreen(),
        ),
      );

      await tester.pumpAndSettle();

      // Badge
      expect(find.text('For Paid Members Only'), findsOneWidget);
      expect(find.byIcon(LucideIcons.fileText), findsWidgets);

      // Heading & Subtitle
      expect(find.text('Resources Center'), findsWidgets);
      expect(
        find.text('Official documents, government orders, and agricultural circulars'),
        findsOneWidget,
      );

      // Explanatory card
      expect(find.text('Subscriber Access Only'), findsOneWidget);
      expect(
        find.text(
          'Access to TAGA official resources and government orders is reserved for active paid subscribers. Please complete or renew your annual subscription to view resources.',
        ),
        findsOneWidget,
      );
      expect(find.byIcon(LucideIcons.alertCircle), findsOneWidget);

      // Benefits list
      expect(
        find.text('Access to official Tamil Nadu Agricultural Engineering GOs'),
        findsOneWidget,
      );
      expect(
        find.text('Scheme guidelines, department orders, and job charts'),
        findsOneWidget,
      );
      expect(
        find.text('Direct download of official leave and application formats'),
        findsOneWidget,
      );

      // Action buttons
      expect(find.text('Manage Subscription'), findsOneWidget);
      expect(find.text('Back to Home'), findsOneWidget);
    });

    testWidgets('calls onUpgrade callback when Manage Subscription is tapped', (tester) async {
      bool upgraded = false;

      await tester.pumpWidget(
        MaterialApp(
          home: ResourcesAccessRestrictedScreen(
            onUpgrade: () {
              upgraded = true;
            },
          ),
        ),
      );

      await tester.pumpAndSettle();

      final manageBtn = find.text('Manage Subscription');
      await tester.ensureVisible(manageBtn);
      await tester.pumpAndSettle();

      await tester.tap(manageBtn);
      await tester.pumpAndSettle();

      expect(upgraded, isTrue);
    });

    testWidgets('calls onBackToHome callback when Back to Home is tapped', (tester) async {
      bool returned = false;

      await tester.pumpWidget(
        MaterialApp(
          home: ResourcesAccessRestrictedScreen(
            onBackToHome: () {
              returned = true;
            },
          ),
        ),
      );

      await tester.pumpAndSettle();

      final backBtn = find.text('Back to Home');
      await tester.ensureVisible(backBtn);
      await tester.pumpAndSettle();

      await tester.tap(backBtn);
      await tester.pumpAndSettle();

      expect(returned, isTrue);
    });

    testWidgets('renders without AppBar when showAppBar is false', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: ResourcesAccessRestrictedScreen(
              showAppBar: false,
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Heading in hero banner should still be present
      expect(find.text('Resources Center'), findsOneWidget);
      expect(find.text('Subscriber Access Only'), findsOneWidget);
    });
  });
}
