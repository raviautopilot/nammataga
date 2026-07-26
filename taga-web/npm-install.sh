#!/bin/bash
set -e

echo "Cleaning old dependencies..."
rm -rf node_modules package-lock.json

echo "Installing all dependencies from package.json..."
npm install

echo "Ensuring required dev dependencies exist..."

# Example of installing a single, specific package:
  npm install @radix-ui/react-accordion
  npm install lucide-react
  npm install -D @vitejs/plugin-react@4
  npm install -D @tailwindcss/vite
  npm install tw-animate-css

# Example of installing a dev dependency:
 npm install --save-dev typescript
 npm install --save-dev @types/react