import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'app_exit_confirm_dialog.dart';

/// Production-grade exit guard wrapper.
///
/// Features:
/// 1. Double back-press / swipe to exit within [doubleBackCooldown] (default 2 seconds)
///    with a refined, non-intrusive floating SnackBar and light haptic feedback.
/// 2. Modal exit confirmation dialog with an instant "Exit Now" action and option to show
///    it explicitly or when requested.
/// 3. Safely verifies whether inner navigators (modals, sheets, sub-routes) have history to pop
///    before triggering exit logic, avoiding trapped flows or false exits.
class AppExitScope extends StatefulWidget {
  final Widget child;

  /// Custom duration window within which a second back-press exits the app.
  final Duration doubleBackCooldown;

  /// Whether the exit confirmation dialog should be presented directly on back-press
  /// instead of / in addition to double-press.
  final bool showConfirmDialog;

  const AppExitScope({
    super.key,
    required this.child,
    this.doubleBackCooldown = const Duration(seconds: 2),
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

  Future<void> _handleBackPress() async {
    // 1. If an inner navigator (e.g. nested modal, bottom sheet, or sub-route) can pop,
    // let it pop first.
    final navigator = Navigator.maybeOf(context);
    if (navigator != null && navigator.canPop()) {
      navigator.maybePop();
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

    // 3. Double-back / swipe again to exit mode with popup option on notice:
    final now = DateTime.now();
    final isWithinCooldown = _lastBackPressTime != null &&
        now.difference(_lastBackPressTime!) <= widget.doubleBackCooldown;

    if (isWithinCooldown) {
      // Second back within cooldown -> Close the application cleanly
      AppExitScope.exitApp();
    } else {
      _lastBackPressTime = now;
      HapticFeedback.lightImpact();

      if (!mounted) return;
      final messenger = ScaffoldMessenger.maybeOf(context);
      if (messenger != null) {
        messenger.clearSnackBars();
        messenger.showSnackBar(
          SnackBar(
            content: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(4),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: const Icon(
                    LucideIcons.undo2,
                    color: Colors.white,
                    size: 15,
                  ),
                ),
                const SizedBox(width: 10),
                const Expanded(
                  child: Text(
                    'Press back or swipe again to exit',
                    style: TextStyle(
                      fontFamily: 'Roboto',
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                      color: Colors.white,
                    ),
                  ),
                ),
              ],
            ),
            action: SnackBarAction(
              label: 'Exit',
              textColor: const Color(0xFFFFB4AB),
              onPressed: () async {
                messenger.hideCurrentSnackBar();
                final shouldExit =
                    await AppExitScope.showExitConfirmation(context);
                if (shouldExit && mounted) {
                  AppExitScope.exitApp();
                }
              },
            ),
            behavior: SnackBarBehavior.floating,
            backgroundColor: const Color(0xFF1B4332),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
              side: const BorderSide(color: Color(0x33D8F3DC), width: 1),
            ),
            duration: widget.doubleBackCooldown,
            margin: const EdgeInsets.fromLTRB(16, 0, 16, 18),
          ),
        );
      }
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
