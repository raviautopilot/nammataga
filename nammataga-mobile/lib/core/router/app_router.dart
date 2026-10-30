import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../core/constants/app_colors.dart';
import '../../core/storage/secure_storage_service.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/events/presentation/events_screen.dart';
import '../../features/grievance/presentation/grievance_form_screen.dart';
import '../../features/grievance/presentation/grievance_intro_screen.dart';
import '../../features/grievance/presentation/grievance_success_screen.dart';
import '../../features/grievance/presentation/widgets/grievance_access_restricted_screen.dart';
import '../../features/home/presentation/home_screen.dart';
import '../../features/profile/presentation/account_settings_screen.dart';
import '../../features/profile/presentation/address_details_screen.dart';
import '../../features/profile/presentation/announcements_screen.dart';
import '../../features/profile/presentation/contact_details_screen.dart';
import '../../features/profile/presentation/employment_details_screen.dart';
import '../../features/profile/presentation/membership_details_screen.dart';
import '../../features/profile/presentation/official_details_screen.dart';
import '../../features/profile/presentation/personal_info_screen.dart';
import '../../features/profile/presentation/profile_overview_screen.dart';
import '../../features/profile/presentation/request_profile_edit_screen.dart';
import '../../features/profile/presentation/subscription_plans_screen.dart';
import '../../features/splash/presentation/splash_screen.dart';

/// Factory to create an isolated GoRouter instance (useful for testing and fresh state)
GoRouter createRouter({String initialLocation = '/splash'}) {
  return GoRouter(
    initialLocation: initialLocation,
  refreshListenable: StorageService.authStateListenable,
  redirect: (context, state) {
    final hasValidToken = StorageService.isAuthenticatedSync;
    final loc = state.matchedLocation;
    final isAuthRoute = loc == '/login' || loc == '/splash';

    // If unauthenticated, prohibit entering internal application routes
    if (!hasValidToken) {
      if (!isAuthRoute) {
        return '/login';
      }
      return null;
    }

    // If authenticated and trying to visit login, redirect to home
    if (loc == '/login') {
      return '/home';
    }

    return null;
  },
  routes: [
    GoRoute(
      path: '/splash',
      name: 'splash',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const SplashScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) {
          return FadeTransition(opacity: animation, child: child);
        },
      ),
    ),
    GoRoute(
      path: '/login',
      name: 'login',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: LoginScreen(
          onLoginSuccess: () {
            context.go('/home');
          },
        ),
        transitionsBuilder: (context, animation, secondaryAnimation, child) {
          return FadeTransition(opacity: animation, child: child);
        },
      ),
    ),
    GoRoute(
      path: '/home',
      name: 'home',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const HomeScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) {
          return FadeTransition(opacity: animation, child: child);
        },
      ),
    ),
    // Stubs for forthcoming phases
    GoRoute(
      path: '/office-bearers',
      name: 'office-bearers',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const _PlaceholderScreen(
          title: 'Office Bearers',
          phase: 'Phase 04',
        ),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
    ),
    GoRoute(
      path: '/events',
      name: 'events',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const EventsScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
    ),
    GoRoute(
      path: '/profile',
      name: 'profile',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const ProfileOverviewScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
      routes: [
        GoRoute(
          path: 'personal',
          name: 'profile-personal',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const PersonalInfoScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'contact',
          name: 'profile-contact',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const ContactDetailsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'employment',
          name: 'profile-employment',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const EmploymentDetailsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'address',
          name: 'profile-address',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const AddressDetailsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'official',
          name: 'profile-official',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const OfficialDetailsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'membership',
          name: 'profile-membership',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const MembershipDetailsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'edit-request',
          name: 'profile-edit-request',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const RequestProfileEditScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'settings',
          name: 'profile-settings',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const AccountSettingsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'subscriptions',
          name: 'profile-subscriptions',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const SubscriptionPlansScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'announcements',
          name: 'profile-announcements',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const AnnouncementsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'notifications',
          name: 'profile-notifications',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const AnnouncementsScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
      ],
    ),
    GoRoute(
      path: '/grievance',
      name: 'grievance',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const GrievanceIntroScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
      routes: [
        GoRoute(
          path: 'form',
          name: 'grievance-form',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const GrievanceFormScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'success',
          name: 'grievance-success',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const GrievanceSuccessScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'restricted',
          name: 'grievance-restricted',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const GrievanceAccessRestrictedScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
      ],
    ),
    GoRoute(
      path: '/resources',
      name: 'resources',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const _PlaceholderScreen(
          title: 'Member Resources',
          phase: 'Phase 08',
        ),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
    ),
    GoRoute(
      path: '/taga-towers',
      name: 'taga-towers',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const _PlaceholderScreen(
          title: 'TAGA Towers Booking',
          phase: 'Phase 09',
        ),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
    ),
  ],
  );
}

/// Global application router singleton
final GoRouter appRouter = createRouter();

class _PlaceholderScreen extends StatelessWidget {
  final String title;
  final String phase;

  const _PlaceholderScreen({required this.title, required this.phase});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title), backgroundColor: Colors.white),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: AppColors.secondary,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  phase,
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: AppColors.primaryDark,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              Text(
                title,
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 6),
              const Text(
                'This screen is queued in the development roadmap.',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
              ),
              const SizedBox(height: 20),
              OutlinedButton(
                onPressed: () => context.pop(),
                child: const Text('Back to Dashboard'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
