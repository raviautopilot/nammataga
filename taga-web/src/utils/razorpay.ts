/**
 * Helper to dynamically load the Razorpay checkout.js script.
 * Resolves with true if successfully loaded or already present, false otherwise.
 */
export const loadRazorpayScript = (): Promise<boolean> => {
  return new Promise((resolve) => {
    // If Razorpay is already available globally, resolve immediately
    if ((window as any).Razorpay) {
      resolve(true);
      return;
    }

    // Check if script has already been injected to avoid duplicates
    const existingScript = document.querySelector(
      'script[src="https://checkout.razorpay.com/v1/checkout.js"]'
    );

    if (existingScript) {
      // Set event listeners on existing script in case it's still loading
      const handleLoad = () => {
        cleanup();
        resolve(true);
      };
      const handleError = () => {
        cleanup();
        resolve(false);
      };
      const cleanup = () => {
        existingScript.removeEventListener("load", handleLoad);
        existingScript.removeEventListener("error", handleError);
      };

      existingScript.addEventListener("load", handleLoad);
      existingScript.addEventListener("error", handleError);
      return;
    }

    // Inject the script element
    const script = document.createElement("script");
    script.src = "https://checkout.razorpay.com/v1/checkout.js";
    script.async = true;
    script.onload = () => resolve(true);
    script.onerror = () => resolve(false);
    document.body.appendChild(script);
  });
};
