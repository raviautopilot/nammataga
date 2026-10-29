import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/payment/razorpay_checkout_manager.dart';
import '../../../../core/widgets/bouncy_button.dart';
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

  // Custom amount controllers per plan ID
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
      _customAmountControllers[planId] = TextEditingController(text: '1000');
    }
    return _customAmountControllers[planId]!;
  }

  Future<void> _handlePayment({
    required SubscriptionPlan plan,
    required int amountInRupees,
    required String memberEmail,
    required String memberName,
    required String memberTagaId,
  }) async {
    if (amountInRupees <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please enter an amount greater than zero.'),
          backgroundColor: AppColors.destructive,
        ),
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
        'name': 'Namma TAGA',
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
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text('Payment Error: $errorMessage'),
                backgroundColor: AppColors.destructive,
              ),
            );
          }
        },
        onDismiss: () {
          if (mounted) {
            setState(() {
              _isProcessingPayment = false;
              _processingPlanId = null;
            });
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Payment cancelled.'),
                backgroundColor: AppColors.textSecondary,
                duration: Duration(seconds: 2),
              ),
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
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to initiate payment: $e'),
            backgroundColor: AppColors.destructive,
          ),
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
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Payment verification issue: $e'),
            backgroundColor: AppColors.destructive,
          ),
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
            mainAxisSize: MainAxisSize.min,
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
          // 1. Current Membership Status Card
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
                children: filtered.map((plan) {
                  final isPaid = paidSubs.contains(plan.id);
                  return _buildPlanCard(
                    plan: plan,
                    isPaid: isPaid,
                    profile: profile,
                  );
                }).toList(),
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

  Widget _buildStatusBanner(dynamic profile) {
    final isSubscribed =
        profile != null && (profile.isPaid || profile.subscriptionActive);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: isSubscribed
              ? const Color(0xFF52B788).withValues(alpha: 0.4)
              : const Color(0xFFFDE68A),
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
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: isSubscribed
                  ? const Color(0xFFE8F5EE)
                  : const Color(0xFFFEF3C7),
              shape: BoxShape.circle,
            ),
            child: Icon(
              isSubscribed ? LucideIcons.shieldCheck : LucideIcons.crown,
              color: isSubscribed
                  ? AppColors.primary
                  : const Color(0xFFD97706),
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
                        isSubscribed
                            ? 'Active Subscription'
                            : 'General Member',
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(width: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 7,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: isSubscribed
                            ? const Color(0xFFE8F5EE)
                            : const Color(0xFFFEF3C7),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        isSubscribed ? 'Active' : 'Unpaid',
                        style: TextStyle(
                          fontSize: 10.5,
                          fontWeight: FontWeight.bold,
                          color: isSubscribed
                              ? AppColors.primary
                              : const Color(0xFFB45309),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 3),
                Text(
                  isSubscribed
                      ? 'Annual dues paid. Valid through March 31, 2027.'
                      : 'Pay your annual subscription to activate all subscriber benefits.',
                  style: const TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
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

  Widget _buildPlanCard({
    required SubscriptionPlan plan,
    required bool isPaid,
    required dynamic profile,
  }) {
    final isProcessing =
        _isProcessingPayment && _processingPlanId == plan.id;
    final isCustom = plan.allowCustomAmount || plan.amount == null;

    final memberEmail = profile?.emailId ?? '';
    final memberName = profile?.name ?? '';
    final memberTagaId = profile?.tagaId ?? '';

    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 14),
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: isPaid
              ? const Color(0xFF52B788)
              : const Color(0xFFE2EBE5),
          width: isPaid ? 1.4 : 1.1,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 12,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header: Name & Frequency / Status
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
              if (isPaid)
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFFE8F5EE),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: const Color(0xFF52B788)),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        LucideIcons.check,
                        size: 11,
                        color: AppColors.primary,
                      ),
                      SizedBox(width: 4),
                      Text(
                        'Paid',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: AppColors.primary,
                        ),
                      ),
                    ],
                  ),
                )
              else
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF3F4F6),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    plan.frequency,
                    style: const TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF4B5563),
                    ),
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

          // Amount Row / Custom Amount Input
          if (!isCustom) ...[
            Row(
              crossAxisAlignment: CrossAxisAlignment.baseline,
              textBaseline: TextBaseline.alphabetic,
              children: [
                Text(
                  plan.displayAmount,
                  style: const TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.bold,
                    color: AppColors.primaryDark,
                  ),
                ),
                const SizedBox(width: 6),
                Text(
                  plan.frequency == 'Annual' ? '/ year' : 'one-time fee',
                  style: const TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ] else ...[
            // Quick preset chips
            const Text(
              'Select or Enter Contribution Amount:',
              style: TextStyle(
                fontSize: 11.5,
                fontWeight: FontWeight.w600,
                color: Color(0xFF4B5563),
              ),
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [500, 1000, 2500, 5000].map((preset) {
                return InkWell(
                  borderRadius: BorderRadius.circular(8),
                  onTap: () {
                    _getController(plan.id).text = preset.toString();
                    setState(() {});
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: _getController(plan.id).text == preset.toString()
                          ? const Color(0xFFE8F5EE)
                          : const Color(0xFFF9FAF7),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: _getController(plan.id).text == preset.toString()
                            ? AppColors.primary
                            : const Color(0xFFD8E3DC),
                      ),
                    ),
                    child: Text(
                      '₹$preset',
                      style: TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.bold,
                        color: _getController(plan.id).text == preset.toString()
                            ? AppColors.primary
                            : AppColors.textPrimary,
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 10),
            TextField(
              controller: _getController(plan.id),
              keyboardType: TextInputType.number,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              decoration: const InputDecoration(
                prefixText: '₹ ',
                prefixStyle: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: AppColors.primaryDark,
                ),
                hintText: 'Enter custom amount',
                contentPadding: EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 10,
                ),
              ),
            ),
          ],
          const SizedBox(height: 16),

          // Pay Button
          if (isPaid && !isCustom) ...[
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
                    'Subscribed & Current for 2026-27',
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
                      : AppColors.primaryDark,
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [
                    BoxShadow(
                      color: AppColors.primaryDark.withValues(alpha: 0.2),
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
                              'Opening Razorpay...',
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
                            const Icon(
                              LucideIcons.lock,
                              size: 14,
                              color: Color(0xFFFDE68A),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              isCustom
                                  ? 'Contribute via Razorpay'
                                  : 'Pay ${plan.displayAmount} with Razorpay',
                              style: const TextStyle(
                                fontSize: 13.5,
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
        ],
      ),
    );
  }
}
