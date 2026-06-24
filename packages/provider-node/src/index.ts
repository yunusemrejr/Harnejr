import { ProviderProfile } from "@harnejr/shared";

export type SmokeResult = {
  providerId: string;
  ok: boolean;
  reason: string;
};

export async function smokeTestProvider(profile: ProviderProfile): Promise<SmokeResult> {
  if (!profile.enabled) {
    return { providerId: profile.id, ok: false, reason: "provider is disabled" };
  }
  if (profile.authMode !== "none" && !profile.apiKeyEnv && !profile.apiKeySecretRef) {
    return { providerId: profile.id, ok: false, reason: "provider has no apiKeyEnv or apiKeySecretRef" };
  }
  return { providerId: profile.id, ok: true, reason: "static profile validation passed" };
}
