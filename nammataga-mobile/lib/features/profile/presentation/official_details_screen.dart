import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/profile_provider.dart';
import 'widgets/detail_card.dart';
import 'widgets/profile_subpage_scaffold.dart';

class OfficialDetailsScreen extends ConsumerWidget {
  const OfficialDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Official Details',
      watermarkIcon: LucideIcons.fileBadge,
      body: profileAsync.when(
        data: (profile) {
          return DetailCard(
            title: 'Official Details',
            subtitle: 'Your official and educational information.',
            icon: LucideIcons.fileBadge,
            items: [
              DetailItem(
                label: 'Educational Qualification',
                value: profile.educationalQualification,
              ),
              DetailItem(label: 'TBF Number', value: profile.tbfNumber),
              DetailItem(label: 'CPS/GPF Number', value: profile.cpsGpfNumber),
            ],
          );
        },
        loading: () => const Center(
          child: Padding(
            padding: EdgeInsets.all(40),
            child: CircularProgressIndicator(color: AppColors.primary),
          ),
        ),
        error: (error, stack) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Text(
              'Unable to load official details.\n$error',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: AppColors.destructive,
                fontSize: 13,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
