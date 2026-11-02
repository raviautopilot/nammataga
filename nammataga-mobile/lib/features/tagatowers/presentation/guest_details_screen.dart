import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';
import '../domain/booking_model.dart';
import '../domain/room_model.dart';
import '../providers/booking_flow_provider.dart';
import 'theme/tagatowers_design_tokens.dart';
import 'widgets/booking_step_progress_bar.dart';

/// Form state representation for an individual guest form card
class _GuestFormEntry {
  final TextEditingController nameController;
  final TextEditingController ageController;
  final TextEditingController contactController;
  String gender; // 'male' | 'female' | ''
  bool isSelfApplied;

  String? nameError;
  String? ageError;
  String? contactError;
  String? genderError;

  _GuestFormEntry({
    required this.nameController,
    required this.ageController,
    required this.contactController,
    this.gender = '',
  }) : isSelfApplied = false;

  void dispose() {
    nameController.dispose();
    ageController.dispose();
    contactController.dispose();
  }

  void clear({String defaultGender = ''}) {
    nameController.clear();
    ageController.clear();
    contactController.clear();
    gender = defaultGender;
    isSelfApplied = false;
    nameError = null;
    ageError = null;
    contactError = null;
    genderError = null;
  }
}

/// Step 2: Guest Details Screen (Agent 7)
///
/// Features:
/// 1. Wizard Progress Header (`BookingStepProgressBar(currentStep: 2)`)
/// 2. Exactly N dynamic guest form cards based on `bedCount` (Guest 1 to Guest N)
/// 3. Fields per card: Full Name (min 3 chars), Age (1-120), Contact (+91 prefix), Gender segmented buttons
/// 4. Guest 1 Quick "Fill as Self" Toggle with DOB age calculation from `MemberProfile`
/// 5. Comprehensive Gender Restriction Validation (Dormitories, partial occupancy, mixed couples)
/// 6. Riverpod sync with `bookingFlowNotifierProvider.setGuestDetails()` and GoRouter navigation
class GuestDetailsScreen extends ConsumerStatefulWidget {
  const GuestDetailsScreen({super.key});

  @override
  ConsumerState<GuestDetailsScreen> createState() => _GuestDetailsScreenState();
}

class _GuestDetailsScreenState extends ConsumerState<GuestDetailsScreen> {
  final List<_GuestFormEntry> _guestForms = [];
  final ScrollController _scrollController = ScrollController();
  final DateFormat _dateFormatter = DateFormat('dd MMM yyyy');

  @override
  void initState() {
    super.initState();
    _initializeFormEntries();
  }

  @override
  void dispose() {
    for (final entry in _guestForms) {
      entry.dispose();
    }
    _scrollController.dispose();
    super.dispose();
  }

  /// Initializes form controllers from current booking flow state and handles data preservation
  void _initializeFormEntries() {
    final flowState = ref.read(bookingFlowNotifierProvider);
    final room = flowState.selectedRoom;
    final avail = flowState.roomAvailability;

    final targetBedCount = flowState.bedCount > 0 ? flowState.bedCount : 1;
    final defaultGender = _getRoomDefaultGender(room, avail);

    for (int i = 0; i < targetBedCount; i++) {
      final existing = i < flowState.guestDetails.length ? flowState.guestDetails[i] : null;

      final name = existing?.name ?? '';
      final ageStr = (existing != null && existing.age > 0) ? existing.age.toString() : '';

      // Format contact nicely
      var rawContact = existing?.contact ?? '';
      if (rawContact.startsWith('+91 ')) {
        rawContact = rawContact.replaceFirst('+91 ', '');
      } else if (rawContact.startsWith('+91')) {
        rawContact = rawContact.replaceFirst('+91', '');
      }

      final gender = (existing != null && existing.gender.isNotEmpty)
          ? existing.gender
          : defaultGender;

      final entry = _GuestFormEntry(
        nameController: TextEditingController(text: name),
        ageController: TextEditingController(text: ageStr),
        contactController: TextEditingController(text: rawContact),
        gender: gender,
      );

      // Detect if Guest 1 was already self-filled
      if (i == 0 && name.isNotEmpty) {
        final profile = ref.read(memberProfileProvider).value;
        if (profile != null) {
          final cleanProfileName = profile.name.trim().toLowerCase();
          final cleanEntryName = name.trim().toLowerCase();
          if (cleanProfileName == cleanEntryName ||
              profile.fullNameWithInitial.trim().toLowerCase() == cleanEntryName) {
            entry.isSelfApplied = true;
          }
        }
      }

      _guestForms.add(entry);
    }
  }

  /// Calculates real age in years from date of birth string
  int _calculateAge(String? dobStr) {
    if (dobStr == null || dobStr.trim().isEmpty) return 0;
    try {
      final clean = dobStr.trim();
      DateTime? birthDate;
      if (clean.contains('-')) {
        final parts = clean.split('-');
        if (parts[0].length == 4) {
          birthDate = DateTime.tryParse(clean);
        } else if (parts.length == 3) {
          final day = int.tryParse(parts[0]);
          final month = int.tryParse(parts[1]);
          final year = int.tryParse(parts[2]);
          if (day != null && month != null && year != null) {
            birthDate = DateTime(year, month, day);
          }
        }
      } else if (clean.contains('/')) {
        final parts = clean.split('/');
        if (parts[0].length == 4) {
          birthDate = DateTime.tryParse(clean.replaceAll('/', '-'));
        } else if (parts.length == 3) {
          final day = int.tryParse(parts[0]);
          final month = int.tryParse(parts[1]);
          final year = int.tryParse(parts[2]);
          if (day != null && month != null && year != null) {
            birthDate = DateTime(year, month, day);
          }
        }
      } else {
        birthDate = DateTime.tryParse(clean);
      }

      if (birthDate == null) return 0;
      final today = DateTime.now();
      var age = today.year - birthDate.year;
      if (today.month < birthDate.month ||
          (today.month == birthDate.month && today.day < birthDate.day)) {
        age--;
      }
      return age > 0 ? age : 0;
    } catch (_) {
      return 0;
    }
  }

  /// Returns the default locked gender for dormitories or existing room restrictions
  String _getRoomDefaultGender(Room? room, RoomAvailability? avail) {
    if (room == null) return '';
    final type = room.type.toLowerCase();
    final id = room.id.toLowerCase();
    if (type == 'gents-dorm' || id == 'gents-dorm') return 'male';
    if (type == 'ladies-dorm' || id == 'ladies-dorm') return 'female';
    if (avail?.genderRestriction != null && avail!.genderRestriction!.isNotEmpty) {
      return avail.genderRestriction!.toLowerCase();
    }
    return '';
  }

  /// Returns effective gender restriction if applicable
  String? _getEffectiveGenderRestriction(Room? room, RoomAvailability? avail) {
    if (room == null) return null;
    final type = room.type.toLowerCase();
    final id = room.id.toLowerCase();
    if (type == 'gents-dorm' || id == 'gents-dorm') return 'male';
    if (type == 'ladies-dorm' || id == 'ladies-dorm') return 'female';
    if (avail?.genderRestriction != null && avail!.genderRestriction!.isNotEmpty) {
      return avail.genderRestriction!.toLowerCase();
    }
    return null;
  }

  /// Quick toggle to auto-fill Guest 1 with authenticated member profile or clear it
  void _toggleGuest1SelfFill() {
    HapticFeedback.lightImpact();
    if (_guestForms.isEmpty) return;

    final guest1 = _guestForms[0];
    final flowState = ref.read(bookingFlowNotifierProvider);
    final room = flowState.selectedRoom;
    final avail = flowState.roomAvailability;
    final defaultGender = _getRoomDefaultGender(room, avail);

    if (guest1.isSelfApplied) {
      // Toggle off: Clear Guest 1
      setState(() {
        guest1.clear(defaultGender: defaultGender);
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Guest 1 cleared'),
          duration: Duration(seconds: 2),
        ),
      );
      return;
    }

    // Toggle on: Populate from MemberProfile or session
    final profile = ref.read(memberProfileProvider).value;
    final session = ref.read(authProvider).value;

    final name = profile?.fullNameWithInitial.isNotEmpty == true
        ? profile!.fullNameWithInitial
        : (profile?.name.isNotEmpty == true
            ? profile!.name
            : (session?.displayName ?? ''));

    var mobile = profile?.mobileNumber ?? session?.user?['mobileNumber']?.toString() ?? '';
    mobile = mobile.replaceAll(RegExp(r'\D'), '');
    if (mobile.length > 10 && mobile.startsWith('91')) {
      mobile = mobile.substring(2);
    }

    final dob = profile?.dateOfBirth ?? session?.user?['dateOfBirth']?.toString();
    final calculatedAge = _calculateAge(dob);

    final rawGender = (profile?.gender.isNotEmpty == true
            ? profile!.gender
            : (session?.user?['gender']?.toString() ?? 'male'))
        .toLowerCase();
    final memberGender = rawGender == 'female' ? 'female' : 'male';

    final restriction = _getEffectiveGenderRestriction(room, avail);
    if (restriction != null) {
      if (restriction == 'male' && memberGender == 'female') {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Your profile is Female, but this room is restricted to Male guests only.'),
            backgroundColor: Colors.redAccent,
          ),
        );
        return;
      }
      if (restriction == 'female' && memberGender == 'male') {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Your profile is Male, but this room is restricted to Female guests only.'),
            backgroundColor: Colors.redAccent,
          ),
        );
        return;
      }
    }

    setState(() {
      guest1.nameController.text = name;
      guest1.contactController.text = mobile;
      if (calculatedAge > 0) {
        guest1.ageController.text = calculatedAge.toString();
      }
      guest1.gender = restriction ?? memberGender;
      guest1.isSelfApplied = true;
      guest1.nameError = null;
      guest1.ageError = null;
      guest1.contactError = null;
      guest1.genderError = null;
    });

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Guest 1 auto-filled from your profile'),
        duration: Duration(seconds: 2),
        backgroundColor: TagaTowersDesignTokens.primaryGreen,
      ),
    );
  }

  /// Validates all input fields, room restrictions, and navigates to Step 3 (Review)
  void _onProceedToReview() {
    HapticFeedback.mediumImpact();
    final flowState = ref.read(bookingFlowNotifierProvider);
    final room = flowState.selectedRoom;
    final avail = flowState.roomAvailability;
    final restriction = _getEffectiveGenderRestriction(room, avail);

    bool hasValidationError = false;

    for (int i = 0; i < _guestForms.length; i++) {
      final form = _guestForms[i];
      final name = form.nameController.text.trim();
      final ageText = form.ageController.text.trim();
      final contact = form.contactController.text.trim();
      final age = int.tryParse(ageText) ?? 0;

      String? nameErr;
      String? ageErr;
      String? contactErr;
      String? genderErr;

      // 1. Name validation
      if (name.isEmpty) {
        nameErr = 'Full name is required';
        hasValidationError = true;
      } else if (name.length < 3) {
        nameErr = 'Name must be at least 3 characters';
        hasValidationError = true;
      }

      // 2. Age validation
      if (ageText.isEmpty) {
        ageErr = 'Age is required';
        hasValidationError = true;
      } else if (age < 1 || age > 120) {
        ageErr = 'Age must be between 1 and 120';
        hasValidationError = true;
      }

      // 3. Contact validation
      final cleanDigits = contact.replaceAll(RegExp(r'\D'), '');
      if (contact.isEmpty) {
        contactErr = 'Contact number is required';
        hasValidationError = true;
      } else if (cleanDigits.length < 10) {
        contactErr = 'Enter a valid 10-digit mobile number';
        hasValidationError = true;
      }

      // 4. Gender selection
      final effectiveGender = restriction ?? form.gender;
      if (effectiveGender.isEmpty) {
        genderErr = 'Please select a gender';
        hasValidationError = true;
      }

      setState(() {
        form.nameError = nameErr;
        form.ageError = ageErr;
        form.contactError = contactErr;
        form.genderError = genderErr;
      });
    }

    if (hasValidationError) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please correct the highlighted errors before continuing'),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    // Gender restriction checks across all guests
    final hasMale = _guestForms.any((f) => (restriction ?? f.gender) == 'male');
    final hasFemale = _guestForms.any((f) => (restriction ?? f.gender) == 'female');

    final isGentsDorm = room?.type == 'gents-dorm' || room?.id == 'gents-dorm';
    final isLadiesDorm = room?.type == 'ladies-dorm' || room?.id == 'ladies-dorm';

    if (isGentsDorm && hasFemale) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Gents Dormitory is strictly for male guests only.'),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    if (isLadiesDorm && hasMale) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Ladies Dormitory is strictly for female guests only.'),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    if (restriction != null && restriction.isNotEmpty) {
      if (restriction == 'male' && hasFemale) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'This room is partially occupied by male guests — only male guests can book remaining beds.',
            ),
            backgroundColor: Colors.redAccent,
          ),
        );
        return;
      }
      if (restriction == 'female' && hasMale) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'This room is partially occupied by female guests — only female guests can book remaining beds.',
            ),
            backgroundColor: Colors.redAccent,
          ),
        );
        return;
      }
      if (hasMale && hasFemale) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'This room is partially occupied by $restriction guests — mixed groups cannot share with existing single occupants.',
            ),
            backgroundColor: Colors.redAccent,
          ),
        );
        return;
      }
    }

    // Dormitories cannot have mixed guests even if fully available
    if ((isGentsDorm || isLadiesDorm) && hasMale && hasFemale) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Dormitories cannot accommodate mixed gender bookings.'),
          backgroundColor: Colors.redAccent,
        ),
      );
      return;
    }

    // Construct final GuestDetail objects
    final guests = _guestForms.map((f) {
      final cleanContact = f.contactController.text.trim();
      final formattedContact = cleanContact.startsWith('+91 ')
          ? cleanContact
          : '+91 $cleanContact';

      return GuestDetail(
        name: f.nameController.text.trim(),
        age: int.tryParse(f.ageController.text.trim()) ?? 0,
        contact: formattedContact,
        gender: restriction ?? f.gender,
      );
    }).toList();

    // Persist details to Riverpod booking flow state
    ref.read(bookingFlowNotifierProvider.notifier).setGuestDetails(guests);

    // Proceed to Step 3: Review
    context.push('/taga-towers/booking/review');
  }

  @override
  Widget build(BuildContext context) {
    final flowState = ref.watch(bookingFlowNotifierProvider);
    final room = flowState.selectedRoom;
    final avail = flowState.roomAvailability;
    final restriction = _getEffectiveGenderRestriction(room, avail);
    final isDorm = room?.type == 'gents-dorm' || room?.type == 'ladies-dorm';

    final totalBeds = flowState.bedCount > 0 ? flowState.bedCount : 1;
    final advanceAmount = flowState.advanceAmount;

    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        title: const Text(
          'Step 2: Guest Details',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            color: TagaTowersDesignTokens.deepEmerald,
            fontSize: 18,
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          onPressed: () => context.pop(),
        ),
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                controller: _scrollController,
                physics: const BouncingScrollPhysics(),
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // 1. Wizard Progress Bar
                    BookingStepProgressBar(
                      currentStep: 2,
                      onStepTapped: (step) {
                        if (step < 2) {
                          context.pop();
                        }
                      },
                    ),
                    const SizedBox(height: 14),

                    // 2. Room Summary & Booking Dates Banner
                    if (room != null) _buildRoomSummaryCard(room, flowState),

                    // 3. Gender Restriction Banner (if applicable)
                    if (restriction != null || isDorm) ...[
                      const SizedBox(height: 12),
                      _buildGenderNoticeBanner(room, restriction),
                    ],

                    const SizedBox(height: 16),

                    // 4. Section Header with Guest Count
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Row(
                          children: [
                            const Icon(
                              LucideIcons.users,
                              size: 18,
                              color: TagaTowersDesignTokens.primaryGreen,
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'Guest Information ($totalBeds required)',
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                          ],
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                          decoration: BoxDecoration(
                            color: TagaTowersDesignTokens.softMint,
                            borderRadius: BorderRadius.circular(20),
                            border: Border.all(
                              color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.5),
                            ),
                          ),
                          child: Text(
                            '$totalBeds Bed${totalBeds > 1 ? 's' : ''}',
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: TagaTowersDesignTokens.deepEmerald,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),

                    // 5. Dynamic N Guest Form Cards
                    for (int i = 0; i < _guestForms.length; i++) ...[
                      _buildGuestCard(index: i, entry: _guestForms[i], restriction: restriction),
                      const SizedBox(height: 14),
                    ],

                    const SizedBox(height: 20),
                  ],
                ),
              ),
            ),

            // 6. Bottom Sticky Navigation & Price Breakdown Bar
            _buildBottomActionBar(totalBeds: totalBeds, advanceAmount: advanceAmount),
          ],
        ),
      ),
    );
  }

  /// Compact header summarizing selected room, stay dates, and stay policy
  Widget _buildRoomSummaryCard(Room room, BookingFlowState flowState) {
    final checkIn = flowState.checkInDate != null
        ? _dateFormatter.format(flowState.checkInDate!)
        : 'N/A';
    final checkOut = flowState.checkOutDate != null
        ? _dateFormatter.format(flowState.checkOutDate!)
        : 'N/A';
    final nights = flowState.totalNights;

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 14,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  LucideIcons.hotel,
                  size: 18,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      room.name,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    Text(
                      TagaTowersDesignTokens.getRoomDisplayName(room.type),
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.mutedBackground,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  '$nights Night${nights > 1 ? 's' : ''}',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          const Divider(height: 1, color: Color(0x1A2D6A4F)),
          const SizedBox(height: 8),
          Row(
            children: [
              const Icon(LucideIcons.calendar, size: 13, color: AppColors.textSecondary),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  '$checkIn → $checkOut (12:00 AM Midnight Cycle)',
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                    color: AppColors.textPrimary,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// Informative banner warning the user about dormitory or partial occupancy gender restriction
  Widget _buildGenderNoticeBanner(Room? room, String? restriction) {
    final isGentsDorm = room?.type == 'gents-dorm' || room?.id == 'gents-dorm';
    final isLadiesDorm = room?.type == 'ladies-dorm' || room?.id == 'ladies-dorm';

    String title;
    String description;
    Color bgColor;
    Color borderColor;
    Color textColor;
    IconData icon;

    if (isGentsDorm) {
      title = 'Gents Dormitory Policy';
      description = 'This dormitory is strictly for Male guests only. Mixed gender bookings are not permitted.';
      bgColor = const Color(0xFFEFF6FF);
      borderColor = const Color(0xFFBFDBFE);
      textColor = const Color(0xFF1E40AF);
      icon = LucideIcons.shieldAlert;
    } else if (isLadiesDorm) {
      title = 'Ladies Dormitory Policy';
      description = 'This dormitory is strictly for Female guests only. Mixed gender bookings are not permitted.';
      bgColor = const Color(0xFFFDF2F8);
      borderColor = const Color(0xFFFBCFE8);
      textColor = const Color(0xFF9D174D);
      icon = LucideIcons.shieldAlert;
    } else if (restriction != null) {
      title = 'Partial Occupancy Gender Restriction';
      description =
          'This room is partially occupied by ${restriction.toLowerCase()} guests. Only ${restriction.toLowerCase()} guests can book the remaining beds.';
      bgColor = TagaTowersDesignTokens.warningAmberLight;
      borderColor = TagaTowersDesignTokens.warningAmberBorder;
      textColor = TagaTowersDesignTokens.warningAmberDark;
      icon = LucideIcons.alertCircle;
    } else {
      return const SizedBox.shrink();
    }

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: borderColor),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 18, color: textColor),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.bold,
                    color: textColor,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  description,
                  style: TextStyle(
                    fontSize: 12,
                    color: textColor.withValues(alpha: 0.9),
                    height: 1.35,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Builds an individual guest card for Guest [index + 1]
  Widget _buildGuestCard({
    required int index,
    required _GuestFormEntry entry,
    required String? restriction,
  }) {
    final isGuest1 = index == 0;
    final isLockedByRoom = restriction != null && restriction.isNotEmpty;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Card Header: Guest index badge and Guest 1 "Fill as Self" pill toggle
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  CircleAvatar(
                    radius: 14,
                    backgroundColor: entry.isSelfApplied
                        ? TagaTowersDesignTokens.primaryGreen
                        : TagaTowersDesignTokens.softMint,
                    child: Text(
                      '${index + 1}',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.bold,
                        color: entry.isSelfApplied
                            ? Colors.white
                            : TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Text(
                    'Guest ${index + 1}',
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                  if (entry.isSelfApplied) ...[
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: TagaTowersDesignTokens.softMint,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: const Text(
                        'Self',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: TagaTowersDesignTokens.primaryGreen,
                        ),
                      ),
                    ),
                  ],
                ],
              ),

              // Guest 1 "Fill as Self" pill button
              if (isGuest1) _buildFillAsSelfButton(entry.isSelfApplied),
            ],
          ),
          const SizedBox(height: 14),

          // Field 1: Full Name
          _buildTextField(
            controller: entry.nameController,
            label: 'Full Name *',
            hint: 'Enter full name (min 3 chars)',
            icon: LucideIcons.user,
            errorText: entry.nameError,
            onChanged: (val) {
              if (entry.nameError != null && val.trim().length >= 3) {
                setState(() => entry.nameError = null);
              }
              if (entry.isSelfApplied) {
                setState(() => entry.isSelfApplied = false);
              }
            },
          ),
          const SizedBox(height: 12),

          // Row with Age & Contact Number
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Field 2: Age (1-120)
              Expanded(
                flex: 2,
                child: _buildTextField(
                  controller: entry.ageController,
                  label: 'Age *',
                  hint: 'Age',
                  icon: LucideIcons.calendar,
                  keyboardType: TextInputType.number,
                  inputFormatters: [
                    FilteringTextInputFormatter.digitsOnly,
                    LengthLimitingTextInputFormatter(3),
                  ],
                  errorText: entry.ageError,
                  onChanged: (val) {
                    final age = int.tryParse(val.trim());
                    if (entry.ageError != null && age != null && age >= 1 && age <= 120) {
                      setState(() => entry.ageError = null);
                    }
                  },
                ),
              ),
              const SizedBox(width: 12),

              // Field 3: Contact Number (+91 prefix)
              Expanded(
                flex: 4,
                child: _buildPhoneField(
                  controller: entry.contactController,
                  errorText: entry.contactError,
                  onChanged: (val) {
                    final digits = val.replaceAll(RegExp(r'\D'), '');
                    if (entry.contactError != null && digits.length >= 10) {
                      setState(() => entry.contactError = null);
                    }
                    if (entry.isSelfApplied) {
                      setState(() => entry.isSelfApplied = false);
                    }
                  },
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Field 4: Gender Segmented Button
          _buildGenderSelector(
            guestIndex: index,
            currentGender: isLockedByRoom ? restriction : entry.gender,
            isLocked: isLockedByRoom,
            lockedGender: restriction,
            errorText: entry.genderError,
            onGenderSelected: (gender) {
              if (isLockedByRoom) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(
                      'Gender is locked to ${restriction.toUpperCase()} by room policy.',
                    ),
                    duration: const Duration(seconds: 2),
                  ),
                );
                return;
              }
              HapticFeedback.lightImpact();
              setState(() {
                entry.gender = gender;
                entry.genderError = null;
              });
            },
          ),
        ],
      ),
    );
  }

  /// Guest 1 "Fill as Self" Pill Button with animated toggle feedback
  Widget _buildFillAsSelfButton(bool isApplied) {
    return BouncyButton(
      onTap: _toggleGuest1SelfFill,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: isApplied
              ? TagaTowersDesignTokens.primaryGreen
              : TagaTowersDesignTokens.softMint.withValues(alpha: 0.5),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isApplied
                ? TagaTowersDesignTokens.primaryGreen
                : TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.4),
          ),
          boxShadow: isApplied
              ? [
                  BoxShadow(
                    color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.25),
                    blurRadius: 6,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              isApplied ? LucideIcons.check : LucideIcons.userCheck,
              size: 14,
              color: isApplied ? Colors.white : TagaTowersDesignTokens.primaryGreen,
            ),
            const SizedBox(width: 5),
            Text(
              isApplied ? '✓ Self Applied' : 'Fill as Self',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: isApplied ? Colors.white : TagaTowersDesignTokens.primaryGreen,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Clean styled Text Input Field
  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    required IconData icon,
    TextInputType keyboardType = TextInputType.text,
    List<TextInputFormatter>? inputFormatters,
    String? errorText,
    ValueChanged<String>? onChanged,
  }) {
    final hasError = errorText != null && errorText.isNotEmpty;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 6),
        TextField(
          controller: controller,
          keyboardType: keyboardType,
          inputFormatters: inputFormatters,
          onChanged: onChanged,
          style: const TextStyle(fontSize: 14, color: AppColors.textPrimary),
          decoration: InputDecoration(
            hintText: hint,
            hintStyle: TextStyle(fontSize: 13, color: Colors.grey.shade400),
            prefixIcon: Icon(icon, size: 16, color: TagaTowersDesignTokens.primaryGreen),
            contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            filled: true,
            fillColor: const Color(0xFFF9FBF9),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(
                color: hasError ? Colors.redAccent : Colors.grey.shade300,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(
                color: hasError ? Colors.redAccent : Colors.grey.shade300,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(
                color: hasError ? Colors.redAccent : TagaTowersDesignTokens.primaryGreen,
                width: 1.5,
              ),
            ),
          ),
        ),
        if (hasError) ...[
          const SizedBox(height: 4),
          Row(
            children: [
              const Icon(LucideIcons.alertCircle, size: 12, color: Colors.redAccent),
              const SizedBox(width: 4),
              Expanded(
                child: Text(
                  errorText,
                  style: const TextStyle(fontSize: 11, color: Colors.redAccent),
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }

  /// Contact number input with dedicated `+91 ` prefix badge
  Widget _buildPhoneField({
    required TextEditingController controller,
    String? errorText,
    ValueChanged<String>? onChanged,
  }) {
    final hasError = errorText != null && errorText.isNotEmpty;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Contact Number *',
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 6),
        TextField(
          controller: controller,
          keyboardType: TextInputType.phone,
          inputFormatters: [
            FilteringTextInputFormatter.digitsOnly,
            LengthLimitingTextInputFormatter(10),
          ],
          onChanged: onChanged,
          style: const TextStyle(fontSize: 14, color: AppColors.textPrimary),
          decoration: InputDecoration(
            hintText: '9876543210',
            hintStyle: TextStyle(fontSize: 13, color: Colors.grey.shade400),
            prefixIcon: Container(
              padding: const EdgeInsets.only(left: 10, right: 6),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(LucideIcons.phone, size: 15, color: TagaTowersDesignTokens.primaryGreen),
                  const SizedBox(width: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.6),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: const Text(
                      '+91',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  Container(width: 1, height: 16, color: Colors.grey.shade300),
                ],
              ),
            ),
            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
            filled: true,
            fillColor: const Color(0xFFF9FBF9),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(
                color: hasError ? Colors.redAccent : Colors.grey.shade300,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(
                color: hasError ? Colors.redAccent : Colors.grey.shade300,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(
                color: hasError ? Colors.redAccent : TagaTowersDesignTokens.primaryGreen,
                width: 1.5,
              ),
            ),
          ),
        ),
        if (hasError) ...[
          const SizedBox(height: 4),
          Row(
            children: [
              const Icon(LucideIcons.alertCircle, size: 12, color: Colors.redAccent),
              const SizedBox(width: 4),
              Expanded(
                child: Text(
                  errorText,
                  style: const TextStyle(fontSize: 11, color: Colors.redAccent),
                ),
              ),
            ],
          ),
        ],
      ],
    );
  }

  /// Segmented gender selector supporting Male/Female with room lock enforcement
  Widget _buildGenderSelector({
    required int guestIndex,
    required String currentGender,
    required bool isLocked,
    required String? lockedGender,
    required String? errorText,
    required ValueChanged<String> onGenderSelected,
  }) {
    final hasError = errorText != null && errorText.isNotEmpty;
    final isMale = currentGender == 'male';
    final isFemale = currentGender == 'female';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Text(
              'Gender *',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
            ),
            if (isLocked) ...[
              const SizedBox(width: 6),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.amber.shade50,
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: Colors.amber.shade200),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(LucideIcons.lock, size: 10, color: Colors.amber.shade800),
                    const SizedBox(width: 3),
                    Text(
                      'Locked: ${lockedGender?.toUpperCase()}',
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        color: Colors.amber.shade800,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        ),
        const SizedBox(height: 8),

        Row(
          children: [
            // Male Option
            Expanded(
              child: _buildGenderOption(
                label: 'Male',
                icon: LucideIcons.user,
                isSelected: isMale,
                isDisabled: isLocked && lockedGender == 'female',
                onTap: () => onGenderSelected('male'),
              ),
            ),
            const SizedBox(width: 10),

            // Female Option
            Expanded(
              child: _buildGenderOption(
                label: 'Female',
                icon: LucideIcons.user,
                isSelected: isFemale,
                isDisabled: isLocked && lockedGender == 'male',
                onTap: () => onGenderSelected('female'),
              ),
            ),
          ],
        ),

        if (hasError) ...[
          const SizedBox(height: 4),
          Row(
            children: [
              const Icon(LucideIcons.alertCircle, size: 12, color: Colors.redAccent),
              const SizedBox(width: 4),
              Text(
                errorText,
                style: const TextStyle(fontSize: 11, color: Colors.redAccent),
              ),
            ],
          ),
        ],
      ],
    );
  }

  /// Individual gender button item
  Widget _buildGenderOption({
    required String label,
    required IconData icon,
    required bool isSelected,
    required bool isDisabled,
    required VoidCallback onTap,
  }) {
    return BouncyButton(
      onTap: isDisabled
          ? () {
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text('$label is prohibited by this room\'s gender restriction.'),
                  duration: const Duration(seconds: 2),
                ),
              );
            }
          : onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(vertical: 10),
        decoration: BoxDecoration(
          color: isDisabled
              ? Colors.grey.shade100
              : (isSelected
                  ? TagaTowersDesignTokens.primaryGreen
                  : Colors.white),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(
            color: isDisabled
                ? Colors.grey.shade300
                : (isSelected
                    ? TagaTowersDesignTokens.primaryGreen
                    : Colors.grey.shade300),
            width: isSelected ? 1.5 : 1.0,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.2),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              isDisabled ? LucideIcons.lock : icon,
              size: 15,
              color: isDisabled
                  ? Colors.grey.shade400
                  : (isSelected ? Colors.white : AppColors.textPrimary),
            ),
            const SizedBox(width: 6),
            Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                color: isDisabled
                    ? Colors.grey.shade400
                    : (isSelected ? Colors.white : AppColors.textPrimary),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Bottom Sticky Action Bar with price preview and continue button
  Widget _buildBottomActionBar({required int totalBeds, required int advanceAmount}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: const Border(top: BorderSide(color: Color(0x1A2D6A4F))),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 10,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      child: Row(
        children: [
          // Left: Advance Breakdown Summary
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                const Text(
                  'Advance Payable',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w500,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 2),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.baseline,
                  textBaseline: TextBaseline.alphabetic,
                  children: [
                    Text(
                      '₹$advanceAmount',
                      style: const TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '($totalBeds bed${totalBeds > 1 ? 's' : ''})',
                      style: const TextStyle(
                        fontSize: 11,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Right: "Proceed to Review" Bouncy Button
          Expanded(
            child: BouncyButton(
              onTap: _onProceedToReview,
              child: Container(
                height: 48,
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [
                      TagaTowersDesignTokens.primaryGreen,
                      TagaTowersDesignTokens.deepEmerald,
                    ],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [
                    BoxShadow(
                      color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.3),
                      blurRadius: 8,
                      offset: const Offset(0, 3),
                    ),
                  ],
                ),
                alignment: Alignment.center,
                child: const Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Review Booking',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    SizedBox(width: 6),
                    Icon(LucideIcons.arrowRight, size: 16, color: Colors.white),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
