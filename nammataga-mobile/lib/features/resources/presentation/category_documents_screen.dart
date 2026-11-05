import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../domain/resource_category.dart';
import '../providers/resources_provider.dart';
import 'widgets/document_card.dart';
import 'widgets/external_links_view.dart';
import 'widgets/resources_shimmer.dart';
import 'widgets/subcategory_chips_bar.dart';

/// Screen displaying the documents of a selected [ResourceCategory].
///
/// Features:
/// - Executive header displaying the selected category name and document count
/// - Subcategory filter chips bar (e.g. 'All', 'Central', 'State') if category has subcategories
/// - Direct rendering of [ExternalLinksView] when category is 'links'
/// - Reactive document list with PDF preview, download actions, share, print, and empty states
/// - Pull-to-refresh integration
/// - Seamless back navigation returning to the Category Directory
class CategoryDocumentsScreen extends ConsumerStatefulWidget {
  final ResourceCategory category;

  const CategoryDocumentsScreen({
    super.key,
    required this.category,
  });

  @override
  ConsumerState<CategoryDocumentsScreen> createState() =>
      _CategoryDocumentsScreenState();
}

class _CategoryDocumentsScreenState
    extends ConsumerState<CategoryDocumentsScreen> {
  String? _selectedSubcategory;

  Future<void> _handleRefresh() async {
    HapticFeedback.lightImpact();
    if (widget.category.id == 'links') {
      ref.invalidate(externalLinksProvider);
    } else {
      ref.invalidate(
        categoryDocumentsProvider((
          categoryId: widget.category.id,
          subcategory: _selectedSubcategory,
        )),
      );
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
      context.go('/resources');
    } catch (_) {
      Navigator.of(context).maybePop();
    }
  }

  @override
  Widget build(BuildContext context) {
    final hasSubcategories = widget.category.subcategories != null &&
        widget.category.subcategories!.isNotEmpty;

    // Document count for header
    int? documentCount;
    if (widget.category.id != 'links') {
      final docsAsync = ref.watch(categoryDocumentsProvider((
        categoryId: widget.category.id,
        subcategory: _selectedSubcategory,
      )));
      documentCount = docsAsync.asData?.value.length;
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: Column(
        children: [
          // Executive Header Card with Selected Category Name
          _buildExecutiveHeader(documentCount: documentCount),

          // Subcategory Filter Chips Bar (if available for this category)
          if (hasSubcategories) ...[
            const SizedBox(height: 10),
            SubcategoryChipsBar(
              subcategories: widget.category.subcategories!,
              selectedSubcategory: _selectedSubcategory,
              onSubcategorySelected: (newSubcat) {
                setState(() {
                  _selectedSubcategory = newSubcat;
                });
              },
            ),
            const SizedBox(height: 6),
          ] else
            const SizedBox(height: 8),

          // Main Document List or External Links View
          Expanded(
            child: RefreshIndicator(
              onRefresh: _handleRefresh,
              color: AppColors.primary,
              backgroundColor: Colors.white,
              child: (widget.category.id == 'links')
                  ? const ExternalLinksView()
                  : _buildDocumentsList(),
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the executive top header card with rich emerald gradient,
  /// back navigation to directory, selected category name, and document count.
  Widget _buildExecutiveHeader({int? documentCount}) {
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

          // Foreground Header Content
          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Navigation Bar Row (Back Button, Official Badge, Refresh Button)
                  Row(
                    children: [
                      // Back Button returning to Category Directory
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

                      // Official Badge
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

                  // Selected Category Name in Header
                  Text(
                    widget.category.name,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 24,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.4,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    widget.category.id == 'links'
                        ? 'Official Government Portals & Department Dashboards'
                        : 'Official documents, circulars & government orders',
                    style: const TextStyle(
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

  /// Builds the reactive document cards list for this category.
  Widget _buildDocumentsList() {
    final docsAsync = ref.watch(categoryDocumentsProvider((
      categoryId: widget.category.id,
      subcategory: _selectedSubcategory,
    )));

    return docsAsync.when(
      loading: () => const ResourcesShimmer(itemCount: 6),
      error: (err, stack) => _buildErrorState(
        err.toString().replaceAll('Exception: ', ''),
        onRetry: () => ref.invalidate(categoryDocumentsProvider((
          categoryId: widget.category.id,
          subcategory: _selectedSubcategory,
        ))),
      ),
      data: (documents) {
        if (documents.isEmpty) {
          return _buildEmptyState(
            title: 'No Documents Found',
            message: _selectedSubcategory != null
                ? 'No documents found for "$_selectedSubcategory" in this category.'
                : 'No documents have been uploaded for ${widget.category.name} yet.',
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
              categoryName: widget.category.name,
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
                'Failed to load documents',
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
