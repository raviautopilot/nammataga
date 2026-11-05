import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';
import 'package:nammataga_mobile/features/resources/domain/external_link.dart';
import 'package:nammataga_mobile/features/resources/domain/resource_category.dart';
import 'package:nammataga_mobile/features/resources/domain/resource_document.dart';
import 'package:nammataga_mobile/features/resources/presentation/resources_access_restricted_screen.dart';
import 'package:nammataga_mobile/features/resources/presentation/resources_screen.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/category_chips_bar.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/document_card.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/external_links_view.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/resources_shimmer.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/subcategory_chips_bar.dart';
import 'package:nammataga_mobile/features/resources/providers/resources_provider.dart';

void main() {
  group('CategoryChipsBar Widget Tests', () {
    final testCategories = [
      const ResourceCategory(id: 'establishment', name: 'Establishment'),
      const ResourceCategory(
        id: 'scheme-gos',
        name: 'Scheme G.Os',
        subcategories: ['Central', 'State'],
      ),
      const ResourceCategory(id: 'forms', name: 'Official Forms'),
      const ResourceCategory(id: 'links', name: 'Useful Links'),
    ];

    testWidgets('renders all category chips with names and icons', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CategoryChipsBar(
              categories: testCategories,
              selectedCategoryId: 'establishment',
              onCategorySelected: (_) {},
            ),
          ),
        ),
      );

      expect(find.text('Establishment'), findsOneWidget);
      expect(find.text('Scheme G.Os'), findsOneWidget);
      expect(find.text('Official Forms'), findsOneWidget);
      expect(find.text('Useful Links'), findsOneWidget);

      expect(find.byIcon(LucideIcons.building2), findsOneWidget);
      expect(find.byIcon(LucideIcons.landmark), findsOneWidget);
      expect(find.byIcon(LucideIcons.fileSpreadsheet), findsOneWidget);
      expect(find.byIcon(LucideIcons.link2), findsOneWidget);
    });

    testWidgets('triggers onCategorySelected when a chip is tapped', (tester) async {
      String? selected;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CategoryChipsBar(
              categories: testCategories,
              selectedCategoryId: 'establishment',
              onCategorySelected: (id) => selected = id,
            ),
          ),
        ),
      );

      await tester.tap(find.text('Scheme G.Os'));
      await tester.pumpAndSettle();

      expect(selected, 'scheme-gos');
    });

    testWidgets('returns empty widget when categories list is empty', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CategoryChipsBar(
              categories: const [],
              selectedCategoryId: null,
              onCategorySelected: (_) {},
            ),
          ),
        ),
      );

      expect(find.byType(ListView), findsNothing);
    });
  });

  group('SubcategoryChipsBar Widget Tests', () {
    final subcategories = ['Central', 'State'];

    testWidgets('renders All chip and subcategory chips', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SubcategoryChipsBar(
              subcategories: subcategories,
              selectedSubcategory: null,
              onSubcategorySelected: (_) {},
            ),
          ),
        ),
      );

      expect(find.text('All'), findsOneWidget);
      expect(find.text('Central'), findsOneWidget);
      expect(find.text('State'), findsOneWidget);
      // "All" is active, so it has check icon
      expect(find.byIcon(LucideIcons.check), findsOneWidget);
    });

    testWidgets('tapping All triggers callback with null', (tester) async {
      String? selected = 'Central';

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SubcategoryChipsBar(
              subcategories: subcategories,
              selectedSubcategory: selected,
              onSubcategorySelected: (sub) => selected = sub,
            ),
          ),
        ),
      );

      await tester.tap(find.text('All'));
      await tester.pumpAndSettle();

      expect(selected, isNull);
    });

    testWidgets('tapping specific subcategory triggers callback with its name', (tester) async {
      String? selected;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SubcategoryChipsBar(
              subcategories: subcategories,
              selectedSubcategory: selected,
              onSubcategorySelected: (sub) => selected = sub,
            ),
          ),
        ),
      );

      await tester.tap(find.text('State'));
      await tester.pumpAndSettle();

      expect(selected, 'State');
    });
  });

  group('DocumentCard Widget Tests', () {
    const testDoc = ResourceDocument(
      title: 'Solar Pump Scheme Guidelines 2026',
      year: '2026',
      subcategory: 'State',
      url: '/docs/solar_pump.pdf',
    );

    testWidgets('renders title, year, subcategory and official PDF badge', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: DocumentCard(
              document: testDoc,
              categoryName: 'Scheme G.Os',
            ),
          ),
        ),
      );

      expect(find.text('Solar Pump Scheme Guidelines 2026'), findsOneWidget);
      expect(find.text('2026'), findsOneWidget);
      expect(find.text('State'), findsOneWidget);
      expect(find.text('Official PDF'), findsOneWidget);
      expect(find.text('PDF'), findsOneWidget);
      expect(find.byIcon(LucideIcons.fileText), findsOneWidget);
      expect(find.byIcon(LucideIcons.calendar), findsOneWidget);
      expect(find.byIcon(LucideIcons.tag), findsOneWidget);
      expect(find.byIcon(LucideIcons.shieldCheck), findsOneWidget);
      expect(find.byIcon(LucideIcons.moreVertical), findsOneWidget);
    });

    testWidgets('triggers custom onTap and onActionsTap when provided', (tester) async {
      bool cardTapped = false;
      bool actionsTapped = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: DocumentCard(
              document: testDoc,
              categoryName: 'Scheme G.Os',
              onTap: () => cardTapped = true,
              onActionsTap: () => actionsTapped = true,
            ),
          ),
        ),
      );

      await tester.tap(find.text('Solar Pump Scheme Guidelines 2026'));
      await tester.pumpAndSettle();
      expect(cardTapped, isTrue);

      await tester.tap(find.byIcon(LucideIcons.moreVertical));
      await tester.pumpAndSettle();
      expect(actionsTapped, isTrue);
    });
  });

  group('ResourcesShimmer Widget Tests', () {
    testWidgets('renders shimmer boxes and document card skeletons', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: ResourcesShimmer(
              itemCount: 3,
              includeChips: true,
            ),
          ),
        ),
      );

      expect(find.byType(CategoryChipsShimmer), findsOneWidget);
      expect(find.byType(DocumentCardShimmer), findsNWidgets(3));
      expect(find.byType(ShimmerLoadingBox), findsWidgets);
    });
  });

  group('ResourcesScreen Integration & Access Guard Tests', () {
    testWidgets('shows loading indicator when memberProfile is loading and session not paid', (tester) async {
      final completer = Completer<MemberProfile>();

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authProvider.overrideWith(
              () => _TestAuthNotifier(
                MemberSession(
                  token: 'test-token',
                  role: 'member',
                  isPaid: false,
                  isGuest: false,
                ),
              ),
            ),
            memberProfileProvider.overrideWith((ref) => completer.future),
          ],
          child: const MaterialApp(
            home: ResourcesScreen(),
          ),
        ),
      );

      await tester.pump();
      expect(find.byType(CircularProgressIndicator), findsOneWidget);

      completer.complete(_createMockProfile(isPaid: false));
      await tester.pumpAndSettle();
    });

    testWidgets('renders ResourcesAccessRestrictedScreen when member is not paid', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authProvider.overrideWith(
              () => _TestAuthNotifier(
                MemberSession(
                  token: 'test-token',
                  role: 'member',
                  isPaid: false,
                  isGuest: false,
                ),
              ),
            ),
            memberProfileProvider.overrideWith(
              (ref) async => _createMockProfile(
                isPaid: false,
                subscriptionActive: false,
              ),
            ),
          ],
          child: const MaterialApp(
            home: ResourcesScreen(),
          ),
        ),
      );

      await tester.pumpAndSettle();
      expect(find.byType(ResourcesAccessRestrictedScreen), findsOneWidget);
      expect(find.text('Subscriber Access Only'), findsOneWidget);
    });

    testWidgets('renders full ResourcesScreen when member is paid', (tester) async {
      tester.view.physicalSize = const Size(1080, 2400);
      tester.view.devicePixelRatio = 2.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      final mockCategories = [
        const ResourceCategory(id: 'scheme-gos', name: 'Scheme G.Os', subcategories: ['Central', 'State']),
        const ResourceCategory(id: 'links', name: 'Useful Links'),
      ];

      final mockDocs = [
        const ResourceDocument(
          title: 'Agricultural Mechanization G.O.',
          year: '2026',
          subcategory: 'State',
          url: '/docs/mech.pdf',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authProvider.overrideWith(
              () => _TestAuthNotifier(
                MemberSession(
                  token: 'test-token',
                  role: 'member',
                  isPaid: true,
                  isGuest: false,
                ),
              ),
            ),
            memberProfileProvider.overrideWith(
              (ref) async => _createMockProfile(isPaid: true, subscriptionActive: true),
            ),
            resourceCategoriesProvider.overrideWith((ref) async => mockCategories),
            selectedCategoryProvider.overrideWith(() => _TestSelectedCategoryNotifier('scheme-gos')),
            categoryDocumentsProvider.overrideWith((ref, arg) async => mockDocs),
          ],
          child: const MaterialApp(
            home: ResourcesScreen(),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Resources Center'), findsOneWidget);
      expect(find.text('Government orders, circulars & official applications'), findsOneWidget);
      expect(find.text('1 document available'), findsOneWidget);
      expect(find.text('Scheme G.Os'), findsOneWidget);
      expect(find.text('Useful Links'), findsOneWidget);
      expect(find.text('Central'), findsOneWidget);
      expect(find.text('State'), findsNWidgets(2)); // Subcategory chip and DocumentCard badge
      expect(find.text('Agricultural Mechanization G.O.'), findsOneWidget);
    });

    testWidgets('renders ExternalLinksView when selectedCategory is links', (tester) async {
      tester.view.physicalSize = const Size(1080, 2400);
      tester.view.devicePixelRatio = 2.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      final mockCategories = [
        const ResourceCategory(id: 'scheme-gos', name: 'Scheme G.Os'),
        const ResourceCategory(id: 'links', name: 'Useful Links'),
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authProvider.overrideWith(
              () => _TestAuthNotifier(
                MemberSession(
                  token: 'test-token',
                  role: 'member',
                  isPaid: true,
                  isGuest: false,
                ),
              ),
            ),
            memberProfileProvider.overrideWith(
              (ref) async => _createMockProfile(isPaid: true, subscriptionActive: true),
            ),
            resourceCategoriesProvider.overrideWith((ref) async => mockCategories),
            selectedCategoryProvider.overrideWith(() => _TestSelectedCategoryNotifier('links')),
            externalLinksProvider.overrideWith(
              (ref) async => [
                const ExternalLink(
                  title: 'Agri Dept Portal',
                  url: 'https://www.tnagrisnet.tn.gov.in/',
                ),
              ],
            ),
          ],
          child: const MaterialApp(
            home: ResourcesScreen(),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.byType(ExternalLinksView), findsOneWidget);
      expect(find.text('Agri Dept Portal'), findsOneWidget);
    });
  });
}

class _TestAuthNotifier extends AuthNotifier {
  final MemberSession _session;
  _TestAuthNotifier(this._session);

  @override
  Future<MemberSession> build() async => _session;
}

class _TestSelectedCategoryNotifier extends SelectedCategoryNotifier {
  final String _initial;
  _TestSelectedCategoryNotifier(this._initial);

  @override
  String? build() => _initial;
}

MemberProfile _createMockProfile({
  required bool isPaid,
  bool subscriptionActive = false,
}) {
  return MemberProfile(
    id: '1',
    tagaId: 'TAGA-001',
    username: 'test_user',
    name: 'Test Member',
    initial: 'A',
    gender: 'Male',
    dateOfBirth: '1990-01-01',
    fatherName: 'Father',
    motherName: 'Mother',
    mobileNumber: '9876543210',
    emailId: 'test@nammataga.in',
    designation: 'Assistant Engineer',
    recruitmentBatch: '2020',
    seniorityNumber: '10',
    workingDistrict: 'Chennai',
    nativeDistrict: 'Chennai',
    residentialAddress: 'Chennai',
    permanentAddress: 'Chennai',
    educationalQualification: 'B.E Agriculture',
    tbfNumber: 'TBF123',
    cpsGpfNumber: 'CPS123',
    paymentStatus: isPaid ? 'paid' : 'unpaid',
    isPaid: isPaid,
    subscriptionActive: subscriptionActive,
  );
}
