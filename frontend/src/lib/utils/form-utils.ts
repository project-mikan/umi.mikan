import type { ActionResult } from "$app/forms";
import { refreshAll } from "$app/navigation";

export function createSubmitHandler(
  setLoading: (loading: boolean) => void,
  setSaved?: (saved: boolean) => void,
) {
  return () => {
    setLoading(true);
    return async ({ result }: { result: ActionResult }) => {
      setLoading(false);
      if (result.type === "success") {
        await refreshAll();
        if (setSaved) {
          setSaved(true);
          setTimeout(() => setSaved(false), 1000);
        }
      }
    };
  };
}
