import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/profile_provider.dart';
import 'widgets/detail_card.dart';
import 'widgets/profile_subpage_scaffold.dart';

class EmploymentDetailsScreen extends ConsumerWidget {
  const EmploymentDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Employment Details',
      watermarkIcon: LucideIcons.briefcase,
      body: profileAsync.when(
        data: (profile) {
          return DetailCard(
            title: 'Employment Details',
            subtitle: 'Your current employment information.',
            icon: LucideIcons.briefcase,
            items: [
              DetailItem(label: 'Designation', value: profile.designation),
              DetailItem(
                label: 'Recruitment Batch',
                value: profile.recruitmentBatch,
              ),
              DetailItem(
                label: 'Seniority Number',
                value: profile.seniorityNumber,
              ),
              DetailItem(
                label: 'Working District',
                value: profile.workingDistrict,
              ),
              DetailItem(
                label: 'Native District',
                value: profile.nativeDistrict,
              ),
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
              'Unable to load employment details.\n$error',
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
