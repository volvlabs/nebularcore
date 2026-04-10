module.exports = {
  branches: [
    "main",
    "master",
    "+([0-9])?(.{+([0-9]),x}).x",
    { name: "develop", prerelease: "beta" },
    { name: "alpha", prerelease: "alpha" },
    {
      name: "/^feature\\/.+$/",
      prerelease: process.env.RELEASE_PREID || "feature",
    },
  ],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/exec",
      { verifyReleaseCmd: "echo ${nextRelease.version} > VERSION.txt" },
    ],
    "@semantic-release/github",
  ],
};
