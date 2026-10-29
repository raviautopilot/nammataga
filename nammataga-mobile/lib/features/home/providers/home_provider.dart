import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/home_repository.dart';
import '../domain/association_info.dart';

final homeRepositoryProvider = Provider<HomeRepository>((ref) {
  return HomeRepository();
});

final aboutInfoProvider = FutureProvider<AboutData>((ref) async {
  final repo = ref.watch(homeRepositoryProvider);
  return repo.getAboutInfo();
});

final statsProvider = FutureProvider<List<StatData>>((ref) async {
  final repo = ref.watch(homeRepositoryProvider);
  return repo.getStats();
});

final objectivesProvider = FutureProvider<List<ObjectiveData>>((ref) async {
  final repo = ref.watch(homeRepositoryProvider);
  return repo.getObjectives();
});

final contactInfoProvider = FutureProvider<ContactData>((ref) async {
  final repo = ref.watch(homeRepositoryProvider);
  return repo.getContactInfo();
});
