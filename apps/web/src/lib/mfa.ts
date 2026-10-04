import { twoFactor } from "better-auth/plugins";

// Apply Better Auth's session-removing challenge hook to our Firebase bridge too.
// This keeps token validation, challenge cookies, TOTP and recovery handling in
// Better Auth instead of creating a second MFA implementation.
export function hospitalMfa() {
  const plugin = twoFactor({
    issuer: "ULSHMS",
    skipVerificationOnEnable: false,
    allowPasswordless: false,
    twoFactorCookieMaxAge: 300,
    trustDeviceMaxAge: 0,
    accountLockout: {
      enabled: true,
      maxFailedAttempts: 5,
      durationSeconds: 900,
    },
  });
  return {
    ...plugin,
    hooks: {
      ...plugin.hooks,
      after: plugin.hooks.after.map((hook) => ({
        ...hook,
        matcher: (context: Parameters<typeof hook.matcher>[0]) =>
          hook.matcher(context) || context.path === "/firebase/sign-in",
      })),
    },
  };
}
