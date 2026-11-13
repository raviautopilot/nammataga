import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/office_bearers_repository.dart';
import '../domain/district_office_bearer.dart';
import '../domain/state_executive_member.dart';

/// Repository provider
final officeBearersRepositoryProvider =
    Provider<OfficeBearersRepository>((ref) {
  return OfficeBearersRepository();
});

/// State Executive members provider
final stateExecutiveProvider =
    FutureProvider<List<StateExecutiveMember>>((ref) async {
  final repo = ref.watch(officeBearersRepositoryProvider);
  return repo.getStateExecutive();
});

/// Districts list provider
final bearerDistrictsProvider = FutureProvider<List<String>>((ref) async {
  final repo = ref.watch(officeBearersRepositoryProvider);
  return repo.getDistricts();
});

/// Notifier managing selected district
class SelectedDistrictNotifier extends Notifier<String?> {
  @override
  String? build() => null;

  void selectDistrict(String? district) => state = district;
  void clear() => state = null;
}

/// Selected district state provider
final selectedDistrictProvider =
    NotifierProvider<SelectedDistrictNotifier, String?>(
  SelectedDistrictNotifier.new,
);

/// District bearers family provider keyed by district name
final districtBearersFamily =
    FutureProvider.family<List<DistrictOfficeBearer>, String>(
        (ref, district) async {
  if (district.isEmpty) return [];
  final repo = ref.watch(officeBearersRepositoryProvider);
  return repo.getDistrictBearers(district);
});
