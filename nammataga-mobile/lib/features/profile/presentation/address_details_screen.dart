import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/profile_provider.dart';
import 'widgets/detail_card.dart';
import 'widgets/profile_subpage_scaffold.dart';

class AddressDetailsScreen extends ConsumerWidget {
  const AddressDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(memberProfileProvider);

    return ProfileSubpageScaffold(
      title: 'Address Details',
      watermarkIcon: LucideIcons.mapPin,
      body: profileAsync.when(
        data: (profile) {
          return DetailCard(
            title: 'Address Details',
            subtitle: 'Your address information.',
            icon: LucideIcons.mapPin,
            customContent: Column(
              children: [
                _buildAddressBox(
                  label: 'Residential Address',
                  address: profile.residentialAddress,
                ),
                const SizedBox(height: 14),
                _buildAddressBox(
                  label: 'Permanent Address',
                  address: profile.permanentAddress,
                ),
              ],
            ),
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
              'Unable to load address details.\n$error',
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

  Widget _buildAddressBox({required String label, required String address}) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFF9FAF7),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE5ECE7), width: 1),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: const Color(0xFFE8F5EE),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(
              LucideIcons.home,
              size: 16,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: const TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  address.isNotEmpty ? address : '—',
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: AppColors.textSecondary,
                    height: 1.45,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
