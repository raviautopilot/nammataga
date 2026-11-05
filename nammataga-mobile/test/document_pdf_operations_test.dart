import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/features/resources/domain/resource_document.dart';
import 'package:nammataga_mobile/features/resources/presentation/pdf_viewer_screen.dart';
import 'package:nammataga_mobile/features/resources/presentation/widgets/document_actions_sheet.dart';
import 'package:nammataga_mobile/features/resources/services/document_download_service.dart';

class MockDioAdapter implements HttpClientAdapter {
  final Future<ResponseBody> Function(RequestOptions options) onFetch;

  MockDioAdapter(this.onFetch);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) {
    return onFetch(options);
  }

  @override
  void close({bool force = false}) {}
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('DocumentDownloadService Tests', () {
    test('singleton instance returns identical instance', () {
      final s1 = DocumentDownloadService.instance;
      final s2 = DocumentDownloadService.instance;
      final s3 = DocumentDownloadService();

      expect(identical(s1, s2), isTrue);
      expect(identical(s1, s3), isTrue);
    });

    test('Riverpod provider provides DocumentDownloadService instance', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final service = container.read(documentDownloadServiceProvider);
      expect(service, isNotNull);
      expect(service, isA<DocumentDownloadService>());
    });

    test('isDownloading returns false when no downloads are active', () {
      final service = DocumentDownloadService.instance;
      expect(service.isDownloading('https://example.com/doc.pdf'), isFalse);
    });

    test('downloadDocument handles network 404 error cleanly', () async {
      final dio = Dio();
      dio.httpClientAdapter = MockDioAdapter((options) async {
        return ResponseBody.fromString(
          'Not Found',
          404,
          headers: {
            Headers.contentTypeHeader: ['text/plain'],
          },
        );
      });

      final customService = DocumentDownloadService(dio: dio);

      expect(
        () => customService.downloadDocument(
          'https://test.nammataga.in/docs/missing.pdf',
          'test_missing_${DateTime.now().millisecondsSinceEpoch}.pdf',
        ),
        throwsA(isA<DocumentDownloadException>()),
      );
    });

    test('downloadDocument prevents duplicate concurrent downloads of identical URL', () async {
      final dio = Dio();
      // Simulate slow response
      dio.httpClientAdapter = MockDioAdapter((options) async {
        await Future.delayed(const Duration(milliseconds: 200));
        return ResponseBody.fromString(
          '%PDF-1.4 test',
          200,
          headers: {
            Headers.contentTypeHeader: ['application/pdf'],
          },
        );
      });

      final customService = DocumentDownloadService(dio: dio);
      const testUrl = 'https://test.nammataga.in/docs/duplicate_test.pdf';
      final fileName = 'duplicate_${DateTime.now().millisecondsSinceEpoch}.pdf';

      final future1 = customService.downloadDocument(testUrl, fileName);

      expect(customService.isDownloading(testUrl), isTrue);

      // Second simultaneous attempt must throw duplicate error
      expect(
        () => customService.downloadDocument(testUrl, fileName),
        throwsA(
          predicate<DocumentDownloadException>((e) =>
              e.message.contains('already in progress')),
        ),
      );

      try {
        await future1;
      } catch (_) {}

      expect(customService.isDownloading(testUrl), isFalse);
    });
  });

  group('DocumentActionsSheet Widget Tests', () {
    testWidgets('renders document title and badges correctly', (tester) async {
      const doc = ResourceDocument(
        title: 'Fundamental Rules TN Govt',
        year: '2026',
        subcategory: 'State',
        url: '/docs/Establishment/Fundamental Rules TN Govt.pdf',
      );

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) {
                return ElevatedButton(
                  onPressed: () {
                    DocumentActionsSheet.show(
                      context,
                      document: doc,
                      categoryName: 'Establishment',
                    );
                  },
                  child: const Text('Open Sheet'),
                );
              },
            ),
          ),
        ),
      );

      // Open sheet
      await tester.tap(find.text('Open Sheet'));
      await tester.pumpAndSettle();

      // Check title
      expect(find.text('Fundamental Rules TN Govt'), findsOneWidget);

      // Check badges
      expect(find.text('PDF Document'), findsOneWidget);
      expect(find.text('2026'), findsOneWidget);
      expect(find.text('Establishment'), findsOneWidget);
      expect(find.text('State'), findsOneWidget);

      // Check actions
      expect(find.text('View Document'), findsOneWidget);
      expect(find.text('Download to Device'), findsOneWidget);
      expect(find.text('Open in External App'), findsOneWidget);
      expect(find.text('Share Document'), findsOneWidget);
      expect(find.text('Print Document'), findsOneWidget);
    });
  });

  group('PdfViewerScreen Presentation Tests', () {
    testWidgets('renders AppBar with title and action buttons', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: PdfViewerScreen(
              title: 'Leave Rules TN Govt',
              url: 'https://test.nammataga.in/docs/leave_rules.pdf',
            ),
          ),
        ),
      );

      // Verify title is rendered in AppBar
      expect(find.text('Leave Rules TN Govt'), findsOneWidget);

      // Verify quick action buttons exist
      expect(find.byIcon(LucideIcons.arrowLeft), findsOneWidget);
      expect(find.byIcon(LucideIcons.download), findsOneWidget);
      expect(find.byIcon(LucideIcons.share2), findsOneWidget);
      expect(find.byIcon(LucideIcons.moreVertical), findsOneWidget);

      // Drain Syncfusion's internal network debounce timer
      await tester.pump(const Duration(seconds: 1));
    });
  });
}
