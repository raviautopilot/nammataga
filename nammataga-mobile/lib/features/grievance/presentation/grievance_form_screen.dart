import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../providers/grievance_provider.dart';
import 'widgets/category_selector_sheet.dart';
import 'widgets/priority_selector_sheet.dart';

/// Form screen for submitting a member grievance petition to the TAGA State Executive Committee.
///
/// Implements complete validation, state preservation across failures,
/// emerald design tokens, character counters, priority styling, and interactive bottom sheets.
class GrievanceFormScreen extends ConsumerStatefulWidget {
  const GrievanceFormScreen({super.key});

  @override
  ConsumerState<GrievanceFormScreen> createState() =>
      _GrievanceFormScreenState();
}

class _GrievanceFormScreenState extends ConsumerState<GrievanceFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _scrollController = ScrollController();

  late final TextEditingController _subjectController;
  late final TextEditingController _descriptionController;
  late final TextEditingController _phoneController;

  String? _selectedCategory;
  String _selectedPriority = 'medium';
  String _selectedPreferredResponse = 'email';

  String? _categoryError;

  @override
  void initState() {
    super.initState();

    // Initialize from cached formData in grievanceSubmissionProvider
    final cached = ref.read(grievanceSubmissionProvider).formData;

    _subjectController = TextEditingController(text: cached.subject);
    _descriptionController = TextEditingController(text: cached.description);
    _phoneController = TextEditingController(text: cached.contactPhone ?? '');

    _selectedCategory = cached.category.isNotEmpty ? cached.category : null;
    _selectedPriority =
        cached.priority.isNotEmpty ? cached.priority.toLowerCase() : 'medium';

    if (cached.preferredResponse.isNotEmpty) {
      final norm = cached.preferredResponse.toLowerCase().trim();
      if (norm == 'phone') {
        _selectedPreferredResponse = 'phone';
      } else if (norm == 'letter' || norm.contains('letter')) {
        _selectedPreferredResponse = 'letter';
      } else {
        _selectedPreferredResponse = 'email';
      }
    } else {
      _selectedPreferredResponse = 'email';
    }

    _subjectController.addListener(_onFieldChanged);
    _descriptionController.addListener(_onFieldChanged);
    _phoneController.addListener(_onFieldChanged);
  }

  void _onFieldChanged() {
    // Keep local state responsive for live character counts
    setState(() {});

    // Clear active error banner on user modification
    if (ref.read(grievanceSubmissionProvider).hasError) {
      ref.read(grievanceSubmissionProvider.notifier).clearError();
    }

    _syncFormData();
  }

  void _syncFormData() {
    ref.read(grievanceSubmissionProvider.notifier).updateFormData(
          subject: _subjectController.text,
          category: _selectedCategory ?? '',
          priority: _selectedPriority,
          description: _descriptionController.text,
          contactPhone: _phoneController.text.trim().isNotEmpty
              ? _phoneController.text.trim()
              : null,
          preferredResponse: _selectedPreferredResponse,
        );
  }

  @override
  void dispose() {
    _subjectController.removeListener(_onFieldChanged);
    _descriptionController.removeListener(_onFieldChanged);
    _phoneController.removeListener(_onFieldChanged);

    _subjectController.dispose();
    _descriptionController.dispose();
    _phoneController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _pickCategory() async {
    HapticFeedback.lightImpact();
    FocusScope.of(context).unfocus();

    final result = await CategorySelectorSheet.show(
      context,
      selectedCategory: _selectedCategory,
    );

    if (result != null && mounted) {
      setState(() {
        _selectedCategory = result;
        _categoryError = null;
      });
      _syncFormData();
    }
  }

  Future<void> _pickPriority() async {
    HapticFeedback.lightImpact();
    FocusScope.of(context).unfocus();

    final result = await PrioritySelectorSheet.show(
      context,
      selectedValue: _selectedPriority,
    );

    if (result != null && mounted) {
      setState(() {
        _selectedPriority = result.value.isNotEmpty
            ? result.value.toLowerCase()
            : result.label.toLowerCase();
      });
      _syncFormData();
    }
  }

  Future<void> _submitForm() async {
    HapticFeedback.lightImpact();
    FocusScope.of(context).unfocus();

    // Validate form inputs
    final isFormValid = _formKey.currentState?.validate() ?? false;

    // Validate category selector
    bool isCategoryValid = true;
    if (_selectedCategory == null || _selectedCategory!.trim().isEmpty) {
      setState(() {
        _categoryError = 'Please select a category';
      });
      isCategoryValid = false;
    } else {
      setState(() {
        _categoryError = null;
      });
    }

    if (!isFormValid || !isCategoryValid) {
      _scrollController.animateTo(
        0,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
      return;
    }

    // Submit to grievanceSubmissionProvider
    final success =
        await ref.read(grievanceSubmissionProvider.notifier).submitGrievance(
              subject: _subjectController.text.trim(),
              category: _selectedCategory!.trim(),
              priority: _selectedPriority.trim(),
              description: _descriptionController.text.trim(),
              contactPhone: _phoneController.text.trim().isNotEmpty
                  ? _phoneController.text.trim()
                  : null,
              preferredResponse: _selectedPreferredResponse,
            );

    if (!mounted) return;

    if (success) {
      context.go('/grievance/success');
    } else {
      _scrollController.animateTo(
        0,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final submissionState = ref.watch(grievanceSubmissionProvider);
    final isSubmitting = submissionState.isSubmitting;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: Column(
        children: [
          // Bespoke Emerald Gradient Header
          _buildHeader(context),

          // Scrollable Form Body
          Expanded(
            child: AbsorbPointer(
              absorbing: isSubmitting,
              child: SingleChildScrollView(
                controller: _scrollController,
                physics: const BouncingScrollPhysics(),
                padding: const EdgeInsets.fromLTRB(18, 18, 18, 28),
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Error Banner on Submission Failure
                      if (submissionState.hasError &&
                          submissionState.errorMessage != null) ...[
                        _buildErrorBanner(submissionState.errorMessage!),
                        const SizedBox(height: 14),
                      ],

                      // Card 1: Petition Classification & Subject
                      _buildPetitionDetailsCard(),

                      const SizedBox(height: 16),

                      // Card 2: Grievance Statement & Live Counter
                      _buildDescriptionCard(),

                      const SizedBox(height: 16),

                      // Card 3: Contact & Communication Preferences
                      _buildCommunicationCard(),

                      const SizedBox(height: 16),

                      // Card 4: Official Association Redressal Notice
                      _buildConfidentialityNotice(),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
      bottomNavigationBar: _buildBottomBar(context, isSubmitting),
    );
  }

  /// Builds the bespoke NammaTAGA multi-stop emerald gradient header
  /// with curved bottom corners, leaf watermark, back navigation, and branding.
  Widget _buildHeader(BuildContext context) {
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
          // Multi-stop Emerald Gradient (#04170D -> #0A2B1B -> #14452F -> #1D5A3D)
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

          // Ambient Radial Light Glow (Top-Right)
          Positioned(
            top: -40,
            right: -30,
            width: 200,
            height: 200,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF52B788).withValues(alpha: 0.26),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Ambient Soft Radial Glow (Bottom-Left)
          Positioned(
            bottom: -30,
            left: -20,
            width: 150,
            height: 150,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF2D6A4F).withValues(alpha: 0.28),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Subtle Leaf Watermark Motif
          Positioned(
            right: -10,
            bottom: -15,
            child: IgnorePointer(
              child: Icon(
                LucideIcons.leaf,
                size: 120,
                color: Colors.white.withValues(alpha: 0.045),
              ),
            ),
          ),

          // Header Foreground Content
          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 10, 16, 22),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      BouncyButton(
                        onTap: () {
                          HapticFeedback.lightImpact();
                          if (context.canPop()) {
                            context.pop();
                          } else {
                            context.go('/grievance');
                          }
                        },
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
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 5,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.12),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: Colors.white.withValues(alpha: 0.20),
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 6,
                              height: 6,
                              decoration: const BoxDecoration(
                                color: Color(0xFF52B788),
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 6),
                            const Text(
                              'State Redressal Portal',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 11.5,
                                fontWeight: FontWeight.w600,
                                letterSpacing: 0.3,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),

                  const SizedBox(height: 18),

                  Text(
                    'NAMMATAGA CIVIC CARE',
                    style: TextStyle(
                      color: const Color(0xFF95D5B2).withValues(alpha: 0.9),
                      fontSize: 10.5,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 1.2,
                    ),
                  ),

                  const SizedBox(height: 4),

                  const Text(
                    'Submit Grievance',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 24,
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.4,
                    ),
                  ),

                  const SizedBox(height: 4),

                  Text(
                    'Provide details of your grievance petition for association review.',
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.85),
                      fontSize: 13.5,
                      height: 1.35,
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

  /// Clean red/amber error banner displayed when backend submission fails.
  Widget _buildErrorBanner(String errorMessage) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF2F2),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFFCA5A5),
          width: 1.2,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x0AE11D48),
            blurRadius: 8,
            offset: Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: const BoxDecoration(
              color: Color(0xFFFEE2E2),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.alertCircle,
              color: Color(0xFFDC2626),
              size: 16,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Submission Failed',
                  style: TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF991B1B),
                    letterSpacing: -0.1,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  errorMessage,
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: Color(0xFFB91C1C),
                    height: 1.35,
                  ),
                ),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(
              LucideIcons.x,
              size: 16,
              color: Color(0xFF991B1B),
            ),
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(),
            onPressed: () {
              ref.read(grievanceSubmissionProvider.notifier).clearError();
            },
          ),
        ],
      ),
    );
  }

  /// Card 1: Petition details (Subject, Category, Priority).
  Widget _buildPetitionDetailsCard() {
    return _buildFormCard(
      title: 'Petition Classification',
      icon: LucideIcons.fileText,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 1. Subject Field
          _buildFieldLabel('Subject', isRequired: true),
          const SizedBox(height: 6),
          TextFormField(
            controller: _subjectController,
            textInputAction: TextInputAction.next,
            style: const TextStyle(
              fontSize: 14.5,
              color: AppColors.textPrimary,
              fontWeight: FontWeight.w500,
            ),
            decoration: _inputDecoration(
              hintText: 'Brief title of your grievance',
              prefixIcon: LucideIcons.fileText,
            ),
            validator: (value) {
              if (value == null || value.trim().isEmpty) {
                return 'Please enter a subject';
              }
              return null;
            },
          ),

          const SizedBox(height: 18),

          // 2. Category Selector Field
          _buildFieldLabel('Category', isRequired: true),
          const SizedBox(height: 6),
          _buildCategoryPickerTile(),
          if (_categoryError != null) ...[
            Padding(
              padding: const EdgeInsets.only(top: 6, left: 12),
              child: Text(
                _categoryError!,
                style: const TextStyle(
                  color: AppColors.destructive,
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],

          const SizedBox(height: 18),

          // 3. Priority Selector Field
          _buildFieldLabel('Priority Level', isRequired: true),
          const SizedBox(height: 6),
          _buildPriorityPickerTile(),
        ],
      ),
    );
  }

  /// Category selection button triggering Agent 4's [CategorySelectorSheet].
  Widget _buildCategoryPickerTile() {
    final hasSelection =
        _selectedCategory != null && _selectedCategory!.isNotEmpty;
    final hasError = _categoryError != null;

    return BouncyButton(
      onTap: _pickCategory,
      child: Container(
        constraints: const BoxConstraints(minHeight: 50),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: hasError
                ? AppColors.destructive
                : (hasSelection
                    ? AppColors.primary.withValues(alpha: 0.6)
                    : const Color(0xFFD6DFD9)),
            width: hasError ? 1.5 : (hasSelection ? 1.3 : 1.0),
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.02),
              blurRadius: 4,
              offset: const Offset(0, 1),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                color: hasSelection
                    ? AppColors.primary.withValues(alpha: 0.1)
                    : AppColors.muted,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                LucideIcons.tag,
                size: 16,
                color: hasSelection
                    ? AppColors.primary
                    : AppColors.textSecondary,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                hasSelection ? _selectedCategory! : 'Select Category...',
                style: TextStyle(
                  fontSize: 14.5,
                  fontWeight:
                      hasSelection ? FontWeight.w600 : FontWeight.normal,
                  color: hasSelection
                      ? AppColors.textPrimary
                      : const Color(0xFF8A9A90),
                ),
              ),
            ),
            const Icon(
              LucideIcons.chevronDown,
              size: 18,
              color: AppColors.textSecondary,
            ),
          ],
        ),
      ),
    );
  }

  /// Priority selection button triggering Agent 4's [PrioritySelectorSheet]
  /// and rendering a color-coded priority badge.
  Widget _buildPriorityPickerTile() {
    final style = _getPriorityStyle(_selectedPriority);
    final displayName = _formatPriorityLabel(_selectedPriority);

    return BouncyButton(
      onTap: _pickPriority,
      child: Container(
        constraints: const BoxConstraints(minHeight: 50),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: const Color(0xFFD6DFD9),
            width: 1.0,
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.02),
              blurRadius: 4,
              offset: const Offset(0, 1),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                color: style.bgColor,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                LucideIcons.flag,
                size: 16,
                color: style.textColor,
              ),
            ),
            const SizedBox(width: 12),

            // Color-coded priority badge
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: style.bgColor,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(
                  color: style.textColor.withValues(alpha: 0.3),
                  width: 1,
                ),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 7,
                    height: 7,
                    decoration: BoxDecoration(
                      color: style.textColor,
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    displayName,
                    style: TextStyle(
                      fontSize: 12.5,
                      fontWeight: FontWeight.w700,
                      color: style.textColor,
                      letterSpacing: 0.2,
                    ),
                  ),
                ],
              ),
            ),

            const Spacer(),

            Text(
              'Change',
              style: TextStyle(
                fontSize: 12.5,
                fontWeight: FontWeight.w600,
                color: AppColors.primary.withValues(alpha: 0.85),
              ),
            ),
            const SizedBox(width: 4),
            const Icon(
              LucideIcons.chevronDown,
              size: 16,
              color: AppColors.textSecondary,
            ),
          ],
        ),
      ),
    );
  }

  /// Card 2: Grievance description multi-line text input with live character counter.
  Widget _buildDescriptionCard() {
    final charCount = _descriptionController.text.trim().length;
    final isMinMet = charCount >= 50;

    return _buildFormCard(
      title: 'Grievance Description',
      icon: LucideIcons.alignLeft,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildFieldLabel('Detailed Description', isRequired: true),
          const SizedBox(height: 3),
          const Text(
            'Be specific. Include dates, locations, service cadre, or reference numbers.',
            style: TextStyle(
              fontSize: 12.5,
              color: AppColors.textSecondary,
              height: 1.3,
            ),
          ),
          const SizedBox(height: 8),

          TextFormField(
            controller: _descriptionController,
            minLines: 4,
            maxLines: 8,
            keyboardType: TextInputType.multiline,
            style: const TextStyle(
              fontSize: 14.5,
              color: AppColors.textPrimary,
              height: 1.45,
            ),
            decoration: _inputDecoration(
              hintText:
                  'Provide complete and detailed facts regarding your grievance (minimum 50 characters required)...',
              alignLabelWithHint: true,
            ),
            validator: (value) {
              final length = value?.trim().length ?? 0;
              if (length < 50) {
                return 'Description must be at least 50 characters';
              }
              return null;
            },
          ),

          const SizedBox(height: 8),

          // Live Character Counter
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Text(
                  '$charCount / 50 characters (min 50 required)',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isMinMet
                        ? const Color(0xFF2D6A4F)
                        : (charCount == 0
                            ? AppColors.textMuted
                            : const Color(0xFFD97706)),
                  ),
                ),
              ),
              if (isMinMet) ...[
                const SizedBox(width: 8),
                const Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      LucideIcons.checkCheck,
                      size: 14,
                      color: Color(0xFF2D6A4F),
                    ),
                    SizedBox(width: 4),
                    Text(
                      'Requirement met',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF2D6A4F),
                      ),
                    ),
                  ],
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }

  /// Card 3: Contact phone and selectable preferred response pills.
  Widget _buildCommunicationCard() {
    return _buildFormCard(
      title: 'Communication Preferences',
      icon: LucideIcons.messageSquare,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Contact Phone (Optional)
          _buildFieldLabel('Contact Phone', isRequired: false),
          const SizedBox(height: 6),
          TextFormField(
            controller: _phoneController,
            keyboardType: TextInputType.phone,
            textInputAction: TextInputAction.done,
            style: const TextStyle(
              fontSize: 14.5,
              color: AppColors.textPrimary,
              fontWeight: FontWeight.w500,
            ),
            decoration: _inputDecoration(
              hintText: '10-digit mobile number (optional)',
              prefixIcon: LucideIcons.phone,
            ),
            validator: (value) {
              final text = value?.trim() ?? '';
              if (text.isEmpty) {
                return null; // Optional
              }
              final digitsOnly = text.replaceAll(RegExp(r'\D'), '');
              if (digitsOnly.length != 10) {
                return 'Please enter a valid 10-digit mobile number';
              }
              return null;
            },
          ),

          const SizedBox(height: 20),

          // Preferred Response Channel (Required)
          _buildFieldLabel('Preferred Response Channel', isRequired: true),
          const SizedBox(height: 4),
          const Text(
            'Select how you wish to receive official correspondence regarding this petition.',
            style: TextStyle(
              fontSize: 12.5,
              color: AppColors.textSecondary,
              height: 1.3,
            ),
          ),
          const SizedBox(height: 12),

          // Segmented Pill Chips
          _buildPreferredResponseChips(),
        ],
      ),
    );
  }

  /// Segmented pill chips for 'email', 'phone', and 'letter'.
  Widget _buildPreferredResponseChips() {
    const options = [
      (
        value: 'email',
        label: 'Email',
        icon: LucideIcons.mail,
      ),
      (
        value: 'phone',
        label: 'Phone',
        icon: LucideIcons.phoneCall,
      ),
      (
        value: 'letter',
        label: 'Postal Letter',
        icon: LucideIcons.mailCheck,
      ),
    ];

    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: options.map((opt) {
        final isSelected = _selectedPreferredResponse == opt.value;

        return BouncyButton(
          onTap: () {
            HapticFeedback.lightImpact();
            setState(() {
              _selectedPreferredResponse = opt.value;
            });
            _syncFormData();
          },
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 180),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              color: isSelected ? const Color(0xFFE8F5E9) : Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: isSelected ? AppColors.primary : const Color(0xFFE2EBE5),
                width: isSelected ? 1.6 : 1.1,
              ),
              boxShadow: isSelected
                  ? [
                      BoxShadow(
                        color: AppColors.primary.withValues(alpha: 0.14),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ]
                  : [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.02),
                        blurRadius: 4,
                        offset: const Offset(0, 1),
                      ),
                    ],
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  opt.icon,
                  size: 16,
                  color: isSelected
                      ? AppColors.primary
                      : AppColors.textSecondary,
                ),
                const SizedBox(width: 8),
                Text(
                  opt.label,
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                    color: isSelected
                        ? AppColors.primaryDark
                        : AppColors.textPrimary,
                  ),
                ),
                if (isSelected) ...[
                  const SizedBox(width: 6),
                  const Icon(
                    LucideIcons.check,
                    size: 14,
                    color: AppColors.primary,
                  ),
                ],
              ],
            ),
          ),
        );
      }).toList(),
    );
  }

  /// Official association confidentiality reassurance card.
  Widget _buildConfidentialityNotice() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFE8F5E9).withValues(alpha: 0.6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFF95D5B2).withValues(alpha: 0.4),
          width: 1.0,
        ),
      ),
      child: const Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            LucideIcons.shieldCheck,
            size: 20,
            color: Color(0xFF2D6A4F),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Text(
              'Your grievance petition is treated with utmost discretion and forwarded directly to the TAGA State Executive Committee for inquiry.',
              style: TextStyle(
                fontSize: 12,
                color: Color(0xFF1B4332),
                height: 1.4,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Floating/fixed bottom submit bar with Safe Area, `BouncyButton`, and emerald gradient.
  Widget _buildBottomBar(BuildContext context, bool isSubmitting) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          top: BorderSide(
            color: const Color(0xFFE2EBE5).withValues(alpha: 0.9),
            width: 1.0,
          ),
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF1B4332).withValues(alpha: 0.06),
            blurRadius: 16,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      padding: const EdgeInsets.fromLTRB(18, 12, 18, 12),
      child: SafeArea(
        top: false,
        child: BouncyButton(
          onTap: isSubmitting ? null : _submitForm,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            width: double.infinity,
            height: 52,
            decoration: BoxDecoration(
              gradient: isSubmitting
                  ? const LinearGradient(
                      colors: [
                        Color(0xFF52796F),
                        Color(0xFF354F52),
                      ],
                    )
                  : const LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        Color(0xFF0D2C1B),
                        Color(0xFF1B4332),
                        Color(0xFF2D6A4F),
                      ],
                    ),
              borderRadius: BorderRadius.circular(16),
              boxShadow: isSubmitting
                  ? []
                  : [
                      BoxShadow(
                        color: const Color(0xFF0D2C1B).withValues(alpha: 0.35),
                        blurRadius: 16,
                        offset: const Offset(0, 6),
                      ),
                    ],
              border: Border.all(
                color: const Color(0xFF40916C).withValues(alpha: 0.4),
                width: 1.1,
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (isSubmitting) ...[
                  const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2.2,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  ),
                  const SizedBox(width: 12),
                  const Text(
                    'Submitting Grievance...',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 0.2,
                    ),
                  ),
                ] else ...[
                  const Icon(
                    LucideIcons.send,
                    size: 19,
                    color: Colors.white,
                  ),
                  const SizedBox(width: 10),
                  const Text(
                    'Submit Grievance',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                      letterSpacing: 0.2,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// Reusable card wrapper matching the NammaTAGA 1.1px border and 4% shadow specification.
  Widget _buildFormCard({
    required String title,
    required IconData icon,
    required Widget child,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: const Color(0xFFE2EBE5),
          width: 1.1,
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF1B4332).withValues(alpha: 0.04),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 30,
                height: 30,
                decoration: BoxDecoration(
                  color: AppColors.muted,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, size: 16, color: AppColors.primary),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  style: const TextStyle(
                    fontSize: 15.5,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.2,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          child,
        ],
      ),
    );
  }

  Widget _buildFieldLabel(String label, {required bool isRequired}) {
    return RichText(
      text: TextSpan(
        text: label,
        style: const TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w600,
          color: AppColors.textPrimary,
        ),
        children: [
          if (isRequired)
            const TextSpan(
              text: ' *',
              style: TextStyle(
                color: AppColors.destructive,
                fontWeight: FontWeight.bold,
              ),
            ),
        ],
      ),
    );
  }

  InputDecoration _inputDecoration({
    required String hintText,
    IconData? prefixIcon,
    bool alignLabelWithHint = false,
  }) {
    return InputDecoration(
      filled: true,
      fillColor: Colors.white,
      alignLabelWithHint: alignLabelWithHint,
      hintText: hintText,
      hintStyle: const TextStyle(
        color: Color(0xFF8A9A90),
        fontSize: 13.5,
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      prefixIcon: prefixIcon != null
          ? Icon(
              prefixIcon,
              size: 18,
              color: AppColors.primary,
            )
          : null,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: Color(0xFFD6DFD9), width: 1.0),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: Color(0xFFD6DFD9), width: 1.0),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
      ),
      errorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: AppColors.destructive, width: 1.0),
      ),
      focusedErrorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: AppColors.destructive, width: 1.5),
      ),
    );
  }

  /// Priority styling mapper aligned with [PrioritySelectorSheet].
  _PriorityStyle _getPriorityStyle(String key) {
    final normalized = key.toLowerCase().trim();
    if (normalized.contains('urgent')) {
      return const _PriorityStyle(
        textColor: Color(0xFFD4183D),
        bgColor: Color(0xFFFFEBEB),
      );
    } else if (normalized.contains('high')) {
      return const _PriorityStyle(
        textColor: Color(0xFFE07A5F),
        bgColor: Color(0xFFFFF3EB),
      );
    } else if (normalized.contains('low')) {
      return const _PriorityStyle(
        textColor: Color(0xFF6B7280),
        bgColor: Color(0xFFF3F4F6),
      );
    } else {
      return const _PriorityStyle(
        textColor: Color(0xFF2D6A4F),
        bgColor: Color(0xFFE8F5E9),
      );
    }
  }

  String _formatPriorityLabel(String value) {
    switch (value.toLowerCase().trim()) {
      case 'urgent':
        return 'Urgent';
      case 'high':
        return 'High';
      case 'low':
        return 'Low';
      case 'medium':
      default:
        return 'Medium';
    }
  }
}

class _PriorityStyle {
  final Color textColor;
  final Color bgColor;

  const _PriorityStyle({
    required this.textColor,
    required this.bgColor,
  });
}
