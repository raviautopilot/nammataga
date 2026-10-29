# Namma TAGA Mobile Application — Creative UI/UX Phase-by-Phase Prompts

> **Project Name:** Namma TAGA Mobile App  
> **Target Platforms:** Android & iOS (Cross-Platform)  
> **Framework:** Flutter (Material 3 with iOS Cupertino Adaptations)  
> **Package Identifier / Bundle ID:** `com.ravi.taga`  
> **Directory Scope:** Strictly inside `nammataga-mobile/`  
> **Target Audience:** Association Members Only (Non-Admin) + Guest Public Browsing  
> **Design Theme:** Premium Emerald & Mint Forest Design System (`#2D6A4F`, `#1B4332`, `#FAFFFE`, `#D8F3DC`)  
> **UI Standard:** Advanced, Creative, High-End Motion & Tactile Micro-Interactions (No Minimalist/Basic Designs)  
> **Local Web Preview:** 19.5:9 Ultra-Realistic Phone Mockup Shell with Dynamic Island & Device Chassis  

---

## Creative Design Tokens & Motion Standards (Analyzed from Web & Elevated)

Every single screen must strictly implement these world-class mobile design standards. No basic or generic widgets are permitted.

### 1. Palette & Surface Tinting (Web Harmonized)
| Token Name | Value | Creative Application |
| :--- | :--- | :--- |
| **Primary (Forest Emerald)** | `#2D6A4F` | Gradient anchors, primary CTAs, active tab glow, app headers |
| **Deep Forest (Dark Emerald)** | `#1B4332` | High-contrast typography, hero gradient base, dark chassis |
| **Mint Off-White (Canvas)** | `#FAFFFE` | Clean, atmospheric screen canvas with soft ambient mint lighting |
| **Pure Surface (Cards)** | `#FFFFFF` | Frosted glass cards with `0.88` to `0.95` opacity |
| **Soft Mint (Highlight)** | `#D8F3DC` | Active chip fills, badge backgrounds, selected list highlights |
| **Sage Accent** | `#B7E4C7` | Avatar rings, category pill borders, subtle glowing highlights |
| **Muted Container** | `#F1F8E9` | Textfield fills, inactive container backgrounds, card insets |
| **Slate Olive (Subtitles)** | `#52796F` | Metadata, timestamps, captions, inactive navigation icons |
| **Focus / Ring Glow** | `#74C69D` | Active input border glow, notification pulse rings |
| **Destructive / Error** | `#D4183D` | Danger badges, cancellations, session warnings |
| **Emerald Ambient Shadow** | `Color(0x142D6A4F)` | Multi-layered 16–24px blur shadows tinted with primary green |
| **Frosted Glass Border** | `Color(0x262D6A4F)` | 1.0–1.2px border on all glass cards with 15% opacity |

### 2. Motion, Depth & Tactile Standards
- **Bouncy Button Scale:** Every button must incorporate an interactive press-down scale transformation (`0.96x` scale on tap down, spring release on tap up).
- **Tactile Haptic Feedback:** Trigger `HapticFeedback.lightImpact()` on tab switches, card taps, chips, and toggles; use `HapticFeedback.mediumImpact()` on form submissions and payment triggers.
- **Skeleton / Shimmer Loading:** Absolutely no plain centered spinning wheels for full page content. Use animated shimmering placeholders matching the exact geometry of incoming cards, avatars, and text blocks.
- **Cascading Entrance Animations:** Use `flutter_animate` on all screen sections with staggered fade-in, slide-up (15–30px), and subtle scale reveal (e.g., `delay: 50.ms * index`).
- **Glassmorphic Depth:** Use `BackdropFilter` with `ImageFilter.blur(sigmaX: 10, sigmaY: 10)`, subtle gradient fills (`Colors.white.withValues(alpha: 0.85)`), and thin emerald borders.

---

## Phase-by-Phase Execution Roadmap

The prompts below are designed to be executed strictly **one prompt at a time** with user confirmation before proceeding to the next.

1. [Prompt 00: Project Foundation, Architecture, Theming & Automation Script (`manage-app.sh`)](#prompt-00-project-foundation-architecture-theming--automation-script-manage-appsh)
2. [Prompt 01: Member Login Screen & Public Guest Mode Access](#prompt-01-member-login-screen--public-guest-mode-access)
3. [Prompt 02: Animated Splash Screen & Dynamic Auth Restoration](#prompt-02-animated-splash-screen--dynamic-auth-restoration)
4. [Prompt 03: Home Screen & Association Dashboard](#prompt-03-home-screen--association-dashboard)
5. [Prompt 04: Office Bearers & Leadership Directory Screen](#prompt-04-office-bearers--leadership-directory-screen)
6. [Prompt 05: Events & Multimedia Gallery Screen](#prompt-05-events--multimedia-gallery-screen)
7. [Prompt 06: Member Profile, Subscription & Announcements Screen](#prompt-06-member-profile-subscription--announcements-screen)
8. [Prompt 07: Member Grievance & Issue Tracking Screen](#prompt-07-member-grievance--issue-tracking-screen)
9. [Prompt 08: Member Resources & Document Center Screen](#prompt-08-member-resources--document-center-screen)
10. [Prompt 09: TAGA Towers Room & Dormitory Booking Screen](#prompt-09-taga-towers-room--dormitory-booking-screen)
11. [Prompt 10: 19.5:9 Web Simulation Shell & Bottom Navigation Shell](#prompt-10-1959-web-simulation-shell--bottom-navigation-shell)
12. [Prompt 11: Release Hardening, Play Store AAB, and APK Production Builds](#prompt-11-release-hardening-play-store-aab-and-apk-production-builds)

---

### Prompt 00: Project Foundation, Architecture, Theming & Automation Script (`manage-app.sh`)

```markdown
### TASK: Scaffold Flutter Mobile Project, Establish Advanced Design Tokens, Riverpod Architecture, and Automation Script (`manage-app.sh`)

#### Objectives:
1. Initialize the Flutter project strictly inside `nammataga-mobile/` with package name `com.ravi.taga` for both Android and iOS.
2. Establish Clean Feature-First modular architecture:
   - `core/constants/` (`app_colors.dart`, `app_assets.dart`, `api_endpoints.dart`, `app_typography.dart`)
   - `core/theme/` (`app_theme.dart` with Material 3, custom shadows, input styles, card borders)
   - `core/network/` (`api_client.dart` with Dio, JWT interceptor, 401 token expiration handler)
   - `core/storage/` (`secure_storage_service.dart` with flutter_secure_storage and shared_preferences)
   - `core/widgets/` (`web_phone_frame.dart`, `bouncy_button.dart`, `glass_card.dart`, `shimmer_placeholder.dart`)
3. Install and configure all required production dependencies:
   - `flutter_riverpod`, `go_router`, `dio`, `flutter_secure_storage`, `shared_preferences`, `flutter_animate`, `cached_network_image`, `razorpay_flutter`, `url_launcher`, `open_filex`, `path_provider`, `intl`, `lucide_icons`.
4. Copy branded assets into `nammataga-mobile/assets/images/`:
   - `tagatower.png` (from `taga-api/data/image/tagatower_final.png`)
   - `logo.png` (from `taga-api/data/image/logo/logo1.jpg`)
   - `objectiveimage.png` (from `taga-web/src/assets/objectiveimage.png`)
5. Implement the unified automation script `nammataga-mobile/manage-app.sh`:
   - Interactive CLI menu with ANSI color output.
   - Web view runner configured with 19.5:9 simulation frame on port 1702 (`start-web`, `stop-web`, `status-web`).
   - Interactive build options for Installable APK (Release/Debug) and Play Store AAB with environment selection (`local`, `dev`, `prod`).

#### Verification:
- `./manage-app.sh status-web` runs without errors.
- `flutter analyze` passes with zero issues.
- `flutter test` passes.
```

---

### Prompt 01: Member Login Screen & Public Guest Mode Access

```markdown
### TASK: Build High-End Creative Member Login Screen with Floating Glassmorphic Card, Animated Ambient Gradient & Guest Mode

#### Context & Source Reference:
- Web Component: `taga-web/src/components/MemberLogin.tsx`
- Web API: `taga-web/src/api/memberLoginApi.ts` & `taga-web/src/api/member.ts`
- Endpoints: `POST /api/member/login`, `POST /api/auth/member-forgot-password`, `POST /api/member/change-password`, `GET /api/membership/districts`

#### Creative UI/UX Design Specifications:
1. **Atmospheric Canvas & Ambient Glow:**
   - Background canvas: `#FAFFFE` layered with subtle animated radial glow spots (soft `#D8F3DC` emerald lights in top-right and bottom-left corners).
   - Top Header Section:
     * Association crest in a floating 3D glass card with soft drop shadow, accompanied by a dynamic shimmer badge: "Tamil Nadu Agricultural Graduates Association".
     * Welcoming typography: Display headline "Member Portal" (28pt, bold, `#1B4332`) with caption "Sign in to access your member benefits and association services" (14pt, `#52796F`).
2. **Floating Frosted Glassmorphism Login Card:**
   - Container: `BackdropFilter` with 12px blur, pure white fill with 90% opacity, 24px border radius, and delicate 1.2px tinted border (`rgba(45, 106, 79, 0.15)`).
   - Ambient Drop Shadow: `BoxShadow(color: Color(0x142D6A4F), blurRadius: 28, offset: Offset(0, 12))`.
3. **Inputs & Micro-Interactions:**
   - Email / TBF Number input:
     * Filled background (`#F1F8E9`), rounded 14px shape.
     * Prefix icon: `LucideIcons.mail` or `user` tinted with `#2D6A4F`.
     * Animated border glow on focus (`#74C69D` with 1.8px stroke).
   - Password input:
     * Toggleable eye icon (`LucideIcons.eye` / `eyeOff`) with bouncy spring reaction and tactile haptic feedback.
     * Forgot Password button: Right-aligned styled text link in bold emerald with subtle underline hover effect.
   - Remember Me toggle: Custom animated checkbox with mint fill and haptic vibration on switch.
4. **Primary CTA & Guest Mode Button:**
   - **"Sign In" Button:**
     * Multi-stop gradient fill: `LinearGradient(colors: [Color(0xFF1B4332), Color(0xFF2D6A4F)])`.
     * Subtle top-edge light reflection line.
     * Tap down scales to `0.96x` with `HapticFeedback.lightImpact()`.
     * Loading state: Smooth pulsing gradient shimmer with custom micro-loader, disabling duplicate taps.
   - **"Browse as Guest / Public Pages" Pill:**
     * Elegant secondary button with frosted white fill, 1.2px `#2D6A4F` border, and trailing arrow icon (`LucideIcons.arrowRight`).
     * Prominent explanatory subtitle: *"Explore Association Info, Office Bearers & Events without logging in"*.
     * Clicking routes user directly to Home dashboard in guest browsing mode (`isGuest = true`).
5. **Interactive Sheets:**
   - **Forgot Password Modal Sheet:**
     * Frosted bottom sheet with drag handle.
     * Segmented selector pill: "Registered Email ID" vs "TBF Number".
     * Live input validation and submit feedback with toast notification.
   - **First-Time Login (Force Change Password) Dialog:**
     * Triggered if `res.forceChangePassword == true`.
     * Multi-field security form (Old temporary password, New password, Confirm password) with live password strength meter indicator.

#### Files to Create:
- `lib/features/auth/data/auth_repository.dart`
- `lib/features/auth/domain/member_session.dart`
- `lib/features/auth/presentation/login_screen.dart`
- `lib/features/auth/presentation/widgets/ambient_background.dart`
- `lib/features/auth/presentation/widgets/forgot_password_sheet.dart`
- `lib/features/auth/presentation/widgets/change_password_sheet.dart`
- `lib/features/auth/providers/auth_provider.dart`

#### Verification & Acceptance:
- Smooth cascading entrance animations on load.
- Valid login saves JWT securely and updates Riverpod session state.
- "Browse as Guest" button grants instant read-only access to public routes.
- Tactile feedback triggers on all keypresses and button presses.
```

---

### Prompt 02: Animated Splash Screen & Dynamic Auth Restoration

```markdown
### TASK: Build Cinematic Animated Splash Screen with Breathing TAGA Tower Visual and Session Auto-Routing

#### Context & Source Reference:
- Asset: `assets/images/tagatower.png` (high-res architectural artwork)
- Web Session Logic: `taga-web/src/App.tsx` (token check, `isPaid` sync, 5-minute grace period buffer)
- API Check: `GET /api/member/profile`

#### Creative UI/UX Design Specifications:
1. **Cinematic Visual Hierarchy:**
   - Background: Deep Emerald Ambient Radiance (`LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: [Color(0xFF0F261C), Color(0xFF1B4332), Color(0xFF2D6A4F)])`).
   - Central Visual:
     * High-res TAGA Tower architectural illustration inside a floating circular aura with breathing radial pulse (`flutter_animate` repeating scale `1.0 -> 1.04` and soft blur glow).
     * Soft tinted spotlight illuminating the tower from above.
2. **Typography & Brand Identity:**
   - Association Crest / Tag: "TAGA" in bold high-contrast display lettering with gold/emerald gradient sheen.
   - Subtitle: "Tamil Nadu Agricultural Graduates Association" (medium 13pt with 1.2 letter spacing).
   - Motto reveal: Staggered slide-up animation: *"Empowering Agricultural Professionals Since 1971"*.
3. **Polished Progress Indicator:**
   - Ultra-thin (2.5px) pill progress bar at bottom with smooth gradient fill (`#52B788` to `#95D5B2`) and glowing lead head.
4. **Session Verification & Auto-Routing Engine:**
   - Query `StorageService.isTokenValid()`.
   - If token valid:
     * Fire background fetch to `GET /api/member/profile` to refresh profile data and paid subscription status (`isPaid` / `subscription_active`).
     * Smoothly fade transition to `HomeScreen` (`/home`).
   - If token missing, invalid, or expired:
     * Cleanly wipe obsolete tokens from encrypted storage.
     * Smoothly fade transition to `LoginScreen` (`/login`).
   - Enforce a minimum display duration of 1400ms so users appreciate the polished branding.

#### Files to Create:
- `lib/features/splash/presentation/splash_screen.dart`
- `lib/core/router/app_router.dart` (GoRouter with reactive redirect guard listening to `authProvider`)

#### Verification & Acceptance:
- Breathing animation runs fluidly at 60/120 FPS.
- Fresh app launch transitions without jitter to Login or Home screen based on token validity.
- Error in profile refresh defaults gracefully without blocking user navigation.
```

---

### Prompt 03: Home Screen & Association Dashboard

```markdown
### TASK: Build World-Class Home Screen & Association Dashboard with 3D Hero Banner, Live Stats Ticker, Glassmorphic Mission/Vision Cards & Interactive Quick-Actions

#### Context & Source Reference:
- Web Component: `taga-web/src/components/AboutUs.tsx`
- Backend Endpoints: `GET /api/public/about`, `GET /api/public/about/stats`, `GET /api/public/about/objectives`, `GET /api/public/about/contact`, `GET /api/logo`

#### Creative UI/UX Design Specifications:
1. **Interactive Hero Header Card:**
   - Multi-layered curved card with dark emerald gradient (`#1B4332` to `#2D6A4F`), subtle topographic contour pattern overlay.
   - Association Logo embedded in a frosted glass badge with established year pill (`Est. 1971` with subtle gold border).
   - Association Full Title and dynamic Tagline with smooth letter-spacing.
   - Quick member status chip: Displays "Active Member", "Annual Subscriber", or "Guest Visitor".
2. **Glassmorphic Live Statistics Counter:**
   - Horizontal scrolling card strip featuring real-time counters:
     * `3000+ Registered Members`
     * `50+ Association Events`
     * `38 District Committees`
     * `1 State Headquarters`
   - Shimmer skeleton loaders while fetching data.
   - Numbers animate counting up from 0 on first view (`TweenAnimationBuilder`).
3. **Quick Service Launcher Grid:**
   - 4-column quick action launcher with elevated circular icon containers (`Office Bearers`, `Events`, `Resources`, `Grievances`, `TAGA Towers`, `Profile`):
     * Glassmorphic circular icon pods with dual-tone icon tinting.
     * Micro-bounce on tap + haptic click.
     * Guest mode guard: Tapping member-restricted services (Resources, Grievances, TAGA Towers) opens an attractive frosted bottom sheet invitation to login.
4. **Mission & Vision Duo-Cards:**
   - Side-by-side or stacked frosted cards with 16px radius, subtle mint border, and distinct themed icons (`LucideIcons.target` for Mission, `LucideIcons.heartHandshake` for Vision).
   - Gradient icon badge backdrop.
5. **Interactive Objectives Journey Section:**
   - Visual numbered milestone cards with vertical connector line.
   - Glowing emerald step nodes (`#2D6A4F`) with title and expanded descriptive copy.
   - Side architectural preview thumbnail of TAGA Tower with rounded corners and soft elevation.
6. **Headquarters & Interactive Contact Card:**
   - Modern glassmorphic contact hub:
     * Direct tap action buttons:
       - Email (`LucideIcons.mail`) -> Launches email client.
       - Phone (`LucideIcons.phoneCall`) -> Launches device dialer.
       - Website (`LucideIcons.globe`) -> Launches official web portal.
       - Headquarters Address -> Tap to view Google Maps pin.

#### Files to Create:
- `lib/features/home/data/home_repository.dart`
- `lib/features/home/domain/association_info.dart`
- `lib/features/home/presentation/home_screen.dart`
- `lib/features/home/presentation/widgets/home_hero_card.dart`
- `lib/features/home/presentation/widgets/animated_stats_ticker.dart`
- `lib/features/home/presentation/widgets/service_launcher_grid.dart`
- `lib/features/home/presentation/widgets/objectives_timeline.dart`
- `lib/features/home/presentation/widgets/contact_glass_hub.dart`
- `lib/features/home/providers/home_provider.dart`

#### Verification & Acceptance:
- Pull-to-refresh smoothly re-fetches all home endpoints.
- Quick action shortcuts navigate to respective routes.
- Calling dialer, mail, and maps triggers native OS intents cleanly.
```

---

### Prompt 04: Office Bearers & Leadership Directory Screen

```markdown
### TASK: Build Magazine-Style Office Bearers & Leadership Directory Screen with Sticky District Filter Rail and Contact Action Triggers

#### Context & Source Reference:
- Web Component: `taga-web/src/components/OfficeBearers.tsx`
- Web API: `taga-web/src/api/OfficeBearers.ts`
- Endpoints: `GET /api/office-bearers/state-executive`, `GET /api/office-bearers/districts`, `GET /api/office-bearers/district-office-bearers?district={name}`

#### Creative UI/UX Design Specifications:
1. **Segmented Top Control Bar:**
   - Frosted pill switch: **[State Executive Committee]** | **[District Representatives]**.
   - Search Bar: Floating search box with rounded 16px shape, search icon, clear button, and real-time filtering by Name, Position, or District.
2. **State Executive Spotlight Cards:**
   - Prominent leadership cards featuring:
     * High-res circular portrait with dual-ring emerald gradient border (`#2D6A4F` to `#74C69D`).
     * Fallback initials monogram with gradient background if image is null.
     * Name in bold heading (18pt, `#1B4332`), Position badge in solid emerald pill (`#2D6A4F`), Tenure badge (`2023-2025` in `#D8F3DC`).
     * Direct floating contact action bar at the bottom of the card:
       - Phone icon button (dials phone).
       - Email icon button (opens mail client).
       - Information icon (opens detail bottom sheet).
3. **District Representatives & Sticky Filter Rail:**
   - Horizontal scrolling district chip rail with active count indicator (e.g. `Coimbatore (4)`, `Madurai (3)`, `Thanjavur (5)`).
   - "Filter by District" bottom sheet picker with instant search for quick district selection.
   - Grid or staggered list of district office bearers.
4. **Office Bearer Credential Bottom Sheet:**
   - Tapping any bearer card slides up an executive profile sheet showing:
     * Full photo/avatar, position, department, tenure, educational qualifications, professional experience, and direct contact options.
5. **Loading & Empty States:**
   - Shimmer placeholder cards replicating the bearer card layout with pulsing gradient.
   - Creative empty state illustration if search yields no results.

#### Files to Create:
- `lib/features/office_bearers/data/office_bearers_repository.dart`
- `lib/features/office_bearers/domain/office_bearer.dart`
- `lib/features/office_bearers/presentation/office_bearers_screen.dart`
- `lib/features/office_bearers/presentation/widgets/bearer_executive_card.dart`
- `lib/features/office_bearers/presentation/widgets/district_filter_rail.dart`
- `lib/features/office_bearers/presentation/widgets/bearer_detail_sheet.dart`
- `lib/features/office_bearers/providers/office_bearers_provider.dart`

#### Verification & Acceptance:
- State executive list loads with correct photos and position badges.
- Selecting different districts updates the list dynamically with smooth fade transition.
- Tapping phone or email buttons launches native phone dialer or email client with haptic confirmation.
```

---

### Prompt 05: Events & Multimedia Gallery Screen

```markdown
### TASK: Build Immersive Events & Multimedia Gallery Screen with Calendar Cards, Countdown Badges, and Full-Screen Lightbox

#### Context & Source Reference:
- Web Component: `taga-web/src/components/Events.tsx`
- Web API: `taga-web/src/api/event.ts`
- Endpoints: `GET /api/events/upcoming`, `GET /api/gallery/years`, `GET /api/gallery?year={year}`

#### Creative UI/UX Design Specifications:
1. **Curved Navigation Switcher:**
   - Pill tab bar: **[Upcoming Events]** | **[Association Gallery]** with smooth animated slide indicator and haptics.
2. **Upcoming Events Section:**
   - **Event Card Design:**
     * Left Date Stamp block: Dark emerald block with Day (`28`) in large bold white font and Month (`OCT`) in mint green text.
     * Title in bold headline with category tag (`General Body Meeting`, `Conference`, `Workshop`).
     * Location row with MapPin icon and Time with Clock icon.
     * Dynamic countdown pill: e.g., *"In 5 Days"* or *"Tomorrow"* with subtle pulse animation.
     * "Add to Device Calendar" action button triggering native device calendar event creation.
     * "View Details" sheet expanding complete agenda, speaker list, and location directions.
   - Clean empty state with calendar illustration when no upcoming events are scheduled.
3. **Multimedia Photo Gallery Section:**
   - **Year Filter Rail:**
     * Segmented chips for available gallery years (`2026`, `2025`, `2024`, etc.) with photo count badge.
   - **Staggered Masonry Image Grid:**
     * 2-column Pinterest-style masonry grid with rounded corners (14px) and subtle border.
     * High-performance `cached_network_image` with shimmering placeholder.
     * Image counter badge overlay on hover/tap.
4. **Interactive Full-Screen Lightbox:**
   - Tapping any photo smoothly expands it with a `Hero` transition into a full-screen dark lightbox.
   - Interactive gestures: Double-tap to zoom, pinch-to-zoom (up to 4x), swipe down to dismiss.
   - Bottom floating action bar: Image title/year, Share button, and Save/Download to gallery button.

#### Files to Create:
- `lib/features/events/data/events_repository.dart`
- `lib/features/events/domain/taga_event.dart`
- `lib/features/events/presentation/events_screen.dart`
- `lib/features/events/presentation/widgets/event_card.dart`
- `lib/features/events/presentation/widgets/event_detail_sheet.dart`
- `lib/features/events/presentation/widgets/gallery_masonry_grid.dart`
- `lib/features/events/presentation/widgets/full_screen_lightbox.dart`
- `lib/features/events/providers/events_provider.dart`

#### Verification & Acceptance:
- Events list displays accurately formatted dates and countdown badges.
- Year selector switches gallery photo sets smoothly.
- Full-screen lightbox supports fluid pinch-to-zoom and swipe-down dismissal.
```

---

### Prompt 06: Member Profile, Subscription & Announcements Screen

```markdown
### TASK: Build Luxurious Member Profile Screen with Digital Membership Card, Razorpay Annual Subscription & Announcements Feed

#### Context & Source Reference:
- Web Component: `taga-web/src/components/Membership.tsx`
- Web API: `taga-web/src/api/member.ts`, `taga-web/src/api/subscription.ts`
- Endpoints: `GET /api/member/profile`, `POST /api/member/edit-request`, `GET /api/subscriptions`, `GET /api/subscriptions/status?email={email}`, `POST /api/subscriptions/create-order`, `POST /api/subscriptions/verify-payment`, `GET /api/member/notifications`, `PUT /api/member/notifications/:id/read`

#### Creative UI/UX Design Specifications:
1. **Digital Association Identity Card (Top Hero):**
   - High-end digital membership card with emerald/gold gradient sheen (`#1B4332` to `#2D6A4F` with subtle gold shimmer edge).
   - Member Avatar with initials fallback and active verification checkmark.
   - Full Name, Designation, and TAGA ID in bold embossed font.
   - Active Status Badge: Emerald pill with glowing dot for "Active Member" or "Annual Subscriber".
   - QR Code / Barcode simulation representation at bottom corner with member ID for physical verification.
2. **Three-Tab Sub-Navigation Rail:**
   - **Tab 1: Member Profile**
     * Detailed information tiles with clean icons: TBF Number, CPS/GPF Number, Date of Birth, Gender, Working District, Native District, Recruitment Batch, Seniority Number, Residential & Permanent Addresses.
     * "Request Profile Correction" button -> Opens a modal bottom sheet with pre-filled inputs to submit corrections to admin (`POST /api/member/edit-request`).
   - **Tab 2: Subscriptions & Dues**
     * Subscription status card: Displays current validity, last paid date, and next due date.
     * Subscription packages list (Annual Subscription, Special Relief Fund, etc.).
     * "Pay Now / Renew" CTA:
       - Triggers backend order creation (`POST /api/subscriptions/create-order`).
       - Launches native Razorpay checkout sheet.
       - On success, verifies signature with backend (`POST /api/subscriptions/verify-payment`) and triggers celebratory confetti animation with refreshed status.
   - **Tab 3: Announcements & Notifications**
     * Notification feed with priority indicators (`Urgent` in crimson, `Important` in emerald).
     * Title, formatted timestamp (`2 hours ago`), and read/unread status dot.
     * Tap to view full announcement dialog and automatically mark as read.
3. **Guest Mode Lock Wall:**
   - If user is browsing in guest mode, display an attractive glass card prompting member login to access the digital ID and member profile.

#### Files to Create:
- `lib/features/profile/data/profile_repository.dart`
- `lib/features/profile/domain/member_profile.dart`
- `lib/features/profile/presentation/profile_screen.dart`
- `lib/features/profile/presentation/widgets/digital_membership_card.dart`
- `lib/features/profile/presentation/tabs/profile_info_tab.dart`
- `lib/features/profile/presentation/tabs/subscription_tab.dart`
- `lib/features/profile/presentation/tabs/announcements_tab.dart`
- `lib/features/profile/presentation/widgets/edit_profile_sheet.dart`
- `lib/features/profile/presentation/widgets/razorpay_checkout_handler.dart`
- `lib/features/profile/providers/profile_provider.dart`

#### Verification & Acceptance:
- Digital ID card renders with member details and status badge.
- Profile edit request submits to backend with toast confirmation.
- Subscription tab reflects current `isPaid` status.
- Unread notification badge updates dynamically.
```

---

### Prompt 07: Member Grievance & Issue Tracking Screen

```markdown
### TASK: Build Modern Member Grievance Submission Desk & Interactive Issue Tracking Screen

#### Context & Source Reference:
- Web Component: `taga-web/src/components/Grievance.tsx`
- Web API: `taga-web/src/api/Grievance.ts`
- Endpoints: `GET /api/categories`, `GET /api/priorities`, `POST /api/grievances`, `GET /api/grievances`, `GET /api/grievance-banner`

#### Creative UI/UX Design Specifications:
1. **Access Control Gate:**
   - Restricted to authenticated paid members (`isLoggedIn && isPaid`).
   - For unpaid or guest members, show an informative locked card explaining that Grievance filing is a subscriber benefit with a direct shortcut to the Subscription tab.
2. **Two-Segmented View:**
   - **Segment 1: File New Grievance**
     * Pre-filled Member Name, Email, and Phone from profile state.
     * Subject Line input with clear character counter.
     * Category Picker: Beautiful bottom-sheet selector loaded from `/api/categories` (Service Matters, Pay Anomaly, Transfer Request, etc.).
     * Priority Segmented Bar: Low (Green), Medium (Orange), High (Red) with animated highlight pill.
     * Detailed Description text area with minimum height and focus glow.
     * Preferred Response Method: Toggle buttons for "Email" vs "Phone Call".
     * "Submit Grievance" button with haptic feedback and confirmation animation.
   - **Segment 2: Ticket Tracking History**
     * Timeline-style tracking cards for previously submitted grievances:
       - Ticket ID, Category tag, and Submission Date.
       - Status badge:
         * `Open` (Golden amber pill with clock icon).
         * `In-Progress` (Azure blue pill with active spinner indicator).
         * `Resolved` (Emerald green pill with checkmark icon).
       - Expandable detail accordion showing ticket description, admin notes, and resolution remarks.
3. **Skeleton Loading & Feedback:**
   - Shimmer skeleton loaders for grievance history cards.
   - Toast banner on successful grievance creation.

#### Files to Create:
- `lib/features/grievance/data/grievance_repository.dart`
- `lib/features/grievance/domain/grievance_ticket.dart`
- `lib/features/grievance/presentation/grievance_screen.dart`
- `lib/features/grievance/presentation/widgets/file_grievance_form.dart`
- `lib/features/grievance/presentation/widgets/grievance_history_list.dart`
- `lib/features/grievance/presentation/widgets/grievance_ticket_card.dart`
- `lib/features/grievance/providers/grievance_provider.dart`

#### Verification & Acceptance:
- Submitting valid form sends correct payload to `POST /api/grievances`.
- Grievance history reflects active status tags.
- Non-paid members see polite subscription requirement card.
```

---

### Prompt 08: Member Resources & Document Center Screen

```markdown
### TASK: Build Premium Member Resources & Document Center with In-App PDF Scrubber, Category Rails, and Download Manager

#### Context & Source Reference:
- Web Component: `taga-web/src/components/Resources.tsx`
- Web API: `taga-web/src/api/resources.ts`
- Endpoints: `GET /api/resources`, `GET /api/resources/:id`, `GET /api/resources/external-links`

#### Creative UI/UX Design Specifications:
1. **Access Gate:**
   - Restricted to authenticated paid members (`isLoggedIn && isPaid`).
2. **Interactive Category & Subcategory Filter Rail:**
   - Horizontal scrolling category rail with glowing active indicator:
     * `Government Orders (G.O.)`
     * `Service Rules & Bylaws`
     * `Association Forms`
     * `Official Portals & Links`
   - Subcategory chips for targeted filtering (e.g. `Scheme GOs`, `Pay Commission`).
3. **Document Cards:**
   - Clean card layout with:
     * Document Icon container with PDF badge.
     * Title with high readability (15pt semibold, `#1B4332`).
     * Metadata tags: Year chip (`2026-2027`), Subcategory tag.
     * Action Buttons:
       - **"View PDF"** button (launches in-app PDF viewer).
       - **"Download"** icon button with circular progress indicator during download.
4. **In-App PDF Viewer:**
   - Embedded PDF viewer with pinch-to-zoom, page counter (`Page 3 of 12`), and thumbnail page scrubber.
   - Top action bar: Share PDF, Print, and Open with external app (`open_filex`).
5. **Curated Official Portals & Useful Links Tab:**
   - Visual cards for official Tamil Nadu Agricultural portals (Dept of Agriculture, TNAU, TNPSC, e-Thottam) with logo icon, URL domain badge, and external link launcher.

#### Files to Create:
- `lib/features/resources/data/resources_repository.dart`
- `lib/features/resources/domain/resource_document.dart`
- `lib/features/resources/presentation/resources_screen.dart`
- `lib/features/resources/presentation/widgets/resource_category_rail.dart`
- `lib/features/resources/presentation/widgets/document_card.dart`
- `lib/features/resources/presentation/widgets/pdf_viewer_screen.dart`
- `lib/features/resources/presentation/widgets/external_portal_card.dart`
- `lib/features/resources/providers/resources_provider.dart`

#### Verification & Acceptance:
- Documents fetch properly with JWT authorization.
- PDF download downloads to application cache and opens via device viewer.
- External links launch browser securely.
```

---

### Prompt 09: TAGA Towers Room & Dormitory Booking Screen

```markdown
### TASK: Build Hospitality-Grade TAGA Towers Booking Screen with Midnight Check-in Cycle, Interactive Dorm Bed Selector, and Boarding Pass Confirmation

#### Context & Source Reference:
- Web Component: `taga-web/src/components/TAGATowers.tsx`
- Web API: `taga-web/src/api/tagatower.ts`
- Endpoints: `GET /api/towers/rooms`, `GET /api/towers/availability`, `GET /api/towers/availability-range`, `POST /api/towers/bookings`, `GET /api/towers/bookings`, `DELETE /api/towers/bookings/:id`, `POST /api/towers/create-order`, `POST /api/towers/verify-payment`

#### Creative UI/UX Design Specifications:
1. **Hospitality Hero & Tower Overview:**
   - Hero banner with `assets/images/tagatower.png` architectural artwork.
   - Association Guest House badge: Location, Amenities (Air Conditioning, Wi-Fi, 24/7 Security, Elevator).
   - Prominent Timing Indicator: *"Check-in: 12:00 AM (Midnight) • Check-out: 12:00 AM (Midnight Cycle)"*.
2. **Booking Flow & Room Selection:**
   - **Date Range Picker:**
     * Dual calendar date picker showing Check-in Date and Check-out Date with night count tally.
   - **Room Carousel & Dormitory Bed Grid:**
     * Room cards: Gents Dormitory, Ladies Dormitory, Guest Private Rooms.
     * Strict Gender Restriction Guards: Alert modal preventing male members from booking Ladies Dormitory and vice versa.
     * Bed selection visual grid: Interactive bed icons showing Available (green), Selected (mint glow), and Booked (gray disabled).
3. **Advance Calculation & Razorpay Advance Payment:**
   - Pricing breakdown card:
     * Advance rate: ₹100/bed/night for Dormitory; ₹200/bed/night for Private Rooms.
     * Total Advance Payable calculated in real time.
     * Balance payable upon arrival at TAGA Towers reception.
   - "Confirm & Pay Advance" CTA triggers Razorpay checkout modal.
4. **"My Bookings" Boarding-Pass Cards:**
   - Active reservations formatted as stylish boarding-pass tickets with perforated divider and barcode.
   - Booking ID, Room/Bed number, Guest Name, Dates, and Payment status badge.
   - "Cancel Booking" option with cancellation confirmation dialog.

#### Files to Create:
- `lib/features/tagatowers/data/tagatowers_repository.dart`
- `lib/features/tagatowers/domain/room_booking.dart`
- `lib/features/tagatowers/presentation/tagatowers_screen.dart`
- `lib/features/tagatowers/presentation/widgets/towers_hero_header.dart`
- `lib/features/tagatowers/presentation/widgets/bed_selection_grid.dart`
- `lib/features/tagatowers/presentation/widgets/booking_summary_sheet.dart`
- `lib/features/tagatowers/presentation/widgets/boarding_pass_ticket.dart`
- `lib/features/tagatowers/providers/tagatowers_provider.dart`

#### Verification & Acceptance:
- Date selection updates bed availability from `/api/towers/availability-range`.
- Gender rules enforce proper room selection.
- Advance calculation matches web rules (₹100 dorm / ₹200 rooms).
```

---

### Prompt 10: 19.5:9 Web Simulation Shell & Bottom Navigation Shell

```markdown
### TASK: Build Ultra-Realistic 19.5:9 Smartphone Frame Mockup for Web Preview and Curved Floating Bottom Navigation Shell

#### Context & Source Reference:
- User Requirement: "i will be able to look the mobile appliaction version on web version on my computer before going to build an apk in specif 19.5:9 ratio with beautiful mobile web layout"

#### Creative UI/UX Design Specifications:
1. **19.5:9 Web Simulation Shell (`web_phone_frame.dart`):**
   - Active when `kIsWeb` or desktop browser is detected:
     * Exact 19.5:9 aspect ratio smartphone chassis (e.g. 420px width x 910px height).
     * Rounded 48px bezel corners with subtle 10px dark chassis border and dual-color ambient drop shadow.
     * Top Dynamic Island with camera lens reflection and speaker slit.
     * Bottom gesture home indicator bar.
     * Outer desktop canvas: Dark textured forest gradient (`#0B1D15` to `#1B4332`) with watermark and active environment chip ("LOCAL • http://localhost:1801/api").
   - On physical Android/iOS devices, the app renders edge-to-edge naturally.
2. **Curved Adaptive Bottom Navigation Shell:**
   - Floating glass navigation bar with rounded 24px pill geometry:
     1. **Home** (`LucideIcons.home`)
     2. **Office Bearers** (`LucideIcons.users`)
     3. **Events** (`LucideIcons.calendar`)
     4. **Services** (`LucideIcons.layoutGrid` -> Opens frosted bottom sheet with Resources, Grievances, TAGA Towers)
     5. **Profile** (`LucideIcons.userCheck`)
   - Animated spring indicator pill behind active tab in `#D8F3DC` with active icon in `#2D6A4F`.
   - Micro-haptic vibration on every tab change.
3. **Dynamic Brand App Bar:**
   - Left: Association crest in rounded square badge.
   - Center: Dynamic screen title.
   - Right: Notification Bell with live unread badge count + Member profile avatar.

#### Files to Create / Polish:
- `lib/core/widgets/web_phone_frame.dart`
- `lib/core/widgets/main_scaffold_shell.dart`
- `lib/core/widgets/services_modal_sheet.dart`
- `lib/core/widgets/brand_app_bar.dart`

#### Verification & Acceptance:
- Desktop web view shows realistic 19.5:9 mobile mockup.
- Tab navigation transitions smoothly between screens.
- Services button expands frosted sheet with direct navigation to all secondary features.
```

---

### Prompt 11: Release Hardening, Play Store AAB, and APK Production Builds

```markdown
### TASK: Configure Android/iOS Build Settings, Permissions, ProGuard, Play Store AAB & Production Signing

#### Context & Source Reference:
- Package Name: `com.ravi.taga`
- Build Artifacts: Installable APK (Universal) & Google Play Store App Bundle (`.aab`)

#### Requirements:
1. **Android Hardening:**
   - Package name verified: `com.ravi.taga`.
   - Android target SDK: 34, compile SDK: 34, min SDK: 23.
   - Verified permissions in `android/app/src/main/AndroidManifest.xml`:
     * `INTERNET`, `ACCESS_NETWORK_STATE`, `POST_NOTIFICATIONS`.
   - ProGuard rules configured for Dio, Riverpod, and Razorpay.
2. **iOS Configuration:**
   - Bundle Identifier set to `com.ravi.taga`.
   - Permissions strings configured in `Info.plist`.
3. **Execution via `manage-app.sh`:**
   - Test `./manage-app.sh build` with options:
     * 1) Release APK -> Outputs `build/app/outputs/flutter-apk/app-release.apk`
     * 2) Debug APK -> Outputs `build/app/outputs/flutter-apk/app-debug.apk`
     * 3) Play Store Bundle -> Outputs `build/app/outputs/bundle/release/app-release.aab`
   - Verify environment injection switches URLs cleanly between Localhost, Dev, and Prod.

#### Verification & Acceptance:
- Production APK compiles successfully with zero errors.
- Output APK installs cleanly with bundle ID `com.ravi.taga`.
```

---

## Execution Rule Summary

1. Execute **strictly one prompt at a time**.
2. After finishing each phase:
   - Summarize what was built and tested.
   - Stop and wait for explicit user confirmation before moving to the next prompt.
3. Use `./manage-app.sh start-web` to inspect the 19.5:9 web layout on your computer at any stage.
4. Use `./manage-app.sh build` to create installable APKs or Play Store AAB bundles.
