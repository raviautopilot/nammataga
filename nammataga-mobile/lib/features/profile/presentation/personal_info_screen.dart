import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/profile_provider.dart';
import 'widgets/detail_card.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Screen 2: Personal Information
/// Displays authentic member personal information as per backend records.
class PersonalInfoScreen extends ConsumerWidget {
  const PersonalInfoScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Personal Information',
      watermarkIcon: LucideIcons.user,
      body: profileAsync.when(
        data: (profile) {
          final dobDisplay = profile.calculatedAge != null
              ? '${profile.formattedDob} ${profile.calculatedAge}'
              : profile.formattedDob;

          return DetailCard(
            title: 'Personal Information',
            subtitle:
                'Your basic personal details as per your membership records.',
            icon: LucideIcons.user,
            items: [
              DetailItem(
                label: 'Full Name',
                value: profile.fullNameWithInitial,
              ),
              DetailItem(label: 'Gender', value: profile.gender),
              DetailItem(label: 'Date of Birth', value: dobDisplay),
              DetailItem(label: "Father's Name", value: profile.fatherName),
              DetailItem(label: "Mother's Name", value: profile.motherName),
            ],
          );
        },
        loading: () => Container(
          height: 280,
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
              'Unable to load personal information.\n$error',
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
