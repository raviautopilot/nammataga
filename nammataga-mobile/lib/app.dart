import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';
import 'core/widgets/web_phone_frame.dart';

/// Root application widget configuring theme, router, and web phone frame
class NammaTagaApp extends ConsumerWidget {
  const NammaTagaApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp.router(
      title: 'Namma TAGA',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      routerConfig: appRouter,
      builder: (context, child) {
        return WebPhoneFrame(child: child ?? const SizedBox.shrink());
      },
    );
  }
}
