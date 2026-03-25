"use client";

import { useRouter } from "next/navigation";
import { useToast } from "@/src/components/ui/toaster";
import { isClientError } from "@/src/lib/api/errors";

export function useApiToasts() {
  const { push } = useToast();
  const router = useRouter();

  return {
    showSuccess(title: string, description?: string) {
      push({ title, description });
    },
    showError(err: unknown, fallbackTitle = "Something went wrong") {
      if (isClientError(err)) {
        push({ title: err.code || fallbackTitle, description: err.message, variant: "destructive" });
        if (err.status === 401) {
          router.replace("/login");
        }
        return;
      }
      push({ title: fallbackTitle, description: "Unexpected error.", variant: "destructive" });
    }
  };
}

