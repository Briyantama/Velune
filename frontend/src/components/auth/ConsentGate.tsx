"use client";

import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "@/src/store/hooks";
import { setConsent } from "@/src/store/slices/consentSlice";
import { ConsentModal } from "@/src/components/ConsentModal";
import { clearAllLocalSessionArtifacts, getConsentModeClient, setConsentModeCookie } from "@/src/services/authStorage";
import type { StorageMode } from "@/src/store/slices/consentSlice";
import { useAuthBootstrap } from "@/src/features/auth/hooks/useAuthBootstrap";

export function ConsentGate({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch();
  const consentStatus = useAppSelector((s) => s.consent.status);
  const storageMode = useAppSelector((s) => s.consent.storageMode);

  useAuthBootstrap(storageMode);

  const applyConsent = (mode: StorageMode) => {
    if (mode === "localStorage") {
      // Ensure no stale local tokens remain from a previous preference.
      clearAllLocalSessionArtifacts();
    }
    setConsentModeCookie(mode);
    dispatch(setConsent({ storageMode: mode }));
  };

  useEffect(() => {
    const existing = getConsentModeClient();
    if (existing) dispatch(setConsent({ storageMode: existing }));
  }, [dispatch]);

  const isUnknown = consentStatus === "unknown" || !storageMode;

  return (
    <>
      <ConsentModal
        open={isUnknown}
        onAccept={() => applyConsent("cookie")}
        onReject={() => applyConsent("localStorage")}
      />
      {/* Block interaction until consent is set so login can store tokens correctly. */}
      {isUnknown ? null : children}
    </>
  );
}

