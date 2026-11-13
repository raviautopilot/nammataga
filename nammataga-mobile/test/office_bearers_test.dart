import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/office_bearers/data/office_bearers_repository.dart';
import 'package:nammataga_mobile/features/office_bearers/domain/district_office_bearer.dart';
import 'package:nammataga_mobile/features/office_bearers/domain/state_executive_member.dart';
import 'package:nammataga_mobile/features/office_bearers/presentation/office_bearers_screen.dart';
import 'package:nammataga_mobile/features/office_bearers/presentation/widgets/district_bearer_row.dart';
import 'package:nammataga_mobile/features/office_bearers/presentation/widgets/state_executive_card.dart';
import 'package:nammataga_mobile/features/office_bearers/providers/office_bearers_provider.dart';

class MockOfficeBearersRepository implements OfficeBearersRepository {
  @override
  Future<List<StateExecutiveMember>> getStateExecutive() async {
    return const [
      StateExecutiveMember(
        sno: 1,
        name: 'L.S. Sasikumar',
        designation: 'State President',
        phone: '9965965860',
        location: 'Dharmapuri',
        qualification: 'M.Sc. (Agri), Ph.D',
        experience: 18,
        description: 'Previously served as state secretary of TAGA',
        image: null,
      ),
      StateExecutiveMember(
        sno: 2,
        name: 'R. Muthuram',
        designation: 'State Secretary',
        phone: '8903304616',
        location: 'Tiruvannamalai',
        qualification: 'B.Sc (Agri)',
        experience: 18,
        description: 'Previously served as president of district TAGA',
        image: null,
      ),
    ];
  }

  @override
  Future<List<String>> getDistricts() async {
    return ['Ariyalur', 'Chengalpattu', 'Salem'];
  }

  @override
  Future<List<DistrictOfficeBearer>> getDistrictBearers(String district) async {
    if (district == 'Ariyalur') {
      return const [
        DistrictOfficeBearer(
          name: 'T.Radhakrishnan',
          title: 'President',
          department: '',
          contact: '9487073750',
        ),
        DistrictOfficeBearer(
          name: 'Tamilmani',
          title: 'Secretary',
          department: '',
          contact: '7502821228',
        ),
      ];
    }
    return [];
  }
}

void main() {
  testWidgets(
      'Office Bearers page renders tabs, state executive cards, and switches to district directory',
      (WidgetTester tester) async {
    final mockRepo = MockOfficeBearersRepository();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          officeBearersRepositoryProvider.overrideWithValue(mockRepo),
        ],
        child: const MaterialApp(
          home: OfficeBearersScreen(),
        ),
      ),
    );

    // Initial render & wait for data loading
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pump(const Duration(milliseconds: 400));

    // Verify Header Title & Tabs
    expect(find.text('TAGA Leadership'), findsOneWidget);
    expect(find.text('Term: 2025–2027'), findsOneWidget);
    expect(find.text('State Executive'), findsOneWidget);
    expect(find.text('District Bearers'), findsOneWidget);

    // Verify State Executive Section Heading & Members
    expect(find.text('State Office Bearers'), findsOneWidget);
    expect(find.text('L.S. Sasikumar'), findsOneWidget);
    expect(find.text('State President'), findsOneWidget);
    expect(find.text('R. Muthuram'), findsOneWidget);
    expect(find.text('State Secretary'), findsOneWidget);

    // Verify StateExecutiveCard uses rectangular portrait card
    expect(find.byType(StateExecutiveCard), findsNWidgets(2));

    // Tap on first member card to open Profile
    await tester.tap(find.text('L.S. Sasikumar'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    // Verify Full Screen Profile details
    expect(find.text('Executive Profile'), findsOneWidget);
    expect(find.text('M.Sc. (Agri), Ph.D'), findsOneWidget);
    expect(find.text('18 Years in Agriculture'), findsOneWidget);
    expect(find.text('Dharmapuri'), findsOneWidget);
    expect(find.text('Call Official'), findsOneWidget);
    expect(find.text('WhatsApp'), findsOneWidget);

    // Pop back to Office Bearers Screen
    await tester.tap(find.byTooltip('Back'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    // Switch to District Bearers Tab
    await tester.tap(find.text('District Bearers'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    // Verify District Selection UI
    expect(find.text('Select District'), findsOneWidget);
    expect(find.text('Select a District Above'), findsOneWidget);

    // Open Dropdown and choose 'Ariyalur'
    await tester.tap(find.text('Choose a district...'), warnIfMissed: false);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('Ariyalur').last, findsOneWidget);
    await tester.tap(find.text('Ariyalur').last);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    // Verify District Directory
    expect(find.text('Ariyalur District Directory'), findsOneWidget);
    expect(find.text('2 Bearers'), findsOneWidget);
    expect(find.text('NAME & DESIGNATION'), findsOneWidget);
    expect(find.text('T.Radhakrishnan'), findsOneWidget);
    expect(find.text('Tamilmani'), findsOneWidget);

    // Verify no circular avatars or images in district rows
    expect(find.byType(DistrictBearerRow), findsNWidgets(2));
    expect(find.byType(CircleAvatar), findsNothing);

    // Tap on a District Bearer row to open details
    await tester.tap(find.text('T.Radhakrishnan'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    // Verify District Bearer Profile Screen (No photos, clean structured info)
    expect(find.text('Bearer Details'), findsOneWidget);
    expect(find.text('Ariyalur District'), findsOneWidget);
    expect(find.text('T.Radhakrishnan'), findsOneWidget);
    expect(find.text('Call Official'), findsOneWidget);
    expect(find.text('WhatsApp'), findsOneWidget);

    // Back to directory
    await tester.tap(find.byTooltip('Back'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    expect(find.text('Ariyalur District Directory'), findsOneWidget);
  });
}
