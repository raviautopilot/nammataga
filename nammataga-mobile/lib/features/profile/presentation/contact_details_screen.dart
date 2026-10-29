import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/profile_provider.dart';
import 'widgets/detail_card.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Screen 3: Contact Details
/// Displays authenticated member communication information.
class ContactDetailsScreen extends ConsumerWidget {
  const ContactDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Contact Details',
      watermarkIcon: LucideIcons.phone,
      body: profileAsync.when(
        data: (profile) {
          return DetailCard(
            title: 'Contact Details',
            subtitle: 'Your contact information for communication.',
            icon: LucideIcons.phoneCall,
            items: [
              DetailItem(
                label: 'Mobile Number',
                value: profile.mobileNumber,
                icon: LucideIcons.phone,
              ),
              DetailItem(
                label: 'Email Address',
                value: profile.emailId,
                icon: LucideIcons.mail,
              ),
            ],
          );
        },
        loading: () => Container(
          height: 180,
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
              'Unable to load contact details.\n$error',
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
