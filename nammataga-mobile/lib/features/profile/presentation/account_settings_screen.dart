import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../auth/presentation/widgets/change_password_sheet.dart';
import '../../auth/providers/auth_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Screen 10: Account & Settings
/// Features the dark emerald header with gear watermark, structured settings rows,
/// separate red logout action, and association version footer matching reference.
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
      body: Column(
        children: [
          // 1. Main Settings Card matching reference
          Container(
            width: double.infinity,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x061B4332),
                  blurRadius: 16,
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
                const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
                _buildSettingRow(
                  icon: LucideIcons.megaphone,
                  title: 'Announcements',
                  onTap: () {
                    context.push('/profile/announcements');
                  },
                ),
                const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
                _buildSettingRow(
                  icon: LucideIcons.headphones,
                  title: 'Help & Support',
                  onTap: () => _showHelpSheet(context),
                ),
                const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
                _buildSettingRow(
                  icon: LucideIcons.info,
                  title: 'About NammaTAGA',
                  onTap: () => _showAboutSheet(context),
                ),
                const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
                _buildSettingRow(
                  icon: LucideIcons.shieldCheck,
                  title: 'Privacy Policy',
                  onTap: () => _showPrivacySheet(context),
                ),
                const Divider(height: 1, indent: 58, color: Color(0xFFEFF4F0)),
                _buildSettingRow(
                  icon: LucideIcons.fileText,
                  title: 'Terms & Conditions',
                  onTap: () => _showTermsSheet(context),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // 2. Destructive Logout Card matching reference
          Container(
            width: double.infinity,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: const Color(0xFFFEE2E2), width: 1.1),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x04DC2626),
                  blurRadius: 14,
                  offset: Offset(0, 4),
                ),
              ],
            ),
            child: _buildSettingRow(
              icon: LucideIcons.logOut,
              title: 'Logout',
              isDestructive: true,
              onTap: () => _confirmSignOut(context, ref),
            ),
          ),
          const SizedBox(height: 24),

          // 3. App Version Footer
          Column(
            children: const [
              Text(
                'NammaTAGA Mobile • Version 1.0.0 (Build 100)',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF64748B),
                ),
              ),
              SizedBox(height: 4),
              Text(
                'Tamil Nadu Agricultural Graduate Association (Estd. 1929)',
                style: TextStyle(
                  fontSize: 11,
                  color: Color(0xFF94A3B8),
                ),
              ),
            ],
          ),
        ],
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
      onTap: () {
        HapticFeedback.lightImpact();
        onTap();
      },
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 15),
        child: Row(
          children: [
            Container(
              width: 38,
              height: 38,
              decoration: BoxDecoration(
                color: isDestructive
                    ? const Color(0xFFFEE2E2)
                    : const Color(0xFFE8F5EE),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: isDestructive
                      ? const Color(0xFFFECACA)
                      : const Color(0xFFCBE3D5),
                  width: 1,
                ),
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
                  fontSize: 14.5,
                  fontWeight: FontWeight.w600,
                  color: isDestructive
                      ? AppColors.destructive
                      : AppColors.textPrimary,
                  letterSpacing: -0.1,
                ),
              ),
            ),
            Icon(
              LucideIcons.chevronRight,
              size: 18,
              color: isDestructive
                  ? const Color(0xFFF87171)
                  : const Color(0xFF9CA3AF),
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
            Container(
              width: 40,
              height: 4,
              margin: const EdgeInsets.only(bottom: 16),
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFE2E8F0),
                borderRadius: BorderRadius.circular(2),
              ),
            ),
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
              leading: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5EE),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(LucideIcons.phone, color: AppColors.primary, size: 18),
              ),
              title: const Text('Phone Helpline', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
              subtitle: const Text('044 34919949'),
              onTap: () => _launch('tel:04434919949'),
            ),
            ListTile(
              leading: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5EE),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(LucideIcons.mail, color: AppColors.primary, size: 18),
              ),
              title: const Text('General Inquiries', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
              subtitle: const Text('tagastate@gmail.com'),
              onTap: () => _launch('mailto:tagastate@gmail.com'),
            ),
            ListTile(
              leading: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5EE),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(LucideIcons.globe, color: AppColors.primary, size: 18),
              ),
              title: const Text('Official Portal', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
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
              'Tamil Nadu Government Agricultural Graduates Association (TAGA) is the apex professional body representing agricultural science graduates of the state agriculture department since 1929.\n\nNammaTAGA Mobile Member Portal v1.0.0 (Build 100)',
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
        title: const Text('Confirm Logout'),
        content: const Text(
          'Are you sure you want to log out of your NammaTAGA account?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.destructive,
              foregroundColor: Colors.white,
            ),
            onPressed: () async {
              Navigator.pop(ctx);
              await ref.read(authProvider.notifier).logout();
              if (context.mounted) {
                context.go('/auth/login');
              }
            },
            child: const Text('Logout'),
          ),
        ],
      ),
    );
  }
}
