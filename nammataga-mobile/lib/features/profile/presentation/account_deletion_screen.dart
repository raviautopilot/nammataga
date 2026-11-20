import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/storage/secure_storage_service.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../../../core/widgets/taga_toast.dart';
import '../../auth/providers/auth_provider.dart';
import '../domain/deletion_request_model.dart';
import '../providers/profile_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

/// Screen 11: Account Deletion (Apple Guideline 5.1.1(v) Compliant)
/// Allows members to review association deletion policies, verify their password,
/// and submit an account deletion request for administrator review.
class AccountDeletionScreen extends ConsumerStatefulWidget {
  const AccountDeletionScreen({super.key});

  @override
  ConsumerState<AccountDeletionScreen> createState() =>
      _AccountDeletionScreenState();
}

class _AccountDeletionScreenState extends ConsumerState<AccountDeletionScreen> {
  final _formKey = GlobalKey<FormState>();
  final _passwordController = TextEditingController();
  final _reasonController = TextEditingController();

  bool _showPassword = false;
  bool _agreedToTerms = false;
  bool _isSubmitting = false;
  bool _overrideRejectedView = false;

  @override
  void dispose() {
    _passwordController.dispose();
    _reasonController.dispose();
    super.dispose();
  }

  String _formatDate(DateTime? dt) {
    if (dt == null) return 'N/A';
    return DateFormat('dd MMM yyyy, hh:mm a').format(dt);
  }

  Future<void> _handleLogout() async {
    HapticFeedback.mediumImpact();
    await StorageService.clearAuth();
    if (!mounted) return;
    await ref.read(authProvider.notifier).logout();
    if (mounted) {
      context.go('/login');
    }
  }

  Future<void> _handleSubmitDeletion() async {
    if (!_formKey.currentState!.validate()) {
      HapticFeedback.lightImpact();
      return;
    }
    if (!_agreedToTerms) {
      TagaToast.error(
        context,
        'Please confirm that you understand account deletion is permanent.',
        title: 'Confirmation Required',
      );
      return;
    }

    FocusScope.of(context).unfocus();
    HapticFeedback.heavyImpact();

    // Show initial confirmation dialog before network submission
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
        ),
        backgroundColor: Colors.white,
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFFFEE2E2),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                LucideIcons.alertTriangle,
                color: AppColors.destructive,
                size: 24,
              ),
            ),
            const SizedBox(width: 12),
            const Expanded(
              child: Text(
                'Confirm Deletion',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ),
          ],
        ),
        content: const Text(
          'Are you sure you want to submit an account deletion request? This action is permanent and cannot be undone once submitted.',
          style: TextStyle(
            fontSize: 14,
            height: 1.5,
            color: Color(0xFF475569),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text(
              'Cancel',
              style: TextStyle(
                color: Color(0xFF64748B),
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.destructive,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(10),
              ),
            ),
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Yes, Submit'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    setState(() => _isSubmitting = true);

    try {
      final repo = ref.read(profileRepositoryProvider);
      final fullReason = _reasonController.text.trim();

      await repo.submitAccountDeletionRequest(
        password: _passwordController.text,
        reason: fullReason,
      );

      if (!mounted) return;

      // Show final confirmation dialog then immediately sign out
      await showDialog<void>(
        context: context,
        barrierDismissible: false,
        builder: (ctx) => AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          backgroundColor: Colors.white,
          title: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFFEF3C7),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(
                  LucideIcons.clock,
                  color: Color(0xFFD97706),
                  size: 24,
                ),
              ),
              const SizedBox(width: 12),
              const Expanded(
                child: Text(
                  'Account Deletion Requested',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF92400E),
                  ),
                ),
              ),
            ],
          ),
          content: const Text(
            'Your account deletion request has been submitted successfully and is awaiting administrator review. For security reasons, your active session is concluded and you will now be signed out.',
            style: TextStyle(
              fontSize: 14,
              height: 1.5,
              color: Color(0xFF475569),
            ),
          ),
          actions: [
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFD97706),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                onPressed: () => Navigator.of(ctx).pop(),
                child: const Text(
                  'Sign Out Now',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
          ],
        ),
      );

      // Force sign-out and clear authentication tokens
      await _handleLogout();
    } catch (e) {
      if (mounted) {
        final rawErr = e.toString().replaceAll('Exception: ', '').trim();
        TagaToast.error(
          context,
          rawErr,
          title: 'Request Failed',
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final statusAsync = ref.watch(accountDeletionStatusProvider);
    final eligibilityAsync = ref.watch(accountDeletionEligibilityProvider);

    return ProfileSubpageScaffold(
      title: 'Account Deletion',
      watermarkIcon: LucideIcons.userX,
      showEditNotice: false,
      body: statusAsync.when(
        loading: () => const Center(
          child: Padding(
            padding: EdgeInsets.symmetric(vertical: 60),
            child: CircularProgressIndicator(color: AppColors.primary),
          ),
        ),
        error: (err, stack) => _buildErrorCard(err.toString()),
        data: (status) {
          // 1. Pending State: User already has a pending deletion request
          if (status.isPending && !_overrideRejectedView) {
            return _buildPendingStateCard(status);
          }

          // 2. Rejected State: Previous request was rejected
          if (status.isRejected && !_overrideRejectedView) {
            return _buildRejectedStateCard(status);
          }

          // 3. Normal / Active Stay State
          final eligibility = eligibilityAsync.value ??
              const DeletionEligibility(canDelete: true, hasActiveStay: false);

          if (eligibilityAsync.isLoading && eligibilityAsync.value == null) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.symmetric(vertical: 60),
                child: CircularProgressIndicator(color: AppColors.primary),
              ),
            );
          }

          return _buildDeletionFormView(eligibility);
        },
      ),
    );
  }

  Widget _buildErrorCard(String errorMsg) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF2F2),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFFCA5A5)),
      ),
      child: Column(
        children: [
          const Icon(
            LucideIcons.alertCircle,
            color: AppColors.destructive,
            size: 32,
          ),
          const SizedBox(height: 12),
          Text(
            errorMsg.replaceAll('Exception: ', ''),
            textAlign: TextAlign.center,
            style: const TextStyle(
              fontSize: 13.5,
              color: Color(0xFF991B1B),
              height: 1.4,
            ),
          ),
          const SizedBox(height: 16),
          BouncyButton(
            onTap: () {
              ref.invalidate(accountDeletionStatusProvider);
              ref.invalidate(accountDeletionEligibilityProvider);
            },
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: const Color(0xFFFCA5A5)),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(LucideIcons.refreshCw, size: 14, color: AppColors.destructive),
                  SizedBox(width: 8),
                  Text(
                    'Retry',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.bold,
                      color: AppColors.destructive,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPendingStateCard(AccountDeletionRequestStatus status) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Pending Status Hero Card
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFBEB),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFFCD34D), width: 1.2),
            boxShadow: const [
              BoxShadow(
                color: Color(0x08D97706),
                blurRadius: 16,
                offset: Offset(0, 4),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEF3C7),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: const Color(0xFFFDE68A)),
                    ),
                    child: const Icon(
                      LucideIcons.clock,
                      color: Color(0xFFD97706),
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 14),
                  const Expanded(
                    child: Text(
                      'Request Under Review',
                      style: TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF92400E),
                        letterSpacing: -0.2,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              // Neutral Pending Review Message
              const Text(
                'Your account deletion request is currently awaiting administrator review. Login access remains paused during this period. You will be notified once reviewed.',
                style: TextStyle(
                  fontSize: 14,
                  height: 1.5,
                  fontWeight: FontWeight.w500,
                  color: Color(0xFF78350F),
                ),
              ),
              const SizedBox(height: 20),
              const Divider(height: 1, color: Color(0xFFFDE68A)),
              const SizedBox(height: 16),
              // Request metadata
              if (status.createdAt != null) ...[
                _buildMetaRow('Requested On', _formatDate(status.createdAt)),
                const SizedBox(height: 8),
              ],
              if (status.reason != null && status.reason!.isNotEmpty) ...[
                _buildMetaRow('Reason', status.reason!),
                const SizedBox(height: 8),
              ],
              _buildMetaRow('Status', 'Pending Administrator Review'),
            ],
          ),
        ),
        const SizedBox(height: 20),

        // Irrevocable Warning & No Withdrawal Notice
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: const Color(0xFFF8FAFC),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: const Color(0xFFE2E8F0)),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const [
              Icon(
                LucideIcons.info,
                size: 20,
                color: Color(0xFF64748B),
              ),
              SizedBox(width: 12),
              Expanded(
                child: Text(
                  'Per association policy, deletion requests cannot be withdrawn once submitted. TAGA administrators will inspect active duties, bookings, and financial clearances before final processing.',
                  style: TextStyle(
                    fontSize: 12.5,
                    height: 1.45,
                    color: Color(0xFF475569),
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),

        // Logout Action Button (NO WITHDRAWAL BUTTON)
        BouncyButton(
          onTap: _handleLogout,
          child: Container(
            width: double.infinity,
            height: 52,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: const Color(0xFFFEE2E2), width: 1.2),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x04DC2626),
                  blurRadius: 10,
                  offset: Offset(0, 3),
                ),
              ],
            ),
            child: const Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  LucideIcons.logOut,
                  size: 18,
                  color: AppColors.destructive,
                ),
                SizedBox(width: 10),
                Text(
                  'Sign Out Now',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: AppColors.destructive,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildRejectedStateCard(AccountDeletionRequestStatus status) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Rejection Card
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: const Color(0xFFFEF2F2),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFFCA5A5), width: 1.2),
            boxShadow: const [
              BoxShadow(
                color: Color(0x0AE11D48),
                blurRadius: 16,
                offset: Offset(0, 4),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEE2E2),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: const Color(0xFFFECACA)),
                    ),
                    child: const Icon(
                      LucideIcons.xCircle,
                      color: AppColors.destructive,
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 14),
                  const Expanded(
                    child: Text(
                      'Request Rejected',
                      style: TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF991B1B),
                        letterSpacing: -0.2,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              const Text(
                'Your previous account deletion request was rejected by an administrator with the following remarks:',
                style: TextStyle(
                  fontSize: 13.5,
                  height: 1.5,
                  color: Color(0xFF7F1D1D),
                ),
              ),
              const SizedBox(height: 14),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFFECACA)),
                ),
                child: Text(
                  status.adminRemarks != null && status.adminRemarks!.isNotEmpty
                      ? status.adminRemarks!
                      : 'No specific remarks provided by administrator.',
                  style: const TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF1E293B),
                    height: 1.45,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              if (status.processedAt != null) ...[
                _buildMetaRow('Reviewed On', _formatDate(status.processedAt)),
                const SizedBox(height: 8),
              ],
              if (status.createdAt != null) ...[
                _buildMetaRow('Submitted On', _formatDate(status.createdAt)),
              ],
            ],
          ),
        ),
        const SizedBox(height: 24),

        // Button to submit new request
        BouncyButton(
          onTap: () {
            setState(() {
              _overrideRejectedView = true;
            });
          },
          child: Container(
            width: double.infinity,
            height: 52,
            decoration: BoxDecoration(
              color: const Color(0xFF0D2C1B),
              borderRadius: BorderRadius.circular(14),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x1F0D2C1B),
                  blurRadius: 12,
                  offset: Offset(0, 4),
                ),
              ],
            ),
            child: const Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  LucideIcons.refreshCw,
                  size: 18,
                  color: Colors.white,
                ),
                SizedBox(width: 10),
                Text(
                  'Submit New Request',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDeletionFormView(DeletionEligibility eligibility) {
    final hasActiveStay = eligibility.hasActiveStay;

    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 1. Active Stay Blocked Warning Card (if applicable)
          if (hasActiveStay) ...[
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFBEB),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xFFFCD34D), width: 1.2),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x08D97706),
                    blurRadius: 12,
                    offset: Offset(0, 3),
                  ),
                ],
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFEF3C7),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      LucideIcons.alertTriangle,
                      color: Color(0xFFD97706),
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  const Expanded(
                    child: Text(
                      'You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.',
                      style: TextStyle(
                        fontSize: 13.5,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF92400E),
                        height: 1.45,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
          ],

          // 2. Disclosures Checklist Card
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
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
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: const Color(0xFFFEE2E2),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: const Icon(
                        LucideIcons.shieldAlert,
                        size: 20,
                        color: AppColors.destructive,
                      ),
                    ),
                    const SizedBox(width: 12),
                    const Expanded(
                      child: Text(
                        'Important Information Before You Proceed',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                          letterSpacing: -0.2,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                _buildDisclosureItem(
                  'Account Access During Review: Once your request is submitted, your account login will be disabled while under review. You will not be able to log into your account unless the request is rejected by an administrator.',
                ),
                _buildDisclosureItem(
                  'TAGA Towers Check-in Status: Account deletion cannot be submitted if you are currently checked in at TAGA Towers. Please complete your checkout before requesting deletion.',
                ),
                _buildDisclosureItem(
                  'Subscriber & Association Services: Access to TAGA subscriber benefits, active subscription services, and association features will end once the deletion is processed.',
                ),
                _buildDisclosureItem(
                  'Administrator Review: Requests are reviewed and verified by association administrators before final processing.',
                ),
                _buildDisclosureItem(
                  'Request Finality: Once submitted, an account deletion request cannot be cancelled or withdrawn.',
                  isCritical: true,
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // 3. Deletion Request Form Inputs
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
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
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Reason for Deletion',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                TextFormField(
                  controller: _reasonController,
                  enabled: !hasActiveStay && !_isSubmitting,
                  maxLines: 3,
                  validator: (value) {
                    if (value == null || value.trim().isEmpty) {
                      return 'Please enter your reason for account deletion';
                    }
                    return null;
                  },
                  decoration: InputDecoration(
                    hintText: 'Please write your reason for requesting account deletion...',
                    hintStyle: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF94A3B8),
                    ),
                    filled: true,
                    fillColor: const Color(0xFFF8FAFC),
                    contentPadding: const EdgeInsets.all(14),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                    ),
                    focusedBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: const BorderSide(
                        color: AppColors.primary,
                        width: 1.5,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 18),

                // Password confirmation field
                const Text(
                  'Confirm Password',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                const Text(
                  'Enter your login password to confirm identity.',
                  style: TextStyle(
                    fontSize: 12,
                    color: Color(0xFF64748B),
                  ),
                ),
                const SizedBox(height: 8),
                TextFormField(
                  controller: _passwordController,
                  enabled: !hasActiveStay && !_isSubmitting,
                  obscureText: !_showPassword,
                  validator: (value) {
                    if (value == null || value.trim().isEmpty) {
                      return 'Password is required to request deletion';
                    }
                    return null;
                  },
                  decoration: InputDecoration(
                    hintText: 'Enter account password',
                    hintStyle: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF94A3B8),
                    ),
                    filled: true,
                    fillColor: const Color(0xFFF8FAFC),
                    prefixIcon: const Icon(
                      LucideIcons.lock,
                      size: 18,
                      color: Color(0xFF64748B),
                    ),
                    suffixIcon: IconButton(
                      icon: Icon(
                        _showPassword ? LucideIcons.eyeOff : LucideIcons.eye,
                        size: 18,
                        color: const Color(0xFF64748B),
                      ),
                      onPressed: () {
                        setState(() => _showPassword = !_showPassword);
                      },
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 14,
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                    ),
                    focusedBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: const BorderSide(
                        color: AppColors.primary,
                        width: 1.5,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 18),

                // Confirmation Checkbox
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      width: 24,
                      height: 24,
                      child: Checkbox(
                        value: _agreedToTerms,
                        activeColor: AppColors.destructive,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(6),
                        ),
                        onChanged: hasActiveStay || _isSubmitting
                            ? null
                            : (val) {
                                setState(() => _agreedToTerms = val ?? false);
                              },
                      ),
                    ),
                    const SizedBox(width: 12),
                    const Expanded(
                      child: Text(
                        'I understand that submitting this request will initiate account closure, and once submitted, this request cannot be cancelled or withdrawn.',
                        style: TextStyle(
                          fontSize: 12.5,
                          height: 1.45,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF334155),
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Submit Button (Destructive Red)
          BouncyButton(
            onTap: (hasActiveStay || !_agreedToTerms || _isSubmitting)
                ? () {
                    if (hasActiveStay) {
                      TagaToast.error(
                        context,
                        'Account deletion is blocked while you have an active stay at TAGA Towers.',
                        title: 'Active Stay Active',
                      );
                    } else if (!_agreedToTerms) {
                      TagaToast.error(
                        context,
                        'Please check the confirmation box to proceed.',
                        title: 'Confirmation Required',
                      );
                    }
                  }
                : _handleSubmitDeletion,
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: double.infinity,
              height: 52,
              decoration: BoxDecoration(
                color: (hasActiveStay || !_agreedToTerms || _isSubmitting)
                    ? const Color(0xFFCBD5E1)
                    : AppColors.destructive,
                borderRadius: BorderRadius.circular(14),
                boxShadow: (hasActiveStay || !_agreedToTerms || _isSubmitting)
                    ? null
                    : const [
                        BoxShadow(
                          color: Color(0x28DC2626),
                          blurRadius: 14,
                          offset: Offset(0, 4),
                        ),
                      ],
              ),
              child: Center(
                child: _isSubmitting
                    ? const SizedBox(
                        width: 22,
                        height: 22,
                        child: CircularProgressIndicator(
                          strokeWidth: 2.2,
                          color: Colors.white,
                        ),
                      )
                    : const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            LucideIcons.userX,
                            color: Colors.white,
                            size: 18,
                          ),
                          SizedBox(width: 10),
                          Text(
                            'Submit Deletion Request',
                            style: TextStyle(
                              fontSize: 15,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                              letterSpacing: 0.1,
                            ),
                          ),
                        ],
                      ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDisclosureItem(String text, {bool isCritical = false}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            isCritical ? LucideIcons.alertCircle : LucideIcons.checkCircle2,
            size: 16,
            color: isCritical
                ? AppColors.destructive
                : const Color(0xFF64748B),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              style: TextStyle(
                fontSize: 12.5,
                height: 1.45,
                fontWeight: isCritical ? FontWeight.bold : FontWeight.w500,
                color: isCritical
                    ? const Color(0xFF991B1B)
                    : const Color(0xFF475569),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMetaRow(String label, String value) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 110,
          child: Text(
            label,
            style: const TextStyle(
              fontSize: 12.5,
              fontWeight: FontWeight.w600,
              color: Color(0xFF92400E),
            ),
          ),
        ),
        Expanded(
          child: Text(
            value,
            style: const TextStyle(
              fontSize: 12.5,
              fontWeight: FontWeight.w500,
              color: Color(0xFF78350F),
            ),
          ),
        ),
      ],
    );
  }
}
