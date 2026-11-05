import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:open_filex/open_filex.dart';
import 'package:printing/printing.dart';
import 'package:share_plus/share_plus.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/resource_document.dart';
import '../../services/document_download_service.dart';
import '../pdf_viewer_screen.dart';

/// Modal bottom sheet displaying comprehensive actions for a [ResourceDocument].
///
/// Actions provided:
/// - In-App PDF Viewing (`/resources/pdf-viewer`)
/// - Direct Download to device Downloads directory
/// - Open in External viewer (`OpenFilex`)
/// - Native Sharing sheet (`SharePlus`)
/// - Native Printing (`Printing.layoutPdf`)
class DocumentActionsSheet extends ConsumerStatefulWidget {
  final ResourceDocument document;
  final String? categoryName;

  const DocumentActionsSheet({
    super.key,
    required this.document,
    this.categoryName,
  });

  /// Displays the [DocumentActionsSheet] as a styled modal bottom sheet.
  static Future<void> show(
    BuildContext context, {
    required ResourceDocument document,
    String? categoryName,
  }) {
    return showModalBottomSheet<void>(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => DocumentActionsSheet(
        document: document,
        categoryName: categoryName,
      ),
    );
  }

  @override
  ConsumerState<DocumentActionsSheet> createState() =>
      _DocumentActionsSheetState();
}

class _DocumentActionsSheetState extends ConsumerState<DocumentActionsSheet> {
  bool _isDownloading = false;
  bool _isOpening = false;
  bool _isSharing = false;
  bool _isPrinting = false;

  /// Helper to extract clean filename.
  String _getFileName() {
    final uri = Uri.tryParse(widget.document.fullUrl);
    final pathSegments = uri?.pathSegments ?? [];
    if (pathSegments.isNotEmpty) {
      final lastSegment = pathSegments.last.split('?').first.trim();
      if (lastSegment.toLowerCase().endsWith('.pdf')) {
        return Uri.decodeComponent(lastSegment);
      }
    }
    final sanitized =
        widget.document.title.replaceAll(RegExp(r'[\\/:*?"<>|]'), '_').trim();
    return sanitized.toLowerCase().endsWith('.pdf') ? sanitized : '$sanitized.pdf';
  }

  /// Ensures document is cached in application storage before external dispatch.
  Future<File?> _ensureLocalFile() async {
    final fileName = _getFileName();
    final cached =
        await DocumentDownloadService.instance.getCachedDocument(fileName);
    if (cached != null) return cached;

    return await DocumentDownloadService.instance.downloadDocument(
      widget.document.fullUrl,
      fileName,
    );
  }

  /// Action: View Document in-app
  void _handleViewDocument() {
    HapticFeedback.lightImpact();
    Navigator.of(context).pop();

    try {
      context.push(
        '/resources/pdf-viewer',
        extra: {
          'title': widget.document.title,
          'url': widget.document.fullUrl,
        },
      );
    } catch (_) {
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => PdfViewerScreen(
            title: widget.document.title,
            url: widget.document.fullUrl,
          ),
        ),
      );
    }
  }

  /// Action: Download to Device Downloads directory
  Future<void> _handleDownloadToDevice() async {
    if (_isDownloading) return;
    HapticFeedback.lightImpact();

    setState(() => _isDownloading = true);

    try {
      final fileName = _getFileName();
      final savedFile = await DocumentDownloadService.instance.downloadDocument(
        widget.document.fullUrl,
        fileName,
        saveToPublicDownloads: true,
      );

      if (mounted) {
        setState(() => _isDownloading = false);
        Navigator.of(context).pop();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(LucideIcons.checkCircle2, color: Colors.white, size: 20),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Saved "$fileName" to Downloads',
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
            action: SnackBarAction(
              label: 'Open',
              textColor: const Color(0xFFB7E4C7),
              onPressed: () {
                OpenFilex.open(savedFile.path);
              },
            ),
            backgroundColor: AppColors.primary,
            behavior: SnackBarBehavior.floating,
            duration: const Duration(seconds: 4),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isDownloading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Download failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Action: Open in External App
  Future<void> _handleOpenExternally() async {
    if (_isOpening) return;
    HapticFeedback.lightImpact();

    setState(() => _isOpening = true);

    try {
      final file = await _ensureLocalFile();
      if (file != null) {
        final result = await OpenFilex.open(file.path);
        if (result.type != ResultType.done && mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Could not launch external app: ${result.message}'),
              backgroundColor: AppColors.destructive,
            ),
          );
        } else if (mounted) {
          Navigator.of(context).pop();
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to open document: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isOpening = false);
      }
    }
  }

  /// Action: Share document
  Future<void> _handleShare() async {
    if (_isSharing) return;
    HapticFeedback.lightImpact();

    setState(() => _isSharing = true);

    try {
      final file = await _ensureLocalFile();
      if (file != null) {
        if (mounted) Navigator.of(context).pop();
        await SharePlus.instance.share(
          ShareParams(
            files: [XFile(file.path)],
            text: widget.document.title,
            subject: widget.document.title,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Share failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isSharing = false);
      }
    }
  }

  /// Action: Print document
  Future<void> _handlePrint() async {
    if (_isPrinting) return;
    HapticFeedback.lightImpact();

    setState(() => _isPrinting = true);

    try {
      final file = await _ensureLocalFile();
      if (file != null) {
        final bytes = await file.readAsBytes();
        if (mounted) Navigator.of(context).pop();
        await Printing.layoutPdf(
          name: widget.document.title,
          onLayout: (format) async => bytes,
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Print failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isPrinting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      padding: EdgeInsets.only(
        left: 20,
        right: 20,
        top: 12,
        bottom: 24 + MediaQuery.of(context).padding.bottom,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Centered drag handle
          Center(
            child: Container(
              width: 44,
              height: 4,
              margin: const EdgeInsets.only(bottom: 16),
              decoration: BoxDecoration(
                color: AppColors.textMuted.withValues(alpha: 0.28),
                borderRadius: BorderRadius.circular(10),
              ),
            ),
          ),

          // Header: Document title & dismiss button
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Text(
                  widget.document.title,
                  style: const TextStyle(
                    fontSize: 17,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.2,
                    height: 1.3,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  Navigator.of(context).pop();
                },
                child: Container(
                  width: 32,
                  height: 32,
                  decoration: BoxDecoration(
                    color: AppColors.muted,
                    shape: BoxShape.circle,
                    border: Border.all(color: AppColors.borderLight),
                  ),
                  child: const Icon(
                    LucideIcons.x,
                    size: 16,
                    color: AppColors.textSecondary,
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 12),

          // Badges row: PDF indicator, category, subcategory, publication year
          Wrap(
            spacing: 8,
            runSpacing: 6,
            children: [
              // PDF badge
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFFDE8E8),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: const Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      LucideIcons.fileText,
                      size: 12,
                      color: Color(0xFFE02424),
                    ),
                    SizedBox(width: 4),
                    Text(
                      'PDF Document',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFFE02424),
                      ),
                    ),
                  ],
                ),
              ),

              // Year badge
              if (widget.document.year.isNotEmpty)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.secondary,
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(
                        LucideIcons.calendar,
                        size: 12,
                        color: AppColors.primary,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        widget.document.year,
                        style: const TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                          color: AppColors.primary,
                        ),
                      ),
                    ],
                  ),
                ),

              // Category badge
              if (widget.categoryName != null && widget.categoryName!.isNotEmpty)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.muted,
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: AppColors.borderLight),
                  ),
                  child: Text(
                    widget.categoryName!,
                    style: const TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ),

              // Subcategory badge
              if (widget.document.subcategory != null &&
                  widget.document.subcategory!.isNotEmpty)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.muted,
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: AppColors.borderLight),
                  ),
                  child: Text(
                    widget.document.subcategory!,
                    style: const TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ),
            ],
          ),

          const SizedBox(height: 18),
          const Divider(height: 1, color: AppColors.borderLight),
          const SizedBox(height: 10),

          // Actions list
          _ActionTile(
            icon: LucideIcons.bookOpen,
            title: 'View Document',
            subtitle: 'Open and read inside the application',
            onTap: _handleViewDocument,
          ),
          _ActionTile(
            icon: LucideIcons.download,
            title: 'Download to Device',
            subtitle: 'Save a copy to your Downloads folder',
            isLoading: _isDownloading,
            onTap: _handleDownloadToDevice,
          ),
          _ActionTile(
            icon: LucideIcons.externalLink,
            title: 'Open in External App',
            subtitle: 'Launch in your preferred external PDF viewer',
            isLoading: _isOpening,
            onTap: _handleOpenExternally,
          ),
          _ActionTile(
            icon: LucideIcons.share2,
            title: 'Share Document',
            subtitle: 'Send document via WhatsApp, Mail, etc.',
            isLoading: _isSharing,
            onTap: _handleShare,
          ),
          _ActionTile(
            icon: LucideIcons.printer,
            title: 'Print Document',
            subtitle: 'Print or export to PDF via system printer',
            isLoading: _isPrinting,
            onTap: _handlePrint,
          ),
        ],
      ),
    );
  }
}

class _ActionTile extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;
  final bool isLoading;

  const _ActionTile({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
    this.isLoading = false,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: BouncyButton(
        onTap: isLoading ? null : onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
          decoration: BoxDecoration(
            color: Colors.transparent,
            borderRadius: BorderRadius.circular(14),
          ),
          child: Row(
            children: [
              Container(
                width: 42,
                height: 42,
                decoration: BoxDecoration(
                  color: AppColors.secondary,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  icon,
                  size: 20,
                  color: AppColors.primary,
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w600,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
              if (isLoading)
                const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(AppColors.primary),
                  ),
                )
              else
                Icon(
                  LucideIcons.chevronRight,
                  size: 18,
                  color: AppColors.textMuted.withValues(alpha: 0.5),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
