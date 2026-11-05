import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';
import '../domain/resource_category.dart';
import '../providers/resources_provider.dart';
import 'category_documents_screen.dart';
import 'resources_access_restricted_screen.dart';
import 'widgets/category_card.dart';
import 'widgets/resources_shimmer.dart';

/// Resources Landing Page – Category Directory.
///
/// Features:
/// - Strict Access Guard check for active paid membership privilege
/// - Executive header card with emerald gradient and brand identity
/// - "Browse Categories" section heading with professional enterprise typography
/// - Vertically stacked, full-width category cards with contextual icons and right arrow
/// - Complete category names without truncation (natural text wrapping)
/// - Normal vertical scrolling with pull-to-refresh
/// - Tapping a category navigates to [CategoryDocumentsScreen]
class ResourcesScreen extends ConsumerStatefulWidget {
  const ResourcesScreen({super.key});

  @override
  ConsumerState<ResourcesScreen> createState() => _ResourcesScreenState();
}

class _ResourcesScreenState extends ConsumerState<ResourcesScreen> {
  Future<void> _handleRefresh() async {
    HapticFeedback.lightImpact();
    ref.invalidate(resourceCategoriesProvider);
  }

  void _handleBack() {
    HapticFeedback.lightImpact();
    try {
      if (context.canPop()) {
        context.pop();
        return;
      }
    } catch (_) {}
    try {
      context.go('/home');
    } catch (_) {
      Navigator.of(context).maybePop();
    }
  }

  void _navigateToCategory(ResourceCategory category) {
    HapticFeedback.lightImpact();
    ref.read(selectedCategoryProvider.notifier).selectCategory(category.id);

    try {
      context.push(
        '/resources/category',
        extra: category,
      );
    } catch (_) {
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => CategoryDocumentsScreen(category: category),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    // 1. Access Guard Check
    final profileAsync = ref.watch(memberProfileProvider);
    final authSession = ref.watch(authProvider).value;
    final isPaid = authSession?.isPaid == true ||
        (profileAsync.hasValue &&
            (profileAsync.value!.isPaid ||
                profileAsync.value!.subscriptionActive));

    if (profileAsync.isLoading && authSession?.isPaid != true) {
      return const Scaffold(
        backgroundColor: Color(0xFFF9FAF7),
        body: Center(
          child: CircularProgressIndicator(
            color: AppColors.primary,
          ),
        ),
      );
    }

    if (!isPaid) {
      return const ResourcesAccessRestrictedScreen();
    }

    // 2. Resource Categories
    final categoriesAsync = ref.watch(resourceCategoriesProvider);

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Executive Header Card with Emerald Gradient & App Bar
          _buildExecutiveHeader(),

          // Main Category Directory
          Expanded(
            child: categoriesAsync.when(
              loading: () => Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeading(),
                  const Expanded(child: CategoryDirectoryShimmer()),
                ],
              ),
              error: (err, stack) => _buildErrorState(
                'Unable to load resource categories: $err',
                onRetry: () => ref.invalidate(resourceCategoriesProvider),
              ),
              data: (categories) {
                if (categories.isEmpty) {
                  return _buildEmptyState(
                    title: 'No Resource Categories',
                    message:
                        'No resource categories are available from the server.',
                  );
                }

                return RefreshIndicator(
                  onRefresh: _handleRefresh,
                  color: AppColors.primary,
                  backgroundColor: Colors.white,
                  child: ListView.builder(
                    physics: const AlwaysScrollableScrollPhysics(
                      parent: BouncingScrollPhysics(),
                    ),
                    padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                    itemCount: categories.length + 1,
                    itemBuilder: (context, index) {
                      if (index == 0) {
                        return _buildSectionHeading();
                      }

                      final category = categories[index - 1];
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 10),
                        child: CategoryCard(
                          category: category,
                          onTap: () => _navigateToCategory(category),
                        ),
                      );
                    },
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  /// Section heading: "Browse Categories"
  Widget _buildSectionHeading() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(0, 18, 0, 14),
      child: Row(
        children: [
          Container(
            width: 3.5,
            height: 18,
            decoration: BoxDecoration(
              color: AppColors.primary,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(width: 8),
          const Text(
            'Browse Categories',
            style: TextStyle(
              color: AppColors.textPrimary,
              fontSize: 17,
              fontWeight: FontWeight.w700,
              letterSpacing: -0.2,
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the executive top header card with rich emerald gradient,
  /// navigation controls, title, and subtitle (document-count badge removed).
  Widget _buildExecutiveHeader() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: const BorderRadius.vertical(
          bottom: Radius.circular(28),
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF04190E).withValues(alpha: 0.35),
            blurRadius: 20,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          // Multi-stop Emerald Brand Gradient
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    Color(0xFF04170D),
                    Color(0xFF0A2B1B),
                    Color(0xFF14452F),
                    Color(0xFF1D5A3D),
                  ],
                  stops: [0.0, 0.35, 0.72, 1.0],
                ),
              ),
            ),
          ),

          // Ambient Top-Right Radial Glow
          Positioned(
            top: -45,
            right: -35,
            width: 200,
            height: 200,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF52B788).withValues(alpha: 0.28),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Ambient Bottom-Left Glow
          Positioned(
            bottom: -40,
            left: -30,
            width: 170,
            height: 170,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF2D6A4F).withValues(alpha: 0.30),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Subtle File Text Watermark Motif
          Positioned(
            right: -10,
            bottom: -15,
            child: IgnorePointer(
              child: Icon(
                LucideIcons.fileText,
                size: 130,
                color: Colors.white.withValues(alpha: 0.04),
              ),
            ),
          ),

          // Foreground Content
          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Navigation Bar Row (Back Button, Center Badge, Refresh Button)
                  Row(
                    children: [
                      // Back Button
                      BouncyButton(
                        onTap: _handleBack,
                        scaleDownFactor: 0.90,
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.14),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.25),
                              width: 1.1,
                            ),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.arrowLeft,
                              size: 18,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),

                      // Pill Badge
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color:
                              const Color(0xFF2D6A4F).withValues(alpha: 0.85),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: const Color(0xFF52B788)
                                .withValues(alpha: 0.40),
                            width: 1.0,
                          ),
                        ),
                        child: const Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              LucideIcons.shieldCheck,
                              size: 12,
                              color: Color(0xFFB7E4C7),
                            ),
                            SizedBox(width: 5),
                            Text(
                              'Official Portal',
                              style: TextStyle(
                                color: Color(0xFFD8F3DC),
                                fontSize: 11,
                                fontWeight: FontWeight.w700,
                                letterSpacing: 0.4,
                              ),
                            ),
                          ],
                        ),
                      ),
                      const Spacer(),

                      // Refresh Button
                      BouncyButton(
                        onTap: _handleRefresh,
                        scaleDownFactor: 0.90,
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.14),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.25),
                              width: 1.1,
                            ),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.refreshCw,
                              size: 16,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 14),

                  // Title & Subtitle
                  const Text(
                    'Resources Center',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 24,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.4,
                    ),
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Government orders, circulars & official applications',
                    style: TextStyle(
                      color: Color(0xFFD8F3DC),
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                      height: 1.3,
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

  /// Styled empty state presentation.
  Widget _buildEmptyState({
    required String title,
    required String message,
  }) {
    return Center(
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(
          parent: BouncingScrollPhysics(),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 48),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 68,
              height: 68,
              decoration: BoxDecoration(
                color: const Color(0xFFF1F8E9),
                shape: BoxShape.circle,
                border: Border.all(
                  color: const Color(0xFFD8F3DC),
                  width: 1.2,
                ),
              ),
              child: const Center(
                child: Icon(
                  LucideIcons.fileQuestion,
                  size: 32,
                  color: AppColors.primary,
                ),
              ),
            ),
            const SizedBox(height: 16),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: AppColors.textPrimary,
                fontSize: 17,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 6),
            ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 300),
              child: Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: AppColors.textSecondary,
                  fontSize: 13.5,
                  height: 1.4,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Styled error state presentation with retry action button.
  Widget _buildErrorState(
    String message, {
    required VoidCallback onRetry,
  }) {
    return Center(
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(
          parent: BouncingScrollPhysics(),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 48),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: const Color(0xFFFEF2F2),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: const Color(0xFFFECACA),
              width: 1.2,
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: const BoxDecoration(
                  color: Color(0xFFFEE2E2),
                  shape: BoxShape.circle,
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.alertTriangle,
                    size: 24,
                    color: Color(0xFFDC2626),
                  ),
                ),
              ),
              const SizedBox(height: 12),
              const Text(
                'Failed to load categories',
                style: TextStyle(
                  color: Color(0xFF991B1B),
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: Color(0xFFB91C1C),
                  fontSize: 13,
                  height: 1.35,
                ),
              ),
              const SizedBox(height: 16),
              BouncyButton(
                onTap: onRetry,
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 18,
                    vertical: 9,
                  ),
                  decoration: BoxDecoration(
                    color: AppColors.primary,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        LucideIcons.refreshCw,
                        size: 14,
                        color: Colors.white,
                      ),
                      SizedBox(width: 8),
                      Text(
                        'Retry',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
