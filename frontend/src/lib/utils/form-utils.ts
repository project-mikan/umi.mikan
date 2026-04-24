import { invalidateAll } from "$app/navigation";
import type { ActionResult } from "@sveltejs/kit";
import { triggerHaptic } from "$lib/utils/haptic";

export function createSubmitHandler(
  setLoading: (loading: boolean) => void,
  setSaved?: (saved: boolean) => void,
) {
  return () => {
    setLoading(true);
    return async ({ result }: { result: ActionResult }) => {
      setLoading(false);
      if (result.type === "success") {
        triggerHaptic();
        await invalidateAll();
        if (setSaved) {
          setSaved(true);
          setTimeout(() => setSaved(false), 1000);
        }
      }
    };
  };
}
