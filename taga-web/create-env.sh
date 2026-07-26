#!/bin/bash

# Create a JavaScript file that exposes the VITE_API_BASE_URL environment variable
# This file will be loaded by the browser to configure the API endpoint at runtime

# 1. Start creating the 'env-config.js' file in the 'public' directory.
#    The 'public' directory is served as-is, so the file will be accessible at /env-config.js.
echo "window._env_ = {" > public/env-config.js

# 2. Append the VITE_API_BASE_URL to the configuration object.
#    The value is taken from the shell's environment variable '$VITE_API_BASE_URL'.
#    This allows the API URL to be set dynamically when the container/server starts.
echo "  VITE_API_BASE_URL: '$VITE_API_BASE_URL'," >> public/env-config.js

# 3. Close the JavaScript object.
echo "};" >> public/env-config.js
