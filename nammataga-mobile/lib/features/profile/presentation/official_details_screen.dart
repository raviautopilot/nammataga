import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/profile_provider.dart';
import 'widgets/detail_card.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Screen 6: Official Details
/// Displays authenticated member official numbers and educational qualifications.
class OfficialDetailsScreen extends ConsumerWidget {
  const OfficialDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Official Details',
      watermarkIcon: LucideIcons.fileText,
      body: profileAsync.when(
        data: (profile) {
          return DetailCard(
            title: 'Official Details',
            subtitle: 'Your official and educational information.',
            icon: LucideIcons.fileCheck,
            items: [
              DetailItem(
                label: 'Educational Qualification',
                value: profile.educationalQualification,
              ),
              DetailItem(
                label: 'TBF Number',
                value: profile.tbfNumber,
              ),
              DetailItem(
                label: 'CPS/GPF Number',
                value: profile.cpsGpfNumber,
              ),
            ],
          );
        },
        loading: () => Container(
          height: 220,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFE2EBE5)),
          ),
          child: const Center(
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
