import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';
import '../providers/resources_provider.dart';
import 'resources_access_restricted_screen.dart';
import 'widgets/category_chips_bar.dart';
import 'widgets/document_card.dart';
import 'widgets/external_links_view.dart';
import 'widgets/resources_shimmer.dart';
import 'widgets/subcategory_chips_bar.dart';

/// Main screen for NammaTAGA Resources Center.
///
/// Features:
/// - Access Guard check for paid member privilege
/// - Executive header card with emerald gradient and dynamic document count
/// - App bar with title, back button, and manual refresh trigger
/// - Category chips bar and subcategory filter bar
/// - Interactive documents list with PDF preview, download actions, and empty states
/// - External government links view for 'links' category
/// - Pull-to-refresh integration
class ResourcesScreen extends ConsumerStatefulWidget {
  const ResourcesScreen({super.key});

  @override
  ConsumerState<ResourcesScreen> createState() => _ResourcesScreenState();
}

class _ResourcesScreenState extends ConsumerState<ResourcesScreen> {
  Future<void> _handleRefresh() async {
    HapticFeedback.lightImpact();
    ref.invalidate(resourceCategoriesProvider);
    final catId = ref.read(selectedCategoryProvider);
    if (catId != null) {
      final subcat = ref.read(selectedSubcategoryProvider);
      ref.invalidate(
        categoryDocumentsProvider((categoryId: catId, subcategory: subcat)),
      );
      if (catId == 'links') {
        ref.invalidate(externalLinksProvider);
      }
    }
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

    // 2. Resource Categories & Current Selection
    final categoriesAsync = ref.watch(resourceCategoriesProvider);
    final selectedCatId = ref.watch(selectedCategoryProvider);
    final selectedSubcat = ref.watch(selectedSubcategoryProvider);

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: Column(
        children: [
          // Executive Header Card with Emerald Gradient & App Bar
          _buildExecutiveHeader(
            selectedCatId: selectedCatId,
            selectedSubcat: selectedSubcat,
          ),

          // Main Content Area with Categories & Document List
          Expanded(
            child: categoriesAsync.when(
              loading: () => const ResourcesShimmer(includeChips: true),
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

                // Identify currently active category object
                final currentCat = categories.firstWhere(
                  (c) => c.id == selectedCatId,
                  orElse: () => categories.first,
                );

                final hasSubcategories = currentCat.subcategories != null &&
                    currentCat.subcategories!.isNotEmpty;

                return Column(
                  children: [
                    const SizedBox(height: 8),

                    // Horizontally scrollable Category Chips Bar
                    CategoryChipsBar(
                      categories: categories,
                      selectedCategoryId: selectedCatId ?? currentCat.id,
                      onCategorySelected: (newCatId) {
                        ref
                            .read(selectedCategoryProvider.notifier)
                            .selectCategory(newCatId);
                      },
                    ),

                    // Horizontally scrollable Subcategory Chips Bar (if available)
                    if (hasSubcategories) ...[
                      const SizedBox(height: 6),
                      SubcategoryChipsBar(
                        subcategories: currentCat.subcategories!,
                        selectedSubcategory: selectedSubcat,
                        onSubcategorySelected: (newSubcat) {
                          ref
                              .read(selectedSubcategoryProvider.notifier)
                              .selectSubcategory(newSubcat);
                        },
                      ),
                    ],

                    const SizedBox(height: 6),

                    // Document List or External Links View
                    Expanded(
                      child: RefreshIndicator(
                        onRefresh: _handleRefresh,
                        color: AppColors.primary,
                        backgroundColor: Colors.white,
                        child: (selectedCatId == 'links')
                            ? const ExternalLinksView()
                            : _buildDocumentsList(
                                categoryId: selectedCatId ?? currentCat.id,
                                categoryName: currentCat.name,
                                subcategory: selectedSubcat,
                              ),
                      ),
                    ),
                  ],
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the executive top header card with rich emerald gradient,
  /// navigation controls, and document summary count.
  Widget _buildExecutiveHeader({
    required String? selectedCatId,
    required String? selectedSubcat,
  }) {
    // Watch current documents to extract live document count
    int? documentCount;
    if (selectedCatId != null && selectedCatId != 'links') {
      final docsAsync = ref.watch(categoryDocumentsProvider((
        categoryId: selectedCatId,
        subcategory: selectedSubcat,
      )));
      documentCount = docsAsync.asData?.value.length;
    }

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

                  // Dynamic Document Count Pill (if loaded)
                  if (documentCount != null) ...[
                    const SizedBox(height: 10),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 3.5,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.12),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: Colors.white.withValues(alpha: 0.20),
                          width: 0.9,
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(
                            LucideIcons.fileCheck2,
                            size: 13,
                            color: Color(0xFFB7E4C7),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            '$documentCount ${documentCount == 1 ? 'document' : 'documents'} available',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 11.5,
                              fontWeight: FontWeight.w600,
                              letterSpacing: 0.2,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the reactive document cards list for the active category.
  Widget _buildDocumentsList({
    required String categoryId,
    required String categoryName,
    required String? subcategory,
  }) {
    final docsAsync = ref.watch(categoryDocumentsProvider((
      categoryId: categoryId,
      subcategory: subcategory,
    )));

    return docsAsync.when(
      loading: () => const ResourcesShimmer(itemCount: 6),
      error: (err, stack) => _buildErrorState(
        err.toString().replaceAll('Exception: ', ''),
        onRetry: () => ref.invalidate(categoryDocumentsProvider((
          categoryId: categoryId,
          subcategory: subcategory,
        ))),
      ),
      data: (documents) {
        if (documents.isEmpty) {
          return _buildEmptyState(
            title: 'No Documents Found',
            message: subcategory != null
                ? 'No documents found for "$subcategory" in this category.'
                : 'No documents have been uploaded for $categoryName yet.',
          );
        }

        return ListView.separated(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          itemCount: documents.length,
          separatorBuilder: (context, index) => const SizedBox(height: 12),
          itemBuilder: (context, index) {
            final doc = documents[index];
            return DocumentCard(
              document: doc,
              categoryName: categoryName,
            );
          },
        );
      },
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
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
        child: Container(
          padding: const EdgeInsets.all(22),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(
              color: const Color(0xFFFFCDD2),
              width: 1.1,
            ),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF1B4332).withValues(alpha: 0.04),
                blurRadius: 14,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 52,
                height: 52,
                decoration: const BoxDecoration(
                  color: Color(0xFFFFEBEE),
                  shape: BoxShape.circle,
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.alertCircle,
                    size: 28,
                    color: AppColors.destructive,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              const Text(
                'Unable to Load Resources',
                style: TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 16.5,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: AppColors.textSecondary,
                  fontSize: 13,
                  height: 1.35,
                ),
              ),
              const SizedBox(height: 18),
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  onRetry();
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 10,
                  ),
                  decoration: BoxDecoration(
                    color: AppColors.primary,
                    borderRadius: BorderRadius.circular(12),
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
                          fontSize: 13.5,
                          fontWeight: FontWeight.w700,
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
