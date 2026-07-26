# Dynamic Razorpay SDK Loading Walkthrough

To prevent loading the Razorpay checkout script (and its chunk dependencies) on initial startup when payments are disabled or unused, we have successfully implemented dynamic loading of the script on-demand.

## Changes Made

### 1. Frontend Core
- **Removed static `<script>` tag** in [index.html](file:///home/ubuntu/code/github/nammataga/taga-web/index.html) that loaded `checkout.js` unconditionally.

### 2. Utilities
- **Created a script injection utility** in [razorpay.ts](file:///home/ubuntu/code/github/nammataga/taga-web/src/utils/razorpay.ts) to:
  - Check if the global object `(window as any).Razorpay` already exists.
  - Check if the script tag has already been injected to prevent duplicate loads.
  - Dynamically append the script element to the DOM and resolve/reject a Promise on completion.

### 3. Component Integrations
- **TAGA Towers**: Modified `openRazorpay` in [TAGATowers.tsx](file:///home/ubuntu/code/github/nammataga/taga-web/src/components/TAGATowers.tsx) to dynamically load the Razorpay SDK right when the booking checkout flow is initiated.
- **Membership**: Modified `handlePaymentSubmit` in [Membership.tsx](file:///home/ubuntu/code/github/nammataga/taga-web/src/components/Membership.tsx) to load the SDK only when the payment flow is started.

---

## Verification Results

### Automated Production Build
- Checked compilation and bundle generation using `npm run build:prod`.
- The compilation completed successfully with no lint or type-checking issues.
