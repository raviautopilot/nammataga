import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/grievance_repository.dart';
import '../domain/grievance_model.dart';
import '../domain/grievance_priority_model.dart';

/// Provider for GrievanceRepository singleton.
final grievanceRepositoryProvider = Provider<GrievanceRepository>((ref) {
  return GrievanceRepository();
});

/// FutureProvider that fetches the list of available grievance categories.
final grievanceCategoriesProvider = FutureProvider<List<String>>((ref) async {
  final repo = ref.watch(grievanceRepositoryProvider);
  return repo.getCategories();
});

/// FutureProvider that fetches the list of grievance priorities.
final grievancePrioritiesProvider =
    FutureProvider<List<GrievancePriorityModel>>((ref) async {
  final repo = ref.watch(grievanceRepositoryProvider);
  return repo.getPriorities();
});

/// FutureProvider that fetches existing submitted grievances for the authenticated member.
final memberGrievancesProvider =
    FutureProvider.autoDispose<List<GrievanceModel>>((ref) async {
  final repo = ref.watch(grievanceRepositoryProvider);
  return repo.getMemberGrievances();
});

/// Data container holding grievance form field values.
/// Preserved across failed submission attempts so user input is never lost.
class GrievanceFormData {
  final String subject;
  final String category;
  final String priority;
  final String description;
  final String? contactPhone;
  final String preferredResponse;

  const GrievanceFormData({
    this.subject = '',
    this.category = '',
    this.priority = '',
    this.description = '',
    this.contactPhone,
    this.preferredResponse = 'Email',
  });

  GrievanceFormData copyWith({
    String? subject,
    String? category,
    String? priority,
    String? description,
    String? contactPhone,
    String? preferredResponse,
  }) {
    return GrievanceFormData(
      subject: subject ?? this.subject,
      category: category ?? this.category,
      priority: priority ?? this.priority,
      description: description ?? this.description,
      contactPhone: contactPhone ?? this.contactPhone,
      preferredResponse: preferredResponse ?? this.preferredResponse,
    );
  }

  Map<String, dynamic> toMap() => {
        'subject': subject,
        'category': category,
        'priority': priority,
        'description': description,
        'contactPhone': contactPhone,
        'preferredResponse': preferredResponse,
      };

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is GrievanceFormData &&
          runtimeType == other.runtimeType &&
          subject == other.subject &&
          category == other.category &&
          priority == other.priority &&
          description == other.description &&
          contactPhone == other.contactPhone &&
          preferredResponse == other.preferredResponse;

  @override
  int get hashCode => Object.hash(
        subject,
        category,
        priority,
        description,
        contactPhone,
        preferredResponse,
      );

  @override
  String toString() =>
      'GrievanceFormData(subject: $subject, category: $category, priority: $priority)';
}

/// State representing grievance form submission status, errors, and form cache.
class GrievanceSubmissionState {
  final bool isSubmitting;
  final String? errorMessage;
  final GrievanceModel? submittedGrievance;
  final GrievanceFormData formData;

  const GrievanceSubmissionState({
    this.isSubmitting = false,
    this.errorMessage,
    this.submittedGrievance,
    this.formData = const GrievanceFormData(),
  });

  /// True when submission has succeeded and created grievance model is present.
  bool get isSuccess => submittedGrievance != null;

  /// True when there is an active error message.
  bool get hasError => errorMessage != null && errorMessage!.trim().isNotEmpty;

  // Convenience field accessors delegating to formData
  String get subject => formData.subject;
  String get category => formData.category;
  String get priority => formData.priority;
  String get description => formData.description;
  String? get contactPhone => formData.contactPhone;
  String get preferredResponse => formData.preferredResponse;

  GrievanceSubmissionState copyWith({
    bool? isSubmitting,
    String? errorMessage,
    bool clearError = false,
    GrievanceModel? submittedGrievance,
    bool clearSubmittedGrievance = false,
    GrievanceFormData? formData,
  }) {
    return GrievanceSubmissionState(
      isSubmitting: isSubmitting ?? this.isSubmitting,
      errorMessage: clearError ? null : (errorMessage ?? this.errorMessage),
      submittedGrievance: clearSubmittedGrievance
          ? null
          : (submittedGrievance ?? this.submittedGrievance),
      formData: formData ?? this.formData,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is GrievanceSubmissionState &&
          runtimeType == other.runtimeType &&
          isSubmitting == other.isSubmitting &&
          errorMessage == other.errorMessage &&
          submittedGrievance == other.submittedGrievance &&
          formData == other.formData;

  @override
  int get hashCode => Object.hash(
        isSubmitting,
        errorMessage,
        submittedGrievance,
        formData,
      );

  @override
  String toString() =>
      'GrievanceSubmissionState(isSubmitting: $isSubmitting, isSuccess: $isSuccess, error: $errorMessage)';
}

/// Notifier managing grievance submission lifecycle, duplicate prevention, and form state preservation.
class GrievanceSubmissionNotifier extends Notifier<GrievanceSubmissionState> {
  @override
  GrievanceSubmissionState build() {
    return const GrievanceSubmissionState();
  }

  /// Updates current form values in state as the user types or interacts with fields.
  void updateFormData({
    String? subject,
    String? category,
    String? priority,
    String? description,
    String? contactPhone,
    String? preferredResponse,
  }) {
    state = state.copyWith(
      formData: state.formData.copyWith(
        subject: subject,
        category: category,
        priority: priority,
        description: description,
        contactPhone: contactPhone,
        preferredResponse: preferredResponse,
      ),
    );
  }

  /// Submits the grievance petition to the backend.
  ///
  /// - Prevents duplicate concurrent submissions if already in-progress (`isSubmitting == true`).
  /// - Takes optional field overrides or falls back to values in `state.formData`.
  /// - Preserves entered form values in `state.formData` on failure so the user can retry without re-typing.
  /// - Resets form values and assigns `submittedGrievance` on success.
  ///
  /// Returns `true` if submission succeeded, `false` otherwise.
  Future<bool> submitGrievance({
    String? subject,
    String? category,
    String? priority,
    String? description,
    String? contactPhone,
    String? preferredResponse,
  }) async {
    // 1. Guard against duplicate rapid submissions
    if (state.isSubmitting) {
      return false;
    }

    final effectiveSubject = (subject ?? state.formData.subject).trim();
    final effectiveCategory = (category ?? state.formData.category).trim();
    final effectivePriority = (priority ?? state.formData.priority).trim();
    final effectiveDescription =
        (description ?? state.formData.description).trim();
    final effectiveContactPhone =
        (contactPhone ?? state.formData.contactPhone)?.trim();
    final effectivePreferredResponse =
        (preferredResponse ?? state.formData.preferredResponse).trim();

    // Preserve the exact submitted data in state
    final currentFormData = GrievanceFormData(
      subject: effectiveSubject,
      category: effectiveCategory,
      priority: effectivePriority,
      description: effectiveDescription,
      contactPhone: effectiveContactPhone,
      preferredResponse: effectivePreferredResponse.isNotEmpty
          ? effectivePreferredResponse
          : 'Email',
    );

    // Update state to submitting, clear previous errors and previous success
    state = state.copyWith(
      isSubmitting: true,
      clearError: true,
      clearSubmittedGrievance: true,
      formData: currentFormData,
    );

    try {
      final repo = ref.read(grievanceRepositoryProvider);
      final createdGrievance = await repo.submitGrievance(
        subject: effectiveSubject,
        category: effectiveCategory,
        priority: effectivePriority,
        description: effectiveDescription,
        contactPhone: effectiveContactPhone,
        preferredResponse: currentFormData.preferredResponse,
      );

      // On success: clear form data, preserve created model
      state = GrievanceSubmissionState(
        isSubmitting: false,
        errorMessage: null,
        submittedGrievance: createdGrievance,
        formData: const GrievanceFormData(),
      );

      // Invalidate list of grievances so fresh submission appears immediately
      ref.invalidate(memberGrievancesProvider);

      return true;
    } catch (e) {
      // On failure: preserve all entered form data so user can fix and retry
      final cleanMessage =
          e.toString().replaceFirst(RegExp(r'^Exception:\s*'), '');

      state = state.copyWith(
        isSubmitting: false,
        errorMessage: cleanMessage,
        formData: currentFormData,
      );

      return false;
    }
  }

  /// Clears active error message without resetting form data.
  void clearError() {
    state = state.copyWith(clearError: true);
  }

  /// Clears successful submission state.
  void clearSuccess() {
    state = state.copyWith(clearSubmittedGrievance: true);
  }

  /// Completely resets state to clean initial defaults.
  void reset() {
    state = const GrievanceSubmissionState();
  }
}

/// Provider managing grievance form submission state and lifecycle.
final grievanceSubmissionProvider =
    NotifierProvider<GrievanceSubmissionNotifier, GrievanceSubmissionState>(
  GrievanceSubmissionNotifier.new,
);
