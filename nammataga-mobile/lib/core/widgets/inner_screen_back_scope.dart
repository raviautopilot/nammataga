import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';

/// Wraps any inner application screen to strictly enforce proper back navigation hierarchy:
///
/// 1. If an inner modal, dialog, or subroute can pop, it pops it first cleanly.
/// 2. If at the screen root, it safely navigates back to the Home Page (`/home`).
/// 3. Completely prevents an inner screen from directly exiting the application.
/// 4. Protected with a re-entrancy lock and direct pop calls to guarantee zero-crash
///    and prevent recursive event loop ANR freezes on Android/iOS back swipes.
class InnerScreenBackScope extends StatefulWidget {
  final Widget child;
  final VoidCallback? onBack;

  const InnerScreenBackScope({
    super.key,
    required this.child,
    this.onBack,
  });

  @override
  State<InnerScreenBackScope> createState() => _InnerScreenBackScopeState();
}

class _InnerScreenBackScopeState extends State<InnerScreenBackScope> {
  bool _isNavigating = false;

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop || _isNavigating) return;
        _isNavigating = true;

        HapticFeedback.lightImpact();

        if (widget.onBack != null) {
          widget.onBack!();
          _isNavigating = false;
          return;
        }

        // 1. If an inner modal, dialog, or pushed subroute can pop, pop it cleanly.
        // NOTE: Use pop(), NEVER maybePop(), to prevent recursive pop cycle lockups!
        final navigator = Navigator.maybeOf(context);
        if (navigator != null && navigator.canPop()) {
          navigator.pop(result);
          _isNavigating = false;
          return;
        }

        // 2. If GoRouter has history to pop, pop it.
        try {
          if (context.canPop()) {
            context.pop(result);
            _isNavigating = false;
            return;
          }
        } catch (_) {}

        // 3. Root of inner screen reached -> Safely navigate back to /home
        try {
          context.go('/home');
        } catch (_) {
          navigator?.pop(result);
        } finally {
          _isNavigating = false;
        }
      },
      child: widget.child,
    );
  }
}
