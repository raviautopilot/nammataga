import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/core/constants/api_endpoints.dart';
import 'package:nammataga_mobile/features/resources/data/resources_repository.dart';
import 'package:nammataga_mobile/features/resources/domain/external_link.dart';
import 'package:nammataga_mobile/features/resources/domain/resource_category.dart';
import 'package:nammataga_mobile/features/resources/domain/resource_document.dart';
import 'package:nammataga_mobile/features/resources/providers/resources_provider.dart';

void main() {
  group('ResourceCategory Domain Tests', () {
    test('fromJson and toJson with subcategories', () {
      final json = {
        'id': 'scheme-gos',
        'name': 'Scheme G.Os',
        'subcategories': ['Central', 'State'],
      };

      final cat = ResourceCategory.fromJson(json);
      expect(cat.id, 'scheme-gos');
      expect(cat.name, 'Scheme G.Os');
      expect(cat.subcategories, ['Central', 'State']);

      final serialized = cat.toJson();
      expect(serialized['id'], 'scheme-gos');
      expect(serialized['name'], 'Scheme G.Os');
      expect(serialized['subcategories'], ['Central', 'State']);
    });

    test('fromJson with missing/null subcategories', () {
      final json = {
        'id': 'establishment',
        'name': 'Establishment',
      };

      final cat = ResourceCategory.fromJson(json);
      expect(cat.id, 'establishment');
      expect(cat.name, 'Establishment');
      expect(cat.subcategories, isNull);

      final serialized = cat.toJson();
      expect(serialized.containsKey('subcategories'), isFalse);
    });

    test('equality and hashCode', () {
      const cat1 = ResourceCategory(
        id: '1',
        name: 'Cat 1',
        subcategories: ['A', 'B'],
      );
      const cat2 = ResourceCategory(
        id: '1',
        name: 'Cat 1',
        subcategories: ['A', 'B'],
      );
      const cat3 = ResourceCategory(
        id: '1',
        name: 'Cat 1',
        subcategories: ['A', 'C'],
      );

      expect(cat1, equals(cat2));
      expect(cat1.hashCode, equals(cat2.hashCode));
      expect(cat1 == cat3, isFalse);
    });

    test('copyWith works correctly', () {
      const cat = ResourceCategory(id: '1', name: 'Original');
      final updated = cat.copyWith(
        name: 'Updated',
        subcategories: ['Central', 'State'],
      );

      expect(updated.id, '1');
      expect(updated.name, 'Updated');
      expect(updated.subcategories, ['Central', 'State']);
    });
  });

  group('ResourceDocument Domain Tests', () {
    test('fromJson and toJson', () {
      final json = {
        'title': 'Agroforestry Rules',
        'year': '2026',
        'subcategory': 'Central',
        'url': '/docs/Scheme G.Os/agroforestry.pdf',
      };

      final doc = ResourceDocument.fromJson(json);
      expect(doc.title, 'Agroforestry Rules');
      expect(doc.year, '2026');
      expect(doc.subcategory, 'Central');
      expect(doc.url, '/docs/Scheme G.Os/agroforestry.pdf');

      final serialized = doc.toJson();
      expect(serialized, json);
    });

    test('fullUrl resolves properly', () {
      const docWithUrl = ResourceDocument(
        title: 'Doc',
        year: '2026',
        url: '/docs/Establishment/test.pdf',
      );
      expect(docWithUrl.fullUrl, contains('/api/docs/Establishment/test.pdf'));

      const docWithoutUrl = ResourceDocument(
        title: 'Doc',
        year: '2026',
      );
      expect(docWithoutUrl.fullUrl, isEmpty);

      const docWithHttp = ResourceDocument(
        title: 'Doc',
        year: '2026',
        url: 'https://example.com/doc.pdf',
      );
      expect(docWithHttp.fullUrl, 'https://example.com/doc.pdf');
    });

    test('isPdf detector', () {
      const pdfDoc = ResourceDocument(
        title: 'PDF Doc',
        year: '2026',
        url: '/docs/sample.PDF',
      );
      expect(pdfDoc.isPdf, isTrue);

      const pdfWithQuery = ResourceDocument(
        title: 'PDF Doc',
        year: '2026',
        url: '/docs/sample.pdf?token=123',
      );
      expect(pdfWithQuery.isPdf, isTrue);

      const docxDoc = ResourceDocument(
        title: 'Word Doc',
        year: '2026',
        url: '/docs/sample.docx',
      );
      expect(docxDoc.isPdf, isFalse);

      const noUrlDoc = ResourceDocument(title: 'Doc', year: '2026');
      expect(noUrlDoc.isPdf, isFalse);
    });

    test('sortDocuments follows website sorting logic', () {
      final docs = [
        const ResourceDocument(title: 'Zoos Circular', year: '2025'),
        const ResourceDocument(title: 'Beta Policy', year: '2026'),
        const ResourceDocument(title: 'Alpha Policy', year: '2026'),
        const ResourceDocument(title: 'Old Scheme', year: '2024-25'),
        const ResourceDocument(title: 'Agro Forestry', year: '2025'),
        const ResourceDocument(title: 'Undated Doc', year: ''),
      ];

      final currentYear = DateTime.now().year.toString();
      final sorted = sortDocuments(docs);

      // Undated doc defaults to current year (2026), so year 2026 docs come first:
      // In 2026: 'Alpha Policy', 'Beta Policy', 'Undated Doc' (alphabetical)
      expect(sorted[0].title, 'Alpha Policy');
      expect(sorted[1].title, 'Beta Policy');
      if (currentYear == '2026') {
        expect(sorted[2].title, 'Undated Doc');
      }

      // Then 2025 docs: 'Agro Forestry', 'Zoos Circular'
      final year2025Docs = sorted.where((d) => d.year == '2025').toList();
      expect(year2025Docs[0].title, 'Agro Forestry');
      expect(year2025Docs[1].title, 'Zoos Circular');

      // Then 2024-25 doc:
      expect(sorted.last.title, 'Old Scheme');
      expect(sorted.last.year, '2024-25');
    });
  });

  group('ExternalLink Domain Tests', () {
    test('fromJson and toJson', () {
      final json = {
        'title': 'TN Govt Portal',
        'url': 'https://tn.gov.in',
      };

      final link = ExternalLink.fromJson(json);
      expect(link.title, 'TN Govt Portal');
      expect(link.url, 'https://tn.gov.in');

      expect(link.toJson(), json);
    });

    test('sortExternalLinks sorts alphabetically case-insensitive', () {
      final links = [
        const ExternalLink(title: 'TNPSC Portal', url: 'https://tnpsc.gov.in'),
        const ExternalLink(title: 'agri portal', url: 'https://agri.tn.gov.in'),
        const ExternalLink(title: 'Cyber Security', url: 'https://cert.in'),
      ];

      final sorted = sortExternalLinks(links);
      expect(sorted[0].title, 'agri portal');
      expect(sorted[1].title, 'Cyber Security');
      expect(sorted[2].title, 'TNPSC Portal');
    });
  });

  group('ResourcesRepository Tests with mock Dio', () {
    late Dio dio;
    late ResourcesRepository repository;

    setUp(() {
      dio = Dio(BaseOptions(baseUrl: ApiEndpoints.baseUrl));
      repository = ResourcesRepository(dio: dio);
    });

    test('getCategories injects subcategories for scheme-gos', () async {
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            if (options.path.endsWith('/resources')) {
              return handler.resolve(
                Response(
                  requestOptions: options,
                  statusCode: 200,
                  data: [
                    {'id': 'establishment', 'name': 'Establishment'},
                    {'id': 'scheme-gos', 'name': 'Scheme G.Os'},
                  ],
                ),
              );
            }
            return handler.next(options);
          },
        ),
      );

      final categories = await repository.getCategories();
      expect(categories.length, 2);

      final establishment = categories.firstWhere((c) => c.id == 'establishment');
      expect(establishment.subcategories, isNull);

      final schemeGos = categories.firstWhere((c) => c.id == 'scheme-gos');
      expect(schemeGos.subcategories, ['Central', 'State']);
    });

    test('getDocuments fetches and sorts documents', () async {
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            if (options.path.contains('/resources/scheme-gos')) {
              expect(options.path, contains('subcategory=Central'));
              return handler.resolve(
                Response(
                  requestOptions: options,
                  statusCode: 200,
                  data: [
                    {
                      'title': 'Z Scheme',
                      'year': '2026',
                      'subcategory': 'Central',
                      'url': '/docs/z.pdf',
                    },
                    {
                      'title': 'A Scheme',
                      'year': '2026',
                      'subcategory': 'Central',
                      'url': '/docs/a.pdf',
                    },
                  ],
                ),
              );
            }
            return handler.next(options);
          },
        ),
      );

      final docs = await repository.getDocuments(
        'scheme-gos',
        subcategory: 'Central',
      );
      expect(docs.length, 2);
      expect(docs[0].title, 'A Scheme');
      expect(docs[1].title, 'Z Scheme');
    });

    test('getExternalLinks fetches and sorts links', () async {
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            if (options.path.endsWith('/resources/external-links')) {
              return handler.resolve(
                Response(
                  requestOptions: options,
                  statusCode: 200,
                  data: [
                    {'title': 'TNAU', 'url': 'https://tnau.ac.in'},
                    {'title': 'Agri Dept', 'url': 'https://agritech.tnau.ac.in'},
                  ],
                ),
              );
            }
            return handler.next(options);
          },
        ),
      );

      final links = await repository.getExternalLinks();
      expect(links.length, 2);
      expect(links[0].title, 'Agri Dept');
      expect(links[1].title, 'TNAU');
    });
  });

  group('Resources Provider Riverpod Tests', () {
    test('selectedCategoryProvider auto-selects first category once loaded', () async {
      final mockCategories = [
        const ResourceCategory(id: 'cat-1', name: 'Category 1'),
        const ResourceCategory(id: 'cat-2', name: 'Category 2'),
      ];

      final container = ProviderContainer(
        overrides: [
          resourceCategoriesProvider.overrideWith(
            (ref) => Future.value(mockCategories),
          ),
        ],
      );
      addTearDown(container.dispose);

      // Trigger categories loading
      await container.read(resourceCategoriesProvider.future);

      // Check selected category has defaulted to 'cat-1'
      expect(container.read(selectedCategoryProvider), 'cat-1');

      // User selects cat-2
      container.read(selectedCategoryProvider.notifier).setCategory('cat-2');
      expect(container.read(selectedCategoryProvider), 'cat-2');

      // Subcategory testing
      container.read(selectedSubcategoryProvider.notifier).setSubcategory('Central');
      expect(container.read(selectedSubcategoryProvider), 'Central');

      // Changing category resets subcategory
      container.read(selectedCategoryProvider.notifier).setCategory('cat-1');
      expect(container.read(selectedSubcategoryProvider), isNull);
    });

    test('categoryDocumentsProvider family fetches correctly', () async {
      final mockDocs = [
        const ResourceDocument(
          title: 'Document 1',
          year: '2026',
          url: '/docs/doc1.pdf',
        ),
      ];

      final container = ProviderContainer(
        overrides: [
          categoryDocumentsProvider((
            categoryId: 'cat-1',
            subcategory: 'Central',
          )).overrideWith((ref) => Future.value(mockDocs)),
        ],
      );
      addTearDown(container.dispose);

      final docs = await container.read(
        categoryDocumentsProvider((
          categoryId: 'cat-1',
          subcategory: 'Central',
        )).future,
      );

      expect(docs.length, 1);
      expect(docs.first.title, 'Document 1');
    });

    test('externalLinksProvider fetches correctly', () async {
      final mockLinks = [
        const ExternalLink(title: 'Portal', url: 'https://example.com'),
      ];

      final container = ProviderContainer(
        overrides: [
          externalLinksProvider.overrideWith((ref) => Future.value(mockLinks)),
        ],
      );
      addTearDown(container.dispose);

      final links = await container.read(externalLinksProvider.future);
      expect(links.length, 1);
      expect(links.first.title, 'Portal');
    });
  });
}
