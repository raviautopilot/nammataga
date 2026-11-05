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
import '../../features/resources/presentation/pdf_viewer_screen.dart';
import '../../features/resources/presentation/resources_access_restricted_screen.dart';
import '../../features/resources/presentation/resources_screen.dart';
import '../../features/splash/presentation/splash_screen.dart';
import '../../features/tagatowers/domain/booking_model.dart';
import '../../features/tagatowers/presentation/booking_bed_selection_screen.dart';
import '../../features/tagatowers/presentation/booking_confirmation_screen.dart';
import '../../features/tagatowers/presentation/booking_details_screen.dart';
import '../../features/tagatowers/presentation/booking_review_screen.dart';
import '../../features/tagatowers/presentation/cancellation_screen.dart';
import '../../features/tagatowers/presentation/guest_details_screen.dart';
import '../../features/tagatowers/presentation/my_bookings_screen.dart';
import '../../features/tagatowers/presentation/payment_recovery_screen.dart';
import '../../features/tagatowers/presentation/payment_review_screen.dart';
import '../../features/tagatowers/presentation/room_availability_screen.dart';
import '../../features/tagatowers/presentation/tagatowers_home_screen.dart';

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
        child: const ResourcesScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
      routes: [
        GoRoute(
          path: 'pdf-viewer',
          name: 'resources-pdf-viewer',
          pageBuilder: (context, state) {
            final extra = state.extra as Map<String, dynamic>? ?? {};
            final title = extra['title'] as String? ??
                state.uri.queryParameters['title'] ??
                'Document Viewer';
            final url = extra['url'] as String? ??
                state.uri.queryParameters['url'] ??
                '';
            final localPath = extra['localPath'] as String? ??
                state.uri.queryParameters['localPath'];
            return CustomTransitionPage(
              key: state.pageKey,
              child: PdfViewerScreen(
                title: title,
                url: url,
                localPath: localPath,
              ),
              transitionsBuilder:
                  (context, animation, secondaryAnimation, child) =>
                      FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'restricted',
          name: 'resources-restricted',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const ResourcesAccessRestrictedScreen(),
            transitionsBuilder:
                (context, animation, secondaryAnimation, child) =>
                    FadeTransition(opacity: animation, child: child),
          ),
        ),
      ],
    ),
    GoRoute(
      path: '/taga-towers',
      name: 'taga-towers',
      pageBuilder: (context, state) => CustomTransitionPage(
        key: state.pageKey,
        child: const TagaTowersHomeScreen(),
        transitionsBuilder: (context, animation, secondaryAnimation, child) =>
            FadeTransition(opacity: animation, child: child),
      ),
      routes: [
        GoRoute(
          path: 'availability',
          name: 'taga-towers-availability',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const RoomAvailabilityScreen(),
            transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'my-bookings',
          name: 'taga-towers-my-bookings',
          pageBuilder: (context, state) {
            final tabIndexStr = state.uri.queryParameters['tab'];
            final tabIndex = int.tryParse(tabIndexStr ?? '') ?? 0;
            return CustomTransitionPage(
              key: state.pageKey,
              child: MyBookingsScreen(initialTabIndex: tabIndex),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'confirmation',
          name: 'taga-towers-confirmation',
          pageBuilder: (context, state) {
            final extra = state.extra;
            BookingResponse? booking;
            String? paymentId;
            String? orderId;

            if (extra is BookingResponse) {
              booking = extra;
            } else if (extra is Map<String, dynamic>) {
              if (extra['booking'] is BookingResponse) {
                booking = extra['booking'] as BookingResponse;
              }
              paymentId = extra['paymentId']?.toString();
              orderId = extra['orderId']?.toString();
            }

            booking ??= const BookingResponse(
              id: '',
              roomId: '',
              roomName: 'TAGA Towers Room',
              checkInDate: '',
              checkOutDate: '',
              bookerName: '',
              bookerId: '',
              bookerPhone: '',
              bookingFor: 'self',
              bedCount: 1,
              gender: 'male',
              guestDetails: [],
              paymentStatus: 'confirmed',
              advanceAmount: 0,
              bookingStatus: 'confirmed',
            );

            return CustomTransitionPage(
              key: state.pageKey,
              child: BookingConfirmationScreen(
                booking: booking,
                razorpayPaymentId: paymentId,
                razorpayOrderId: orderId,
              ),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'payment-recovery',
          name: 'taga-towers-payment-recovery',
          pageBuilder: (context, state) {
            final extra = state.extra;
            String? bookingId;
            String? orderId;
            String? paymentId;
            String? signature;
            int? amount;
            String? roomName;

            if (extra is Map<String, dynamic>) {
              if (extra['booking'] is BookingResponse) {
                final b = extra['booking'] as BookingResponse;
                bookingId = b.id;
                amount = b.advanceAmount;
                roomName = b.roomName;
              } else if (extra['booking'] is Map) {
                final bMap = extra['booking'] as Map;
                bookingId = bMap['id']?.toString();
                amount = bMap['advanceAmount'] is int ? bMap['advanceAmount'] as int : null;
                roomName = bMap['roomName']?.toString();
              }
              bookingId ??= extra['bookingId']?.toString();
              orderId = extra['orderId']?.toString() ?? extra['razorpayOrderId']?.toString();
              paymentId = extra['paymentId']?.toString() ?? extra['razorpayPaymentId']?.toString();
              signature = extra['signature']?.toString() ?? extra['razorpaySignature']?.toString();
              amount ??= extra['amount'] is int ? extra['amount'] as int : null;
              roomName ??= extra['roomName']?.toString();
            }

            bookingId ??= state.uri.queryParameters['bookingId'];
            orderId ??= state.uri.queryParameters['orderId'];
            paymentId ??= state.uri.queryParameters['paymentId'];

            return CustomTransitionPage(
              key: state.pageKey,
              child: PaymentRecoveryScreen(
                bookingId: bookingId,
                razorpayOrderId: orderId,
                razorpayPaymentId: paymentId,
                razorpaySignature: signature,
                amount: amount,
                roomName: roomName,
              ),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'booking/bed-selection',
          name: 'taga-towers-bed-selection',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const BookingBedSelectionScreen(),
            transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'booking/guest-details',
          name: 'taga-towers-guest-details',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const GuestDetailsScreen(),
            transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'booking/review',
          name: 'taga-towers-booking-review',
          pageBuilder: (context, state) => CustomTransitionPage(
            key: state.pageKey,
            child: const BookingReviewScreen(),
            transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                FadeTransition(opacity: animation, child: child),
          ),
        ),
        GoRoute(
          path: 'booking/payment-review',
          name: 'taga-towers-payment-review',
          pageBuilder: (context, state) {
            final extra = state.extra;
            BookingResponse? booking;
            if (extra is BookingResponse) {
              booking = extra;
            } else if (extra is Map<String, dynamic> && extra['booking'] is BookingResponse) {
              booking = extra['booking'] as BookingResponse;
            }
            return CustomTransitionPage(
              key: state.pageKey,
              child: PaymentReviewScreen(booking: booking),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'booking/details/:id',
          name: 'taga-towers-booking-details',
          pageBuilder: (context, state) {
            final bookingId = state.pathParameters['id'] ?? '';
            final extra = state.extra;
            BookingResponse? initialBooking;
            if (extra is BookingResponse) {
              initialBooking = extra;
            } else if (extra is Map<String, dynamic> && extra['booking'] is BookingResponse) {
              initialBooking = extra['booking'] as BookingResponse;
            }
            return CustomTransitionPage(
              key: state.pageKey,
              child: BookingDetailsScreen(
                bookingId: bookingId,
                initialBooking: initialBooking,
              ),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'booking/cancel/:id',
          name: 'taga-towers-booking-cancel-id',
          pageBuilder: (context, state) {
            final bookingId = state.pathParameters['id'] ?? '';
            final extra = state.extra;
            BookingResponse? booking;
            if (extra is BookingResponse) {
              booking = extra;
            } else if (extra is Map<String, dynamic> && extra['booking'] is BookingResponse) {
              booking = extra['booking'] as BookingResponse;
            }
            return CustomTransitionPage(
              key: state.pageKey,
              child: CancellationScreen(
                booking: booking,
                bookingId: bookingId,
              ),
              transitionsBuilder: (context, animation, secondaryAnimation, child) =>
                  FadeTransition(opacity: animation, child: child),
            );
          },
        ),
        GoRoute(
          path: 'cancel',
          name: 'taga-towers-cancel',
          pageBuilder: (context, state) {
            final extra = state.extra;
            BookingResponse? booking;
            String? bookingId;
            if (extra is BookingResponse) {
              booking = extra;
              bookingId = extra.id;
            } else if (extra is String) {
              bookingId = extra;
            } else if (extra is Map<String, dynamic>) {
              if (extra['booking'] is BookingResponse) {
                booking = extra['booking'] as BookingResponse;
              }
              bookingId = extra['bookingId']?.toString() ?? booking?.id;
            }
            bookingId ??= state.uri.queryParameters['bookingId'] ??
                state.uri.queryParameters['id'];

            return CustomTransitionPage(
              key: state.pageKey,
              child: CancellationScreen(
                booking: booking,
                bookingId: bookingId,
              ),
              transitionsBuilder:
                  (context, animation, secondaryAnimation, child) =>
                      FadeTransition(opacity: animation, child: child),
            );
          },
        ),
      ],
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
