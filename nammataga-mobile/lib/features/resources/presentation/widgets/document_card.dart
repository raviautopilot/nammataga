import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/resource_document.dart';
import '../pdf_viewer_screen.dart';
import 'document_actions_sheet.dart';

/// Clean, modern, easy-to-scan card representing an official document.
///
/// Features:
/// - Left: PDF icon badge with subtle emerald background
/// - Middle: Document title (bold, max 2 lines, clean ellipsis)
/// - Metadata Row: Year badge with calendar icon, Subcategory badge (if present), "Official PDF" tag
/// - Right: Action button triggering [DocumentActionsSheet]
/// - Card tap: Opens in-app PDF viewer or invokes custom [onTap]
class DocumentCard extends StatelessWidget {
  final ResourceDocument document;
  final String? categoryName;
  final VoidCallback? onTap;
  final VoidCallback? onActionsTap;

  const DocumentCard({
    super.key,
    required this.document,
    this.categoryName,
    this.onTap,
    this.onActionsTap,
  });

  void _handleCardTap(BuildContext context) {
    HapticFeedback.lightImpact();
    if (onTap != null) {
      onTap!();
      return;
    }

    // Default card tap action: Open PDF Viewer directly for fast viewing
    if (document.isPdf && document.fullUrl.isNotEmpty) {
      try {
        context.push(
          '/resources/pdf-viewer',
          extra: {
            'title': document.title,
            'url': document.fullUrl,
          },
        );
        return;
      } catch (_) {
        // Fallback for non-GoRouter test harnesses
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => PdfViewerScreen(
              title: document.title,
              url: document.fullUrl,
            ),
          ),
        );
        return;
      }
    }

    // Non-PDF or general fallback: show full actions sheet
    _handleActionsTap(context);
  }

  void _handleActionsTap(BuildContext context) {
    HapticFeedback.lightImpact();
    if (onActionsTap != null) {
      onActionsTap!();
      return;
    }

    DocumentActionsSheet.show(
      context,
      document: document,
      categoryName: categoryName,
    );
  }

  @override
  Widget build(BuildContext context) {
    final cleanYear = document.year.trim().isEmpty ? null : document.year.trim();
    final cleanSubcat =
        document.subcategory != null && document.subcategory!.trim().isNotEmpty
            ? document.subcategory!.trim()
            : null;

    return BouncyButton(
      onTap: () => _handleCardTap(context),
      scaleDownFactor: 0.98,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: const Color(0xFFE2EBE5),
            width: 1.1,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF1B4332).withValues(alpha: 0.05),
              blurRadius: 12,
              offset: const Offset(0, 3),
            ),
          ],
        ),
        padding: const EdgeInsets.all(14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Left: PDF Icon Badge
            Container(
              width: 44,
              height: 48,
              decoration: BoxDecoration(
                color: const Color(0xFFF1F8E9),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: const Color(0xFFD8F3DC),
                  width: 1.1,
                ),
              ),
              child: const Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    LucideIcons.fileText,
                    size: 20,
                    color: AppColors.primary,
                  ),
                  SizedBox(height: 2),
                  Text(
                    'PDF',
                    style: TextStyle(
                      color: AppColors.primary,
                      fontSize: 8.5,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 0.5,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),

            // Middle Column: Title & Metadata Row
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Document Title
                  Text(
                    document.title,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: AppColors.textPrimary,
                      fontSize: 14.5,
                      fontWeight: FontWeight.w700,
                      height: 1.3,
                      letterSpacing: 0.1,
                    ),
                  ),
                  const SizedBox(height: 8),

                  // Metadata Badges Row
                  Wrap(
                    spacing: 6,
                    runSpacing: 4,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      // Year Badge
                      if (cleanYear != null)
                        _buildBadge(
                          icon: LucideIcons.calendar,
                          label: cleanYear,
                          backgroundColor: const Color(0xFFF1F8E9),
                          borderColor: const Color(0xFFD8F3DC),
                          textColor: AppColors.primary,
                        ),

                      // Subcategory Badge (e.g. State / Central)
                      if (cleanSubcat != null)
                        _buildBadge(
                          icon: LucideIcons.tag,
                          label: cleanSubcat,
                          backgroundColor: const Color(0xFFE8F5E9),
                          borderColor: const Color(0xFFC8E6C9),
                          textColor: const Color(0xFF1B4332),
                        ),

                      // Official PDF Tag
                      _buildBadge(
                        icon: LucideIcons.shieldCheck,
                        label: 'Official PDF',
                        backgroundColor: const Color(0xFFF6FAF7),
                        borderColor: const Color(0xFFE2EBE5),
                        textColor: AppColors.textSecondary,
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),

            // Right: Actions Button
            BouncyButton(
              onTap: () => _handleActionsTap(context),
              scaleDownFactor: 0.92,
              child: Container(
                width: 36,
                height: 36,
                decoration: BoxDecoration(
                  color: const Color(0xFFF6FAF7),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: const Color(0xFFE2EBE5),
                    width: 1.0,
                  ),
                ),
                child: const Icon(
                  LucideIcons.moreVertical,
                  size: 18,
                  color: AppColors.textSecondary,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBadge({
    required IconData icon,
    required String label,
    required Color backgroundColor,
    required Color borderColor,
    required Color textColor,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: borderColor, width: 0.9),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            size: 11,
            color: textColor,
          ),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              color: textColor,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              height: 1.1,
            ),
          ),
        ],
      ),
    );
  }
}
