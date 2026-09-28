import type { NextConfig } from "next";

import "./src/lib/env";

const nextConfig: NextConfig = {
  output: "standalone",
};

export default nextConfig;
