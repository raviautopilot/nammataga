import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../auth/presentation/widgets/change_password_sheet.dart';
import '../../auth/providers/auth_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

class AccountSettingsScreen extends ConsumerWidget {
  const AccountSettingsScreen({super.key});

  Future<void> _launch(String urlString) async {
    final uri = Uri.parse(urlString);
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      }
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessionAsync = ref.watch(authProvider);
    final email = sessionAsync.value?.memberEmail ?? '';

    return ProfileSubpageScaffold(
      title: 'Account & Settings',
      watermarkIcon: LucideIcons.settings,
      showEditNotice: false,
      body: Container(
        width: double.infinity,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
          boxShadow: const [
            BoxShadow(
              color: Color(0x061B4332),
              blurRadius: 14,
              offset: Offset(0, 4),
            ),
          ],
        ),
        child: Column(
          children: [
            _buildSettingRow(
              icon: LucideIcons.lock,
              title: 'Change Password',
              onTap: () {
                ChangePasswordSheet.show(context, email: email);
              },
            ),
            const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            _buildSettingRow(
              icon: LucideIcons.megaphone,
              title: 'Announcements',
              onTap: () {
                context.push('/profile/announcements');
              },
            ),
            const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            _buildSettingRow(
              icon: LucideIcons.headphones,
              title: 'Help & Support',
              onTap: () => _showHelpSheet(context),
            ),
            const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            _buildSettingRow(
              icon: LucideIcons.info,
              title: 'About NammaTAGA',
              onTap: () => _showAboutSheet(context),
            ),
            const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            _buildSettingRow(
              icon: LucideIcons.shieldCheck,
              title: 'Privacy Policy',
              onTap: () => _showPrivacySheet(context),
            ),
            const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            _buildSettingRow(
              icon: LucideIcons.fileText,
              title: 'Terms & Conditions',
              onTap: () => _showTermsSheet(context),
            ),
            const Divider(height: 1, indent: 54, color: Color(0xFFEFF4F0)),
            _buildSettingRow(
              icon: LucideIcons.logOut,
              title: 'Logout',
              isDestructive: true,
              onTap: () => _confirmSignOut(context, ref),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSettingRow({
    required IconData icon,
    required String title,
    required VoidCallback onTap,
    bool isDestructive = false,
  }) {
    return BouncyButton(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 15),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(7),
              decoration: BoxDecoration(
                color: isDestructive
                    ? const Color(0xFFFEE2E2)
                    : const Color(0xFFE8F5EE),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                icon,
                size: 18,
                color: isDestructive
                    ? AppColors.destructive
                    : AppColors.primary,
              ),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Text(
                title,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: isDestructive
                      ? AppColors.destructive
                      : AppColors.textPrimary,
                ),
              ),
            ),
            if (!isDestructive)
              const Icon(
                LucideIcons.chevronRight,
                size: 17,
                color: Color(0xFF9CA3AF),
              ),
          ],
        ),
      ),
    );
  }

  void _showHelpSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Help & Support',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 6),
            const Text(
              'Get in touch with the state executive office for any membership or technical queries.',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
            ),
            const SizedBox(height: 18),
            ListTile(
              leading: const Icon(LucideIcons.phone, color: AppColors.primary),
              title: const Text('Phone Helpline'),
              subtitle: const Text('044 34919949'),
              onTap: () => _launch('tel:04434919949'),
            ),
            ListTile(
              leading: const Icon(LucideIcons.mail, color: AppColors.primary),
              title: const Text('General Inquiries'),
              subtitle: const Text('tagastate@gmail.com'),
              onTap: () => _launch('mailto:tagastate@gmail.com'),
            ),
            ListTile(
              leading: const Icon(LucideIcons.globe, color: AppColors.primary),
              title: const Text('Official Portal'),
              subtitle: const Text('www.nammataga.com'),
              onTap: () => _launch('https://www.nammataga.com'),
            ),
          ],
        ),
      ),
    );
  }

  void _showAboutSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: const [
            Text(
              'About NammaTAGA',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 10),
            Text(
              'Tamil Nadu Government Agricultural Graduates Association (TAGA) is the apex professional body representing agricultural science graduates of the state agriculture department since 1929.\n\nNammaTAGA Mobile Member Portal v1.0.0',
              style: TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.5,
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showPrivacySheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: const [
            Text(
              'Privacy Policy',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 10),
            Text(
              'NammaTAGA protects member data with encryption and strict role-based access. Your government service records and contact credentials are used solely for verified association affairs and state communications.',
              style: TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.5,
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showTermsSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: const [
            Text(
              'Terms & Conditions',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 10),
            Text(
              'Use of the NammaTAGA mobile portal is governed by association bylaws. Subscribed members agree to uphold the ethics and collective interests of agricultural officers in Tamil Nadu.',
              style: TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.5,
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _confirmSignOut(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Sign Out'),
        content: const Text(
          'Are you sure you want to sign out of your TAGA member account?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.destructive,
            ),
            onPressed: () async {
              Navigator.of(ctx).pop();
              await ref.read(authProvider.notifier).logout();
              if (context.mounted) {
                context.go('/login');
              }
            },
            child: const Text(
              'Sign Out',
              style: TextStyle(color: Colors.white),
            ),
          ),
        ],
      ),
    );
  }
}
