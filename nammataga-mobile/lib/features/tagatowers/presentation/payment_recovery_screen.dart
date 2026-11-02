import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../domain/booking_model.dart';
import '../domain/payment_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';

/// Status of the payment recovery resolution process
enum PaymentRecoveryStatus {
  verifying,
  pendingConfirmation,
  verified,
  issue,
}

/// A dedicated recovery screen designed to gracefully handle edge cases where:
/// - Payment gateway successfully deducted money from the member's account,
///   but the app crashed, device battery died, or network dropped before
///   the backend verification response was processed.
/// - Razorpay returned a success response, but backend signature verification was delayed.
/// - The member needs to re-verify an existing transaction without initiating a duplicate charge.
///
/// DUPLICATE SAFETY GUARANTEE:
/// This screen strictly NEVER initiates a new payment order or creates duplicate bookings.
class PaymentRecoveryScreen extends ConsumerStatefulWidget {
  /// Preserved booking ID held in draft state
  final String? bookingId;

  /// Preserved Razorpay Order ID
  final String? razorpayOrderId;

  /// Preserved Razorpay Payment ID
  final String? razorpayPaymentId;

  /// Preserved Razorpay Signature
  final String? razorpaySignature;

  /// Advance payment amount in INR
  final int? amount;

  /// Room name or accommodation label (optional display)
  final String? roomName;

  /// Initial status to display upon entering the screen
  final PaymentRecoveryStatus? initialStatus;

  const PaymentRecoveryScreen({
    super.key,
    this.bookingId,
    this.razorpayOrderId,
    this.razorpayPaymentId,
    this.razorpaySignature,
    this.amount,
    this.roomName,
    this.initialStatus,
  });

  @override
  ConsumerState<PaymentRecoveryScreen> createState() =>
      _PaymentRecoveryScreenState();
}

class _PaymentRecoveryScreenState extends ConsumerState<PaymentRecoveryScreen>
    with SingleTickerProviderStateMixin {
  late PaymentRecoveryStatus _status;
  bool _isChecking = false;
  String? _statusMessage;
  String? _errorMessage;
  BookingResponse? _confirmedBooking;

  // Local preserved credentials (from widget props or local storage/flow state)
  String? _bookingId;
  String? _razorpayOrderId;
  String? _razorpayPaymentId;
  String? _razorpaySignature;
  int _amount = 0;
  String? _roomName;

  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  // Storage keys for persisting interrupted payment state across app kills
  static const String _keyInterruptedBookingId = 'tagatowers_recovery_booking_id';
  static const String _keyInterruptedOrderId = 'tagatowers_recovery_order_id';
  static const String _keyInterruptedPaymentId = 'tagatowers_recovery_payment_id';
  static const String _keyInterruptedSignature = 'tagatowers_recovery_signature';
  static const String _keyInterruptedAmount = 'tagatowers_recovery_amount';
  static const String _keyInterruptedRoomName = 'tagatowers_recovery_room_name';

  @override
  void initState() {
    super.initState();
    _status = widget.initialStatus ?? PaymentRecoveryStatus.verifying;

    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1400),
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.96, end: 1.04).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );

    WidgetsBinding.instance.addPostFrameCallback((_) {
      _initAndResolveCredentials();
    });
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  /// Initializes transaction credentials from props, flow state, or disk
  Future<void> _initAndResolveCredentials() async {
    final flow = ref.read(bookingFlowNotifierProvider);

    _bookingId = widget.bookingId ?? flow.pendingBookingId;
    _razorpayOrderId = widget.razorpayOrderId ?? flow.lastPaymentOrderId;
    _razorpayPaymentId = widget.razorpayPaymentId ?? flow.lastPaymentId;
    _razorpaySignature = widget.razorpaySignature ?? flow.lastPaymentSignature;
    _amount = widget.amount ?? flow.advanceAmount;
    _roomName = widget.roomName ?? flow.selectedRoom?.name;

    // Fallback to disk if app was terminated and state was cleared
    if (_bookingId == null || _bookingId!.isEmpty) {
      final prefs = await SharedPreferences.getInstance();
      _bookingId = prefs.getString(_keyInterruptedBookingId) ?? _bookingId;
      _razorpayOrderId = prefs.getString(_keyInterruptedOrderId) ?? _razorpayOrderId;
      _razorpayPaymentId =
          prefs.getString(_keyInterruptedPaymentId) ?? _razorpayPaymentId;
      _razorpaySignature =
          prefs.getString(_keyInterruptedSignature) ?? _razorpaySignature;
      final savedAmount = prefs.getInt(_keyInterruptedAmount);
      if (savedAmount != null && savedAmount > 0) _amount = savedAmount;
      _roomName = prefs.getString(_keyInterruptedRoomName) ?? _roomName;
    } else {
      // Persist credentials in case of subsequent network loss or tab switch
      _persistInterruptedPayment();
    }

    if (mounted) {
      setState(() {});
      // Auto-trigger verification once credentials are initialized only if not provided an explicit initial state
      if (widget.initialStatus == null) {
        _checkPaymentStatus(isInitialCheck: true);
      }
    }
  }

  /// Persists current recovery payload to SharedPreferences
  Future<void> _persistInterruptedPayment() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      if (_bookingId != null) {
        await prefs.setString(_keyInterruptedBookingId, _bookingId!);
      }
      if (_razorpayOrderId != null) {
        await prefs.setString(_keyInterruptedOrderId, _razorpayOrderId!);
      }
      if (_razorpayPaymentId != null) {
        await prefs.setString(_keyInterruptedPaymentId, _razorpayPaymentId!);
      }
      if (_razorpaySignature != null) {
        await prefs.setString(_keyInterruptedSignature, _razorpaySignature!);
      }
      if (_amount > 0) {
        await prefs.setInt(_keyInterruptedAmount, _amount);
      }
      if (_roomName != null) {
        await prefs.setString(_keyInterruptedRoomName, _roomName!);
      }
    } catch (_) {}
  }

  /// Clears persisted recovery payload upon successful confirmation
  Future<void> _clearPersistedRecovery() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(_keyInterruptedBookingId);
      await prefs.remove(_keyInterruptedOrderId);
      await prefs.remove(_keyInterruptedPaymentId);
      await prefs.remove(_keyInterruptedSignature);
      await prefs.remove(_keyInterruptedAmount);
      await prefs.remove(_keyInterruptedRoomName);
    } catch (_) {}
  }

  /// Re-verifies payment status via two coordinated avenues:
  /// 1. Calls backend `verifyPayment` with preserved credentials.
  /// 2. Queries `getUserBookings(bookerId)` to check if payment is already marked confirmed.
  Future<void> _checkPaymentStatus({bool isInitialCheck = false}) async {
    if (_isChecking) return;

    setState(() {
      _isChecking = true;
      _status = PaymentRecoveryStatus.verifying;
      _errorMessage = null;
      _statusMessage = 'Connecting to TAGA Towers server & re-verifying...';
    });

    final repo = ref.read(tagaTowersRepositoryProvider);
    final session = ref.read(authProvider).value;
    final bookerId = session?.tagaId ?? '';

    bool verified = false;

    // ── STEP 1: Attempt direct backend verification if signature is present ──
    if (_bookingId != null &&
        _bookingId!.isNotEmpty &&
        _razorpayOrderId != null &&
        _razorpayOrderId!.isNotEmpty &&
        _razorpayPaymentId != null &&
        _razorpayPaymentId!.isNotEmpty) {
      try {
        final req = VerifyTowersPaymentRequest(
          bookingId: _bookingId!,
          razorpayOrderId: _razorpayOrderId!,
          razorpayPaymentId: _razorpayPaymentId!,
          razorpaySignature: _razorpaySignature ?? '',
        );

        await repo.verifyPayment(req);
        verified = true;
      } catch (e) {
        // Verification endpoint might reject if already confirmed or signature missing.
        // We gracefully fall through to checking the member's live bookings list.
        debugPrint('[PaymentRecovery] verifyPayment error: $e');
      }
    }

    // ── STEP 2: Check user bookings list to inspect live paymentStatus ──
    if (bookerId.isNotEmpty) {
      try {
        final bookings = await repo.getUserBookings(bookerId);

        // Search for this booking by ID, OrderID, or PaymentID
        BookingResponse? matched;
        for (final b in bookings) {
          if ((_bookingId != null && b.id == _bookingId) ||
              (_razorpayPaymentId != null &&
                  b.razorpayPaymentId == _razorpayPaymentId) ||
              (_razorpayOrderId != null &&
                  b.razorpayOrderId == _razorpayOrderId)) {
            matched = b;
            break;
          }
        }

        if (matched != null) {
          final pStatus = matched.paymentStatus.toLowerCase().trim();
          if (pStatus == 'confirmed') {
            verified = true;
            _confirmedBooking = matched;
          } else if (pStatus == 'pending') {
            // Still registered as pending on backend
            if (!verified) {
              if (mounted) {
                setState(() {
                  _status = PaymentRecoveryStatus.pendingConfirmation;
                  _statusMessage =
                      'Your booking reference was found, but backend gateway confirmation is still pending.';
                });
              }
            }
          }
        }
      } catch (e) {
        debugPrint('[PaymentRecovery] getUserBookings error: $e');
      }
    }

    if (!mounted) return;

    if (verified) {
      HapticFeedback.heavyImpact();
      await _clearPersistedRecovery();

      // Refresh Riverpod bookings provider caches
      if (bookerId.isNotEmpty) {
        ref.invalidate(userBookingsProvider(bookerId));
        ref.invalidate(pastBookingsProvider(null));
      }

      setState(() {
        _isChecking = false;
        _status = PaymentRecoveryStatus.verified;
        _statusMessage = 'Payment confirmed and room reservation is secured!';
      });
    } else {
      HapticFeedback.lightImpact();
      setState(() {
        _isChecking = false;
        if (_status != PaymentRecoveryStatus.pendingConfirmation) {
          _status = PaymentRecoveryStatus.issue;
        }
        _errorMessage =
            'We could not verify your payment at this moment. If funds were debited, do not worry — please contact our caretaker team with your Payment ID.';
      });
    }
  }

  /// Copies a transaction ID to the device clipboard with tactile feedback
  void _copyToClipboard(String text, String label) {
    Clipboard.setData(ClipboardData(text: text));
    HapticFeedback.selectionClick();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(LucideIcons.circleCheck, color: Colors.white, size: 18),
            const SizedBox(width: 8),
            Text(
              '$label copied to clipboard',
              style: const TextStyle(fontWeight: FontWeight.w500),
            ),
          ],
        ),
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 2),
        backgroundColor: TagaTowersDesignTokens.deepEmerald,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
    );
  }

  /// Opens telephone dialer with phone number
  Future<void> _makePhoneCall(String phoneNumber) async {
    final clean = phoneNumber.replaceAll(RegExp(r'\s+'), '');
    final uri = Uri.parse('tel:$clean');
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not open phone dialer for $phoneNumber')),
        );
      }
    }
  }

  /// Opens mail client with pre-filled payment dispute details
  Future<void> _sendEmailSupport() async {
    final subject = 'TAGA Towers Room Booking Payment Issue - ${_razorpayPaymentId ?? _bookingId ?? 'Member Assistance'}';
    final body = '''Dear TAGA Towers Caretaker / Admin,

I experienced a network or verification issue during room booking advance payment.

Details:
- Booking Reference ID: ${_bookingId ?? 'N/A'}
- Razorpay Order ID: ${_razorpayOrderId ?? 'N/A'}
- Razorpay Payment ID: ${_razorpayPaymentId ?? 'N/A'}
- Amount Paid: ₹${_amount > 0 ? _amount : 200}
- Room: ${_roomName ?? 'TAGA Towers Room'}

Kindly confirm my booking or verify the transaction.

Thank you.''';

    final uri = Uri(
      scheme: 'mailto',
      path: TagaTowersDesignTokens.helpdeskEmail,
      query: 'subject=${Uri.encodeComponent(subject)}&body=${Uri.encodeComponent(body)}',
    );

    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Could not open email client')),
        );
      }
    }
  }

  /// Navigates to booking confirmation screen
  void _navigateToConfirmation() {
    // Clear interrupted flow state
    ref.read(bookingFlowNotifierProvider.notifier).reset();

    // Prefer navigating to confirmation route with booking or ID
    context.go(
      '/taga-towers/confirmation',
      extra: _confirmedBooking ?? _bookingId,
    );
  }

  /// Navigates safely back to TAGA Towers Home
  void _navigateToTowersHome() {
    context.go('/taga-towers');
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: true,
      onPopInvokedWithResult: (didPop, result) {
        // Prevent accidental hardware back button loops
      },
      child: Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: _buildAppBar(),
        body: SafeArea(
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(
              parent: BouncingScrollPhysics(),
            ),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // 1. Dynamic Status Banner
                _buildStatusBanner(),
                const SizedBox(height: 20),

                // 2. Transaction Reference Card
                _buildTransactionCard(),
                const SizedBox(height: 20),

                // 3. Primary Action Buttons
                _buildActionButtons(),
                const SizedBox(height: 24),

                // 4. Duplicate Safety & Member Advisory Notice
                if (_status != PaymentRecoveryStatus.verified) ...[
                  _buildMemberAdvisoryNotice(),
                  const SizedBox(height: 24),
                ],

                // 5. On-site Caretakers & Helpdesk Contacts Hub
                _buildSupportHub(),
                const SizedBox(height: 20),

                // 6. Duplicate Safety Guarantee Footnote
                _buildDuplicateSafetyBadge(),
                const SizedBox(height: 32),
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// Clean, accessible AppBar with status indicator chip
  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: Colors.white,
      surfaceTintColor: Colors.transparent,
      elevation: 0,
      centerTitle: false,
      leading: IconButton(
        icon: const Icon(LucideIcons.arrowLeft, color: AppColors.textPrimary),
        tooltip: 'Back to TAGA Towers',
        onPressed: _navigateToTowersHome,
      ),
      title: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Payment Recovery',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
              letterSpacing: -0.2,
            ),
          ),
          Text(
            'TAGA Towers Egmore',
            style: TextStyle(
              fontSize: 11.5,
              fontWeight: FontWeight.w500,
              color: AppColors.textSecondary,
            ),
          ),
        ],
      ),
      actions: [
        Padding(
          padding: const EdgeInsets.only(right: 16),
          child: _buildStatusPill(),
        ),
      ],
      bottom: PreferredSize(
        preferredSize: const Size.fromHeight(1),
        child: Container(
          color: TagaTowersDesignTokens.subtleBorder,
          height: 1,
        ),
      ),
    );
  }

  /// Compact status chip displayed in AppBar
  Widget _buildStatusPill() {
    Color bg;
    Color border;
    Color textColor;
    String label;

    switch (_status) {
      case PaymentRecoveryStatus.verifying:
        bg = TagaTowersDesignTokens.softMint;
        border = TagaTowersDesignTokens.mintHighlight;
        textColor = TagaTowersDesignTokens.deepEmerald;
        label = 'Verifying';
        break;
      case PaymentRecoveryStatus.pendingConfirmation:
        bg = TagaTowersDesignTokens.warningAmberLight;
        border = TagaTowersDesignTokens.warningAmberBorder;
        textColor = TagaTowersDesignTokens.warningAmberDark;
        label = 'Pending';
        break;
      case PaymentRecoveryStatus.verified:
        bg = TagaTowersDesignTokens.softMint;
        border = TagaTowersDesignTokens.primaryGreen;
        textColor = TagaTowersDesignTokens.deepEmerald;
        label = 'Verified';
        break;
      case PaymentRecoveryStatus.issue:
        bg = const Color(0xFFFEF2F2);
        border = const Color(0xFFFCA5A5);
        textColor = const Color(0xFF991B1B);
        label = 'Action Needed';
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: border, width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (_status == PaymentRecoveryStatus.verifying)
            const SizedBox(
              width: 10,
              height: 10,
              child: CircularProgressIndicator(
                strokeWidth: 1.5,
                valueColor: AlwaysStoppedAnimation<Color>(
                  TagaTowersDesignTokens.deepEmerald,
                ),
              ),
            )
          else
            Container(
              width: 7,
              height: 7,
              decoration: BoxDecoration(color: textColor, shape: BoxShape.circle),
            ),
          const SizedBox(width: 5),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              color: textColor,
            ),
          ),
        ],
      ),
    );
  }

  /// Dynamic hero status banner rendering the 4 distinct states
  Widget _buildStatusBanner() {
    IconData iconData;
    Color iconColor;
    Color iconBg;
    Color cardBg;
    Color borderColor;
    String bannerTitle;
    String bannerDesc;

    switch (_status) {
      case PaymentRecoveryStatus.verifying:
        iconData = LucideIcons.shieldCheck;
        iconColor = TagaTowersDesignTokens.primaryGreen;
        iconBg = TagaTowersDesignTokens.softMint;
        cardBg = Colors.white;
        borderColor = TagaTowersDesignTokens.mintHighlight;
        bannerTitle = 'Verifying Payment...';
        bannerDesc =
            'We are securely querying the reservation server to confirm your Razorpay payment and room hold. Please keep this screen open.';
        break;

      case PaymentRecoveryStatus.pendingConfirmation:
        iconData = LucideIcons.clock;
        iconColor = TagaTowersDesignTokens.warningAmberDark;
        iconBg = TagaTowersDesignTokens.warningAmberLight;
        cardBg = const Color(0xFFFFFDF5);
        borderColor = TagaTowersDesignTokens.warningAmberBorder;
        bannerTitle = 'Payment Pending Backend Confirmation';
        bannerDesc =
            'Your payment was received by the bank gateway, but backend confirmation is taking a few moments. Do not pay again — click Re-verify below.';
        break;

      case PaymentRecoveryStatus.verified:
        iconData = LucideIcons.circleCheck;
        iconColor = const Color(0xFF15803D);
        iconBg = const Color(0xFFDCFCE7);
        cardBg = const Color(0xFFF0FDF4);
        borderColor = const Color(0xFF86EFAC);
        bannerTitle = 'Payment Verified Successfully';
        bannerDesc =
            'Your booking advance payment has been validated! Your room reservation at TAGA Towers, Egmore is officially confirmed.';
        break;

      case PaymentRecoveryStatus.issue:
        iconData = LucideIcons.alertTriangle;
        iconColor = const Color(0xFFB91C1C);
        iconBg = const Color(0xFFFEE2E2);
        cardBg = const Color(0xFFFFF5F5);
        borderColor = const Color(0xFFFCA5A5);
        bannerTitle = 'Payment Verification Issue';
        bannerDesc =
            'We were unable to automatically verify this transaction. If funds were debited from your account, our caretaker team will verify your payment manually.';
        break;
    }

    return Container(
      decoration: BoxDecoration(
        color: cardBg,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: borderColor, width: 1.5),
        boxShadow: TagaTowersDesignTokens.softShadow,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Animated icon container
              if (_status == PaymentRecoveryStatus.verifying)
                ScaleTransition(
                  scale: _pulseAnimation,
                  child: Container(
                    width: 48,
                    height: 48,
                    decoration: BoxDecoration(
                      color: iconBg,
                      shape: BoxShape.circle,
                      border: Border.all(color: borderColor, width: 1.5),
                    ),
                    child: Center(
                      child: SizedBox(
                        width: 22,
                        height: 22,
                        child: CircularProgressIndicator(
                          strokeWidth: 2.5,
                          valueColor: AlwaysStoppedAnimation<Color>(iconColor),
                        ),
                      ),
                    ),
                  ),
                )
              else
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: iconBg,
                    shape: BoxShape.circle,
                    border: Border.all(color: borderColor, width: 1.5),
                  ),
                  child: Icon(iconData, size: 24, color: iconColor),
                ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      bannerTitle,
                      style: TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.w800,
                        color: _status == PaymentRecoveryStatus.issue
                            ? const Color(0xFF991B1B)
                            : (_status == PaymentRecoveryStatus.pendingConfirmation
                                ? TagaTowersDesignTokens.warningAmberDark
                                : TagaTowersDesignTokens.deepEmerald),
                        letterSpacing: -0.3,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Text(
                      bannerDesc,
                      style: TextStyle(
                        fontSize: 12.5,
                        height: 1.4,
                        fontWeight: FontWeight.w400,
                        color: Colors.black.withValues(alpha: 0.75),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          if (_errorMessage != null && _status == PaymentRecoveryStatus.issue) ...[
            const SizedBox(height: 12),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFFFEF2F2),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: const Color(0xFFFCA5A5),
                  width: 1,
                ),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    LucideIcons.alertCircle,
                    size: 14,
                    color: Color(0xFFB91C1C),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _errorMessage!,
                      style: const TextStyle(
                        fontSize: 11.5,
                        color: Color(0xFF991B1B),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ] else if (_statusMessage != null && _status != PaymentRecoveryStatus.verified) ...[
            const SizedBox(height: 12),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.7),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: borderColor.withValues(alpha: 0.5),
                  width: 1,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    LucideIcons.info,
                    size: 14,
                    color: Colors.black.withValues(alpha: 0.6),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _statusMessage!,
                      style: TextStyle(
                        fontSize: 11.5,
                        color: Colors.black.withValues(alpha: 0.75),
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  /// Transaction and preserved identifiers card with copy actions
  Widget _buildTransactionCard() {
    final effectiveBookingId = _bookingId ?? 'Draft Holding';
    final effectiveOrderId = _razorpayOrderId ?? 'Pending Order ID';
    final effectivePaymentId = _razorpayPaymentId ?? 'Pending Payment ID';
    final effectiveAmount = _amount > 0 ? _amount : 200;

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Row(
                children: [
                  Icon(
                    LucideIcons.receipt,
                    size: 18,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'Transaction Record',
                    style: TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '₹$effectiveAmount Advance',
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          const Divider(height: 1, color: Color(0xFFE5E7EB)),
          const SizedBox(height: 14),

          // Booking Reference ID
          _buildDetailRow(
            label: 'Booking Reference ID',
            value: effectiveBookingId,
            canCopy: _bookingId != null && _bookingId!.isNotEmpty,
            onCopy: () => _copyToClipboard(effectiveBookingId, 'Booking ID'),
          ),
          const SizedBox(height: 10),

          // Razorpay Order ID
          _buildDetailRow(
            label: 'Razorpay Order ID',
            value: effectiveOrderId,
            canCopy: _razorpayOrderId != null && _razorpayOrderId!.isNotEmpty,
            onCopy: () => _copyToClipboard(effectiveOrderId, 'Order ID'),
          ),
          const SizedBox(height: 10),

          // Razorpay Payment ID
          _buildDetailRow(
            label: 'Razorpay Payment ID',
            value: effectivePaymentId,
            canCopy: _razorpayPaymentId != null && _razorpayPaymentId!.isNotEmpty,
            onCopy: () => _copyToClipboard(effectivePaymentId, 'Payment ID'),
            highlight: true,
          ),
          const SizedBox(height: 10),

          // Room / Stay Type if available
          if (_roomName != null && _roomName!.isNotEmpty) ...[
            _buildDetailRow(
              label: 'Selected Room',
              value: _roomName!,
              canCopy: false,
            ),
            const SizedBox(height: 10),
          ],

          // Facility Location
          _buildDetailRow(
            label: 'Facility Location',
            value: 'Velanmai Illam, Egmore, Chennai',
            canCopy: false,
          ),
        ],
      ),
    );
  }

  /// Individual row helper for transaction metadata with optional copy button
  Widget _buildDetailRow({
    required String label,
    required String value,
    required bool canCopy,
    VoidCallback? onCopy,
    bool highlight = false,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Expanded(
          flex: 4,
          child: Text(
            label,
            style: const TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w500,
              color: AppColors.textSecondary,
            ),
          ),
        ),
        Expanded(
          flex: 6,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              Flexible(
                child: Text(
                  value,
                  overflow: TextOverflow.ellipsis,
                  textAlign: TextAlign.right,
                  style: TextStyle(
                    fontSize: 12.5,
                    fontWeight: highlight ? FontWeight.w700 : FontWeight.w600,
                    color: highlight
                        ? TagaTowersDesignTokens.deepEmerald
                        : AppColors.textPrimary,
                    fontFamily: (canCopy && !highlight) ? 'monospace' : null,
                  ),
                ),
              ),
              if (canCopy) ...[
                const SizedBox(width: 6),
                InkWell(
                  onTap: onCopy,
                  borderRadius: BorderRadius.circular(6),
                  child: Padding(
                    padding: const EdgeInsets.all(4.0),
                    child: Icon(
                      LucideIcons.copy,
                      size: 13,
                      color: highlight
                          ? TagaTowersDesignTokens.primaryGreen
                          : AppColors.textSecondary,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  /// Action buttons adapting to verified vs. unverified status
  Widget _buildActionButtons() {
    if (_status == PaymentRecoveryStatus.verified) {
      return Column(
        children: [
          // View Booking Confirmation Button (Primary Success Action)
          BouncyButton(
            onTap: _navigateToConfirmation,
            child: Container(
              width: double.infinity,
              height: 52,
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: [
                    TagaTowersDesignTokens.deepEmerald,
                    TagaTowersDesignTokens.primaryGreen,
                  ],
                ),
                borderRadius: BorderRadius.circular(14),
                boxShadow: TagaTowersDesignTokens.floatingGlow,
              ),
              child: const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(LucideIcons.ticketCheck, color: Colors.white, size: 20),
                  SizedBox(width: 10),
                  Text(
                    'View Booking Confirmation',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                      letterSpacing: -0.2,
                    ),
                  ),
                  SizedBox(width: 8),
                  Icon(LucideIcons.arrowRight, color: Colors.white, size: 18),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          // Secondary button to towers home
          BouncyButton(
            onTap: _navigateToTowersHome,
            child: Container(
              width: double.infinity,
              height: 48,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                  color: TagaTowersDesignTokens.subtleBorder,
                  width: 1.2,
                ),
              ),
              child: const Center(
                child: Text(
                  'Back to TAGA Towers Home',
                  style: TextStyle(
                    color: AppColors.textPrimary,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
          ),
        ],
      );
    }

    return Column(
      children: [
        // Check / Re-verify Payment Status Button
        BouncyButton(
          onTap: _isChecking ? null : () => _checkPaymentStatus(),
          child: Container(
            width: double.infinity,
            height: 52,
            decoration: BoxDecoration(
              gradient: _isChecking
                  ? LinearGradient(
                      colors: [
                        Colors.grey.shade400,
                        Colors.grey.shade500,
                      ],
                    )
                  : const LinearGradient(
                      colors: [
                        TagaTowersDesignTokens.deepEmerald,
                        TagaTowersDesignTokens.primaryGreen,
                      ],
                    ),
              borderRadius: BorderRadius.circular(14),
              boxShadow: _isChecking
                  ? null
                  : TagaTowersDesignTokens.floatingGlow,
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (_isChecking)
                  const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                else
                  const Icon(LucideIcons.refreshCw, color: Colors.white, size: 18),
                const SizedBox(width: 10),
                Text(
                  _isChecking
                      ? 'Re-verifying Payment...'
                      : 'Check / Re-verify Payment Status',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    letterSpacing: -0.2,
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 12),

        // Back to TAGA Towers Home (Safe Exit)
        BouncyButton(
          onTap: _navigateToTowersHome,
          child: Container(
            width: double.infinity,
            height: 48,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: TagaTowersDesignTokens.border,
                width: 1.2,
              ),
            ),
            child: const Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  LucideIcons.building2,
                  size: 16,
                  color: AppColors.textSecondary,
                ),
                SizedBox(width: 8),
                Text(
                  'Back to TAGA Towers Home',
                  style: TextStyle(
                    color: AppColors.textPrimary,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  /// Critical Member Advisory Card alerting the user NOT to pay again
  Widget _buildMemberAdvisoryNotice() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBEB), // Amber-50
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFFDE68A), // Amber-200
          width: 1.2,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x0A000000),
            blurRadius: 8,
            offset: Offset(0, 2),
          ),
        ],
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(
                LucideIcons.shieldAlert,
                color: Color(0xFFB45309), // Amber-700
                size: 20,
              ),
              SizedBox(width: 8),
              Text(
                'Important Member Safety Advisory',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                  color: Color(0xFF92400E), // Amber-800
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          const Text(
            'Do not pay again. If money was deducted from your account, our team will verify your payment automatically or you can contact support with your Payment ID.',
            style: TextStyle(
              fontSize: 13,
              height: 1.45,
              fontWeight: FontWeight.w600,
              color: Color(0xFF78350F), // Amber-900
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Your booking reference remains preserved in the TAGA Towers system. Re-initiating payment may cause duplicate debits from your bank account.',
            style: TextStyle(
              fontSize: 11.5,
              height: 1.4,
              color: const Color(0xFF78350F).withValues(alpha: 0.85),
            ),
          ),
        ],
      ),
    );
  }

  /// Caretaker phone numbers and helpdesk email support hub
  Widget _buildSupportHub() {
    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(
                LucideIcons.headset,
                size: 18,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
              SizedBox(width: 8),
              Text(
                'On-Site Caretaker & Support Hub',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          const Text(
            'Available 24/7 at Velanmai Illam, Egmore to assist with payment verification',
            style: TextStyle(
              fontSize: 12,
              color: AppColors.textSecondary,
              height: 1.35,
            ),
          ),
          const SizedBox(height: 14),

          // Caretaker 1: Mr. Muthu
          _buildContactTile(
            name: 'Mr. Muthu',
            role: 'Caretaker, TAGA Towers',
            phoneDisplay: '96007 63744',
            phoneNumber: '+919600763744',
            icon: LucideIcons.userCheck,
          ),
          const SizedBox(height: 10),

          // Caretaker 2: Mr. Mariyappan
          _buildContactTile(
            name: 'Mr. Mariyappan',
            role: 'Caretaker, TAGA Towers',
            phoneDisplay: '96770 10300',
            phoneNumber: '+919677010300',
            icon: LucideIcons.userCheck,
          ),
          const SizedBox(height: 10),

          // Facility Reception Desk
          _buildContactTile(
            name: 'Reception & Helpdesk',
            role: 'Egmore Office Landline',
            phoneDisplay: '044 34919949',
            phoneNumber: '04434919949',
            icon: LucideIcons.phoneCall,
          ),
          const SizedBox(height: 12),

          // Email Support Button
          InkWell(
            onTap: _sendEmailSupport,
            borderRadius: BorderRadius.circular(12),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.mutedBackground,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.5),
                  width: 1,
                ),
              ),
              child: const Row(
                children: [
                  Icon(
                    LucideIcons.mail,
                    size: 16,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                  SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Email Payment Support',
                          style: TextStyle(
                            fontSize: 12.5,
                            fontWeight: FontWeight.bold,
                            color: TagaTowersDesignTokens.deepEmerald,
                          ),
                        ),
                        Text(
                          TagaTowersDesignTokens.helpdeskEmail,
                          style: TextStyle(
                            fontSize: 11.5,
                            color: AppColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Icon(
                    LucideIcons.externalLink,
                    size: 14,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Individual caretaker telephone contact tile with direct dial button
  Widget _buildContactTile({
    required String name,
    required String role,
    required String phoneDisplay,
    required String phoneNumber,
    required IconData icon,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFF9FAFB),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE5E7EB), width: 1),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.softMint,
              shape: BoxShape.circle,
            ),
            child: Icon(icon, size: 16, color: TagaTowersDesignTokens.deepEmerald),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                Text(
                  role,
                  style: const TextStyle(
                    fontSize: 11,
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
          BouncyButton(
            onTap: () => _makePhoneCall(phoneNumber),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 7),
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.deepEmerald,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(LucideIcons.phone, size: 13, color: Colors.white),
                  const SizedBox(width: 6),
                  Text(
                    phoneDisplay,
                    style: const TextStyle(
                      fontSize: 11.5,
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
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

  /// Footnote reminding member of zero duplicate payment risk
  Widget _buildDuplicateSafetyBadge() {
    return Center(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(999),
          border: Border.all(
            color: TagaTowersDesignTokens.subtleBorder,
            width: 1,
          ),
        ),
        child: const Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              LucideIcons.lock,
              size: 13,
              color: TagaTowersDesignTokens.primaryGreen,
            ),
            SizedBox(width: 6),
            Text(
              'Duplicate Safety Active: No new payment can be triggered here',
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w500,
                color: AppColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
