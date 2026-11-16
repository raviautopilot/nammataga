import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/payment/razorpay_checkout_manager.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../../../core/widgets/taga_toast.dart';
import '../domain/subscription_plan.dart';
import '../providers/profile_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

class SubscriptionPlansScreen extends ConsumerStatefulWidget {
  const SubscriptionPlansScreen({super.key});

  @override
  ConsumerState<SubscriptionPlansScreen> createState() =>
      _SubscriptionPlansScreenState();
}

class _SubscriptionPlansScreenState
    extends ConsumerState<SubscriptionPlansScreen> {
  final RazorpayCheckoutManager _checkoutManager = RazorpayCheckoutManager();
  String _selectedCategory = 'all'; // 'all', 'fees', 'contributions'
  bool _isProcessingPayment = false;
  String? _processingPlanId;

  // Custom amount controllers per plan ID (starts empty)
  final Map<String, TextEditingController> _customAmountControllers = {};

  @override
  void dispose() {
    _checkoutManager.dispose();
    for (final controller in _customAmountControllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  TextEditingController _getController(String planId) {
    if (!_customAmountControllers.containsKey(planId)) {
      _customAmountControllers[planId] = TextEditingController();
    }
    return _customAmountControllers[planId]!;
  }

  /// Website-aligned Grace Period calculation:
  /// April 1 to May 31 of each year is the 2-month renewal grace period.
  bool _isInGracePeriod() {
    final now = DateTime.now();
    final graceStart = DateTime(now.year, 4, 1);
    final graceEnd = DateTime(now.year, 5, 31, 23, 59, 59);
    return !now.isBefore(graceStart) && !now.isAfter(graceEnd);
  }

  String _formatISODate(String? dateStr) {
    if (dateStr == null || dateStr.isEmpty) return '-';
    try {
      final date = DateTime.parse(dateStr);
      return DateFormat('dd MMM yyyy').format(date);
    } catch (_) {
      return dateStr;
    }
  }

  Future<void> _handlePayment({
    required SubscriptionPlan plan,
    required int amountInRupees,
    required String memberEmail,
    required String memberName,
    required String memberTagaId,
  }) async {
    if (amountInRupees <= 0) {
      TagaToast.error(
        context,
        'Please enter an amount greater than zero.',
        title: 'Invalid Amount',
      );
      return;
    }

    setState(() {
      _isProcessingPayment = true;
      _processingPlanId = plan.id;
    });

    HapticFeedback.mediumImpact();

    try {
      final repo = ref.read(profileRepositoryProvider);
      final amountInPaise = amountInRupees * 100;

      final notes = {
        'subscription_id': plan.id,
        'subscription_name': plan.name,
        'member_name': memberName,
        'member_taga_id': memberTagaId,
        'member_email': memberEmail,
      };

      // 1. Create Order on Backend
      final order = await repo.createSubscriptionOrder(
        subscriptionId: plan.id,
        amountInPaise: amountInPaise,
        email: memberEmail,
        notes: notes,
      );

      final options = {
        'key': order.key,
        'amount': order.amount,
        'currency': order.currency,
        'name': 'TAGA',
        'description': plan.name,
        'order_id': order.orderId,
        'prefill': {
          'name': memberName,
          'email': memberEmail,
        },
        'theme': {
          'color': '#1B4332',
        },
      };

      // 2. Open Razorpay Checkout (Web modal or Native SDK)
      _checkoutManager.openCheckout(
        options: options,
        onSuccess: (paymentId, orderId, signature) async {
          await _onPaymentSuccess(
            plan: plan,
            orderId: orderId.isNotEmpty ? orderId : order.orderId,
            paymentId: paymentId,
            signature: signature,
            memberEmail: memberEmail,
            amountInPaise: amountInPaise,
          );
        },
        onError: (errorMessage) {
          if (mounted) {
            setState(() {
              _isProcessingPayment = false;
              _processingPlanId = null;
            });
            TagaToast.error(
              context,
              errorMessage,
              title: 'Payment Error',
            );
          }
        },
        onDismiss: () {
          if (mounted) {
            setState(() {
              _isProcessingPayment = false;
              _processingPlanId = null;
            });
            TagaToast.info(
              context,
              'Payment cancelled.',
              title: 'Cancelled',
            );
          }
        },
      );
    } catch (e) {
      if (mounted) {
        setState(() {
          _isProcessingPayment = false;
          _processingPlanId = null;
        });
        TagaToast.error(
          context,
          'Failed to initiate payment: $e',
          title: 'Payment Error',
        );
      }
    }
  }

  Future<void> _onPaymentSuccess({
    required SubscriptionPlan plan,
    required String orderId,
    required String paymentId,
    required String signature,
    required String memberEmail,
    required int amountInPaise,
  }) async {
    try {
      final repo = ref.read(profileRepositoryProvider);

      // Verify payment with backend
      await repo.verifySubscriptionPayment(
        subscriptionId: plan.id,
        orderId: orderId,
        paymentId: paymentId,
        signature: signature,
        email: memberEmail,
        amountInPaise: amountInPaise,
      );

      // Invalidate profile and subscription state
      ref.invalidate(memberProfileProvider);
      ref.invalidate(memberSubscriptionStatusProvider);
      ref.invalidate(memberPaidSubscriptionsProvider);

      if (mounted) {
        setState(() {
          _isProcessingPayment = false;
          _processingPlanId = null;
        });

        _showSuccessDialog(plan, paymentId);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isProcessingPayment = false;
          _processingPlanId = null;
        });
        TagaToast.error(
          context,
          'Payment verification issue: $e',
          title: 'Verification Issue',
        );
      }
    }
  }

  void _showSuccessDialog(SubscriptionPlan plan, String paymentId) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        content: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: Column(
            mainAxisSize: minAxisSize,
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: const BoxDecoration(
                  color: Color(0xFFE8F5EE),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  LucideIcons.checkCheck,
                  size: 32,
                  color: AppColors.primary,
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Payment Successful!',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Your payment for "${plan.name}" has been confirmed and verified.',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                  height: 1.4,
                ),
              ),
              if (paymentId.isNotEmpty) ...[
                const SizedBox(height: 12),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF3F4F6),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    'Ref ID: $paymentId',
                    style: const TextStyle(
                      fontSize: 11,
                      fontFamily: 'monospace',
                      color: AppColors.textPrimary,
                    ),
                  ),
                ),
              ],
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primaryDark,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 12),
                  ),
                  onPressed: () => Navigator.of(ctx).pop(),
                  child: const Text(
                    'Done',
                    style: TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  static const minAxisSize = MainAxisSize.min;

  @override
  Widget build(BuildContext context) {
    final plansAsync = ref.watch(subscriptionPlansProvider);
    final profileAsync = ref.watch(memberProfileProvider);
    final paidSubsAsync = ref.watch(memberPaidSubscriptionsProvider);

    final profile = profileAsync.value;
    final paidSubs = paidSubsAsync.value ?? [];

    return ProfileSubpageScaffold(
      title: 'Subscription Plans & Dues',
      watermarkIcon: LucideIcons.creditCard,
      showEditNotice: false,
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 1. Current Membership Status Card (Exact website matching)
          _buildStatusBanner(profile),
          const SizedBox(height: 18),

          // 2. Category Filter Chips
          _buildCategoryFilters(),
          const SizedBox(height: 16),

          // 3. Plans List
          plansAsync.when(
            data: (plans) {
              final filtered = _filterPlans(plans);

              if (filtered.isEmpty) {
                return Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(32),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: const Color(0xFFE2EBE5)),
                  ),
                  child: const Center(
                    child: Text(
                      'No subscription plans available in this category.',
                      style: TextStyle(
                        fontSize: 13,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ),
                );
              }

              return Column(
                children: [
                  ...filtered.map((plan) {
                    final isAnnual = plan.id == 'annual-subscription';
                    final inGrace = _isInGracePeriod();
                    final isMemberPaid =
                        profile != null && (profile.isPaid || profile.subscriptionActive);
                    final isPaid = paidSubs.contains(plan.id) ||
                        (isAnnual && isMemberPaid && !inGrace);

                    return _buildPlanCard(
                      plan: plan,
                      isPaid: isPaid,
                      inGrace: inGrace,
                      isMemberPaid: isMemberPaid,
                      profile: profile,
                    );
                  }),
                  const SizedBox(height: 12),
                  // Professional Payment Gateway Security & Advisory Card (Website Aligned)
                  _buildPaymentInformationCard(),
                ],
              );
            },
            loading: () => const Center(
              child: Padding(
                padding: EdgeInsets.all(40),
                child: CircularProgressIndicator(color: AppColors.primary),
              ),
            ),
            error: (err, stack) => Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Text(
                  'Failed to load subscription plans.\n$err',
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    color: AppColors.destructive,
                    fontSize: 13,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Website-aligned Membership Status Card with Grace Period / Active / Unpaid state
  Widget _buildStatusBanner(dynamic profile) {
    final isSubscribed =
        profile != null && (profile.isPaid || profile.subscriptionActive);
    final inGrace = _isInGracePeriod();

    String badgeLabel;
    Color badgeBg;
    Color badgeBorder;
    Color badgeTextColor;
    String statusTitle;
    String statusDesc;
    IconData statusIcon;

    if (isSubscribed && inGrace) {
      // Annual member in grace period (April 1 - May 31)
      badgeLabel = 'Renewal Grace Period';
      badgeBg = const Color(0xFFFFF7ED); // Orange-50
      badgeBorder = const Color(0xFFFDBA74); // Orange-300
      badgeTextColor = const Color(0xFFC2410C); // Orange-700
      statusTitle = 'Renewal Grace Period Active';
      statusDesc = 'Grace period active until May 31. Please renew your annual dues to continue member privileges.';
      statusIcon = LucideIcons.clock;
    } else if (isSubscribed) {
      // Active Paid Member
      badgeLabel = 'Active Member';
      badgeBg = const Color(0xFFE8F5EE);
      badgeBorder = const Color(0xFF86EFAC);
      badgeTextColor = AppColors.primary;
      statusTitle = 'Active Subscription';
      statusDesc = 'Annual dues paid. Valid through March 31, 2027.';
      statusIcon = LucideIcons.shieldCheck;
    } else {
      // Non-active / Unpaid General Member
      badgeLabel = 'Unpaid';
      badgeBg = const Color(0xFFFEF3C7);
      badgeBorder = const Color(0xFFFCD34D);
      badgeTextColor = const Color(0xFFB45309);
      statusTitle = 'General Member';
      statusDesc = 'Pay your annual subscription to activate all subscriber benefits & room booking privileges.';
      statusIcon = LucideIcons.crown;
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: badgeBorder,
          width: 1.2,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 12,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: badgeBg,
              shape: BoxShape.circle,
            ),
            child: Icon(
              statusIcon,
              color: badgeTextColor,
              size: 22,
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Flexible(
                      child: Text(
                        statusTitle,
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 2.5,
                      ),
                      decoration: BoxDecoration(
                        color: badgeBg,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: badgeBorder, width: 1),
                      ),
                      child: Text(
                        badgeLabel,
                        style: TextStyle(
                          fontSize: 10.5,
                          fontWeight: FontWeight.bold,
                          color: badgeTextColor,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 5),
                Text(
                  statusDesc,
                  style: const TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                    height: 1.35,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCategoryFilters() {
    final categories = [
      {'id': 'all', 'label': 'All Plans'},
      {'id': 'fees', 'label': 'Membership Fees'},
      {'id': 'contributions', 'label': 'Voluntary Funds'},
    ];

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      physics: const BouncingScrollPhysics(),
      child: Row(
        children: categories.map((cat) {
          final isSelected = _selectedCategory == cat['id'];
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: InkWell(
              borderRadius: BorderRadius.circular(20),
              onTap: () {
                setState(() => _selectedCategory = cat['id']!);
              },
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
                decoration: BoxDecoration(
                  color: isSelected ? AppColors.primaryDark : Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: isSelected
                        ? AppColors.primaryDark
                        : const Color(0xFFD8E3DC),
                  ),
                ),
                child: Text(
                  cat['label']!,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                    color: isSelected ? Colors.white : AppColors.textPrimary,
                  ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  List<SubscriptionPlan> _filterPlans(List<SubscriptionPlan> plans) {
    if (_selectedCategory == 'fees') {
      return plans.where((p) => !p.needBased && !p.allowCustomAmount).toList();
    }
    if (_selectedCategory == 'contributions') {
      return plans.where((p) => p.needBased || p.allowCustomAmount).toList();
    }
    return plans;
  }

  /// Website-aligned Plan Card
  Widget _buildPlanCard({
    required SubscriptionPlan plan,
    required bool isPaid,
    required bool inGrace,
    required bool isMemberPaid,
    required dynamic profile,
  }) {
    final isProcessing =
        _isProcessingPayment && _processingPlanId == plan.id;
    final isAnnual = plan.id == 'annual-subscription';
    final isGraceRenewal = isAnnual && inGrace && isMemberPaid && !isPaid;
    final isCustom = plan.allowCustomAmount || plan.amount == null;

    final memberEmail = profile?.emailId ?? '';
    final memberName = profile?.name ?? '';
    final memberTagaId = profile?.tagaId ?? '';

    // Only oneTime items that are paid, OR annual subscription when paid (not in grace)
    // Need-based voluntary funds are never permanently marked as 'Paid'
    final isEffectivelyPaid = (!plan.needBased && isPaid && !isGraceRenewal);

    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: isAnnual
              ? const Color(0xFF86EFAC)
              : isEffectivelyPaid
                  ? const Color(0xFF52B788)
                  : const Color(0xFFE2EBE5),
          width: isAnnual ? 1.5 : (isEffectivelyPaid ? 1.3 : 1.1),
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 12,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Top highlight banner for annual subscription (matching website)
            if (isAnnual)
              Container(
                width: double.infinity,
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
                color: isGraceRenewal ? const Color(0xFFC2410C) : const Color(0xFF1B4332),
                child: Row(
                  children: [
                    Icon(
                      isGraceRenewal ? LucideIcons.alertTriangle : LucideIcons.checkCheck,
                      color: Colors.white,
                      size: 13,
                    ),
                    const SizedBox(width: 6),
                    Text(
                      isGraceRenewal
                          ? 'Renewal Period • Grace until May 31'
                          : 'Required Annual Membership Dues',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 0.2,
                      ),
                    ),
                  ],
                ),
              ),

            Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Header: Name & Status / Frequency
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          plan.name,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                            letterSpacing: -0.2,
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      if (isEffectivelyPaid)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFFE8F5EE),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(color: const Color(0xFF52B788)),
                          ),
                          child: const Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                LucideIcons.check,
                                size: 12,
                                color: AppColors.primary,
                              ),
                              SizedBox(width: 4),
                              Text(
                                'Paid',
                                style: TextStyle(
                                  fontSize: 11.5,
                                  fontWeight: FontWeight.bold,
                                  color: AppColors.primary,
                                ),
                              ),
                            ],
                          ),
                        )
                      else if (isGraceRenewal)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: const Color(0xFFFFF7ED),
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(color: const Color(0xFFFDBA74)),
                          ),
                          child: const Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                LucideIcons.clock,
                                size: 12,
                                color: Color(0xFFC2410C),
                              ),
                              SizedBox(width: 4),
                              Text(
                                'Renew Due',
                                style: TextStyle(
                                  fontSize: 11.5,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFFC2410C),
                                ),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 6),

                  // Description
                  Text(
                    plan.description,
                    style: const TextStyle(
                      fontSize: 12.5,
                      color: AppColors.textSecondary,
                      height: 1.35,
                    ),
                  ),
                  const SizedBox(height: 14),

                  // Website-aligned 3-column metadata row (Amount, Frequency, Next Due)
                  Container(
                    padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 14),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF8FAFC),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: const Color(0xFFE2E8F0)),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        // Amount column
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'Amount',
                                style: TextStyle(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w500,
                                  color: Color(0xFF64748B),
                                ),
                              ),
                              const SizedBox(height: 2),
                              Text(
                                isCustom ? 'Flexible' : plan.displayAmount,
                                style: const TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF0F172A),
                                ),
                              ),
                            ],
                          ),
                        ),

                        // Frequency column (Website-style exact frequency: Annual / One-time / Flexible)
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'Frequency',
                                style: TextStyle(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w500,
                                  color: Color(0xFF64748B),
                                ),
                              ),
                              const SizedBox(height: 2),
                              Text(
                                plan.frequency,
                                style: const TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF0F172A),
                                ),
                              ),
                            ],
                          ),
                        ),

                        // Next Due Date column (Shown for Annual Subscription)
                        if (isAnnual && plan.nextDueDate != null)
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                const Text(
                                  'Next Due',
                                  style: TextStyle(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w500,
                                    color: Color(0xFF64748B),
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  _formatISODate(plan.nextDueDate),
                                  style: const TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.bold,
                                    color: Color(0xFF0F172A),
                                  ),
                                ),
                              ],
                            ),
                          ),
                      ],
                    ),
                  ),

                  // If custom amount allowed: clean dynamic input field (preset 500, 1000 chips removed)
                  if (isCustom) ...[
                    const SizedBox(height: 14),
                    const Text(
                      'Contribution Amount (₹)',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF334155),
                      ),
                    ),
                    const SizedBox(height: 6),
                    TextField(
                      controller: _getController(plan.id),
                      keyboardType: TextInputType.number,
                      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: AppColors.primaryDark,
                      ),
                      decoration: InputDecoration(
                        prefixIcon: const Padding(
                          padding: EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          child: Text(
                            '₹',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              color: AppColors.primaryDark,
                            ),
                          ),
                        ),
                        prefixIconConstraints: const BoxConstraints(minWidth: 0, minHeight: 0),
                        hintText: 'Enter amount (e.g. 1000)',
                        hintStyle: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.normal,
                          color: Colors.grey.shade400,
                        ),
                        filled: true,
                        fillColor: const Color(0xFFF8FAFC),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 12,
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
                            color: AppColors.primaryDark,
                            width: 1.5,
                          ),
                        ),
                      ),
                    ),
                  ],

                  const SizedBox(height: 16),

                  // Action Button
                  if (isEffectivelyPaid) ...[
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF0FDF4),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFBBF7D0)),
                      ),
                      child: const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            LucideIcons.circleCheck,
                            size: 16,
                            color: AppColors.primary,
                          ),
                          SizedBox(width: 8),
                          Text(
                            'Paid & Active',
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: AppColors.primaryDark,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ] else ...[
                    BouncyButton(
                      onTap: isProcessing
                          ? null
                          : () {
                              final amount = isCustom
                                  ? int.tryParse(_getController(plan.id).text.trim()) ?? 0
                                  : (plan.amount ?? 0);

                              _handlePayment(
                                plan: plan,
                                amountInRupees: amount,
                                memberEmail: memberEmail,
                                memberName: memberName,
                                memberTagaId: memberTagaId,
                              );
                            },
                      child: Container(
                        width: double.infinity,
                        height: 48,
                        decoration: BoxDecoration(
                          color: isProcessing
                              ? const Color(0xFF9CA3AF)
                              : isGraceRenewal
                                  ? const Color(0xFFC2410C)
                                  : AppColors.primaryDark,
                          borderRadius: BorderRadius.circular(12),
                          boxShadow: [
                            BoxShadow(
                              color: (isGraceRenewal
                                      ? const Color(0xFFC2410C)
                                      : AppColors.primaryDark)
                                  .withValues(alpha: 0.2),
                              blurRadius: 8,
                              offset: const Offset(0, 3),
                            ),
                          ],
                        ),
                        child: Center(
                          child: isProcessing
                              ? const Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    SizedBox(
                                      width: 18,
                                      height: 18,
                                      child: CircularProgressIndicator(
                                        color: Colors.white,
                                        strokeWidth: 2,
                                      ),
                                    ),
                                    SizedBox(width: 10),
                                    Text(
                                      'Connecting Secure Payment...',
                                      style: TextStyle(
                                        fontSize: 13.5,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.white,
                                      ),
                                    ),
                                  ],
                                )
                              : Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(
                                      isGraceRenewal
                                          ? LucideIcons.refreshCw
                                          : LucideIcons.shieldCheck,
                                      size: 15,
                                      color: const Color(0xFFFDE68A),
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      isGraceRenewal
                                          ? 'Renew Subscription'
                                          : isCustom
                                              ? 'Proceed to Contribute'
                                              : 'Pay ${plan.displayAmount}',
                                      style: const TextStyle(
                                        fontSize: 14,
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
                    if (isGraceRenewal)
                      const Padding(
                        padding: EdgeInsets.only(top: 6),
                        child: Center(
                          child: Text(
                            'Grace period active until May 31',
                            style: TextStyle(
                              fontSize: 11.5,
                              color: Color(0xFFC2410C),
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Website-aligned Payment Information Card (Matches taga-web/src/components/Membership.tsx)
  Widget _buildPaymentInformationCard() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFEFF6FF), // Blue-50
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFBFDBFE)), // Blue-200
      ),
      child: const Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            LucideIcons.info,
            size: 20,
            color: Color(0xFF2563EB), // Blue-600
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Payment & Membership Guidelines',
                  style: TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF1E3A8A), // Blue-900
                  ),
                ),
                SizedBox(height: 6),
                Text(
                  '• All transactions are encrypted and processed securely via Razorpay.',
                  style: TextStyle(fontSize: 12, color: Color(0xFF1E40AF), height: 1.4),
                ),
                Text(
                  '• Annual subscription period: April 1 - March 31.',
                  style: TextStyle(fontSize: 12, color: Color(0xFF1E40AF), height: 1.4),
                ),
                Text(
                  '• Grace period: 2 months (until May 31) for annual renewal.',
                  style: TextStyle(fontSize: 12, color: Color(0xFF1E40AF), height: 1.4),
                ),
                Text(
                  '• Official payment receipt and email confirmation delivered upon successful verification.',
                  style: TextStyle(fontSize: 12, color: Color(0xFF1E40AF), height: 1.4),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
