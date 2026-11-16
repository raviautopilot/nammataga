import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'app_exit_confirm_dialog.dart';
import 'taga_toast.dart';

/// Production-grade exit guard wrapper enforcing the two-step exit flow:
///
/// 1st Back Press: Displays the improvised Option A floating capsule toast notice
///                 for 5 seconds. If the user does not press back again, it vanishes.
/// 2nd Back Press (within 5 seconds): Displays the Exit Confirmation pop-up dialog
///                 before closing the application.
class AppExitScope extends StatefulWidget {
  final Widget child;

  /// Custom duration window within which a second back-press triggers the exit dialog.
  /// Defaults to 5 seconds per requirements.
  final Duration doubleBackCooldown;

  /// Whether the exit confirmation dialog should be presented directly on back-press.
  final bool showConfirmDialog;

  const AppExitScope({
    super.key,
    required this.child,
    this.doubleBackCooldown = const Duration(seconds: 5),
    this.showConfirmDialog = false,
  });

  /// Static helper to trigger the exit confirmation popup from anywhere in the app.
  static Future<bool> showExitConfirmation(BuildContext context) async {
    HapticFeedback.mediumImpact();
    final shouldExit = await showDialog<bool>(
      context: context,
      barrierDismissible: true,
      builder: (dialogCtx) => const AppExitConfirmDialog(),
    );
    return shouldExit == true;
  }

  /// Exits the application cleanly across Android/iOS platforms.
  static void exitApp() {
    SystemNavigator.pop();
  }

  @override
  State<AppExitScope> createState() => _AppExitScopeState();
}

class _AppExitScopeState extends State<AppExitScope> {
  DateTime? _lastBackPressTime;
  bool _isNavigating = false;

  Future<void> _handleBackPress() async {
    if (_isNavigating) return;
    _isNavigating = true;

    try {
      // 1. If an inner navigator (e.g. nested modal, bottom sheet, or sub-route) can pop,
      // let it pop first cleanly without recursive loops.
      final navigator = Navigator.maybeOf(context);
      if (navigator != null && navigator.canPop()) {
        navigator.pop();
        return;
      }

      // 2. If explicit confirmation dialog mode is active:
      if (widget.showConfirmDialog) {
        final shouldExit = await AppExitScope.showExitConfirmation(context);
        if (shouldExit && mounted) {
          AppExitScope.exitApp();
        }
        return;
      }

      // 3. Two-step back-press flow
      final now = DateTime.now();
      final isWithinCooldown = _lastBackPressTime != null &&
          now.difference(_lastBackPressTime!) <= widget.doubleBackCooldown;

      if (isWithinCooldown) {
        // 2nd back press within 5s -> Show Exit Confirmation pop-up before closing!
        _lastBackPressTime = null;
        if (!mounted) return;
        ScaffoldMessenger.maybeOf(context)?.hideCurrentSnackBar();
        final shouldExit = await AppExitScope.showExitConfirmation(context);
        if (shouldExit && mounted) {
          AppExitScope.exitApp();
        }
      } else {
        // 1st back press -> Record time, light haptic feedback, and show 5-second exit toast
        _lastBackPressTime = now;
        HapticFeedback.lightImpact();

        if (!mounted) return;
        TagaToast.exitNotice(
          context,
          duration: widget.doubleBackCooldown,
          onExitTapped: () async {
            ScaffoldMessenger.maybeOf(context)?.hideCurrentSnackBar();
            final shouldExit = await AppExitScope.showExitConfirmation(context);
            if (shouldExit && mounted) {
              AppExitScope.exitApp();
            }
          },
        );
      }
    } finally {
      _isNavigating = false;
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        _handleBackPress();
      },
      child: widget.child,
    );
  }
}
