import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:open_filex/open_filex.dart';
import 'package:printing/printing.dart';
import 'package:share_plus/share_plus.dart';
import 'package:syncfusion_flutter_pdfviewer/pdfviewer.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../services/document_download_service.dart';

/// High-performance, buttery-smooth in-app PDF viewing experience for NammaTAGA documents.
///
/// Optimizations:
/// - Instant streaming direct from network via [SfPdfViewer.network] (zero hanging on repeat opens)
/// - Ultra-smooth 60/120 fps pinch-to-zoom and pan (scroll heads and text selection collision disabled)
/// - Isolated page number notifiers (no frame-dropping rebuilds during gesture interactions)
/// - Direct export to device's public Downloads directory with quick "Open" action
/// - Full suite of enterprise actions: Native Share, Wireless Print, Open in External App
class PdfViewerScreen extends ConsumerStatefulWidget {
  final String title;
  final String url;
  final String? localPath;

  const PdfViewerScreen({
    super.key,
    required this.title,
    required this.url,
    this.localPath,
  });

  @override
  ConsumerState<PdfViewerScreen> createState() => _PdfViewerScreenState();
}

class _PdfViewerScreenState extends ConsumerState<PdfViewerScreen> {
  late final PdfViewerController _pdfViewerController;
  final ValueNotifier<int> _currentPageNotifier = ValueNotifier<int>(1);
  final ValueNotifier<int> _pageCountNotifier = ValueNotifier<int>(0);
  final ValueNotifier<double> _currentZoomNotifier = ValueNotifier<double>(1.0);
  final ValueNotifier<bool> _isDocumentLoadedNotifier = ValueNotifier<bool>(false);

  bool _isDownloading = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _pdfViewerController = PdfViewerController();
  }

  @override
  void dispose() {
    _pdfViewerController.dispose();
    _currentPageNotifier.dispose();
    _pageCountNotifier.dispose();
    _currentZoomNotifier.dispose();
    _isDocumentLoadedNotifier.dispose();
    super.dispose();
  }

  /// Extracts a safe filename from the document URL or title.
  String _getFileName() {
    final uri = Uri.tryParse(widget.url);
    final pathSegments = uri?.pathSegments ?? [];
    if (pathSegments.isNotEmpty) {
      final lastSegment = pathSegments.last.split('?').first.trim();
      if (lastSegment.toLowerCase().endsWith('.pdf')) {
        return Uri.decodeComponent(lastSegment);
      }
    }
    final sanitized =
        widget.title.replaceAll(RegExp(r'[\\/:*?"<>|]'), '_').trim();
    return sanitized.toLowerCase().endsWith('.pdf')
        ? sanitized
        : '$sanitized.pdf';
  }

  /// Downloads and caches a local copy on demand for Share, Print, or Open External.
  Future<File?> _fetchFileForAction(String actionName) async {
    // If localPath was passed and exists, reuse it
    if (widget.localPath != null && widget.localPath!.isNotEmpty) {
      final explicitFile = File(widget.localPath!);
      if (await explicitFile.exists() && await explicitFile.length() > 0) {
        return explicitFile;
      }
    }

    final fileName = _getFileName();

    if (!mounted) return null;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Preparing document for $actionName...'),
        duration: const Duration(seconds: 2),
        backgroundColor: AppColors.primary,
        behavior: SnackBarBehavior.floating,
      ),
    );

    try {
      return await DocumentDownloadService.instance.downloadDocument(
        widget.url,
        fileName,
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to prepare document: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
      return null;
    }
  }

  /// Downloads the document directly to the device's public Downloads directory.
  Future<void> _handleDownloadToDevice() async {
    if (_isDownloading) return;
    HapticFeedback.lightImpact();

    setState(() => _isDownloading = true);
    final fileName = _getFileName();

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              const SizedBox(
                width: 18,
                height: 18,
                child: CircularProgressIndicator(
                  strokeWidth: 2.2,
                  color: Colors.white,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  'Downloading "$fileName" to device Downloads...',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          duration: const Duration(seconds: 4),
          backgroundColor: AppColors.primary,
          behavior: SnackBarBehavior.floating,
        ),
      );
    }

    try {
      final savedFile =
          await DocumentDownloadService.instance.downloadDocument(
        widget.url,
        fileName,
        saveToPublicDownloads: true,
      );

      if (mounted) {
        setState(() => _isDownloading = false);
        ScaffoldMessenger.of(context).hideCurrentSnackBar();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(
                  LucideIcons.checkCircle2,
                  color: Colors.white,
                  size: 20,
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Saved "$fileName" to device Downloads',
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
            duration: const Duration(seconds: 5),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isDownloading = false);
        ScaffoldMessenger.of(context).hideCurrentSnackBar();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Download failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Triggers the native system share sheet.
  Future<void> _handleShare() async {
    HapticFeedback.lightImpact();
    final file = await _fetchFileForAction('sharing');
    if (file == null) return;

    try {
      await SharePlus.instance.share(
        ShareParams(
          files: [XFile(file.path)],
          text: widget.title,
          subject: widget.title,
        ),
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Share failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Sends the document to the system print dialog.
  Future<void> _handlePrint() async {
    HapticFeedback.lightImpact();
    final file = await _fetchFileForAction('printing');
    if (file == null) return;

    try {
      final bytes = await file.readAsBytes();
      await Printing.layoutPdf(
        name: widget.title,
        onLayout: (format) async => bytes,
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Print failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Opens the PDF with the device's installed PDF viewers.
  Future<void> _handleOpenExternally() async {
    HapticFeedback.lightImpact();
    final file = await _fetchFileForAction('external viewer');
    if (file == null) return;

    try {
      final result = await OpenFilex.open(file.path);
      if (result.type != ResultType.done && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Could not open external app: ${result.message}'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error launching external viewer: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Prompts a dialog allowing user to jump directly to a page number.
  void _showPageJumpDialog() {
    final pageCount = _pageCountNotifier.value;
    if (pageCount <= 1) return;

    HapticFeedback.lightImpact();
    final textController = TextEditingController();

    showDialog<void>(
      context: context,
      builder: (dialogContext) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(18),
          ),
          title: const Text(
            'Jump to Page',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Enter page number (1 – $pageCount):',
                style: const TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: textController,
                keyboardType: TextInputType.number,
                autofocus: true,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: InputDecoration(
                  filled: true,
                  fillColor: AppColors.muted,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(color: AppColors.borderLight),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(
                      color: AppColors.primary,
                      width: 1.5,
                    ),
                  ),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: const Text(
                'Cancel',
                style: TextStyle(color: AppColors.textSecondary),
              ),
            ),
            ElevatedButton(
              onPressed: () {
                final target = int.tryParse(textController.text.trim());
                if (target != null && target >= 1 && target <= pageCount) {
                  _pdfViewerController.jumpToPage(target);
                  Navigator.of(dialogContext).pop();
                }
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
              child: const Text('Go'),
            ),
          ],
        );
      },
    );
  }

  /// Opens the PDF URL directly in the device's external browser.
  Future<void> _openInBrowser() async {
    final uri = Uri.tryParse(widget.url);
    if (uri != null) {
      try {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Could not launch browser: $e'),
              backgroundColor: AppColors.destructive,
            ),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        titleSpacing: 0,
        backgroundColor: AppColors.surface,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: AppColors.primary),
          onPressed: () => Navigator.of(context).pop(),
          tooltip: 'Back',
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              widget.title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            ValueListenableBuilder<int>(
              valueListenable: _pageCountNotifier,
              builder: (context, pageCount, _) {
                if (pageCount == 0) return const SizedBox.shrink();
                return ValueListenableBuilder<int>(
                  valueListenable: _currentPageNotifier,
                  builder: (context, currentPage, _) {
                    return Text(
                      'Page $currentPage of $pageCount',
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                    );
                  },
                );
              },
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: _isDownloading
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: AppColors.primary,
                    ),
                  )
                : const Icon(LucideIcons.download, size: 20),
            tooltip: 'Download to device',
            onPressed: _isDownloading ? null : _handleDownloadToDevice,
          ),
          IconButton(
            icon: const Icon(LucideIcons.share2, size: 20),
            tooltip: 'Share',
            onPressed: _handleShare,
          ),
          PopupMenuButton<String>(
            icon: const Icon(LucideIcons.moreVertical, size: 20),
            tooltip: 'More actions',
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
            onSelected: (value) {
              switch (value) {
                case 'print':
                  _handlePrint();
                  break;
                case 'external':
                  _handleOpenExternally();
                  break;
                case 'browser':
                  _openInBrowser();
                  break;
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'print',
                child: Row(
                  children: [
                    Icon(
                      LucideIcons.printer,
                      size: 18,
                      color: AppColors.textPrimary,
                    ),
                    SizedBox(width: 12),
                    Text('Print Document', style: TextStyle(fontSize: 14)),
                  ],
                ),
              ),
              const PopupMenuItem(
                value: 'external',
                child: Row(
                  children: [
                    Icon(
                      LucideIcons.externalLink,
                      size: 18,
                      color: AppColors.textPrimary,
                    ),
                    SizedBox(width: 12),
                    Text(
                      'Open in External App',
                      style: TextStyle(fontSize: 14),
                    ),
                  ],
                ),
              ),
              const PopupMenuItem(
                value: 'browser',
                child: Row(
                  children: [
                    Icon(
                      LucideIcons.globe,
                      size: 18,
                      color: AppColors.textPrimary,
                    ),
                    SizedBox(width: 12),
                    Text('Open in Browser', style: TextStyle(fontSize: 14)),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
      body: _buildBody(),
      bottomNavigationBar: ValueListenableBuilder<bool>(
        valueListenable: _isDocumentLoadedNotifier,
        builder: (context, isLoaded, _) {
          if (!isLoaded || _errorMessage != null) {
            return const SizedBox.shrink();
          }
          return _buildBottomBar();
        },
      ),
    );
  }

  Widget _buildBody() {
    if (_errorMessage != null) {
      return _buildErrorView();
    }

    // Direct, buttery-smooth PDF rendering
    // Disables gesture-fighting text selection and scroll heads for seamless pinch-to-zoom
    final explicitFile = widget.localPath != null && widget.localPath!.isNotEmpty
        ? File(widget.localPath!)
        : null;

    final viewerWidget = (explicitFile != null && explicitFile.existsSync())
        ? SfPdfViewer.file(
            explicitFile,
            controller: _pdfViewerController,
            canShowScrollHead: false,
            canShowScrollStatus: false,
            canShowPaginationDialog: false,
            canShowPageLoadingIndicator: false,
            enableDoubleTapZooming: true,
            enableTextSelection: false,
            pageSpacing: 4.0,
            maxZoomLevel: 3.0,
            interactionMode: PdfInteractionMode.pan,
            onZoomLevelChanged: (details) {
              _currentZoomNotifier.value = details.newZoomLevel;
            },
            onDocumentLoaded: (details) {
              _pageCountNotifier.value = details.document.pages.count;
              _currentPageNotifier.value = _pdfViewerController.pageNumber;
              _currentZoomNotifier.value = _pdfViewerController.zoomLevel;
              _isDocumentLoadedNotifier.value = true;
            },
            onPageChanged: (details) {
              _currentPageNotifier.value = details.newPageNumber;
            },
            onDocumentLoadFailed: (details) {
              if (mounted) {
                setState(() {
                  _errorMessage = details.description;
                });
              }
            },
          )
        : SfPdfViewer.network(
            widget.url,
            controller: _pdfViewerController,
            canShowScrollHead: false,
            canShowScrollStatus: false,
            canShowPaginationDialog: false,
            canShowPageLoadingIndicator: false,
            enableDoubleTapZooming: true,
            enableTextSelection: false,
            pageSpacing: 4.0,
            maxZoomLevel: 3.0,
            interactionMode: PdfInteractionMode.pan,
            onZoomLevelChanged: (details) {
              _currentZoomNotifier.value = details.newZoomLevel;
            },
            onDocumentLoaded: (details) {
              _pageCountNotifier.value = details.document.pages.count;
              _currentPageNotifier.value = _pdfViewerController.pageNumber;
              _currentZoomNotifier.value = _pdfViewerController.zoomLevel;
              _isDocumentLoadedNotifier.value = true;
            },
            onPageChanged: (details) {
              _currentPageNotifier.value = details.newPageNumber;
            },
            onDocumentLoadFailed: (details) {
              if (mounted) {
                setState(() {
                  _errorMessage = details.description;
                });
              }
            },
          );

    return viewerWidget;
  }

  /// Error view with Retry and Open in Browser fallbacks.
  Widget _buildErrorView() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: AppColors.destructive.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                LucideIcons.fileX,
                color: AppColors.destructive,
                size: 40,
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Unable to Open Document',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 10),
            Text(
              _errorMessage ??
                  'The document could not be loaded. Please verify your connection or open it in your browser.',
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 28),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                BouncyButton(
                  onTap: () {
                    HapticFeedback.lightImpact();
                    setState(() {
                      _errorMessage = null;
                      _isDocumentLoadedNotifier.value = false;
                      _currentZoomNotifier.value = 1.0;
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 20,
                      vertical: 12,
                    ),
                    decoration: BoxDecoration(
                      color: AppColors.primary,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          LucideIcons.rotateCcw,
                          size: 16,
                          color: Colors.white,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Retry',
                          style: TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.w600,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 14),
                BouncyButton(
                  onTap: () {
                    HapticFeedback.lightImpact();
                    _openInBrowser();
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 18,
                      vertical: 12,
                    ),
                    decoration: BoxDecoration(
                      color: AppColors.muted,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: AppColors.borderLight),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          LucideIcons.globe,
                          size: 16,
                          color: AppColors.textPrimary,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Open in Browser',
                          style: TextStyle(
                            color: AppColors.textPrimary,
                            fontWeight: FontWeight.w600,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  /// Zoom in by discrete responsive step.
  void _handleZoomIn() {
    HapticFeedback.lightImpact();
    final current = _pdfViewerController.zoomLevel;
    double target;
    if (current < 1.45) {
      target = 1.5;
    } else if (current < 1.95) {
      target = 2.0;
    } else if (current < 2.45) {
      target = 2.5;
    } else {
      target = 3.0;
    }
    _pdfViewerController.zoomLevel = target;
    _currentZoomNotifier.value = target;
  }

  /// Zoom out by discrete responsive step.
  void _handleZoomOut() {
    HapticFeedback.lightImpact();
    final current = _pdfViewerController.zoomLevel;
    double target;
    if (current > 2.55) {
      target = 2.5;
    } else if (current > 2.05) {
      target = 2.0;
    } else if (current > 1.55) {
      target = 1.5;
    } else {
      target = 1.0;
    }
    _pdfViewerController.zoomLevel = target;
    _currentZoomNotifier.value = target;
  }

  /// Resets zoom level to 100%.
  void _resetZoom() {
    HapticFeedback.lightImpact();
    _pdfViewerController.zoomLevel = 1.0;
    _currentZoomNotifier.value = 1.0;
  }

  /// Bottom navigation and zoom toolbar.
  Widget _buildBottomBar() {
    return Container(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 8,
        bottom: 8 + MediaQuery.of(context).padding.bottom,
      ),
      decoration: BoxDecoration(
        color: AppColors.surface,
        border: const Border(
          top: BorderSide(color: AppColors.borderLight),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Previous page button
          ValueListenableBuilder<int>(
            valueListenable: _currentPageNotifier,
            builder: (context, currentPage, _) {
              return IconButton(
                icon: const Icon(LucideIcons.chevronLeft),
                color: AppColors.primary,
                disabledColor: AppColors.textMuted.withValues(alpha: 0.3),
                tooltip: 'Previous page',
                onPressed: currentPage > 1
                    ? () {
                        HapticFeedback.lightImpact();
                        _pdfViewerController.previousPage();
                      }
                    : null,
              );
            },
          ),

          // Tappable Page indicator
          GestureDetector(
            onTap: _showPageJumpDialog,
            child: Container(
              padding: const EdgeInsets.symmetric(
                horizontal: 14,
                vertical: 6,
              ),
              decoration: BoxDecoration(
                color: AppColors.muted,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: AppColors.borderLight),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(
                    LucideIcons.bookOpen,
                    size: 14,
                    color: AppColors.primary,
                  ),
                  const SizedBox(width: 6),
                  ValueListenableBuilder<int>(
                    valueListenable: _pageCountNotifier,
                    builder: (context, pageCount, _) {
                      return ValueListenableBuilder<int>(
                        valueListenable: _currentPageNotifier,
                        builder: (context, currentPage, _) {
                          return Text(
                            pageCount > 0
                                ? 'Page $currentPage of $pageCount'
                                : 'Loading...',
                            style: const TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textPrimary,
                            ),
                          );
                        },
                      );
                    },
                  ),
                ],
              ),
            ),
          ),

          // Next page button
          ValueListenableBuilder<int>(
            valueListenable: _pageCountNotifier,
            builder: (context, pageCount, _) {
              return ValueListenableBuilder<int>(
                valueListenable: _currentPageNotifier,
                builder: (context, currentPage, _) {
                  return IconButton(
                    icon: const Icon(LucideIcons.chevronRight),
                    color: AppColors.primary,
                    disabledColor: AppColors.textMuted.withValues(alpha: 0.3),
                    tooltip: 'Next page',
                    onPressed: currentPage < pageCount
                        ? () {
                            HapticFeedback.lightImpact();
                            _pdfViewerController.nextPage();
                          }
                        : null,
                  );
                },
              );
            },
          ),

          // Zoom controls divider
          Container(
            height: 24,
            width: 1,
            color: AppColors.borderLight,
          ),

          // Zoom Out
          ValueListenableBuilder<double>(
            valueListenable: _currentZoomNotifier,
            builder: (context, zoom, _) {
              final canZoomOut = zoom > 1.05;
              return IconButton(
                icon: const Icon(LucideIcons.zoomOut, size: 20),
                color: AppColors.textPrimary,
                disabledColor: AppColors.textMuted.withValues(alpha: 0.3),
                tooltip: 'Zoom out',
                onPressed: canZoomOut ? _handleZoomOut : null,
              );
            },
          ),

          // Zoom percentage indicator
          ValueListenableBuilder<double>(
            valueListenable: _currentZoomNotifier,
            builder: (context, zoom, _) {
              final percent = (zoom * 100).round();
              final isZoomed = zoom > 1.05;
              return GestureDetector(
                onTap: isZoomed ? _resetZoom : null,
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: isZoomed
                        ? AppColors.primaryLight.withValues(alpha: 0.15)
                        : AppColors.muted,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    '$percent%',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: isZoomed
                          ? AppColors.primary
                          : AppColors.textSecondary,
                    ),
                  ),
                ),
              );
            },
          ),

          // Zoom In
          ValueListenableBuilder<double>(
            valueListenable: _currentZoomNotifier,
            builder: (context, zoom, _) {
              final canZoomIn = zoom < 2.95;
              return IconButton(
                icon: const Icon(LucideIcons.zoomIn, size: 20),
                color: AppColors.textPrimary,
                disabledColor: AppColors.textMuted.withValues(alpha: 0.3),
                tooltip: 'Zoom in',
                onPressed: canZoomIn ? _handleZoomIn : null,
              );
            },
          ),
        ],
      ),
    );
  }
}
